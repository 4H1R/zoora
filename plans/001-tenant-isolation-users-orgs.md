# Plan 001: Tenant isolation on users + organizations single-object methods

> **Executor instructions**: Follow this plan step by step. Run every verification
> command and confirm the expected result before moving to the next step. If anything in
> the "STOP conditions" section occurs, stop and report — do not improvise. When done,
> update the status row for this plan in `plans/README.md`.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/users internal/organizations internal/domain`
> If any in-scope file changed since this plan was written, compare the "Current state"
> excerpts against the live code before proceeding; on a mismatch, treat it as a STOP
> condition.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

Zoora is multi-tenant. The auth middleware proves the caller belongs to the request's
host org and holds a permission, but several `users` and `organizations` service methods
then fetch a row by `:id` and mutate/return it **without** verifying that row belongs to
the caller's org. A non-admin holding an org-role permission (`users:view_any`,
`users:create`, `roles:update`, `organizations:update`) can therefore read or modify
records in *other* tenants by supplying their UUID — a cross-tenant data breach — and in
one case (`AssignRole`) can escalate their own privileges to the Manager preset. The
correct guard already exists on sibling methods in the same files; this plan applies it
uniformly.

## Current state

Files:
- `internal/users/service.go` — user service; several methods miss the org guard.
- `internal/organizations/service.go` — org service; `Update` misses the org guard.
- `internal/users/handler.go` — routes (context only, not modified).

**The correct guard already used by `users.Update` (`internal/users/service.go:119-123`)** — copy this shape:
```go
if !caller.IsAdmin && caller.OrgID != nil {
    if user.OrganizationID == nil || *user.OrganizationID != *caller.OrgID {
        return nil, domain.ErrForbidden
    }
}
```
**And by `organizations.GetByID` (`internal/organizations/service.go:83-85`)**:
```go
if !caller.IsAdmin && (caller.OrgID == nil || *caller.OrgID != id) {
    return nil, domain.ErrForbidden
}
```
**And the manager-role guard already used by `users.Create`/`users.Update` (`internal/users/service.go:76-78`)**:
```go
} else if !caller.IsAdmin && s.isManagerRole(ctx, *dto.RoleID) {
    return nil, domain.ErrForbidden
}
```

Vulnerable methods:

1. `organizations.Update` — `internal/organizations/service.go:89-117`. Only checks a
   caller exists; no `id == *caller.OrgID` match. Lets any `organizations:update` holder
   rewrite any org's name/**slug** (slug is the subdomain routing key, see
   `internal/middleware/tenant.go:51`).

2. `users.GetByID` — `internal/users/service.go:101-106`. Only checks a caller exists,
   then returns `repo.FindByID(id)`. Route `internal/users/handler.go:43` gates it with
   `RequireSelfOrPermission(PermUsersView, PermUsersViewAny, "id")`, so a `users:view_any`
   holder reaches any user in any org.

3. `users.Create` — `internal/users/service.go:81-88`. Builds the user with
   `OrganizationID: dto.OrganizationID` straight from the request body. The seat-limit
   check (`service.go:62-66`) runs against `caller.OrgID`, but the row lands in
   `dto.OrganizationID`. A non-admin can create accounts in another org.

4. `users.AssignRole` — `internal/users/service.go:248-270`. Checks `roles:update`, loads
   the target user, then sets `user.RoleID = &dto.RoleID` with **no** `isManagerRole`
   guard and **no** target-org check. A non-admin with `roles:update` can assign the
   Manager preset (privilege escalation, including to themselves) and to users in other
   orgs.

5. `users.RemoveRole` — `internal/users/service.go:272-294`. Loads the user by ID and
   nulls `RoleID` with no target-org check (cross-tenant privilege strip).

Conventions: services return `domain.ErrForbidden` for authz failures (handlers map it
to 403). Admins (`caller.IsAdmin`) always bypass tenant scoping. `caller` comes from
`domain.CallerFromCtx(ctx)`.

## Commands you will need

| Purpose | Command | Expected on success |
|---------|---------|---------------------|
| Build | `go build ./...` | exit 0 |
| Unit tests (users) | `go test -race -count=1 ./internal/users/...` | ok, all pass |
| Unit tests (orgs) | `go test -race -count=1 ./internal/organizations/...` | ok, all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope** (modify only these):
- `internal/organizations/service.go`
- `internal/users/service.go`
- `internal/users/service_*_test.go` and/or `internal/organizations/*_test.go` (add tests; create a `service_authz_test.go` if no suitable file exists)

**Out of scope** (do NOT touch):
- `internal/auth/middleware.go` and any middleware — the fix belongs in the service layer.
- `admin_service.go` files — the admin surface already re-checks `requireAdmin` and correctly accepts arbitrary orgs; leave it.
- Repository files — do not add org filters to repos here; keep the guard in the service (matches the existing pattern).
- Any DTO shape / response shape — clients depend on it.

## Git workflow

- Branch: `advisor/001-tenant-isolation-users-orgs`
- Commit style matches repo (conventional commits, e.g. `fix(users): enforce tenant scope on GetByID/AssignRole`). See `git log --oneline -5`.
- Do NOT push or open a PR unless the operator instructed it.

## Steps

### Step 1: Guard `organizations.Update`

In `internal/organizations/service.go` `Update` (starts line 89), replace the bare
existence check with a caller+org guard mirroring `GetByID` (line 83). After
`caller, ok := domain.CallerFromCtx(ctx)` fails-closed, add:
```go
if !caller.IsAdmin && (caller.OrgID == nil || *caller.OrgID != id) {
    return nil, domain.ErrForbidden
}
```
(You will need to bind `caller` instead of discarding it with `_`.)

**Verify**: `go build ./...` → exit 0.

### Step 2: Guard `users.GetByID`

In `internal/users/service.go` `GetByID` (line 101), bind the caller, load the user, then
apply the same guard `Update` uses (lines 119-123) before returning:
```go
caller, ok := domain.CallerFromCtx(ctx)
if !ok {
    return nil, domain.ErrForbidden
}
user, err := s.repo.FindByID(ctx, id)
if err != nil {
    return nil, err
}
if !caller.IsAdmin && caller.OrgID != nil {
    if user.OrganizationID == nil || *user.OrganizationID != *caller.OrgID {
        return nil, domain.ErrForbidden
    }
}
return user, nil
```

**Verify**: `go build ./...` → exit 0.

### Step 3: Force org on `users.Create` for non-admins

In `internal/users/service.go` `Create` (line 52), right after the `if !caller.IsAdmin { dto.IsAdmin = false }`
block, constrain the target org so a non-admin cannot create users in another org:
```go
if !caller.IsAdmin {
    dto.IsAdmin = false
    if caller.OrgID != nil {
        dto.OrganizationID = caller.OrgID
    }
}
```
Leave the admin path untouched (admins may set any org — `AdminCreate` handles that
surface). Confirm the `user := &domain.User{ OrganizationID: dto.OrganizationID, ... }`
now receives the forced value.

**Verify**: `go build ./...` → exit 0.

### Step 4: Guard `users.AssignRole` (manager-guard + target-org)

In `internal/users/service.go` `AssignRole` (line 248), after loading the target user and
before `user.RoleID = &dto.RoleID`, add both guards that `Update` already has:
```go
if !caller.IsAdmin && caller.OrgID != nil {
    if user.OrganizationID == nil || *user.OrganizationID != *caller.OrgID {
        return nil, domain.ErrForbidden
    }
}
if !caller.IsAdmin && s.isManagerRole(ctx, dto.RoleID) {
    return nil, domain.ErrForbidden
}
```
(`isManagerRole` is defined at `internal/users/service.go:44` and takes a `uuid.UUID`.)

**Verify**: `go build ./...` → exit 0.

### Step 5: Guard `users.RemoveRole` (target-org)

In `internal/users/service.go` `RemoveRole` (line 272), after loading the target user and
before `user.RoleID = nil`, add:
```go
if !caller.IsAdmin && caller.OrgID != nil {
    if user.OrganizationID == nil || *user.OrganizationID != *caller.OrgID {
        return nil, domain.ErrForbidden
    }
}
```

**Verify**: `go build ./...` → exit 0.

### Step 6: Run the full package suites

**Verify**:
- `go test -race -count=1 ./internal/users/...` → all pass
- `go test -race -count=1 ./internal/organizations/...` → all pass
- `make lint` → exit 0

## Test plan

Add table-driven tests. Model them structurally after the existing authz tests in
`internal/users/service_auth_test.go` (see how it builds a `domain.Caller` and injects it
via `domain.WithCaller(ctx, caller)`), and use the repo's fake/in-memory repositories or
mocks already used there.

New cases (one per fixed method):
- `GetByID`: caller in org A requesting a user in org B → `domain.ErrForbidden`; caller in org A requesting a user in org A → success; admin requesting any → success.
- `Create`: non-admin caller in org A with `dto.OrganizationID = orgB` → created user's `OrganizationID == orgA` (forced), not orgB.
- `AssignRole`: non-admin assigning the Manager preset role → `ErrForbidden`; non-admin assigning a normal role to a user in another org → `ErrForbidden`; non-admin assigning a normal role to a same-org user → success.
- `RemoveRole`: non-admin targeting a user in another org → `ErrForbidden`.
- `organizations.Update`: non-admin whose `OrgID != id` → `ErrForbidden`; matching → success.

Verification: `go test -race -count=1 ./internal/users/... ./internal/organizations/...` → all pass, including the new cases.

## Done criteria

ALL must hold:

- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/users/... ./internal/organizations/...` exits 0, with new authz tests present and passing
- [ ] `make lint` exits 0
- [ ] Each of the 5 methods (`organizations.Update`, `users.GetByID`, `users.Create`, `users.AssignRole`, `users.RemoveRole`) has an explicit non-admin org/manager guard
- [ ] `git status` shows only in-scope files modified
- [ ] `plans/README.md` status row for 001 updated

## STOP conditions

Stop and report (do not improvise) if:

- The excerpts in "Current state" don't match the live code (drift since `0071d2e`).
- `isManagerRole` no longer exists or changed signature — the manager-guard step depends on it.
- Removing the `_` and binding `caller` in `organizations.Update` reveals `caller` is already bound elsewhere in the function (name collision) — reconcile carefully, don't shadow.
- Any existing test *expects* the old cross-tenant behavior (e.g. asserts a non-admin can read another org's user) — that would mean the behavior is intentional; stop and report rather than deleting the test.
- A step's verification fails twice after a reasonable fix attempt.

## Maintenance notes

- Reviewer: confirm admins still bypass (every guard is `!caller.IsAdmin`), and that the `AdminCreate`/admin service paths are untouched.
- This is the same guard Plan 010 will later hoist into shared `authz`/`Caller` helpers. Until then, keep the per-method guards — they are the enforcement.
- Any *new* single-object method added to these services must carry the same guard; call it out in review.
