import { ChannelType } from "discord.js";
import { beforeEach, describe, expect, it, vi } from "vitest";
import sendToChannel from "./sendToChannel.js";

const { fetchChannel } = vi.hoisted(() => ({
  fetchChannel: vi.fn(),
}));

vi.mock("../globals/DiscordJSClient.js", () => ({
  default: {
    channels: {
      fetch: fetchChannel,
    },
    users: {
      fetch: vi.fn(),
    },
  },
}));

describe("sendToChannel.ts", () => {
  beforeEach(() => {
    fetchChannel.mockReset();
  });

  it("logs and returns when channel is missing", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);
    fetchChannel.mockResolvedValue(null);

    await sendToChannel("missing-channel", "hello world");

    expect(fetchChannel).toHaveBeenCalledWith("missing-channel");
    expect(errorSpy).toHaveBeenCalledWith("Channel with id missing-channel not found");
    errorSpy.mockRestore();
  });

  it("logs and returns when channel is not a guild text channel", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);
    fetchChannel.mockResolvedValue({ type: ChannelType.DM });

    await sendToChannel("dm-channel", "hello world");

    expect(errorSpy).toHaveBeenCalledWith(
      "Channel with id dm-channel is not a text channel",
    );
    errorSpy.mockRestore();
  });
});
