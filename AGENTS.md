# AGENTS.md

Before scanning the repository for any assigned task, read the files in `.maestro/` first:

- `.maestro/maestro-interface/api-contracts.md` — HTTP endpoints, WebSocket channels, and gRPC methods
- `.maestro/maestro-interface/dependencies.md` — consumed gRPC services and upstream dependency map
- `.maestro/maestro-interface/db-schema.md` — database tables, columns, and relationships
- `.maestro/<service-name>/maestro-interface/api-contracts.md` — service-specific path
- `.maestro/<service-name>/maestro-interface/dependencies.md` — service-specific path
- `.maestro/<service-name>/maestro-interface/db-schema.md` — service-specific path

These files map the full surface of the service and include source file references. Use them to orient before exploring the codebase.

Snapped services:

- none yet
