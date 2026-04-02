---
module: package-alternatives
version: "1.0.0"
last_updated: "2026-04-02"
category: reference
---

# Package Alternatives (대체 패키지 대응표)

금지 패키지에서 대체 패키지로의 전환 가이드.

## 전체 대응표

| 금지 패턴 | 대체 패턴 | 비고 |
|---|---|---|
| `IRequest<TResponse>` | Wolverine POCO record | 인터페이스 불필요 |
| `IRequestHandler<TReq, TRes>` | `static Task<TRes> Handle(TReq, ...)` | POCO static handler |
| `IMediator.Send()` | `IMessageBus.InvokeAsync<T>()` | Wolverine bus |
| `INotification` | Wolverine Event (POCO record) | 인터페이스 불필요 |
| `INotificationHandler<T>` | `static Task Handle(TEvent, ...)` | POCO handler |
| `IMapper.Map<T>()` | `source.Adapt<T>()` | Mapster extension |
| `CreateMap<S,D>()` | `IMapFrom<T>.ConfigureMapping()` | Mapster interface |
| `Profile` (AutoMapper) | `IRegister` (Mapster) | 매핑 등록 |
| `ForMember().MapFrom()` | `TypeAdapterConfig` fluent API | Mapster config |
| `using FluentAssertions` | `using AwesomeAssertions` | Drop-in replacement |
| `new Mock<T>()` | `Substitute.For<T>()` | NSubstitute (신규 프로젝트) |
| `mock.Setup(x => x.Method())` | `sub.Method().Returns(value)` | NSubstitute 구문 |
| `mock.Verify(x => x.Method())` | `sub.Received().Method()` | NSubstitute 검증 |
| `UseInMemoryDatabase()` | `Testcontainers.MsSql` / `Testcontainers.PostgreSql` | 실제 DB 컨테이너 |

---

## CQRS: MediatR → Wolverine

### Before (금지)

MediatR 패턴은 인터페이스 기반으로, 모든 Command/Query에 `IRequest<T>` 구현이 필요했다.

### After (Wolverine)

```csharp
// Command - 인터페이스 없는 POCO record
public record CreateOrder(string CustomerId, List<OrderLineDto> Lines);

// Response
public record OrderCreated(Guid OrderId, DateTimeOffset CreatedAt);

// Handler - static POCO handler
public static class CreateOrderHandler
{
    public static async Task<OrderCreated> Handle(
        CreateOrder command,
        AppDbContext db,
        ILogger logger,
        CancellationToken ct)
    {
        Guard.Against.NullOrEmpty(command.CustomerId, nameof(command.CustomerId));

        var order = Order.Create(command.CustomerId, command.Lines);
        db.Orders.Add(order);
        await db.SaveChangesAsync(ct);

        logger.Information("Order {OrderId} created", order.Id);
        return new OrderCreated(order.Id, order.CreatedAt);
    }
}

// 호출
var result = await bus.InvokeAsync<OrderCreated>(new CreateOrder("cust-1", lines), ct);
```

---

## Mapping: AutoMapper → Mapster

### After (Mapster)

```csharp
// DTO with IMapFrom<T>
public class OrderDto : IMapFrom<Order>
{
    public Guid Id { get; init; }
    public string CustomerName { get; init; } = default!;
    public decimal TotalAmount { get; init; }

    public void ConfigureMapping(TypeAdapterConfig config)
    {
        config.NewConfig<Order, OrderDto>()
            .Map(dest => dest.CustomerName, src => src.Customer.FullName)
            .Map(dest => dest.TotalAmount, src => src.Lines.Sum(l => l.Amount));
    }
}

// 사용
var dto = order.Adapt<OrderDto>();
var projected = await db.Orders.ProjectToType<OrderDto>().ToListAsync(ct);
```

---

## Assertions: FluentAssertions → AwesomeAssertions

### After (AwesomeAssertions)

```csharp
using AwesomeAssertions;

[Fact]
public void Order_Create_WithValidInput_ReturnsSuccess()
{
    var result = Order.Create("customer-1", [new OrderLineDto("SKU-001", 2, 100m)]);

    result.IsSuccess.Should().BeTrue();
    result.Value.CustomerId.Should().Be("customer-1");
    result.Value.Status.Should().Be(OrderStatus.Draft);
}
```

---

## Mocking: Moq → NSubstitute

### After (NSubstitute)

```csharp
using NSubstitute;
using AwesomeAssertions;

[Fact]
public async Task Handle_WithValidCommand_SavesOrder()
{
    // Arrange
    var db = Substitute.For<IApplicationDbContext>();
    db.SaveChangesAsync(Arg.Any<CancellationToken>()).Returns(1);

    var command = new CreateOrder("customer-1", [new OrderLineDto("SKU-001", 2)]);

    // Act
    var result = await CreateOrderHandler.Handle(command, db, CancellationToken.None);

    // Assert
    result.Should().NotBeNull();
    await db.Received(1).SaveChangesAsync(Arg.Any<CancellationToken>());
}
```

---

## Integration Test: InMemoryDatabase → Testcontainers

### After (Testcontainers + Respawn)

```csharp
// 실제 SQL Server 컨테이너 사용
private readonly MsSqlContainer _dbContainer = new MsSqlBuilder()
    .WithImage("mcr.microsoft.com/mssql/server:2022-latest")
    .Build();

// Respawn으로 테스트 간 DB 상태 리셋
var respawner = await Respawner.CreateAsync(conn, new RespawnerOptions
{
    DbAdapter = DbAdapter.SqlServer,
    TablesToIgnore = ["__EFMigrationsHistory"]
});
await respawner.ResetAsync(conn);
```
