---
name: maestro-snap
description: Snapshot a service's technical surface area into living markdown docs. Document HTTP, WebSocket, gRPC, and database contracts by orienting from the project manifest, then using direct contract artifacts when present. Supports single-service repos and monorepos, writing per-service `.maestro/<service-name>/maestro-interface/` output when needed, plus a root `AGENTS.md` that points agents at `.maestro/` before they scan code. Use when a user wants living interface docs, API contract snapshots, or runs /maestro-snap.
---

# Maestro Snap

Generate living markdown files that capture the service surface area, plus a root `AGENTS.md` that points agents at `.maestro/` before they scan code.

Invocation:
- `/maestro-snap` snaps the detected service in a single-service repo.
- `/maestro-snap` snaps every detected service at a monorepo root.
- `/maestro-snap [service-name]` limits the run to one named service.

**Output:**
- Single-service repos:
  - `.maestro/maestro-interface/api-contracts.md` — HTTP, WebSocket, and gRPC contracts
  - `.maestro/maestro-interface/db-schema.md` — database tables, columns, relationships, indexes
- Monorepos:
  - `.maestro/<service-name>/maestro-interface/api-contracts.md`
  - `.maestro/<service-name>/maestro-interface/db-schema.md`
- `AGENTS.md` — single navigation rule: check `.maestro/` first, with a snapped-services list at the root

## Workflow

### Step 1 - Orient

Read the project manifest first to identify the stack and runtime shape. Then determine whether the repo is a single service or a monorepo by checking for multiple first-level service directories that each contain their own manifest. Run in parallel:

```bash
# Project manifest
{
  cat go.mod 2>/dev/null || cat package.json 2>/dev/null || cat Cargo.toml 2>/dev/null || cat build.gradle.kts 2>/dev/null || cat pom.xml 2>/dev/null
} | head -40

# Package metadata when present
cat package.json 2>/dev/null | head -80
```

Use the manifest and built-in knowledge of the detected stack to infer where routes, handlers, models, migrations, and config usually live. Do not rely on hardcoded framework grep patterns.

### Step 1.5 - Resolve Scope

- If the repo has a single service, write output to `.maestro/maestro-interface/`.
- If the repo is a monorepo, identify every service directory that has its own manifest.
- If a service name argument is provided, limit the run to that named service.
- If no service name argument is provided at a monorepo root, snap every detected service.
- Use the service directory name as the output directory name under `.maestro/`.

### Step 2 — Contract Artifact Fast Path

Run these in parallel before broader code reading:

```bash
# gRPC source of truth
find . -name "*.proto" -not -path "*/build/*" -not -path "*/node_modules/*"

# HTTP contract source of truth
find . \( -name "openapi.yml" -o -name "openapi.yaml" -o -name "openapi.json" \
  -o -name "swagger.yml" -o -name "swagger.yaml" -o -name "swagger.json" \) \
  -not -path "*/node_modules/*" -not -path "*/build/*"
```

- Proto files found → read them directly as the source of truth for gRPC contracts.
- OpenAPI/Swagger spec found → read it directly as the source of truth for HTTP contracts.
- Neither found → continue to the broader code-reading pass.

### Step 3 — Read and Extract

Read every discovered file. Use the detected stack from Step 1 to understand where to look next.

- Prefer direct contract artifacts over inferred endpoints when they exist.
- Read proto files before writing output so service and message definitions are authoritative.
- If no direct contract artifact exists, read the code that defines routes, handlers, models, and migrations:
  - Open the server bootstrap, router setup, controllers, handlers, and adjacent tests.
  - Follow imports and registration calls to find the code paths that define routes or RPCs.
  - Infer path params, query params, request bodies, response bodies, and auth from the code and tests.
  - Find migration directories by name heuristics: `migrations/`, `db/migrate/`, `database/migrations/`, `schema/migrations/`, `resources/db/migration/`, and similar. Read raw SQL migration files in version order regardless of migration tool.
  - Cross-reference ORM or schema model files for relationships and type information that SQL does not make explicit.

### Step 4 - Write Output

```bash
mkdir -p .maestro/<service-name>/maestro-interface  # use .maestro/maestro-interface/ for a single-service repo
```

Write both files using the exact structure in [OUTPUT-TEMPLATES.md](OUTPUT-TEMPLATES.md).

Rules:
- Regenerate fully each run; never append to existing files.
- Every section has a slug anchor for deep-linking.
- Include `file:line` source references for traceability.
- Mark auth on every HTTP endpoint and websocket entry with the clearest available signal: `public`, `🔒 JWT`, `🔒 OAuth2`, `🔒 API Key`, `🔒 Role: ADMIN`, or `🔒 (inferred)` when certainty is low.
- If a DTO cannot be resolved, write `(type unresolved)` rather than omitting the row.
- Sort HTTP endpoints by path then method. Sort entities alphabetically.
- For single-service repos, write the files under `.maestro/maestro-interface/`.
- For monorepos, write each service under `.maestro/<service-name>/maestro-interface/`.

### Step 5 - Write AGENTS.md

Create `AGENTS.md` at the repository root if it does not already exist. If it already exists, update it in place so the snapped-services list stays current.

The file must be minimal and purely navigational, but it must list all snapped services and their `.maestro/<service-name>/` paths:

```markdown
# AGENTS.md

Before scanning the repository for any assigned task, read the files in `.maestro/` first:

- `.maestro/maestro-interface/api-contracts.md` — HTTP endpoints, WebSocket channels, and gRPC methods
- `.maestro/maestro-interface/db-schema.md` — database tables, columns, and relationships
- `.maestro/<service-name>/maestro-interface/api-contracts.md` — monorepo service path
- `.maestro/<service-name>/maestro-interface/db-schema.md` — monorepo service path

These files map the full surface of the service and include source file references. Use them to orient before exploring the codebase.

Snapped services:

- `<service-name>` → `.maestro/<service-name>/`
```

### Step 6 - Report

Print a brief summary:
- Files written and line counts
- Counts: N HTTP endpoints, N WS events, N gRPC methods, N entities
- Any gaps, such as unresolved DTO types or inferred auth markers
