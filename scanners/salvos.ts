import {Client, TextChannel} from "discord.js";
import puppeteer from 'puppeteer';
import SalvosQuery from "../schema/salvosQuery.js";

class Salvos {
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
            let cursor = SalvosQuery.find().cursor();
            for(let item = await cursor.next(); item != null; item = await cursor.next()) { 
                try {
                    console.log("SALVOS")
                    await page.goto(item.name);
                    await page.waitForTimeout(Math.random()*10000 + 5000); //Waits before continuing. (Trying not to get IP banned)
                    let selector = await page.waitForSelector('.line-clamp-3');
                    let salvosItem = await selector?.evaluate(el => el.textContent);
                    if(salvosItem != item.lastItemFound) {
                        this.client.channels.fetch(this.channelId)
                            .then(channel => <TextChannel>channel)
                            .then(channel => {
                                if(channel) channel.send(`<@&${this.roleId}> Please know that a ${salvosItem} is available at  ${item.name}`);
                            })
                            .catch(console.error)
                        if(salvosItem != undefined) {
                            item.lastItemFound = salvosItem;
                            item.save();
                        }
                    }
                }
                catch (e) {
                    console.error;
                }
            }
        }
        catch(e) {
            console.error;
        }
        finally {
            await page.close();
            await browser.disconnect();
            await browser.close();
        }
    }   
}

export default Salvos;