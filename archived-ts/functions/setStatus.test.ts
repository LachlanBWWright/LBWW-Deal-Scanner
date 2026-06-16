import { beforeEach, describe, expect, it, vi } from "vitest";
import setStatus from "./setStatus.js";
import client from "../globals/DiscordJSClient.js";

const { setPresence } = vi.hoisted(() => ({
  setPresence: vi.fn(),
}));

vi.mock("../globals/DiscordJSClient.js", () => ({
  default: {
    user: {
      setPresence,
    },
  },
}));

describe("setStatus.ts", () => {
  beforeEach(() => {
    setPresence.mockReset();
  });

  it("updates discord rich presence text", () => {
    setStatus("Scanning Gumtree");

    expect(client.user?.setPresence).toBeDefined();
    expect(setPresence).toHaveBeenCalledTimes(1);
    expect(setPresence).toHaveBeenCalledWith({
      status: "online",
      activities: [
        {
          name: "Scanning Gumtree",
          type: 4,
        },
      ],
    });
  });
});
