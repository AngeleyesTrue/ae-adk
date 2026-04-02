---
module: project-templates
version: "1.0.0"
last_updated: "2026-04-02"
category: infrastructure
---

# Project Templates

Clean Architecture 4-layer 솔루션 구조 및 프로젝트 템플릿.

## Quick Reference

```
MySolution/
├── src/
│   ├── MyApp.Domain/           # 도메인 모델, Value Objects, Aggregate Roots
│   ├── MyApp.Application/      # Commands, Queries, Handlers (Wolverine)
│   ├── MyApp.Infrastructure/   # EF Core, External Services, Repositories
│   └── MyApp.Web/              # ASP.NET Core, Blazor, Endpoints
├── tests/
│   ├── MyApp.Domain.UnitTests/
│   ├── MyApp.Application.UnitTests/
│   ├── MyApp.Application.IntegrationTests/
│   └── MyApp.Web.E2ETests/
├── Directory.Build.props
└── MySolution.sln
```

---

## Detailed Patterns

### 프로젝트 참조 방향

```
Domain ← Application ← Infrastructure ← Web
  (없음)    (Domain만)    (App+Domain)    (모두)
```

```xml
<!-- MyApp.Application.csproj -->
<ItemGroup>
  <ProjectReference Include="..\MyApp.Domain\MyApp.Domain.csproj" />
</ItemGroup>

<!-- MyApp.Infrastructure.csproj -->
<ItemGroup>
  <ProjectReference Include="..\MyApp.Application\MyApp.Application.csproj" />
  <ProjectReference Include="..\MyApp.Domain\MyApp.Domain.csproj" />
</ItemGroup>

<!-- MyApp.Web.csproj -->
<ItemGroup>
  <ProjectReference Include="..\MyApp.Infrastructure\MyApp.Infrastructure.csproj" />
  <ProjectReference Include="..\MyApp.Application\MyApp.Application.csproj" />
</ItemGroup>
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
    <EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>
  </PropertyGroup>

  <PropertyGroup>
    <Authors>Angeleyes</Authors>
    <Company>SPK</Company>
  </PropertyGroup>
</Project>
```

### Layer별 패키지

```xml
<!-- MyApp.Domain.csproj - 최소 의존성 -->
<ItemGroup>
  <PackageReference Include="Ardalis.GuardClauses" Version="5.*" />
</ItemGroup>

<!-- MyApp.Application.csproj -->
<ItemGroup>
  <PackageReference Include="WolverineFx" Version="3.*" />
  <PackageReference Include="Mapster" Version="7.*" />
  <PackageReference Include="FluentValidation" Version="11.*" />
  <PackageReference Include="Serilog.Extensions.Logging" Version="8.*" />
</ItemGroup>

<!-- MyApp.Infrastructure.csproj -->
<ItemGroup>
  <PackageReference Include="Microsoft.EntityFrameworkCore.SqlServer" Version="9.*" />
  <PackageReference Include="Microsoft.EntityFrameworkCore.Design" Version="9.*" />
  <PackageReference Include="WolverineFx.SqlServer" Version="3.*" />
  <PackageReference Include="Serilog.Sinks.Console" Version="6.*" />
  <PackageReference Include="Serilog.Sinks.Seq" Version="8.*" />
</ItemGroup>

<!-- MyApp.Web.csproj -->
<ItemGroup>
  <PackageReference Include="Serilog.AspNetCore" Version="8.*" />
  <PackageReference Include="Swashbuckle.AspNetCore" Version="7.*" />
</ItemGroup>
```

### Program.cs (Web 프로젝트)

```csharp
using Serilog;

Log.Logger = new LoggerConfiguration()
    .WriteTo.Console()
    .CreateBootstrapLogger();

try
{
    var builder = WebApplication.CreateBuilder(args);

    builder.Host.UseSerilog((ctx, cfg) =>
        cfg.ReadFrom.Configuration(ctx.Configuration));

    // Layer별 DI 등록
    builder.Services
        .AddApplication()
        .AddInfrastructure(builder.Configuration)
        .AddPresentation();

    // Wolverine
    builder.Host.UseWolverine(opts =>
    {
        opts.Discovery.IncludeAssembly(typeof(CreateOrderHandler).Assembly);
        opts.PersistMessagesWithSqlServer(
            builder.Configuration.GetConnectionString("Default")!);
    });

    var app = builder.Build();

    if (app.Environment.IsDevelopment())
    {
        app.UseSwagger();
        app.UseSwaggerUI();
    }

    app.UseSerilogRequestLogging();
    app.UseAuthentication();
    app.UseAuthorization();

    // 엔드포인트 매핑
    app.MapOrderEndpoints();
    app.MapCustomerEndpoints();
    app.MapHealthChecks("/health");

    app.Run();
}
catch (Exception ex)
{
    Log.Fatal(ex, "Application terminated unexpectedly");
}
finally
{
    Log.CloseAndFlush();
}
```

---

## Advanced Topics

### Feature Folders

```
src/MyApp.Application/
├── Orders/
│   ├── Commands/
│   │   ├── CreateOrder.cs
│   │   ├── CreateOrderHandler.cs
│   │   └── CreateOrderValidator.cs
│   ├── Queries/
│   │   ├── GetOrder.cs
│   │   └── GetOrderHandler.cs
│   ├── Events/
│   │   ├── OrderPlaced.cs
│   │   └── OrderPlacedHandler.cs
│   └── Dtos/
│       └── OrderDto.cs
├── Customers/
│   ├── Commands/
│   └── Queries/
└── Shared/
    ├── Interfaces/
    └── Behaviors/
```

### Multi-Tenant 구조

```csharp
// Tenant별 DbContext 생성
public class TenantDbContextFactory(
    IHttpContextAccessor httpContext,
    IConfiguration config)
{
    public AppDbContext CreateContext()
    {
        var tenantId = httpContext.HttpContext?.User.FindFirst("tenant_id")?.Value
            ?? throw new UnauthorizedException("Tenant not found");

        var connectionString = config.GetConnectionString(tenantId);
        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseSqlServer(connectionString)
            .Options;

        return new AppDbContext(options);
    }
}
```
