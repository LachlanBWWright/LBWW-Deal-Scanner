import { ActionRowBuilder, ButtonBuilder, ButtonStyle } from "discord.js";
import { db } from "../globals/PrismaClient.js";
import { actionRegistry, generateRandomKey } from "./actionRegistry.js";
import {
  getResponsePrelude,
  getFailurePrelude,
} from "../functions/messagePreludes.js";
import { resultAsync } from "../functions/neverthrowUtils.js";

interface ButtonInteractionInput {
  customId: string;
  user: {
    id: string;
  };
  deferReply(options: { ephemeral: boolean }): Promise<unknown>;
  editReply(options: {
    content: string;
    components?: unknown[];
  }): Promise<unknown>;
}

// Global action registry for button actions
// Use shared actionRegistry and generateRandomKey from actionRegistry.ts

function isPrismaError(error: unknown): error is { code: unknown } {
  return typeof error === "object" && error !== null && "code" in error;
}

function isErrorWithMessage(error: unknown): error is { message: string } {
  if (typeof error !== "object" || error === null) return false;
  if (!("message" in error)) return false;

  if (isRecord(error)) {
    return typeof error.message === "string";
  }
  return false;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export async function buttonInteractionHandler(
  interaction: ButtonInteractionInput,
) {
  const handlerResult = await resultAsync(async () => {
    await interaction.deferReply({ ephemeral: true });

    const customId = interaction.customId;

    // customId is now the short action key
    const actionKey = customId;
    const action = await actionRegistry.get(actionKey);
    if (!action) {
      await interaction.editReply({
        content: `${getFailurePrelude()} This action has expired or is invalid.`,
      });
      return;
    }

    // Dispatch based on stored action type
    if (action.type === "delete") {
      // For delete we open a confirmation dialog (create confirm/cancel actions)
      await handleDeleteQuery(interaction, actionKey);
    } else if (action.type === "confirm_delete") {
      await handleConfirmDelete(interaction, actionKey);
    } else if (action.type === "cancel_delete") {
      await handleCancelDelete(interaction);
    } else if (action.type === "subscribe_dm") {
      await handleSubscribeDM(interaction, actionKey);
    } else if (action.type === "unsubscribe_dm") {
      await handleUnsubscribeDM(interaction, actionKey);
    }
  }, "Button interaction error");

  if (handlerResult.isErr()) {
    console.error("Button interaction error:", handlerResult.error.message);
    const replyResult = await resultAsync(
      () =>
        interaction.editReply({
          content: `${getFailurePrelude()} An error occurred while processing your request.`,
        }),
      "Failed to send error message",
    );
    if (replyResult.isErr()) {
      console.error("Failed to send error message:", replyResult.error.message);
    }
  }
}

async function handleDeleteQuery(
  interaction: ButtonInteractionInput,
  actionKey: string,
) {
  const action = await actionRegistry.get(actionKey);
  if (!action) {
    await interaction.editReply({
      content: `${getFailurePrelude()} Invalid or expired action.`,
      components: [],
    });
    return;
  }

  const { queryType, queryId } = action;

  // Create confirmation buttons which reference new short keys
  const confirmKey = generateRandomKey();
  const cancelKey = generateRandomKey();
  const now = Date.now();

  await actionRegistry.set(confirmKey, {
    type: "confirm_delete",
    queryType: action.queryType,
    queryId: action.queryId,
    userId: interaction.user.id,
    timestamp: now,
    relatedKey: actionKey,
  });

  await actionRegistry.set(cancelKey, {
    type: "cancel_delete",
    queryType: action.queryType,
    queryId: action.queryId,
    userId: interaction.user.id,
    timestamp: now,
    relatedKey: actionKey,
  });

  const confirmButton = new ButtonBuilder()
    .setCustomId(confirmKey)
    .setLabel("Yes, Delete Query")
    .setStyle(ButtonStyle.Danger);

  const cancelButton = new ButtonBuilder()
    .setCustomId(cancelKey)
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

  // Clean up old actions (older than 10 minutes)
  await actionRegistry.cleanupExpired(10 * 60 * 1000);
}

async function handleConfirmDelete(
  interaction: ButtonInteractionInput,
  actionKey: string,
) {
  const actionData = await actionRegistry.get(actionKey);
  if (!actionData || actionData.type !== "confirm_delete") {
    await interaction.editReply({
      content: `${getFailurePrelude()} This deletion request has expired or is invalid.`,
      components: [],
    });
    return;
  }

  const { queryType, queryId } = actionData;

  const deleteResult = await resultAsync(async () => {
    // Delete from database based on query type
    let deleted = false;

    switch (queryType) {
      case "ebay": {
        await db.ebay.delete({
          where: { url: queryId },
        });
        deleted = true;
        break;
      }
      case "gumtree": {
        await db.gumtree.delete({
          where: { url: queryId },
        });
        deleted = true;
        break;
      }
      case "cashConverters": {
        await db.cashConverters.delete({
          where: { url: queryId },
        });
        deleted = true;
        break;
      }
      case "salvos": {
        await db.salvos.delete({
          where: { name: queryId },
        });
        deleted = true;
        break;
      }
      case "steamMarket": {
        await db.steamMarket.delete({
          where: { name: queryId },
        });
        deleted = true;
        break;
      }
      case "csTradeBot": {
        await db.csTradeBot.delete({
          where: { name: queryId },
        });
        deleted = true;
        break;
      }
      case "csMarket": {
        await db.csMarket.delete({
          where: { url: queryId },
        });
        deleted = true;
        break;
      }
      default: {
        await interaction.editReply({
          content: `${getFailurePrelude()} Unknown query type: ${queryType}`,
          components: [],
        });
        return;
      }
    }

    // Remove the confirm action and its related pending action
    await actionRegistry.delete(actionKey);
    if (actionData.relatedKey)
      await actionRegistry.delete(actionData.relatedKey);

    if (deleted) {
      await interaction.editReply({
        content: `${getResponsePrelude()} Successfully deleted the ${queryType} query.`,
        components: [],
      });
    }
  }, "Delete query error");

  if (deleteResult.isErr()) {
    await actionRegistry.delete(actionKey);
    const error = deleteResult.error;
    console.error("Delete query error:", error.message);

    // Narrow error shape for Prisma
    if (isPrismaError(error) && error.code === "P2025") {
      // Prisma record not found error
      await interaction.editReply({
        content: `${getFailurePrelude()} Query not found. It may have already been deleted.`,
        components: [],
      });
    } else {
      const message = isErrorWithMessage(error)
        ? error.message
        : "Unknown error";
      await interaction.editReply({
        content: `${getFailurePrelude()} Failed to delete query: ${message}`,
        components: [],
      });
    }
  }
}

async function handleCancelDelete(interaction: ButtonInteractionInput) {
  const actionKey = interaction.customId;
  const action = await actionRegistry.get(actionKey);
  if (action?.relatedKey) await actionRegistry.delete(action.relatedKey);
  await actionRegistry.delete(actionKey);

  await interaction.editReply({
    content: `Deletion cancelled.`,
    components: [],
  });
}

async function handleSubscribeDM(
  interaction: ButtonInteractionInput,
  actionKey: string,
) {
  const action = await actionRegistry.get(actionKey);
  if (!action) {
    await interaction.editReply({
      content: `${getFailurePrelude()} Invalid or expired action.`,
      components: [],
    });
    return;
  }

  const { queryType, queryId } = action;

  const subscribeResult = await resultAsync(async () => {
    // Get the Query record for this specific query
    const query = await getQueryByTypeAndId(queryType, queryId);
    if (!query?.queryId) {
      await interaction.editReply({
        content: `${getFailurePrelude()} Query not found.`,
      });
      return;
    }

    // Check if user is already subscribed
    const existing = await db.userQuery.findUnique({
      where: {
        userId_queryId: {
          userId: interaction.user.id,
          queryId: query.queryId,
        },
      },
    });

    if (existing) {
      await interaction.editReply({
        content: `${getResponsePrelude()} You are already subscribed to DMs for this query.`,
      });
      return;
    }

    // Create subscription
    await db.userQuery.create({
      data: {
        userId: interaction.user.id,
        queryId: query.queryId,
        queryType,
      },
    });

    await actionRegistry.delete(actionKey);

    await interaction.editReply({
      content: `${getResponsePrelude()} You have been subscribed to DM notifications for this ${queryType} query. You will receive a DM when new items are found.`,
    });
  }, "Subscribe DM error");

  if (subscribeResult.isErr()) {
    const error = subscribeResult.error;
    console.error("Subscribe DM error:", error.message);
    const message = isErrorWithMessage(error) ? error.message : "Unknown error";
    await interaction.editReply({
      content: `${getFailurePrelude()} Failed to subscribe: ${message}`,
    });
  }
}

async function handleUnsubscribeDM(
  interaction: ButtonInteractionInput,
  actionKey: string,
) {
  const action = await actionRegistry.get(actionKey);
  if (!action) {
    await interaction.editReply({
      content: `${getFailurePrelude()} Invalid or expired action.`,
      components: [],
    });
    return;
  }

  const { queryType, queryId } = action;

  const unsubscribeResult = await resultAsync(async () => {
    // Get the Query record for this specific query
    const query = await getQueryByTypeAndId(queryType, queryId);
    if (!query?.queryId) {
      await interaction.editReply({
        content: `${getFailurePrelude()} Query not found.`,
      });
      return;
    }

    // Delete subscription
    await db.userQuery.delete({
      where: {
        userId_queryId: {
          userId: interaction.user.id,
          queryId: query.queryId,
        },
      },
    });

    await actionRegistry.delete(actionKey);

    await interaction.editReply({
      content: `${getResponsePrelude()} You have been unsubscribed from DM notifications for this query.`,
    });
  }, "Unsubscribe DM error");

  if (unsubscribeResult.isErr()) {
    const error = unsubscribeResult.error;
    console.error("Unsubscribe DM error:", error.message);

    if (isPrismaError(error) && error.code === "P2025") {
      // Prisma record not found error
      await interaction.editReply({
        content: `${getFailurePrelude()} You are not subscribed to this query.`,
      });
    } else {
      const message = isErrorWithMessage(error)
        ? error.message
        : "Unknown error";
      await interaction.editReply({
        content: `${getFailurePrelude()} Failed to unsubscribe: ${message}`,
      });
    }
  }
}

async function getQueryByTypeAndId(queryType: string, queryId: string) {
  switch (queryType) {
    case "salvos":
      return await db.salvos.findUnique({
        where: { name: queryId },
        include: { query: true },
      });
    case "ebay":
      return await db.ebay.findUnique({
        where: { url: queryId },
        include: { query: true },
      });
    case "gumtree":
      return await db.gumtree.findUnique({
        where: { url: queryId },
        include: { query: true },
      });
    case "cashConverters":
      return await db.cashConverters.findUnique({
        where: { url: queryId },
        include: { query: true },
      });
    case "steamMarket":
      return await db.steamMarket.findUnique({
        where: { name: queryId },
        include: { query: true },
      });
    case "csTradeBot":
      return await db.csTradeBot.findUnique({
        where: { name: queryId },
        include: { query: true },
      });
    case "csMarket":
      return await db.csMarket.findUnique({
        where: { url: queryId },
        include: { query: true },
      });
    default:
      return null;
  }
}
