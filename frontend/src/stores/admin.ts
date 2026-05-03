import type { GithubCom4H1RZooraInternalDomainOrganization } from "@/api/model"

import { create } from "zustand"
import { persist } from "zustand/middleware"

type Organization = GithubCom4H1RZooraInternalDomainOrganization

interface AdminState {
  activeOrganization: Organization | null
  activeOrganizationId: string | null
  setActiveOrganization: (org: Organization | null) => void
}

export const useAdminStore = create<AdminState>()(
  persist(
    (set) => ({
      activeOrganization: null,
      activeOrganizationId: null,
      setActiveOrganization: (org) => set({ activeOrganization: org, activeOrganizationId: org?.id ?? null }),
    }),
    { name: "admin" }
  )
)
