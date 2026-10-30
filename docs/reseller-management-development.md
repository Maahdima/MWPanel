# Reseller contribution development status

> Current source status (2026-10-30): contribution tests, test-only database helpers and Mock RouterOS were removed at the user’s request after the activation/speed checks. Test results below record pre-removal verification; current included verification consists of builds, static analysis and whitespace checks. No test or mock tooling is shipped.

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
- Embedded production assets include underscore-prefixed dynamic-route chunks; the production binary build is verified.

- Resumable cross-router migration, destination preflight, preserved UUID/ownership/accounting/usage, final source sampling, verified metadata switching, refreshed private configurations/QR codes, safe cleanup and compensation before switching. Source traffic enforcement continues during preparation.
- Complete scoped ledger/invoice CSV exports with exact rials, formula-safe text and private temporary-file cleanup.

## Activation, provisioning and speed follow-up — 2026-10-30

- Equivalent ISO and legacy RouterOS scheduler dates now compare by calendar date. Verified creation posts its charge before invoking the existing restriction reconciler immediately. A prepaid account remains active with zero credit; previously paid eligible peers remain enabled.
- Provisioning failures persist sanitized stage/error information and retry times. Admin retry only resets backoff; it cannot bypass ownership, live leases, verification, existing restrictions or accounting. Worker ticks have bounded deadlines and skip unavailable routers for the remainder of a batch.
- Admin detail uses the original cards, tabs, tables and dialogs to separate account, billing, provisioning and actual synchronization. Resource assignments, limits, wallet/debt, invoices and configuration counts are visible. Reseller lists retain read-only/payment access under suspension.
- Per-configuration download/upload speeds are configurable during creation and through **Manage → Change speed limits**. Units use the existing decimal RouterOS K/M/G convention, blank means unlimited. Positive values up to 1 Tbit/s are accepted with three decimal places. Changes use the existing free durable-operation outbox, preserve accounting and restrictions, target the owned peer's router/queue and verify the resulting limit before saving. Missing queue metadata is recovered by deterministic UUID name; unexpected targets are refused.

Pre-removal checks passed: `GOMAXPROCS=2 go test -p 1 -timeout 180s ./...` (SQLite service/HTTP suites), `go vet -p 1 ./...`, embedded Go binary build, UI `npm run build`, `npm run lint`, changed-source Prettier check, 12 Node UI-state tests, and `git diff --check`. The removed regression suite covered zero-credit activation, unapproved/approved receipts, scheduler normalization, retry/lease/restart recovery, mixed restrictions, speed validation/creation, lost-response retries, zero speed-change charges and two-router isolation.

Earlier in this follow-up, browser checks covered admin/reseller login on the user-run app and disposable-data desktop/mobile dark-mode activation screens. The browser transport became unavailable during the final speed-form review, so the new speed form has build/static verification but no completed final browser check. The user-run port 3000 process and its data were not restarted or changed. No physical router, current PostgreSQL run or current race-detector run is claimed.

## Historical verification — 2026-10-26

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
