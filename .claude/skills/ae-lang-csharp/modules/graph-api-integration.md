---
module: graph-api-integration
version: "1.0.0"
last_updated: "2026-04-02"
category: infrastructure
---

# Microsoft Graph API Integration

GraphServiceClient, 인증, Batch 요청, Result 패턴.

## Quick Reference

```csharp
// 인증 설정
var credential = new ClientSecretCredential(tenantId, clientId, clientSecret);
var graphClient = new GraphServiceClient(credential, ["https://graph.microsoft.com/.default"]);

// 사용자 조회
var user = await graphClient.Users[userId].GetAsync();

// 서비스 등록
services.AddSingleton(graphClient);
```

---

## Detailed Patterns

### Authentication Setup

```csharp
// ClientSecretCredential (Application 권한)
public static class GraphDependencyInjection
{
    public static IServiceCollection AddGraphApi(
        this IServiceCollection services, IConfiguration config)
    {
        var graphConfig = config.GetSection("MicrosoftGraph");

        var credential = new ClientSecretCredential(
            graphConfig["TenantId"],
            graphConfig["ClientId"],
            graphConfig["ClientSecret"]);

        var graphClient = new GraphServiceClient(credential,
            ["https://graph.microsoft.com/.default"]);

        services.AddSingleton(graphClient);
        services.AddScoped<IGraphUserService, GraphUserService>();
        services.AddScoped<IGraphGroupService, GraphGroupService>();

        return services;
    }
}

// appsettings.json
// {
//   "MicrosoftGraph": {
//     "TenantId": "your-tenant-id",
//     "ClientId": "your-client-id",
//     "ClientSecret": "your-client-secret" // -> Azure Key Vault 사용 권장
//   }
// }

// 프로덕션: Azure Key Vault 사용
// builder.Configuration.AddAzureKeyVault(
//     new Uri("https://your-vault.vault.azure.net/"),
//     new DefaultAzureCredential());

// 개발 환경: User Secrets 사용
// dotnet user-secrets set "MicrosoftGraph:ClientSecret" "your-secret"
```

### Service Abstraction with Result Pattern

```csharp
public interface IGraphUserService
{
    Task<Result<GraphUser>> GetUserAsync(string userId, CancellationToken ct = default);
    Task<Result<IReadOnlyList<GraphUser>>> GetGroupMembersAsync(
        string groupId, CancellationToken ct = default);
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
        catch (ODataError ex) when (ex.ResponseStatusCode == 404)
        {
            logger.LogWarning("Graph user {UserId} not found", userId);
            return Result<GraphUser>.NotFound($"User {userId} not found");
        }
        catch (ODataError ex) when (ex.ResponseStatusCode == 429)
        {
            logger.LogWarning("Graph API throttled for user {UserId}", userId);
            return Result<GraphUser>.Error("Rate limited. Please retry later.");
        }
        catch (ODataError ex)
        {
            logger.LogError(ex, "Graph API error for user {UserId}", userId);
            return Result<GraphUser>.Error($"Graph API error: {ex.Message}");
        }
    }

    public async Task<Result<IReadOnlyList<GraphUser>>> GetGroupMembersAsync(
        string groupId, CancellationToken ct)
    {
        Guard.Against.NullOrWhiteSpace(groupId, nameof(groupId));

        try
        {
            var members = new List<GraphUser>();
            var response = await graphClient.Groups[groupId].Members.GetAsync(
                config => config.QueryParameters.Top = 100, ct);

            if (response?.Value is null)
                return Result<IReadOnlyList<GraphUser>>.Success([]);

            // 페이지네이션 처리
            var pageIterator = PageIterator<DirectoryObject, DirectoryObjectCollectionResponse>
                .CreatePageIterator(graphClient, response, item =>
                {
                    if (item is User user)
                        members.Add(user.Adapt<GraphUser>());
                    return true;
                });

            await pageIterator.IterateAsync(ct);

            return Result<IReadOnlyList<GraphUser>>.Success(members.AsReadOnly());
        }
        catch (ODataError ex)
        {
            logger.LogError(ex, "Failed to get group {GroupId} members", groupId);
            return Result<IReadOnlyList<GraphUser>>.Error(ex.Message);
        }
    }
}
```

### Batch Requests

```csharp
// 여러 Graph 요청을 하나의 배치로 실행
public async Task<Result<BatchResult>> BatchGetUsersAsync(
    IReadOnlyList<string> userIds, CancellationToken ct)
{
    var batchContent = new BatchRequestContentCollection(graphClient);

    foreach (var userId in userIds)
    {
        var request = graphClient.Users[userId].ToGetRequestInformation(config =>
            config.QueryParameters.Select = ["id", "displayName", "mail"]);

        await batchContent.AddBatchRequestStepAsync(request);
    }

    var batchResponse = await graphClient.Batch.PostAsync(batchContent, ct);
    // 응답 처리...

    return Result<BatchResult>.Success(new BatchResult(/* ... */));
}
```

---

## Advanced Topics

### Delegated vs Application Permissions

```csharp
// Application 권한 (앱 자체로 인증, 사용자 컨텍스트 없음)
// - 백그라운드 서비스, 데몬
// - User.Read.All, Group.Read.All 등
var credential = new ClientSecretCredential(tenantId, clientId, clientSecret);

// Delegated 권한 (사용자 대신 인증, 사용자 컨텍스트 있음)
// - 웹 앱, API (사용자 로그인 후)
// - User.Read, Mail.Send 등
var credential = new OnBehalfOfCredential(tenantId, clientId, clientSecret, userAssertion);
```

### Throttling Handling

```csharp
// 재시도 정책 (Graph SDK 내장 + Polly 보강)
public class GraphRetryHandler(ILogger logger) : DelegatingHandler
{
    protected override async Task<HttpResponseMessage> SendAsync(
        HttpRequestMessage request, CancellationToken ct)
    {
        HttpResponseMessage response;
        var retryCount = 0;
        const int maxRetries = 3;

        do
        {
            response = await base.SendAsync(request, ct);

            if (response.StatusCode != HttpStatusCode.TooManyRequests)
                break;

            if (response.Headers.RetryAfter?.Delta is { } delay)
            {
                logger.LogWarning("Throttled, retrying after {Delay}s", delay.TotalSeconds);
                await Task.Delay(delay, ct);
            }

            retryCount++;
        } while (retryCount < maxRetries);

        return response;
    }
}
```

### Change Notifications (Webhooks)

```csharp
// Graph 변경 알림 구독
public async Task SubscribeToUserChangesAsync(CancellationToken ct)
{
    var subscription = new Subscription
    {
        ChangeType = "updated",
        NotificationUrl = "https://myapp.com/api/graph/notifications",
        Resource = "users",
        ExpirationDateTime = DateTimeOffset.UtcNow.AddDays(2),
        ClientState = Guid.CreateVersion7().ToString()
    };

    await graphClient.Subscriptions.PostAsync(subscription, cancellationToken: ct);
}
```
