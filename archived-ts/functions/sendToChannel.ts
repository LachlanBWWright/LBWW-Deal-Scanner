import {
  ChannelType,
  ActionRowBuilder,
  ButtonBuilder,
  ButtonStyle,
  MessagePayload,
  MessagePayloadOption,
} from "discord.js";
import client from "../globals/DiscordJSClient.js";
import {
  actionRegistry,
  generateRandomKey,
} from "../commandManager/actionRegistry.js";
import { db } from "../globals/PrismaClient.js";
import { resultAsync } from "./neverthrowUtils.js";

interface sendToChannelOptions {
  files?: string[];
  queryId?: string;
  queryType?: string;
}

export default async function sendToChannel(
  channelId: string,
  message: string,
  { files, queryId, queryType }: sendToChannelOptions = {}, //Optional param for embedding images
) {
  // First, send DMs to subscribed users
  if (queryId && queryType) {
    await sendDMsToSubscribedUsers(queryId, queryType, message, files);

    // Check if this query is DM-only
    const query = await getQueryByTypeAndId(queryType, queryId);
    if (query?.query?.dmOnly) {
      // Don't send to channel, only DMs
      return;
    }
  }

  const channel = await client.channels.fetch(channelId);

  if (!channel) {
    console.error(`Channel with id ${channelId} not found`);
    return;
  }

  if (channel.type !== ChannelType.GuildText) {
    console.error(`Channel with id ${channelId} is not a text channel`);
    return;
  }

  const messageOptions: MessagePayloadOption = {
    content: message,
    ...(files && { files: files }),
  };

  // Add buttons if query info is provided
  if (queryId && queryType) {
    const deleteActionKey = generateRandomKey();
    const subscribeDMActionKey = generateRandomKey();
    const now = Date.now();

    await actionRegistry.set(deleteActionKey, {
      type: "delete",
      queryType,
      queryId,
      timestamp: now,
    });

    await actionRegistry.set(subscribeDMActionKey, {
      type: "subscribe_dm",
      queryType,
      queryId,
      timestamp: now,
    });

    const deleteButton = new ButtonBuilder()
      .setCustomId(deleteActionKey)
      .setLabel("Delete Query")
      .setStyle(ButtonStyle.Danger);

    const subscribeDMButton = new ButtonBuilder()
      .setCustomId(subscribeDMActionKey)
      .setLabel("Subscribe to DM")
      .setStyle(ButtonStyle.Primary);

    const row = new ActionRowBuilder<ButtonBuilder>().addComponents(
      deleteButton,
      subscribeDMButton,
    );

    messageOptions.components = [row];

    // cleanup old actions (10 minutes)
    await actionRegistry.cleanupExpired(10 * 60 * 1000);
  }

  const messagePayload = new MessagePayload(channel, messageOptions); // Create a new MessagePayload instance to ensure proper formatting

  //https://discord.js.org/docs/packages/discord.js/14.14.1/BaseGuildTextChannel:Class#send
  await channel.send(messagePayload);
}

async function sendDMsToSubscribedUsers(
  queryId: string,
  queryType: string,
  message: string,
  files?: string[],
) {
  // Get the actual Query record to find subscribed users
  const query = await getQueryByTypeAndId(queryType, queryId);
  if (!query?.queryId) return;

  const userQueries = await db.userQuery.findMany({
    where: { queryId: query.queryId },
  });

  for (const userQuery of userQueries) {
    const dmResult = await resultAsync(async () => {
      const user = await client.users.fetch(userQuery.userId);

      const unsubscribeActionKey = generateRandomKey();
      await actionRegistry.set(unsubscribeActionKey, {
        type: "unsubscribe_dm",
        queryType,
        queryId,
        timestamp: Date.now(),
      });

      const unsubscribeButton = new ButtonBuilder()
        .setCustomId(unsubscribeActionKey)
        .setLabel("Unsubscribe from DM")
        .setStyle(ButtonStyle.Secondary);

      const row = new ActionRowBuilder<ButtonBuilder>().addComponents(
        unsubscribeButton,
      );

      await user.send({
        content: message,
        files: files,
        components: [row],
      });
    }, `Failed to send DM to user ${userQuery.userId}`);

    if (dmResult.isErr()) {
      console.error(dmResult.error.message);
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
