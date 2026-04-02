---
module: mapster-advanced
version: "1.0.0"
last_updated: "2026-04-02"
category: infrastructure
---

# Mapster Advanced Mapping

Mapster 7.x IRegister, TypeAdapterConfig, EF Core Projection.

## Quick Reference

```csharp
// 단순 매핑
var dto = order.Adapt<OrderDto>();

// IRegister 인터페이스로 매핑 등록
public class OrderMappingRegister : IRegister { ... }

// EF Core Projection
var projected = await db.Orders.ProjectToType<OrderDto>().ToListAsync(ct);

// TypeAdapterConfig
config.NewConfig<Order, OrderDto>()
    .Map(dest => dest.CustomerName, src => src.Customer.FullName);
```

---

## Detailed Patterns

### IRegister Interface

```csharp
// IRegister 인터페이스로 매핑 설정을 별도 클래스에 정의
public class OrderMappingRegister : IRegister
{
    public void Register(TypeAdapterConfig config)
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
```

### TypeAdapterConfig Fluent API

```csharp
// 복잡한 매핑 설정
public static class MappingConfig
{
    public static void Configure()
    {
        TypeAdapterConfig<Order, OrderSummaryDto>.NewConfig()
            .Map(dest => dest.ItemCount, src => src.Lines.Count)
            .Map(dest => dest.TotalAmount, src => src.Lines.Sum(l => l.UnitPrice * l.Quantity))
            .Map(dest => dest.IsHighValue, src => src.Lines.Sum(l => l.UnitPrice * l.Quantity) > 1000)
            .Ignore(dest => dest.InternalNotes);

        // 조건부 매핑
        TypeAdapterConfig<Customer, CustomerDto>.NewConfig()
            .Map(dest => dest.Email, src => src.Email.Value)
            .Map(dest => dest.FullName,
                src => $"{src.FirstName} {src.LastName}".Trim())
            .Map(dest => dest.MemberSince,
                src => src.CreatedAt.ToString("yyyy-MM-dd"));
    }
}
```

### EF Core Projection

```csharp
// 쿼리 시점에서 매핑 - SELECT 절 최적화
public static class GetOrdersHandler
{
    public static async Task<PagedResult<OrderDto>> Handle(
        GetOrders query,
        AppDbContext db,
        CancellationToken ct)
    {
        var queryable = db.Orders.AsQueryable();

        // ProjectToType은 SQL SELECT 절을 최적화
        // 필요한 컬럼만 조회 (N+1 방지)
        var items = await queryable
            .OrderByDescending(o => o.CreatedAt)
            .Skip((query.Page - 1) * query.PageSize)
            .Take(query.PageSize)
            .ProjectToType<OrderDto>()
            .ToListAsync(ct);

        var total = await queryable.CountAsync(ct);
        return new PagedResult<OrderDto>(items, total, query.Page, query.PageSize);
    }
}
```

### Collection Mapping

```csharp
// 리스트 매핑
var dtos = orders.Adapt<List<OrderDto>>();

// 배열 매핑
var array = orders.Adapt<OrderDto[]>();

// Dictionary 매핑
var lookup = orders.Adapt<Dictionary<Guid, OrderDto>>();
```

### Nested Object Mapping

```csharp
// IRegister로 중첩 객체 매핑 설정
public class OrderDetailMappingRegister : IRegister
{
    public void Register(TypeAdapterConfig config)
    {
        config.NewConfig<Order, OrderDetailDto>()
            .Map(dest => dest.Customer, src => src.Customer)
            .Map(dest => dest.Lines, src => src.Lines)
            .Map(dest => dest.ShippingAddress, src => src.ShippingAddress);

        // 중첩 객체도 별도 매핑 설정
        config.NewConfig<OrderLine, OrderLineDto>()
            .Map(dest => dest.TotalPrice,
                src => src.UnitPrice * src.Quantity);
    }
}
```

### DI 등록

```csharp
// Program.cs 또는 DependencyInjection.cs
public static IServiceCollection AddMapsterConfig(this IServiceCollection services)
{
    var config = TypeAdapterConfig.GlobalSettings;

    // IRegister 인터페이스를 구현한 모든 타입 자동 스캔
    config.Scan(typeof(OrderMappingRegister).Assembly);

    // 추가 설정
    MappingConfig.Configure();

    services.AddSingleton(config);
    services.AddScoped<IMapper, ServiceMapper>();

    return services;
}
```

---

## Advanced Topics

### Two-Way Mapping

```csharp
// 양방향 매핑
// TwoWays()는 단순 1:1 속성에만 사용. 복합 매핑은 별도 역방향 설정 필요
TypeAdapterConfig<Order, OrderDto>.NewConfig()
    .TwoWays()
    .Map(dest => dest.Name, src => src.Name);
```

### Conditional Mapping

```csharp
TypeAdapterConfig<Order, OrderDto>.NewConfig()
    .Map(dest => dest.Discount,
        src => src.Customer.Tier == CustomerTier.Gold
            ? src.TotalAmount * 0.1m
            : 0m);
```

### Custom Value Resolver

```csharp
TypeAdapterConfig<Order, OrderDto>.NewConfig()
    .Map(dest => dest.FormattedTotal,
        src => $"{src.TotalAmount.Currency} {src.TotalAmount.Amount:N2}");
```

### Compiled Mapping (Performance)

```csharp
// 컴파일된 매핑 함수
var compiledMapper = TypeAdapterConfig<Order, OrderDto>
    .NewConfig()
    .Compile();
var dto = compiledMapper(order);

// EF Core 프로젝션
var dtos = await db.Orders
    .ProjectToType<OrderDto>()
    .ToListAsync();
```
