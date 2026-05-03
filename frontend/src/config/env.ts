import { z } from "zod"

const clientEnvSchema = z.object({
  VITE_API_URL: z.string().url(),
})

export const clientEnv = clientEnvSchema.parse(import.meta.env)
