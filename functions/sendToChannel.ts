import { ChannelType } from "discord.js";
import client from "../globals/DiscordJSClient.js";

interface sendToChannelOptions {
  files?: string[];
}

export default async function sendToChannel(
  channelId: string,
  message: string,
  { files }: sendToChannelOptions = {} //Optional param for embedding images
) {
  const channel = await client.channels.fetch(channelId);

  if (!channel) throw new Error(`Channel with id ${channelId} not found`);

  if (channel.type !== ChannelType.GuildText)
    throw new Error(`Channel with id ${channelId} is not a text channel`);

  //https://discord.js.org/docs/packages/discord.js/14.14.1/BaseGuildTextChannel:Class#send
  await channel.send({
    content: message,
    ...(files && { files: files }),
  });
}
