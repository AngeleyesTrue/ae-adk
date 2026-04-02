---
module: aggregate-patterns
version: "1.0.0"
last_updated: "2026-04-02"
category: ddd
---

# Aggregate Root Patterns

Aggregate 경계, 불변식 강제, 트랜잭션 범위.

## Quick Reference

핵심 규칙:
- Aggregate 외부에서 내부 Entity를 직접 수정하지 않는다
- Aggregate 간 참조는 ID로만 한다 (객체 참조 금지)
- 하나의 트랜잭션에서 하나의 Aggregate만 수정한다
- Repository는 Aggregate Root 단위로 작성한다

```csharp
// Aggregate Root 기본 구조
public class Order : AggregateRoot
{
    private readonly List<OrderLine> _lines = [];
    public IReadOnlyList<OrderLine> Lines => _lines.AsReadOnly();

    // 내부 Entity 수정은 Aggregate Root를 통해서만
    public Result AddLine(string sku, int quantity, decimal unitPrice) { ... }
    public Result RemoveLine(Guid lineId) { ... }
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

    public byte[] RowVersion { get; private set; } = []; // 낙관적 동시성

    protected void AddDomainEvent(object domainEvent)
        => _domainEvents.Add(domainEvent);

    public void ClearDomainEvents() => _domainEvents.Clear();
}

public abstract class BaseEntity : IEquatable<BaseEntity>
{
    public Guid Id { get; protected set; } = Guid.CreateVersion7();
    public DateTimeOffset CreatedAt { get; protected set; } = DateTimeOffset.UtcNow;
    public DateTimeOffset? UpdatedAt { get; protected set; }

    public bool Equals(BaseEntity? other) => other is not null && Id == other.Id;
    public override bool Equals(object? obj) => Equals(obj as BaseEntity);
    public override int GetHashCode() => Id.GetHashCode();
}
```

### Aggregate Boundary (Order + OrderLine)

```csharp
public class Order : AggregateRoot
{
    public string CustomerId { get; private set; } = default!;
    public OrderStatus Status { get; private set; }

    private readonly List<OrderLine> _lines = [];
    public IReadOnlyList<OrderLine> Lines => _lines.AsReadOnly();

    private Order() { }

    // 간소화 버전 - 전체 Factory Method 패턴은 rich-domain-modeling.md 참조
    public static Result<Order> Create(string customerId)
    {
        Guard.Against.NullOrWhiteSpace(customerId, nameof(customerId));

        return Result<Order>.Success(new Order
        {
            CustomerId = customerId,
            Status = OrderStatus.Draft
        });
    }

    // 내부 Entity 조작은 Aggregate Root를 통해서만
    public Result AddLine(string sku, int quantity, decimal unitPrice)
    {
        Guard.Against.NullOrWhiteSpace(sku, nameof(sku));
        Guard.Against.NegativeOrZero(quantity, nameof(quantity));
        Guard.Against.Negative(unitPrice, nameof(unitPrice));

        if (Status != OrderStatus.Draft)
            return Result.Error("Can only add lines to draft orders");

        // 동일 SKU 중복 체크 (불변식)
        if (_lines.Any(l => l.Sku == sku))
            return Result.Error($"SKU {sku} already exists in this order");

        _lines.Add(OrderLine.Create(Id, sku, quantity, unitPrice));

        UpdatedAt = DateTimeOffset.UtcNow;
        return Result.Success();
    }

    public Result RemoveLine(Guid lineId)
    {
        if (Status != OrderStatus.Draft)
            return Result.Error("Can only remove lines from draft orders");

        var line = _lines.FirstOrDefault(l => l.Id == lineId);
        if (line is null)
            return Result.Error("Line item not found");

        _lines.Remove(line);
        UpdatedAt = DateTimeOffset.UtcNow;
        return Result.Success();
    }

    public Result UpdateLineQuantity(Guid lineId, int newQuantity)
    {
        Guard.Against.NegativeOrZero(newQuantity, nameof(newQuantity));

        if (Status != OrderStatus.Draft)
            return Result.Error("Can only modify draft orders");

        var line = _lines.FirstOrDefault(l => l.Id == lineId);
        if (line is null)
            return Result.Error("Line item not found");

        line.UpdateQuantity(newQuantity);
        UpdatedAt = DateTimeOffset.UtcNow;
        return Result.Success();
    }

    public decimal CalculateTotal()
        => _lines.Sum(l => l.UnitPrice * l.Quantity);
}

// 내부 Entity - Aggregate 밖에서 직접 접근 불가
public class OrderLine : BaseEntity
{
    internal OrderLine() { } // EF Core용

    public Guid OrderId { get; private set; }
    public string Sku { get; private set; } = default!;
    public int Quantity { get; private set; }
    public decimal UnitPrice { get; private set; }

    internal static OrderLine Create(Guid orderId, string sku, int quantity, decimal unitPrice)
        => new() { Id = Guid.CreateVersion7(), OrderId = orderId, Sku = sku, Quantity = quantity, UnitPrice = unitPrice };

    internal void UpdateQuantity(int newQuantity) => Quantity = newQuantity;
}
```

### Repository (Aggregate 단위)

```csharp
// Repository는 Aggregate Root 단위로만 작성
public interface IOrderRepository
{
    Task<Order?> FindByIdAsync(Guid id, CancellationToken ct = default);
    Task<Order?> FindByIdWithLinesAsync(Guid id, CancellationToken ct = default);
    Task AddAsync(Order order, CancellationToken ct = default);
    Task SaveChangesAsync(CancellationToken ct = default);
}

public class OrderRepository(AppDbContext db) : IOrderRepository
{
    public async Task<Order?> FindByIdAsync(Guid id, CancellationToken ct)
        => await db.Orders.FindAsync([id], ct);

    public async Task<Order?> FindByIdWithLinesAsync(Guid id, CancellationToken ct)
        => await db.Orders
            .Include(o => o.Lines)
            .FirstOrDefaultAsync(o => o.Id == id, ct);

    public async Task AddAsync(Order order, CancellationToken ct)
        => await db.Orders.AddAsync(order, ct);

    public async Task SaveChangesAsync(CancellationToken ct)
        => await db.SaveChangesAsync(ct);
}
```

### Concurrency Control

```csharp
// EF Core Configuration
public class OrderConfiguration : IEntityTypeConfiguration<Order>
{
    public void Configure(EntityTypeBuilder<Order> builder)
    {
        builder.Property(o => o.RowVersion).IsRowVersion();
    }
}

// 동시성 충돌 처리
public static class UpdateOrderHandler
{
    public static async Task Handle(
        UpdateOrder command,
        IOrderRepository repo,
        ILogger logger,
        CancellationToken ct)
    {
        var order = await repo.FindByIdWithLinesAsync(command.OrderId, ct)
            ?? throw new NotFoundException();

        var result = order.AddLine(command.Sku, command.Quantity, command.UnitPrice);
        if (!result.IsSuccess) throw new DomainException(result.Errors);

        try
        {
            await repo.SaveChangesAsync(ct);
        }
        catch (DbUpdateConcurrencyException)
        {
            logger.LogWarning("Concurrency conflict on Order {OrderId}", command.OrderId);
            throw; // Wolverine 재시도 정책이 처리
        }
    }
}
```

---

## Advanced Topics

### Small Aggregates 원칙

```csharp
// 나쁜 예: 너무 큰 Aggregate (주문 + 고객 + 결제 + 배송)
// 좋은 예: 작은 Aggregate (주문만, 고객ID로 참조)
public class Order : AggregateRoot
{
    // 고객 ID로만 참조 (객체 참조 X)
    public string CustomerId { get; private set; } = default!;

    // 배송 정보는 별도 Aggregate
    public Guid? ShipmentId { get; private set; }
}

public class Shipment : AggregateRoot
{
    // 주문 ID로만 참조
    public Guid OrderId { get; private set; }
    public string TrackingNumber { get; private set; } = default!;
    public ShipmentStatus Status { get; private set; }
}
```

### Eventual Consistency Between Aggregates

```csharp
// Aggregate 간 일관성은 도메인 이벤트로 보장
// Order Aggregate에서 이벤트 발행
public Result Confirm()
{
    Status = OrderStatus.Confirmed;
    AddDomainEvent(new OrderConfirmed(Id, DateTimeOffset.UtcNow));
    return Result.Success();
}

// Shipment Aggregate에서 이벤트 처리
public static class OrderConfirmedHandler
{
    public static async Task Handle(
        OrderConfirmed @event,
        AppDbContext db,
        CancellationToken ct)
    {
        var shipment = Shipment.CreateForOrder(@event.OrderId);
        db.Shipments.Add(shipment);
        await db.SaveChangesAsync(ct);
    }
}
```

---
**관련 모듈**: [Domain Events](domain-events.md) | [EF Core Conventions](efcore-conventions.md) | [Rich Domain Modeling](rich-domain-modeling.md)
