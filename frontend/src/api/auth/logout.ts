import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { AUTH_TOKEN_KEY } from "../mutator/custom-instance"

export function useLogout() {
  const queryClient = useQueryClient()
  const { t } = useTranslation()
  const navigate = useNavigate()

  return useMutation({
    mutationFn: () => new Promise<void>((resolve) => resolve()),
    onSuccess: () => {
      localStorage.removeItem(AUTH_TOKEN_KEY)
      queryClient.clear()
      toast.success(t("logout.success"))
      navigate({ to: "/" })
    },
  })
}
