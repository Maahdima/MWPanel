# Reseller contribution development status

> Current source status: the 23 contribution test files and dedicated PostgreSQL fixture helper have been removed. Test results below are historical runs made before removal. The resulting source builds and passes `go vet`; Mock RouterOS remains available for manual development checks.


Local implementation and review are prepared on `feat/reseller-billing-management`, based on `c935efa`. The proposal is published as [#6](https://github.com/Maahdima/MWPanel/issues/6). Existing Git identity is retained.

## Implemented and verified

- Explicit server identity and legacy-safe ownership migration; SQLite/PostgreSQL-compatible GORM models; current database role resolution and admin guards.
- Genuine per-router adaptors, peer/interface/IP-pool operations, synchronization, exports, counters, expiry and session jobs. Two Mock RouterOS servers test colliding local IDs and isolated failures.
- Admin reseller CRUD, resource assignments, limits and audited ownership transfers. Legacy/imported unassigned peers remain admin-owned. Transfers can explicitly record the original plan duration for full resets of legacy peers.
- Exact integer-rial pricing, byte quantities, deterministic half-up rounding, rate snapshots, prepaid holds and immutable ledger. Anchored fixed-day periods, invoices, recurring fees, partial settlement, grace and audited overrides.
- Private validated PNG/JPEG receipts, pending/approved/rejected states, globally unique bank references and atomic idempotent approval. No wallet credit before approval.
- Independent manual, quota, expiry, billing, administrative and provisioning restrictions. Durable router reconciliation with leases/backoff and safe restoration.
- Durable peer creation, quota purchases, expiry extensions, usage resets and full resets. Holds and intents commit together; router state is verified before financial posting. Recoverable ambiguous responses and database failures; preserved original pricing on retry. Unsupported unlimited reseller plans are rejected.
- Durable free rename, address change, key regeneration, explicit manual status, public-share toggle, quota reduction and deletion. No charge or wallet change. Suspended reseller writes are blocked; authorized admin actions remain possible. Legacy alternate billed/reset paths cannot bypass accounting.
- Existing-style admin reseller list/detail, exact toman inputs/display, receipt review, financial settings/actions, ownership filtering/transfers, scoped reseller billing/configuration pages, purchase and management forms. English UI and existing themes retained.
- Embedded production assets include underscore-prefixed dynamic-route chunks; a regression test checks generated JS/CSS/HTML against embedded files.

- Resumable cross-router migration, destination preflight, preserved UUID/ownership/accounting/usage, final source sampling, verified metadata switching, refreshed private configurations/QR codes, safe cleanup and compensation before switching. Source traffic enforcement continues during preparation.
- Complete scoped ledger/invoice CSV exports with exact rials, formula-safe text and private temporary-file cleanup.

## Verification executed — 2026-10-26

- Entire current Go suite: `GOMAXPROCS=2 go test -p 1 -timeout 180s ./...` passed.
- Focused race tests for concurrent prepaid reservations, receipt approval, purchase replay and concurrent migration workers passed. Migration quota/counter continuation tests passed under the race detector.
- Service and HTTP suites passed against an isolated official PostgreSQL 17.6 container, using independent connection pools and private fixture schemas. The disposable container was removed afterward. SQLite tests also passed.
- `go vet ./...`, UI `npm run build` (TypeScript and Vite), `npm run lint`, embedded production binary build and `git diff --check` passed. No dependency changes were required.
- An earlier full race suite passed; a later full race run under a two-CPU budget exceeded its 180-second package timeout. That timed run is not a pass; current normal-suite and focused race evidence are distinguished above.

A local embedded binary ran against disposable SQLite data and loopback-only Mock RouterOS. Browser checks verified admin/reseller login and role navigation, scoped billing/configurations, migration history and a custom `ADMIN_PANEL_PATH`. The repeated-403 admin request loop was fixed. New dialogs use the existing scrollable, padded dialog body. No physical router or production data was used.

## Material limits and publication

- Same-router moves between different interfaces are rejected by migration preflight. Cross-router migrations are supported and tested; no zero-downtime promise is made.
- A legacy peer with unknown original duration needs an explicit audited duration before full reset.
- Physical RouterOS acceptance and real-world interruption behavior remain operator validation tasks; automated checks use configurable REST mocks.
- The upstream proposal is published as [#6](https://github.com/Maahdima/MWPanel/issues/6); the user authorized pushing the branch and opening one PR. GitHub authentication was renewed for AryanFekri.

## Recovery and operational boundaries

`PeerPurchase` is the shared durable operation outbox; free mutations have a null charge ID. A held amount is not a ledger debit. Failed/ambiguous external changes stay pending with bounded per-tick work and persisted backoff. Workers inventory stable public keys and UUID resource names before retrying creation. A finalized charge and the matching peer metadata are committed in one account transaction. SQLite serializes writers; PostgreSQL account updates lock rows. Source/router objects with unexpected identities are never changed or deleted.

Pending operations prevent ownership changes and conflicting legacy mutations. New configurations are provisioned disabled and activated only by restriction reconciliation. Billing or admin suspension can still disable an existing peer while a purchase is pending. A legacy peer without a known original duration cannot be fully reset until an admin records that duration explicitly; it is never guessed.

## Architectural alignment

The follow-up refactor follows the original repository boilerplate without changing business rules. See [alignment and verification](reseller-management-refactor.md) for the inspected patterns, structural changes and current test evidence.
