# Proposal: reseller billing, tenant isolation and concurrent MikroTik routing

Add reseller accounts to the existing MWPanel installation, with server-enforced ownership, exact integer-rial prepaid/postpaid accounting, manually approved card-to-card receipts, grace-based suspension/recovery, and resumable router migration.

Implementation will preserve Echo/GORM, React/TanStack, English UI, SQLite/PostgreSQL and the embedded single binary. Legacy peers remain superadmin-owned. Router-local identifiers become composite with server identity. Financial operations use immutable rate snapshots, transactional idempotency and recoverable router intents. Existing share UUIDs remain stable after migration; clients may need refreshed configurations.

Acceptance checklist:
- [ ] Tenant scoping covers direct API calls, exports, statistics, bulk operations and privileged paths.
- [ ] All pricing rules use exact bytes and integer rials, with deterministic rounding and snapshot history.
- [ ] Concurrent prepaid operations never overspend; postpaid periods/partial settlement/restarts do not duplicate obligations.
- [ ] Receipt upload never credits an account before atomic, idempotent admin approval.
- [ ] Billing recovery preserves manual, expiry and quota restrictions.
- [ ] Two mock routers with colliding IDs remain isolated; migrations recover from partial failure.
- [ ] Existing workflows, frontend checks and embedded production builds pass.

See `docs/reseller-management-design.md` for the inspected architecture and phased design. This is a local proposal pending review, not a published issue.
