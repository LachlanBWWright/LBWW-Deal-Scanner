import {Client, TextChannel} from "discord.js";
import puppeteer, { ElementHandle } from 'puppeteer';
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

                    let foundName: string | undefined
                    let foundPrice: number | undefined
                    
                    let result = await page.$("div[class='srp-river-results clearfix']"); //#s-item__wrapper
                    if(result) {
                        let resName = await result.$eval("span[role='heading']", res => res.textContent);
                        if(resName) foundName = resName;
                        let resPrice = await result.$eval("span[class='s-item__price']", res => res.textContent);
                        if(resPrice) foundPrice = parseFloat(resPrice.replace(/[^0-9.-]+/g,""))
                    }
                    
                    if(foundName !== undefined && foundPrice !== undefined && foundName != item.lastItemFound && foundPrice <= item.maxPrice) {
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
                catch (e) {
                    if(e instanceof Error && e.name !== 'TimeoutError') console.error(e);
                }
            }
        }
        catch(e) {
            console.error(e);
        }
        finally {
            await page.close();
            await browser.disconnect();
            await browser.close();
        }
    }   
}

export default Ebay;