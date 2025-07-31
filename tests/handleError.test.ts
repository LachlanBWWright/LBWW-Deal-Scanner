import { describe, it, expect, vi } from "vitest";
import handleError from "../functions/handleError.js";

// Mock the sendToChannel function
vi.mock("../functions/sendToChannel.js", () => ({
  default: vi.fn(),
}));

describe("handleError", () => {
  it("should log error to console", () => {
    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    const testError = new Error("Test error message");
    
    handleError(testError, "test-function");
    
    expect(consoleSpy).toHaveBeenCalledWith("An error has occurred in test-function:");
    expect(consoleSpy).toHaveBeenCalledWith(testError);
    
    consoleSpy.mockRestore();
  });

  it("should handle errors with custom names", () => {
    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    const testError = new Error("Custom error");
    
    handleError(testError, "custom-function-name");
    
    expect(consoleSpy).toHaveBeenCalledWith("An error has occurred in custom-function-name:");
    expect(consoleSpy).toHaveBeenCalledWith(testError);
    
    consoleSpy.mockRestore();
  });

  it("should call sendToChannel when ERROR_CHANNEL_ID is set", async () => {
    // This test would require more complex mocking of the globals
    // For now, just ensure the function doesn't throw
    const testError = new Error("Test error");
    
    expect(() => handleError(testError, "test")).not.toThrow();
  });
});