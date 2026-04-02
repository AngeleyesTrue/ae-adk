---
module: wolverine-cqrs
version: "1.0.0"
last_updated: "2026-04-02"
category: cqrs
---

# Wolverine CQRS (Command/Query/Event)

Wolverine 3.x POCO handler 기반 CQRS 패턴. 인터페이스 없는 순수 함수 핸들러.

## Quick Reference

핵심 원칙:
- Command/Query/Event = **POCO record** (인터페이스 구현 불필요)
- Handler = **static 메서드** (클래스 인스턴스 불필요)
- DI = **메서드 파라미터 주입** (생성자 주입 불필요)
- 호출 = `IMessageBus.InvokeAsync<T>()` (Command/Query), `PublishAsync()` (Event)

```csharp
// Command 정의 - POCO record
public record CreateOrder(string CustomerId, List<OrderLineDto> Lines);
public record OrderCreated(Guid OrderId, DateTimeOffset CreatedAt);

// Handler - static method, 파라미터로 DI 주입
// 참고: 간결성을 위해 AppDbContext를 직접 사용합니다.
// Clean Architecture에서는 IOrderRepository 추상화를 권장합니다.
// 상세 패턴: clean-architecture.md 참조.
public static class CreateOrderHandler
{
    public static async Task<OrderCreated> Handle(
        CreateOrder command,
        AppDbContext db,
        ILogger logger,
        CancellationToken ct)
    {
        Guard.Against.NullOrEmpty(command.CustomerId, nameof(command.CustomerId));
        Guard.Against.NullOrEmpty(command.Lines, nameof(command.Lines));

        var order = Order.Create(command.CustomerId, command.Lines);

        db.Orders.Add(order);
        await db.SaveChangesAsync(ct);

        logger.LogInformation("Order {OrderId} created for customer {CustomerId}",
            order.Id, command.CustomerId);

        return new OrderCreated(order.Id, order.CreatedAt);
    }
}

// 호출
var result = await bus.InvokeAsync<OrderCreated>(
    new CreateOrder("cust-1", lines), ct);
```

---

## Detailed Patterns

### Command Handler

상태를 변경하는 명령. 응답을 반환하거나 void일 수 있다.

```csharp
// Command
public record UpdateOrderStatus(Guid OrderId, OrderStatus NewStatus);

// Handler (void 반환)
public static class UpdateOrderStatusHandler
{
    public static async Task Handle(
        UpdateOrderStatus command,
        AppDbContext db,
        ILogger logger,
        CancellationToken ct)
    {
        var order = await db.Orders.FindAsync([command.OrderId], ct)
            ?? throw new NotFoundException($"Order {command.OrderId} not found");

        var result = order.UpdateStatus(command.NewStatus);
        if (!result.IsSuccess)
            throw new DomainException(result.Errors);

        await db.SaveChangesAsync(ct);
        logger.LogInformation("Order {OrderId} status updated to {Status}",
            command.OrderId, command.NewStatus);
    }
}
```

### Query Handler

읽기 전용 조회. 항상 응답을 반환한다.

```csharp
// Query
public record GetOrder(Guid OrderId);

// Query Handler
public static class GetOrderHandler
{
    public static async Task<OrderDto?> Handle(
        GetOrder query,
        AppDbContext db,
        CancellationToken ct)
    {
        return await db.Orders
            .Where(o => o.Id == query.OrderId)
            .ProjectToType<OrderDto>()
            .FirstOrDefaultAsync(ct);
    }
}

// 페이지네이션 쿼리
public record GetOrders(int Page = 1, int PageSize = 20, string? Status = null);

public static class GetOrdersHandler
{
    public static async Task<PagedResult<OrderDto>> Handle(
        GetOrders query,
        AppDbContext db,
        CancellationToken ct)
    {
        var queryable = db.Orders.AsQueryable();

        if (query.Status is not null)
            queryable = queryable.Where(o => o.Status.ToString() == query.Status);

        var total = await queryable.CountAsync(ct);
        var items = await queryable
            .OrderByDescending(o => o.CreatedAt)
            .Skip((query.Page - 1) * query.PageSize)
            .Take(query.PageSize)
            .ProjectToType<OrderDto>()
            .ToListAsync(ct);

        return new PagedResult<OrderDto>(items, total, query.Page, query.PageSize);
    }
}
```

### Event Handler

도메인 이벤트 처리. 반환값 없음.

```csharp
// Event - POCO record
public record OrderPlaced(Guid OrderId, string CustomerId, decimal TotalAmount);

// Event Handler
public static class OrderPlacedHandler
{
    public static async Task Handle(
        OrderPlaced @event,
        INotificationService notifications,
        ILogger logger,
        CancellationToken ct)
    {
        logger.LogInformation("Processing OrderPlaced for {OrderId}", @event.OrderId);
        await notifications.SendOrderConfirmationAsync(@event.OrderId, ct);
    }
}

// 이벤트 발행
await bus.PublishAsync(new OrderPlaced(order.Id, customerId, totalAmount), ct);
```

### Wolverine 등록

```csharp
// Program.cs
var builder = WebApplication.CreateBuilder(args);

builder.Host.UseWolverine(opts =>
{
    // 핸들러 자동 디스커버리 (컨벤션 기반)
    opts.Discovery.IncludeAssembly(typeof(CreateOrderHandler).Assembly);

    // 로컬 큐 설정
    opts.LocalQueue("orders")
        .Sequential()
        .UseDurableInbox();

    // 외부 전송 스텁 (테스트용)
    // opts.StubAllExternalTransports();
});
```

---

## Advanced Topics

### Cascading Messages

핸들러 반환값으로 후속 메시지를 자동 발행한다.

```csharp
// 튜플 반환 = Cascading Messages
public static class PlaceOrderHandler
{
    public static async Task<(OrderPlaced, OrderConfirmationEmail)> Handle(
        PlaceOrder command,
        AppDbContext db,
        CancellationToken ct)
    {
        var order = await db.Orders.FindAsync([command.OrderId], ct)
            ?? throw new NotFoundException();

        order.Place();
        await db.SaveChangesAsync(ct);

        // 반환된 두 메시지가 자동으로 발행됨
        return (
            new OrderPlaced(order.Id, order.CustomerId, order.TotalAmount),
            new OrderConfirmationEmail(order.CustomerId, order.Id)
        );
    }
}
```

### Retry Policies

```csharp
builder.Host.UseWolverine(opts =>
{
    // 특정 핸들러에 재시도 정책 적용
    opts.Policies.ForMessagesOfType<CreateOrder>()
        .RetryWithCooldown(
            TimeSpan.FromMilliseconds(100),
            TimeSpan.FromMilliseconds(500),
            TimeSpan.FromSeconds(1));

    // 전역 에러 핸들링
    opts.Policies.OnException<DbUpdateConcurrencyException>()
        .RetryTimes(3);

    opts.Policies.OnException<TimeoutException>()
        .Requeue();
});
```

### Dead Letter Queue

```csharp
builder.Host.UseWolverine(opts =>
{
    // 처리 실패 메시지를 Dead Letter Queue로 이동
    opts.Policies.OnException<Exception>()
        .MoveToErrorQueue();

    // Dead Letter Queue 모니터링
    opts.LocalQueue("dead-letters")
        .UseDurableInbox();
});
```

---
**관련 모듈**: [Wolverine Middleware](wolverine-middleware.md) | [Clean Architecture](clean-architecture.md) | [Domain Events](domain-events.md)
