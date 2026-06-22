import { defineConfig } from "orval"

export default defineConfig({
  zoora: {
    output: {
      mode: "tags-split",
      target: "src/api/backend.ts",
      schemas: "src/api/model",
      client: "react-query",
      override: {
        mutator: {
          path: "./src/api/mutator/custom-instance.ts",
          name: "customInstance",
        },
      },
    },
    input: {
      target: "../docs/swagger.json",
    },
  },
})
