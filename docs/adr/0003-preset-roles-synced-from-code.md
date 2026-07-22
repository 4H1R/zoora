# Preset-Role Permissions Reconciled from Code at Startup

The three Preset Roles (Manager, Teacher, Student) and their Permission sets are defined in `internal/domain/preset_roles.go`. At API startup, `SyncPermissions` reconciles each Preset Role's `role_permissions` grants to match those code-defined sets. Granting or revoking a Preset Role permission is therefore an edit to that file, not a database migration.

We chose startup reconciliation so preset access is versioned with the code that relies on it and cannot drift per environment. The trade-off: the permission tables are partly authoritative-in-code, so editing `role_permissions` directly for a preset role is pointless — the next boot overwrites it.

## Consequences

- To change a Preset Role's permissions, edit `preset_roles.go`; no migration needed.
- Custom Roles are unaffected — they are org-scoped and never reconciled.
