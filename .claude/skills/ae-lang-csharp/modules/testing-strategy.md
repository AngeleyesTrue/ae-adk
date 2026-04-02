---
module: testing-strategy
version: "1.0.0"
last_updated: "2026-04-02"
category: testing
---

# Testing Strategy

테스트 피라미드, 프레임워크 비교, AAA 패턴, 네이밍 규칙.

## Quick Reference

테스트 피라미드:
- **Unit Tests**: 70% (도메인 로직, 핸들러, Value Objects)
- **Integration Tests**: 25% (DB, 외부 서비스, 메시지 처리)
- **E2E Tests**: 5% (사용자 시나리오, API 엔드투엔드)

커버리지 타겟:
- Line Coverage: 80%+
- Branch Coverage: 75%+

SUT 네이밍: `_sut{ClassName}` (e.g., `_sutOrderService`, `_sutCreateOrderHandler`)

테스트 메서드 네이밍: `MethodName_Scenario_ExpectedBehavior`

```csharp
using AwesomeAssertions;

[Fact]
public void Create_WithValidInput_ReturnsSuccess()
{
    // Arrange
    var customerId = "customer-1";
    var lines = new List<OrderLineDto> { new("SKU-001", 2, 100m) };

    // Act
    var result = Order.Create(customerId, lines);

    // Assert
    result.IsSuccess.Should().BeTrue();
    result.Value.CustomerId.Should().Be(customerId);
    result.Value.Status.Should().Be(OrderStatus.Draft);
}
```

---

## Detailed Patterns

### AAA 패턴 (Arrange-Act-Assert)

```csharp
using AwesomeAssertions;
using NSubstitute;

[Fact]
public async Task Handle_WithValidCommand_CreatesOrderAndReturnsResult()
{
    // Arrange - 테스트 데이터 및 의존성 설정
    var db = Substitute.For<IApplicationDbContext>();
    var logger = Substitute.For<ILogger>();
    var command = new CreateOrder("customer-1", [new OrderLineDto("SKU-001", 2)]);

    // Act - 테스트 대상 실행
    var result = await CreateOrderHandler.Handle(command, db, logger, CancellationToken.None);

    // Assert - 결과 검증
    result.Should().NotBeNull();
    result.OrderId.Should().NotBeEmpty();
    await db.Received(1).SaveChangesAsync(Arg.Any<CancellationToken>());
}
```

### 테스트 프로젝트 구조

```
tests/
├── MyApp.Domain.UnitTests/           # 도메인 모델 단위 테스트
│   ├── Entities/
│   │   └── OrderTests.cs
│   └── ValueObjects/
│       └── MoneyTests.cs
├── MyApp.Application.UnitTests/      # 핸들러 단위 테스트
│   └── Orders/
│       ├── CreateOrderHandlerTests.cs
│       └── GetOrderHandlerTests.cs
├── MyApp.Application.IntegrationTests/ # Wolverine 통합 테스트
│   ├── Fixtures/
│   │   └── IntegrationTestFixture.cs
│   └── Orders/
│       └── OrderWorkflowTests.cs
├── MyApp.Infrastructure.IntegrationTests/ # DB 통합 테스트
│   ├── Fixtures/
│   │   └── DatabaseFixture.cs
│   └── Persistence/
│       └── OrderRepositoryTests.cs
└── MyApp.Architecture.Tests/         # 아키텍처 규칙 테스트
    └── DependencyTests.cs
```

### SUT 네이밍 규칙

```csharp
public class OrderServiceTests
{
    // SUT = System Under Test
    // _sut{ClassName} 형식으로 네이밍
    private readonly OrderService _sutOrderService;
    private readonly IOrderRepository _orderRepo;

    public OrderServiceTests()
    {
        _orderRepo = Substitute.For<IOrderRepository>();
        _sutOrderService = new OrderService(_orderRepo);
    }
}
```

### 네이밍 규칙 예시

```csharp
// 패턴: MethodName_Scenario_ExpectedBehavior
[Fact] public void Create_WithValidInput_ReturnsSuccess() { }
[Fact] public void Create_WithEmptyCustomerId_ReturnsError() { }
[Fact] public void Create_WithNoLines_ReturnsError() { }
[Fact] public void Confirm_WhenDraftStatus_UpdatesToConfirmed() { }
[Fact] public void Confirm_WhenAlreadyShipped_ReturnsError() { }
[Fact] public async Task Handle_WithValidCommand_SavesOrderToDatabase() { }
[Fact] public async Task Handle_WithDuplicateSku_ThrowsDomainException() { }
```

### xUnit vs TUnit 비교

| Feature | xUnit v3 | TUnit |
|---------|----------|-------|
| Test Discovery | Runtime reflection | Compile-time source gen |
| Parallelism | Collection-level | Test-level (fine-grained) |
| Async | Task return | async-first, ValueTask |
| Native AOT | Limited | Full support |
| Maturity | Stable, ecosystem mature | Emerging, growing |
| NuGet Weekly | 2M+ downloads | 50K+ downloads |
| Assertions | Assert.Equal (기본) | Assert.That (기본) |
| Lifecycle | IAsyncLifetime | async lifecycle hooks |
| Recommended | 기존 프로젝트 | 신규 프로젝트 고려 |

```csharp
// xUnit v3 스타일
public class OrderTests
{
    [Fact]
    public void Create_WithValidInput_ReturnsSuccess()
    {
        var result = Order.Create("cust-1", [new OrderLineDto("SKU-001", 2, 100m)]);
        result.IsSuccess.Should().BeTrue();
    }

    [Theory]
    [InlineData("")]
    [InlineData(null)]
    [InlineData("   ")]
    public void Create_WithInvalidCustomerId_ReturnsError(string? customerId)
    {
        var result = Order.Create(customerId!, [new OrderLineDto("SKU-001", 2, 100m)]);
        result.IsSuccess.Should().BeFalse();
    }
}

// TUnit 스타일 (신규 프로젝트 고려)
public class OrderTests
{
    [Test]
    public async Task Create_WithValidInput_ReturnsSuccess()
    {
        var result = Order.Create("cust-1", [new OrderLineDto("SKU-001", 2, 100m)]);
        await Assert.That(result.IsSuccess).IsTrue();
    }

    [Test]
    [Arguments("")]
    [Arguments("   ")]
    public async Task Create_WithInvalidCustomerId_ReturnsError(string customerId)
    {
        var result = Order.Create(customerId, [new OrderLineDto("SKU-001", 2, 100m)]);
        await Assert.That(result.IsSuccess).IsFalse();
    }
}
```

---

## Advanced Topics

### Test Data Builders

```csharp
public class OrderBuilder
{
    private string _customerId = "default-customer";
    private List<OrderLineDto> _lines = [new("SKU-001", 1, 100m)];

    public OrderBuilder WithCustomerId(string id) { _customerId = id; return this; }
    public OrderBuilder WithLines(params OrderLineDto[] lines) { _lines = [..lines]; return this; }
    public OrderBuilder WithNoLines() { _lines = []; return this; }

    public Result<Order> Build() => Order.Create(_customerId, _lines);
    public Order BuildValid() => Build().Value;
}

// 사용
var order = new OrderBuilder()
    .WithCustomerId("vip-customer")
    .WithLines(new("SKU-001", 5, 200m), new("SKU-002", 3, 150m))
    .BuildValid();
```

### Test Categories

```csharp
// Trait으로 카테고리 분류
[Fact]
[Trait("Category", "Unit")]
public void Domain_UnitTest() { }

[Fact]
[Trait("Category", "Integration")]
public async Task Database_IntegrationTest() { }

// 실행: dotnet test --filter "Category=Unit"
// 실행: dotnet test --filter "Category=Integration"
```
