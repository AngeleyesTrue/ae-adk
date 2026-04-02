---
module: efcore-conventions
version: "1.0.0"
last_updated: "2026-04-02"
category: data
---

# EF Core Conventions

Entity Framework Core 컨벤션, 설정, 마이그레이션 패턴.

## Quick Reference

```csharp
// DbContext 기본 설정
public class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
{
    public DbSet<Order> Orders => Set<Order>();
    public DbSet<Customer> Customers => Set<Customer>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        // 현재 어셈블리의 모든 IEntityTypeConfiguration 자동 적용
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(AppDbContext).Assembly);
    }
}
```

마이그레이션 명령:

```bash
dotnet ef migrations add InitialCreate --project src/MyApp.Infrastructure --startup-project src/MyApp.Web
dotnet ef database update --project src/MyApp.Infrastructure --startup-project src/MyApp.Web
dotnet ef migrations script --idempotent -o migration.sql
```

---

## Detailed Patterns

### Entity Configuration (IEntityTypeConfiguration)

```csharp
// 엔티티별 설정 분리
public class OrderConfiguration : IEntityTypeConfiguration<Order>
{
    public void Configure(EntityTypeBuilder<Order> builder)
    {
        builder.ToTable("Orders");
        builder.HasKey(o => o.Id);

        builder.Property(o => o.CustomerId)
            .IsRequired()
            .HasMaxLength(100);

        builder.Property(o => o.Status)
            .HasConversion<string>()
            .HasMaxLength(50);

        builder.Property(o => o.TotalAmount)
            .HasPrecision(18, 2);

        // Owned Type (Value Object)
        builder.OwnsOne(o => o.ShippingAddress, address =>
        {
            address.Property(a => a.Street).HasMaxLength(200);
            address.Property(a => a.City).HasMaxLength(100);
            address.Property(a => a.ZipCode).HasMaxLength(20);
        });

        // 컬렉션 관계
        builder.HasMany(o => o.Lines)
            .WithOne()
            .HasForeignKey(l => l.OrderId)
            .OnDelete(DeleteBehavior.Cascade);

        // 인덱스
        builder.HasIndex(o => o.CustomerId);
        builder.HasIndex(o => o.CreatedAt);
    }
}
```

### Naming Conventions

```csharp
// 스네이크 케이스 컨벤션 (PostgreSQL 권장)
protected override void ConfigureConventions(ModelConfigurationBuilder configBuilder)
{
    configBuilder.Properties<string>().HaveMaxLength(256);
    configBuilder.Properties<decimal>().HavePrecision(18, 2);
}

// 또는 EFCore.NamingConventions 패키지
services.AddDbContext<AppDbContext>(opts =>
    opts.UseNpgsql(connectionString)
        .UseSnakeCaseNamingConvention());
```

### Timestamp Conventions

```csharp
// 자동 타임스탬프 관리
public interface IAuditable
{
    DateTimeOffset CreatedAt { get; set; }
    string? CreatedBy { get; set; }
    DateTimeOffset? UpdatedAt { get; set; }
    string? UpdatedBy { get; set; }
}

// SaveChanges에서 자동 설정
public override async Task<int> SaveChangesAsync(CancellationToken ct = default)
{
    foreach (var entry in ChangeTracker.Entries<IAuditable>())
    {
        if (entry.State == EntityState.Added)
            entry.Entity.CreatedAt = DateTimeOffset.UtcNow;

        if (entry.State == EntityState.Modified)
            entry.Entity.UpdatedAt = DateTimeOffset.UtcNow;
    }
    return await base.SaveChangesAsync(ct);
}
```

### Value Object Mapping

```csharp
// Value Object을 Owned Type으로 매핑
public class Money : ValueObject
{
    public decimal Amount { get; private set; }
    public string Currency { get; private set; } = default!;

    protected override IEnumerable<object> GetEqualityComponents()
    {
        yield return Amount;
        yield return Currency;
    }
}

// Configuration
builder.OwnsOne(o => o.TotalAmount, money =>
{
    money.Property(m => m.Amount).HasColumnName("TotalAmount").HasPrecision(18, 2);
    money.Property(m => m.Currency).HasColumnName("TotalCurrency").HasMaxLength(3);
});
```

### Soft Delete Filter

```csharp
// Global Query Filter로 소프트 삭제 구현
// 전체 정의는 [Soft Delete Filters](soft-delete-filters.md) 참조
public interface ISoftDeletable
{
    bool IsDeleted { get; set; }
    DateTimeOffset? DeletedAt { get; set; }
    string? DeletedBy { get; set; }
}

// OnModelCreating에서 필터 적용
foreach (var entityType in modelBuilder.Model.GetEntityTypes())
{
    if (typeof(ISoftDeletable).IsAssignableFrom(entityType.ClrType))
    {
        var parameter = Expression.Parameter(entityType.ClrType, "e");
        var property = Expression.Property(parameter, nameof(ISoftDeletable.IsDeleted));
        var condition = Expression.Equal(property, Expression.Constant(false));
        var lambda = Expression.Lambda(condition, parameter);

        modelBuilder.Entity(entityType.ClrType).HasQueryFilter(lambda);
    }
}
```

---

## Advanced Topics

### Bulk Operations

```csharp
// ExecuteUpdate - SQL UPDATE 직접 실행
await db.Orders
    .Where(o => o.Status == OrderStatus.Expired)
    .ExecuteUpdateAsync(s => s
        .SetProperty(o => o.Status, OrderStatus.Cancelled)
        .SetProperty(o => o.UpdatedAt, DateTimeOffset.UtcNow), ct);

// ExecuteDelete - SQL DELETE 직접 실행
await db.AuditLogs
    .Where(l => l.CreatedAt < cutoffDate)
    .ExecuteDeleteAsync(ct);
```

### Split Queries

```csharp
// 카테시안 곱 방지를 위한 Split Query
var orders = await db.Orders
    .Include(o => o.Lines)
    .Include(o => o.Payments)
    .AsSplitQuery()
    .ToListAsync(ct);

// 글로벌 설정
services.AddDbContext<AppDbContext>(opts =>
    opts.UseSqlServer(connectionString,
        sql => sql.UseQuerySplittingBehavior(QuerySplittingBehavior.SplitQuery)));
```

### Concurrency Token

```csharp
// 낙관적 동시성 제어
public class Order : AggregateRoot
{
    public byte[] RowVersion { get; private set; } = [];
}

// Configuration
builder.Property(o => o.RowVersion)
    .IsRowVersion();
```

---
**관련 모듈**: [EF Core Advanced](efcore-advanced.md) | [Soft Delete Filters](soft-delete-filters.md) | [Aggregate Patterns](aggregate-patterns.md)
