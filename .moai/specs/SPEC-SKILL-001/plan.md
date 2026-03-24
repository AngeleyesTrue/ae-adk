---
spec_id: SPEC-SKILL-001
type: plan
version: "1.0.0"
---

# SPEC-SKILL-001 구현 계획: C# 스킬 고도화

## 1. 개요

### 1.1. 전략: Fork-and-Replace

기존 `moai-lang-csharp` 스킬은 **수정하지 않고 유지**한다. 대신 `ae-lang-csharp` 스킬을 새로 생성하여 동일한 trigger 조건에서 **우선 로드**되도록 구성한다. 이로써:

- moai-adk 업스트림 업데이트와 충돌이 발생하지 않는다.
- ae-adk 고유의 기술 스택이 완전히 반영된다.
- 필요 시 moai-lang-csharp로 롤백이 가능하다.

### 1.2. 대상 경로

```
.claude/skills/ae-lang-csharp/
```

### 1.3. 총 산출물

- SKILL.md: 1개
- modules/: 16개 (마이그레이션 5 + 신규 11)
- references/: 2개
- **합계: 19개 파일**

---

## 2. 구현 단계 (Milestones)

### Phase 1 (Primary): 스킬 골격 + 5개 마이그레이션 모듈

**목표**: ae-lang-csharp 스킬의 기본 구조를 수립하고, spk-dotnet-guide에서 5개 모듈을 마이그레이션한다.

**산출물**:

| 순서 | 파일 | 작업 | 원본 |
|---|---|---|---|
| 1 | `SKILL.md` | 신규 작성 | - |
| 2 | `references/banned-packages.md` | 신규 작성 | - |
| 3 | `references/package-alternatives.md` | 신규 작성 | - |
| 4 | `modules/dotnet-platform.md` | 마이그레이션 | spk-dotnet-guide |
| 5 | `modules/clean-architecture.md` | 마이그레이션 | spk-dotnet-guide |
| 6 | `modules/wolverine-cqrs.md` | 마이그레이션 | spk-dotnet-guide |
| 7 | `modules/efcore-conventions.md` | 마이그레이션 | spk-dotnet-guide |
| 8 | `modules/aspnet-blazor.md` | 마이그레이션 | spk-dotnet-guide |

**마이그레이션 규칙**:

- 기존 Obsidian 콘텐츠의 핵심 내용을 보존한다.
- YAML frontmatter를 추가한다 (module, version, last_updated, category).
- Progressive Disclosure 3단계 구조로 재구성한다.
- 금지 패키지가 포함된 코드 예시가 있으면 대체 패키지로 교체한다.

**완료 기준**: SKILL.md + 7개 파일 = 8개 파일 생성, 금지 패키지 0건.

---

### Phase 2 (Secondary): 6개 Core DDD 신규 모듈

**목표**: DDD/CQRS 핵심 패턴 모듈을 신규 작성한다.

**산출물**:

| 순서 | 파일 | 핵심 내용 |
|---|---|---|
| 1 | `modules/domain-events.md` | Publish-after-save, Wolverine outbox |
| 2 | `modules/wolverine-middleware.md` | Before/After middleware, FluentValidation 연동 |
| 3 | `modules/rich-domain-modeling.md` | Factory Method, Value Object, Entity 패턴 |
| 4 | `modules/mapster-advanced.md` | TypeAdapterConfig, Projection, IMapFrom<T> |
| 5 | `modules/soft-delete-filters.md` | ISoftDeletable, Global Query Filter |
| 6 | `modules/aggregate-patterns.md` | Aggregate Root, 경계, 불변식, 트랜잭션 |

**핵심 코드 예시 - Wolverine POCO Handler**:

```csharp
// Command - 인터페이스 없는 POCO record
public record CreateOrder(string CustomerId, List<OrderLineDto> Lines);

// Response
public record OrderCreated(Guid OrderId, DateTimeOffset CreatedAt);

// Handler - static POCO handler, 인터페이스 구현 불필요
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

        logger.Information("Order {OrderId} created for customer {CustomerId}",
            order.Id, command.CustomerId);

        return new OrderCreated(order.Id, order.CreatedAt);
    }
}
```

**핵심 코드 예시 - Mapster IMapFrom<T>**:

```csharp
// DTO with IMapFrom<T> 인터페이스
public class OrderDto : IMapFrom<Order>
{
    public Guid Id { get; init; }
    public string CustomerName { get; init; } = default!;
    public decimal TotalAmount { get; init; }
    public string StatusDisplay { get; init; } = default!;

    public void ConfigureMapping(TypeAdapterConfig config)
    {
        config.NewConfig<Order, OrderDto>()
            .Map(dest => dest.CustomerName, src => src.Customer.FullName)
            .Map(dest => dest.TotalAmount, src => src.Lines.Sum(l => l.Amount))
            .Map(dest => dest.StatusDisplay, src => src.Status.ToDisplayString());
    }
}

// 사용법 - Mapster extension method
var dto = order.Adapt<OrderDto>();
var dtos = orders.Select(o => o.Adapt<OrderDto>()).ToList();

// EF Core Projection (쿼리 시점 매핑)
var projected = await db.Orders
    .ProjectToType<OrderDto>()
    .ToListAsync(ct);
```

**핵심 코드 예시 - Domain Events (Publish-After-Save)**:

```csharp
// Domain Event - POCO record (인터페이스 없음, Wolverine이 라우팅)
public record OrderPlaced(Guid OrderId, string CustomerId, decimal TotalAmount);

// Entity에서 이벤트 등록
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

    public void ClearDomainEvents() => _domainEvents.Clear();
}

// SaveChanges 후 이벤트 발행 (EF Core Interceptor)
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

// Wolverine Event Handler
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
```

**핵심 코드 예시 - Rich Domain Model (Factory Method)**:

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

    protected override IEnumerable<object> GetEqualityComponents()
    {
        yield return Amount;
        yield return Currency;
    }
}

public class Order : AggregateRoot
{
    public OrderStatus Status { get; private set; }
    public Money TotalAmount { get; private set; } = default!;

    private Order() { } // EF Core용

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
}
```

**완료 기준**: 6개 파일 생성, 각 모듈에 실행 가능한 코드 예시 포함, 금지 패키지 0건.

---

### Phase 3 (Tertiary): 5개 Infrastructure 신규 모듈

**목표**: 인프라 및 운영 관련 모듈을 신규 작성한다.

**산출물**:

| 순서 | 파일 | 핵심 내용 |
|---|---|---|
| 1 | `modules/service-abstractions.md` | DI 패턴, Options pattern, Factory 패턴 |
| 2 | `modules/efcore-advanced.md` | Interceptor, Value Converter, Compiled Query |
| 3 | `modules/project-templates.md` | 솔루션 구조, Layer별 프로젝트 구성 |
| 4 | `modules/graph-api-integration.md` | Microsoft Graph SDK, Auth, Batch 요청 |
| 5 | `modules/background-workers.md` | IHostedService, Wolverine Scheduled Jobs |

**핵심 코드 예시 - Background Worker (Wolverine Scheduled)**:

```csharp
// Wolverine Scheduled Job
public record CleanupExpiredSessions;

public static class CleanupExpiredSessionsHandler
{
    // Wolverine이 스케줄에 따라 자동 호출
    public static async Task Handle(
        CleanupExpiredSessions command,
        AppDbContext db,
        ILogger logger,
        CancellationToken ct)
    {
        var cutoff = DateTimeOffset.UtcNow.AddHours(-24);

        var expired = await db.Sessions
            .Where(s => s.LastActivity < cutoff)
            .ExecuteDeleteAsync(ct);

        logger.Information("Cleaned up {Count} expired sessions", expired);
    }
}

// Program.cs에서 스케줄 등록
builder.Host.UseWolverine(opts =>
{
    opts.Publish(pub =>
    {
        pub.Message<CleanupExpiredSessions>()
            .ScheduledAt(new CronExpression("0 0 2 * * ?"));  // 매일 02:00
    });
});
```

**핵심 코드 예시 - Graph API Integration**:

```csharp
// Graph API 서비스 추상화
public interface IGraphUserService
{
    Task<Result<GraphUser>> GetUserAsync(string userId, CancellationToken ct = default);
    Task<Result<IReadOnlyList<GraphUser>>> GetGroupMembersAsync(string groupId, CancellationToken ct = default);
}

public class GraphUserService(GraphServiceClient graphClient, ILogger logger)
    : IGraphUserService
{
    public async Task<Result<GraphUser>> GetUserAsync(string userId, CancellationToken ct)
    {
        Guard.Against.NullOrWhiteSpace(userId, nameof(userId));

        try
        {
            var user = await graphClient.Users[userId]
                .GetAsync(config =>
                {
                    config.QueryParameters.Select =
                        ["id", "displayName", "mail", "userPrincipalName"];
                }, ct);

            if (user is null)
                return Result<GraphUser>.NotFound($"User {userId} not found");

            return Result<GraphUser>.Success(user.Adapt<GraphUser>());
        }
        catch (ServiceException ex) when (ex.ResponseStatusCode == 404)
        {
            logger.Warning("Graph user {UserId} not found", userId);
            return Result<GraphUser>.NotFound($"User {userId} not found");
        }
    }
}
```

**완료 기준**: 5개 파일 생성, 실전 코드 예시 포함, 금지 패키지 0건.

---

### Phase 4 (Final): 품질 검증

**목표**: 전체 스킬의 품질을 검증하고 최종 확인한다.

**검증 항목**:

| 순서 | 검증 | 방법 |
|---|---|---|
| 1 | 금지 패키지 Grep | `Grep` 도구로 MediatR, AutoMapper, FluentAssertions 검색 (0건 확인) |
| 2 | 파일 수 확인 | `Glob` 도구로 19개 파일 존재 확인 |
| 3 | YAML frontmatter | `Read` 도구로 각 모듈의 frontmatter 확인 |
| 4 | Progressive Disclosure | `Grep` 도구로 3단계 구조 헤딩 확인 |
| 5 | Context7 매핑 | SKILL.md metadata 확인 |
| 6 | `references/` 디렉토리 | banned-packages.md, package-alternatives.md 존재 및 내용 확인 |

**완료 기준**: 모든 검증 항목 통과.

---

## 3. 기술 접근법

### 3.1. 실전 기반 원칙

모든 코드 예시는 실제 프로젝트(SPK Cloud Archive, SPK ID Manager)에서 사용되는 패턴을 기반으로 한다. 이론적 설명보다 **복사-붙여넣기 가능한 코드**를 우선한다.

### 3.2. 금지 패키지 제로 톨러런스

모든 코드 블록에서 금지 패키지의 API가 단 하나도 포함되어서는 안 된다. 코드 예시의 주석에서도 금지 패키지 이름을 언급할 때는 "사용하지 않음" 맥락에서만 허용한다.

### 3.3. Progressive Disclosure 3단계

모든 모듈은 다음 3단계 구조를 따른다:

1. **Quick Reference**: 핵심 패턴을 즉시 참조할 수 있는 짧은 코드 스니펫과 규칙.
2. **Detailed Patterns**: 패턴의 상세 설명, 사용 시나리오, 전체 코드 예시.
3. **Advanced Topics**: 고급 사용법, 엣지 케이스, 성능 최적화, 실전 팁.

### 3.4. Context7 호환

SKILL.md의 메타데이터 구조는 Context7 MCP 서버가 인식할 수 있는 형식을 유지한다. 각 모듈의 YAML frontmatter에 `module`, `version`, `category` 필드를 포함한다.

### 3.5. 코드 예시 원칙

- **.NET 10 Preview / C# 13+ 문법** 사용 (primary constructor, collection expression, etc.)
- **Guard.Against** (Ardalis.GuardClauses) 사용
- **Serilog** structured logging 사용 (`ILogger` with message template)
- **Result\<T\> pattern** 사용 (exception 대신)
- **async/await + CancellationToken** 항상 포함
- **`Guid.CreateVersion7()`** 사용 (sequential GUID)

---

## 4. 리스크

| 리스크 | 영향 | 완화 방안 |
|---|---|---|
| Wolverine 3.x API 변경 | 코드 예시 무효화 | Wolverine 공식 문서 기반 작성, 버전 명시 |
| Mapster 문서 부족 | 고급 기능 구현 어려움 | GitHub 소스 코드 참조, 실전 프로젝트 경험 반영 |
| .NET 10 Preview 상태 | Breaking changes 가능성 | .NET 9 기반으로 작성 후 출시 시 업그레이드 |
| 16개 모듈 범위 | 일관성 유지 어려움 | Phase별 점진적 작성, 각 Phase 완료 시 검증 |
| AwesomeAssertions API 차이 | Drop-in이 아닌 부분 존재 | 핵심 assertion만 사용, 검증 후 문서화 |

---

## 5. 모듈 의존성 그래프

```
dotnet-platform (기반)
  ├── clean-architecture
  │     ├── domain-events
  │     ├── rich-domain-modeling
  │     │     ├── aggregate-patterns
  │     │     └── soft-delete-filters
  │     ├── service-abstractions
  │     └── project-templates
  ├── wolverine-cqrs
  │     ├── wolverine-middleware
  │     ├── domain-events
  │     └── background-workers
  ├── efcore-conventions
  │     ├── efcore-advanced
  │     └── soft-delete-filters
  ├── aspnet-blazor
  ├── mapster-advanced
  └── graph-api-integration
```

---

## 6. 추적성

| Phase | 요구사항 | 산출물 | 수락 기준 |
|---|---|---|---|
| Phase 1 | REQ-CORE-001, REQ-BAN-001~003, REQ-MIG-001~002 | SKILL.md, references/2, modules/5 | AC-STRUCT-001~003, AC-BAN-001~003, AC-MIG-001~002 |
| Phase 2 | REQ-NEW-001~006, REQ-CORE-002~004 | modules/6 | AC-CORE-001~005, AC-NEW-001 |
| Phase 3 | REQ-NEW-007~011 | modules/5 | AC-NEW-002, AC-INFRA-001~004 |
| Phase 4 | REQ-QUAL-001~003, REQ-COMPAT-001~003 | (검증 결과) | AC-CTX7-001, AC-ALT-001 |
