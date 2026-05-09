import { describe, expect, it } from "vitest";
import { db } from "../../globals/PrismaClient.js";
import createMultiSearchQuery from "./createMultiSearchQuery.js";

interface InteractionOptions {
  skinName: string;
  maxPrice: number;
  minFloat: number;
  maxFloat: number;
  dmOnly?: boolean;
}

function makeInteraction(input: InteractionOptions) {
  const replies: string[] = [];
  const interaction = {
    options: {
      getString: (name: string) => (name === "skinname" ? input.skinName : null),
      getNumber: (name: string) => {
        if (name === "maxprice") return input.maxPrice;
        if (name === "minfloat") return input.minFloat;
        if (name === "maxfloat") return input.maxFloat;
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

describe("createMultiSearchQuery.ts", () => {
  it("creates a cs trade bot query for valid float/price input", async () => {
    const skinName = `AK-47 | Test Skin ${Date.now()}`;
    const { interaction, replies } = makeInteraction({
      skinName,
      minFloat: 0.05,
      maxFloat: 0.3,
      maxPrice: 99,
      dmOnly: true,
    });

    await createMultiSearchQuery(interaction);

    const created = await db.csTradeBot.findUnique({ where: { name: skinName } });
    expect(created).not.toBeNull();
    expect(created?.maxPrice).toBe(99);
    expect(created?.minFloat).toBe(0.05);
    expect(created?.maxFloat).toBe(0.3);
    expect(replies[0]).toContain("search has been created");

    await db.query.deleteMany({ where: { csTradeBot: { is: { name: skinName } } } });
  });

  it("rejects invalid float ranges", async () => {
    const { interaction, replies } = makeInteraction({
      skinName: "AK-47 | Broken",
      minFloat: 0.8,
      maxFloat: 0.2,
      maxPrice: 40,
    });

    await createMultiSearchQuery(interaction);

    expect(replies[0]).toContain("minimum float cannot be higher");
  });
});
