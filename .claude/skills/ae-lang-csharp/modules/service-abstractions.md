---
module: service-abstractions
version: "1.0.0"
last_updated: "2026-04-02"
category: infrastructure
---

# Service Abstractions

DI 패턴, Options pattern, Factory 패턴.

## Quick Reference

```csharp
// DI 등록
services.AddScoped<IOrderService, OrderService>();
services.AddSingleton<ICacheService, RedisCacheService>();
services.AddTransient<IEmailSender, SmtpEmailSender>();

// Options Pattern
services.Configure<SmtpOptions>(config.GetSection("Smtp"));

// Keyed Services (.NET 8+)
services.AddKeyedScoped<INotificationService, EmailService>("email");
services.AddKeyedScoped<INotificationService, SmsService>("sms");
```

---

## Detailed Patterns

### Layer별 DI 등록

```csharp
// Domain - 순수 도메인 로직, DI 없음

// Application Layer DI
public static class ApplicationDependencyInjection
{
    public static IServiceCollection AddApplication(this IServiceCollection services)
    {
        // Mapster 설정
        var config = TypeAdapterConfig.GlobalSettings;
        config.Scan(typeof(ApplicationDependencyInjection).Assembly);
        services.AddSingleton(config);
        services.AddScoped<IMapper, ServiceMapper>();

        // FluentValidation
        services.AddValidatorsFromAssemblyContaining<CreateOrderValidator>();

        return services;
    }
}

// Infrastructure Layer DI
public static class InfrastructureDependencyInjection
{
    public static IServiceCollection AddInfrastructure(
        this IServiceCollection services, IConfiguration config)
    {
        // EF Core
        services.AddDbContext<AppDbContext>(opts =>
            opts.UseSqlServer(config.GetConnectionString("Default")));

        // Repositories
        services.AddScoped<IOrderRepository, OrderRepository>();
        services.AddScoped<ICustomerRepository, CustomerRepository>();

        // External Services
        services.AddScoped<INotificationService, NotificationService>();

        // Interceptors
        services.AddScoped<DomainEventInterceptor>();
        services.AddScoped<AuditInterceptor>();

        return services;
    }
}

// Presentation Layer DI
public static class PresentationDependencyInjection
{
    public static IServiceCollection AddPresentation(this IServiceCollection services)
    {
        services.AddEndpointsApiExplorer();
        services.AddSwaggerGen();

        // ⚠️ 개발 환경 전용. 프로덕션에서는 특정 Origin만 허용할 것
        services.AddCors(opts =>
            opts.AddDefaultPolicy(policy =>
                policy.AllowAnyOrigin().AllowAnyMethod().AllowAnyHeader()));

        return services;
    }
}
```

### Options Pattern

```csharp
// 설정 클래스
public class SmtpOptions
{
    public const string SectionName = "Smtp";
    public required string Host { get; init; }
    public int Port { get; init; } = 587;
    public required string Username { get; init; }
    public required string Password { get; init; }
    public bool UseSsl { get; init; } = true;
}

// 유효성 검증 포함 등록
services.AddOptions<SmtpOptions>()
    .BindConfiguration(SmtpOptions.SectionName)
    .ValidateDataAnnotations()
    .ValidateOnStart();

// 사용 - IOptions<T> (Singleton, 변경 불가)
public class EmailSender(IOptions<SmtpOptions> options)
{
    private readonly SmtpOptions _smtp = options.Value;
}

// 사용 - IOptionsSnapshot<T> (Scoped, 요청마다 최신값)
public class ConfigurableService(IOptionsSnapshot<FeatureFlags> flags)
{
    public bool IsEnabled => flags.Value.NewFeatureEnabled;
}

// 사용 - IOptionsMonitor<T> (Singleton, 변경 감지)
public class MonitoredService(IOptionsMonitor<AppSettings> monitor)
{
    private readonly IDisposable? _changeToken = monitor.OnChange(settings =>
        Log.Information("Settings changed: {Setting}", settings.SomeValue));
}
```

### Factory Pattern

```csharp
// Keyed Services (.NET 8+)
services.AddKeyedScoped<IPaymentProcessor, StripeProcessor>("stripe");
services.AddKeyedScoped<IPaymentProcessor, PayPalProcessor>("paypal");
services.AddKeyedScoped<IPaymentProcessor, TossProcessor>("toss");

// 사용
public static class ProcessPaymentHandler
{
    public static async Task<PaymentResult> Handle(
        ProcessPayment command,
        [FromKeyedServices("stripe")] IPaymentProcessor stripe,
        CancellationToken ct)
    {
        return await stripe.ProcessAsync(command.Amount, command.Currency, ct);
    }
}

// 또는 Factory 패턴
public interface IPaymentProcessorFactory
{
    IPaymentProcessor Create(string provider);
}

public class PaymentProcessorFactory(IServiceProvider sp) : IPaymentProcessorFactory
{
    public IPaymentProcessor Create(string provider)
        => sp.GetRequiredKeyedService<IPaymentProcessor>(provider);
}
```

---

## Advanced Topics

### Decorator Pattern with DI

```csharp
// 기존 서비스에 캐싱 데코레이터 추가
services.AddScoped<OrderRepository>();
services.AddScoped<IOrderRepository>(sp =>
    new CachedOrderRepository(
        sp.GetRequiredService<OrderRepository>(),
        sp.GetRequiredService<IDistributedCache>()));

public class CachedOrderRepository(
    IOrderRepository inner,
    IDistributedCache cache) : IOrderRepository
{
    public async Task<Order?> FindByIdAsync(Guid id, CancellationToken ct)
    {
        var cacheKey = $"order:{id}";
        var cached = await cache.GetStringAsync(cacheKey, ct);
        if (cached is not null)
            return JsonSerializer.Deserialize<Order>(cached);

        var order = await inner.FindByIdAsync(id, ct);
        if (order is not null)
            await cache.SetStringAsync(cacheKey,
                JsonSerializer.Serialize(order),
                new DistributedCacheEntryOptions { AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(5) },
                ct);

        return order;
    }

    public Task AddAsync(Order order, CancellationToken ct) => inner.AddAsync(order, ct);
    public Task SaveChangesAsync(CancellationToken ct) => inner.SaveChangesAsync(ct);
}
```

### Health Checks

```csharp
builder.Services.AddHealthChecks()
    .AddDbContextCheck<AppDbContext>("database")
    .AddCheck<WolverineHealthCheck>("wolverine")
    .AddUrlGroup(new Uri("https://graph.microsoft.com/v1.0/$metadata"), "graph-api");

app.MapHealthChecks("/health", new HealthCheckOptions
{
    ResponseWriter = UIResponseWriter.WriteHealthCheckUIResponse
});
```
