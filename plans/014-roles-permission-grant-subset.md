# Plan 014: Restrict permissions a non-admin may grant to a role

> **Executor instructions**: This is an **investigate + product-decision + fix** plan. It
> changes an authorization *policy*, so Step 0 requires a decision before code. If the
> decision is unclear, STOP and ask rather than guessing. Update this plan's row in
> `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/roles internal/domain`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P3
- **Effort**: M
- **Risk**: MED
- **Depends on**: 001
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

`roles.Create` and `roles.Update` accept an arbitrary `PermissionIDs` list, validating only
that the IDs *exist* — not that the caller already holds those permissions. A non-admin with
`roles:create`/`roles:update` can therefore mint an org role carrying high-value permissions
(`users:delete_any`, `organizations:update`, `roles:update`, …) that they were never
delegated. Combined with role assignment (Plan 001 fixes the manager-preset escalation, but
a **custom** role built this way can still carry elevated perms and be assigned to oneself),
this is an intra-org privilege-escalation path. The fix is a delegation rule: a non-admin
may only grant permissions they themselves hold. This is a policy change, so it needs a
product decision first.

## Current state

File: `internal/roles/service.go`:
```go
// Create (line 41): perms exist-check only
perms, err := s.permRepo.FindByIDs(ctx, dto.PermissionIDs)
if len(perms) != len(dto.PermissionIDs) {
    return nil, fmt.Errorf("some permission IDs not found: %w", domain.ErrValidation)
}
// Update (line 124): same exist-only check for dto.PermissionIDs
```
- Custom-role creation is already a plan feature (`FeatureCustomRoles`, line 58) and
  admin-bypassed. Preset roles are admin-only.
- `domain.Caller` exposes `HasPermission(name)` and `Permissions []string`
  (`internal/domain/caller.go:15,30`).
- Permissions are addressed by UUID in the DTO but by name on the caller — you will need to
  resolve the requested permission IDs to names (via `s.permRepo`) to compare against
  `caller.Permissions`.

Depends on 001 (which establishes the manager-guard/self-assignment fixes); do 001 first so
the escalation surface is already partly closed and this plan is the remaining piece.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (roles) | `go test -race -count=1 ./internal/roles/...` | all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/roles/service.go` (`Create`, `Update`)
- `internal/roles/*_test.go`

**Out of scope**:
- Preset-role handling (admin-only already).
- Admin callers — they may grant anything; the rule applies only to non-admins.
- The permission catalog / DTO shape.

## Git workflow

- Branch: `advisor/014-roles-permission-grant-subset`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 0 (decision — do this first): Confirm the delegation rule

The proposed rule: **a non-admin may only grant permissions they currently hold** (subset
check). Verify this is the intended product policy. Consider the Manager preset — does a
Manager hold the full set they need to delegate to sub-roles? If the product wants a
different model (e.g. a fixed "delegatable" subset, or managers may grant anything within a
capped list), that changes the implementation.

**If the intended policy is not clearly "grant-only-what-you-hold", STOP and ask.** Do not
ship a policy guess — over-restricting breaks legitimate role authoring; under-restricting
leaves the escalation open.

### Step 1: Enforce the subset check in `Create` and `Update`

For non-admin callers, after resolving `dto.PermissionIDs` to permission **names**, reject
any name the caller does not hold:
```go
if !caller.IsAdmin {
    for _, p := range perms { // perms resolved from permRepo.FindByIDs
        if !caller.HasPermission(domain.PermissionName(p.Name)) {
            return nil, domain.ErrForbidden
        }
    }
}
```
Apply the same block in `Update` (using the `newPerms` it resolves when
`dto.PermissionIDs != nil`). Admins bypass.

**Verify**: `go build ./...` → exit 0.

### Step 2: Run the suite

**Verify**:
- `go test -race -count=1 ./internal/roles/...` → all pass
- `make lint` → exit 0

## Test plan

Model after existing roles tests. Cases:
- Non-admin caller holding {A, B} creating a role with perms {A} → success.
- Non-admin holding {A, B} creating a role with perms {A, C} where C ∉ caller perms → `ErrForbidden`, no role created.
- Same for `Update` adding a not-held permission → `ErrForbidden`.
- Admin granting any permission → success (bypass).

Verification: `go test -race -count=1 ./internal/roles/...` → all pass, new cases included.

## Done criteria

- [ ] Step 0 decision recorded (in PR description): the delegation rule is confirmed
- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/roles/...` exits 0 with new subset tests passing
- [ ] `make lint` exits 0
- [ ] Non-admins cannot grant a permission they don't hold in `Create`/`Update`
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 014 updated

## STOP conditions

- Excerpts don't match live code (drift).
- Step 0 policy is ambiguous or the Manager preset would be unable to author its normal sub-roles under the subset rule — STOP and ask for the intended model.
- `permRepo.FindByIDs` does not return permission names (only IDs) — resolve names another way or STOP.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: confirm the Manager preset's own permission set is a superset of what managers are expected to delegate; otherwise the rule blocks a legitimate workflow.
- This pairs with Plan 001 (role assignment guards). Together they close the "build an elevated custom role and assign it to yourself" path.
- If a future "delegatable subset" concept is introduced, replace the hold-check with it here.
