---
module: dotnet-platform
version: "1.0.0"
last_updated: "2026-04-02"
category: platform
---

# .NET Platform (.NET 10 / C# 13-14)

.NET 10 LTS 및 C# 13/14 최신 기능 기반 플랫폼 가이드.

## Quick Reference

Target Framework: `net10.0` (.NET 10 LTS, 2026-11 출시 예정, 현재 .NET 9 기반 개발)
Language Version: C# 13/14

핵심 C# 13/14 기능:

```csharp
// Primary Constructor (C# 12+)
public class OrderService(IOrderRepository repo, ILogger logger)
{
    public async Task<Result<Order>> GetAsync(Guid id, CancellationToken ct)
    {
        var order = await repo.FindByIdAsync(id, ct);
        return order is null
            ? Result<Order>.NotFound($"Order {id} not found")
            : Result<Order>.Success(order);
    }
}

// Collection Expression (C# 12+)
List<string> names = ["Alice", "Bob", "Charlie"];
int[] numbers = [1, 2, 3, 4, 5];
ReadOnlySpan<byte> bytes = [0x01, 0x02, 0x03];

// Pattern Matching 강화
var discount = customer switch
{
    { Tier: CustomerTier.Gold, OrderCount: > 100 } => 0.20m,
    { Tier: CustomerTier.Silver, OrderCount: > 50 } => 0.10m,
    { IsNewCustomer: true } => 0.05m,
    _ => 0m
};

// Guid.CreateVersion7() - 순차적 GUID
var id = Guid.CreateVersion7();
```

SDK 명령:

```bash
dotnet new webapi -n MyApp.Web --framework net10.0
dotnet new classlib -n MyApp.Domain --framework net10.0
dotnet build
dotnet test
dotnet run --project src/MyApp.Web
```

---

## Detailed Patterns

### Global Usings

```csharp
// GlobalUsings.cs (프로젝트 루트)
global using System.Collections.Immutable;
global using Ardalis.GuardClauses;
global using Serilog;
global using Mapster;
```

### Directory.Build.props

```xml
<Project>
  <PropertyGroup>
    <TargetFramework>net10.0</TargetFramework>
    <LangVersion>latest</LangVersion>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
    <AnalysisLevel>latest-recommended</AnalysisLevel>
  </PropertyGroup>
</Project>
```

### Nullable Reference Types

```csharp
// null 안전 패턴
public class Customer
{
    public required string Name { get; init; }
    public string? MiddleName { get; init; }
    public required Email Email { get; init; }

    // null 검증은 Guard.Against 사용
    public static Customer Create(string name, string email)
    {
        Guard.Against.NullOrWhiteSpace(name, nameof(name));
        Guard.Against.NullOrWhiteSpace(email, nameof(email));

        return new Customer
        {
            Name = name,
            Email = Email.Create(email)
        };
    }
}
```

### File-Scoped Namespaces

```csharp
// 파일 전체에 단일 네임스페이스 적용
namespace MyApp.Domain.Entities;

public class Order : AggregateRoot
{
    // 들여쓰기 한 단계 줄어듦
}
```

### Result Pattern

```csharp
// 예외 대신 Result<T> 패턴 사용
public abstract record Result<T>
{
    public bool IsSuccess { get; init; }
    public T? Value { get; init; }
    public string[] Errors { get; init; } = [];

    public static Result<T> Success(T value) => new SuccessResult<T>(value);
    public static Result<T> Error(params string[] errors) => new ErrorResult<T>(errors);
    public static Result<T> NotFound(string message) => new NotFoundResult<T>(message);
}
```

---

## Advanced Topics

### Native AOT 고려사항

```csharp
// Native AOT 호환을 위한 JSON 직렬화 컨텍스트
[JsonSerializable(typeof(OrderDto))]
[JsonSerializable(typeof(List<OrderDto>))]
public partial class AppJsonContext : JsonSerializerContext { }

// Program.cs
builder.Services.ConfigureHttpJsonOptions(opts =>
    opts.SerializerOptions.TypeInfoResolverChain.Add(AppJsonContext.Default));
```

### Performance Tips

```csharp
// Span<T>과 Memory<T> 활용
public static bool ContainsDigit(ReadOnlySpan<char> text)
{
    foreach (var c in text)
        if (char.IsDigit(c)) return true;
    return false;
}

// FrozenDictionary (읽기 전용, 빠른 조회)
var lookup = items.ToFrozenDictionary(x => x.Key, x => x.Value);

// SearchValues (문자열 검색 최적화)
private static readonly SearchValues<char> _digits = SearchValues.Create("0123456789");
public static bool HasDigit(ReadOnlySpan<char> text) => text.ContainsAny(_digits);
```

### Guid.CreateVersion7

```csharp
// UUID v7: 타임스탬프 기반 순차적 GUID
// DB 인덱스 성능 향상 (B-tree 순차 삽입)
var id = Guid.CreateVersion7();

// Entity 기본 키에 사용
public abstract class BaseEntity
{
    public Guid Id { get; protected set; } = Guid.CreateVersion7();
    public DateTimeOffset CreatedAt { get; protected set; } = DateTimeOffset.UtcNow;
}
```
