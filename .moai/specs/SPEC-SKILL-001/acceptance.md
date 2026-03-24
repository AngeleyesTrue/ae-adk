---
spec_id: SPEC-SKILL-001
type: acceptance
version: "1.0.0"
---

# SPEC-SKILL-001 수락 기준

## 1. 금지 패키지 제거 검증

### AC-BAN-001: MediatR 금지 검증

**조건**: `.claude/skills/ae-lang-csharp/` 내 모든 파일의 코드 블록(` ``` ` 내부)에서 MediatR 관련 API가 존재하지 않아야 한다.

**검증 대상 패턴**:
- `using MediatR`
- `IRequest<`
- `IRequestHandler<`
- `IMediator`
- `ISender`
- `INotification`
- `INotificationHandler<`
- `AddMediatR`

**검증 방법**:
```
Grep: pattern="IRequest<|IRequestHandler<|IMediator|ISender|INotification[^s]|INotificationHandler<|AddMediatR|using MediatR"
      path=".claude/skills/ae-lang-csharp/"
      output_mode="count"
```

**합격 기준**: 일치 건수 = 0

**예외**: 주석 또는 본문 텍스트에서 "MediatR 대신 Wolverine 사용" 등 금지 맥락 설명은 허용한다. 단, 코드 블록 내부에서는 절대 불허.

---

### AC-BAN-002: AutoMapper 금지 검증

**조건**: 코드 블록에서 AutoMapper 관련 API가 존재하지 않아야 한다.

**검증 대상 패턴**:
- `using AutoMapper`
- `IMapper` (AutoMapper context - Mapster의 `IMapper`와 구분)
- `CreateMap<`
- `ForMember(`
- `MapFrom(` (AutoMapper context)
- `AddAutoMapper`
- `Profile` (AutoMapper mapping profile context)

**검증 방법**:
```
Grep: pattern="using AutoMapper|AddAutoMapper|CreateMap<"
      path=".claude/skills/ae-lang-csharp/"
      output_mode="count"
```

**합격 기준**: 일치 건수 = 0

**참고**: `ForMember`와 `MapFrom`은 Mapster에서도 사용 가능하므로, AutoMapper context인지 확인이 필요하다. `using AutoMapper`와 `CreateMap<` 검색으로 1차 필터링한다.

---

### AC-BAN-003: FluentAssertions 금지 검증

**조건**: 코드 블록에서 FluentAssertions 네임스페이스가 존재하지 않아야 한다.

**검증 대상 패턴**:
- `using FluentAssertions`

**검증 방법**:
```
Grep: pattern="using FluentAssertions"
      path=".claude/skills/ae-lang-csharp/"
      output_mode="count"
```

**합격 기준**: 일치 건수 = 0

---

## 2. 스킬 구조 검증

### AC-STRUCT-001: 디렉토리 구조 검증

**조건**: 아래 디렉토리 구조가 존재해야 한다.

```
.claude/skills/ae-lang-csharp/
├── SKILL.md
├── modules/
│   └── (16 .md files)
└── references/
    ├── banned-packages.md
    └── package-alternatives.md
```

**검증 방법**:
```
Glob: pattern=".claude/skills/ae-lang-csharp/**/*.md"
```

**합격 기준**: 총 19개 .md 파일이 존재한다.

---

### AC-STRUCT-002: YAML Frontmatter 검증

**조건**: 모든 모듈 파일(`modules/*.md`)에 YAML frontmatter가 포함되어야 한다.

**필수 필드**:
- `module`: 모듈 식별자
- `version`: 버전 문자열
- `last_updated`: 날짜 문자열
- `category`: 카테고리 (platform / architecture / cqrs / data / web / ddd / infrastructure)

**검증 방법**:
```
Grep: pattern="^---"
      path=".claude/skills/ae-lang-csharp/modules/"
      output_mode="count"
```

각 모듈 파일을 `Read`로 열어 frontmatter 필드 존재를 확인한다.

**합격 기준**: 16개 모듈 파일 모두 4개 필수 필드를 포함한다.

---

### AC-STRUCT-003: Progressive Disclosure 3단계 검증

**조건**: 모든 모듈 파일이 Progressive Disclosure 3단계 구조를 포함해야 한다.

**필수 섹션 헤딩**:
1. `Quick Reference` (또는 `빠른 참조`)
2. `Detailed Patterns` (또는 `상세 패턴`)
3. `Advanced Topics` (또는 `고급 주제`)

**검증 방법**:
```
Grep: pattern="Quick Reference|빠른 참조"
      path=".claude/skills/ae-lang-csharp/modules/"
      output_mode="count"
```

**합격 기준**: 각 검색에서 16건 (모듈당 1건).

---

## 3. 핵심 기능 검증

### AC-CORE-001: Wolverine CQRS 패턴 검증

**조건**: `wolverine-cqrs.md`에 Wolverine POCO handler 패턴이 포함되어야 한다.

**필수 포함 내용**:
- Command handler 예시 (POCO record + static handler method)
- Query handler 예시
- `IMessageBus` 사용 예시
- `CancellationToken` 포함

**검증 방법**:
```
Read: file_path=".claude/skills/ae-lang-csharp/modules/wolverine-cqrs.md"
```

**합격 기준**: POCO handler 패턴이 코드 블록으로 포함되고, MediatR API가 사용되지 않는다.

---

### AC-CORE-002: Mapster 매핑 패턴 검증

**조건**: `mapster-advanced.md`에 Mapster 매핑 패턴이 포함되어야 한다.

**필수 포함 내용**:
- `IMapFrom<T>` 인터페이스 구현 예시
- `Adapt<T>()` extension method 사용
- `TypeAdapterConfig` 설정
- `ProjectToType<T>()` EF Core projection

**검증 방법**:
```
Grep: pattern="IMapFrom<|Adapt<|TypeAdapterConfig|ProjectToType<"
      path=".claude/skills/ae-lang-csharp/modules/mapster-advanced.md"
      output_mode="count"
```

**합격 기준**: 4개 패턴 모두 1건 이상 일치.

---

### AC-CORE-003: AwesomeAssertions 패턴 검증

**조건**: 테스트 관련 코드 예시에서 AwesomeAssertions를 사용해야 한다.

**검증 방법**:
```
Grep: pattern="using AwesomeAssertions|\.Should\(\)"
      path=".claude/skills/ae-lang-csharp/"
      output_mode="count"
```

**합격 기준**: 1건 이상 일치. FluentAssertions는 0건.

---

### AC-CORE-004: Rich Domain Model 검증

**조건**: `rich-domain-modeling.md`에 Rich Domain Model 패턴이 포함되어야 한다.

**필수 포함 내용**:
- Factory Method 패턴 (static Create method)
- Value Object 패턴 (GetEqualityComponents)
- Entity 패턴 (private setter, 상태 전이 메서드)
- `Result<T>` 반환 패턴
- `Guard.Against` 사용

**검증 방법**:
```
Read: file_path=".claude/skills/ae-lang-csharp/modules/rich-domain-modeling.md"
```

**합격 기준**: Factory Method, Value Object, Entity 패턴이 각각 코드 블록으로 포함된다.

---

### AC-CORE-005: Domain Events 검증

**조건**: `domain-events.md`에 도메인 이벤트 패턴이 포함되어야 한다.

**필수 포함 내용**:
- Publish-after-save 패턴 (SaveChanges 이후 발행)
- Wolverine `IMessageBus.PublishAsync()` 사용
- `SaveChangesInterceptor` 또는 동등한 메커니즘
- Event handler 예시

**검증 방법**:
```
Grep: pattern="PublishAsync|SaveChangesInterceptor|DomainEvent"
      path=".claude/skills/ae-lang-csharp/modules/domain-events.md"
      output_mode="count"
```

**합격 기준**: 3개 패턴 모두 1건 이상 일치.

---

## 4. 마이그레이션 모듈 검증

### AC-MIG-001: 5개 모듈 존재 검증

**조건**: spk-dotnet-guide에서 마이그레이션한 5개 모듈이 존재해야 한다.

**대상 파일**:
1. `modules/dotnet-platform.md`
2. `modules/clean-architecture.md`
3. `modules/wolverine-cqrs.md`
4. `modules/efcore-conventions.md`
5. `modules/aspnet-blazor.md`

**검증 방법**:
```
Glob: pattern=".claude/skills/ae-lang-csharp/modules/{dotnet-platform,clean-architecture,wolverine-cqrs,efcore-conventions,aspnet-blazor}.md"
```

**합격 기준**: 5개 파일 모두 존재.

---

### AC-MIG-002: 마이그레이션 형식 검증

**조건**: 마이그레이션된 모듈이 ae-lang-csharp 형식을 따라야 한다.

**필수 조건**:
- YAML frontmatter 포함 (AC-STRUCT-002 기준)
- Progressive Disclosure 3단계 (AC-STRUCT-003 기준)
- 금지 패키지 미사용 (AC-BAN-001~003 기준)

**합격 기준**: AC-STRUCT-002, AC-STRUCT-003, AC-BAN-001~003 기준을 모두 충족.

---

## 5. 신규 모듈 검증

### AC-NEW-001: Core DDD 모듈 존재 검증

**조건**: 6개 Core DDD 신규 모듈이 존재해야 한다.

**대상 파일**:
1. `modules/domain-events.md`
2. `modules/wolverine-middleware.md`
3. `modules/rich-domain-modeling.md`
4. `modules/mapster-advanced.md`
5. `modules/soft-delete-filters.md`
6. `modules/aggregate-patterns.md`

**검증 방법**:
```
Glob: pattern=".claude/skills/ae-lang-csharp/modules/{domain-events,wolverine-middleware,rich-domain-modeling,mapster-advanced,soft-delete-filters,aggregate-patterns}.md"
```

**합격 기준**: 6개 파일 모두 존재하며, 각 파일에 최소 1개 이상의 C# 코드 블록(```` ```csharp ````)이 포함된다.

---

### AC-NEW-002: Infrastructure 모듈 존재 검증

**조건**: 5개 Infrastructure 신규 모듈이 존재해야 한다.

**대상 파일**:
1. `modules/service-abstractions.md`
2. `modules/efcore-advanced.md`
3. `modules/project-templates.md`
4. `modules/graph-api-integration.md`
5. `modules/background-workers.md`

**검증 방법**:
```
Glob: pattern=".claude/skills/ae-lang-csharp/modules/{service-abstractions,efcore-advanced,project-templates,graph-api-integration,background-workers}.md"
```

**합격 기준**: 5개 파일 모두 존재하며, 각 파일에 최소 1개 이상의 C# 코드 블록이 포함된다.

---

## 6. 인프라 모듈 상세 검증

### AC-INFRA-001: 프로젝트 템플릿 및 서비스 추상화 검증

**조건**: `project-templates.md`와 `service-abstractions.md`에 실용적인 내용이 포함되어야 한다.

**project-templates.md 필수 포함**:
- Clean Architecture 4-layer 솔루션 구조
- 각 Layer별 프로젝트 참조 방향
- `.csproj` 또는 `Directory.Build.props` 예시

**service-abstractions.md 필수 포함**:
- DI 등록 패턴 (AddScoped/AddTransient/AddSingleton)
- Options pattern (`IOptions<T>`)
- Factory pattern 예시

**합격 기준**: 위 항목이 코드 블록으로 포함된다.

---

### AC-INFRA-002: EF Core 고급 검증

**조건**: `efcore-advanced.md`에 EF Core 고급 패턴이 포함되어야 한다.

**필수 포함 내용**:
- `SaveChangesInterceptor` 또는 `DbCommandInterceptor` 예시
- Value Converter (`HasConversion`) 예시
- Compiled Query (`EF.CompileAsyncQuery`) 예시 (선택)

**검증 방법**:
```
Grep: pattern="Interceptor|HasConversion|CompileAsyncQuery"
      path=".claude/skills/ae-lang-csharp/modules/efcore-advanced.md"
      output_mode="count"
```

**합격 기준**: Interceptor와 HasConversion이 각 1건 이상.

---

### AC-INFRA-003: Background Workers 검증

**조건**: `background-workers.md`에 백그라운드 작업 패턴이 포함되어야 한다.

**필수 포함 내용**:
- `IHostedService` 또는 `BackgroundService` 패턴
- Wolverine Scheduled Jobs 패턴 (선택)

**검증 방법**:
```
Grep: pattern="IHostedService|BackgroundService|ScheduledAt"
      path=".claude/skills/ae-lang-csharp/modules/background-workers.md"
      output_mode="count"
```

**합격 기준**: 1건 이상 일치.

---

### AC-INFRA-004: Graph API 통합 검증

**조건**: `graph-api-integration.md`에 Microsoft Graph API 통합 패턴이 포함되어야 한다.

**필수 포함 내용**:
- `GraphServiceClient` 사용 예시
- 인증(Authentication) 설정 패턴
- `Result<T>` 반환 패턴

**검증 방법**:
```
Grep: pattern="GraphServiceClient|ClientSecretCredential|Result<"
      path=".claude/skills/ae-lang-csharp/modules/graph-api-integration.md"
      output_mode="count"
```

**합격 기준**: GraphServiceClient가 1건 이상.

---

## 7. 대체 패키지 대응 검증

### AC-ALT-001: 대응표 완전성 검증

**조건**: `references/package-alternatives.md`에 금지 패키지와 대체 패키지의 대응표가 완전해야 한다.

**필수 대응 항목**:

| 금지 패턴 | 대체 패턴 |
|---|---|
| `IRequest<T>` | Wolverine POCO record |
| `IRequestHandler` | Wolverine static handler |
| `IMediator.Send` | `IMessageBus.InvokeAsync` |
| `IMapper.Map<T>` | `Adapt<T>()` |
| `CreateMap<S,D>` | `IMapFrom<T>` |
| `using FluentAssertions` | `using AwesomeAssertions` |

**검증 방법**:
```
Read: file_path=".claude/skills/ae-lang-csharp/references/package-alternatives.md"
```

**합격 기준**: 위 6개 대응 항목이 모두 포함된다.

---

## 8. Context7 통합 검증

### AC-CTX7-001: SKILL.md 메타데이터 검증

**조건**: `SKILL.md`의 YAML frontmatter에 Context7 호환 메타데이터가 포함되어야 한다.

**필수 필드**:
- `name`: 스킬 식별자
- `description`: 스킬 설명
- `triggers`: 파일 패턴 리스트
- `metadata.version`: 버전
- `metadata.replaces`: 대체 대상 스킬명

**검증 방법**:
```
Read: file_path=".claude/skills/ae-lang-csharp/SKILL.md"
```

**합격 기준**: 위 필드가 모두 존재하며 유효한 값을 가진다.

---

## 9. 품질 게이트

전체 스킬이 릴리스 가능한 수준인지 판단하는 최종 품질 게이트:

| 게이트 | 기준 | 검증 ID |
|---|---|---|
| 파일 수 | 19개 .md 파일 | AC-STRUCT-001 |
| 금지 패키지 | 코드 블록 내 금지 패키지 0건 | AC-BAN-001, AC-BAN-002, AC-BAN-003 |
| YAML frontmatter | 모든 모듈에 필수 필드 포함 | AC-STRUCT-002 |
| Progressive Disclosure | 모든 모듈에 3단계 구조 | AC-STRUCT-003 |
| 핵심 패턴 | Wolverine POCO, Mapster IMapFrom<T>, AwesomeAssertions, Rich Domain Model, Domain Events | AC-CORE-001~005 |
| 마이그레이션 | 5개 모듈 존재 및 형식 준수 | AC-MIG-001, AC-MIG-002 |
| 신규 모듈 | 11개 모듈 존재 및 최소 품질 | AC-NEW-001, AC-NEW-002 |
| 대응표 | 6개 필수 대응 항목 완전 | AC-ALT-001 |
| Context7 | SKILL.md 메타데이터 유효 | AC-CTX7-001 |

**최종 합격 기준**: 모든 게이트 통과.

---

## 10. 검증 도구

수락 기준 검증에 사용하는 도구:

| 도구 | 용도 |
|---|---|
| **Glob** | 파일 존재 여부 확인 (패턴 매칭으로 파일 목록 조회) |
| **Grep** | 금지 패키지 검색, 패턴 존재 확인, 건수 카운트 |
| **Read** | 파일 내용 직접 확인 (frontmatter, 코드 블록, 구조 검증) |

### 검증 실행 순서

1. `Glob`으로 19개 파일 존재 확인
2. `Grep`으로 금지 패키지 3종 검색 (0건 확인)
3. `Grep`으로 Progressive Disclosure 헤딩 검색 (16건 확인)
4. `Read`로 SKILL.md 메타데이터 확인
5. `Read`로 핵심 모듈 (wolverine-cqrs, mapster-advanced, rich-domain-modeling, domain-events) 내용 확인
6. `Read`로 references/ 파일 내용 확인
7. `Grep`으로 핵심 패턴 (IMapFrom, Adapt, PublishAsync, Guard.Against) 존재 확인
