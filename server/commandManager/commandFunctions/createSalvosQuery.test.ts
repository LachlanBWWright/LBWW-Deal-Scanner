import { describe, expect, it } from "vitest";
import { db } from "../../globals/PrismaClient.js";
import createSalvosQuery from "./createSalvosQuery.js";

interface InteractionOptions {
  query: string | null;
  minPrice?: number;
  maxPrice?: number;
  dmOnly?: boolean;
}

function makeInteraction(input: InteractionOptions) {
  const replies: string[] = [];
  const interaction = {
    options: {
      getString: (name: string) => (name === "query" ? input.query : null),
      getNumber: (name: string) => {
        if (name === "minprice") return input.minPrice ?? 0;
        if (name === "maxprice") return input.maxPrice ?? 99999;
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

describe("createSalvosQuery.ts", () => {
  it("creates a salvos query and confirms it", async () => {
    const query = `test-salvos-${Date.now()}`;
    const { interaction, replies } = makeInteraction({
      query,
      minPrice: 10,
      maxPrice: 30,
      dmOnly: true,
    });

    await createSalvosQuery(interaction);

    const created = await db.salvos.findUnique({ where: { name: query } });
    expect(created).not.toBeNull();
    expect(created?.minPrice).toBe(10);
    expect(created?.maxPrice).toBe(30);
    expect(replies[0]).toContain("search has been created");
    expect(replies[0]).toContain("DM only");

    await db.query.deleteMany({ where: { salvos: { is: { name: query } } } });
  });

  it("rejects URL input for salvos query text", async () => {
    const { interaction, replies } = makeInteraction({
      query: "https://www.salvosstores.com.au/search?search=chair",
    });

    await createSalvosQuery(interaction);

    expect(replies[0]).toContain("Do not enter a URL");
  });
});
