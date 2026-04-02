---
name: ae-lang-csharp
description: >
  Angeleyes-specific C# development specialist - Wolverine 3.x CQRS, Mapster 7.x mapping,
  AwesomeAssertions testing, NSubstitute mocking, Clean Architecture 4-layer,
  Rich Domain Model, enterprise development guide targeting .NET 10 / C# 13-14.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "language"
  status: "active"
  updated: "2026-04-02"
  modularized: "true"
  tags: "language, csharp, dotnet, wolverine, mapster, clean-architecture, ddd"
  context7-libraries: "/dotnet/aspnetcore, /dotnet/efcore, /dotnet/runtime, /wolverine, /mapster"
  author: "Angeleyes"
  extends: "moai-lang-csharp"
  target_framework: "net10.0"
  language_version: "C# 13/14"
  related-skills: "moai-lang-csharp"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["C#", "Csharp", ".NET", "ASP.NET", "Entity Framework", "Blazor", "Wolverine", "Mapster", "AwesomeAssertions", ".cs", ".csproj", ".sln", "dotnet"]
  languages: ["csharp", "c#"]
---

# Angeleyes C# Development Specialist

Wolverine 3.x / Mapster 7.x / AwesomeAssertions 기반 엔터프라이즈 C# 개발 가이드.
.NET 10 (Preview) / C# 13-14 타겟. Clean Architecture 4-layer + Rich Domain Model.

## Quick Reference

Auto-Triggers: `.cs`, `.csproj`, `.sln`, `.razor`, `appsettings*.json`, `Program.cs`

Core Stack:

| 영역 | 패키지 | 버전 |
|------|--------|------|
| CQRS/Messaging | Wolverine | 3.x |
| Object Mapping | Mapster | 7.x |
| Test Assertions | AwesomeAssertions | latest |
| Mocking | NSubstitute | latest |
| Guard Clauses | Ardalis.GuardClauses | latest |
| Logging | Serilog | latest |
| Runtime | .NET 10 (Preview) | net10.0 |
| Language | C# 13/14 (Preview) | latest |
| Test Framework | xUnit v3 (기존) / TUnit (신규 권장) | latest |
| Integration Test | Testcontainers + Respawn | latest |
| Architecture Test | NetArchTest | latest |

Quick Commands:

- 새 Web API: `dotnet new webapi -n MyApp.Web --framework net10.0`
- Wolverine 추가: `dotnet add package WolverineFx`
- Mapster 추가: `dotnet add package Mapster`
- AwesomeAssertions 추가: `dotnet add package AwesomeAssertions`
- NSubstitute 추가: `dotnet add package NSubstitute`

---

## CRITICAL: Banned Packages

다음 패키지는 상용 라이선스 전환으로 인해 **절대 사용 금지**:

| 금지 패키지 | 사유 | 대체 패키지 |
|---|---|---|
| **MediatR** | 상용 라이선스 전환 | **Wolverine 3.x** |
| **AutoMapper** | 상용 라이선스 전환 | **Mapster 7.x** |
| **FluentAssertions** | 라이선스 문제 | **AwesomeAssertions** |

코드 블록에서 금지 패키지 API를 사용하면 안 됩니다.
상세 규칙: [references/banned-packages.md](references/banned-packages.md)
대체 패턴: [references/package-alternatives.md](references/package-alternatives.md)

---

## Module Index

### Platform

- [.NET Platform](modules/dotnet-platform.md) - .NET 10 / C# 13-14 플랫폼 기초
- [Project Templates](modules/project-templates.md) - 솔루션 구조 및 프로젝트 템플릿

### Architecture

- [Clean Architecture](modules/clean-architecture.md) - Clean Architecture 4-layer 구조
- [Rich Domain Modeling](modules/rich-domain-modeling.md) - Factory Method, Value Object, Entity
- [Aggregate Patterns](modules/aggregate-patterns.md) - Aggregate Root, 경계, 불변식

### CQRS / Messaging

- [Wolverine CQRS](modules/wolverine-cqrs.md) - POCO handler, Command/Query/Event
- [Wolverine Middleware](modules/wolverine-middleware.md) - 미들웨어 체인 패턴
- [Domain Events](modules/domain-events.md) - Publish-after-save, Wolverine 연동

### Data Access

- [EF Core Conventions](modules/efcore-conventions.md) - 컨벤션, 설정, 마이그레이션
- [EF Core Advanced](modules/efcore-advanced.md) - Interceptor, Value Converter
- [Soft Delete Filters](modules/soft-delete-filters.md) - Global Query Filter 패턴

### Mapping

- [Mapster Advanced](modules/mapster-advanced.md) - IRegister, TypeAdapterConfig, Projection

### Web

- [ASP.NET + Blazor](modules/aspnet-blazor.md) - Minimal API, Blazor, 미들웨어

### Infrastructure

- [Service Abstractions](modules/service-abstractions.md) - DI, Options, Factory 패턴
- [Graph API Integration](modules/graph-api-integration.md) - Microsoft Graph SDK
- [Background Workers](modules/background-workers.md) - BackgroundService, Wolverine Scheduled

### Testing

- [Testing Strategy](modules/testing-strategy.md) - 피라미드, 프레임워크 비교, AAA 패턴
- [Testing Wolverine](modules/testing-wolverine.md) - 핸들러 테스팅, Tracked Sessions
- [Testing Infrastructure](modules/testing-infrastructure.md) - Testcontainers, NetArchTest

---

## References

- [Banned Packages](references/banned-packages.md) - 금지 패키지 규칙
- [Package Alternatives](references/package-alternatives.md) - 대체 패키지 대응표
