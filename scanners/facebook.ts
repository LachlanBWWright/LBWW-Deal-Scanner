import {Client, TextChannel} from "discord.js";
import { Cursor } from "mongoose";
import puppeteer, { ElementHandle } from 'puppeteer';
import FacebookQuery from "../schema/facebookQuery.js";

class Facebook {
    client: Client;
    channelId: string;
    roleId: string;
    recentlyFound: Map<string, Map<string, number>>; //Date() - Returns current time, 
    cursor: any;

    constructor(client: Client, channelId: string, roleId: string) {
        this.client = client;
        this.channelId = channelId;
        this.roleId = roleId;
        this.recentlyFound = new Map();
        this.cursor = FacebookQuery.find().cursor();

        this.scan = this.scan.bind(this);
    }

    async scan(page: puppeteer.Page) {
        try {       
            let item = await this.cursor.next();
            if(item === null) {
                this.cursor = FacebookQuery.find().cursor();
                item = await this.cursor.next();
            }
            
            await page.goto(item.name);
            await page.waitForTimeout(Math.random()*3000); //Waits before continuing. (Trying not to get IP banned)

            let res = await page.$('div.x9f619.x78zum5.xdt5ytf.x1qughib.x1rdy4ex.xz9dl7a.xsag5q8.xh8yej3.xp0eagm.x1nrcals')
            if(!res) return;
            let results = await res.$$eval('div.x1gslohp.xkh6y0r', nodes => nodes.map(n => n.textContent));
            //[ 'A$450', 'item name', 'Sydney, NSW' ] NOTE: 'A$290A$350' is how a discount will appear
            if(!results) return;
            let foundPrice = parseInt(results[0]?.split('A$')[1].replace(',', '') ?? '9999999', 10); //The text starts with A$, so [0] will be nothing, while [1] will be the price, and [2] will be the original price if discounted
            let foundName = results[1];
            
            if(foundName !== undefined && foundPrice !== undefined && foundName != item.lastItemFound) {
                this.client.channels.fetch(this.channelId)
                    .then(channel => <TextChannel>channel)
                    .then(channel => {
                        if(channel) channel.send(`<@&${this.roleId}> Please know that a ${foundName} priced at $${foundPrice} is available at ${item.name}`);
                    })
                    .catch(e => console.error(e))
                if(foundName != undefined) {
                    item.lastItemFound = foundName;
                    item.save();
                }
            }
        }
        catch(e) {
            console.error(e);
        }
    }   
}

export default Facebook;