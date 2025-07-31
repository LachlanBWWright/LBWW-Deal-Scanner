import { describe, it, expect, vi } from "vitest";

// Mock Discord.js types and client
vi.mock("../globals/DiscordJSClient.js", () => ({
  default: {
    channels: {
      fetch: vi.fn(),
    },
  },
}));

import sendToChannel from "../functions/sendToChannel.js";
import client from "../globals/DiscordJSClient.js";
import { ChannelType } from "discord.js";

describe("sendToChannel", () => {
  it("should send message to valid text channel", async () => {
    const mockChannel = {
      type: ChannelType.GuildText,
      send: vi.fn().mockResolvedValue({}),
    };

    vi.mocked(client.channels.fetch).mockResolvedValue(mockChannel as any);

    await sendToChannel("test-channel-id", "Test message");

    expect(client.channels.fetch).toHaveBeenCalledWith("test-channel-id");
    expect(mockChannel.send).toHaveBeenCalledWith({
      content: "Test message",
    });
  });

  it("should send message with files when provided", async () => {
    const mockChannel = {
      type: ChannelType.GuildText,
      send: vi.fn().mockResolvedValue({}),
    };

    vi.mocked(client.channels.fetch).mockResolvedValue(mockChannel as any);

    await sendToChannel("test-channel-id", "Test message", { files: ["image.png"] });

    expect(mockChannel.send).toHaveBeenCalledWith({
      content: "Test message",
      files: ["image.png"],
    });
  });

  it("should throw error when channel is not found", async () => {
    vi.mocked(client.channels.fetch).mockResolvedValue(null);

    await expect(sendToChannel("invalid-channel-id", "Test message"))
      .rejects.toThrow("Channel with id invalid-channel-id not found");
  });

  it("should throw error when channel is not a text channel", async () => {
    const mockChannel = {
      type: ChannelType.GuildVoice, // Not a text channel
    };

    vi.mocked(client.channels.fetch).mockResolvedValue(mockChannel as any);

    await expect(sendToChannel("voice-channel-id", "Test message"))
      .rejects.toThrow("Channel with id voice-channel-id is not a text channel");
  });

  it("should handle empty message", async () => {
    const mockChannel = {
      type: ChannelType.GuildText,
      send: vi.fn().mockResolvedValue({}),
    };

    vi.mocked(client.channels.fetch).mockResolvedValue(mockChannel as any);

    await sendToChannel("test-channel-id", "");

    expect(mockChannel.send).toHaveBeenCalledWith({
      content: "",
    });
  });
});