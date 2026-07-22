# Per-Organization User Identity

A User is scoped to exactly one Organization: the `org_id` foreign key lives on the user row, and there is no global User identity and no separate Member/Membership join entity. Usernames are unique only *within* an Organization, so the same real person in two Organizations is two distinct User rows. A Platform Admin is the sole exception — a User with `org_id = NULL`.

We chose this over a global-identity-plus-membership model because tenancy is strict and users almost never span Organizations; inlining membership on the User row keeps every query naturally tenant-scoped and avoids a join on the hottest path (auth). The trade-off: cross-org identity (one login, many orgs) is not possible without a schema change and data migration, and "member" as an org-level concept has no table to point at.

## Consequences

- Look-ups and login are always `(org_id, username)` scoped; see `FindByUsernameAndOrg`.
- Introducing SSO-style shared identity later means adding a real identity/membership layer — a hard, non-local change.
