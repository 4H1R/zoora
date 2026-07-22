# Plan 011: Scope media download authz to resource membership

> **Executor instructions**: This is an **investigate-then-fix** plan. Do the investigation
> in Step 1 first; if it reveals the assumptions below are wrong, STOP and report rather
> than forcing a fix. Verify each step. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/media`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21
- **Execution status (2026-07-21)**: **BLOCKED at Step 1 (by design).** Vuln confirmed real
  (`PresignDownload`/`GetByID` gate only on org-wide `authorizeOrgAccess`). Fix blocked on a
  cross-feature design decision the owner must make — see "Blocking findings" below.

## Blocking findings (from Step 1 investigation)

The plan's remediation cannot proceed without an owner design decision:

1. **No membership checker reaches the media package.** `media.NewService`
   (`internal/media/service.go:49`) takes only `(MediaRepository, objectStorage,
   entitlements.Service, storageUsageReader, logger)`. Resource-membership authz needs new
   injected dependencies.
2. **The plan's cited reference is inaccurate.** `ValidateAttachments` lives in
   `internal/conversations/media_adapter.go:25` (not the media service) and is a *binding*
   check (does this media belong to this conversation), not a *membership* check. It presumes
   the caller was already proven a member upstream (`conversations/service.go:117
   requireMember`). It cannot serve as the download-authz reference.
3. **The existing `domain.ModelAuthorizer` does not cover media's model types.** Media rows use
   model types `organization`, `conversation`, `ticket` (ModelID = **class** ID), `live_room`,
   `offline_room` (see `internal/domain/media.go:14-35`). The only `ModelAuthorizer`
   implementations (`livesessions`, `polls`) resolve **only** `"live_session"` — a different
   string from media's `"live_room"` — and none resolve `conversation`/`ticket`/`offline_room`.

**Model-type classification for the fix:**
- `organization` → org-wide Shared folder; existing org check is sufficient (keep open).
- nil `OrganizationID` (platform/changelog) → global (keep open).
- `conversation` → needs `ConversationMemberRepository.Exists(convID, caller)`.
- `ticket` (ModelID = class ID) → needs class-membership resolution.
- `live_room` → needs room→session→class→member resolution.
- `offline_room` → needs offline-room/class membership resolution.

**Owner decision required — pick an injection strategy:**
- (a) Extend/add `ModelAuthorizer` resolvers covering all four membership-scoped media model
  strings, inject one into `media.NewService`; or
- (b) Inject several narrow membership-checker ports into `media.NewService`.
Either path edits files beyond this plan's in-scope list (`livesessions`/`conversations`/class/
offline resolvers + `cmd/api/main.go` + `cmd/worker/main.go`). **This plan must be rewritten
around the chosen strategy before re-dispatch.**

## Why this matters

Media download authorization is **org-wide, not resource-scoped**. `authorizeOrgAccess`
(`internal/media/service.go:124`) permits any media row whose `OrganizationID` matches the
caller's org; `PresignDownload` (`:134`) and `GetByID` (`:158`) gate solely on that. The
download route (`internal/media/handler.go:30`) requires only `media:view`. So any org
member with `media:view` can mint a presigned download URL for **any** media in their org
given its ID — including another user's private DM/conversation attachments and other
classes' materials — without being a member of the owning conversation/class. IDs are
UUIDv4 (not blindly enumerable) but leak via shared threads and API responses, so this is a
real object-level authorization gap.

## Current state

File: `internal/media/service.go`:
```go
// line 124
func authorizeOrgAccess(caller domain.Caller, m *domain.Media) error {
    if m.OrganizationID == nil || caller.IsAdmin {
        return nil
    }
    if caller.OrgID == nil || *caller.OrgID != *m.OrganizationID {
        return domain.ErrNotFound   // hides existence cross-org
    }
    return nil
}
// PresignDownload (134) and GetByID (158) call only authorizeOrgAccess.
```
- A `domain.Media` row carries `ModelType` and (a model) `ModelID` linking it to its owning
  resource (conversation, ticket, live room, class collection, etc.). Confirm the exact
  field names in `internal/domain/*media*.go`.
- Write-time validation already resolves membership per model type — see
  `ValidateAttachments` in the media service (search the package). That is the reference for
  which model types are membership-scoped and how membership is checked; **read it before
  writing the fix** and mirror its branching.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (media) | `go test -race -count=1 ./internal/media/...` | all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/media/service.go`
- `cmd/api/main.go` (only if new membership-checker dependencies must be injected)
- `internal/media/*_test.go`

**Out of scope**:
- The org-wide check for genuinely org-global collections (e.g. changelog assets with nil `OrganizationID`, "Shared" collections) — those stay reachable; do not over-restrict them.
- Upload path — unchanged (already validates on write).
- Storage/presign helpers — unchanged.

## Git workflow

- Branch: `advisor/011-media-download-resource-scope`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1 (investigate): Enumerate media model types and their membership rule

Read `ValidateAttachments` and the `domain.Media` definition. Produce a list of each
`ModelType` value and classify it:
- **membership-scoped** (conversation, ticket, live-room/class-scoped) → download must
  verify the caller is a member/participant of that specific model.
- **org-wide** (shared org collection) → the existing org check is sufficient.
- **global** (nil `OrganizationID`) → stays open.

**If** the model types cannot be cleanly classified, or the membership checkers for them are
not already available to the media package, **STOP and report** the finding with your
classification so the owner can decide the injection strategy. Do not guess a membership
rule.

### Step 2 (fix): Add a resource-scoped authorization branch

Extend the authorization used by `PresignDownload` and `GetByID`: after the org check
passes, for membership-scoped model types, call the owning feature's membership check
(the same one `ValidateAttachments` uses) and return `domain.ErrNotFound` when the caller is
not a member. Keep admin bypass and the org-wide/global fast-paths. Prefer injecting narrow
membership-checker interfaces (mirroring how `ValidateAttachments` obtains them) over
importing feature packages directly (feature packages must not import each other — see
CLAUDE.md dependency rules; cross-feature interaction goes through domain interfaces).

**Verify**: `go build ./...` → exit 0.

### Step 3: Run the suite

**Verify**:
- `go test -race -count=1 ./internal/media/...` → all pass
- `make lint` → exit 0

## Test plan

Model after existing media service tests with fake membership checkers. Cases:
- Conversation-attachment media: caller is org member but NOT a conversation member → `ErrNotFound` (no presigned URL).
- Same media, caller IS a conversation member → success (URL issued).
- Org-wide collection media, same-org caller → success (unchanged).
- Global media (nil org) → success for any caller (unchanged).
- Admin → success for all (unchanged).

Verification: `go test -race -count=1 ./internal/media/...` → all pass, new cases included.

## Done criteria

- [ ] Step 1 classification recorded (in the PR description or a comment)
- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/media/...` exits 0 with new membership tests passing
- [ ] `make lint` exits 0
- [ ] `PresignDownload`/`GetByID` verify resource membership for membership-scoped model types
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 011 updated

## STOP conditions

- Excerpts/field names don't match live code (drift).
- Step 1 shows model types can't be cleanly classified or the needed membership checkers aren't reachable without a feature-to-feature import — STOP; report options.
- A change would break a legitimate org-wide download flow — STOP; refine the classification.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: verify the model-type classification matches `ValidateAttachments` exactly — a drift between write-time and read-time membership rules re-opens the gap.
- Any new media `ModelType` must be classified here and in `ValidateAttachments` together.
