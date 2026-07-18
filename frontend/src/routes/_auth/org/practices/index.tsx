import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

import { PracticesHubPage } from "@/components/org/practices/hub/PracticesHubPage"
import { adminSearchSchema } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

const practicesSearchSchema = adminSearchSchema.extend({
  window: z.enum(["upcoming", "open", "ended"]).optional(),
  needs_grading: z.boolean().optional(),
  class_id: z.string().optional(),
  class_session_id: z.string().optional(),
})

export const Route = createFileRoute("/_auth/org/practices/")({
  head: () => orgHead("org.nav.practices"),
  validateSearch: practicesSearchSchema,
  component: PracticesHubPage,
})
