---
module: background-workers
version: "1.0.0"
last_updated: "2026-04-02"
category: infrastructure
---

# Background Workers

BackgroundService, IHostedService, Wolverine Scheduled Jobs.

## Quick Reference

```csharp
// BackgroundService (주기적 작업)
public class DataSyncWorker : BackgroundService { ... }

// IHostedService (시작/종료 시 실행)
public class StartupInitializer : IHostedService { ... }

// Wolverine Scheduled Job (크론 기반)
opts.Publish(pub => pub.Message<CleanupJob>()
    .ScheduledAt(new CronExpression("0 0 2 * * ?")));
```

---

## Detailed Patterns

### BackgroundService

```csharp
public class DataSyncWorker(
    IServiceScopeFactory scopeFactory,
    ILogger logger) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken ct)
    {
        logger.Information("DataSyncWorker 시작");

        while (!ct.IsCancellationRequested)
        {
            try
            {
                using var scope = scopeFactory.CreateScope();
                var syncService = scope.ServiceProvider
                    .GetRequiredService<IDataSyncService>();

                await syncService.SyncAsync(ct);
                logger.Information("데이터 동기화 완료");
            }
            catch (OperationCanceledException) when (ct.IsCancellationRequested)
            {
                break; // 정상 종료
            }
            catch (Exception ex)
            {
                logger.Error(ex, "데이터 동기화 실패");
            }

            await Task.Delay(TimeSpan.FromMinutes(30), ct);
        }

        logger.Information("DataSyncWorker 종료");
    }
}

// 등록
builder.Services.AddHostedService<DataSyncWorker>();
```

### Wolverine Scheduled Jobs

```csharp
// Job 메시지 정의
public record CleanupExpiredSessions;

// Handler
public static class CleanupExpiredSessionsHandler
{
    public static async Task Handle(
        CleanupExpiredSessions command,
        AppDbContext db,
        ILogger logger,
        CancellationToken ct)
    {
        var cutoff = DateTimeOffset.UtcNow.AddHours(-24);

        var expired = await db.Sessions
            .Where(s => s.LastActivity < cutoff)
            .ExecuteDeleteAsync(ct);

        logger.Information("Cleaned up {Count} expired sessions", expired);
    }
}

// 스케줄 등록
builder.Host.UseWolverine(opts =>
{
    opts.Publish(pub =>
    {
        pub.Message<CleanupExpiredSessions>()
            .ScheduledAt(new CronExpression("0 0 2 * * ?"));  // 매일 02:00
    });

    // 다른 스케줄
    opts.Publish(pub =>
    {
        pub.Message<GenerateDailyReport>()
            .ScheduledAt(new CronExpression("0 0 6 * * ?"));  // 매일 06:00
    });
});
```

### Channel-Based Producer/Consumer

```csharp
public class QueueProcessorService(
    Channel<WorkItem> channel,
    IServiceScopeFactory scopeFactory,
    ILogger logger) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken ct)
    {
        await foreach (var item in channel.Reader.ReadAllAsync(ct))
        {
            try
            {
                using var scope = scopeFactory.CreateScope();
                var processor = scope.ServiceProvider
                    .GetRequiredService<IWorkItemProcessor>();

                await processor.ProcessAsync(item, ct);
            }
            catch (Exception ex)
            {
                logger.Error(ex, "Failed to process work item {ItemId}", item.Id);
            }
        }
    }
}

// 등록
builder.Services.AddSingleton(Channel.CreateBounded<WorkItem>(
    new BoundedChannelOptions(100)
    {
        FullMode = BoundedChannelFullMode.Wait
    }));
builder.Services.AddHostedService<QueueProcessorService>();

// Producer (다른 서비스에서 아이템 추가)
public class OrderService(Channel<WorkItem> channel)
{
    public async Task EnqueueAsync(WorkItem item, CancellationToken ct)
        => await channel.Writer.WriteAsync(item, ct);
}
```

---

## Advanced Topics

### Graceful Shutdown

```csharp
public class GracefulWorker(ILogger logger) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken ct)
    {
        // CancellationToken을 존중하여 정상 종료
        ct.Register(() =>
            logger.Information("Shutdown requested, finishing current work..."));

        while (!ct.IsCancellationRequested)
        {
            await DoWorkAsync(ct);
            await Task.Delay(TimeSpan.FromSeconds(10), ct);
        }
    }

    public override async Task StopAsync(CancellationToken ct)
    {
        logger.Information("GracefulWorker stopping...");
        await base.StopAsync(ct);
        logger.Information("GracefulWorker stopped");
    }
}
```

### Health Check for Background Services

```csharp
public class WorkerHealthCheck(DataSyncWorker worker) : IHealthCheck
{
    public Task<HealthCheckResult> CheckHealthAsync(
        HealthCheckContext context,
        CancellationToken ct = default)
    {
        if (worker.LastRunAt is null)
            return Task.FromResult(HealthCheckResult.Degraded("Never run"));

        var timeSinceLastRun = DateTimeOffset.UtcNow - worker.LastRunAt.Value;
        return timeSinceLastRun > TimeSpan.FromHours(1)
            ? Task.FromResult(HealthCheckResult.Unhealthy(
                $"Last run {timeSinceLastRun.TotalMinutes:F0} minutes ago"))
            : Task.FromResult(HealthCheckResult.Healthy());
    }
}
```

### Wolverine Durable Messaging

```csharp
// 서버 재시작 시에도 메시지 유실 방지
builder.Host.UseWolverine(opts =>
{
    opts.PersistMessagesWithSqlServer(connectionString);

    // 스케줄된 작업도 Durable
    opts.Publish(pub =>
    {
        pub.Message<CleanupExpiredSessions>()
            .ScheduledAt(new CronExpression("0 0 2 * * ?"));
    });

    opts.Policies.UseDurableLocalQueues();
});
```
