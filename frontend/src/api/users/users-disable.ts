// Hand-written mutation hooks for the user disable/enable endpoints.
//
// These mirror orval's generated hook surface but are maintained by hand
// because the endpoints could not be regenerated at the time of writing.
// Replace with the generated hooks (usePostUsersIdDisable, etc.) once the
// OpenAPI spec is regenerated via `pnpm generate`.
import type { GithubCom4H1RZooraInternalDomainUser as User } from "@/api/model"
import type { ErrorType } from "@/api/mutator/custom-instance"
import type { UseMutationOptions } from "@tanstack/react-query"

import { useMutation } from "@tanstack/react-query"

import { customInstance } from "@/api/mutator/custom-instance"

export interface DisableUserResponse {
  data: { success?: boolean; data?: User }
  status: number
}

export interface DisableUserVariables {
  id: string
  reason?: string
}

export interface EnableUserVariables {
  id: string
}

const disableRequest = (base: string, { id, reason }: DisableUserVariables) =>
  customInstance<DisableUserResponse>(`${base}/${id}/disable`, {
    method: "POST",
    body: JSON.stringify({ reason: reason ?? "" }),
  })

const enableRequest = (base: string, { id }: EnableUserVariables) =>
  customInstance<DisableUserResponse>(`${base}/${id}/enable`, { method: "POST" })

type DisableOptions<TError> = UseMutationOptions<DisableUserResponse, TError, DisableUserVariables>
type EnableOptions<TError> = UseMutationOptions<DisableUserResponse, TError, EnableUserVariables>

// Org-panel surface (/users/:id).
export const useDisableUser = <TError = ErrorType<unknown>>(options?: DisableOptions<TError>) =>
  useMutation<DisableUserResponse, TError, DisableUserVariables>({
    mutationFn: (variables) => disableRequest("/users", variables),
    ...options,
  })

export const useEnableUser = <TError = ErrorType<unknown>>(options?: EnableOptions<TError>) =>
  useMutation<DisableUserResponse, TError, EnableUserVariables>({
    mutationFn: (variables) => enableRequest("/users", variables),
    ...options,
  })

// Admin-panel surface (/admin/users/:id).
export const useDisableAdminUser = <TError = ErrorType<unknown>>(options?: DisableOptions<TError>) =>
  useMutation<DisableUserResponse, TError, DisableUserVariables>({
    mutationFn: (variables) => disableRequest("/admin/users", variables),
    ...options,
  })

export const useEnableAdminUser = <TError = ErrorType<unknown>>(options?: EnableOptions<TError>) =>
  useMutation<DisableUserResponse, TError, EnableUserVariables>({
    mutationFn: (variables) => enableRequest("/admin/users", variables),
    ...options,
  })
