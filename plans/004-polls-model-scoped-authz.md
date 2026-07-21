# Plan 004: Polls — enforce model-scoped authorization

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/polls internal/domain/poll.go cmd/api/main.go`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

The `internal/polls` service applies **no** tenant/object authorization on reads, votes, or
result access. `GetByID`, `Answer`, `Results`, and `ListAnswers` only check that a caller
exists, then operate on any poll by ID. `Create` attaches a poll to any `model_id` the
request names. `List` for a `polls:update_any` holder resolves to an empty scope that
returns polls across **every** org. The `Poll` model has no `organization_id` column, so
nothing scopes it implicitly. Result: any authenticated user can read another org's poll,
see who voted for what, read tallies, and cast/stuff votes by guessing/knowing a poll
UUID. Polls attach to a `class` or a `live_session`, both of which resolve to a class with
an owner and enrolled members — so the fix is to authorize against that owning class, the
same way `qa` does via a `ModelAuthorizer`.

## Current state

Files:
- `internal/polls/service.go`:
  - `Create` (line 34): sets `ModelType`/`ModelID` from the DTO, no ownership check.
  - `GetByID` (line 59): existence check only.
  - `Update` (line 67) / `Delete` (line 94): use `canManagePoll` (line 30) =
    `caller.CanManage(poll.UserID, domain.PermPollsUpdateAny)` — owner-or-`update_any`, but
    **no** org/class check.
  - `List` (line 116) + `resolveListScope` (line 130): `polls:update_any` ⇒
    `domain.PollListScope{}` (AllOrgs=false, OwnerID=nil) ⇒ repo applies no `WHERE` ⇒
    cross-org results.
  - `Answer` (line 143): validates options/closed, no participation check.
  - `ListAnswers` (line 193) / `Results` (line 201): existence check only.
  - Constructor `NewService(repo, answers, logger)` (line 18) — no class/member repos yet.
- `internal/domain/poll.go`: `Poll.ModelType` (string) + `Poll.ModelID` (uuid). Create DTO
  requires them (lines 47-48). Poll model types seen in use: `"class"` (model_id = class
  ID) and `domain.ChatModelLiveSession` = `"live_session"` (model_id = live room ID; polls
  are closed by model at `internal/livesessions/service.go:513`).
- Only elevation permission is `PermPollsUpdateAny` (`internal/domain/permissions.go:58`).
  There is no `polls:view_any`.

Reference pattern — `internal/livesessions/model_authorizer.go` resolves a `live_session`
model_id (a LiveRoom ID) → session → class, then applies participate/moderate rules. Build
the poll equivalent, extended to also handle `"class"` (model_id **is** the class ID).

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (polls) | `go test -race -count=1 ./internal/polls/...` | all pass |
| Build cmd | `go build ./cmd/...` | exit 0 |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/polls/model_authorizer.go` (create)
- `internal/polls/service.go`
- `cmd/api/main.go` (poll service wiring — inject the repos)
- `internal/polls/*_test.go`

**Out of scope**:
- Adding an `organization_id` column / migration to polls — not needed; authorize via the owning class.
- `internal/domain/poll.go` shape changes beyond adding model-type constants if absent.
- The poll HTTP routes / capability perms in `handler.go` — keep as the capability gate.

## Git workflow

- Branch: `advisor/004-polls-model-scoped-authz`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 0: Confirm the model-type universe

Run: `grep -rn "CreatePollDTO\|ModelType:" internal/ --include='*.go' | grep -iv _test`
plus inspect callers of the poll service's `Create`. Confirm polls are only created with
`model_type` in {`"class"`, `"live_session"`}. **If any other model_type exists, STOP** —
the authorizer below must cover every type or it will wrongly 403 legitimate polls.

### Step 1: Build a poll ModelAuthorizer

Create `internal/polls/model_authorizer.go`. It resolves `(modelType, modelID)` to the
owning class and applies participate/moderate rules using poll permissions:
```go
package polls

type modelAuthorizer struct {
    rooms    domain.LiveRoomRepository
    sessions domain.ClassSessionRepository
    classes  domain.ClassRepository
    members  domain.ClassMemberRepository
}

func (a *modelAuthorizer) classForModel(ctx context.Context, modelType string, modelID uuid.UUID) (*domain.Class, error) {
    switch modelType {
    case "class":
        return a.classes.FindByID(ctx, modelID)
    case domain.ChatModelLiveSession: // "live_session": modelID is a LiveRoom ID
        room, err := a.rooms.FindByID(ctx, modelID)
        if err != nil { return nil, err }
        session, err := a.sessions.FindByID(ctx, room.ClassSessionID)
        if err != nil { return nil, err }
        return a.classes.FindByID(ctx, session.ClassID)
    default:
        return nil, domain.ErrUnsupportedModelType
    }
}

// canModerate: admin, polls:update_any holder, or the owning teacher.
func (a *modelAuthorizer) canModerate(ctx context.Context, caller domain.Caller, modelType string, modelID uuid.UUID) (bool, error) {
    class, err := a.classForModel(ctx, modelType, modelID)
    if err != nil { return false, err }
    return caller.CanManage(class.UserID, domain.PermPollsUpdateAny), nil
}

// canParticipate: moderator, or an enrolled member of the owning class.
func (a *modelAuthorizer) canParticipate(ctx context.Context, caller domain.Caller, modelType string, modelID uuid.UUID) (bool, error) {
    class, err := a.classForModel(ctx, modelType, modelID)
    if err != nil { return false, err }
    if caller.CanManage(class.UserID, domain.PermPollsUpdateAny) {
        return true, nil
    }
    return a.members.Exists(ctx, class.ID, caller.UserID)
}
```
Use a literal `"class"` or add a `domain.PollModelClass` constant next to
`ChatModelLiveSession` if you prefer — either is fine, keep it consistent.

**Verify**: `go build ./...` → exit 0.

### Step 2: Inject the authorizer into the poll service

- Add a `modelAuth *modelAuthorizer` (or an interface) field to the poll `service`.
- Extend `NewService` to accept `rooms`, `sessions`, `classes`, `members` repos (or a
  prebuilt authorizer) and construct/assign it.
- In `cmd/api/main.go`, update the poll service construction to pass
  `liveRoomRepo, classSessionRepo, classRepo, classMemberRepo` (all already declared,
  main.go:140-153).

**Verify**: `go build ./...` and `go build ./cmd/...` → exit 0.

### Step 3: Gate `Create` on moderation

In `Create` (line 34), after binding the caller and before building the poll:
```go
ok, err := s.modelAuth.canModerate(ctx, caller, dto.ModelType, dto.ModelID)
if err != nil { return nil, err }
if !ok { return nil, domain.ErrForbidden }
```

**Verify**: `go build ./...` → exit 0.

### Step 4: Gate participation reads/votes

In `GetByID`, `Answer`, `Results`, and `ListAnswers`: after loading the poll (all except
`ListAnswers` already load it; for `ListAnswers` load via `s.repo.FindByID(ctx, pollID)`),
add:
```go
allowed, err := s.modelAuth.canParticipate(ctx, caller, poll.ModelType, poll.ModelID)
if err != nil { return ..., err }
if !allowed { return ..., domain.ErrForbidden }
```
`ListAnswers` currently discards the caller (`_, ok := ...`) — bind it. For `ListAnswers`
(per-voter disclosure) require **moderation**, not just participation:
`s.modelAuth.canModerate(...)`.

**Verify**: `go build ./...` → exit 0.

### Step 5: Tighten `Update`/`Delete` and `List`

- `Update` (line 67) / `Delete` (line 94): after loading the poll, replace the
  `canManagePoll` check with `s.modelAuth.canModerate(ctx, caller, poll.ModelType, poll.ModelID)`
  (keep returning `ErrForbidden` on false). This adds the missing org/class dimension.
- `List` (line 116): for non-admins, require `q.ModelType != nil && q.ModelID != nil`,
  authorize that model via `canParticipate`, and keep the existing per-model filter. Reject
  (`ErrForbidden`) a non-admin listing with no model filter, so the empty-scope cross-org
  path is unreachable. Admins keep `AllOrgs`.

**Verify**: `go build ./...` → exit 0.

### Step 6: Run suites

**Verify**:
- `go test -race -count=1 ./internal/polls/...` → all pass
- `go build ./cmd/...` → exit 0
- `make lint` → exit 0

## Test plan

Existing tests in `internal/polls/service_test.go` build polls with `ModelType: "class"` —
extend them with a fake authorizer (or fake class/member repos) so you can control
participate/moderate. Cases:
- `GetByID`/`Answer`/`Results`: non-participant → `ErrForbidden`; enrolled member → success; `Answer` by non-participant creates no answer rows.
- `ListAnswers`: participant-but-not-moderator → `ErrForbidden`; moderator → success.
- `Create`: non-moderator of the target model → `ErrForbidden`.
- `Update`/`Delete`: non-moderator → `ErrForbidden`; owner/`update_any` → success.
- `List`: non-admin with nil model filter → `ErrForbidden`; with a model they can't access → `ErrForbidden`; admin → returns.

Verification: `go test -race -count=1 ./internal/polls/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` and `go build ./cmd/...` exit 0
- [ ] `go test -race -count=1 ./internal/polls/...` exits 0 with new authz tests passing
- [ ] `make lint` exits 0
- [ ] All poll read/vote/result/mutate methods authorize against the owning class
- [ ] `List` refuses unscoped non-admin listing; `resolveListScope` no longer yields an all-org empty scope for `update_any`
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 004 updated

## STOP conditions

- Excerpts don't match live code (drift).
- Step 0 finds a poll `model_type` beyond `class`/`live_session` — the authorizer would 403 it; STOP and report so the type can be added.
- `ChatModelLiveSession` is not `"live_session"` or its model_id is not a LiveRoom ID — STOP; the resolution chain is wrong.
- An existing test asserts a non-member can read/vote on a poll — indicates intended behavior; STOP.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Every new poll `model_type` must get a branch in `classForModel`, or it will be denied. Call this out in review and keep it in sync with wherever polls are created.
- Reviewer: confirm `ListAnswers` (voter identity) requires moderation, not mere participation — that's a privacy boundary.
- If polls ever need to attach to a non-class model, this authorizer needs redesign; note it here.
