# Reseller contribution development status

Implementation remains in progress on `feat/reseller-billing-management`, based on `c935efa`. No remote issue, push or PR has been published. The issue proposal is prepared locally. Existing Git identity is retained.

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

## Verification executed

On 2026-10-25, `go test -race ./...` passed after integrating durable purchases and free operations. Focused service tests additionally cover initial 15/30/45/60-day pricing, exact reset charges, expiry extension pricing independent of added days, key-change recovery, lost-success router responses at peer/queue/scheduler creation, posting failure after router success, concurrent replay and wallet protection. HTTP tests cover forged price/owner, IDOR, missing idempotency keys and receipt authorization.

`go vet ./...`, UI `npm run build`, UI `npm run lint` and `git diff --check` have passed for completed stages. Subsequent changes require focused re-verification. UI build runs TypeScript checking and Vite; Go embeds that output. No dependency change was necessary.

A real local embedded binary has run with a disposable SQLite database and loopback-only Mock RouterOS. The admin reseller/billing detail page was inspected in the browser and captured. Browser review found an unconditional admin shop-status request in the command menu during reseller login. Role-aware navigation, login/2FA destinations, root-route guards and logout cache handling now prevent it. An actual reseller login subsequently made only five successful auth/billing requests, with no 403 or admin API call. The test browser and local servers were stopped afterward to avoid unnecessary load.

## Remaining work

- Resumable default-router migration with destination preflight, preserved accounting/usage/UUIDs, verified metadata switching, updated configurations/QR codes and safe source cleanup. Unsafe default-interface changes currently return a conflict rather than silently moving metadata.
- Finish browser regression checks of reseller login, suspension/payment-first behavior, purchase/management forms, responsive layout and themes after the latest changes.
- PostgreSQL execution tests; Docker daemon availability has not yet been established. SQLite success is not a claim of PostgreSQL verification.
- Complete history pagination, failure/recovery documentation, final security review, final production build and PR description. Do not describe the entire requested feature as finished until these are verified.

## Recovery and operational boundaries

`PeerPurchase` is the shared durable operation outbox; free mutations have a null charge ID. A held amount is not a ledger debit. Failed/ambiguous external changes stay pending with bounded per-tick work and persisted backoff. Workers inventory stable public keys and UUID resource names before retrying creation. A finalized charge and the matching peer metadata are committed in one account transaction. SQLite serializes writers; PostgreSQL account updates lock rows. Source/router objects with unexpected identities are never changed or deleted.

Pending operations prevent ownership changes and conflicting legacy mutations. New configurations are provisioned disabled and activated only by restriction reconciliation. Billing or admin suspension can still disable an existing peer while a purchase is pending. A legacy peer without a known original duration cannot be fully reset until an admin records that duration explicitly; it is never guessed.
