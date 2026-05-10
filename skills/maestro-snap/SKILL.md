---
name: maestro-snap
description: Snapshot a Kotlin repository's full technical surface area into living markdown docs. Scans and documents HTTP/REST endpoints, WebSocket events, gRPC services (with request/response DTOs), and database schemas (entities, columns, relations, indexes). Targets Kotlin repos using Spring Boot, Ktor, JPA/Hibernate, Exposed, and Flyway/Liquibase. Writes .maestro/maestro-interface/api-contracts.md, db-schema.md, and a minimal AGENTS.md at the repo root pointing agents at .maestro/ before they scan code. Use when user wants to document a service, create living interface docs, snapshot API contracts, or runs /maestro-snap.
---

# Maestro Snap

Generate living-document markdown files that capture the complete technical surface of this service, plus a minimal `AGENTS.md` at the repo root that points agents at `.maestro/` before they scan code.

**Output:**
- `.maestro/maestro-interface/api-contracts.md` — HTTP, WebSocket, gRPC contracts + DTOs
- `.maestro/maestro-interface/db-schema.md` — entities, columns, relationships, indexes
- `AGENTS.md` — single navigation rule: check `.maestro/` first

## Workflow

### Step 1 — Orient

Identify the service name, framework, and ORM in use. Run in parallel:

```bash
# Service name
grep -m1 "rootProject.name\|artifactId" settings.gradle.kts settings.gradle pom.xml 2>/dev/null | head -3

# Framework
grep -r "spring-boot\|ktor\|micronaut" build.gradle.kts build.gradle pom.xml 2>/dev/null | head -5

# ORM / DB layer
grep -r "spring-data-jpa\|hibernate\|exposed\|jooq\|r2dbc" build.gradle.kts build.gradle pom.xml 2>/dev/null | head -5

# Migration tool
find . \( -name "*.sql" -path "*/flyway/*" -o -name "*.sql" -path "*/migration/*" -o -name "*.xml" -path "*/liquibase/*" \) \
  -not -path "*/node_modules/*" | head -10
```

### Step 2 — Discover (run all in parallel)

```bash
# HTTP — Spring Boot controllers
grep -rn "@RestController\|@Controller\b" --include="*.kt" -l | grep -v "/test/"

# HTTP — Ktor routing
grep -rn "fun Route\.\|routing\s*{\|route(\|get(\|post(\|put(\|delete(\|patch(" \
  --include="*.kt" -l | grep -v "/test/"

# WebSocket — Spring
grep -rn "@MessageMapping\|@SubscribeMapping\|@EnableWebSocket" \
  --include="*.kt" -l | grep -v "/test/"

# WebSocket — Ktor
grep -rn "webSocket(\|WebSocket" --include="*.kt" -l | grep -v "/test/"

# gRPC — proto files
find . -name "*.proto" -not -path "*/build/*"

# gRPC — Kotlin stubs / service impls
grep -rn "GrpcService\|BindableService\|ImplBase\|@GrpcMethod" \
  --include="*.kt" -l | grep -v "/test/"

# JPA entities
grep -rn "@Entity\b" --include="*.kt" -l | grep -v "/test/"

# Exposed tables
grep -rn "object .* : Table\(\|: IntIdTable\|: LongIdTable\|: UUIDTable" \
  --include="*.kt" -l | grep -v "/test/"

# Flyway migrations
find . -name "V*.sql" -path "*/migration*" -not -path "*/build/*" | sort

# Liquibase
find . \( -name "*.xml" -o -name "*.yaml" -o -name "*.yml" \) \
  -path "*/liquibase/*" -not -path "*/build/*"

# DTOs / request-response data classes
grep -rn "data class .* {" --include="*.kt" -l | grep -v "/test/"
```

### Step 3 — Read and extract

Read every discovered file. For each, extract what is documented in [SCAN-PATTERNS.md](SCAN-PATTERNS.md).

- Read DTO/data class files **before** writing output so you can inline field tables.
- For gRPC, read `.proto` files directly — they are the source of truth.
- For Flyway, read each migration SQL file in version order to catch columns not in entities.

### Step 4 — Write output

```bash
mkdir -p .maestro/maestro-interface
```

Write both files using the exact structure in [OUTPUT-TEMPLATES.md](OUTPUT-TEMPLATES.md).

Rules:
- Regenerate fully each run — never append to existing files.
- Every section has a slug anchor for deep-linking.
- Include `file:line` source references for traceability.
- Mark auth on every HTTP endpoint: `public`, `🔒 JWT`, `🔒 OAuth2`, `🔒 Role: ADMIN`, etc.
- If a DTO cannot be resolved, write `(type unresolved)` rather than omitting the row.
- Sort HTTP endpoints by path then method. Sort entities alphabetically.

### Step 5 — Write AGENTS.md

Create `AGENTS.md` at the repository root if it does not already exist. If it already exists, leave it untouched.

The file must be minimal — no service-specific details, no domain knowledge, no snapshot data. It is purely a navigation rule for agents:

```markdown
# AGENTS.md

Before scanning the repository for any assigned task, read the files in `.maestro/` first:

- `.maestro/maestro-interface/api-contracts.md` — HTTP endpoints, WebSocket channels, and gRPC methods
- `.maestro/maestro-interface/db-schema.md` — database tables, columns, and relationships

These files map the full surface of the service and include source file references. Use them to orient before exploring the codebase.
```

### Step 6 — Report

Print a brief summary:
- Files written and line counts
- Counts: N HTTP endpoints, N WS events, N gRPC methods, N entities
- Any gaps (e.g. "2 controllers had unresolved DTO types — marked in output")
