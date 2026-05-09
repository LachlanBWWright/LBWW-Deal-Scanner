import { describe, expect, it } from "vitest";
import { db } from "../../globals/PrismaClient.js";
import createCashQuery from "./createCashQuery.js";

interface InteractionOptions {
  query: string;
  dmOnly?: boolean;
}

function makeInteraction(input: InteractionOptions) {
  const replies: string[] = [];
  const interaction = {
    options: {
      getString: (name: string) => (name === "query" ? input.query : null),
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
    });

    await createCashQuery(interaction);

    const created = await db.cashConverters.findUnique({ where: { url } });
    expect(created).not.toBeNull();
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
