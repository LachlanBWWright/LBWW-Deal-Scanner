import {Client, TextChannel} from "discord.js";
import puppeteer, { Puppeteer } from 'puppeteer';
import SalvosQuery from "../schema/salvosQuery.js";

class Salvos {
    client: Client;
    channelId: string;
    roleId: string;
    cursor: any;

    constructor(client: Client, channelId: string, roleId: string) {
        this.client = client;
        this.channelId = channelId;
        this.roleId = roleId;
        this.cursor = SalvosQuery.find().cursor();
        this.scan = this.scan.bind(this);
    }

    async scan(page: puppeteer.Page) {
        try {
            let item = await this.cursor.next()
            if(item === null) {
                this.cursor = SalvosQuery.find().cursor();
                item = await this.cursor.next();
            }
             
            await page.goto(item.name);
            await page.waitForTimeout(Math.random()*3000); //Waits before continuing. (Trying not to get IP banned)
            let selector = await page.waitForSelector('.line-clamp-3');
            let salvosItem = await selector?.evaluate(el => el.textContent);
            if(salvosItem != item.lastItemFound) {
                this.client.channels.fetch(this.channelId)
                    .then(channel => <TextChannel>channel)
                    .then(channel => {
                        if(channel) channel.send(`<@&${this.roleId}> Please know that a ${salvosItem} is available at  ${item.name}`);
                    })
                    .catch(e => console.error(e))
                if(salvosItem != undefined) {
                    item.lastItemFound = salvosItem;
                    item.save();
                }
            } 
        }
        catch(e: any) {
            if(e.constructor.name !== 'TimeoutError') console.error(e);
        }
    }   
}

export default Salvos;