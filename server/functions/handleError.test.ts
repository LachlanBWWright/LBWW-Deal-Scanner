import { describe, expect, it, vi } from "vitest";
import handleError from "./handleError.js";

describe("handleError.ts", () => {
  it("logs the error and publishes via notification service when provided", () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);
    const publish = vi.fn(async () => undefined);
    const error = new Error("Something broke");

    handleError(error, "UnitTestSource", { publish });

    expect(errorSpy).toHaveBeenCalledTimes(2);
    expect(publish).toHaveBeenCalledTimes(1);
    expect(publish).toHaveBeenCalledWith({
      kind: "error",
      source: "UnitTestSource",
      message: "Something broke",
      stack: error.stack,
    });

    errorSpy.mockRestore();
  });
});
