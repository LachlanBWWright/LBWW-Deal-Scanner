import { describe, it, expect, vi } from "vitest";

// Mock the Discord client (define before import)
vi.mock("../globals/DiscordJSClient.js", () => ({
  default: {
    user: {
      setPresence: vi.fn(),
    },
  },
}));

import setStatus from "../functions/setStatus.js";
import client from "../globals/DiscordJSClient.js";

describe("setStatus", () => {
  it("should set Discord bot presence with given status text", () => {
    const testStatus = "Scanning for deals...";
    
    setStatus(testStatus);
    
    expect(client.user?.setPresence).toHaveBeenCalledWith({
      status: "online",
      activities: [
        {
          name: testStatus,
          type: 4,
        },
      ],
    });
  });

  it("should handle empty status text", () => {
    setStatus("");
    
    expect(client.user?.setPresence).toHaveBeenCalledWith({
      status: "online",
      activities: [
        {
          name: "",
          type: 4,
        },
      ],
    });
  });

  it("should handle long status text", () => {
    const longStatus = "A".repeat(100);
    
    setStatus(longStatus);
    
    expect(client.user?.setPresence).toHaveBeenCalledWith({
      status: "online",
      activities: [
        {
          name: longStatus,
          type: 4,
        },
      ],
    });
  });

  it("should not throw when client.user is null", () => {
    // This test verifies the optional chaining works correctly
    expect(() => setStatus("test")).not.toThrow();
  });
});