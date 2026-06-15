import type { GithubCom4H1RZooraInternalDomainOrganization } from "@/api/model"

import { create } from "zustand"
import { persist } from "zustand/middleware"

type Organization = GithubCom4H1RZooraInternalDomainOrganization

interface AdminState {
  activeOrganization: Organization | null
  activeOrganizationId: string | null
  setActiveOrganization: (org: Organization | null) => void
}

function asUUID(value: unknown): string | null {
  if (typeof value === "string" && value.length > 0) return value
  if (Array.isArray(value) && typeof value[0] === "string") return value[0]
  return null
}

export const useAdminStore = create<AdminState>()(
  persist(
    (set) => ({
      activeOrganization: null,
      activeOrganizationId: null,
      setActiveOrganization: (org) =>
        set({ activeOrganization: org, activeOrganizationId: asUUID(org?.id) }),
    }),
    {
      name: "admin",
      migrate: (state) => {
        const s = state as Partial<AdminState> | undefined
        if (!s) return { activeOrganization: null, activeOrganizationId: null }
        const id = asUUID(s.activeOrganizationId ?? s.activeOrganization?.id)
        const org = s.activeOrganization
          ? { ...s.activeOrganization, id: asUUID(s.activeOrganization.id) ?? undefined }
          : null
        return { activeOrganization: org as Organization | null, activeOrganizationId: id }
      },
      version: 1,
    }
  )
)
