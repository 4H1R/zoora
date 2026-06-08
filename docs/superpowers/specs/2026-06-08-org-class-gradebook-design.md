# Org Class Gradebook — Design

**Date:** 2026-06-08
**Scope:** Frontend only. Add a gradebook page to the org class area at
`/org/{orgId}/classes/{classId}/gradebook`. Backend (`internal/gradebook`) already exists.

## Goal

Let teachers/managers manage a class gradebook, and enrolled students view it,
from the org-facing class UI — reusing the existing backend API and matching the
org class page's visual aesthetic (eyebrow labels, `rounded-2xl`, decorative bg).

## Existing assets (reused)

- **API hooks** (auto-generated, `frontend/src/api/gradebook/gradebook.ts`):
  - `useGetClassesIdGradebook(classId)` — full matrix (`columns` + `rows`)
  - `usePostClassesIdGradebookColumns`, `usePutClassesIdGradebookColumnsColumnId`,
    `useDeleteClassesIdGradebookColumnsColumnId`
  - cell upsert via column dialog flow
  - `getGetClassesIdGradebookQueryKey` for invalidation
- **Dialogs** (`components/admin/gradebook/`) — both use *generic* API hooks
  (classes/gradebook/practices/quizzes), not admin-coupled. Reused as-is:
  - `GradebookColumnDialog` — create/edit column, incl. auto-source picker
    (session/practice/quiz)
  - `GradebookCellDialog` — upsert a manual cell value
  - These use `admin.gradebook.*` i18n keys. Left as-is (functional, not worth a
    move for this task). New org-styled wrapper uses fresh `org.class.gradebook.*`
    keys for page chrome.

Note: the existing admin `GradebookMatrixView` is **not** reused — we build an
org-styled view to match the class page. The dialogs it depends on are reused.

## Backend authz (mirrored on frontend — `internal/gradebook/service.go`)

- **View** matrix: admin OR `gradebook:view_any` OR class owner
  (`class.user_id == caller`) OR enrolled member (view-only).
- **Manage** columns/cells (create column, upsert cell, edit column): admin OR
  `gradebook:update_any` OR class owner.
- **Delete** column: admin OR `gradebook:delete_any` OR class owner.

Frontend can't cheaply determine enrollment, so visibility is API-driven (below).

## Architecture

### 1. Route

New file: `frontend/src/routes/_auth/org/$orgId/classes/$classId_.gradebook.tsx`

- Trailing `_` on `$classId_` opts the route OUT of nesting under
  `$classId.tsx` (which renders its own page with no `<Outlet/>`). This is the
  same convention already used by
  `classsessions/$classSessionId_.quizzes.$quizId.take.tsx`.
- Resulting path: `/org/{orgId}/classes/{classId}/gradebook`.
- `head: () => orgHead("org.class.gradebook.title")`.
- Component:
  - guards with `useOrgGuard(["classes:view", "classes:view_any"])` (same as the
    class detail page) — returns null if not allowed.
  - reads `classId`/`orgId` from `Route.useParams()`.
  - fetches class via `useGetClassesId(classId)` (for header name + owner check).
  - renders `<OrgGradebookView classId={classId} cls={cls} />`.
  - back link → `/org/$orgId/classes/$classId`.

### 2. Entry point on class detail page

In `$classId.tsx`, add a **Gradebook** link in the top bar (next to the existing
"back to classes" / shortId row) or class header actions:

- Always rendered for class viewers (any user who can see the page). Per the
  decision, visibility is not pre-gated on gradebook perms; the gradebook page
  itself resolves access via the API.
- `<Link to="/org/$orgId/classes/$classId/gradebook" params={{ orgId, classId }}>`
  styled as an outline `Button` with a `TrophyIcon` (lucide).

### 3. New component — `components/org/classes/OrgGradebookView.tsx`

Props: `{ classId: string; cls?: Class }`.

Responsibilities:
- `useGetClassesIdGradebook(classId)` → matrix.
- **Access resolution (API-driven):**
  - loading → skeleton.
  - `status === 200` → render matrix (grid).
  - `status === 403` (or error with 403) → "no access" empty state
    (`org.class.gradebook.noAccess*` keys).
- **Manage gating** (controls shown only when allowed):
  - compute `canManage = can("gradebook:update_any") || isOwner` where
    `isOwner = !!cls?.user_id && cls.user_id === user.id` (from `useAccess()`).
  - compute `canDelete = can("gradebook:delete_any") || isOwner`.
  - when `!canManage`: matrix is read-only — no "New column" button, no cell
    click-to-edit, no column dropdown menu. Students/viewers see grades only.
- **Visual style** (org aesthetic, matching `$classId.tsx`):
  - section wrapper with `Eyebrow` + `h2` title + stat cells
    (columns count, students count) in a `rounded-2xl ring-1` bar.
  - matrix table inside a `rounded-2xl` card, horizontal scroll, sticky first
    column (student name w/ avatar initials via `getEntityColor`/`getInitials`).
  - auto columns render values as `Badge`; manual columns render plain text and
    (when `canManage`) are click-to-edit.
  - empty state (no columns + no rows) with `TrophyIcon` + hint.
- **Dialogs** (only mounted when `canManage`):
  - `GradebookColumnDialog` (create/edit), `GradebookCellDialog` (cell upsert),
    `DeleteConfirmDialog` (delete column, gated `canDelete`).
  - invalidate `getGetClassesIdGradebookQueryKey(classId)` on success (handled
    by the dialogs already).

### 4. i18n keys (new) — `org.class.gradebook.*`

Add to both `en.json` and `fa.json`:

- `title`, `eyebrow`, `subtitle`
- `open` (link label on class page)
- `newColumn`, `student`
- `stats.columns`, `stats.students`
- `noResults`, `noResultsHint`
- `noAccessTitle`, `noAccessHint`
- `types.*` — reuse mapping for the 6 column types (auto_attendance,
  auto_practice, auto_quiz, manual_grade, manual_attendance, manual_text). If
  duplicating `admin.gradebook.types.*` is undesirable, the view may read the
  type label from a shared spot; default is to add `org.class.gradebook.types.*`.

Use logical CSS props (`ms-`/`me-`/`ps-`/`pe-`/`start`/`end`) for RTL.

## Data flow

```
class page ──[Gradebook link]──▶ /classes/{id}/gradebook route
  route ──useGetClassesId──▶ class (name, owner) ──┐
  route ──renders──▶ OrgGradebookView              │
    OrgGradebookView ──useGetClassesIdGradebook──▶ matrix (200 grid / 403 no-access)
    manage controls gated by can(update_any/delete_any) || owner
    column/cell dialogs ──mutations──▶ invalidate matrix query
```

## Error handling

- 403 on matrix fetch → friendly "no access" state (not a crash).
- Mutation errors → toast (existing dialog behavior).
- Missing class → header falls back to em dash; page still attempts matrix.

## Testing / verification

- `pnpm typecheck` must pass.
- Manual: as class owner — see Gradebook link, open page, create a manual
  column, edit a cell, delete a column. As a non-owner without perms — link
  present, page shows grid read-only or no-access per backend.
- No new backend changes → no Go tests needed.

## Out of scope

- Reordering columns via drag (uses existing `order_index` number field).
- Bulk grade entry / CSV import.
- Any backend change (API already complete).
- Restructuring the class detail page into tabs.
