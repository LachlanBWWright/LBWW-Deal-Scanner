import { ChannelType } from "discord.js";
import client from "../globals/DiscordJSClient.js";

export default async function sendToChannel(
  channelId: string,
  message: string
) {
  const channel = await client.channels.fetch(channelId);

  if (!channel) throw new Error(`Channel with id ${channelId} not found`);

  if (channel.type !== ChannelType.GuildText)
    throw new Error(`Channel with id ${channelId} is not a text channel`);

  await channel.send(message);
}
