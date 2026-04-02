---
module: rich-domain-modeling
version: "1.0.0"
last_updated: "2026-04-02"
category: ddd
---

# Rich Domain Modeling

Factory Method, Value Object, Entity 패턴. Result&lt;T&gt; 기반 도메인 로직.

## Quick Reference

핵심 원칙:
- **Factory Method**: `static Create()` - 불변식을 보장하는 유일한 생성 경로
- **Value Object**: 불변, 동등성 비교, `GetEqualityComponents()`
- **Entity**: private setter, 상태 전이 메서드, 비즈니스 규칙 캡슐화
- **Result&lt;T&gt;**: 예외 대신 명시적 결과 반환
- **Guard.Against**: 입력 검증에 Ardalis.GuardClauses 사용

---

## Detailed Patterns

### Value Object

```csharp
public class Money : ValueObject
{
    public decimal Amount { get; }
    public string Currency { get; }

    private Money(decimal amount, string currency)
    {
        Amount = amount;
        Currency = currency;
    }

    public static Money Create(decimal amount, string currency)
    {
        Guard.Against.Negative(amount, nameof(amount));
        Guard.Against.NullOrWhiteSpace(currency, nameof(currency));
        Guard.Against.InvalidInput(currency, nameof(currency),
            c => c.Length == 3, "Currency must be ISO 4217 code");

        return new Money(amount, currency.ToUpperInvariant());
    }

    public Money Add(Money other)
    {
        Guard.Against.Expression(
            m => m.Currency != Currency, other,
            "Cannot add different currencies");

        return new Money(Amount + other.Amount, Currency);
    }

    public Money Subtract(Money other)
    {
        Guard.Against.Expression(
            m => m.Currency != Currency, other,
            "Cannot subtract different currencies");

        var result = Amount - other.Amount;
        Guard.Against.Negative(result, message: "Insufficient funds");

        return new Money(result, Currency);
    }

    protected override IEnumerable<object> GetEqualityComponents()
    {
        yield return Amount;
        yield return Currency;
    }
}
```

### Email Value Object

```csharp
public class Email : ValueObject
{
    public string Value { get; }

    private Email(string value) => Value = value;

    public static Result<Email> Create(string email)
    {
        if (string.IsNullOrWhiteSpace(email))
            return Result<Email>.Error("Email is required");

        email = email.Trim().ToLowerInvariant();

        if (!email.Contains('@') || email.Length > 256)
            return Result<Email>.Error("Invalid email format");

        return Result<Email>.Success(new Email(email));
    }

    protected override IEnumerable<object> GetEqualityComponents()
    {
        yield return Value;
    }

    public override string ToString() => Value;
    public static implicit operator string(Email email) => email.Value;
}
```

### Entity with Factory Method

```csharp
public class Order : AggregateRoot
{
    public string CustomerId { get; private set; } = default!;
    public OrderStatus Status { get; private set; }
    public Money TotalAmount { get; private set; } = default!;

    private readonly List<OrderLine> _lines = [];
    public IReadOnlyList<OrderLine> Lines => _lines.AsReadOnly();

    private readonly List<object> _domainEvents = [];
    public IReadOnlyList<object> DomainEvents => _domainEvents.AsReadOnly();

    private Order() { } // EF Core용 private 생성자

    // Factory Method - 불변식을 보장하는 유일한 생성 경로
    public static Result<Order> Create(string customerId, List<OrderLineDto> lines)
    {
        if (string.IsNullOrWhiteSpace(customerId))
            return Result<Order>.Error("Customer ID is required");

        if (lines is not { Count: > 0 })
            return Result<Order>.Error("At least one line item is required");

        var order = new Order
        {
            Id = Guid.CreateVersion7(),
            CustomerId = customerId,
            Status = OrderStatus.Draft,
            CreatedAt = DateTimeOffset.UtcNow
        };

        foreach (var line in lines)
        {
            var addResult = order.AddLine(line);
            if (!addResult.IsSuccess)
                return Result<Order>.Error(addResult.Errors.ToArray());
        }

        order.TotalAmount = order.CalculateTotal();
        return Result<Order>.Success(order);
    }

    // 상태 전이 - 비즈니스 규칙을 도메인에 캡슐화
    public Result Confirm()
    {
        if (Status != OrderStatus.Draft)
            return Result.Error($"Cannot confirm order in {Status} status");

        Status = OrderStatus.Confirmed;
        _domainEvents.Add(new OrderConfirmed(Id, DateTimeOffset.UtcNow));
        return Result.Success();
    }

    public Result Ship(string trackingNumber)
    {
        Guard.Against.NullOrWhiteSpace(trackingNumber, nameof(trackingNumber));

        if (Status != OrderStatus.Confirmed)
            return Result.Error($"Cannot ship order in {Status} status");

        Status = OrderStatus.Shipped;
        _domainEvents.Add(new OrderShipped(Id, trackingNumber));
        return Result.Success();
    }

    public Result Cancel(string reason)
    {
        if (Status is OrderStatus.Shipped or OrderStatus.Delivered)
            return Result.Error("Cannot cancel shipped or delivered order");

        Status = OrderStatus.Cancelled;
        _domainEvents.Add(new OrderCancelled(Id, reason));
        return Result.Success();
    }

    private Result AddLine(OrderLineDto dto)
    {
        if (dto.Quantity <= 0)
            return Result.Error("Quantity must be positive");

        _lines.Add(new OrderLine
        {
            Id = Guid.CreateVersion7(),
            Sku = dto.Sku,
            Quantity = dto.Quantity,
            UnitPrice = Money.Create(dto.UnitPrice, "KRW")
        });

        return Result.Success();
    }

    private Money CalculateTotal()
        => _lines.Aggregate(
            Money.Create(0, "KRW"),
            (sum, line) => sum.Add(
                Money.Create(line.UnitPrice.Amount * line.Quantity, "KRW")));

    public void ClearDomainEvents() => _domainEvents.Clear();
}
```

---

## Advanced Topics

### Specification Pattern

```csharp
// 재사용 가능한 비즈니스 규칙 캡슐화
public abstract class Specification<T>
{
    public abstract Expression<Func<T, bool>> ToExpression();

    public bool IsSatisfiedBy(T entity)
        => ToExpression().Compile()(entity);
}

public class ActiveOrderSpec : Specification<Order>
{
    public override Expression<Func<Order, bool>> ToExpression()
        => order => order.Status != OrderStatus.Cancelled
                  && order.Status != OrderStatus.Delivered;
}

public class HighValueOrderSpec : Specification<Order>
{
    private readonly decimal _threshold;
    public HighValueOrderSpec(decimal threshold) => _threshold = threshold;

    public override Expression<Func<Order, bool>> ToExpression()
        => order => order.TotalAmount.Amount >= _threshold;
}

// 사용
var spec = new ActiveOrderSpec();
var activeOrders = await db.Orders.Where(spec.ToExpression()).ToListAsync(ct);
```

### Domain Service

```csharp
// 여러 Aggregate에 걸친 도메인 로직
public class OrderPricingService
{
    public static Result<Money> CalculateDiscount(
        Order order,
        Customer customer,
        IReadOnlyList<Promotion> activePromotions)
    {
        var baseAmount = order.TotalAmount;
        var discount = Money.Create(0, baseAmount.Currency);

        // 고객 등급 할인
        var tierDiscount = customer.Tier switch
        {
            CustomerTier.Gold => 0.10m,
            CustomerTier.Silver => 0.05m,
            _ => 0m
        };

        if (tierDiscount > 0)
            discount = discount.Add(
                Money.Create(baseAmount.Amount * tierDiscount, baseAmount.Currency));

        // 프로모션 할인
        foreach (var promo in activePromotions.Where(p => p.IsApplicable(order)))
        {
            discount = discount.Add(promo.CalculateDiscount(baseAmount));
        }

        // 최대 할인 한도
        var maxDiscount = Money.Create(baseAmount.Amount * 0.30m, baseAmount.Currency);
        if (discount.Amount > maxDiscount.Amount)
            discount = maxDiscount;

        return Result<Money>.Success(discount);
    }
}
```

### Entity Equality

```csharp
public abstract class BaseEntity : IEquatable<BaseEntity>
{
    public Guid Id { get; protected set; } = Guid.CreateVersion7();
    public DateTimeOffset CreatedAt { get; protected set; } = DateTimeOffset.UtcNow;

    public bool Equals(BaseEntity? other)
    {
        if (other is null) return false;
        if (ReferenceEquals(this, other)) return true;
        return Id == other.Id;
    }

    public override bool Equals(object? obj) => Equals(obj as BaseEntity);
    public override int GetHashCode() => Id.GetHashCode();

    public static bool operator ==(BaseEntity? left, BaseEntity? right)
        => Equals(left, right);
    public static bool operator !=(BaseEntity? left, BaseEntity? right)
        => !Equals(left, right);
}
```
