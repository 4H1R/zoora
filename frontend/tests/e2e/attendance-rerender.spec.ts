import { expect, test } from "@playwright/test"

// Regression: the org attendance page always mounts DataTablePagination over a
// client-side table. useClientTable used to feed react-table controlled
// `pagination` state with no onPaginationChange and autoReset left on, so reading
// the pagination row model (getRowCount in DataTablePagination) fired react-table's
// internal state setter every render with nothing to absorb it — an unbounded
// re-render loop. Guard it by counting React commits during an idle window.
test("org attendance page does not re-render in a loop at idle", async ({ page }) => {
  const orgId = "44444444-4444-4444-8444-444444444444"

  await page.route("**/api/v1/users/me", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          id: "55555555-5555-4555-8555-555555555555",
          organization_id: orgId,
          username: "teacher",
          name: "Teacher One",
          is_admin: false,
          role: { name: "teacher", permissions: [{ name: "attendance:view" }] },
        },
      }),
    })
  })

  await page.route(`**/api/v1/organizations/${orgId}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: { id: orgId, name: "Demo Organization", status: "active", total_users: 3 },
      }),
    })
  })

  await page.route("**/api/v1/attendance/me**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          summary: { present: 5, absent: 1, late: 2, excused: 0 },
          items: [
            {
              id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
              status: "present",
              created_at: "2026-01-04T00:00:00Z",
              class: { name: "Algebra" },
              class_session: { name: "Session 1" },
            },
          ],
        },
      }),
    })
  })

  await page.goto("/org/attendance?page=1&page_size=20")
  await page.waitForTimeout(1500)

  const patched = await page.evaluate(() => {
    const w = window as unknown as Record<string, unknown>
    w.__commits = 0
    const hook = w.__REACT_DEVTOOLS_GLOBAL_HOOK__ as { onCommitFiberRoot?: (...a: unknown[]) => void } | undefined
    if (!hook) return false
    const orig = hook.onCommitFiberRoot
    hook.onCommitFiberRoot = (...a: unknown[]) => {
      ;(w.__commits as number)++
      return orig?.(...a)
    }
    return true
  })
  expect(patched, "react devtools hook present").toBeTruthy()

  await page.waitForTimeout(3000) // quiet window — no interaction
  const commits = await page.evaluate(() => (window as unknown as { __commits: number }).__commits)
  console.log("COMMITS_during_quiet_3s", commits)
  // Healthy idle page commits a handful of times at most; the loop bug produced 500+.
  expect(commits, "React commits during 3s idle").toBeLessThan(30)
})
