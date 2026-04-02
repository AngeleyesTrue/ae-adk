---
module: testing-wolverine
version: "1.0.0"
last_updated: "2026-04-02"
category: testing
---

# Testing Wolverine Handlers

순수 함수 핸들러 단위 테스트, Tracked Sessions 통합 테스트, NSubstitute 기반.

## Quick Reference

```csharp
using AwesomeAssertions;
using NSubstitute;

// 핵심: Wolverine 핸들러는 순수 함수 → 직접 호출로 테스트
var result = await CreateOrderHandler.Handle(command, db, CancellationToken.None);

// Tracked Session으로 전체 메시지 흐름 검증
var session = await Host.InvokeMessageAndWaitAsync(command);

// 외부 전송 스텁
opts.StubAllExternalTransports();
```

---

## Detailed Patterns

### 순수 함수 핸들러 단위 테스트

```csharp
using AwesomeAssertions;
using NSubstitute;

public class CreateOrderHandlerTests
{
    [Fact]
    public async Task CreateOrder_WithValidInput_ReturnsOrderCreatedAndEmail()
    {
        // Arrange
        var db = Substitute.For<IApplicationDbContext>();
        var orders = new List<Order>();
        // ToAsyncDbSet(): DbSet<T> 비동기 모킹용 확장 메서드
        // NuGet: MockQueryable.NSubstitute 패키지 또는 프로젝트 자체 구현 필요
        // See: https://github.com/romantitov/MockQueryable
        db.Orders.Returns(orders.ToAsyncDbSet());

        var command = new CreateOrder("customer-1", [new OrderLineDto("SKU-001", 2)]);

        // Act
        var (created, email) = await CreateOrderHandler.Handle(
            command, db, CancellationToken.None);

        // Assert
        created.Should().NotBeNull();
        created.OrderId.Should().NotBeEmpty();
        email.CustomerId.Should().Be("customer-1");
        await db.Received(1).SaveChangesAsync(Arg.Any<CancellationToken>());
    }

    [Fact]
    public async Task CreateOrder_WithEmptyCustomerId_ThrowsException()
    {
        // Arrange
        var db = Substitute.For<IApplicationDbContext>();
        var command = new CreateOrder("", [new OrderLineDto("SKU-001", 2)]);

        // Act & Assert
        var act = () => CreateOrderHandler.Handle(command, db, CancellationToken.None);
        await act.Should().ThrowAsync<ArgumentException>();
    }
}
```

### Cascading Messages 검증

```csharp
using AwesomeAssertions;

public class ConfirmOrderHandlerTests
{
    [Fact]
    public void ConfirmOrder_WithDraftOrder_ReturnsCascadingMessages()
    {
        // Arrange
        var order = new OrderBuilder().BuildValid();

        // Act - 핸들러 직접 호출
        var (confirmed, shipmentRequested) = ConfirmOrderHandler.Handle(
            new ConfirmOrder(order.Id), order);

        // Assert - Cascading Messages 검증
        confirmed.Should().NotBeNull();
        confirmed.OrderId.Should().Be(order.Id);

        shipmentRequested.Should().NotBeNull();
        shipmentRequested.OrderId.Should().Be(order.Id);
    }

    [Fact]
    public void ConfirmOrder_WithShippedOrder_ThrowsDomainException()
    {
        // Arrange
        var order = new OrderBuilder().BuildValid();
        order.Confirm(); // Draft → Confirmed
        order.Ship("TRACK-001"); // Confirmed → Shipped

        // Act & Assert
        var act = () => ConfirmOrderHandler.Handle(
            new ConfirmOrder(order.Id), order);
        act.Should().Throw<DomainException>();
    }
}
```

### Tracked Sessions 통합 테스트

```csharp
using AwesomeAssertions;

public class OrderWorkflowTests : IAsyncLifetime
{
    private IHost _host = default!;

    public async Task InitializeAsync()
    {
        _host = await Host.CreateDefaultBuilder()
            .UseWolverine(opts =>
            {
                opts.Discovery.IncludeAssembly(typeof(CreateOrderHandler).Assembly);
                opts.StubAllExternalTransports();
            })
            .StartAsync();
    }

    [Fact]
    public async Task CreateOrder_TrackedSession_CompletesFullWorkflow()
    {
        // Arrange
        var command = new CreateOrder("customer-1", [new OrderLineDto("SKU-001", 2)]);

        // Act - Tracked Session으로 모든 cascading 메시지 처리 완료까지 대기
        var session = await _host.InvokeMessageAndWaitAsync(command);

        // Assert - 발행된 메시지 유형 검증
        session.Sent.SingleMessage<OrderCreated>().Should().NotBeNull();
        session.Sent.SingleMessage<OrderConfirmationEmail>()
            .CustomerId.Should().Be("customer-1");
    }

    [Fact]
    public async Task PlaceOrder_TrackedSession_PublishesEventAndEmail()
    {
        // Arrange
        var placeCommand = new PlaceOrder(Guid.CreateVersion7());

        // Act
        var session = await _host.InvokeMessageAndWaitAsync(placeCommand);

        // Assert
        session.Sent.SingleMessage<OrderPlaced>().Should().NotBeNull();
    }

    public async Task DisposeAsync() => await _host.StopAsync();
}
```

### Handler Stubbing

```csharp
using AwesomeAssertions;

[Fact]
public async Task ExternalTransport_IsStubbed_DoesNotSendActualMessages()
{
    // Arrange - 외부 전송 스텁 처리
    using var host = await Host.CreateDefaultBuilder()
        .UseWolverine(opts =>
        {
            opts.Discovery.IncludeAssembly(typeof(CreateOrderHandler).Assembly);
            opts.StubAllExternalTransports(); // 외부 전송 차단
        })
        .StartAsync();

    var command = new CreateOrder("test-customer", [new OrderLineDto("SKU-001", 1)]);

    // Act
    var session = await host.InvokeMessageAndWaitAsync(command);

    // Assert - 메시지가 발행되었지만 실제 전송은 차단됨
    session.Sent.MessagesOf<OrderCreated>().Should().HaveCount(1);
}
```

### NSubstitute 의존성 모킹

```csharp
using NSubstitute;
using AwesomeAssertions;

public class OrderPlacedHandlerTests
{
    [Fact]
    public async Task Handle_SendsNotification()
    {
        // Arrange - NSubstitute로 의존성 모킹
        var notifications = Substitute.For<INotificationService>();
        var logger = Substitute.For<ILogger>();

        notifications.SendOrderConfirmationAsync(
            Arg.Any<Guid>(), Arg.Any<CancellationToken>())
            .Returns(Task.CompletedTask);

        var @event = new OrderPlaced(Guid.CreateVersion7(), "customer-1", 500m);

        // Act
        await OrderPlacedHandler.Handle(@event, notifications, logger, CancellationToken.None);

        // Assert
        await notifications.Received(1).SendOrderConfirmationAsync(
            @event.OrderId, Arg.Any<CancellationToken>());
    }

    [Fact]
    public async Task Handle_WithFailedNotification_LogsError()
    {
        // Arrange
        var notifications = Substitute.For<INotificationService>();
        var logger = Substitute.For<ILogger>();

        notifications.SendOrderConfirmationAsync(
            Arg.Any<Guid>(), Arg.Any<CancellationToken>())
            .ThrowsAsync(new HttpRequestException("Service unavailable"));

        var @event = new OrderPlaced(Guid.CreateVersion7(), "customer-1", 500m);

        // Act & Assert
        var act = () => OrderPlacedHandler.Handle(
            @event, notifications, logger, CancellationToken.None);
        await act.Should().ThrowAsync<HttpRequestException>();
    }
}
```

---

## Advanced Topics

### Testing Middleware

```csharp
using NSubstitute;
using AwesomeAssertions;

[Fact]
public async Task ValidationMiddleware_WithInvalidInput_ReturnsProblemDetails()
{
    // Arrange
    var validator = new CreateOrderValidator();
    var validators = new List<IValidator<CreateOrder>> { validator };
    var invalidCommand = new CreateOrder("", []); // 빈 입력

    // Act
    var result = await FluentValidationMiddleware.BeforeAsync(
        invalidCommand, validators, CancellationToken.None);

    // Assert
    result.Should().NotBeNull();
    result!.Status.Should().Be(400);
}
```

### Custom Test Host Configuration

```csharp
public class WolverineTestHost : IAsyncLifetime
{
    private IHost _host = default!;

    public IHost Host => _host;

    public async Task InitializeAsync()
    {
        _host = await Microsoft.Extensions.Hosting.Host.CreateDefaultBuilder()
            .UseWolverine(opts =>
            {
                opts.Discovery.IncludeAssembly(typeof(CreateOrderHandler).Assembly);
                opts.StubAllExternalTransports();

                // 테스트용 로컬 큐
                opts.LocalQueue("test-queue").Sequential();

                // 인메모리 메시지 저장 (테스트 전용)
                opts.Policies.UseDurableLocalQueues();
            })
            .ConfigureServices(services =>
            {
                // 테스트용 서비스 등록
                services.AddScoped<INotificationService>(_ =>
                    Substitute.For<INotificationService>());
            })
            .StartAsync();
    }

    public async Task DisposeAsync() => await _host.StopAsync();
}
```
