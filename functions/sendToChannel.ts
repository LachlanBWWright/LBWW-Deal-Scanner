import { ChannelType, ActionRowBuilder, ButtonBuilder, ButtonStyle } from "discord.js";
import client from "../globals/DiscordJSClient.js";

interface sendToChannelOptions {
  files?: string[];
  queryId?: string;
  queryType?: string;
}

export default async function sendToChannel(
  channelId: string,
  message: string,
  { files, queryId, queryType }: sendToChannelOptions = {} //Optional param for embedding images
) {
  const channel = await client.channels.fetch(channelId);

  if (!channel) throw new Error(`Channel with id ${channelId} not found`);

  if (channel.type !== ChannelType.GuildText)
    throw new Error(`Channel with id ${channelId} is not a text channel`);

  const messageOptions: any = {
    content: message,
    ...(files && { files: files }),
  };

  // Add delete button if query info is provided
  if (queryId && queryType) {
    const deleteButton = new ButtonBuilder()
      .setCustomId(`delete_query_${queryType}_${queryId}`)
      .setLabel('Delete Query')
      .setStyle(ButtonStyle.Danger);

    const row = new ActionRowBuilder<ButtonBuilder>()
      .addComponents(deleteButton);

    messageOptions.components = [row];
  }

  //https://discord.js.org/docs/packages/discord.js/14.14.1/BaseGuildTextChannel:Class#send
  await channel.send(messageOptions);
}
