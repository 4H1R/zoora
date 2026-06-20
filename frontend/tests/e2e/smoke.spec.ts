import { expect, test } from "@playwright/test"

test("renders the app shell without a client-side crash", async ({ page }) => {
  const errors: string[] = []
  page.on("pageerror", (error) => errors.push(error.message))

  await page.goto("/")
  await expect(page.locator("#app")).toBeAttached()
  await expect(page.getByText("Index")).toBeVisible()

  expect(errors).toEqual([])
})

test("renders login and validates required credentials on submit", async ({ page }) => {
  await page.route("**/api/v1/users/me", async (route) => {
    await route.fulfill({
      status: 401,
      contentType: "application/json",
      body: JSON.stringify({ success: false, error: { code: "UNAUTHORIZED", message: "unauthorized" } }),
    })
  })

  await page.goto("/login")

  await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible()
  await expect(page.getByLabel("Username")).toBeVisible()
  await expect(page.locator("#password")).toBeVisible()

  await page.getByRole("button", { name: "Sign in" }).click()

  await expect(page.getByText("Username must be at least 3 characters")).toBeVisible()
  await expect(page.getByText("Password must be at least 8 characters")).toBeVisible()

  await page.getByRole("button", { name: "Show password" }).click()
  await expect(page.locator("#password")).toHaveAttribute("type", "text")
})

test("renders admin dashboard for an authenticated admin", async ({ page }) => {
  await page.route("**/api/v1/users/me", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          id: "11111111-1111-4111-8111-111111111111",
          username: "admin",
          name: "Admin User",
          is_admin: true,
        },
      }),
    })
  })

  await page.route("**/api/v1/admin/organizations/stats", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          total_organizations: 2,
          active_count: 1,
          trial_count: 1,
          suspended_count: 0,
          archived_count: 0,
          deleted_organizations: 0,
          total_users: 4,
        },
      }),
    })
  })

  await page.route("**/api/v1/admin/users**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          items: [
            {
              id: "22222222-2222-4222-8222-222222222222",
              username: "teacher",
              name: "Teacher One",
              is_admin: false,
              created_at: "2026-01-01T00:00:00Z",
              role: { name: "teacher" },
            },
          ],
          total: 1,
          page: 1,
          page_size: 5,
        },
      }),
    })
  })

  await page.route("**/api/v1/admin/organizations**", async (route) => {
    if (route.request().url().includes("/admin/organizations/stats")) {
      await route.fallback()
      return
    }
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          items: [
            {
              id: "33333333-3333-4333-8333-333333333333",
              name: "Demo Org",
              status: "active",
              created_at: "2026-01-02T00:00:00Z",
            },
          ],
          total: 1,
          page: 1,
          page_size: 5,
        },
      }),
    })
  })

  await page.route("**/api/v1/admin/{classes,live-rooms,quizzes,polls,question-banks}**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ success: true, data: { items: [], total: 0, page: 1, page_size: 8 } }),
    })
  })

  await page.goto("/admin/dashboard")

  await expect(page.getByRole("heading", { name: "Platform Overview" })).toBeVisible()
  await expect(page.getByText("Total Members")).toBeVisible()
  await expect(page.getByText("4")).toBeVisible()
  await expect(page.getByText("Demo Org")).toBeVisible()
  await expect(page.getByText("Teacher One")).toBeVisible()
  await expect(page.getByText("Admin User")).toBeVisible()
})

test("renders organization dashboard for an authenticated member", async ({ page }) => {
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
          role: {
            name: "teacher",
            permissions: [
              { name: "classes:view" },
              { name: "classes:create" },
              { name: "quizzes:view" },
              { name: "users:view" },
            ],
          },
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
        data: {
          id: orgId,
          name: "Demo Organization",
          status: "active",
          total_users: 3,
        },
      }),
    })
  })

  await page.route("**/api/v1/classes**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          items: [
            {
              id: "66666666-6666-4666-8666-666666666666",
              name: "Algebra",
              total_users: 20,
              created_at: "2026-01-03T00:00:00Z",
              user: { name: "Teacher One" },
            },
          ],
          total: 1,
          page: 1,
          page_size: 8,
        },
      }),
    })
  })

  await page.route("**/api/v1/quizzes**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          items: [
            {
              id: "77777777-7777-4777-8777-777777777777",
              title: "Quiz One",
              created_at: "2026-01-04T00:00:00Z",
            },
          ],
          total: 1,
          page: 1,
          page_size: 8,
        },
      }),
    })
  })

  await page.route("**/api/v1/users?**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        success: true,
        data: {
          items: [],
          total: 3,
          page: 1,
          page_size: 8,
        },
      }),
    })
  })

  await page.goto(`/org/${orgId}/dashboard`)

  await expect(page.getByRole("heading", { name: /Teacher/i })).toBeVisible()
  await expect(page.getByText("Demo Organization")).toBeVisible()
  await expect(page.getByRole("link", { name: /Algebra/ })).toBeVisible()
  await expect(page.getByText("Quiz One").first()).toBeVisible()
  await expect(page.getByRole("button", { name: /New Class/i })).toBeVisible()
})
