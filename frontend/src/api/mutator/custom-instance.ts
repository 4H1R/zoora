import redaxios from "redaxios"

import { clientEnv } from "@/config/env"

export const AUTH_TOKEN_KEY = "access_token"

export const apiClient = redaxios.create({
  baseURL: clientEnv.VITE_API_URL,
  validateStatus: () => true,
})

export const customInstance = async <T>(url: string, init?: RequestInit): Promise<T> => {
  const token = localStorage.getItem(AUTH_TOKEN_KEY)
  const headers = new Headers(init?.headers)
  headers.set("Content-Type", "application/json")
  if (token) headers.set("Authorization", `Bearer ${token}`)

  const res = await apiClient(url, {
    method: (init?.method ?? "GET") as import("redaxios").Options["method"],
    headers: Object.fromEntries(headers.entries()),
    data: init?.body,
  })

  if (res.status >= 400) {
    const err = new Error(`HTTP ${res.status}`) as ErrorType<unknown>
    err.response = res as import("redaxios").Response<unknown>
    throw err
  }

  return {
    data: res.data,
    status: res.status,
    headers: res.headers,
  } as T
}

export type ErrorType<T> = Error & {
  response?: import("redaxios").Response<T>
}
