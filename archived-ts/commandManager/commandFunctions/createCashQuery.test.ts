import { describe, expect, it } from "vitest";
import { db } from "../../globals/PrismaClient.js";
import createCashQuery from "./createCashQuery.js";

interface InteractionOptions {
  query: string;
  dmOnly?: boolean;
  maxPrice?: number;
  scanMode?: string;
}

function makeInteraction(input: InteractionOptions) {
  const replies: string[] = [];
  const interaction = {
    options: {
      getString: (name: string) => {
        if (name === "query") return input.query;
        if (name === "scanmode") return input.scanMode ?? null;
        return null;
      },
      getNumber: (name: string) =>
        name === "maxprice" ? (input.maxPrice ?? null) : null,
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

describe("createCashQuery.ts", () => {
  it("creates a cash converters query and confirms it", async () => {
    const url = `https://www.cashconverters.com.au/shop/?q=test-${Date.now()}`;
    const { interaction, replies } = makeInteraction({
      query: url,
      dmOnly: true,
      maxPrice: 120,
      scanMode: "siteWide",
    });

    await createCashQuery(interaction);

    const created = await db.cashConverters.findUnique({ where: { url } });
    expect(created).not.toBeNull();
    expect(created?.maxPrice).toBe(120);
    expect(created?.scanMode).toBe("siteWide");
    expect(replies[0]).toContain("search has been created");
    expect(replies[0]).toContain("DM only");

    await db.query.deleteMany({ where: { cashConverters: { is: { url } } } });
  });

  it("rejects non-cash-converters URLs", async () => {
    const { interaction, replies } = makeInteraction({
      query: "https://example.com/search",
    });

    await createCashQuery(interaction);

    expect(replies[0]).toContain("invalid");
  });
});
