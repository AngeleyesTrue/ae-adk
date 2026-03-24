---
id: SPEC-SKILL-001
version: "1.0.0"
status: draft
created: "2026-03-23"
updated: "2026-03-23"
author: Angeleyes
priority: high
issue_number: 0
---

# SPEC-SKILL-001: C# 스킬 고도화 및 Angeleyes 전용 커스터마이징

## 1. 현황 및 환경

### 1.1. 프로젝트 구조

ae-adk는 moai-adk의 포크 레포지토리이다. moai-adk에는 `.claude/skills/moai-lang-csharp/` 경로에 기본 C# 스킬이 존재하며, 이는 범용 C# 가이드를 제공한다. 그러나 Angeleyes의 실제 개발 환경과 기술 스택은 기본 스킬과 상당한 차이가 있어, 전용 스킬을 별도로 구축해야 한다.

### 1.2. 핵심 문제: 금지 패키지 사용

moai-lang-csharp 기본 스킬 및 일반적인 AI 코드 생성에서 빈번하게 등장하는 3개 패키지가 **상용 라이선스 전환 또는 라이선스 문제**로 인해 사용이 금지되었다:

| 금지 패키지 | 사유 | 대체 패키지 |
|---|---|---|
| **MediatR** | 상용 라이선스 전환 | **Wolverine 3.x** |
| **AutoMapper** | 상용 라이선스 전환 | **Mapster 7.x** |
| **FluentAssertions** | 라이선스 문제 | **AwesomeAssertions** |

이 금지 패키지들이 코드 생성에 포함되면 프로젝트에 직접적인 법적/비용 리스크를 야기한다. 따라서 **제로 톨러런스** 정책을 적용한다.

### 1.3. 기술 스택 불일치

| 영역 | 기본 스킬 (moai) | Angeleyes 실제 스택 |
|---|---|---|
| CQRS/Messaging | MediatR `IRequest<T>` | **Wolverine 3.x** POCO handlers |
| Object Mapping | AutoMapper `Profile` | **Mapster 7.x** `IMapFrom<T>` |
| Test Assertions | FluentAssertions | **AwesomeAssertions** (drop-in) |
| Guard Clauses | 수동 검증 | **Ardalis.GuardClauses** |
| Logging | `ILogger<T>` | **Serilog** (structured logging) |
| Target Runtime | .NET 8 | **.NET 10 Preview** (2026-11 LTS 출시 예정, 현재 .NET 9 기반 개발 후 업그레이드) |
| Error Handling | Exceptions | **Result\<T\> pattern** |
| Language Version | C# 12 | **C# 13/14** |

### 1.4. 기존 자산

**Obsidian 지식 저장소**: `D:\Obsidian\spk-dotnet-guide`에 5개 모듈이 존재한다:

1. `dotnet-platform` - .NET 플랫폼 기초
2. `clean-architecture` - Clean Architecture 4-layer 구조
3. `wolverine-cqrs` - Wolverine CQRS 패턴
4. `efcore-conventions` - EF Core 컨벤션
5. `aspnet-blazor` - ASP.NET + Blazor 패턴

이 모듈들은 이미 Angeleyes의 기술 스택으로 작성되어 있으며, ae-lang-csharp 스킬의 마이그레이션 기반이 된다.

### 1.5. 실제 프로젝트

- **SPK Cloud Archive**: M365 아카이빙 솔루션 (Exchange, SharePoint, Teams 데이터 아카이빙)
- **SPK ID Manager**: Wolverine Bus + Microsoft Graph API 기반 ID 관리 시스템

---

## 2. 가정 사항 (Assumptions)

1. **Wolverine 3.x**는 stable 릴리스 상태이며, POCO handler 패턴이 안정적으로 동작한다.
2. **Mapster 7.x**는 `IMapFrom<T>` 인터페이스를 통한 매핑 구성을 지원한다.
3. **AwesomeAssertions**는 FluentAssertions의 drop-in replacement로, 동일한 API surface를 제공한다.
4. **.NET 10 LTS**는 2026-11에 출시 예정이며, 이것이 프로덕션 타겟이다. 현재는 .NET 9 기반으로 코드를 작성하고 출시 시 업그레이드한다. C# 13/14 문법을 사용한다.
5. **Clean Architecture 4-layer** 구조를 기본으로 한다: Domain → Application → Infrastructure → Presentation.
6. **DDD Rich Domain Model**을 채택하며, Anemic Domain Model은 지양한다.
7. **Wolverine POCO handlers**가 MediatR의 `IRequest`/`IRequestHandler` 패턴을 완전히 대체한다.
8. `ae-lang-csharp` 스킬은 `moai-lang-csharp` 스킬을 **대체(replace)** 한다 (병존하지만 ae가 우선).

---

## 3. 요구사항 (Requirements)

### 3.1. 금지 패키지 규칙 (Hard, Unwanted)

| ID | 요구사항 | 유형 | 우선순위 |
|---|---|---|---|
| REQ-BAN-001 | MediatR 및 관련 API (`IRequest`, `IRequestHandler`, `IMediator`, `ISender`, `INotification`, `INotificationHandler`)를 코드 블록에서 절대 사용 금지 | Hard | Critical |
| REQ-BAN-002 | AutoMapper 및 관련 API (`IMapper`, `CreateMap`, `Profile`, `ForMember`, `MapFrom`)를 코드 블록에서 절대 사용 금지 | Hard | Critical |
| REQ-BAN-003 | FluentAssertions 네임스페이스 (`using FluentAssertions`)를 코드 블록에서 절대 사용 금지 | Hard | Critical |

### 3.2. 핵심 스킬 교체 (Hard, Functional)

| ID | 요구사항 | 유형 | 우선순위 |
|---|---|---|---|
| REQ-CORE-001 | `.claude/skills/ae-lang-csharp/SKILL.md`를 생성하여 ae-adk에서 C# 관련 작업 시 자동 로드되도록 한다 | Hard | High |
| REQ-CORE-002 | Wolverine 3.x POCO handler 패턴을 CQRS 기본 패턴으로 제공한다 (Command/Query/Event handler) | Hard | High |
| REQ-CORE-003 | Mapster 7.x `IMapFrom<T>` 패턴을 객체 매핑 기본 패턴으로 제공한다 | Hard | High |
| REQ-CORE-004 | AwesomeAssertions를 테스트 assertion 기본 라이브러리로 제공한다 | Hard | High |

### 3.3. 마이그레이션 모듈 (Hard, Functional)

| ID | 요구사항 | 유형 | 우선순위 |
|---|---|---|---|
| REQ-MIG-001 | spk-dotnet-guide의 5개 모듈을 ae-lang-csharp 형식에 맞춰 마이그레이션한다 | Hard | High |
| REQ-MIG-002 | 마이그레이션 시 기존 콘텐츠를 보존하되, YAML frontmatter와 Progressive Disclosure 형식을 적용한다 | Hard | Medium |

대상 모듈:

1. `dotnet-platform.md` - .NET 플랫폼 기초
2. `clean-architecture.md` - Clean Architecture 4-layer
3. `wolverine-cqrs.md` - Wolverine CQRS 패턴
4. `efcore-conventions.md` - EF Core 컨벤션
5. `aspnet-blazor.md` - ASP.NET + Blazor

### 3.4. 신규 모듈 (Hard, Functional)

| ID | 요구사항 | 유형 | 우선순위 |
|---|---|---|---|
| REQ-NEW-001 | `domain-events.md` - 도메인 이벤트 패턴 (publish-after-save, Wolverine integration) | Hard | High |
| REQ-NEW-002 | `wolverine-middleware.md` - Wolverine 미들웨어 체인 패턴 | Hard | High |
| REQ-NEW-003 | `rich-domain-modeling.md` - Rich Domain Model (Factory Method, Value Object, Entity) | Hard | High |
| REQ-NEW-004 | `mapster-advanced.md` - Mapster 고급 매핑 (TypeAdapterConfig, Projection) | Hard | Medium |
| REQ-NEW-005 | `soft-delete-filters.md` - Soft Delete + Global Query Filter 패턴 | Hard | Medium |
| REQ-NEW-006 | `aggregate-patterns.md` - Aggregate Root 패턴 (경계, 불변식, 트랜잭션) | Hard | Medium |
| REQ-NEW-007 | `service-abstractions.md` - 서비스 추상화 패턴 (DI, Options, Factory) | Hard | Medium |
| REQ-NEW-008 | `efcore-advanced.md` - EF Core 고급 (Interceptor, Value Converter, Compiled Query) | Hard | Medium |
| REQ-NEW-009 | `project-templates.md` - 프로젝트 템플릿 및 솔루션 구조 | Hard | Medium |
| REQ-NEW-010 | `graph-api-integration.md` - Microsoft Graph API 통합 패턴 | Hard | Low |
| REQ-NEW-011 | `background-workers.md` - Background Worker 패턴 (IHostedService, Wolverine Scheduled) | Hard | Low |

### 3.5. 품질 요구사항 (Hard, Non-Functional)

| ID | 요구사항 | 유형 | 우선순위 |
|---|---|---|---|
| REQ-QUAL-001 | 모든 모듈에 Progressive Disclosure 3단계 적용: Quick Reference → Detailed Patterns → Advanced Topics | Hard | High |
| REQ-QUAL-002 | 모든 모듈 파일에 YAML frontmatter 포함 (module, version, last_updated, category) | Hard | Medium |
| REQ-QUAL-003 | Context7 MCP 서버와의 매핑을 위한 메타데이터 구조 유지 | Hard | Medium |

### 3.6. 호환성 요구사항 (Soft)

| ID | 요구사항 | 유형 | 우선순위 |
|---|---|---|---|
| REQ-COMPAT-001 | moai-lang-csharp 파일을 직접 수정하지 않는다 (Fork-and-Replace 전략) | Soft | High |
| REQ-COMPAT-002 | ae-lang-csharp가 moai-lang-csharp보다 우선 로드되도록 trigger 조건을 설정한다 | Soft | High |
| REQ-COMPAT-003 | moai-adk 업스트림 업데이트 시 충돌이 발생하지 않도록 독립적 경로를 사용한다 | Soft | Medium |

---

## 4. 사양 (Specifications)

### 4.1. 디렉토리 구조

```
.claude/skills/ae-lang-csharp/
├── SKILL.md                          # 스킬 엔트리포인트
├── modules/
│   ├── dotnet-platform.md            # [MIG] .NET 플랫폼 기초
│   ├── clean-architecture.md         # [MIG] Clean Architecture 4-layer
│   ├── wolverine-cqrs.md             # [MIG] Wolverine CQRS 패턴
│   ├── efcore-conventions.md         # [MIG] EF Core 컨벤션
│   ├── aspnet-blazor.md              # [MIG] ASP.NET + Blazor
│   ├── domain-events.md              # [NEW] 도메인 이벤트
│   ├── wolverine-middleware.md       # [NEW] Wolverine 미들웨어
│   ├── rich-domain-modeling.md       # [NEW] Rich Domain Model
│   ├── mapster-advanced.md           # [NEW] Mapster 고급 매핑
│   ├── soft-delete-filters.md        # [NEW] Soft Delete 패턴
│   ├── aggregate-patterns.md         # [NEW] Aggregate Root 패턴
│   ├── service-abstractions.md       # [NEW] 서비스 추상화
│   ├── efcore-advanced.md            # [NEW] EF Core 고급
│   ├── project-templates.md          # [NEW] 프로젝트 템플릿
│   ├── graph-api-integration.md      # [NEW] Graph API 통합
│   └── background-workers.md         # [NEW] Background Workers
└── references/
    ├── banned-packages.md            # 금지 패키지 규칙 (Constitution)
    └── package-alternatives.md       # 대체 패키지 대응표
```

총 파일 수: **19개** (SKILL.md 1 + modules 16 + references 2)

### 4.2. SKILL.md 메타데이터 사양

SKILL.md 파일은 다음 YAML frontmatter를 포함한다:

```yaml
---
name: ae-lang-csharp
description: "Angeleyes 전용 C# 스킬 - Wolverine/Mapster/AwesomeAssertions 기반"
triggers:
  - "*.cs"
  - "*.csproj"
  - "*.sln"
  - "*.razor"
  - "appsettings*.json"
  - "Program.cs"
metadata:
  author: Angeleyes
  version: "1.0.0"
  replaces: moai-lang-csharp
  target_framework: net10.0
  language_version: "C# 13/14"
---
```

### 4.3. 금지 패키지 규칙 (Constitution)

`references/banned-packages.md`에 정의되는 금지 규칙의 상세 내용:

#### MediatR 금지 API 목록

- `using MediatR`
- `IRequest`, `IRequest<T>`
- `IRequestHandler<TRequest, TResponse>`
- `IMediator`, `ISender`
- `INotification`, `INotificationHandler<T>`
- `services.AddMediatR()`

#### AutoMapper 금지 API 목록

- `using AutoMapper`
- `IMapper`, `IMapper.Map<T>()`
- `CreateMap<TSource, TDestination>()`
- `Profile` (AutoMapper context)
- `ForMember()`, `MapFrom()`
- `services.AddAutoMapper()`

#### FluentAssertions 금지 네임스페이스

- `using FluentAssertions`
- `using FluentAssertions.*` (모든 하위 네임스페이스)

### 4.4. 대체 패키지 대응표

| 금지 패턴 | 대체 패턴 | 비고 |
|---|---|---|
| `IRequest<TResponse>` | Wolverine POCO record | 인터페이스 불필요 |
| `IRequestHandler<TReq, TRes>` | `static Task<TRes> Handle(TReq, ...)` | POCO static handler |
| `IMediator.Send()` | `IMessageBus.InvokeAsync<T>()` | Wolverine bus |
| `INotification` | Wolverine Event (POCO record) | 인터페이스 불필요 |
| `INotificationHandler<T>` | `static Task Handle(TEvent, ...)` | POCO handler |
| `IMapper.Map<T>()` | `source.Adapt<T>()` | Mapster extension |
| `CreateMap<S,D>()` | `IMapFrom<T>.ConfigureMapping()` | Mapster interface |
| `Profile` (AutoMapper) | `IRegister` (Mapster) | 매핑 등록 |
| `ForMember().MapFrom()` | `TypeAdapterConfig` fluent API | Mapster config |
| `using FluentAssertions` | `using AwesomeAssertions` | Drop-in replacement |

---

## 5. 추적성 매트릭스 (Traceability Matrix)

| 요구사항 ID | 사양 섹션 | 대상 파일 | 수락 기준 |
|---|---|---|---|
| REQ-BAN-001 | 4.3 | references/banned-packages.md | AC-BAN-001 |
| REQ-BAN-002 | 4.3 | references/banned-packages.md | AC-BAN-002 |
| REQ-BAN-003 | 4.3 | references/banned-packages.md | AC-BAN-003 |
| REQ-CORE-001 | 4.2 | SKILL.md | AC-STRUCT-001 |
| REQ-CORE-002 | 4.4 | modules/wolverine-cqrs.md | AC-CORE-001 |
| REQ-CORE-003 | 4.4 | modules/mapster-advanced.md | AC-CORE-002 |
| REQ-CORE-004 | 4.4 | modules/dotnet-platform.md | AC-CORE-003 |
| REQ-MIG-001 | 4.1 | modules/*.md (5 files) | AC-MIG-001 |
| REQ-MIG-002 | 4.1 | modules/*.md (5 files) | AC-MIG-002 |
| REQ-NEW-001 | 4.1 | modules/domain-events.md | AC-CORE-005 |
| REQ-NEW-002 | 4.1 | modules/wolverine-middleware.md | AC-NEW-001 |
| REQ-NEW-003 | 4.1 | modules/rich-domain-modeling.md | AC-CORE-004 |
| REQ-NEW-004 | 4.1 | modules/mapster-advanced.md | AC-NEW-001 |
| REQ-NEW-005 | 4.1 | modules/soft-delete-filters.md | AC-NEW-001 |
| REQ-NEW-006 | 4.1 | modules/aggregate-patterns.md | AC-NEW-001 |
| REQ-NEW-007 | 4.1 | modules/service-abstractions.md | AC-INFRA-001 |
| REQ-NEW-008 | 4.1 | modules/efcore-advanced.md | AC-INFRA-002 |
| REQ-NEW-009 | 4.1 | modules/project-templates.md | AC-INFRA-001 |
| REQ-NEW-010 | 4.1 | modules/graph-api-integration.md | AC-INFRA-004 |
| REQ-NEW-011 | 4.1 | modules/background-workers.md | AC-INFRA-003 |
| REQ-QUAL-001 | 4.1 | modules/*.md (all) | AC-STRUCT-003 |
| REQ-QUAL-002 | 4.2 | modules/*.md (all) | AC-STRUCT-002 |
| REQ-QUAL-003 | 4.2 | SKILL.md | AC-CTX7-001 |
| REQ-COMPAT-001 | 4.1 | (directory path) | AC-STRUCT-001 |
| REQ-COMPAT-002 | 4.2 | SKILL.md triggers | AC-STRUCT-001 |
| REQ-COMPAT-003 | 4.1 | (directory path) | AC-STRUCT-001 |
