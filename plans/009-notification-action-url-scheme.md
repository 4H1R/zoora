# Plan 009: Validate notification `action_url` scheme

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/notifications internal/domain/notification.go`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P2
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

A notification's `action_url` is validated only for length (`omitempty,max=500`), never for
scheme. It is stored, echoed to inbox clients, passed as the push `link`
(`internal/notifications/service.go:641-645`), and appended to Telegram/Bale/SMS message
bodies (`:709`). A sender (anyone with `notifications:send` / `send_any`) can set it to a
`javascript:` URI or an attacker origin — enabling stored-XSS-on-click or phishing against
every recipient in their authorized audience, depending on how clients render/handle it.
Constraining the scheme to `http`/`https` (or a relative path) at the write boundary closes
this server-side regardless of client behavior.

## Current state

DTO — `internal/domain/notification.go:163`:
```go
ActionURL *string `json:"action_url" binding:"omitempty,max=500"`
```
(part of `SendNotificationDTO`; also present as a stored field, `notification.go:95`, and on
the model at `:228`).

Consumers (why unvalidated input is dangerous):
- `internal/notifications/service.go:641-645` — `link := *n.ActionURL` passed to `Push.SendMulticast(...)`.
- `internal/notifications/service.go:707-712` (`botMessage`) — appends `*n.ActionURL` verbatim to the message body.

The `Send` service method reads `dto.ActionURL` at `internal/notifications/service.go:129`
(and there is a system-notification path, `SendSystem`, nearby). Validation belongs at this
service boundary so both HTTP and system callers are covered, and both the stored value and
every delivery channel are safe.

The repo uses `go-playground/validator/v10` for binding tags and registers custom
validators (see `cmd/api/main.go` around line 84, "failed to register validators") — but a
plain `url` tag would reject legitimate **relative** paths (the code uses `"/"` as the
default link), so validate in the service instead.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (notifications) | `go test -race -count=1 ./internal/notifications/...` | all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/notifications/service.go`
- `internal/notifications/*_test.go`

**Out of scope**:
- The delivery methods (`DeliverPush`, `botMessage`, `smsMessage`) — they become safe once the input is validated; don't change them.
- DTO/model shape — keep `action_url` a `*string`.
- Client rendering — out of scope (backend defensive fix).

## Git workflow

- Branch: `advisor/009-notification-action-url-scheme`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1: Add a scheme validator helper

In `internal/notifications/service.go`, add:
```go
// validateActionURL rejects action URLs whose scheme is not http/https. A
// relative path (no scheme, e.g. "/org/quizzes/123") is allowed. Nil/empty is
// allowed (no link).
func validateActionURL(raw *string) error {
    if raw == nil || *raw == "" {
        return nil
    }
    u, err := url.Parse(*raw)
    if err != nil {
        return domain.NewValidationError(map[string]string{"action_url": "invalid URL"})
    }
    if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
        return domain.NewValidationError(map[string]string{"action_url": "must be an http(s) or relative URL"})
    }
    return nil
}
```
Add `"net/url"` to the imports.

**Verify**: `go build ./...` → exit 0.

### Step 2: Call it in the send path(s)

In `Send` (and `SendSystem` if it also accepts an `ActionURL`), before persisting/enqueuing,
add:
```go
if err := validateActionURL(dto.ActionURL); err != nil {
    return err
}
```
Place it early — after the caller/authz checks, before the notification row is built.
Confirm every code path that sets `ActionURL` from external input flows through this check
(grep `ActionURL` in the package).

**Verify**: `go build ./...` → exit 0.

### Step 3: Run the suite

**Verify**:
- `go test -race -count=1 ./internal/notifications/...` → all pass
- `make lint` → exit 0

## Test plan

Model after existing notification service tests. Cases (unit-test `validateActionURL`
directly plus one `Send` integration case):
- `action_url` = `javascript:alert(1)` → validation error, notification NOT created.
- `action_url` = `data:text/html,...` → validation error.
- `action_url` = `https://example.com/x` → accepted.
- `action_url` = `/org/quizzes/123` (relative) → accepted.
- `action_url` = nil / empty → accepted.

Verification: `go test -race -count=1 ./internal/notifications/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/notifications/...` exits 0 with new validation tests passing
- [ ] `make lint` exits 0
- [ ] Every send path rejects non-http(s), non-relative `action_url`
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 009 updated

## STOP conditions

- Excerpts don't match live code (drift).
- There is a send path that sets `ActionURL` but bypasses `Send`/`SendSystem` (e.g. a repo-direct insert from another package) — STOP and report; the validation must cover it or the fix is incomplete.
- An existing test/system feature legitimately sends a non-http scheme (e.g. a deep-link `myapp://`) — STOP; the allow-list needs widening deliberately.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: confirm `title`/`body` are rendered as text by clients (not HTML) — that's the other half of the XSS surface, owned by the frontend; note it in the PR but it's out of scope here.
- If mobile deep-links (custom schemes) are ever needed, extend the allow-list explicitly rather than removing the check.
