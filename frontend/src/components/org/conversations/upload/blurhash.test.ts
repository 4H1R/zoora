import { describe, expect, it } from "vitest"

import { blurhashComponents } from "./blurhash"

// The canvas/decode paths of encodeBlurhash + imageDimensions rely on real
// pixel rendering, which jsdom does not implement — those are intentionally
// guarded (return null) rather than unit-tested. We only pin the pure bits.
describe("blurhashComponents", () => {
  it("uses a 4x3 component grid", () => {
    expect(blurhashComponents()).toEqual([4, 3])
  })
})
