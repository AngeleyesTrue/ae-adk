---
module: efcore-advanced
version: "1.0.0"
last_updated: "2026-04-02"
category: data
---

# EF Core Advanced

Interceptor, Value Converter, Compiled Query 고급 패턴.

## Quick Reference

```csharp
// SaveChangesInterceptor - 저장 전/후 로직
public class AuditInterceptor : SaveChangesInterceptor { ... }

// Value Converter - 타입 변환
builder.Property(o => o.Status).HasConversion<string>();

// Compiled Query - 성능 최적화
private static readonly Func<AppDbContext, Guid, Task<Order?>> _getById =
    EF.CompileAsyncQuery((AppDbContext db, Guid id) =>
        db.Orders.FirstOrDefault(o => o.Id == id));
```

---

## Detailed Patterns

### AuditInterceptor (SaveChangesInterceptor)

```csharp
public class AuditInterceptor(ICurrentUserService currentUser) : SaveChangesInterceptor
{
    public override ValueTask<InterceptionResult<int>> SavingChangesAsync(
        DbContextEventData eventData,
        InterceptionResult<int> result,
        CancellationToken ct = default)
    {
        var context = eventData.Context;
        if (context is null) return ValueTask.FromResult(result);

        foreach (var entry in context.ChangeTracker.Entries<IAuditable>())
        {
            if (entry.State == EntityState.Added)
            {
                entry.Entity.CreatedAt = DateTimeOffset.UtcNow;
                entry.Entity.CreatedBy = currentUser.UserId;
            }

            if (entry.State == EntityState.Modified)
            {
                entry.Entity.UpdatedAt = DateTimeOffset.UtcNow;
                entry.Entity.UpdatedBy = currentUser.UserId;
            }
        }

        return ValueTask.FromResult(result);
    }
}

// 등록
services.AddDbContext<AppDbContext>((sp, opts) =>
{
    opts.UseSqlServer(connectionString);
    opts.AddInterceptors(
        sp.GetRequiredService<AuditInterceptor>(),
        sp.GetRequiredService<DomainEventInterceptor>(),
        sp.GetRequiredService<SoftDeleteInterceptor>());
});
```

### Value Converter

```csharp
// Strongly Typed ID
public readonly record struct OrderId(Guid Value)
{
    public static OrderId New() => new(Guid.CreateVersion7());
    public static implicit operator Guid(OrderId id) => id.Value;
}

// Configuration
builder.Property(o => o.Id)
    .HasConversion(
        id => id.Value,
        value => new OrderId(value));

// Enum to String
builder.Property(o => o.Status)
    .HasConversion<string>()
    .HasMaxLength(50);

// JSON Column
builder.Property(o => o.Metadata)
    .HasConversion(
        v => JsonSerializer.Serialize(v, JsonSerializerOptions.Default),
        v => JsonSerializer.Deserialize<Dictionary<string, string>>(v,
            JsonSerializerOptions.Default) ?? new())
    .HasColumnType("nvarchar(max)");

// 또는 EF Core 8+ JSON Column 네이티브 지원
builder.OwnsOne(o => o.Address, address =>
{
    address.ToJson();
});
```

### DbCommandInterceptor (쿼리 로깅)

```csharp
public class SlowQueryInterceptor(ILogger logger) : DbCommandInterceptor
{
    public override async ValueTask<DbDataReader> ReaderExecutedAsync(
        DbCommand command,
        CommandExecutedEventData eventData,
        DbDataReader result,
        CancellationToken ct = default)
    {
        if (eventData.Duration.TotalMilliseconds > 200)
        {
            logger.Warning(
                "Slow query ({Duration}ms): {CommandText}",
                eventData.Duration.TotalMilliseconds,
                command.CommandText);
        }

        return result;
    }
}
```

---

## Advanced Topics

### Compiled Queries

```csharp
// Compiled Query - LINQ 표현식 사전 컴파일로 성능 향상
public static class OrderQueries
{
    public static readonly Func<AppDbContext, Guid, Task<Order?>> GetByIdAsync =
        EF.CompileAsyncQuery((AppDbContext db, Guid id) =>
            db.Orders
                .Include(o => o.Lines)
                .FirstOrDefault(o => o.Id == id));

    public static readonly Func<AppDbContext, string, IAsyncEnumerable<Order>>
        GetByCustomerAsync =
            EF.CompileAsyncQuery((AppDbContext db, string customerId) =>
                db.Orders
                    .Where(o => o.CustomerId == customerId)
                    .OrderByDescending(o => o.CreatedAt));
}

// 사용
var order = await OrderQueries.GetByIdAsync(db, orderId);
await foreach (var o in OrderQueries.GetByCustomerAsync(db, customerId))
{
    // 처리
}
```

### Batch Operations (ExecuteUpdate / ExecuteDelete)

```csharp
// 대량 업데이트 - 엔티티 로딩 없이 직접 SQL
await db.Orders
    .Where(o => o.Status == OrderStatus.Draft && o.CreatedAt < cutoff)
    .ExecuteUpdateAsync(s => s
        .SetProperty(o => o.Status, OrderStatus.Expired)
        .SetProperty(o => o.UpdatedAt, DateTimeOffset.UtcNow), ct);

// 대량 삭제
var deletedCount = await db.AuditLogs
    .Where(l => l.CreatedAt < retentionDate)
    .ExecuteDeleteAsync(ct);
```

### Temporal Tables

```csharp
// SQL Server Temporal Table 설정
builder.ToTable("Orders", tb => tb.IsTemporal(ttb =>
{
    ttb.HasPeriodStart("ValidFrom");
    ttb.HasPeriodEnd("ValidTo");
    ttb.UseHistoryTable("OrdersHistory");
}));

// 시점 조회
var orderAtTime = await db.Orders
    .TemporalAsOf(DateTime.UtcNow.AddDays(-7))
    .FirstOrDefaultAsync(o => o.Id == orderId, ct);

// 변경 이력 조회
var history = await db.Orders
    .TemporalAll()
    .Where(o => o.Id == orderId)
    .OrderByDescending(o => EF.Property<DateTime>(o, "ValidFrom"))
    .ToListAsync(ct);
```
