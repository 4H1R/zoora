# Audit Log Coverage

What the org audit log records, and — just as important — what it deliberately
does not. This matrix reflects the code as actually instrumented (Phase 2), not
an aspirational list.

See `docs/adr/0004-audit-log-synchronous-service-capture.md` and CONTEXT.md
(Audit section). The guard test `internal/audit/guard_test.go` keeps this matrix
honest: every `domain.AuditTargetType` must either be emitted by a service
success path or be listed as a read-only exception below.

## Recorded (success, same-tx, hard-fail)

Success entries are written explicitly in the service layer, inside the **same DB
transaction** as the change. If the audit insert fails, the change is rolled
back (hard-fail).

| Target type   | Actions                                     | Service       | Notes |
|---------------|---------------------------------------------|---------------|-------|
| class         | created, updated, deleted                   | classes       | cascade child counts in `metadata` |
| enrollment    | enrolled, unenrolled                        | classes       | label = student name |
| user          | created, updated, deleted, disabled, enabled| users         | |
| role          | created, updated, deleted                   | roles         | org-facing custom roles; preset/platform-admin paths excluded |
| quiz          | created, updated, deleted                   | quizzes       | |
| question_bank | created, updated, deleted                   | questionbanks | bank CRUD **and** question CRUD both file under `question_bank` (label = bank name) |
| gradebook     | created, updated, deleted, graded           | gradebook     | column CRUD + cell grade set/override; student in `metadata` |
| billing       | updated                                     | billing       | plan activation; **System** actor on the gateway callback path, human actor on the admin manual path |
| live_session  | created, updated                            | livesessions  | **no delete** — there is no user-facing delete; rooms are lifecycle-ended |
| offline_room  | created, updated, deleted                   | offlines      | |
| practice      | created, updated, deleted, graded           | practices     | |
| attendance    | created, updated, deleted                   | attendance    | Mark (created/updated), Update, Delete; bulk/auto-mark **excluded** (computed) |
| org_settings  | updated                                     | orgsettings   | org-facing settings update; platform-admin paths excluded |
| organization  | updated                                     | organizations | org profile update; platform-admin provisioning excluded |
| custom_field  | created, updated, deleted                   | customfields  | definition CRUD (archive maps to deleted) |
| connector     | created, deleted                            | connectors    | link (create) + unlink (delete); delivery/webhook traffic **excluded** |
| ticket        | created, updated                            | tickets       | create + close (close records `updated`); message posts **excluded** |
| poll          | created, updated, deleted                   | polls         | votes **excluded** |
| qa            | created, updated, deleted                   | qa            | question create/moderate/delete; votes **excluded** |
| import        | created, updated                            | imports       | job create + finalize (**finalize is worker-side**, System actor) |
| media         | deleted                                     | media         | delete only; presign/read **excluded** |

## Declared but not emitted (read-only projection)

| Target type    | Why not emitted |
|----------------|-----------------|
| calendar_event | The calendar is a **read-only projection** over classes, live sessions, quizzes, etc. Nothing creates/updates/deletes a calendar event directly, so no service success path emits it. The constant is retained because the denied-attempt middleware's route map uses it to file `denied` 403s against calendar routes, and `AuditTargetType.Valid()` must accept `calendar_event` for filtering those denied entries. Listed as a `readOnlyException` in the guard test. |

## Recorded (denied, best-effort, no tx)

Any mutating request (POST/PUT/PATCH/DELETE) that resolves to 403 is captured by
`middleware.AuditDenied` with `outcome=denied`, actor + org from the Caller, and a
route-derived target type. No friendly label (the resource was never loaded).
This is best-effort and runs outside any transaction — a recorder error is logged,
never surfaced to the client.

## Deliberately excluded (record nothing)

These are excluded by decision, not oversight. Removing an exclusion is a
deliberate change.

- All reads (GET/HEAD).
- Chat & conversation message content (send/edit/delete) — already durable in its
  own tables.
- Poll votes and Q&A votes — telemetry/content, not accountability.
- Presence / heartbeats.
- Notification "seen" markers.
- Attendance bulk/auto-mark (computed, not an accountable human action).
- Connector delivery/webhook traffic (operational, not structural).
- Ticket message posts (content; only the ticket lifecycle create/close is
  accountable).
