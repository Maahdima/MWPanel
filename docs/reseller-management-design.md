# Reseller billing and multi-router design

Status: implementation in progress. This is not a completed feature announcement.

## Existing architecture

- `api/cmd/main.go` constructs GORM, the shared RouterOS client registry, scheduled jobs and Echo.
- `api/dataservice/db.go` uses GORM AutoMigrate for SQLite/PostgreSQL. Existing admins have no roles; peers/interfaces have no server or tenant foreign keys.
- `api/http/api.go` registers JWT groups, public UUID share routes, shop proxy and service-token routes.
- `api/service/peer.go` creates/updates/deletes peers, queues, schedulers and config assets. `shop_peer.go` also creates and renews peers; `service_peer.go` exposes lookups.
- `api/service/sync.go` and `sync_preview.go` import/update peers and interfaces and remove stale records. They must never reconcile one router against another router's records.
- `api/cmd/jobs/{traffic,expire,sessions}.go` read counters, reset usage, disable peers and record sessions. `service/interface.go` can cascade peer deletion.
- `api/common/client.go` caches HTTP clients by server name; GetClient(nil) currently selects by map iteration. Every adaptor method uses that default.
- `api/http/wg_peer.go` includes bulk resets, exports, share controls, config/QR downloads and session reads; each is a tenant boundary.
- `api/service/{config_file,qr_code,excel}.go` read peer data and generate files independently of peer controllers.
- `ui/src/features`, `hooks`, `schema`, `routes` and shared `components` provide the existing UI patterns. Preserve English, responsive layout and both themes.

## Mutation inventory

Peer create/update/status/delete; usage reset (single/bulk); share and share expiry; interface cascade deletion; full/selected sync import/update/stale deletion; service-token shop create/charge; traffic/expiry jobs; generated config/QR replacement; ownership transfer; reseller suspension/recovery; destination migration and source cleanup. Financial guards must cover every entry, including privileged service routes and background tasks.

## Proposed boundaries and files

Extend GORM models/migration with reseller ownership (nullable = superadmin), explicit server identity, financial records and durable operation/reconciliation records. Extend the existing JWT account model, resolving current role and status from the database on each protected request. Keep public customer shares unchanged.

Centralize exact pricing in `api/service/billing_pricing.go`; snapshot byte quantities, applied rate, rule and rounded rials at posting. Keep all account updates, ledger records, invoice allocations and idempotency decisions in one database transaction. SQLite uses conditional updates/serialized writers; PostgreSQL additionally uses row locks. No request may finalize a charge for an unsuccessful operation.

Use explicit immutable router selection, never mutable global selection. RouterOS IDs are local to a router: indexes and queries must become composite. A durable intent must precede router mutation, with deterministic object identifiers, verification, bounded reconciliation and compensation. Migration commits destination routing metadata only after destination verification, then cleans source objects. Share UUIDs stay stable; clients may need updated downloads.

Track independent restrictions for manual disablement, quota, expiration, administrative suspension and billing suspension. A reconciliation job derives desired router state from all restrictions and records actual synchronization separately.

Affected existing files: GORM models/db/seeds; authentication/totp services and schemas; API route groups and middleware; peer/interface/pool/server/sync/export/config services and controllers; client registry and all RouterOS adaptors; traffic/expiry/session jobs; UI auth store, routing/sidebar, peer forms/table/hooks/schemas, plus reseller/billing/payment views using shared components.

## Pricing contract

Money is integer rials; 1 toman = 10 rials. Default rate = 100,000 rials/GiB/30 days; 1 GiB = 1,073,741,824 bytes (existing traffic convention). Round nonnegative rational results half up to nearest rial once at posting; use arbitrary precision intermediates and reject int64 overflow.

- Create: bytes × rate × max(days,30) / (GiB × 30).
- Quota increase: added bytes × rate / GiB; reductions never refund.
- Expiry extension: total configured bytes × rate / GiB, independent of added days. Add days to remaining validity; reject invalid/unlimited reseller plans.
- Usage reset: consumed bytes since prior reset × rate / GiB, expiry unchanged.
- Full reset: configured bytes × rate / GiB once; restart configured duration now; no additional consumed-traffic charge.
- Rename, view, config download, key changes, permitted IP changes, manual status and deletion are free; deletion never refunds.

## Ordered implementation and verification

1. Audit, baseline builds/tests, local execution, issue proposal and project guidance.
2. Add nullable legacy-safe ownership/models; migration tests for existing installations.
3. Resolve identities and enforce scoped authorization on all HTTP/service boundaries; IDOR tests before exposing reseller login.
4. Explicit server routing, composite identities and router-specific jobs; two independent Mock RouterOS instances with colliding IDs.
5. Admin reseller CRUD/resource assignment/limits/ownership transfer.
6. Exact pricing, immutable ledger, prepaid conditional debit and postpaid invoices/settlements; concurrent/idempotent tests.
7. Protected receipt uploads and atomic admin approval; MIME/content/size/authorization tests.
8. Grace/suspension/restoration and durable reconciliation; preserve other restrictions.
9. Resumable destination migration, stable shares and generated assets; failure injection tests.
10. Existing-style admin/reseller/billing UI with role-aware navigation and payment-first suspension.
11. Regression/security tests, SQLite/PostgreSQL verification, embedded UI binary build, review and one PR.

## Risks and decisions still requiring implementation

Never infer ownership from user payloads. Historical records survive deletion. Unique payment references and per-account operation keys prevent replay with changed payloads. Failed or uncertain RouterOS requests require verification before compensation or retry. Offline peers remain readable. Existing unassigned peers are superadmin-owned.

The brief specifies both exactly 30-day anchored periods and configurable intervals. Use default 30-day fixed-duration periods, permit explicit per-account interval configuration, preserve the activation anchor and never use calendar-month arithmetic. Mode changes preserve liabilities and credit atomically; overdue obligations remain blocking until settled or explicitly overridden by an audited admin action.

No production data or real routers are used for local verification. No push, upstream issue publication or PR publication without the user's authorization.
