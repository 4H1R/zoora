# Plan 003: LiveRoom chat — enforce tenant + room-membership authorization

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/chat cmd/api/main.go internal/livesessions/model_authorizer.go`
> Mismatch vs "Current state" excerpts = STOP.

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

The `internal/chat` package (live-room chat) performs **no** per-object tenant or
membership authorization. Every read/list/send/get method only checks that *a* caller
exists or that they hold a broad `chats:*` capability permission — never that the chat's
live room belongs to the caller's org or that the caller is enrolled in the owning class.
Any authenticated user with `chats:view` can enumerate and read **every organization's**
live-room chats and messages by ID, and `chats:write` holders can inject messages that are
broadcast into another tenant's live LiveKit room. This is a cross-tenant confidentiality
and integrity breach. The sibling `qa` feature already solves the identical
polymorphic-authz problem via `domain.ModelAuthorizer`; this plan reuses it.

## Current state

Files:
- `internal/chat/service.go` — chat service. Vulnerable methods:
  - `GetChat` (line 129), `ListChats` (line 191), `SendMessage` (line 198), `GetMessage`
    (line 257): existence/capability check only, no object authz.
  - `UpdateChat` (line 141), `DeleteChat` (line 175), `UpdateMessage` (line 269), plus
    `DeleteMessage`: check `PermChatsManage`/`PermChatsWrite` capability but not object org.
  - Constructor `NewService` (line 32) already receives `liveRooms domain.LiveRoomRepository`.
  - A chat carries `LiveRoomID uuid.UUID` (see `broadcastToLiveRoom`, line 66:
    `s.liveRooms.FindByID(ctx, chat.LiveRoomID)`).
- `internal/chat/handler.go` — routes (lines 41-51), each gated by a `perm(domain.PermChats*)` capability. Keep these; add object authz in the service.
- `internal/livesessions/model_authorizer.go` — the reuse target. `domain.ModelAuthorizer`
  has `CanModerate(ctx, caller, modelType, modelID)` and `CanParticipate(...)`. Passing
  `domain.QAModelLiveSession` ("live_session") + a **LiveRoom ID** resolves room → session
  → class and enforces: manage-permission/owner ⇒ moderator; enrolled member / view-any ⇒
  participant. A chat's `LiveRoomID` is exactly that LiveRoom ID.
- `cmd/api/main.go`:
  - Line 214: `chat.NewService(chatRepo, chatMessageRepo, transactor, log, livekitClient, liveRoomRepo)`.
  - Line 223: `modelAuthorizer := livesessions.NewModelAuthorizer(liveRoomRepo, classSessionRepo, classRepo, classMemberRepo)`.
  - All four repos (`liveRoomRepo`, `classSessionRepo`, `classRepo`, `classMemberRepo`) are
    constructed earlier (lines 140-153), so `modelAuthorizer` can be built **before** the
    chat service.

Reference pattern — how `qa` layers a route capability check with an object-level
`ModelAuthorizer` call: see `internal/livesessions/model_authorizer.go` and its consumers
in `internal/qa/service.go` (search for `CanParticipate` / `CanModerate`). Match that.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (chat) | `go test -race -count=1 ./internal/chat/...` | all pass |
| Tests (build-wide) | `go build ./cmd/...` | exit 0 |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/chat/service.go`
- `cmd/api/main.go` (reorder + one new constructor arg only)
- `internal/chat/*_test.go` (add authz tests)

**Out of scope**:
- `internal/livesessions/model_authorizer.go` — reuse, do not change.
- `internal/chat/repository.go` — do not add org columns/filters here (the resolution goes through the room→class chain via the authorizer).
- The `chats:*` permission definitions and route wiring in `handler.go` — keep as the capability gate.
- The realtime `broadcastToLiveRoom` fanout — unchanged.

## Git workflow

- Branch: `advisor/003-chat-tenant-membership-authz`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1: Inject `domain.ModelAuthorizer` into the chat service

In `internal/chat/service.go`:
- Add field `modelAuth domain.ModelAuthorizer` to the `service` struct.
- Add it as a parameter to `NewService` (append after `liveRooms`) and assign it.

In `cmd/api/main.go`:
- Move the `modelAuthorizer := livesessions.NewModelAuthorizer(...)` line (currently ~223)
  to **above** the `chat.NewService(...)` call (currently ~214). Confirm the four repo
  variables it needs are already declared above that point (they are, lines 140-153).
- Pass `modelAuthorizer` as the new final argument to `chat.NewService(...)`.

**Verify**: `go build ./...` → exit 0.

### Step 2: Add a private authz helper in the chat service

Add two helpers that resolve a chat to its live room and delegate to the authorizer:
```go
func (s *service) canParticipate(ctx context.Context, caller domain.Caller, chat *domain.LiveRoomChat) (bool, error) {
    return s.modelAuth.CanParticipate(ctx, caller, domain.QAModelLiveSession, chat.LiveRoomID)
}
func (s *service) canModerate(ctx context.Context, caller domain.Caller, chat *domain.LiveRoomChat) (bool, error) {
    return s.modelAuth.CanModerate(ctx, caller, domain.QAModelLiveSession, chat.LiveRoomID)
}
```

**Verify**: `go build ./...` → exit 0.

### Step 3: Gate the read/send methods on participation

In `GetChat`, `SendMessage`, `GetMessage`, and `ListMessages`, after loading the chat
(for `GetMessage`/`ListMessages`, load the parent chat via `s.chatRepo.FindByID`), add:
```go
ok, err := s.canParticipate(ctx, caller, chat)
if err != nil {
    return nil, err
}
if !ok {
    return nil, domain.ErrForbidden
}
```
For `GetMessage`, resolve the message's `ChatID` to its chat first, then participate-check.
Keep the existing `FeatureChat` check in `SendMessage`.

**Verify**: `go build ./...` → exit 0.

### Step 4: Gate chat mutation on moderation

In `UpdateChat` and `DeleteChat`, after loading the chat, replace the bare
`PermChatsManage` capability check with a `canModerate` object check (keep the capability
check too, as a cheap pre-filter):
```go
ok, err := s.canModerate(ctx, caller, chat)
if err != nil {
    return nil, err
}
if !ok {
    return nil, domain.ErrForbidden
}
```
(`DeleteChat` currently deletes by ID without loading the chat — load it first via
`s.chatRepo.FindByID(ctx, id)` so the room is known.)

For `UpdateMessage` and `DeleteMessage`: resolve the message → chat, then allow if the
caller is the message author (`msg.SenderID != nil && *msg.SenderID == caller.UserID`) OR
`canModerate` is true; otherwise `ErrForbidden`.

**Verify**: `go build ./...` → exit 0.

### Step 5: Scope `ListChats` to an authorized room

`ListChats` (line 191) currently returns the unscoped repo list. For non-admins, require
the query to target a specific room and authorize it:
```go
caller, ok := domain.CallerFromCtx(ctx)
if !ok { return nil, 0, domain.ErrForbidden }
if !caller.IsAdmin {
    if q.LiveRoomID == nil {
        return nil, 0, domain.ErrForbidden
    }
    allowed, err := s.modelAuth.CanParticipate(ctx, caller, domain.QAModelLiveSession, *q.LiveRoomID)
    if err != nil {
        return nil, 0, err
    }
    if !allowed {
        return nil, 0, domain.ErrForbidden
    }
}
return s.chatRepo.List(ctx, q)
```

**Verify**: `go build ./...` → exit 0. If the frontend relies on an unfiltered
`GET /chats` for non-admins, that call will now 403 — note it in the PR (see Maintenance).

### Step 6: Run suites

**Verify**:
- `go test -race -count=1 ./internal/chat/...` → all pass
- `go build ./cmd/...` → exit 0
- `make lint` → exit 0

## Test plan

Model tests after existing chat tests (or `internal/qa` service tests) and mock
`domain.ModelAuthorizer` (a tiny fake returning configurable bool/err). Cases:
- `GetChat`/`GetMessage`/`ListMessages`/`SendMessage`: authorizer says not-participant → `ErrForbidden`; participant → success.
- `SendMessage` non-participant → no message created, no broadcast.
- `UpdateChat`/`DeleteChat`: non-moderator → `ErrForbidden`.
- `UpdateMessage`/`DeleteMessage`: non-author non-moderator → `ErrForbidden`; author → success.
- `ListChats`: non-admin with nil `LiveRoomID` → `ErrForbidden`; non-admin with a room they can't access → `ErrForbidden`; admin → returns list.

Verification: `go test -race -count=1 ./internal/chat/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` and `go build ./cmd/...` exit 0
- [ ] `go test -race -count=1 ./internal/chat/...` exits 0 with new authz tests passing
- [ ] `make lint` exits 0
- [ ] Every chat read/send/mutate method performs an object-level `canParticipate`/`canModerate` check (not just capability)
- [ ] `ListChats` refuses unscoped non-admin listing
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 003 updated

## STOP conditions

- Excerpts don't match live code (drift).
- `chat.LiveRoomID` turns out **not** to be a LiveRoom ID that `ModelAuthorizer` can resolve (e.g. `classForModel` returns `ErrUnsupportedModelType` for it) — then the reuse assumption is wrong; STOP and report so a chat-specific resolver can be designed instead.
- Reordering `modelAuthorizer` in `main.go` requires moving a repo declaration that isn't yet defined at that point — STOP; do not reorder repo construction.
- An existing test asserts a non-member can read/list chats — indicates intended behavior; STOP.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- **Frontend coordination**: Step 5 changes `GET /chats` semantics for non-admins (now requires `live_room_id`). Confirm the web client always passes it; if it depended on unfiltered listing, that must change too. Flag in the PR.
- Reviewer: verify the participate/moderate mapping matches product intent — a live-room chat's audience is the live session's participants. If chats ever attach to non-live-session models, `ModelAuthorizer` will need another `model_type` branch.
- This mirrors `qa`; keep the two in sync if the authz rules evolve.
