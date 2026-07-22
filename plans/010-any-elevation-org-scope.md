# Plan 010: Org-scope the `_any` elevation in the shared authz resolver

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/platform/authz`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: 001, 002
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

The shared `authz.Resolver.Scope` / `decideScope` grant `ScopeAll` to any holder of the
resource's `*_any` permission **without** checking the resource's org against the caller's.
The *list* paths already guard this (`ListScope`, `resolver.go:44-53`, forces
`OrganizationID: caller.OrgID` for `_any` holders, and the code comment warns about
"unfiltered (cross-tenant) scan") — but the single-object `Scope` path does not. So a
holder of an `*_any` permission (a staff/org-admin tier role) in org A can read or mutate a
single class-scoped object (quiz, gradebook cell, practice/offline room, attendance row)
belonging to org B by ID. This centralizes the missing org check in the one resolver every
class-scoped feature already funnels through.

## Current state

File: `internal/platform/authz/resolver.go`.

```go
// line 24 — pure decision
func decideScope(caller domain.Caller, class *domain.Class, isMember bool, anyPerm domain.PermissionName) Scope {
    if caller.IsAdmin || caller.HasPermission(anyPerm) {   // <-- no org check
        return ScopeAll
    }
    if caller.UserID == class.UserID {
        return ScopeClass
    }
    if isMember {
        return ScopeOwn
    }
    return ScopeNone
}

// line 66 — resolver entrypoint
func (r *Resolver) Scope(ctx context.Context, caller domain.Caller, class *domain.Class, anyPerm domain.PermissionName) (Scope, error) {
    if caller.IsAdmin || caller.HasPermission(anyPerm) {   // <-- no org check
        return ScopeAll, nil
    }
    if caller.UserID == class.UserID {
        return ScopeClass, nil
    }
    isMember, err := r.members.Exists(ctx, class.ID, caller.UserID)
    if err != nil {
        return ScopeNone, err
    }
    return decideScope(caller, class, isMember, anyPerm), nil
}
```
- `class` (`*domain.Class`) carries `OrganizationID` (confirmed by `class.OrganizationID`
  usage at `internal/offlines/service.go:126`).
- `ListScope` (line 44) is the correct reference: `_any` elevation additionally requires a
  non-nil `OrgID` and stamps `OrganizationID: caller.OrgID`.

Consumers of `Scope` include `internal/gradebook/service.go:80`, quiz/practice/offline/
attendance read paths (search `resolver.Scope` and `.Scope(ctx`).

Depends on 001/002 because those already add per-method org guards for users/orgs/roles;
this plan makes the class-scoped resolver consistent with them. Landing them first avoids
churn overlap.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (authz) | `go test -race -count=1 ./internal/platform/authz/...` | all pass |
| Tests (consumers) | `go test -race -count=1 ./internal/gradebook/... ./internal/quizzes/... ./internal/offlines/... ./internal/practices/... ./internal/attendance/...` | all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/platform/authz/resolver.go`
- `internal/platform/authz/resolver_test.go`

**Out of scope**:
- `domain.Caller.CanManage` / `CanManageOwned` (`internal/domain/caller.go`) — they receive only an `ownerID`, not the resource org, so they cannot be org-scoped centrally. Call sites that use them keep their own per-method org guards (see 005 for the classes example). Do NOT change `caller.go` here.
- Feature services — no changes; they call the resolver.
- `ListScope` — already correct.

## Git workflow

- Branch: `advisor/010-any-elevation-org-scope`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1: Add an org-scoped elevation helper

In `resolver.go`, add a small predicate:
```go
// hasOrgScopedAny reports whether the caller may act org-wide on class: admin
// always, or an anyPerm holder whose org matches the class's org.
func hasOrgScopedAny(caller domain.Caller, class *domain.Class, anyPerm domain.PermissionName) bool {
    if caller.IsAdmin {
        return true
    }
    return caller.HasPermission(anyPerm) &&
        caller.OrgID != nil && class.OrganizationID != nil &&
        *caller.OrgID == *class.OrganizationID
}
```

### Step 2: Use it in `Scope` and `decideScope`

Replace the two `if caller.IsAdmin || caller.HasPermission(anyPerm)` elevation checks (lines
25 and 67) with `if hasOrgScopedAny(caller, class, anyPerm)`. The owner
(`caller.UserID == class.UserID`) and member fallbacks are unchanged, so a cross-org `_any`
holder now correctly degrades to owner/member/none instead of `ScopeAll`.

**Verify**: `go build ./...` → exit 0.

### Step 3: Run authz + consumer suites

**Verify**:
- `go test -race -count=1 ./internal/platform/authz/...` → all pass
- `go test -race -count=1 ./internal/gradebook/... ./internal/quizzes/... ./internal/offlines/... ./internal/practices/... ./internal/attendance/...` → all pass
- `make lint` → exit 0

## Test plan

Extend `internal/platform/authz/resolver_test.go`. Cases:
- `_any` holder whose `OrgID` matches `class.OrganizationID` → `ScopeAll`.
- `_any` holder whose `OrgID` differs from `class.OrganizationID` → falls through to owner/member logic (`ScopeClass` if owner, `ScopeOwn` if member, else `ScopeNone`) — NOT `ScopeAll`.
- `_any` holder with nil `OrgID` → not elevated.
- Admin → `ScopeAll` regardless of org.

Verification: `go test -race -count=1 ./internal/platform/authz/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` exits 0
- [ ] authz + listed consumer suites pass under `-race`
- [ ] `make lint` exits 0
- [ ] `Scope`/`decideScope` grant `ScopeAll` for `_any` only when org matches (or admin)
- [ ] `caller.go` unchanged; only `authz/resolver.go` + its test modified (`git status`)
- [ ] `plans/README.md` row for 010 updated

## STOP conditions

- Excerpts don't match live code (drift).
- Any consumer test breaks in a way showing a legitimate flow depends on a **cross-org** `_any` holder reaching `ScopeAll` (i.e. a real super-admin that is not `IsAdmin`) — STOP and report; that would mean the tenancy model has a non-admin cross-org role this plan would break.
- `domain.Class` has no `OrganizationID` — STOP.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: confirm no non-admin cross-org role exists in the product; if one does, it must be modeled explicitly, not via unscoped `_any`.
- Remaining `CanManage`/`CanManageOwned` single-object call sites (gradebook column source, offline/practice/attendance point reads) still need per-method org guards where the resource org is available — track them as follow-ups (plans 012/013 cover two of them).
- Keep `ListScope` and `Scope` elevation logic in sync going forward.
