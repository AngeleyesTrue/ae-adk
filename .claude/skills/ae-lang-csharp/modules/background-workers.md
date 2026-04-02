---
module: background-workers
version: "1.0.0"
last_updated: "2026-04-02"
category: infrastructure
---

# Background Workers

BackgroundService, IHostedService, Wolverine Scheduled/Recurring Messages.

## Quick Reference

```csharp
// BackgroundService (주기적 작업)
public class DataSyncWorker : BackgroundService { ... }

// IHostedService (시작/종료 시 실행)
public class StartupInitializer : IHostedService { ... }

// Wolverine 반복 작업 (BackgroundService + IMessageBus)
public class ScheduledCleanupService(IServiceScopeFactory scopeFactory) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken ct)
    {
        using var timer = new PeriodicTimer(TimeSpan.FromHours(1));
        while (await timer.WaitForNextTickAsync(ct))
        {
            using var scope = scopeFactory.CreateScope();
            var bus = scope.ServiceProvider.GetRequiredService<IMessageBus>();
            await bus.PublishAsync(new CleanupJob());
        }
    }
}
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
        logger.LogInformation("DataSyncWorker 시작");

        while (!ct.IsCancellationRequested)
        {
            try
            {
                using var scope = scopeFactory.CreateScope();
                var syncService = scope.ServiceProvider
                    .GetRequiredService<IDataSyncService>();

                await syncService.SyncAsync(ct);
                logger.LogInformation("데이터 동기화 완료");
            }
            catch (OperationCanceledException) when (ct.IsCancellationRequested)
            {
                break; // 정상 종료
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "데이터 동기화 실패");
            }

            await Task.Delay(TimeSpan.FromMinutes(30), ct);
        }

        logger.LogInformation("DataSyncWorker 종료");
    }
}

// 등록
builder.Services.AddHostedService<DataSyncWorker>();
```

### Wolverine Scheduled / Recurring Messages

Wolverine 3.x는 내장 크론 스케줄러가 없다. 반복 작업은 `BackgroundService` + `IMessageBus` 조합을 사용한다.

```csharp
// 1. Job 메시지 정의
public record CleanupExpiredSessions;
public record GenerateDailyReport;

// 2. Handler (Wolverine가 자동 라우팅)
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

        logger.LogInformation("Cleaned up {Count} expired sessions", expired);
    }
}

// 3. 반복 실행: BackgroundService + IMessageBus
public class ScheduledCleanupService(IServiceScopeFactory scopeFactory) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken ct)
    {
        // 매일 02:00까지 대기 후 실행하거나, 단순 주기 사용
        using var timer = new PeriodicTimer(TimeSpan.FromHours(24));
        while (await timer.WaitForNextTickAsync(ct))
        {
            using var scope = scopeFactory.CreateScope();
            var bus = scope.ServiceProvider.GetRequiredService<IMessageBus>();
            await bus.PublishAsync(new CleanupExpiredSessions());
        }
    }
}

public class ScheduledReportService(IServiceScopeFactory scopeFactory) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken ct)
    {
        using var timer = new PeriodicTimer(TimeSpan.FromHours(24));
        while (await timer.WaitForNextTickAsync(ct))
        {
            using var scope = scopeFactory.CreateScope();
            var bus = scope.ServiceProvider.GetRequiredService<IMessageBus>();
            await bus.PublishAsync(new GenerateDailyReport());
        }
    }
}

// 4. 등록
builder.Services.AddHostedService<ScheduledCleanupService>();
builder.Services.AddHostedService<ScheduledReportService>();

// 5. 일회성 지연 발행 (Wolverine 내장 지원)
await bus.ScheduleAsync(new SendReminder(orderId), DateTimeOffset.UtcNow.AddHours(24));
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
                logger.LogError(ex, "Failed to process work item {ItemId}", item.Id);
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
            logger.LogInformation("Shutdown requested, finishing current work..."));

        while (!ct.IsCancellationRequested)
        {
            await DoWorkAsync(ct);
            await Task.Delay(TimeSpan.FromSeconds(10), ct);
        }
    }

    public override async Task StopAsync(CancellationToken ct)
    {
        logger.LogInformation("GracefulWorker stopping...");
        await base.StopAsync(ct);
        logger.LogInformation("GracefulWorker stopped");
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
    opts.Policies.UseDurableLocalQueues();
});

// Durable 환경에서 일회성 지연 메시지 (서버 재시작 후에도 보존)
await bus.ScheduleAsync(new CleanupExpiredSessions(), DateTimeOffset.UtcNow.AddHours(2));

// 반복 작업은 BackgroundService로 분리 (위 'Wolverine Scheduled / Recurring Messages' 참조)
builder.Services.AddHostedService<ScheduledCleanupService>();
```
