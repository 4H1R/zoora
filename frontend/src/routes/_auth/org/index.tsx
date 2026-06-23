import { createFileRoute, redirect } from "@tanstack/react-router"

export const Route = createFileRoute("/_auth/org/")({
  beforeLoad: () => {
    throw redirect({ to: "/org/dashboard" })
  },
})
