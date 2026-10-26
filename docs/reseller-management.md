# Resellers, billing and router migration

MWPanel uses one installation and database. Existing administrators remain superadmins, and existing/imported peers without an explicit owner remain superadmin-owned. Customer public-share pages do not require administrative accounts.

## Administration and resources

Use **Resellers** to create an account, choose prepaid or postpaid billing, set limits and assign interfaces. The interface's existing IP pool supplies client addresses. Resellers receive no server credentials or permission to change global resources. The peer table's owner filter and ownership action let superadmins inspect and transfer peers. Transfers are audited and cannot interrupt pending operations.

Creation, quota changes and renewals for owned peers use the dedicated purchase controls. Legacy update/reset/service paths reject changes that could bypass billing. **Manage** provides free renaming, client-address changes within an assigned pool, key regeneration, explicit manual status, public sharing, quota reduction and deletion. Deletion and reductions do not refund historical charges. A fully suspended reseller has payment submission and read-only configuration access; an admin retains authorized management controls.

Limits can cover active peers, per-peer/total configured traffic and remaining validity. Unlimited reseller traffic or duration is rejected. When transferring a legacy peer, record its original plan duration if full reset is needed; an unknown original duration is never guessed.

## Exact prices and accounting

Money is stored as signed integer rials. The UI displays toman, with ten rials per toman and a decimal digit when needed. The default price is 100,000 rials (10,000 toman) per GiB; one GiB is 1,073,741,824 bytes. A reseller override takes precedence over the global rate. Exact rational arithmetic rounds half up to one rial once; every purchase retains its rate and quantity snapshot.

| Operation | Charge |
| --- | --- |
| Create | quota bytes × rate × max(days, 30) / (GiB × 30) |
| Add traffic | additional bytes × rate / GiB |
| Extend validity | full configured quota × rate / GiB, independent of added days |
| Reset usage | exact consumed bytes sampled at submission × rate / GiB; expiration unchanged |
| Full reset | full configured quota × rate / GiB once; reset usage and restart the recorded duration today |

Prepaid purchases reserve credit before external changes and finalize the debit only when the operation is verified. Concurrent reservations cannot spend the same available credit. Zero remaining credit does not suspend existing paid peers. Postpaid charges accumulate in an anchored fixed-duration period: default 30 days, configurable per account. Early payments do not move that anchor. Recurring fees default to zero and are posted once per period.

At an unpaid period boundary, new billable operations stop immediately. Existing eligible peers remain available during the default five-day grace period. After grace, billing suspension is reconciled to the actual routers. Partial settlement reduces debt but does not clear overdue restrictions. Full settlement clears only billing restrictions; manual disablement, expiration, quota and admin restrictions survive. Mode changes preserve liabilities, credit and history. Admin adjustments and overrides require reasons and are audited.

The dashboard shows recent invoices and ledger entries. **Download full ledger/invoices (CSV)** retrieves the complete scoped history, with exact rial fields. Receipt history is also ownership-scoped. Export files are private temporary files and are removed after the response; text is protected against spreadsheet-formula interpretation.

## Card-to-card receipts

The superadmin configures card number, cardholder and instructions in payment settings. A reseller submits a PNG/JPEG up to 5 MiB, amount, payment date and transaction reference. Image content is decoded and re-encoded, stored privately and served only through authorized routes. A pending receipt has no financial effect. Only admin approval creates credit/settlement; approval is atomic and idempotent, and an approved bank reference cannot be credited twice. Rejected receipts remain in history and can be replaced by a new submission.

## Durable operations and recovery

Peer operations have durable intents, payload-bound idempotency keys and database leases. Paid holds and operation records are created together. A router timeout does not justify a finalized charge or an unverified refund. Workers inventory stable public keys and UUID resource names before retrying creation, so a lost success response can be recovered without duplicate objects. A ledger charge and the matching peer metadata commit together. Free operations use the same outbox without a charge.

Router requests have bounded operation deadlines shorter than their leases. Failed attempts retain their state and use persisted backoff. The purchase worker rotates through pending IDs, and peer reconciliation avoids repeatedly blocking a batch behind one failing router. Operator-visible pending/error states are distinct from manual, quota, expiration, administrative, billing and provisioning restrictions.

SQLite serializes writers; PostgreSQL account-row updates provide financial serialization. Busy/deadlock retries are bounded. An unresolved external operation must be verified or repaired before a hold can be released. Keep the database, private receipt directory and peer-files directory backed up together. SQL diagnostics omit parameter values because they can contain passwords, keys or financial details.

## Changing the default router

1. Assign the destination interface and its pool to the reseller while retaining source resources.
2. Open **Migrate router**, select the destination, optionally provide its public VPN endpoint and enter a reason. The endpoint may differ from the router's management address.
3. Preflight verifies destination connectivity, interface identity, source peer/queue/scheduler identities, available addressing and duplicate keys. A free address is allocated from the assigned destination pool when the old address conflicts or is outside that pool.
4. The workflow records all affected peers before preparing disabled destination peers, queues and schedulers. Consumption and quota enforcement continue on the source while preparation is pending.
5. After destination verification, a peer lease protects the final source sample. The source is disabled and routing metadata, counter baselines and accumulated usage switch atomically. Ownership, private/public/preshared keys, quota, expiration, original duration, billing and public UUID are preserved. Actual queue limits are carried across.
6. Configurations and QR codes are rebuilt privately and atomically. Restriction reconciliation restores only eligible destination peers. Source objects are removed only after destination verification.
7. The reseller default changes after every peer completes. Failed cleanup resumes without re-importing old source objects as admin-owned peers.

**Cancel before switch** requests durable compensation only while no peer has switched routing metadata and no worker is holding the migration lease. Prepared destination objects are verified and removed; eligible source connectivity is restored without undoing other restrictions. After any routing switch, recovery proceeds forward rather than attempting an unsafe automatic rollback.

Clients may need updated downloads. Migration is not promised to be interruption-free. A move between different interfaces on the same router is explicitly refused during preflight; this workflow targets cross-router migration. Conflicting/reused object identities are never silently overwritten. While routing metadata and assets are being switched, configuration downloads return a temporary unavailable response; the public UUID stays stable.

## Database upgrades and verification

Back up before upgrading. AutoMigrate adds ownership, router and billing/workflow fields without deleting legacy peers or assigning them to a reseller. Router-local identities use composite indexes. Legacy router ownership can be inferred only when there is exactly one server; ambiguous installations require an explicit valid `LEGACY_ROUTER_SERVER_ID` and are not partially guessed.

Build `ui/` before Go because production embeds its output, including dynamic-route chunks. The final contribution does not include automated test sources or the PostgreSQL fixture helper. Historical test results in the development notes were obtained before those files were removed; they are not a claim of regression coverage in the final source tree.

The development Mock RouterOS and local test harness have also been removed. Build the application and run `go vet ./...` for current-source static verification. See `reseller-management-development.md` for the validation record and `reseller-management-pr.md` for the contribution description.
