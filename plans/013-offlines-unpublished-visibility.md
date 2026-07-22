# Plan 013: Hide unpublished offline rooms from students

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/offlines`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P3
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

`canViewRoom` grants an enrolled student access to an offline room regardless of its
`PublishedAt` timestamp. Since `CreateRoom` accepts a `PublishedAt`, scheduled/unpublished
release is clearly intended — but a student can fetch an unpublished room (title,
description, attachments) by ID before its release time, defeating scheduled release. This
is a same-tenant, member-only exposure (low severity), fixed by gating the member branch on
publication.

## Current state

File: `internal/offlines/service.go`:
```go
// line 96
func (s *service) canViewRoom(ctx context.Context, caller domain.Caller, room *domain.OfflineRoom) (bool, error) {
    if canManageRoom(caller, room) {           // admin / offlines:update_any / creator
        return true, nil
    }
    if caller.HasPermission(domain.PermOfflinesViewAny) {
        return true, nil
    }
    return s.members.Exists(ctx, room.ClassID, caller.UserID)   // <-- ignores PublishedAt
}
```
- `room.PublishedAt` is a `*time.Time` (nil = never published / draft), set in `CreateRoom`
  (`internal/offlines/service.go:132`).
- Managers (`canManageRoom`) and `offlines:view_any` holders must still see unpublished
  rooms (to preview/schedule).
- `GetRoom` (line 145) gates on `canViewRoom`.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (offlines) | `go test -race -count=1 ./internal/offlines/...` | all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/offlines/service.go` (`canViewRoom` only)
- `internal/offlines/*_test.go`

**Out of scope**:
- Manager/`_any`/creator visibility — must keep seeing unpublished rooms.
- List endpoints — check whether the list path already filters unpublished for students (it likely does; see the class list-scope handling). If it does NOT, note it, but the point-read `GetRoom` is the confirmed gap; do not expand scope without a STOP note.

## Git workflow

- Branch: `advisor/013-offlines-unpublished-visibility`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1: Gate the member branch on publication

Change the member fallback so a plain enrolled student only sees published rooms:
```go
func (s *service) canViewRoom(ctx context.Context, caller domain.Caller, room *domain.OfflineRoom) (bool, error) {
    if canManageRoom(caller, room) {
        return true, nil
    }
    if caller.HasPermission(domain.PermOfflinesViewAny) {
        return true, nil
    }
    isMember, err := s.members.Exists(ctx, room.ClassID, caller.UserID)
    if err != nil {
        return false, err
    }
    if !isMember {
        return false, nil
    }
    // Members see a room only once it is published.
    return room.PublishedAt != nil && !room.PublishedAt.After(time.Now()), nil
}
```
Ensure `time` is imported (it is used elsewhere in the package).

**Verify**: `go build ./...` → exit 0.

### Step 2: Run the suite

**Verify**:
- `go test -race -count=1 ./internal/offlines/...` → all pass
- `make lint` → exit 0

## Test plan

Model after existing offlines tests. Cases:
- Enrolled student + room with `PublishedAt = nil` → `GetRoom` returns `ErrForbidden`.
- Enrolled student + room with `PublishedAt` in the future → `ErrForbidden`.
- Enrolled student + room with `PublishedAt` in the past → success.
- Creator/manager + unpublished room → success (regression: still visible).
- `offlines:view_any` holder + unpublished room → success.

Verification: `go test -race -count=1 ./internal/offlines/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/offlines/...` exits 0 with new visibility tests passing
- [ ] `make lint` exits 0
- [ ] Students cannot `GetRoom` an unpublished/future room; managers/`_any`/creators still can
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 013 updated

## STOP conditions

- Excerpts don't match live code (drift).
- `OfflineRoom` has no `PublishedAt` field, or its semantics differ from "nil/future = not yet released" — STOP and report.
- An existing test asserts a student sees an unpublished room — indicates intended behavior; STOP.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: confirm the list endpoint applies the same publication filter for students (so unpublished rooms don't leak via listing either); if not, file a follow-up.
- The same publish-visibility rule likely applies to practice rooms / other scheduled resources — worth a consistency pass later.
