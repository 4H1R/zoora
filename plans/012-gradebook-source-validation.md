# Plan 012: Gradebook auto-column — validate SourceID belongs to the class

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/gradebook cmd/api/main.go`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

`CreateColumn` validates only that an auto-type column has a non-nil `SourceID` — never that
the referenced quiz / practice room / session belongs to the column's class. `fetchAutoData`
then queries submissions/attendance for that attacker-chosen `SourceID` and folds the
results into the class matrix. A teacher can therefore create an auto column whose source is
a quiz/room/session they do **not** own and surface those foreign grades for any student who
also appears on their own roster — a horizontal grade-disclosure path. Validating that the
source belongs to the class at column-create time closes it.

## Current state

File: `internal/gradebook/service.go`.

Service already holds the repos needed for two of three source types
(`internal/gradebook/service.go:23-35`): `quizzes domain.QuizRepository`,
`practiceRooms domain.PracticeRoomRepository`. It does NOT hold a class-session repo.

`CreateColumn` (line 83):
```go
class, err := s.classes.FindByID(ctx, classID)
...
if !canManageGradebook(caller, class) { return nil, domain.ErrForbidden }
if dto.Type.IsAuto() && dto.SourceID == nil {
    return nil, domain.NewValidationError(map[string]string{"source_id": "required for auto column types"})
}
col := &domain.GradebookColumn{ ClassID: classID, ..., SourceID: dto.SourceID, ... }
```
`fetchAutoData` (line 440) switches on `col.Type`:
- `GradebookColumnAutoAttendance` → `attendance.ListBySession(sourceID, ...)` (source = session ID)
- `GradebookColumnAutoPractice` → `practiceSubs.ListByRoom(sourceID, ...)` (source = practice room ID)
- `GradebookColumnAutoQuiz` → `quizSubs.ListByQuiz(sourceID, ...)` (source = quiz ID)

To validate the source's `ClassID` you need: quiz→`ClassID`, practice room→`ClassID`,
session→`ClassID`. Quiz and practice-room repos are present; a `ClassSessionRepository` is
not injected yet (but `classSessionRepo` exists in `cmd/api/main.go:141`).

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (gradebook) | `go test -race -count=1 ./internal/gradebook/...` | all pass |
| Build cmd | `go build ./cmd/...` | exit 0 |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/gradebook/service.go`
- `cmd/api/main.go` (inject the session repo into the gradebook service)
- `internal/gradebook/*_test.go`

**Out of scope**:
- `fetchAutoData` query logic — unchanged (it becomes safe once the source is validated at create time).
- Manual (non-auto) column types — no source, no change.
- Adding org columns anywhere.

## Git workflow

- Branch: `advisor/012-gradebook-source-validation`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1: Inject a ClassSessionRepository

Add `sessions domain.ClassSessionRepository` to the gradebook `service` struct and
`NewService` params; assign it. In `cmd/api/main.go`, pass `classSessionRepo` (declared at
main.go:141) into the gradebook service constructor.

**Verify**: `go build ./...` and `go build ./cmd/...` → exit 0.

### Step 2: Validate the source belongs to the class

In `CreateColumn`, after the `SourceID == nil` check and before building `col`, add a
per-type validation that the source's `ClassID` equals `classID`:
```go
if dto.Type.IsAuto() {
    if err := s.validateSource(ctx, dto.Type, *dto.SourceID, classID); err != nil {
        return nil, err
    }
}
```
And add the helper:
```go
func (s *service) validateSource(ctx context.Context, typ domain.GradebookColumnType, sourceID, classID uuid.UUID) error {
    var srcClassID uuid.UUID
    switch typ {
    case domain.GradebookColumnAutoQuiz:
        q, err := s.quizzes.FindByID(ctx, sourceID)
        if err != nil { return err }
        srcClassID = q.ClassID
    case domain.GradebookColumnAutoPractice:
        r, err := s.practiceRooms.FindByID(ctx, sourceID)
        if err != nil { return err }
        srcClassID = r.ClassID
    case domain.GradebookColumnAutoAttendance:
        sess, err := s.sessions.FindByID(ctx, sourceID)
        if err != nil { return err }
        srcClassID = sess.ClassID
    default:
        return nil
    }
    if srcClassID != classID {
        return domain.NewValidationError(map[string]string{"source_id": "source does not belong to this class"})
    }
    return nil
}
```
Confirm the actual field names (`quiz.ClassID`, `practiceRoom.ClassID`, `session.ClassID`)
and repo `FindByID` signatures; adjust if they differ. **If a source model has no ClassID,
STOP** and report.

**Verify**: `go build ./...` → exit 0.

### Step 3: Run suites

**Verify**:
- `go test -race -count=1 ./internal/gradebook/...` → all pass
- `go build ./cmd/...` → exit 0
- `make lint` → exit 0

## Test plan

Model after existing gradebook service tests (fake repos). Cases per auto type:
- SourceID references a quiz/room/session in a **different** class → `CreateColumn` returns validation error, no column created.
- SourceID references a quiz/room/session in the **same** class → success.
- Manual column type → no source validation, success.

Verification: `go test -race -count=1 ./internal/gradebook/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` and `go build ./cmd/...` exit 0
- [ ] `go test -race -count=1 ./internal/gradebook/...` exits 0 with new validation tests passing
- [ ] `make lint` exits 0
- [ ] `CreateColumn` rejects an auto SourceID whose owning ClassID != the column's class
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 012 updated

## STOP conditions

- Excerpts/field names don't match live code (drift).
- A source model (quiz/practice room/session) has no `ClassID` field — STOP; the ownership can't be validated as designed.
- `UpdateColumn` also lets `SourceID` change (check line 117 onward) — if so, note it; apply the same validation there, or STOP and report if that expands scope significantly.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: check whether `UpdateColumn` can repoint `SourceID`; if it can, it needs the same guard (add it or file a follow-up).
- Any new auto column type must extend `validateSource`.
- Defense-in-depth follow-up (deferred): also re-check ownership in `fetchAutoData` before querying, in case a column predates this validation.
