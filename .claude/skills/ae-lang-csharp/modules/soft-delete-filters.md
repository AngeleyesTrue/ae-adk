---
module: soft-delete-filters
version: "1.0.0"
last_updated: "2026-04-02"
category: data
---

# Soft Delete + Global Query Filters

ISoftDeletable 인터페이스와 EF Core Global Query Filter 패턴.

## Quick Reference

```csharp
// 인터페이스
public interface ISoftDeletable
{
    bool IsDeleted { get; set; }
    DateTimeOffset? DeletedAt { get; set; }
    string? DeletedBy { get; set; }
}

// Global Query Filter (자동으로 IsDeleted = false 조건 추가)
builder.HasQueryFilter(e => !e.IsDeleted);

// 삭제된 레코드 포함 조회
db.Orders.IgnoreQueryFilters().ToListAsync(ct);
```

---

## Detailed Patterns

### ISoftDeletable Interface

```csharp
public interface ISoftDeletable
{
    bool IsDeleted { get; set; }
    DateTimeOffset? DeletedAt { get; set; }
    string? DeletedBy { get; set; }
}

// Entity에 구현
public class Order : AggregateRoot, ISoftDeletable
{
    // ... 기존 속성 ...
    public bool IsDeleted { get; set; }
    public DateTimeOffset? DeletedAt { get; set; }
    public string? DeletedBy { get; set; }
}
```

### SaveChangesInterceptor로 Soft Delete 구현

```csharp
public class SoftDeleteInterceptor(ICurrentUserService currentUser)
    : SaveChangesInterceptor
{
    public override ValueTask<InterceptionResult<int>> SavingChangesAsync(
        DbContextEventData eventData,
        InterceptionResult<int> result,
        CancellationToken ct = default)
    {
        var context = eventData.Context;
        if (context is null) return ValueTask.FromResult(result);

        foreach (var entry in context.ChangeTracker.Entries<ISoftDeletable>())
        {
            if (entry.State == EntityState.Deleted)
            {
                // 실제 삭제 대신 소프트 삭제로 변환
                entry.State = EntityState.Modified;
                entry.Entity.IsDeleted = true;
                entry.Entity.DeletedAt = DateTimeOffset.UtcNow;
                entry.Entity.DeletedBy = currentUser.UserId;
            }
        }

        return ValueTask.FromResult(result);
    }
}
```

### Global Query Filter 설정

```csharp
// DbContext의 OnModelCreating에서 자동 적용
protected override void OnModelCreating(ModelBuilder modelBuilder)
{
    base.OnModelCreating(modelBuilder);

    // ISoftDeletable을 구현한 모든 엔티티에 필터 적용
    foreach (var entityType in modelBuilder.Model.GetEntityTypes())
    {
        if (typeof(ISoftDeletable).IsAssignableFrom(entityType.ClrType))
        {
            var parameter = Expression.Parameter(entityType.ClrType, "e");
            var property = Expression.Property(parameter,
                nameof(ISoftDeletable.IsDeleted));
            var condition = Expression.Equal(property,
                Expression.Constant(false));
            var lambda = Expression.Lambda(condition, parameter);

            modelBuilder.Entity(entityType.ClrType).HasQueryFilter(lambda);
        }
    }
}
```

### 삭제된 레코드 조회

```csharp
// 기본 조회 - 삭제된 레코드 자동 제외
var activeOrders = await db.Orders.ToListAsync(ct);

// 삭제된 레코드 포함 조회
var allOrders = await db.Orders
    .IgnoreQueryFilters()
    .ToListAsync(ct);

// 삭제된 레코드만 조회
var deletedOrders = await db.Orders
    .IgnoreQueryFilters()
    .Where(o => o.IsDeleted)
    .ToListAsync(ct);
```

### Audit Trail 통합

```csharp
// Soft Delete 시 감사 로그 남기기
public class SoftDeleteAuditInterceptor(
    ICurrentUserService currentUser,
    ILogger logger) : SaveChangesInterceptor
{
    public override ValueTask<InterceptionResult<int>> SavingChangesAsync(
        DbContextEventData eventData,
        InterceptionResult<int> result,
        CancellationToken ct = default)
    {
        var context = eventData.Context;
        if (context is null) return ValueTask.FromResult(result);

        foreach (var entry in context.ChangeTracker.Entries<ISoftDeletable>())
        {
            if (entry.State == EntityState.Deleted)
            {
                entry.State = EntityState.Modified;
                entry.Entity.IsDeleted = true;
                entry.Entity.DeletedAt = DateTimeOffset.UtcNow;
                entry.Entity.DeletedBy = currentUser.UserId;

                logger.LogInformation(
                    "Soft deleted {EntityType} {EntityId} by {UserId}",
                    entry.Entity.GetType().Name,
                    ((BaseEntity)entry.Entity).Id,
                    currentUser.UserId);
            }
        }

        return ValueTask.FromResult(result);
    }
}
```

---

## Advanced Topics

### Cascade Soft Delete

```csharp
// 부모 삭제 시 자식도 소프트 삭제
public class CascadeSoftDeleteService(AppDbContext db)
{
    public async Task SoftDeleteWithCascadeAsync<T>(
        T entity, CancellationToken ct) where T : BaseEntity, ISoftDeletable
    {
        entity.IsDeleted = true;
        entity.DeletedAt = DateTimeOffset.UtcNow;

        // 네비게이션 프로퍼티를 통한 자식 엔티티 소프트 삭제
        var entry = db.Entry(entity);
        foreach (var navigation in entry.Navigations)
        {
            if (navigation.CurrentValue is IEnumerable<ISoftDeletable> children)
            {
                foreach (var child in children)
                {
                    child.IsDeleted = true;
                    child.DeletedAt = DateTimeOffset.UtcNow;
                }
            }
        }

        await db.SaveChangesAsync(ct);
    }
}
```

### Unique Index with Soft Delete

```csharp
// 소프트 삭제된 레코드를 제외한 유니크 인덱스
builder.HasIndex(e => e.Email)
    .HasFilter("IsDeleted = 0")
    .IsUnique();
```

### Restore (Undelete)

```csharp
public static class RestoreOrderHandler
{
    public static async Task Handle(
        RestoreOrder command,
        AppDbContext db,
        ILogger logger,
        CancellationToken ct)
    {
        var order = await db.Orders
            .IgnoreQueryFilters()
            .FirstOrDefaultAsync(o => o.Id == command.OrderId && o.IsDeleted, ct)
            ?? throw new NotFoundException("Deleted order not found");

        order.IsDeleted = false;
        order.DeletedAt = null;
        order.DeletedBy = null;

        await db.SaveChangesAsync(ct);
        logger.LogInformation("Order {OrderId} restored", command.OrderId);
    }
}
```
