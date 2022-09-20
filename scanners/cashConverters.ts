import {Client, TextChannel} from "discord.js";
import axios from "axios";
import puppeteer, { Browser, HTTPResponse } from 'puppeteer';
import CashConvertersQuery from "../schema/cashConvertersQuery.js";

class CashConverters {
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
            let cursor = CashConvertersQuery.find().cursor();
            for(let item = await cursor.next(); item != null; item = await cursor.next()) { 
                try {
                    await page.goto(item.name);
                    await page.waitForTimeout(Math.random()*10000 + 5000); //Waits before continuing. (Trying not to get IP banned)
                    let selector = await page.waitForSelector('.product-item__title__description');
                    let cashConvertersItem = await selector?.evaluate(el => el.textContent);
                    //.evaluate(el => el.textContent);
                    if(cashConvertersItem != item.lastItemFound) {
                        this.client.channels.fetch(this.channelId)
                            .then(channel => <TextChannel>channel)
                            .then(channel => {
                                if(channel) channel.send(`<@&${this.roleId}> Please know that a ${cashConvertersItem} is available at  ${item.name}`);
                            })
                            .catch(console.error)
                        if(cashConvertersItem != undefined) {
                            item.lastItemFound = cashConvertersItem;
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

export default CashConverters;