import { Client, TextChannel } from "discord.js";
import axios from "axios";

class XboxBigW {
  wasFound: boolean;
  client: Client;
  channelId: string;
  roleId: string;

  constructor(client: Client, channelId: string, roleId: string) {
    this.wasFound = false;
    this.client = client;
    this.channelId = channelId;
    this.roleId = roleId;

    this.scan = this.scan.bind(this);
  }

  scan() {
    axios
      .get("https://api.bigw.com.au/api/products/v0/product/124385")
      .then((res) => {
        let finding: string = res.data.products[124385].listingStatus;
        if (finding !== "LISTEDINSTOREONLY") {
          //wasFound stops messages from repeatedly being sent if the PS5 is in stock
          if (!this.wasFound) {
            this.wasFound = true;
            this.client.channels
              .fetch(this.channelId)
              .then((channel) => <TextChannel>channel)
              .then((channel) => {
                if (channel)
                  channel.send(
                    `<@&${this.roleId}> Please know that an XBox Series X is available at: https://www.bigw.com.au/product/xbox-series-x-1tb-console/p/124385`,
                  );
              })
              .catch((e) => console.error(e));
          }
        } else this.wasFound = false;
      })
      .catch((e) => console.error(e));
  }
}

export default XboxBigW;
