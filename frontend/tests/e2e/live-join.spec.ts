import { expect, test } from "@playwright/test"

const LIVE_ID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"

const viewerMe = {
  success: true,
  data: {
    id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
    username: "student",
    name: "Student One",
    is_admin: false,
    role: {
      name: "student",
      permissions: [{ name: "classes:view" }],
    },
  },
}

const activeRoom = {
  success: true,
  data: {
    id: LIVE_ID,
    status: "active",
    class_session: {
      name: "Algebra 101",
      class: {
        name: "Math",
        user: { name: "Prof. Green" },
      },
    },
  },
}

const createdRoom = {
  success: true,
  data: {
    id: LIVE_ID,
    status: "created",
    class_session: {
      name: "Algebra 101",
      class: {
        name: "Math",
        user: { name: "Prof. Green" },
      },
    },
  },
}

test("viewer sees a simplified join card with a Join button", async ({ page }) => {
  const errors: string[] = []
  page.on("pageerror", (error) => errors.push(error.message))

  await page.route("**/api/v1/users/me", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(viewerMe),
    })
  })

  await page.route(`**/api/v1/live-rooms/${LIVE_ID}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(activeRoom),
    })
  })

  await page.goto(`/live/${LIVE_ID}`)

  await expect(page.getByText("Algebra 101")).toBeVisible()
  await expect(page.getByText("Prof. Green")).toBeVisible()
  await expect(page.getByRole("button", { name: /join/i })).toBeVisible()

  // Viewer should see NO microphone or camera aria-label controls
  const mediaControls = page.locator('[aria-label*="microphone" i], [aria-label*="camera" i]')
  await expect(mediaControls).toHaveCount(0)

  expect(errors).toEqual([])
})

test("non-host sees Waiting for host on a created room", async ({ page }) => {
  await page.route("**/api/v1/users/me", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(viewerMe),
    })
  })

  await page.route(`**/api/v1/live-rooms/${LIVE_ID}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(createdRoom),
    })
  })

  await page.goto(`/live/${LIVE_ID}`)

  await expect(page.getByText(/waiting for the host to start/i)).toBeVisible()
})
