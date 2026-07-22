# Zoora

Video conferencing, virtual classroom, and LMS SaaS. Multi-tenant: every user, class, role, and billing record belongs to exactly one Organization.

This glossary covers the **core spine** — the tenancy, identity, and access concepts every feature depends on. Feature-specific vocabulary (quizzes, live sessions, billing internals) is added in later sessions.

## Language

### Tenancy & Identity

**Organization**:
The tenant boundary. Every user, class, role, and billing record belongs to exactly one Organization. Identified by a unique slug (its DNS label). Nothing sits above it.
_Avoid_: Tenant, Workspace, Account, Company, School

**User**:
A person's account and identity, scoped to exactly one Organization. Identity is per-org, not global — the same person in two Organizations is two distinct Users. Carries the person's role, permissions, and plan entitlements for that Organization.
_Avoid_: Member, Account (at the Organization level — "member" means class enrollment; see ClassMember)

**Platform Admin**:
A User that belongs to no Organization (org_id NULL, is_admin) and operates across all tenants, bypassing permission checks. Not a role or entity — a special kind of User that lives outside tenancy.
_Avoid_: Super Admin, Superuser, Admin (bare)

### Access Control

**Permission**:
A single granted capability, keyed as `resource:action` (e.g. `users:view`). ~140 exist. An `_any` suffix widens scope from own-resources to the whole Organization. A User's effective permissions are the flattened set from their Role.
_Avoid_: Grant, Scope, Ability

**Role**:
A named bundle of Permissions assigned to a User. Either Preset or Custom.
_Avoid_: Group, Profile

**Caller**:
The authenticated principal for a request — a User plus resolved Permissions and Entitlements, with Platform Admin bypass. Authorization checks run against the Caller, not the User directly.
_Avoid_: Principal, Current User, Session

**Preset Role**:
A global, code-defined Role shared by every Organization — Manager, Teacher, Student. Its Permissions are reconciled from code (`preset_roles.go`) at startup, not via migrations.
_Avoid_: System Role, Built-in Role, Default Role

**Custom Role**:
An Organization-defined Role scoped to that Organization, created and edited by a Manager.
_Avoid_: System Role, Built-in Role

**Manager**:
The Preset Role of an Organization steward — manages users, roles, billing, and every class org-wide.
_Avoid_: Admin, Owner, Org Admin

**Teacher**:
The Preset Role for running one's own classes — creates classes, live sessions, and quizzes, and grades their own students. Defined by role intent, not by owning a class (see Class Owner).
_Avoid_: Instructor, Tutor

**Student**:
The Preset Role for a learner — joins classes and live sessions, takes quizzes, and sees their own grades. Its Permissions are relation-scoped to "own" by the authz resolver.
_Avoid_: Learner, Pupil, Attendee

### Classes

**Class**:
A cohort inside an Organization: one Class Owner plus enrolled Students. NOT a course or curriculum — there is no syllabus or catalog entity.
_Avoid_: Course, Classroom, Cohort, Group, Section

**Class Owner**:
The single User who owns a Class (`Class.UserID`) — whoever held the create permission at creation, not necessarily a Teacher (a Manager can own a Class).
_Avoid_: conflating with Teacher

**Enrollment**:
A Student's current membership in a Class, realized by a ClassMember record. Current-state only — unenrolling hard-deletes the record, leaving no history.
_Avoid_: Registration, Subscription, Membership (bare)

**Class Session**:
A scheduled meeting within a Class that organizes the live rooms held under that Class.
_Avoid_: Lesson, Meeting (bare)

### Plans & Entitlements

**Plan**:
An Organization's `tier_size` key (e.g. `pro_200`) — a Plan Tier plus a member-capacity size. Held inline on the Organization with `PlanExpiresAt`; expiry silently downgrades to Free. There is no Subscription entity.
_Avoid_: Subscription, Package, Tier (alone)

**Plan Tier**:
The capability level of a Plan — free, plus, pro, or max. Determines which Features are granted; capacity size is separate.
_Avoid_: Level, Grade

**Entitlements**:
The capability set an Organization has under its current Plan. Derived in memory from the plan catalog on each request — never persisted, and there is no per-org override.
_Avoid_: Grant, Override

**Feature**:
A boolean capability gate keyed by name (e.g. `FeatureAI`, `FeatureRecording`). An Organization has a Feature purely by virtue of its Plan Tier.
_Avoid_: Flag, Toggle

**Limit**:
A numeric quota within Entitlements (e.g. max users, storage). Convention: **`0` means unlimited**, not zero. The `entitlements` package enforces only the count-based Limits that need a live DB count.
_Avoid_: Cap, Quota (bare)

### Audit

**Audit Log**:
An Organization's append-only, immutable stream of structural mutations, kept for accountability ("who deleted this Class"). Forensic record only — not a user-facing activity feed and not the platform "What's New" (see Changelog). Read by holders of `audit:view_any` (Manager by default); never edited or deleted, not even by a Manager or Platform Admin.
_Avoid_: Activity Log, Changelog, History, Event Log

**Audit Entry**:
One record in the Audit Log: an Actor performed an Action on a Target at a time, with an Outcome. Carries denormalized snapshots (`actor_name`, `target_label`) so it survives the deletion of the User or resource it describes. Its `org_id` is always the **target's** Organization, never the actor's — so a Platform Admin's action is filed under the org it touched.
_Avoid_: Log Line, Audit Record (bare), History Item

**Audit Action**:
A code-defined, closed-set verb on an Audit Entry — `Created`, `Updated`, `Deleted` (extendable with domain verbs like `Enrolled`, `Graded`). A typed constant, not a free string, so filtering is drift-free.
_Avoid_: Event Type, Verb (bare)

**Audit Target Type**:
The code-defined, closed-set kind of resource an Audit Entry concerns (`Class`, `Quiz`, `User`, `Role`, …) — one constant per auditable resource. The choke point that a coverage guard test asserts against.
_Avoid_: Entity Type, Resource Kind (bare)

**Outcome**:
Whether an Audit Entry records a committed change (`success`) or a blocked one (`denied`). Success entries are written service-side in the same DB transaction as the change (hard-fail: no entry → the action is rejected). Denied entries are authorization refusals (403) captured centrally in middleware — best-effort, no transaction, route-derived Target.
_Avoid_: Status, Result (bare)

**System Actor**:
The reserved Actor for mutations with no human Caller — worker jobs, plan-expiry downgrade, cascades. Recorded with a nil `actor_id` and `actor_name = "System"`. Distinguishes "the platform did this" from any User's action.
_Avoid_: Automated, Cron, Robot
