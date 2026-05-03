import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"

import { useLogout } from "@/api/auth/logout"
import { AUTH_TOKEN_KEY } from "@/api/mutator/custom-instance"
import { Button } from "@/components/ui/button"

export const Route = createFileRoute("/")({
  component: RouteComponent,
})

function RouteComponent() {
  const { mutate: logout } = useLogout()

  const handleLogout = () => {
    logout()
  }

  return (
    <div>
      <Button onClick={handleLogout}>Logout</Button>
    </div>
  )
}
