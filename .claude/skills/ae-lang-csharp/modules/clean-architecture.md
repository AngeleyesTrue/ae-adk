---
module: clean-architecture
version: "1.0.0"
last_updated: "2026-04-02"
category: architecture
---

# Clean Architecture 4-Layer

Domain -> Application -> Infrastructure -> Presentation 의존성 방향 기반 아키텍처.

## Quick Reference

의존성 방향 (안쪽에서 바깥쪽):

```
Domain (핵심, 의존성 없음)
  ← Application (유스케이스, Domain만 참조)
    ← Infrastructure (구현, Application + Domain 참조)
      ← Presentation (UI/API, 모두 참조)
```

프로젝트 참조:

| 프로젝트 | 참조 대상 |
|---------|----------|
| MyApp.Domain | (없음) |
| MyApp.Application | MyApp.Domain |
| MyApp.Infrastructure | MyApp.Application, MyApp.Domain |
| MyApp.Web | MyApp.Infrastructure, MyApp.Application, MyApp.Domain |

---

## Detailed Patterns

### Domain Layer

순수 도메인 모델. 외부 의존성 없음.

```csharp
// MyApp.Domain/Entities/Order.cs
namespace MyApp.Domain.Entities;

public class Order : AggregateRoot
{
    public string CustomerId { get; private set; } = default!;
    public OrderStatus Status { get; private set; }
    public Money TotalAmount { get; private set; } = default!;

    private readonly List<OrderLine> _lines = [];
    public IReadOnlyList<OrderLine> Lines => _lines.AsReadOnly();

    private Order() { } // EF Core

    public static Result<Order> Create(string customerId, List<OrderLineDto> lines)
    {
        Guard.Against.NullOrWhiteSpace(customerId, nameof(customerId));
        Guard.Against.NullOrEmpty(lines, nameof(lines));

        var order = new Order
        {
            Id = Guid.CreateVersion7(),
            CustomerId = customerId,
            Status = OrderStatus.Draft,
            CreatedAt = DateTimeOffset.UtcNow
        };

        foreach (var line in lines)
            order._lines.Add(OrderLine.Create(line));

        order.TotalAmount = order.CalculateTotal();
        return Result<Order>.Success(order);
    }
}

// MyApp.Domain/Interfaces/IOrderRepository.cs
public interface IOrderRepository
{
    Task<Order?> FindByIdAsync(Guid id, CancellationToken ct = default);
    Task AddAsync(Order order, CancellationToken ct = default);
}
```

### Application Layer

유스케이스와 Wolverine 핸들러.

```csharp
// MyApp.Application/Orders/Commands/CreateOrder.cs
namespace MyApp.Application.Orders.Commands;

// Command - POCO record
public record CreateOrder(string CustomerId, List<OrderLineDto> Lines);
public record OrderCreated(Guid OrderId, DateTimeOffset CreatedAt);

// Handler - static POCO
public static class CreateOrderHandler
{
    public static async Task<OrderCreated> Handle(
        CreateOrder command,
        IOrderRepository repo,
        ILogger logger,
        CancellationToken ct)
    {
        var result = Order.Create(command.CustomerId, command.Lines);
        if (!result.IsSuccess)
            throw new DomainException(result.Errors);

        await repo.AddAsync(result.Value, ct);

        logger.Information("Order {OrderId} created for {CustomerId}",
            result.Value.Id, command.CustomerId);

        return new OrderCreated(result.Value.Id, result.Value.CreatedAt);
    }
}
```

### Infrastructure Layer

구현 세부사항. EF Core, 외부 서비스 연동.

```csharp
// MyApp.Infrastructure/Persistence/OrderRepository.cs
namespace MyApp.Infrastructure.Persistence;

public class OrderRepository(AppDbContext db) : IOrderRepository
{
    public async Task<Order?> FindByIdAsync(Guid id, CancellationToken ct)
        => await db.Orders
            .Include(o => o.Lines)
            .FirstOrDefaultAsync(o => o.Id == id, ct);

    public async Task AddAsync(Order order, CancellationToken ct)
    {
        db.Orders.Add(order);
        await db.SaveChangesAsync(ct);
    }
}

// MyApp.Infrastructure/DependencyInjection.cs
public static class DependencyInjection
{
    public static IServiceCollection AddInfrastructure(
        this IServiceCollection services, IConfiguration config)
    {
        services.AddDbContext<AppDbContext>(opts =>
            opts.UseSqlServer(config.GetConnectionString("Default")));

        services.AddScoped<IOrderRepository, OrderRepository>();
        return services;
    }
}
```

### Presentation Layer

API 엔드포인트와 Wolverine 연동.

```csharp
// MyApp.Web/Endpoints/OrderEndpoints.cs
namespace MyApp.Web.Endpoints;

public static class OrderEndpoints
{
    public static void MapOrderEndpoints(this WebApplication app)
    {
        var group = app.MapGroup("/api/orders").WithTags("Orders");

        group.MapPost("/", async (CreateOrder command, IMessageBus bus, CancellationToken ct) =>
        {
            var result = await bus.InvokeAsync<OrderCreated>(command, ct);
            return Results.Created($"/api/orders/{result.OrderId}", result);
        });

        group.MapGet("/{id:guid}", async (Guid id, IMessageBus bus, CancellationToken ct) =>
        {
            var result = await bus.InvokeAsync<OrderDto?>(new GetOrder(id), ct);
            return result is not null ? Results.Ok(result) : Results.NotFound();
        });
    }
}
```

---

## Advanced Topics

### Cross-Cutting Concerns

```csharp
// 각 레이어에서 DI Extension Method 패턴
// Program.cs
var builder = WebApplication.CreateBuilder(args);

builder.Services
    .AddDomain()          // Domain 서비스 (도메인 이벤트 등)
    .AddApplication()     // Application 서비스 (Wolverine, Mapster)
    .AddInfrastructure(builder.Configuration)  // Infrastructure
    .AddPresentation();   // Presentation (Swagger, CORS)

builder.Host.UseWolverine();
```

### Feature Folders vs Layer Folders

```
// Layer Folders (기본 Clean Architecture)
src/MyApp.Application/
├── Commands/
├── Queries/
└── Handlers/

// Feature Folders (복잡한 도메인에 권장)
src/MyApp.Application/
├── Orders/
│   ├── Commands/
│   ├── Queries/
│   └── Handlers/
├── Customers/
│   ├── Commands/
│   └── Queries/
└── Shared/
```

### Shared Kernel

```csharp
// 여러 Bounded Context에서 공유하는 핵심 타입
// MyApp.SharedKernel 프로젝트
namespace MyApp.SharedKernel;

public abstract class AggregateRoot : BaseEntity
{
    private readonly List<object> _domainEvents = [];
    public IReadOnlyList<object> DomainEvents => _domainEvents.AsReadOnly();

    protected void AddDomainEvent(object domainEvent) => _domainEvents.Add(domainEvent);
    public void ClearDomainEvents() => _domainEvents.Clear();
}

public abstract class ValueObject
{
    protected abstract IEnumerable<object> GetEqualityComponents();

    public override bool Equals(object? obj)
    {
        if (obj is not ValueObject other) return false;
        return GetEqualityComponents().SequenceEqual(other.GetEqualityComponents());
    }

    public override int GetHashCode()
        => GetEqualityComponents()
            .Aggregate(0, (hash, component) => HashCode.Combine(hash, component));
}
```
