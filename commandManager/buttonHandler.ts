import {
  ButtonInteraction,
  ActionRowBuilder,
  ButtonBuilder,
  ButtonStyle,
} from "discord.js";
import { db } from "../globals/PrismaClient.js";
import {
  getResponsePrelude,
  getFailurePrelude,
} from "../functions/messagePreludes.js";

// Store pending deletions with user ID and timestamp
const pendingDeletions = new Map<
  string,
  { queryType: string; queryId: string; timestamp: number }
>();

export async function buttonInteractionHandler(interaction: ButtonInteraction) {
  try {
    await interaction.deferReply({ ephemeral: true });

    const customId = interaction.customId;

    if (customId.startsWith("delete_query_")) {
      await handleDeleteQuery(interaction);
    } else if (customId.startsWith("confirm_delete_")) {
      await handleConfirmDelete(interaction);
    } else if (customId.startsWith("cancel_delete_")) {
      await handleCancelDelete(interaction);
    }
  } catch (error) {
    console.error("Button interaction error:", error);
    try {
      await interaction.editReply({
        content: `${getFailurePrelude()} An error occurred while processing your request.`,
      });
    } catch (e) {
      console.error("Failed to send error message:", e);
    }
  }
}

async function handleDeleteQuery(interaction: ButtonInteraction) {
  const customId = interaction.customId;
  const parts = customId.split("_");

  if (parts.length < 4) {
    await interaction.editReply({
      content: `${getFailurePrelude()} Invalid button format.`,
    });
    return;
  }

  const queryType = parts[2];
  const queryId = parts.slice(3).join("_"); // Rejoin in case URL contains underscores

  // Store pending deletion
  const pendingKey = `${interaction.user.id}_${Date.now()}`;
  pendingDeletions.set(pendingKey, {
    queryType,
    queryId,
    timestamp: Date.now(),
  });

  // Create confirmation buttons
  const confirmButton = new ButtonBuilder()
    .setCustomId(`confirm_delete_${pendingKey}`)
    .setLabel("Yes, Delete Query")
    .setStyle(ButtonStyle.Danger);

  const cancelButton = new ButtonBuilder()
    .setCustomId(`cancel_delete_${pendingKey}`)
    .setLabel("Cancel")
    .setStyle(ButtonStyle.Secondary);

  const row = new ActionRowBuilder<ButtonBuilder>().addComponents(
    confirmButton,
    cancelButton,
  );

  await interaction.editReply({
    content: `Are you sure you want to delete this ${queryType} query?\n\`${queryId}\`\n\n**This action cannot be undone.**`,
    components: [row],
  });

  // Clean up old pending deletions (older than 10 minutes)
  const tenMinutesAgo = Date.now() - 10 * 60 * 1000;
  for (const [key, value] of pendingDeletions.entries()) {
    if (value.timestamp < tenMinutesAgo) {
      pendingDeletions.delete(key);
    }
  }
}

async function handleConfirmDelete(interaction: ButtonInteraction) {
  const customId = interaction.customId;
  const pendingKey = customId.replace("confirm_delete_", "");

  const pendingDeletion = pendingDeletions.get(pendingKey);
  if (!pendingDeletion) {
    await interaction.editReply({
      content: `${getFailurePrelude()} This deletion request has expired or is invalid.`,
      components: [],
    });
    return;
  }

  const { queryType, queryId } = pendingDeletion;

  try {
    // Delete from database based on query type
    let deleted = false;

    switch (queryType) {
      case "ebay":
        const ebayResult = await db.ebay.delete({
          where: { url: queryId },
        });
        deleted = true;
        break;
      case "gumtree":
        const gumtreeResult = await db.gumtree.delete({
          where: { url: queryId },
        });
        deleted = true;
        break;
      case "cashConverters":
        const cashResult = await db.cashConverters.delete({
          where: { url: queryId },
        });
        deleted = true;
        break;
      case "salvos":
        const salvosResult = await db.salvos.delete({
          where: { name: queryId },
        });
        deleted = true;
        break;
      case "steamMarket":
        const steamResult = await db.steamMarket.delete({
          where: { name: queryId },
        });
        deleted = true;
        break;
      case "csTradeBot":
        const csTradeBotResult = await db.csTradeBot.delete({
          where: { name: queryId },
        });
        deleted = true;
        break;
      case "csMarket":
        const csMarketResult = await db.csMarket.delete({
          where: { url: queryId },
        });
        deleted = true;
        break;
      default:
        throw new Error(`Unknown query type: ${queryType}`);
    }

    pendingDeletions.delete(pendingKey);

    if (deleted) {
      await interaction.editReply({
        content: `${getResponsePrelude()} Successfully deleted the ${queryType} query.`,
        components: [],
      });
    }
  } catch (error: unknown) {
    pendingDeletions.delete(pendingKey);
    console.error("Delete query error:", error);

    // Narrow error shape for Prisma
    const errAny = error as { code?: string; message?: string } | undefined;
    if (errAny?.code === "P2025") {
      // Prisma record not found error
      await interaction.editReply({
        content: `${getFailurePrelude()} Query not found. It may have already been deleted.`,
        components: [],
      });
    } else {
      await interaction.editReply({
        content: `${getFailurePrelude()} Failed to delete query: ${
          errAny?.message || "Unknown error"
        }`,
        components: [],
      });
    }
  }
}

async function handleCancelDelete(interaction: ButtonInteraction) {
  const customId = interaction.customId;
  const pendingKey = customId.replace("cancel_delete_", "");

  pendingDeletions.delete(pendingKey);

  await interaction.editReply({
    content: `Deletion cancelled.`,
    components: [],
  });
}
