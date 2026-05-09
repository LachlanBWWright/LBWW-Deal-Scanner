import { describe, expect, it, vi } from "vitest";
import {
  getFailurePrelude,
  getNotificationPrelude,
  getResponsePrelude,
} from "./messagePreludes.js";

describe("messagePreludes.ts", () => {
  it("returns standard response prelude when random is above rare threshold", () => {
    const randomSpy = vi.spyOn(Math, "random").mockReturnValue(0.5);
    const result = getResponsePrelude();
    expect(
      [
        "Listen up, Jack,",
        "My fellow Americans,",
        "Folks,",
        "Here's the deal,",
      ].includes(result),
    ).toBe(true);
    randomSpy.mockRestore();
  });

  it("returns rare prelude when random is below threshold", () => {
    const randomSpy = vi.spyOn(Math, "random");
    randomSpy.mockReturnValueOnce(0).mockReturnValueOnce(0);
    expect(getResponsePrelude()).toBe("Look, fat,");
    randomSpy.mockReturnValueOnce(0).mockReturnValueOnce(0);
    expect(getFailurePrelude()).toBe("Stupid son of a bitch,");
    randomSpy.mockReturnValueOnce(0).mockReturnValueOnce(0);
    expect(getNotificationPrelude()).toBe(
      "This is a big fu- ...uh... flippable deal,",
    );

    randomSpy.mockRestore();
  });
});
