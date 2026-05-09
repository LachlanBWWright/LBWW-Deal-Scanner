import { describe, expect, it } from "vitest";
import { db } from "../../globals/PrismaClient.js";
import createCSMarket from "./createCSMarket.js";

interface InteractionOptions {
  query: string;
  maxFloat?: number;
  maxPrice?: number;
  dmOnly?: boolean;
}

function makeInteraction(input: InteractionOptions) {
  const replies: string[] = [];
  const interaction = {
    options: {
      getString: (name: string) => (name === "query" ? input.query : null),
      getNumber: (name: string) => {
        if (name === "maxfloat") return input.maxFloat ?? 1;
        if (name === "maxprice") return input.maxPrice ?? 1;
        return null;
      },
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

describe("createCSMarket.ts", () => {
  it("creates a cs market query for steam listing URLs", async () => {
    const query = `https://steamcommunity.com/market/listings/730/AK-47%20%7C%20Redline%20%28Field-Tested%29-${Date.now()}`;
    const { interaction, replies } = makeInteraction({
      query,
      maxFloat: 0.4,
      maxPrice: 120,
      dmOnly: true,
    });

    await createCSMarket(interaction);

    const created = await db.csMarket.findFirst({ where: { displayUrl: query } });
    expect(created).not.toBeNull();
    expect(created?.maxPrice).toBe(120);
    expect(created?.maxFloat).toBe(0.4);
    expect(replies[0]).toContain("search has been created");
    expect(replies[0]).toContain("DM only");

    await db.query.deleteMany({ where: { csMarket: { is: { displayUrl: query } } } });
  });

  it("rejects non-steam listing URLs", async () => {
    const { interaction, replies } = makeInteraction({
      query: "https://example.com/not-a-steam-listing",
    });

    await createCSMarket(interaction);

    expect(replies[0]).toContain("invalid");
  });
});
