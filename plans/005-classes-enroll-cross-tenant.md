# Plan 005: classes.Enroll — block cross-tenant self-enrollment

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/classes`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

`classes.Enroll` loads a class by ID with no org filter, and its "students may self-enroll"
branch lets any user holding the org-wide `classes:join` permission create a `class_members`
row for **any class in any organization** by supplying its UUID. Class membership is the key
that unlocks student-level reads across the app — `canViewQuiz`, `canViewRoom` (live +
offline), gradebook `ScopeOwn`, practice rooms — all resolve access via
`members.Exists(classID, caller.UserID)`. So a self-enrolled outsider gains read access to
another tenant's class content. This closes the cross-tenant boundary on enrollment.

## Current state

File: `internal/classes/service.go`, `Enroll` (lines 305-339):
```go
func (s *service) Enroll(ctx context.Context, classID uuid.UUID, dto domain.EnrollClassMemberDTO) (*domain.ClassMember, error) {
    caller, ok := domain.CallerFromCtx(ctx)
    if !ok {
        return nil, domain.ErrForbidden
    }
    class, err := s.repo.FindByID(ctx, classID)   // <-- no org filter
    if err != nil {
        return nil, err
    }
    // Authorization: teacher/staff/admin may enroll any user. A student may
    // only self-enroll.
    if !canManageClass(caller, class) && dto.UserID != caller.UserID {
        return nil, domain.ErrForbidden
    }
    ...
```
- `class` has `OrganizationID` (see how `CreateRoom` reads `class.OrganizationID` in
  `internal/offlines/service.go:126`).
- `canManageClass(caller, class)` is the manage-tier helper (admin / `classes:*_any` /
  owning teacher) used elsewhere in this file.
- Route: `internal/classes/handler.go:69` — `POST /classes/:id/members`, gated only by
  `perm(domain.PermClassesJoin)`.

Convention: authz failure → `domain.ErrForbidden`; admins bypass via `caller.IsAdmin`
(already inside `canManageClass`).

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (classes) | `go test -race -count=1 ./internal/classes/...` | all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/classes/service.go` (the `Enroll` method only)
- `internal/classes/*_test.go`

**Out of scope**:
- `canManageClass` and other class methods — unchanged.
- Repository files — keep the guard in the service.
- Any self-enrollment product feature (e.g. an "open enrollment" flag) — out of scope; this plan only enforces the tenant boundary. If the product wants self-enroll gated further, that's a separate change (note in Maintenance).

## Git workflow

- Branch: `advisor/005-classes-enroll-cross-tenant`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1: Reject enrollment into a class outside the caller's org

In `Enroll`, immediately after the class is loaded (after line 313) and before the
authorization branch, add a tenant guard for non-admins:
```go
if !caller.IsAdmin {
    if caller.OrgID == nil || class.OrganizationID == nil || *class.OrganizationID != *caller.OrgID {
        return nil, domain.ErrForbidden
    }
}
```
This runs for both the manage path and the self-enroll path, so neither can cross tenants.
Admins keep the ability to enroll into any org.

**Verify**: `go build ./...` → exit 0.

### Step 2: Run the suite

**Verify**:
- `go test -race -count=1 ./internal/classes/...` → all pass
- `make lint` → exit 0

## Test plan

Model after existing `classes` service tests (build a `domain.Caller`, inject via
`domain.WithCaller`; use the package's fake repos). Cases:
- Self-enroll (`dto.UserID == caller.UserID`) into a class in a **different** org → `domain.ErrForbidden`, no member row created.
- Self-enroll into a class in the **caller's** org → success.
- Teacher/manager enrolling another user into a class in a different org → `ErrForbidden`.
- Admin enrolling into any org → success (regression: admin bypass preserved).

Verification: `go test -race -count=1 ./internal/classes/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/classes/...` exits 0 with the new cross-tenant test passing
- [ ] `make lint` exits 0
- [ ] `Enroll` rejects any non-admin enrollment where `class.OrganizationID != caller.OrgID`
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 005 updated

## STOP conditions

- Excerpt doesn't match live code (drift).
- `domain.Class` has no `OrganizationID` field (contradicts `offlines/service.go:126`) — STOP.
- An existing test asserts a user can self-enroll into another org's class — indicates intended behavior; STOP.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: confirm the guard runs before both the manage and self-enroll branches.
- Product follow-up (deferred): even within an org, self-enrollment is currently allowed for anyone with `classes:join`. If the product wants class-level opt-in (invite-only vs open), that's a separate feature — this plan only fixes the tenant boundary.
- Any future bulk-enroll or invite endpoint must carry the same org guard.
