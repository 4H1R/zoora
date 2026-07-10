import { afterEach, describe, expect, it, vi } from "vitest"

import { resolveWsUrl } from "./ws-url"

function stubLocation(protocol: string, host: string) {
  vi.stubGlobal("location", { protocol, host } as Location)
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("resolveWsUrl", () => {
  it("passes absolute ws:// through unchanged", () => {
    stubLocation("https:", "acme.zoora.ir")
    expect(resolveWsUrl("ws://localhost:8080/api/v1/ws")).toBe(
      "ws://localhost:8080/api/v1/ws"
    )
  })

  it("passes absolute wss:// through unchanged", () => {
    stubLocation("http:", "localhost:3000")
    expect(resolveWsUrl("wss://example.test/ws")).toBe("wss://example.test/ws")
  })

  it("derives wss + current host from a relative path on https", () => {
    stubLocation("https:", "acme.zoora.ir")
    expect(resolveWsUrl("/api/v1/ws")).toBe("wss://acme.zoora.ir/api/v1/ws")
  })

  it("derives ws + current host from a relative path on http", () => {
    stubLocation("http:", "acme.localhost:3000")
    expect(resolveWsUrl("/api/v1/ws")).toBe("ws://acme.localhost:3000/api/v1/ws")
  })

  it("prefixes a missing leading slash", () => {
    stubLocation("https:", "acme.zoora.ir")
    expect(resolveWsUrl("api/v1/ws")).toBe("wss://acme.zoora.ir/api/v1/ws")
  })
})
