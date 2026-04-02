---
module: domain-events
version: "1.0.0"
last_updated: "2026-04-02"
category: ddd
---

# Domain Events

Publish-after-save 패턴과 Wolverine 연동 도메인 이벤트.

## Quick Reference

```csharp
// 도메인 이벤트 = POCO record (인터페이스 없음)
public record OrderPlaced(Guid OrderId, string CustomerId, decimal TotalAmount);

// AggregateRoot에서 이벤트 등록
order._domainEvents.Add(new OrderPlaced(order.Id, customerId, order.TotalAmount));

// SaveChanges 후 자동 발행 (DomainEventInterceptor)
// Wolverine Event Handler로 처리
public static class OrderPlacedHandler
{
    public static async Task Handle(OrderPlaced @event, INotificationService svc, CancellationToken ct)
        => await svc.SendOrderConfirmationAsync(@event.OrderId, ct);
}
```

---

## Detailed Patterns

### AggregateRoot Base Class

```csharp
public abstract class AggregateRoot : BaseEntity
{
    private readonly List<object> _domainEvents = [];
    public IReadOnlyList<object> DomainEvents => _domainEvents.AsReadOnly();

    protected void AddDomainEvent(object domainEvent) => _domainEvents.Add(domainEvent);
    public void ClearDomainEvents() => _domainEvents.Clear();
}
```

### Domain Event Records

```csharp
// 주문 관련 이벤트
public record OrderPlaced(Guid OrderId, string CustomerId, decimal TotalAmount);
public record OrderConfirmed(Guid OrderId, DateTimeOffset ConfirmedAt);
public record OrderShipped(Guid OrderId, string TrackingNumber);
public record OrderCancelled(Guid OrderId, string Reason);
```

### Entity에서 이벤트 등록

```csharp
public class Order : AggregateRoot
{
    private readonly List<object> _domainEvents = [];
    public IReadOnlyList<object> DomainEvents => _domainEvents.AsReadOnly();

    public static Order Create(string customerId, List<OrderLineDto> lines)
    {
        var order = new Order
        {
            Id = Guid.CreateVersion7(),
            CustomerId = customerId,
            Status = OrderStatus.Placed,
            CreatedAt = DateTimeOffset.UtcNow
        };

        order.AddLines(lines);
        order._domainEvents.Add(new OrderPlaced(order.Id, customerId, order.TotalAmount));

        return order;
    }

    public Result Confirm()
    {
        if (Status != OrderStatus.Placed)
            return Result.Error($"Cannot confirm order in {Status} status");

        Status = OrderStatus.Confirmed;
        _domainEvents.Add(new OrderConfirmed(Id, DateTimeOffset.UtcNow));
        return Result.Success();
    }

    public void ClearDomainEvents() => _domainEvents.Clear();
}
```

### DomainEventInterceptor (Publish-After-Save)

```csharp
// SaveChanges 완료 후 이벤트를 Wolverine으로 발행
public class DomainEventInterceptor(IMessageBus bus) : SaveChangesInterceptor
{
    public override async ValueTask<int> SavedChangesAsync(
        SaveChangesCompletedEventData eventData,
        int result,
        CancellationToken ct = default)
    {
        var context = eventData.Context;
        if (context is null) return result;

        var aggregates = context.ChangeTracker
            .Entries<AggregateRoot>()
            .Where(e => e.Entity.DomainEvents.Count > 0)
            .Select(e => e.Entity)
            .ToList();

        foreach (var aggregate in aggregates)
        {
            foreach (var domainEvent in aggregate.DomainEvents)
            {
                await bus.PublishAsync(domainEvent, ct);
            }
            aggregate.ClearDomainEvents();
        }

        return result;
    }
}
```

### Interceptor 등록

```csharp
// Program.cs 또는 DependencyInjection.cs
services.AddDbContext<AppDbContext>((sp, opts) =>
{
    opts.UseSqlServer(connectionString);
    opts.AddInterceptors(sp.GetRequiredService<DomainEventInterceptor>());
});

services.AddScoped<DomainEventInterceptor>();
```

### Event Handler

```csharp
// Wolverine Event Handler - static POCO
public static class OrderPlacedHandler
{
    public static async Task Handle(
        OrderPlaced @event,
        INotificationService notifications,
        ILogger logger,
        CancellationToken ct)
    {
        logger.Information("Processing OrderPlaced event for {OrderId}", @event.OrderId);
        await notifications.SendOrderConfirmationAsync(@event.OrderId, ct);
    }
}

// 여러 핸들러가 동일 이벤트 구독 가능
public static class OrderPlacedAnalyticsHandler
{
    public static async Task Handle(
        OrderPlaced @event,
        IAnalyticsService analytics,
        CancellationToken ct)
    {
        await analytics.TrackOrderAsync(@event.OrderId, @event.TotalAmount, ct);
    }
}
```

---

## Advanced Topics

### Outbox Pattern (Wolverine 내장)

```csharp
// Wolverine의 Durable Outbox는 메시지 유실 방지
builder.Host.UseWolverine(opts =>
{
    // PostgreSQL 기반 Durable Inbox/Outbox
    opts.PersistMessagesWithPostgresql(connectionString);

    // 또는 SQL Server
    // opts.PersistMessagesWithSqlServer(connectionString);

    opts.Policies.UseDurableInboxOnAllListeners();
    opts.Policies.UseDurableOutboxOnAllSendingEndpoints();
});
```

### Event Versioning

```csharp
// V2 이벤트 - 기존 V1과 호환
public record OrderPlacedV2(
    Guid OrderId,
    string CustomerId,
    decimal TotalAmount,
    DateTimeOffset PlacedAt,    // V2에 추가
    string? CouponCode = null); // V2에 추가, 기본값으로 호환

// Upcaster 패턴 (V1 → V2 변환)
public static class OrderPlacedUpcaster
{
    public static OrderPlacedV2 Handle(OrderPlaced v1)
        => new(v1.OrderId, v1.CustomerId, v1.TotalAmount,
               DateTimeOffset.UtcNow, null);
}
```

### Idempotency

```csharp
// 멱등성 보장 - 이벤트 ID 기반 중복 처리 방지
public record OrderPlaced(Guid EventId, Guid OrderId, string CustomerId, decimal TotalAmount)
{
    public OrderPlaced(Guid orderId, string customerId, decimal totalAmount)
        : this(Guid.CreateVersion7(), orderId, customerId, totalAmount) { }
}

public static class OrderPlacedHandler
{
    public static async Task Handle(
        OrderPlaced @event,
        AppDbContext db,
        ILogger logger,
        CancellationToken ct)
    {
        // 중복 체크
        if (await db.ProcessedEvents.AnyAsync(e => e.EventId == @event.EventId, ct))
        {
            logger.Warning("Duplicate event {EventId} skipped", @event.EventId);
            return;
        }

        // 처리 로직
        await db.ProcessedEvents.AddAsync(new ProcessedEvent(@event.EventId), ct);
        await db.SaveChangesAsync(ct);
    }
}
```
