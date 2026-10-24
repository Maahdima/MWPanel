# Reseller contribution: current development state

This branch is an incremental implementation, not a finished reseller release.
Existing administrative workflows remain available. Reseller accounts can identify
and manage their own authentication settings, but existing administrative routes
are deliberately forbidden to resellers until scoped replacements are complete.
Do not deploy this branch as the completed feature described in the proposal.

## Implemented and verified

- Legacy-safe nullable peer ownership and database-backed account roles.
- Current role/active-account verification after JWT signature validation, including
  immediate revocation of old administrative tokens. Profile changes are bound to
  the authenticated username. `/api/auth/me` returns the current principal.
- Exact byte/rial pricing, half-up rounding at a single posting boundary,
  overflow rejection, and historical rate/quantity snapshots.
- Transactional prepaid holds, signed balances for credit/liabilities, immutable
  ledger/audit records, payload-bound idempotency, explicit adjustments, recoverable
  release/commit transitions, and atomic mode changes preserving balances.
- Router-bound REST adaptors that can operate concurrently without shared mutable
  selection; deterministic legacy default selection; persisted TLS configuration;
  correct server-specific connectivity checks and errors for missing routers.
- A loopback-only development Mock RouterOS command, reusable REST test fixture,
  authentication checks and deterministic failure injection for tests.

The ledger primitives are **not yet connected to peer HTTP mutations**. They must
not be represented as end-to-end billing. Likewise, bound adaptors do **not yet**
make existing peer services, sync and jobs fully multi-router: explicit persisted
server relationships, composite local identifiers and per-server service/job
routing remain necessary.

## Remaining feature work

1. Scoped reseller peer operations, CRUD, resource assignment, limits and audited
   ownership transfer; role-aware UI and tenant-scoped exports/statistics.
2. Explicit peer/interface server identity and safe legacy assignment, composite
   uniqueness, and router-specific peer/sync/queue/scheduler/background operations.
3. Global/individual pricing administration and integration of durable financial
   intents with every paid peer path, including full reset and expiry extension.
4. Anchored postpaid periods, invoices, early/partial settlement, recurring fees,
   grace enforcement and exactly-once period processing.
5. Card instructions, protected receipt upload/download, approval/rejection and
   atomic, idempotent payment allocation. Submissions must never credit balances.
6. Separate restriction reasons, payment-first suspended UI and bounded durable
   router suspension/recovery reconciliation preserving manual/expiry/quota state.
7. Resumable migration preflight, destination verification, metadata switch,
   configuration/QR regeneration and safe source cleanup with stable public UUIDs.
8. Complete UI, regression/security coverage, PostgreSQL execution, screenshots
   where relevant, and a final upstream issue/PR after authorization.

## Build and test

Build the UI first because Go embeds `ui/dist`:

```sh
cd ui
npm ci --no-audit --no-fund
npm run build
cd ..
mkdir -p .local/test-data
DATA_DIR="$PWD/.local/test-data" go test -race ./...
DATA_DIR="$PWD/.local/test-data" go vet ./...
git diff --check
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o .local/mwp ./api/cmd/main.go
```

Use a disposable local database and loopback listen address. Follow README env
loading (`ENV_FILE`), and set `DB_DIALECT=sqlite`, `HOST=127.0.0.1`, a separate
`DATA_DIR`, and first-run admin credentials. The implementation currently reads
`HOST`/`PORT`; README's configuration table also mentions `SERVER_HOST`/`SERVER_PORT`,
so inspect `api/config/config.go` when preparing settings.

For a mock router, choose a local-only test password and launch explicitly:

```sh
MOCK_ROUTEROS_PASSWORD='local-test-only' go run ./api/cmd/mock-routeros -listen 127.0.0.1:18080
```

Add server `127.0.0.1`, REST port `18080`, username `mockadmin` and that test password
in the local panel, with SSL disabled. This command is a development tool: it does
not emulate the entire RouterOS operating system, keys, networking or real traffic.
Tests use two independent instances and deliberately colliding local object IDs.
No production path automatically falls back to a mock.

## Verification results (2026-10-24)

- `npm ci`: succeeded without dependency-file changes.
- `npm run build`: succeeded (includes `tsc -b` and Vite production build).
- `go test -race ./...`: succeeded for the implemented foundation.
- `go vet ./...`: succeeded.
- Embedded production UI/Go build: succeeded.
- Local HTTP smoke check: login, current identity, mock server registration,
  server availability and peer listing all returned HTTP 200.
- Five local project skills passed the skill validator. Local execution guidance
  is excluded using `.git/info/exclude` rather than added to upstream source.
- `npm run lint`: baseline failure before UI modifications: 10 errors and one
  warning across existing interface/peer/server table cells and shared form
  components. No UI source changes have been made in this implementation stage.
- First Go baseline run started before UI assets existed and failed with
  `pattern dist/*: no matching files found`; subsequent runs after the documented
  UI build succeeded.
- PostgreSQL was not executed; Docker's daemon was unavailable. The tests above
  exercised SQLite. PostgreSQL acceptance remains outstanding.

No issue, branch or PR has been published remotely. The issue proposal is local at
`docs/reseller-management-issue.md`. A final PR is not ready while the items above
remain incomplete.
