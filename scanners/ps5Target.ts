//https://www.target.com.au/ws-api/v1/target/stock?baseProductCode=64226187&locations=5286&deliveryTypes=HD,CC DISC
//https://www.target.com.au/ws-api/v1/target/stock?baseProductCode=64226170&locations=5286&deliveryTypes=HD,CC DIGITAL
//https://www.target.com.au/p/xbox-series-s-console-digital/64445021
//https://www.target.com.au/ws-api/v1/target/delivery/estimate/64226187?postalCode=2052&storeNumber=5286&mode=CC THIS MIGHT WORK BETTER

/* fetch("https://www.target.com.au/ws-api/v1/target/stock?baseProductCode=P64445021&locations=5286&deliveryTypes=HD,ED,CC,CONSOLIDATED_STORES_SOH", {
  "referrerPolicy": "strict-origin-when-cross-origin",
  "body": null,
  "method": "GET"
}); */

import {Client, TextChannel} from "discord.js";
import axios from "axios";

class Ps5Target {
    wasFound: boolean;
    wasFound2: boolean;
    client: Client;
    channelId: string;
    roleId: string;

    constructor(client: Client, channelId: string, roleId: string) {
        this.wasFound = false;
        this.wasFound2 = false;
        this.client = client;
        this.channelId = channelId;
        this.roleId = roleId;

        this.scan = this.scan.bind(this);
    }

    scan() {
        axios.get("https://www.target.com.au/ws-api/v1/target/delivery/estimate/64226187?postalCode=2052&storeNumber=5286&mode=CC"/* "https://www.target.com.au/ws-api/v1/target/delivery/estimate/64226187?postalCode=2052&storeNumber=5286&mode=CC" */)
            .then(res => {
                if(res.data.data.deliveryModes[0].locations !== undefined) {
                    if(!this.wasFound) {
                        this.wasFound = true;
                        this.client.channels.fetch(this.channelId)
                        .then(channel => <TextChannel>channel)
                        .then(channel => {
                            if(channel) channel.send(`<@&${this.roleId}> Please know that a Disc PS5 is available at: https://www.target.com.au/p/playstation-reg-5-console/64226187`);
                        })
                        .catch(e => console.error(e))
                    }
                }
                else this.wasFound = false;

            })
            .catch(e => console.error(e));

        axios.get("https://www.target.com.au/ws-api/v1/target/delivery/estimate/64226170?postalCode=2052&storeNumber=5286&mode=CC"/* "https://www.target.com.au/ws-api/v1/target/delivery/estimate/64226187?postalCode=2052&storeNumber=5286&mode=CC" */)
            .then(res => {
                if(res.data.data.deliveryModes[0].locations !== undefined) {
                    if(!this.wasFound2) {
                        this.wasFound2 = true;
                        this.client.channels.fetch(this.channelId)
                        .then(channel => <TextChannel>channel)
                        .then(channel => {
                            if(channel) channel.send(`<@&${this.roleId}> Please know that a Digital PS5 is available at: https://www.target.com.au/p/playstation-reg-5-console-digital-edition/64226170`);
                        })
                        .catch(e => console.error(e))
                    }
                }
                else this.wasFound2 = false;

            })
            .catch(e => console.error(e));
    }   
}

export default Ps5Target;