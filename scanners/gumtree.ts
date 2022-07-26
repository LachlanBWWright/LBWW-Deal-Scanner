import {Client, TextChannel} from "discord.js";
import puppeteer, { ElementHandle } from 'puppeteer';
import GumtreeQuery from "../schema/gumtreeQuery.js";

class Gumtree {
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
            let cursor = GumtreeQuery.find().cursor();
            for(let item = await cursor.next(); item != null; item = await cursor.next()) { 
                try {
                    await page.goto(item.name);
                    await page.waitForTimeout(Math.random()*10000 + 5000); //Waits before continuing. (Trying not to get IP banned)

                    let foundName: string | undefined
                    let foundPrice: number | undefined

                    let result = await page.$('#user-ad-collection-new-design');
                    if(result) {
                        //#user-ad-1298634939 > div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span
                        let resName = await result.$eval('div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span', res => res.textContent);
                        if(resName) foundName = resName;
                        //#user-ad-1298634939 > div.user-ad-row-new-design__right-content > div:nth-child(1) > div > span.user-ad-price-new-design__price
                        let resPrice = await result.$eval('div.user-ad-row-new-design__right-content > div:nth-child(1) > div > span.user-ad-price-new-design__price', res => res.textContent);
                        if(resPrice) foundPrice = parseFloat(resPrice.replace(/[^0-9.-]+/g,""))
                        console.log('foundName: ' + foundName + 'foundPrice ' + foundPrice)
                    }
                    
                    if(foundName !== undefined && foundPrice !== undefined && foundName != item.lastItemFound && foundPrice <= item.maxPrice) {
                        this.client.channels.fetch(this.channelId)
                            .then(channel => <TextChannel>channel)
                            .then(channel => {
                                if(channel) channel.send(`<@&${this.roleId}> Please know that a ${foundName} priced at $${foundPrice} is available at ${item.name}`);
                            })
                            .catch(console.error)
                        if(foundName != undefined) {
                            item.lastItemFound = foundName;
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

export default Gumtree;