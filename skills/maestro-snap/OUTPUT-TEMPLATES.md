# Output Templates

Exact markdown structure for the two generated files. Fill in extracted data; omit entire sections if none exist, such as no gRPC or no WebSocket support.

For single-service repos, write these files under `.maestro/maestro-interface/`. For monorepos, write one set per service under `.maestro/<service-name>/maestro-interface/`.

Use `🔒 (inferred)` whenever auth cannot be determined with confidence from explicit configuration.

---

## `.maestro/maestro-interface/api-contracts.md`

```markdown
# API Contracts
> **Service:** {service-name} | **Generated:** {YYYY-MM-DD} | **Branch:** {git-branch}

## Table of Contents
- [HTTP Endpoints](#http-endpoints)
- [WebSocket Events](#websocket-events)
- [gRPC Services](#grpc-services)

---

## HTTP Endpoints

### `{METHOD} {/full/path}` {#{method-full-path}}
**Source:** `{file}:{line}`
**Auth:** {public | 🔒 JWT | 🔒 OAuth2 | 🔒 API Key | 🔒 Role: X | 🔒 (inferred)}
**Handler:** `{HandlerName}.{methodName}`

**Request Body** — `{RequestType}` _(omit section if no body)_

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| {field} | {type} | yes/no | {rules or "—"} |

**Path Params** _(omit if none)_

| Param | Type |
|-------|------|
| {id} | uuid |

**Query Params** _(omit if none)_

| Param | Type | Required | Default |
|-------|------|----------|---------|
| {page} | number | no | 1 |

**Response** — `{ResponseType}` | Status `{200 | 201 | 204}`

| Field | Type | Notes |
|-------|------|-------|
| {field} | {type} | {notes or "—"} |

**Error responses** _(omit if unknown)_
- `400` - validation error
- `401` - unauthenticated
- `403` - forbidden

---

<!-- Repeat block per endpoint, sorted by path then method -->

---

## WebSocket Events

**Namespace:** `{/namespace or /}`
**Source:** `{gateway-file}:{line}`
**Auth:** {public | 🔒 JWT on connection | 🔒 (inferred)}

### Incoming: `{event:name}` {#{ws-in-event-name}}

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| {field} | {type} | yes/no | {rules or "—"} |

**Emits back:** `{ack-event-name}` _(omit if void)_

| Field | Type |
|-------|------|
| {field} | {type} |

---

### Outgoing (server-push): `{event:name}` {#{ws-out-event-name}}
**Triggered by:** {describe trigger}

| Field | Type |
|-------|------|
| {field} | {type} |

---

## gRPC Services

**Proto:** `{path/to/file.proto}`
**Package:** `{proto-package}`

### Service: `{ServiceName}` {#{grpc-servicename}}

#### `{MethodName}` {#{grpc-servicename-methodname}}
**Signature:** `rpc {MethodName} ({RequestMsg}) returns ({stream?} {ResponseMsg})`
**Streaming:** {unary | server-streaming | client-streaming | bidirectional}

**Request - `{RequestMsg}`**

| Field | Type | Field # | Notes |
|-------|------|---------|-------|
| {field} | {proto-type} | {N} | {optional / required / repeated} |

**Response - `{ResponseMsg}`**

| Field | Type | Field # | Notes |
|-------|------|---------|-------|
| {field} | {proto-type} | {N} | |

---
```

---

## `.maestro/maestro-interface/db-schema.md`

```markdown
# Database Schema
> **Service:** {service-name} | **Generated:** {YYYY-MM-DD} | **ORM:** {ORM or storage style}

## Table of Contents
- [Entities](#entities)
- [Relationship Map](#relationship-map)

---

## Entities

### `{table_name}` {#{entity-table-name}}
**Entity class:** `{EntityClass}` - `{file}:{line}`

| Column | Type | Constraints | Default | Notes |
|--------|------|-------------|---------|-------|
| id | uuid | PK | gen_random_uuid() | |
| {col} | {db-type} | {NOT NULL / UNIQUE / FK→table.col} | {value or "—"} | {notes or "—"} |
| created_at | timestamptz | NOT NULL | now() | |
| updated_at | timestamptz | NOT NULL | now() | |
| deleted_at | timestamptz | nullable | — | soft-delete |

**Relations:**

| Type | Field | Target Entity | FK Column | Notes |
|------|-------|---------------|-----------|-------|
| ManyToOne | user | `users` | `user_id` | cascade delete |
| OneToMany | orders | `orders` | — | inverse of orders.user |
| ManyToMany | roles | `roles` | join: `user_roles` | |

**Indexes:**

| Name | Columns | Unique | Notes |
|------|---------|--------|-------|
| idx_{table}_email | email | yes | login lookup |
| idx_{table}_created | created_at | no | |

---

<!-- Repeat per entity, sorted alphabetically -->

---

## Relationship Map

Text ERD - one line per relation, entity names as written above.

```
users           ||--o{ orders          : "places"
users           }o--|| organizations   : "belongs to"
orders          ||--|| order_fills     : "filled by"
users           }o--o{ roles           : "assigned"
```

Use crow's foot notation:
- `||` exactly one
- `o|` zero or one
- `}o` zero or many
- `}|` one or many

_(This section is best-effort - include only relations you can confirm from entity files or migrations.)_
```
