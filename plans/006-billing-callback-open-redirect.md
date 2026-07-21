# Plan 006: Fix open redirect in the public billing gateway callback

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/billing/handler.go internal/domain/tenant.go`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

`GET /billing/callback/:gateway` is a **public, unauthenticated** endpoint (the payment
gateway redirects the user's browser to it). It reads `?org=<slug>` from the query and
substitutes it, unvalidated, into the redirect host via
`strings.ReplaceAll(template, "{slug}", slug)`. Because `/` terminates the host portion of
a URL, a crafted `org` value makes the browser redirect to an attacker-controlled host —
an open redirect on a public endpoint, usable for phishing (the link looks like it points
at the real payment domain). The fix validates the slug against the same strict slug
charset the rest of the app already enforces.

## Current state

File: `internal/billing/handler.go`:
```go
// line 159
func (h *Handler) Callback(c *gin.Context) {
    gateway := domain.GatewayName(c.Param("gateway"))
    authority := c.Query("Authority")
    status := c.Query("Status")
    slug := c.Query("org")                          // <-- attacker-controlled
    inv, err := h.svc.HandleCallback(c.Request.Context(), gateway, authority, status == "OK")
    if err != nil {
        c.Redirect(http.StatusFound, h.resultURL(slug, "error", nil))
        return
    }
    ...
    c.Redirect(http.StatusFound, h.resultURL(slug, outcome, inv))
}

// line 179
func (h *Handler) resultURL(slug, outcome string, inv *domain.Invoice) string {
    base := strings.ReplaceAll(h.appURLTemplate, "{slug}", slug) + "/org/billing/result?status=" + outcome
    if inv != nil {
        base += "&invoice=" + inv.ID.String()
    }
    return base
}
```
`h.appURLTemplate` is e.g. `http://{slug}.localhost:5173` / `https://{slug}.<domain>`.

Existing validator to reuse — `internal/domain/tenant.go:58`:
```go
var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
func ValidateSlug(slug string) error {
    if len(slug) < 2 || !slugPattern.MatchString(slug) { return ErrInvalidSlug }
    if _, reserved := ReservedSlugs[slug]; reserved { return ErrInvalidSlug }
    return nil
}
```
The pattern permits only `[a-z0-9-]` — no `/`, `.`, `@`, `:`, so a valid slug cannot escape
the host label.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (billing) | `go test -race -count=1 ./internal/billing/...` | all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/billing/handler.go`
- `internal/billing/*_test.go`

**Out of scope**:
- `internal/billing/service.go` `HandleCallback` — the server-side payment verification is correct; do not touch it.
- `internal/domain/tenant.go` — reuse `ValidateSlug`, don't change it.
- The redirect flow itself (302 to a result page) — keep it; only validate the slug.

## Git workflow

- Branch: `advisor/006-billing-callback-open-redirect`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1: Validate the slug before it reaches the redirect

In `Callback`, immediately after reading `slug := c.Query("org")`, reject an invalid slug
so it never enters `resultURL`:
```go
slug := c.Query("org")
if domain.ValidateSlug(slug) != nil {
    c.String(http.StatusBadRequest, "invalid organization")
    return
}
```
This runs before `HandleCallback`, so a malformed `org` gets a plain 400 with no redirect.
A legitimate gateway callback always carries the valid slug set at checkout, so this does
not affect real payments.

**Verify**: `go build ./...` → exit 0.

### Step 2: Run the suite

**Verify**:
- `go test -race -count=1 ./internal/billing/...` → all pass
- `make lint` → exit 0

## Test plan

Add handler tests (model after existing billing handler tests; use `httptest` +
`gin.CreateTestContext` as the package already does). Cases:
- `org` = a value containing `/` or a full external host → response is 400, **no** `Location` redirect header pointing off-template.
- `org` = a value with `.` or `@` or `:` → 400.
- `org` = a valid slug (e.g. `acme`) → behaves as before (302 to the templated host).
- Missing `org` (empty) → 400 (empty fails `ValidateSlug`).

Verification: `go test -race -count=1 ./internal/billing/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/billing/...` exits 0 with the new redirect tests passing
- [ ] `make lint` exits 0
- [ ] An invalid `org` slug yields a 400 and never reaches `resultURL`/`c.Redirect`
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 006 updated

## STOP conditions

- Excerpts don't match live code (drift).
- `domain.ValidateSlug` no longer exists or its charset now permits `/`/`.` — STOP and report; the fix relies on the strict charset.
- Legitimate callbacks are found to send `org` values that fail `ValidateSlug` (e.g. slugs with uppercase or dots) — STOP; validation would break real payments, and the slug source needs reconsidering.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: confirm the 400 path is before `HandleCallback` so no partial state depends on the redirect.
- Stronger follow-up (deferred): derive the redirect slug from the settled invoice's organization server-side instead of trusting the query param at all. That needs an org-slug lookup wired into the handler; out of scope here but noted as the ideal end-state.
- Any other handler that substitutes a request value into a redirect host must apply the same validation.
