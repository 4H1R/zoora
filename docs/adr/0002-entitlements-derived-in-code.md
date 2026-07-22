# Entitlements Derived in Code, Not Persisted

The plan catalog and the tier→feature/limit mapping (`tierSpecs`, `PlanCatalog` in `internal/domain/plan.go`) are defined in Go, not in the database. An Organization stores only a `Plan` key (`tier_size`, e.g. `pro_200`) plus `PlanExpiresAt` inline. Entitlements — the capability set (Features + Limits) an Organization has — are **computed in memory per request** from the catalog and attached to the Caller. There is no `plans` table, no Subscription entity, and no per-org feature override.

We chose this because the plan catalog changes with releases, not with tenant data, and code is the natural home for release-versioned rules (reviewable, testable, deployed atomically). The trade-off: changing what a plan grants requires a deploy, and there is deliberately no mechanism to grant one Organization a Feature off-plan.

## Consequences

- Feature checks run against derived state via `Caller.HasFeature` → `Entitlements.Can`; the `internal/entitlements` package enforces only count-based Limits that need a live DB count (users, storage, rooms).
- Adding per-org overrides or a true Subscription lifecycle is a real modeling change, not a config tweak — do not assume a table exists to edit.
- `Limit` convention: `0` means **unlimited**, not zero.
