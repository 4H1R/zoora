import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { Eye, EyeOff } from "lucide-react"
import { useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { usePostAuthLogin } from "@/api/auth/auth"
import { AUTH_TOKEN_KEY } from "@/api/mutator/custom-instance"
import { getGetUsersMeQueryKey } from "@/api/users/users"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { InputGroup, InputGroupButton, InputGroupInput } from "@/components/ui/input-group"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

const loginSchema = z.object({
  username: z.string().min(3),
  password: z.string().min(8),
})

type LoginFormValues = z.infer<typeof loginSchema>

export function LoginForm({ className, ...props }: React.ComponentProps<"div">) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const loginMutation = usePostAuthLogin()
  const [showPassword, setShowPassword] = useState(false)

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: "", password: "" },
  })

  const onSubmit = handleSubmit((values) => {
    loginMutation.mutate(
      { data: values },
      {
        onSuccess: (response) => {
          if (response.status === 200 && response.data.data?.token) {
            localStorage.setItem(AUTH_TOKEN_KEY, response.data.data.token)

            const user = response.data.data.user

            toast.success(t("login.success"))

            if (user?.is_admin) {
              navigate({ to: "/admin/dashboard" })
            } else if (user?.organization_id) {
              navigate({ to: "/org/dashboard" })
            }

            queryClient.invalidateQueries({ queryKey: getGetUsersMeQueryKey() })
          }
        },
        onError: () => {
          setError("username", { message: t("login.error") })
        },
      }
    )
  })

  const isPending = isSubmitting || loginMutation.isPending

  return (
    <div className={cn("flex flex-col", className)} {...props}>
      <h1 className="text-2xl font-semibold tracking-tight">{t("login.title")}</h1>
      <p className="text-muted-foreground mt-1.5 text-sm">{t("login.subtitle")}</p>

      <form onSubmit={onSubmit} noValidate className="mt-6">
        <FieldGroup>
          <Field data-invalid={!!errors.username || undefined}>
            <FieldLabel htmlFor="username" className="text-xs">
              {t("login.username")}
            </FieldLabel>
            <Input
              id="username"
              type="text"
              autoComplete="username"
              placeholder={t("login.usernamePlaceholder")}
              className="h-10.5"
              aria-invalid={!!errors.username}
              {...register("username")}
            />
            {errors.username && <FieldError>{errors.username.message}</FieldError>}
          </Field>

          <Field data-invalid={!!errors.password || undefined}>
            <FieldLabel htmlFor="password" className="text-xs">
              {t("login.password")}
            </FieldLabel>
            <InputGroup className="h-10.5">
              <InputGroupInput
                id="password"
                type={showPassword ? "text" : "password"}
                autoComplete="current-password"
                placeholder={t("login.passwordPlaceholder")}
                aria-invalid={!!errors.password}
                {...register("password")}
              />
              <InputGroupButton
                size="icon-xs"
                variant="ghost"
                onClick={() => setShowPassword(!showPassword)}
                aria-label={showPassword ? t("login.hidePassword") : t("login.showPassword")}
              >
                {showPassword ? <EyeOff /> : <Eye />}
              </InputGroupButton>
            </InputGroup>
            {errors.password && <FieldError>{errors.password.message}</FieldError>}
          </Field>

          <Button type="submit" disabled={isPending} className="mt-2 h-10.5 w-full text-sm font-semibold">
            {isPending && <Spinner />}
            {isPending ? t("login.submitting") : t("login.submit")}
          </Button>
        </FieldGroup>
      </form>
    </div>
  )
}
