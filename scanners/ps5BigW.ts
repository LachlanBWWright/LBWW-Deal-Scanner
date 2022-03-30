import {Client, TextChannel} from "discord.js";
import axios from "axios";

class Ps5BigW {
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
        axios.get("https://api.bigw.com.au/api/products/v0/product/124625")
            .then(res => {
                let finding: string = res.data.products[124625].listingStatus;
                if(finding !== "LISTEDINSTOREONLY") {
                    //wasFound stops messages from repeatedly being sent if the PS5 is in stock
                    if(!this.wasFound) {
                        this.wasFound = true;
                        this.client.channels.fetch(this.channelId)
                        .then(channel => <TextChannel>channel)
                        .then(channel => {
                            if(channel) channel.send(`<@&${this.roleId}> Please know that a Disc PS5 is available at: https://www.bigw.com.au/product/playstation-5-console/p/124625`);
                        })
                        .catch(console.error)
                    }
                }
                else this.wasFound = false;
            })
            .catch(err => console.log("Error: " + err));

        axios.get("https://api.bigw.com.au/api/products/v0/product/124626")
            .then(res => {
                let finding: string = res.data.products[124626].listingStatus;
                if(finding !== "LISTEDINSTOREONLY") {
                    //wasFound stops messages from repeatedly being sent if the PS5 is in stock
                    if(!this.wasFound2) {
                        this.wasFound2 = true;
                        this.client.channels.fetch(this.channelId)
                        .then(channel => <TextChannel>channel)
                        .then(channel => {
                            if(channel) channel.send(`<@&${this.roleId}> Please know that a Digital PS5 is available at: https://www.bigw.com.au/product/playstation-5-console/p/124626`);
                        })
                        .catch(console.error)
                    }
                }
                else this.wasFound2 = false;
            })
            .catch(err => console.log("Error: " + err));
    }   
}

export default Ps5BigW;