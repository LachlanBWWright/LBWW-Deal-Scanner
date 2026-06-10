import { beforeEach, describe, expect, it, vi } from "vitest";

const { actionGet } = vi.hoisted(() => ({
  actionGet: vi.fn(),
}));

vi.mock("./actionRegistry.js", () => ({
  actionRegistry: {
    get: actionGet,
    set: vi.fn(),
    delete: vi.fn(),
    cleanupExpired: vi.fn(),
  },
  generateRandomKey: vi.fn(() => "test-key"),
}));

describe("buttonHandler.ts", () => {
  beforeEach(() => {
    actionGet.mockReset();
  });

  it("responds with failure message when action key is missing", async () => {
    actionGet.mockResolvedValue(null);
    const deferReply = vi.fn(async () => undefined);
    const editReply = vi.fn(async () => undefined);

    const { buttonInteractionHandler } = await import("./buttonHandler.js");

    await buttonInteractionHandler({
      customId: "missing-key",
      user: { id: "user-1" },
      deferReply,
      editReply,
    });

    expect(deferReply).toHaveBeenCalledWith({ ephemeral: true });
    expect(editReply).toHaveBeenCalledTimes(1);
    expect(editReply.mock.calls[0]?.[0]?.content).toContain(
      "expired or is invalid",
    );
  });
});
