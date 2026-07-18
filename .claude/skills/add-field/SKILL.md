---
name: add-field
description: >
  Add or change a field on a backend domain model. Touches migration (edit in place — dev mode, not production),
  domain struct, DTOs, service mapping, repository filters, handler swagger annotations, factory, and seeder.
  Use when user says "add field", "add column", "change field", "remove field", or asks to modify a model's schema.
---

You are modifying a field on a domain model in Zoora backend (Go + GORM + gin + postgres).
Dev mode — no new migration files. Edit the existing migration in place.

## Identify scope first

Before touching any file, identify:
1. Which entity/table is changing (e.g. `users`, `classes`, `quizzes`)
2. Which migration file owns that table: `migrations/000NNN_<desc>.up.sql`
3. All domain files: `internal/domain/<entity>.go`
4. Feature package: `internal/<feature>/` (repository.go, service.go, handler.go, admin_handler.go if exists)
5. Factory: `internal/factory/<entity>.go`
6. Seeder: `cmd/seed/main.go`

Ask the user to clarify if entity or operation is ambiguous.

## Step 1 — Migration (edit in place)

Find the existing `.up.sql` that owns the table. Edit the `CREATE TABLE` statement directly.
- Adding field: add column line with type + constraints (NOT NULL, DEFAULT, CHECK, index)
- Removing field: delete the column line + any related INDEX/CONSTRAINT lines
- Changing field: modify the column definition in-place

Common SQL patterns:
```sql
-- nullable
field_name       VARCHAR(255),
-- not null with default
field_name       VARCHAR(255) NOT NULL DEFAULT '',
-- not null no default
field_name       INTEGER NOT NULL,
-- enum via CHECK
field_name       VARCHAR(20) NOT NULL DEFAULT 'value' CHECK (field_name IN ('a', 'b', 'c')),
-- uuid FK
field_name       UUID REFERENCES other_table(id) ON DELETE CASCADE,
-- optional FK
field_name       UUID,   -- + CONSTRAINT line below
-- boolean
field_name       BOOLEAN NOT NULL DEFAULT FALSE,
-- jsonb array
field_name       JSONB NOT NULL DEFAULT '[]',
-- index (add after CREATE TABLE block)
CREATE INDEX idx_<table>_<field> ON <table> (<field>);
```

After editing: tell user to run `make migrate-reset` to apply.

## Step 2 — Domain model struct

File: `internal/domain/<entity>.go`

Update the struct field with correct GORM + JSON tags:
```go
// required string
FieldName string `gorm:"not null" json:"field_name"`
// optional string
FieldName string `gorm:"default:''" json:"field_name"`
// nullable
FieldName *string `json:"field_name,omitempty"`
// bool
FieldName bool `gorm:"not null;default:false" json:"field_name"`
// uuid FK required
FieldName uuid.UUID `gorm:"type:uuid;not null" json:"field_name"`
// uuid FK optional
FieldName *uuid.UUID `gorm:"type:uuid" json:"field_name,omitempty"`
// enum string
FieldName FieldType `gorm:"not null;default:'value'" json:"field_name"`
// jsonb
FieldName []uuid.UUID `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"field_name"`
// hide from JSON (e.g. password)
FieldName string `gorm:"not null" json:"-"`
```

Define enum types near the struct if needed:
```go
type FieldType string
const (
    FieldTypeA FieldType = "a"
    FieldTypeB FieldType = "b"
)
```

## Step 3 — DTOs

File: `internal/domain/<entity>.go` (same file, DTOs below the model struct)

**CreateDTO** — add required field:
```go
FieldName string `json:"field_name" binding:"required,min=2"`
```

**UpdateDTO** — add optional field (pointer for omitempty patch semantics):
```go
FieldName *string `json:"field_name" binding:"omitempty,min=2"`
```

Common binding tags:
- `binding:"required"` — must be present
- `binding:"omitempty,min=N"` — optional but validated if present
- `binding:"required,oneof=a b c"` — enum
- `binding:"required,gt=0"` — positive int
- `binding:"gte=0"` — non-negative
- `binding:"required,uuid"` — UUID
- `binding:"omitempty,uuid"` — optional UUID

Skip adding to UpdateDTO if the field should never change after creation.

## Step 4 — Service (DTO → model mapping)

File: `internal/<feature>/service.go`

In `Create()`: add field mapping from DTO to model struct literal:
```go
model := &domain.Entity{
    // existing fields...
    FieldName: dto.FieldName,  // add this
}
```

In `Update()`: add conditional patch:
```go
if dto.FieldName != nil {
    entity.FieldName = *dto.FieldName
}
```

For required non-pointer UpdateDTO fields:
```go
entity.FieldName = dto.FieldName
```

## Step 5 — Repository (filters/search)

File: `internal/<feature>/repository.go`

Only touch if the field needs to be filterable or searchable.

Add to filter query struct in domain if needed:
```go
// in domain/<entity>.go or domain/filters.go
type ListQuery struct {
    FieldName *string `form:"field_name"`
    // ...
}
```

Add WHERE clause in List():
```go
if q.FieldName != nil {
    base = base.Where("field_name = ?", *q.FieldName)
}
```

For text search fields, update `AllowedSearchFields` in handler's `listConfig`.

## Step 6 — Handler (swagger annotations)

File: `internal/<feature>/handler.go` (or `admin_handler.go`)

For new request body fields — swagger picks them up automatically from the DTO struct, no annotation change needed.

For new query params (filters), add `@Param` annotation:
```go
// @Param field_name query string false "Description"
// @Param field_name query bool false "Filter by field"
```

If field is searchable/orderable, update the `listConfig`:
```go
var entityListConfig = domain.ListConfig{
    AllowedSearchFields: []string{"name", "field_name"},   // add here
    AllowedOrderFields:  []string{"created_at", "field_name"},  // add here if orderable
}
```

After handler changes: run `make swagger` to regenerate OpenAPI spec.

## Step 7 — Factory

File: `internal/factory/<entity>.go`

Add the new field to the `NewEntity()` function with realistic fake data:
```go
func NewEntity(orgID uuid.UUID, opts ...func(*domain.Entity)) *domain.Entity {
    e := &domain.Entity{
        // existing...
        FieldName: fake.SomeMethod(),  // add this
    }
    for _, o := range opts {
        o(e)
    }
    return e
}
```

Common fake data methods (using `github.com/brianvoe/gofakeit/v6` or similar):
- `fake.Word()`, `fake.Sentence(N)`, `fake.Name()`
- `fake.IntRange(min, max)`
- `fake.Bool()`
- `fake.UUID()`
- For enums: pick one constant value as default, let opts override

## Step 8 — Seeder

File: `cmd/seed/main.go`

If factory default covers it, no seeder change needed.

If seeder explicitly sets fields (e.g. overriding factory defaults for specific seed scenarios), update the `func(e *domain.Entity)` option lambdas:
```go
entity := factory.NewEntity(orgID, func(e *domain.Entity) {
    e.FieldName = "specific-seed-value"  // add if needed
})
```

## Step 9 — Regenerate frontend client

After all Go changes + `make swagger`:
```bash
cd frontend && bun run generate
```

This regenerates `frontend/src/api/model/*.ts` DTO types and hooks automatically.

## Step 10 — Verify

```bash
make lint        # catch type errors, unused imports
make test        # run unit tests
```

Fix any compilation errors before reporting done.

## Order of operations summary

1. Edit migration SQL in place
2. Update domain struct + DTOs
3. Update service mapping
4. Update repository filters (if needed)
5. Update handler swagger (if new query params)
6. Update factory
7. Update seeder (if needed)
8. `make migrate-reset` → `make swagger` → `cd frontend && bun run generate`
9. `make lint` + `make test`
