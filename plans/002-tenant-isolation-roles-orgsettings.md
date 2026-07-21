# Plan 002: Tenant isolation on roles + orgsettings single-object methods

> **Executor instructions**: Follow step by step. Run every verification command and
> confirm the expected result before the next step. On any "STOP condition", stop and
> report. When done, update this plan's row in `plans/README.md`.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/roles internal/orgsettings`
> On any change to an in-scope file, compare the "Current state" excerpts to live code
> before proceeding; mismatch = STOP.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

Same class as Plan 001, in the `roles` and `orgsettings` services. A non-admin holding
`roles:view` / `roles:update` / `roles:delete` or `organizations:update` can read, modify,
or delete another organization's custom role or settings by supplying its UUID, because
these methods only check for a preset flag (roles) or nothing at all (orgsettings) — never
that the object belongs to the caller's org. `roles.GetByID` performs **no caller check at
all**. Fixing this closes cross-tenant disclosure and tampering of another tenant's
authorization model and configuration.

## Current state

Files:
- `internal/roles/service.go` — role service.
- `internal/orgsettings/service.go` — org-settings service.

Roles model has `OrganizationID *uuid.UUID` (nil for preset roles). `roles.List`
(`internal/roles/service.go:183-193`) is the correct pattern — for non-admins it forces
`f.OrganizationID = caller.OrgID` and `IncludePreset = true`.

Vulnerable methods:

1. `roles.GetByID` — `internal/roles/service.go:95-97`:
   ```go
   func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
       return s.roleRepo.FindByID(ctx, id)
   }
   ```
   No caller, no org check. Route `internal/roles/handler.go:29` gates only `PermRolesView`.

2. `roles.Update` — `internal/roles/service.go:99-157`. Blocks non-admins from touching
   **preset** roles (`role.IsPreset && !caller.IsAdmin`, line 110) but for a *custom* role
   there is no `role.OrganizationID == *caller.OrgID` check.

3. `roles.Delete` — `internal/roles/service.go:159-181`. Same: preset-only guard, no
   org match for custom roles.

4. `orgsettings.Get` — `internal/orgsettings/service.go:23-32`. Takes `orgID` from the
   path; no caller check at all.

5. `orgsettings.Update` — `internal/orgsettings/service.go:39-58`. Same; no caller/org
   match. (`AdminUpdate` at line 61 is correctly `caller.IsAdmin`-gated — leave it.)

Note: `orgsettings.GetByOrgID` (line 35) is an *internal provider* method called by other
services (`domain.OrganizationSettingsProvider`), NOT an HTTP entrypoint — it must stay
open (no caller in that ctx). Only guard the HTTP-facing `Get`/`Update`, and route the
guard so `GetByOrgID` still calls an unguarded path.

Conventions: authz failure → `domain.ErrForbidden`. Admin bypass via `caller.IsAdmin`.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (roles) | `go test -race -count=1 ./internal/roles/...` | all pass |
| Tests (orgsettings) | `go test -race -count=1 ./internal/orgsettings/...` | all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/roles/service.go`
- `internal/orgsettings/service.go`
- test files in `internal/roles/` and `internal/orgsettings/` (add/create `service_authz_test.go` if none fits)

**Out of scope**:
- `roles/role_repository.go` / any repository — keep the guard in the service.
- `orgsettings.AdminUpdate` — already correctly admin-gated.
- `orgsettings.GetByOrgID` provider method — must remain unguarded (internal caller).
- Preset-role handling — the existing preset guards stay as-is.

## Git workflow

- Branch: `advisor/002-tenant-isolation-roles-orgsettings`
- Conventional-commit messages. No push/PR unless instructed.

## Steps

### Step 1: Guard `roles.GetByID`

Replace the two-line body with a caller+org check. Preset roles stay readable by anyone
(they are shared); custom roles must match the caller's org:
```go
func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
    caller, ok := domain.CallerFromCtx(ctx)
    if !ok {
        return nil, domain.ErrForbidden
    }
    role, err := s.roleRepo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if !caller.IsAdmin && !role.IsPreset {
        if role.OrganizationID == nil || caller.OrgID == nil || *role.OrganizationID != *caller.OrgID {
            return nil, domain.ErrForbidden
        }
    }
    return role, nil
}
```

**Verify**: `go build ./...` → exit 0.

### Step 2: Guard `roles.Update` and `roles.Delete`

In both methods, after the existing `role.IsPreset && !caller.IsAdmin` check and before
any mutation, add the custom-role org guard:
```go
if !caller.IsAdmin && !role.IsPreset {
    if role.OrganizationID == nil || caller.OrgID == nil || *role.OrganizationID != *caller.OrgID {
        return nil, domain.ErrForbidden
    }
}
```
(In `Update` this goes right after line 115; in `Delete` right after line 171.)

**Verify**: `go build ./...` → exit 0.

### Step 3: Guard `orgsettings.Get` and `orgsettings.Update`

Add a caller+org guard at the top of each of the HTTP-facing `Get` (line 23) and `Update`
(line 39). Do NOT add it to `GetByOrgID` (line 35):
```go
caller, ok := domain.CallerFromCtx(ctx)
if !ok {
    return nil, domain.ErrForbidden
}
if !caller.IsAdmin && (caller.OrgID == nil || *caller.OrgID != orgID) {
    return nil, domain.ErrForbidden
}
```
Because `GetByOrgID` calls `Get`, and `GetByOrgID`'s ctx has no caller, moving the guard
into `Get` would break the internal provider. **Solution**: keep an unguarded internal
helper and put the guard only on the exported HTTP methods. Concretely: rename the current
guard-free body of `Get` to a private `get(ctx, orgID)`, have `GetByOrgID` call `get`
directly, and have the exported `Get` do the caller check then call `get`. Apply the same
caller check at the top of `Update`.

**Verify**: `go build ./...` → exit 0 (confirm `GetByOrgID` still compiles and is not
gated).

### Step 4: Run suites

**Verify**:
- `go test -race -count=1 ./internal/roles/... ./internal/orgsettings/...` → all pass
- `make lint` → exit 0

## Test plan

Model after existing tests in each package (build a `domain.Caller`, inject via
`domain.WithCaller`). New cases:
- `roles.GetByID`: non-admin caller org A reading a custom role owned by org B → `ErrForbidden`; reading a preset role → success; reading own-org custom role → success.
- `roles.Update` / `roles.Delete`: non-admin targeting org B's custom role → `ErrForbidden`.
- `orgsettings.Get` / `Update`: non-admin whose `OrgID != orgID` → `ErrForbidden`; matching → success.
- Regression: `GetByOrgID` (provider) still returns settings when the ctx has no caller (must NOT 403).

Verification: `go test -race -count=1 ./internal/roles/... ./internal/orgsettings/...` → all pass.

## Done criteria

- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/roles/... ./internal/orgsettings/...` exits 0 with new tests passing
- [ ] `make lint` exits 0
- [ ] `roles.GetByID/Update/Delete` guard custom roles by org; presets unchanged
- [ ] `orgsettings.Get/Update` are caller-guarded; `GetByOrgID` provider is NOT (regression test proves it)
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 002 updated

## STOP conditions

- Excerpts don't match live code (drift).
- `GetByOrgID` is called from a ctx that *does* carry a caller everywhere (then the split may be unnecessary) — verify by grepping callers of `GetByOrgID`/`OrganizationSettingsProvider`; if unsure, STOP and report.
- An existing test expects a non-admin to read another org's role/settings — indicates intended behavior; STOP.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: verify the `GetByOrgID` internal path stays open — a 403 there would silently break every consumer that reads org settings (attendance thresholds, SMS gate).
- Preset roles are intentionally global; the guard must never block reading them.
- Plan 010 generalizes this org check; keep these guards until then.
