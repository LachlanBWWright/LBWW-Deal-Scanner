import {Client, TextChannel} from "discord.js";
import puppeteer from 'puppeteer';
import EbayQuery from "../schema/ebayQuery.js";

class Ebay {
    client: Client;
    channelId: string;
    roleId: string;

    constructor(client: Client, channelId: string, roleId: string) {
        this.client = client;
        this.channelId = channelId;
        this.roleId = roleId;

        this.scan = this.scan.bind(this);
    }

    async scan() {
        const browser = await puppeteer.launch({headless: true, args: ['--no-sandbox']});
        const page = await browser.newPage();
        try {
            let cursor = EbayQuery.find().cursor();
            for(let item = await cursor.next(); item != null; item = await cursor.next()) { 
                try {
                    await page.goto(item.name);
                    await page.waitForTimeout(Math.random()*10000 + 5000); //Waits before continuing. (Trying not to get IP banned)
                    let selector = await page.waitForSelector('#srp-river-results > ul > li:nth-child(3) > div > div.s-item__info.clearfix > a > h3');
                    let ebayItem = await selector?.evaluate(el => el.textContent);
                    //.evaluate(el => el.textContent);
                    if(ebayItem != item.lastItemFound) {
                        this.client.channels.fetch(this.channelId)
                            .then(channel => <TextChannel>channel)
                            .then(channel => {
                                if(channel) channel.send(`<@&${this.roleId}> Please know that a ${ebayItem} is available at  ${item.name}`);
                            })
                            .catch(console.error)
                        if(ebayItem != undefined) {
                            item.lastItemFound = ebayItem;
                            item.save();
                        }
                    }
                }
                catch (e) {
                    console.log(e);
                }
            }
        }
        catch(e) {
            console.log(e);
        }
        finally {
            await page.close();
            await browser.disconnect();
            await browser.close();
        }
    }   
}

export default Ebay;