import { describe, expect, it } from "vitest";
import { capExponential } from "./client";

describe("capExponential", () => {
  it("returns the base delay on the first attempt", () => {
    expect(capExponential(0, 500, 10_000)).toBe(500);
  });

  it("doubles with each subsequent attempt", () => {
    expect(capExponential(1, 500, 10_000)).toBe(1000);
    expect(capExponential(2, 500, 10_000)).toBe(2000);
    expect(capExponential(3, 500, 10_000)).toBe(4000);
  });

  it("caps at the configured maximum", () => {
    expect(capExponential(10, 500, 10_000)).toBe(10_000);
  });
});
