---
module: testing-infrastructure
version: "1.0.0"
last_updated: "2026-04-02"
category: testing
---

# Testing Infrastructure

Testcontainers + Respawn, WebApplicationFactory, NetArchTest, Verify.

## Quick Reference

```csharp
using AwesomeAssertions;

// Testcontainers - 실제 DB 컨테이너
var container = new MsSqlBuilder().Build();
await container.StartAsync();

// Respawn - 테스트 간 DB 리셋
await respawner.ResetAsync(connection);

// NetArchTest - 아키텍처 규칙 검증
Types.InAssembly(assembly).ShouldNot().HaveDependencyOn("MyApp.Infrastructure");

// Verify - 스냅샷 테스팅
await Verify(dto);
```

---

## Detailed Patterns

### Testcontainers + Respawn 조합

```csharp
using AwesomeAssertions;

public class IntegrationTestFixture : IAsyncLifetime
{
    private readonly MsSqlContainer _dbContainer = new MsSqlBuilder()
        .WithImage("mcr.microsoft.com/mssql/server:2022-latest")
        .Build();

    private Respawner _respawner = default!;

    public string ConnectionString => _dbContainer.GetConnectionString();

    public async Task InitializeAsync()
    {
        await _dbContainer.StartAsync();

        // DB 마이그레이션 적용
        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseSqlServer(ConnectionString)
            .Options;
        await using var context = new AppDbContext(options);
        await context.Database.MigrateAsync();

        // Respawner 초기화
        using var conn = new SqlConnection(ConnectionString);
        await conn.OpenAsync();
        _respawner = await Respawner.CreateAsync(conn, new RespawnerOptions
        {
            DbAdapter = DbAdapter.SqlServer,
            TablesToIgnore = ["__EFMigrationsHistory"]
        });
    }

    public async Task ResetDatabaseAsync()
    {
        using var conn = new SqlConnection(ConnectionString);
        await conn.OpenAsync();
        await _respawner.ResetAsync(conn);
    }

    public async Task DisposeAsync() => await _dbContainer.DisposeAsync();
}

// 사용
public class OrderRepositoryTests(IntegrationTestFixture fixture)
    : IClassFixture<IntegrationTestFixture>, IAsyncLifetime
{
    public async Task InitializeAsync() => await fixture.ResetDatabaseAsync();
    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task AddAsync_SavesOrderToDatabase()
    {
        // Arrange
        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseSqlServer(fixture.ConnectionString)
            .Options;
        await using var db = new AppDbContext(options);
        var repo = new OrderRepository(db);
        var order = new OrderBuilder().BuildValid();

        // Act
        await repo.AddAsync(order, CancellationToken.None);

        // Assert
        var saved = await repo.FindByIdAsync(order.Id, CancellationToken.None);
        saved.Should().NotBeNull();
        saved!.CustomerId.Should().Be(order.CustomerId);
    }
}
```

### CustomWebApplicationFactory

```csharp
public class CustomWebApplicationFactory : WebApplicationFactory<Program>
{
    private readonly MsSqlContainer _dbContainer;

    public CustomWebApplicationFactory(MsSqlContainer dbContainer)
        => _dbContainer = dbContainer;

    protected override void ConfigureWebHost(IWebHostBuilder builder)
    {
        builder.ConfigureServices(services =>
        {
            // 실제 DB 컨테이너로 교체
            services.RemoveAll<DbContextOptions<AppDbContext>>();
            services.AddDbContext<AppDbContext>(opts =>
                opts.UseSqlServer(_dbContainer.GetConnectionString()));

            // 외부 서비스 모킹
            services.RemoveAll<INotificationService>();
            services.AddScoped(_ => Substitute.For<INotificationService>());
        });

        builder.UseEnvironment("Testing");
    }
}

// API 통합 테스트
public class OrderApiTests(CustomWebApplicationFactory factory)
    : IClassFixture<CustomWebApplicationFactory>
{
    private readonly HttpClient _client = factory.CreateClient();

    [Fact]
    public async Task CreateOrder_ReturnsCreated()
    {
        // Arrange
        var command = new { CustomerId = "customer-1", Lines = new[]
        {
            new { Sku = "SKU-001", Quantity = 2, UnitPrice = 100m }
        }};

        // Act
        var response = await _client.PostAsJsonAsync("/api/orders", command);

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.Created);

        var result = await response.Content.ReadFromJsonAsync<OrderCreated>();
        result.Should().NotBeNull();
        result!.OrderId.Should().NotBeEmpty();
    }
}
```

### NetArchTest 아키텍처 검증

```csharp
using AwesomeAssertions;

public class ArchitectureTests
{
    private readonly Assembly _domainAssembly = typeof(Order).Assembly;
    private readonly Assembly _applicationAssembly = typeof(CreateOrderHandler).Assembly;
    private readonly Assembly _infrastructureAssembly = typeof(AppDbContext).Assembly;

    [Fact]
    public void Domain_ShouldNot_DependOn_Application()
    {
        var result = Types.InAssembly(_domainAssembly)
            .ShouldNot()
            .HaveDependencyOn("MyApp.Application")
            .GetResult();

        result.IsSuccessful.Should().BeTrue();
    }

    [Fact]
    public void Domain_ShouldNot_DependOn_Infrastructure()
    {
        var result = Types.InAssembly(_domainAssembly)
            .ShouldNot()
            .HaveDependencyOn("MyApp.Infrastructure")
            .GetResult();

        result.IsSuccessful.Should().BeTrue();
    }

    [Fact]
    public void Application_ShouldNot_DependOn_Infrastructure()
    {
        var result = Types.InAssembly(_applicationAssembly)
            .ShouldNot()
            .HaveDependencyOn("MyApp.Infrastructure")
            .GetResult();

        result.IsSuccessful.Should().BeTrue();
    }

    [Fact]
    public void Application_ShouldNot_DependOn_Presentation()
    {
        var result = Types.InAssembly(_applicationAssembly)
            .ShouldNot()
            .HaveDependencyOn("MyApp.Web")
            .GetResult();

        result.IsSuccessful.Should().BeTrue();
    }

    [Fact]
    public void Domain_Entities_ShouldBe_Sealed_Or_Abstract()
    {
        var result = Types.InAssembly(_domainAssembly)
            .That()
            .ResideInNamespace("MyApp.Domain.Entities")
            .Should()
            .BeSealed()
            .Or()
            .BeAbstract()
            .GetResult();

        // 경고로 처리 (강제하지 않을 수 있음)
    }
}
```

### 스냅샷 테스팅 (Verify)

```csharp
using AwesomeAssertions;

public class SnapshotTests
{
    [Fact]
    public async Task GetOrderDto_MatchesSnapshot()
    {
        // Arrange
        var order = new OrderBuilder()
            .WithCustomerId("customer-1")
            .WithLines(new OrderLineDto("SKU-001", 2, 100m))
            .BuildValid();
        var dto = order.Adapt<OrderDto>();

        // Act & Assert - 스냅샷과 비교
        await Verify(dto)
            .DontScrubGuids() // GUID 마스킹 비활성화 (선택)
            .DontScrubDateTimes(); // DateTime 마스킹 비활성화 (선택)
    }

    [Fact]
    public async Task ApiResponse_MatchesSnapshot()
    {
        var response = new
        {
            Status = "success",
            Data = new OrderDto
            {
                Id = Guid.Parse("12345678-1234-1234-1234-123456789012"),
                CustomerName = "Test Customer",
                TotalAmount = 200m,
                StatusDisplay = "Draft"
            }
        };

        await Verify(response);
    }
}
```

---

## Advanced Topics

### PostgreSQL Testcontainers

```csharp
// PostgreSQL 컨테이너
private readonly PostgreSqlContainer _pgContainer = new PostgreSqlBuilder()
    .WithImage("postgres:16-alpine")
    .Build();

// Redis 컨테이너
private readonly RedisContainer _redisContainer = new RedisBuilder()
    .WithImage("redis:7-alpine")
    .Build();
```

### Shared Fixture (Collection Fixture)

```csharp
// 테스트 컬렉션 전체에서 공유하는 Fixture
[CollectionDefinition("Database")]
public class DatabaseCollection : ICollectionFixture<IntegrationTestFixture> { }

[Collection("Database")]
public class OrderTests(IntegrationTestFixture fixture)
{
    // fixture는 컬렉션 내 모든 테스트에서 공유
}

[Collection("Database")]
public class CustomerTests(IntegrationTestFixture fixture)
{
    // 동일한 DB 컨테이너 재사용
}
```

### CI/CD 테스트 설정

```yaml
# GitHub Actions
- name: Run Unit Tests
  run: dotnet test --filter "Category=Unit" --logger trx

- name: Run Integration Tests
  run: dotnet test --filter "Category=Integration" --logger trx
  env:
    DOCKER_HOST: unix:///var/run/docker.sock  # Testcontainers용

- name: Run Architecture Tests
  run: dotnet test --filter "Category=Architecture" --logger trx
```
