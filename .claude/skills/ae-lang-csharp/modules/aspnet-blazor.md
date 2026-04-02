---
module: aspnet-blazor
version: "1.0.0"
last_updated: "2026-04-02"
category: web
---

# ASP.NET Core + Blazor

Minimal API, Wolverine 통합, Blazor 컴포넌트 패턴.

## Quick Reference

```csharp
// Program.cs - ASP.NET Core + Wolverine
var builder = WebApplication.CreateBuilder(args);

builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();
builder.Services.AddInfrastructure(builder.Configuration);
builder.Host.UseWolverine();
builder.Host.UseSerilog((ctx, cfg) =>
    cfg.ReadFrom.Configuration(ctx.Configuration));

var app = builder.Build();

if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseAuthentication();
app.UseAuthorization();

app.MapOrderEndpoints();
app.MapCustomerEndpoints();

app.Run();
```

Blazor 렌더 모드:

| 모드 | 설명 |
|------|------|
| `@rendermode InteractiveServer` | Server-side 렌더링, SignalR |
| `@rendermode InteractiveWebAssembly` | Client-side WASM |
| `@rendermode InteractiveAuto` | Server로 시작, WASM 다운로드 후 전환 |
| SSR (기본) | Static Server Rendering |

---

## Detailed Patterns

### Minimal API with Wolverine

```csharp
// 엔드포인트 그룹화
public static class OrderEndpoints
{
    public static void MapOrderEndpoints(this WebApplication app)
    {
        var group = app.MapGroup("/api/orders")
            .WithTags("Orders")
            .RequireAuthorization();

        group.MapPost("/", CreateOrder)
            .WithName("CreateOrder")
            .Produces<OrderCreated>(StatusCodes.Status201Created)
            .ProducesValidationProblem();

        group.MapGet("/{id:guid}", GetOrder)
            .WithName("GetOrder")
            .Produces<OrderDto>()
            .Produces(StatusCodes.Status404NotFound);

        group.MapGet("/", GetOrders)
            .WithName("GetOrders")
            .Produces<PagedResult<OrderDto>>();
    }

    private static async Task<IResult> CreateOrder(
        CreateOrder command,
        IMessageBus bus,
        CancellationToken ct)
    {
        var result = await bus.InvokeAsync<OrderCreated>(command, ct);
        return Results.Created($"/api/orders/{result.OrderId}", result);
    }

    private static async Task<IResult> GetOrder(
        Guid id,
        IMessageBus bus,
        CancellationToken ct)
    {
        var result = await bus.InvokeAsync<OrderDto?>(new GetOrder(id), ct);
        return result is not null ? Results.Ok(result) : Results.NotFound();
    }

    private static async Task<IResult> GetOrders(
        [AsParameters] GetOrders query,
        IMessageBus bus,
        CancellationToken ct)
    {
        var result = await bus.InvokeAsync<PagedResult<OrderDto>>(query, ct);
        return Results.Ok(result);
    }
}
```

### Middleware Pipeline

```csharp
// 커스텀 미들웨어
public class RequestLoggingMiddleware(RequestDelegate next, ILogger logger)
{
    public async Task InvokeAsync(HttpContext context)
    {
        var requestId = Guid.CreateVersion7().ToString("N")[..8];
        using (LogContext.PushProperty("RequestId", requestId))
        {
            logger.LogInformation("HTTP {Method} {Path} started",
                context.Request.Method, context.Request.Path);

            var sw = Stopwatch.StartNew();
            await next(context);
            sw.Stop();

            logger.LogInformation("HTTP {Method} {Path} completed {StatusCode} in {Elapsed}ms",
                context.Request.Method, context.Request.Path,
                context.Response.StatusCode, sw.ElapsedMilliseconds);
        }
    }
}

// 등록
app.UseMiddleware<RequestLoggingMiddleware>();
```

### Authentication / Authorization

```csharp
// JWT Bearer 인증
builder.Services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
    .AddJwtBearer(opts =>
    {
        opts.Authority = builder.Configuration["Auth:Authority"];
        opts.Audience = builder.Configuration["Auth:Audience"];
    });

builder.Services.AddAuthorization(opts =>
{
    opts.AddPolicy("AdminOnly", policy =>
        policy.RequireRole("Admin"));

    opts.AddPolicy("OrderOwner", policy =>
        policy.AddRequirements(new OrderOwnerRequirement()));
});

// 엔드포인트에 정책 적용
group.MapDelete("/{id:guid}", DeleteOrder)
    .RequireAuthorization("AdminOnly");
```

### Blazor Component

```razor
@* OrderList.razor *@
@rendermode InteractiveServer

<h3>주문 목록</h3>

@if (_orders is null)
{
    <p>로딩 중...</p>
}
else if (_orders.Count == 0)
{
    <p>주문이 없습니다.</p>
}
else
{
    <table class="table">
        <thead>
            <tr>
                <th>주문 ID</th>
                <th>고객</th>
                <th>금액</th>
                <th>상태</th>
            </tr>
        </thead>
        <tbody>
            @foreach (var order in _orders)
            {
                <tr>
                    <td>@order.Id.ToString()[..8]</td>
                    <td>@order.CustomerName</td>
                    <td>@order.TotalAmount.ToString("C")</td>
                    <td><StatusBadge Status="@order.Status" /></td>
                </tr>
            }
        </tbody>
    </table>
}

@code {
    [Inject] private HttpClient Http { get; set; } = default!;

    private List<OrderDto>? _orders;

    protected override async Task OnInitializedAsync()
    {
        _orders = await Http.GetFromJsonAsync<List<OrderDto>>("/api/orders");
    }
}
```

---

## Advanced Topics

### SSR Streaming

```razor
@* 스트리밍 렌더링으로 점진적 UI 표시 *@
@attribute [StreamRendering]

@if (_data is null)
{
    <LoadingSpinner />
}
else
{
    <DataGrid Items="@_data" />
}

@code {
    private List<OrderDto>? _data;

    protected override async Task OnInitializedAsync()
    {
        // 비동기 데이터 로딩 - UI가 점진적으로 업데이트됨
        _data = await OrderService.GetOrdersAsync();
    }
}
```

### Enhanced Navigation

```razor
@* 클라이언트 사이드 네비게이션 (SPA-like 경험) *@
<NavLink href="/orders" Match="NavLinkMatch.Prefix">
    주문 관리
</NavLink>

@* 폼 제출 후 네비게이션 *@
<EditForm Model="@_model" OnValidSubmit="HandleSubmit" Enhance>
    <DataAnnotationsValidator />
    <InputText @bind-Value="_model.Name" />
    <button type="submit">저장</button>
</EditForm>
```

### Error Boundary

```razor
<ErrorBoundary @ref="_errorBoundary">
    <ChildContent>
        <OrderDetails OrderId="@OrderId" />
    </ChildContent>
    <ErrorContent Context="ex">
        <div class="alert alert-danger">
            <h4>오류 발생</h4>
            <p>@ex.Message</p>
            <button @onclick="() => _errorBoundary?.Recover()">다시 시도</button>
        </div>
    </ErrorContent>
</ErrorBoundary>

@code {
    private ErrorBoundary? _errorBoundary;
}
```
