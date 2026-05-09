import { describe, expect, it } from "vitest";
import { db } from "../../globals/PrismaClient.js";
import createEbayQuery from "./createEbayQuery.js";

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
      getNumber: (name: string) =>
        name === "maxprice" ? (input.maxPrice ?? 1000) : null,
      getBoolean: (name: string) =>
        name === "dmonly" ? (input.dmOnly ?? false) : null,
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

describe("createEbayQuery.ts", () => {
  it("creates an eBay query and confirms it", async () => {
    const url = `https://www.ebay.com.au/sch/i.html?_nkw=test-${Date.now()}`;
    const { interaction, replies } = makeInteraction({
      query: url,
      maxPrice: 321,
      dmOnly: true,
    });

    await createEbayQuery(interaction);

    const created = await db.ebay.findUnique({ where: { url } });
    expect(created).not.toBeNull();
    expect(created?.maxPrice).toBe(321);
    expect(replies[0]).toContain("search has been created");
    expect(replies[0]).toContain("DM only");

    await db.query.deleteMany({ where: { ebay: { is: { url } } } });
  });

  it("rejects non-ebay URLs", async () => {
    const { interaction, replies } = makeInteraction({
      query: "https://example.com/search",
    });

    await createEbayQuery(interaction);

    expect(replies[0]).toContain("invalid");
  });
});
