# Alignment with the existing MWPanel architecture

The source of truth is this repository's original `c935efa` architecture, compared with the completed feature at `78b0561`. No separate boilerplate repository was supplied or substituted. Existing examples include `api/http/server.go`, `api/http/schema/server.go`, `api/service/server.go`, `ui/src/api/servers.ts`, server hooks, `features/servers/components`, and the shared form/table/dialog controls.

## Findings and changes

| Inconsistency | Alignment |
| --- | --- |
| Reseller request/view types and custom JSON parsing lived in services | Moved to `api/http/schema`, used by controllers and services as in the original server flow; exact integer strings/numbers and optional values remain supported |
| Database-role JWT policy was embedded in route registration | Moved the existing configuration to `api/http/middleware`; current-role resolution, fail-closed route allowlist and HTTP status behavior are preserved |
| Components performed Axios calls and assembled query functions | Added domain API modules and named query/mutation hooks; components retain interaction state, exact form conversion and payload-bound retry keys |
| Response schemas were declared in components | Consolidated them in the existing schema directory |
| Download code was duplicated in a query hook and the existing traffic export | Shared blob saving in `lib/download.ts`; authenticated file requests live in `api/files.ts` |
| Reseller controls lived beside page entries and list columns were inline | Reused the existing feature `components` organization and extracted reseller columns |
| Scoped configurations had their own table/loading implementation | Reused `DataTable` and `DataTableSkeleton`, preserving row actions and tenant-scoped data; filtering, pagination and visibility use the original table controls |
| Administration used a separate form implementation and custom selects | Reused React Hook Form, Zod, `FormField`, `SelectDropdown`, `PasswordInput` and shared `Select` controls, preserving resource assignments and exact submission values |

Echo controllers, direct GORM services, existing response envelopes, routing, authentication and RouterOS adapters remain the same architecture. No repository layer, new dependency or alternative application was introduced. The pricing, ledger/holds, invoice settlement, receipt verification, independent restrictions, outbox and migration/reconciliation algorithms were not changed. Their transactions and durable records are necessary for financial correctness and external-router recovery; removing them to resemble simple CRUD would weaken integrity.

## Verification on 2026-10-26

Before refactoring, the normal Go suite, vet, UI TypeScript/Vite build and lint passed. After the changes:

- `GOMAXPROCS=2 go test -p 1 -timeout 180s ./...` passed, including authorization/IDOR, exact billing, receipt approval, suspension/restoration, Mock RouterOS, multi-router migration, public configurations and embedded assets.
- `GOMAXPROCS=2 go vet -p 1 ./...` passed.
- Service and HTTP suites passed against isolated PostgreSQL fixture schemas. Docker was initially stopped; after starting the installed runtime, the disposable test container completed and was removed.
- Targeted race tests passed for prepaid concurrency, receipt approval, purchase replay and concurrent migration workers. The full race suite was not rerun for this structural change.
- `npm run build`, `npm run lint`, the embedded production binary build and `git diff --check` passed. A new Go schema regression test covers optional/null values, numbers above JavaScript's exact range, int64 limits and malformed/overflowing input.
- A real browser checked admin login on a custom panel prefix, shared form selection and successful unchanged settings submission, light/dark appearance, logout, reseller login, scoped billing and the shared configuration table. All 43 recorded API requests returned 200; no repeated 403 or administrative API request occurred during the reseller flow. The browser used disposable demo data with an empty peer list; populated peer actions remain covered by service/HTTP tests rather than a claim of exhaustive browser automation.

No database was reset, no real router was contacted, and the temporary application/Mock servers were stopped. The existing limitations remain: same-router interface migration is rejected and real RouterOS acceptance is still required before deployment.

## UI spacing follow-up

Reseller pages now use the same heading/action spacing as the existing server page. Dialog fields group labels and controls consistently; responsive creation fields stack at small widths. Shared `DialogFooter` keeps Cancel and submit actions visible while long bodies scroll, using explicit form association without changing request payloads. Dialog descriptions match the original accessibility pattern and histories use one scrolling body.

Prettier checked all 65 non-generated UI source files changed by the contribution; Go formatting required no correction. UI build, lint and embedded-asset tests passed. A browser verified scrolling the long settings form while its footer stayed visible, and submitting unchanged demo settings through that footer succeeded. The prior backend/financial suites were not repeated for this presentation-only pass. Temporary application and Mock servers were stopped afterward.
