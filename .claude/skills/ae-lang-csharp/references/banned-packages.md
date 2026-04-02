---
module: banned-packages
version: "1.0.0"
last_updated: "2026-04-02"
category: reference
---

# Banned Packages Constitution (금지 패키지 규칙)

이 문서는 ae-lang-csharp 스킬에서 **절대 사용 금지**인 패키지를 정의한다.
코드 블록 내에서 이 패키지의 API가 단 하나라도 포함되면 안 된다.

## 금지 사유

세 패키지 모두 **상용 라이선스 전환** 또는 **라이선스 문제**로 인해 금지되었다.
프로젝트에 직접적인 법적/비용 리스크를 야기하므로 **제로 톨러런스** 정책을 적용한다.

---

## MediatR (금지)

**사유**: 상용 라이선스 전환
**대체**: Wolverine 3.x

### 금지 API 목록

- `using MediatR`
- `IRequest`, `IRequest<T>`
- `IRequestHandler<TRequest, TResponse>`
- `IMediator`, `ISender`
- `INotification`, `INotificationHandler<T>`
- `services.AddMediatR()`

### 대체 패턴

| 금지 패턴 | Wolverine 대체 |
|---|---|
| `IRequest<TResponse>` | POCO record (인터페이스 불필요) |
| `IRequestHandler<TReq, TRes>` | `static Task<TRes> Handle(TReq, ...)` |
| `IMediator.Send()` | `IMessageBus.InvokeAsync<T>()` |
| `INotification` | POCO record event |
| `INotificationHandler<T>` | `static Task Handle(TEvent, ...)` |
| `services.AddMediatR()` | `host.UseWolverine()` |

---

## AutoMapper (금지)

**사유**: 상용 라이선스 전환
**대체**: Mapster 7.x

### 금지 API 목록

- `using AutoMapper`
- `IMapper` (AutoMapper context)
- `CreateMap<TSource, TDestination>()`
- `Profile` (AutoMapper mapping profile)
- `ForMember()`, `MapFrom()` (AutoMapper context)
- `services.AddAutoMapper()`

### 대체 패턴

| 금지 패턴 | Mapster 대체 |
|---|---|
| `IMapper.Map<T>()` | `source.Adapt<T>()` |
| `CreateMap<S,D>()` | `IMapFrom<T>.ConfigureMapping()` |
| `Profile` (AutoMapper) | `IRegister` (Mapster) |
| `ForMember().MapFrom()` | `TypeAdapterConfig` fluent API |
| `services.AddAutoMapper()` | `services.AddMapster()` |

---

## FluentAssertions (금지)

**사유**: 라이선스 문제
**대체**: AwesomeAssertions (drop-in replacement)

### 금지 API 목록

- `using FluentAssertions`
- `using FluentAssertions.*` (모든 하위 네임스페이스)

### 대체 패턴

| 금지 패턴 | 대체 패턴 |
|---|---|
| `using FluentAssertions` | `using AwesomeAssertions` |

AwesomeAssertions는 FluentAssertions의 drop-in replacement로, `.Should()` 등 동일한 API를 제공한다.

---

## Moq (권장 대체)

**사유**: SponsorLink 논란 (2023), 커뮤니티 신뢰 하락
**대체**: NSubstitute (신규 프로젝트 권장)

Moq는 기존 프로젝트에서 허용하되, 신규 프로젝트에서는 NSubstitute를 사용한다.

| 금지 패턴 (신규 프로젝트) | NSubstitute 대체 |
|---|---|
| `new Mock<T>()` | `Substitute.For<T>()` |
| `mock.Setup(x => x.Method())` | `sub.Method().Returns(value)` |
| `mock.Verify(x => x.Method())` | `sub.Received().Method()` |
| `mock.Object` | (직접 사용, .Object 불필요) |

---

## InMemory DB (권장 대체)

**사유**: InMemory DB는 실제 DB 동작과 차이가 크며, 프로덕션 마이그레이션 실패 위험
**대체**: Testcontainers + Respawn

| 금지 패턴 (신규 프로젝트) | 대체 패턴 |
|---|---|
| `UseInMemoryDatabase()` | `Testcontainers.MsSql` / `Testcontainers.PostgreSql` |
