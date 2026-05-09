import puppeteer from "puppeteer";
import { afterEach, describe, expect, it, vi } from "vitest";
import createSCMQuery from "./createSCMQuery.js";

interface InteractionOptions {
  query: string;
  maxPrice?: number;
  dmOnly?: boolean;
}

function makeInteraction(input: InteractionOptions) {
  const replies: string[] = [];
  const interaction = {
    options: {
      getString: (name: string) => (name === "query" ? input.query : null),
      getNumber: (name: string) => (name === "maxprice" ? (input.maxPrice ?? 1) : null),
      getBoolean: (name: string) => (name === "dmonly" ? (input.dmOnly ?? false) : null),
    },
    editReply: async (message: string) => {
      replies.push(message);
      return message;
    },
  };

  return {
    interaction,
    replies,
  };
}

describe("createSCMQuery.ts", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("reports failure when browser startup fails", async () => {
    vi.spyOn(puppeteer, "launch").mockRejectedValueOnce(new Error("launch failed"));
    const { interaction, replies } = makeInteraction({
      query: "AK-47 | Redline",
      maxPrice: 25,
      dmOnly: false,
    });

    await createSCMQuery(interaction);

    expect(replies[0]).toContain("invalid");
  });
});
