import { describe, expect, it } from "vitest";
import { db } from "../../globals/PrismaClient.js";
import createGumtreeQuery from "./createGumtreeQuery.js";

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

describe("createGumtreeQuery.ts", () => {
  it("creates a gumtree query and confirms it", async () => {
    const url = `https://www.gumtree.com.au/s-test/k0?query=test-${Date.now()}`;
    const { interaction, replies } = makeInteraction({
      query: url,
      maxPrice: 111,
      dmOnly: true,
    });

    await createGumtreeQuery(interaction);

    const created = await db.gumtree.findUnique({ where: { url } });
    expect(created).not.toBeNull();
    expect(created?.maxPrice).toBe(111);
    expect(replies[0]).toContain("search has been created");
    expect(replies[0]).toContain("DM only");

    await db.query.deleteMany({ where: { gumtree: { is: { url } } } });
  });

  it("rejects non-gumtree URLs", async () => {
    const { interaction, replies } = makeInteraction({
      query: "https://example.com/search",
    });

    await createGumtreeQuery(interaction);

    expect(replies[0]).toContain("invalid");
  });
});
