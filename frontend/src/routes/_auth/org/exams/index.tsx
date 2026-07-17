import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

import { ExamsHubPage } from "@/components/org/exams/ExamsHubPage"
import { adminSearchSchema } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

// status (from adminSearchSchema) carries the student state tab;
// class_id/class_session_id drive the server-side filters.
const examsSearchSchema = adminSearchSchema.extend({
  class_id: z.string().optional(),
  class_session_id: z.string().optional(),
})

export const Route = createFileRoute("/_auth/org/exams/")({
  head: () => orgHead("org.nav.exams"),
  validateSearch: examsSearchSchema,
  component: ExamsHubPage,
})
