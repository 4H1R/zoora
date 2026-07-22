# Plan 007: SMS OTP verify — attempt cap + invalidate on wrong guess

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/connectors`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P2
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

`VerifySMSOTP` compares a 6-digit code and only deletes it on **success**; a wrong guess
returns an error but leaves the code live for its full 5-minute TTL, with no per-code
attempt cap. The only rate limit is on *requesting* codes, not *verifying* them. An
authenticated user can therefore brute-force the 6-digit code (10^6 space, 5-minute window)
for a phone number they submitted but do not own, linking a victim's phone as an SMS
connector on their own account. Adding an attempt counter that invalidates the code closes
the brute-force window.

## Current state

File: `internal/connectors/service.go`.

Request side (context — already rate-limited per hour):
```go
// line 148
func (s *service) RequestSMSOTP(ctx context.Context, dto domain.RequestSMSOTPDTO) error {
    ...
    n, err := s.rdb.Incr(ctx, otpRLKey(caller.UserID)).Result()  // per-hour request cap
    ...
    if n > otpMaxPerHour { return domain.ErrRateLimited }         // otpMaxPerHour = 3
    code, err := sixDigitCode()                                  // "%06d" of rand<1e6
    rec, _ := json.Marshal(otpRecord{Phone: dto.Phone, Code: code})
    s.rdb.Set(ctx, otpKey(caller.UserID), rec, otpTTL)           // stored under caller
    return s.sms.SendOTP(ctx, dto.Phone, code)
}
```
Verify side (the vulnerable path):
```go
// line 189
func (s *service) VerifySMSOTP(ctx context.Context, dto domain.VerifySMSOTPDTO) error {
    caller, ok := domain.CallerFromCtx(ctx)
    if !ok { return domain.ErrForbidden }
    raw, err := s.rdb.Get(ctx, otpKey(caller.UserID)).Result()
    if err == redis.Nil {
        return domain.NewValidationError(map[string]string{"code": "no pending verification — request a new code"})
    }
    if err != nil { return fmt.Errorf(...) }
    var rec otpRecord
    json.Unmarshal([]byte(raw), &rec)
    if rec.Code != dto.Code {
        return domain.NewValidationError(map[string]string{"code": "incorrect code"})  // <-- code NOT deleted, no attempt count
    }
    s.rdb.Del(ctx, otpKey(caller.UserID))
    ...
}
```
- The OTP is keyed by `caller.UserID` (`otpKey`). Redis client is `s.rdb`.
- Constants near the top of the file: `otpMaxPerHour = 3`, `otpTTL`, and helpers `otpKey`,
  `otpRLKey`. Follow the same style for a new attempt-counter key.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (connectors) | `go test -race -count=1 ./internal/connectors/...` | all pass |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/connectors/service.go`
- `internal/connectors/*_test.go`

**Out of scope**:
- `RequestSMSOTP` request-rate logic — unchanged.
- The push-token / other connector methods — unchanged.
- The SMS platform client — unchanged.

## Git workflow

- Branch: `advisor/007-sms-otp-attempt-cap`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1: Add an attempt-cap constant and key helper

Near the existing OTP constants, add:
```go
const otpMaxAttempts = 5
```
Add a key helper next to `otpKey` (e.g. `otpAttemptsKey(userID uuid.UUID) string`)
following the same construction as `otpKey`.

**Verify**: `go build ./...` → exit 0.

### Step 2: Count attempts and invalidate on a wrong guess

In `VerifySMSOTP`, on the mismatch branch, increment an attempt counter (bounded to the
OTP's TTL) and delete the code once the cap is reached, so a burned window forces a new
request:
```go
if rec.Code != dto.Code {
    n, _ := s.rdb.Incr(ctx, otpAttemptsKey(caller.UserID)).Result()
    if n == 1 {
        s.rdb.Expire(ctx, otpAttemptsKey(caller.UserID), otpTTL)
    }
    if n >= otpMaxAttempts {
        s.rdb.Del(ctx, otpKey(caller.UserID))
        s.rdb.Del(ctx, otpAttemptsKey(caller.UserID))
    }
    return domain.NewValidationError(map[string]string{"code": "incorrect code"})
}
```

### Step 3: Clear the attempt counter on success

On the success path, delete the attempt counter alongside the code:
```go
s.rdb.Del(ctx, otpKey(caller.UserID))
s.rdb.Del(ctx, otpAttemptsKey(caller.UserID))
```

**Verify**: `go build ./...` → exit 0.

### Step 4: Run the suite

**Verify**:
- `go test -race -count=1 ./internal/connectors/...` → all pass
- `make lint` → exit 0

## Test plan

The connectors tests use `miniredis` (see `github.com/alicebob/miniredis/v2` in go.mod) or a
real Redis via the existing test harness — follow whatever the current connectors tests do.
Cases:
- After `otpMaxAttempts` wrong guesses, the stored OTP key is gone and a subsequent verify (even with the correct code) returns the "no pending verification" error.
- A correct code within the attempt budget succeeds and clears both keys.
- The attempt counter resets after a fresh `RequestSMSOTP` (new code) — confirm a new request lets the user try again (the counter key expired or is reset; if not reset by request, add a `Del(otpAttemptsKey)` in `RequestSMSOTP`).

Verification: `go test -race -count=1 ./internal/connectors/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/connectors/...` exits 0 with new brute-force tests passing
- [ ] `make lint` exits 0
- [ ] After N wrong guesses the OTP is invalidated; success clears both keys
- [ ] `git status` shows only in-scope files
- [ ] `plans/README.md` row for 007 updated

## STOP conditions

- Excerpts don't match live code (drift).
- The connectors tests cannot reach a Redis (real or fake) and there is no existing pattern to follow — STOP and report; do not add a new Redis test dependency unilaterally.
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Ensure `RequestSMSOTP` resets the attempt counter so a legitimate re-request isn't pre-blocked (add `s.rdb.Del(ctx, otpAttemptsKey(caller.UserID))` there if the test in Step 4 shows it's needed).
- Reviewer: confirm the counter TTL is bounded (never a permanent lockout) and that success always clears it.
- Consider (deferred) also capping verify attempts per hour independent of a single code, if abuse continues.
