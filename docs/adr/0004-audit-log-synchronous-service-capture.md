# Audit Log: Synchronous Service-Layer Capture, Same-Transaction Hard-Fail

The per-org Audit Log records structural mutations (create/update/delete of resources) for accountability. Success entries are written **explicitly in the service layer**, inside the **same DB transaction** as the change they record, and **hard-fail**: if the audit insert fails, the whole action is rejected. Actor and org come from `CallerFromCtx`; the audit row joins the caller's transaction via the existing ctx-tx pattern (`Transactor.RunInTx` / `TxFromCtx`), so no tx handle is threaded manually. Services depend on a `domain.AuditRecorder` interface injected via constructor — never on an `audit` package — to honor the "features import only domain/platform/config" rule.

We chose this over the obvious alternatives because forensic value depends on both *completeness* and *meaning*. HTTP middleware would guarantee coverage but produce meaningless entries (no resource name, no domain intent). DB triggers lack actor context. An async Asynq path risks committed-but-unlogged actions. Same-tx + hard-fail makes "in the log or it didn't happen" literally true at negligible availability cost, and the service layer is the only place that knows the human-readable target label and the domain verb.

## Consequences

- Coverage is **explicit and forgettable**: a new mutating endpoint without a `Record` call is a silent gap. A guard test over the `AuditTargetType` enum plus a coverage doc (recording deliberate exclusions — chat messages, poll votes, presence, reads) make gaps loud instead of silent.
- **Denied attempts** (403) cannot share these invariants — there is no transaction and the request is already rejected. They are captured centrally in the error middleware, best-effort/soft-fail, with a route-derived target and no friendly label. `Outcome` distinguishes them from `success`.
- Entries are **immutable** — no update or delete endpoint exists, deliberately, so an actor (including a Manager or Platform Admin) cannot cover their tracks.
- `org_id` is the **target's** org, so Platform Admin cross-tenant actions surface in the affected org's log rather than vanishing (admin has no org).
- Content/telemetry (chat/conversation messages, poll votes, presence, seen-markers) is **out of scope** — it already has its own durable home; auditing it would flood the log for no accountability gain.
