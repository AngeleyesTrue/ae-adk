---
module: wolverine-middleware
version: "1.0.0"
last_updated: "2026-04-02"
category: cqrs
---

# Wolverine Middleware

Wolverine 미들웨어 체인 패턴. Before/After 컨벤션 기반 파이프라인.

## Quick Reference

```csharp
// Before: 핸들러 실행 전 호출
// After: 핸들러 실행 후 호출
public class LoggingMiddleware
{
    public static void Before(ILogger logger, Envelope envelope)
        => logger.LogInformation("Handling {MessageType}", envelope.MessageType);

    public static void After(ILogger logger, Envelope envelope)
        => logger.LogInformation("Handled {MessageType}", envelope.MessageType);
}

// 등록
opts.Policies.AddMiddleware<LoggingMiddleware>();
```

---

## Detailed Patterns

### FluentValidation Middleware

```csharp
// 핸들러 실행 전 유효성 검증
public class FluentValidationMiddleware
{
    public static async Task<ProblemDetails?> BeforeAsync<T>(
        T message,
        IEnumerable<IValidator<T>> validators,
        CancellationToken ct) where T : class
    {
        if (!validators.Any()) return null;

        var context = new ValidationContext<T>(message);
        var results = await Task.WhenAll(
            validators.Select(v => v.ValidateAsync(context, ct)));

        var failures = results
            .SelectMany(r => r.Errors)
            .Where(f => f is not null)
            .ToList();

        if (failures.Count == 0) return null;

        // ProblemDetails 반환 시 핸들러 실행 중단
        return new ProblemDetails
        {
            Status = StatusCodes.Status400BadRequest,
            Title = "Validation failed",
            Extensions = { ["errors"] = failures.Select(f => new
            {
                f.PropertyName,
                f.ErrorMessage
            })}
        };
    }
}

// 등록
builder.Host.UseWolverine(opts =>
{
    opts.Policies.AddMiddleware<FluentValidationMiddleware>();
});

// Validator 예시
public class CreateOrderValidator : AbstractValidator<CreateOrder>
{
    public CreateOrderValidator()
    {
        RuleFor(x => x.CustomerId).NotEmpty().MaximumLength(100);
        RuleFor(x => x.Lines).NotEmpty();
        RuleForEach(x => x.Lines).ChildRules(line =>
        {
            line.RuleFor(l => l.Sku).NotEmpty();
            line.RuleFor(l => l.Quantity).GreaterThan(0);
        });
    }
}
```

### Logging Middleware

```csharp
public class DetailedLoggingMiddleware
{
    public static void Before(ILogger logger, Envelope envelope)
    {
        logger.LogInformation(
            "==> Handling {MessageType} (CorrelationId: {CorrelationId})",
            envelope.MessageType,
            envelope.CorrelationId);
    }

    public static void After(ILogger logger, Envelope envelope)
    {
        logger.LogInformation(
            "<== Handled {MessageType} (CorrelationId: {CorrelationId})",
            envelope.MessageType,
            envelope.CorrelationId);
    }

    // 예외 발생 시
    public static void Finally(ILogger logger, Envelope envelope, Exception? ex)
    {
        if (ex is not null)
        {
            logger.LogError(ex,
                "!!! Error handling {MessageType}: {Error}",
                envelope.MessageType, ex.Message);
        }
    }
}
```

### Transaction Middleware

> **주의**: Wolverine의 `opts.Policies.AutoApplyTransactions()` 사용 시 이 수동 트랜잭션 미들웨어와 충돌할 수 있습니다.
> 수동 트랜잭션 관리가 필요한 경우에만 사용하고, Wolverine 자동 트랜잭션은 비활성화하세요.

```csharp
// DbContext 트랜잭션 래핑
public class TransactionMiddleware
{
    // Before 미들웨어에서 null 반환 시 Wolverine이 핸들러를 계속 실행한다.
    // non-null 반환 시 핸들러 실행을 중단하고 해당 값을 응답으로 사용한다.
    public static async Task<IResult?> BeforeAsync(
        AppDbContext db,
        CancellationToken ct)
    {
        await db.Database.BeginTransactionAsync(ct);
        return null; // null 반환 = 핸들러 계속 실행 (Wolverine Before 컨벤션)
    }

    public static async Task AfterAsync(AppDbContext db, CancellationToken ct)
    {
        await db.Database.CommitTransactionAsync(ct);
    }

    public static async Task FinallyAsync(AppDbContext db, Exception? ex)
    {
        if (ex is not null && db.Database.CurrentTransaction is not null)
        {
            await db.Database.RollbackTransactionAsync();
        }
    }
}
```

### Exception Handling Middleware

```csharp
public class ExceptionHandlingMiddleware
{
    public static void Finally(ILogger logger, Envelope envelope, Exception? ex)
    {
        if (ex is DomainException domainEx)
        {
            logger.LogWarning("Domain exception in {MessageType}: {Errors}",
                envelope.MessageType,
                string.Join(", ", domainEx.Errors));
        }
        else if (ex is NotFoundException notFoundEx)
        {
            logger.LogWarning("Not found in {MessageType}: {Message}",
                envelope.MessageType, notFoundEx.Message);
        }
        else if (ex is not null)
        {
            logger.LogError(ex, "Unhandled exception in {MessageType}",
                envelope.MessageType);
        }
    }
}
```

---

## Advanced Topics

### Conditional Middleware

```csharp
// 특정 메시지 타입에만 미들웨어 적용
builder.Host.UseWolverine(opts =>
{
    // 모든 Command에 유효성 검증 적용
    opts.Policies.AddMiddleware<FluentValidationMiddleware>(chain =>
        chain.MessageType.Name.EndsWith("Command") ||
        chain.MessageType.Name.StartsWith("Create") ||
        chain.MessageType.Name.StartsWith("Update"));

    // 모든 핸들러에 로깅 적용
    opts.Policies.AddMiddleware<LoggingMiddleware>();
});
```

### Middleware Ordering

```csharp
// 미들웨어는 등록 순서대로 실행
builder.Host.UseWolverine(opts =>
{
    // 1. 로깅 (가장 바깥)
    opts.Policies.AddMiddleware<LoggingMiddleware>();
    // 2. 유효성 검증
    opts.Policies.AddMiddleware<FluentValidationMiddleware>();
    // 3. 트랜잭션 (가장 안쪽)
    opts.Policies.AddMiddleware<TransactionMiddleware>();
    // → Handler 실행
});
```

### Performance Middleware

```csharp
public class PerformanceMiddleware
{
    public static Stopwatch Before() => Stopwatch.StartNew();

    public static void After(Stopwatch sw, ILogger logger, Envelope envelope)
    {
        sw.Stop();
        if (sw.ElapsedMilliseconds > 500)
        {
            logger.LogWarning(
                "Slow handler: {MessageType} took {Elapsed}ms",
                envelope.MessageType, sw.ElapsedMilliseconds);
        }
    }
}
```

---
**관련 모듈**: [Wolverine CQRS](wolverine-cqrs.md) | [Domain Events](domain-events.md)
