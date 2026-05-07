import {
  ChannelType,
  ActionRowBuilder,
  ButtonBuilder,
  ButtonStyle,
  MessagePayload,
  type MessagePayloadOption,
} from "discord.js";
import type { AppNotification, DealNotification } from "../../deals/types.js";
import type { NotificationProvider } from "../../notifications/types.js";
import client from "../../globals/DiscordJSClient.js";
import {
  actionRegistry,
  generateRandomKey,
} from "../../commandManager/actionRegistry.js";
import { db } from "../../globals/PrismaClient.js";
import { fromThrowableAsync } from "../../functions/neverthrowUtils.js";
import globals from "../../globals/Globals.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";

export class DiscordNotificationProvider implements NotificationProvider {
  name = "discord";

  isEnabled(): boolean {
    return !!(
      globals.DISCORD_TOKEN &&
      globals.BOT_CLIENT_ID &&
      globals.DISCORD_GUILD_ID
    );
  }

  async send(notification: AppNotification): Promise<void> {
    if (notification.kind === "error") {
      await this.sendError(notification.source, notification.message, notification.stack);
    } else {
      await this.sendDeal(notification);
    }
  }

  private async sendError(source: string, message: string, stack?: string): Promise<void> {
    if (!globals.ERROR_CHANNEL_ID) return;
    await this.sendToDiscordChannel(
      globals.ERROR_CHANNEL_ID,
      `An error has occurred in ${source}:\n\n${message}\n\n${stack ?? ""}`,
    );
  }

  private async sendDeal(notification: DealNotification): Promise<void> {
    const channelId = this.getChannelId(notification);
    const roleId = this.getRoleId(notification);
    if (!channelId) return;

    const prelude = getNotificationPrelude();
    const roleMention = roleId ? `<@&${roleId}> ` : "";
    const message = `${roleMention}${prelude} ${notification.title}`;

    const queryId = notification.query?.id;
    const queryType = notification.query?.type;

    // Check dmOnly before sending to channel
    if (queryId && queryType) {
      await this.sendDMsToSubscribedUsers(queryId, queryType, message, notification.imageUrl ? [notification.imageUrl] : undefined);

      const query = await getQueryByTypeAndId(queryType, queryId);
      if (query?.query?.dmOnly) return;
    }

    await this.sendToDiscordChannel(channelId, message, {
      files: notification.imageUrl ? [notification.imageUrl] : undefined,
      queryId,
      queryType,
    });
  }

  private getChannelId(notification: DealNotification): string | null {
    switch (notification.source) {
      case "cashConverters":
        return globals.CASH_CONVERTERS_CHANNEL_ID;
      case "ebay":
        return globals.EBAY_CHANNEL_ID;
      case "gumtree":
        return globals.GUMTREE_CHANNEL_ID;
      case "salvos":
        return globals.SALVOS_CHANNEL_ID;
      case "steamMarket":
        return globals.STEAM_QUERY_CHANNEL_ID;
      case "csTrade":
      case "lootFarm":
      case "tradeIt":
        return globals.CS_CHANNEL_ID;
      default:
        return null;
    }
  }

  private getRoleId(notification: DealNotification): string | null {
    switch (notification.source) {
      case "cashConverters":
        return globals.CASH_CONVERTERS_ROLE_ID;
      case "ebay":
        return globals.EBAY_ROLE_ID;
      case "gumtree":
        return globals.GUMTREE_ROLE_ID;
      case "salvos":
        return globals.SALVOS_ROLE_ID;
      case "steamMarket":
        return globals.STEAM_QUERY_ROLE_ID;
      case "csTrade":
      case "lootFarm":
      case "tradeIt":
        return globals.CS_ROLE_ID;
      default:
        return null;
    }
  }

  private async sendToDiscordChannel(
    channelId: string,
    message: string,
    options: { files?: string[]; queryId?: string; queryType?: string } = {},
  ): Promise<void> {
    const { files, queryId, queryType } = options;
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
      ...(files && { files }),
    };

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

      await actionRegistry.cleanupExpired(10 * 60 * 1000);
    }

    const messagePayload = new MessagePayload(channel, messageOptions);
    await channel.send(messagePayload);
  }

  private async sendDMsToSubscribedUsers(
    queryId: string,
    queryType: string,
    message: string,
    files?: string[],
  ): Promise<void> {
    const query = await getQueryByTypeAndId(queryType, queryId);
    if (!query?.queryId) return;

    const userQueries = await db.userQuery.findMany({
      where: { queryId: query.queryId },
    });

    for (const userQuery of userQueries) {
      const dmResult = await fromThrowableAsync(async () => {
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
          files,
          components: [row],
        });
      }, `Failed to send DM to user ${userQuery.userId}`);

      if (dmResult.isErr()) {
        console.error(dmResult.error.message);
      }
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
