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
  const channel = await client.channels.fetch(channelId);

  if (!channel) throw new Error(`Channel with id ${channelId} not found`);

  if (channel.type !== ChannelType.GuildText)
    throw new Error(`Channel with id ${channelId} is not a text channel`);

  const messageOptions: MessagePayloadOption = {
    content: message,
    ...(files && { files: files }),
  };

  // Add delete button if query info is provided
  if (queryId && queryType) {
    const actionKey = generateRandomKey();
    const now = Date.now();
    await actionRegistry.set(actionKey, {
      type: "delete",
      queryType,
      queryId,
      timestamp: now,
    });

    const deleteButton = new ButtonBuilder()
      .setCustomId(actionKey)
      .setLabel("Delete Query")
      .setStyle(ButtonStyle.Danger);

    const row = new ActionRowBuilder<ButtonBuilder>().addComponents(
      deleteButton,
    );

    messageOptions.components = [row];

    // cleanup old actions (10 minutes)
    await actionRegistry.cleanupExpired(10 * 60 * 1000);
  }

  const messagePayload = new MessagePayload(channel, messageOptions); // Create a new MessagePayload instance to ensure proper formatting

  //https://discord.js.org/docs/packages/discord.js/14.14.1/BaseGuildTextChannel:Class#send
  await channel.send(messagePayload);
}
