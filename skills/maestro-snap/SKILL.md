---
name: maestro-snap
description: Snapshot a service's technical surface area into living markdown docs. Document HTTP, WebSocket, gRPC, and database contracts by orienting from the project manifest, then using direct contract artifacts when present. Writes .maestro/maestro-interface/api-contracts.md, db-schema.md, and a minimal AGENTS.md at the repo root pointing agents at .maestro/ before they scan code. Use when a user wants living interface docs, API contract snapshots, or runs /maestro-snap.
---

# Maestro Snap

Generate living markdown files that capture the service surface area, plus a minimal `AGENTS.md` at the repo root that points agents at `.maestro/` before they scan code.

**Output:**
- `.maestro/maestro-interface/api-contracts.md` — HTTP, WebSocket, and gRPC contracts
- `.maestro/maestro-interface/db-schema.md` — database tables, columns, relationships, indexes
- `AGENTS.md` — single navigation rule: check `.maestro/` first

## Workflow

### Step 1 — Orient

Read the project manifest first to identify the stack and runtime shape. Run in parallel:

```bash
# Project manifest
cat go.mod 2>/dev/null || cat package.json 2>/dev/null || cat Cargo.toml 2>/dev/null || cat build.gradle.kts 2>/dev/null || cat pom.xml 2>/dev/null | head -40

# Package metadata when present
cat package.json 2>/dev/null | head -80
```

Use the manifest and built-in knowledge of the detected stack to infer where routes, handlers, models, migrations, and config usually live. Do not rely on hardcoded framework grep patterns.

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
- If no direct contract artifact exists, read the code that defines routes, handlers, models, and migrations.

### Step 4 — Write Output

```bash
mkdir -p .maestro/maestro-interface
```

Write both files using the exact structure in [OUTPUT-TEMPLATES.md](OUTPUT-TEMPLATES.md).

Rules:
- Regenerate fully each run; never append to existing files.
- Every section has a slug anchor for deep-linking.
- Include `file:line` source references for traceability.
- Mark auth on every HTTP endpoint and websocket entry with the clearest available signal.
- Use `🔒 (inferred)` when auth cannot be determined with certainty from explicit configuration.
- If a DTO cannot be resolved, write `(type unresolved)` rather than omitting the row.
- Sort HTTP endpoints by path then method. Sort entities alphabetically.

### Step 5 — Write AGENTS.md

Create `AGENTS.md` at the repository root if it does not already exist. If it already exists, leave it untouched.

The file must be minimal and purely navigational:

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
- Any gaps, such as unresolved DTO types or inferred auth markers
