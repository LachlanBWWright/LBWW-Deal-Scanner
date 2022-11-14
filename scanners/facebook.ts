import {Client, TextChannel} from "discord.js";
import puppeteer, { ElementHandle } from 'puppeteer';
import FacebookQuery from "../schema/facebookQuery.js";

class Facebook {
    client: Client;
    channelId: string;
    roleId: string;
    recentlyFound: Map<string, Map<string, number>>; //Date() - Returns current time, 

    constructor(client: Client, channelId: string, roleId: string) {
        this.client = client;
        this.channelId = channelId;
        this.roleId = roleId;
        this.recentlyFound = new Map();

        this.scan = this.scan.bind(this);
    }

    async scan() {
        const browser = await puppeteer.launch({headless: false, args: ['--no-sandbox']});
        const page = await browser.newPage();

        try {
            let cursor = FacebookQuery.find().cursor();
            for(let item = await cursor.next(); item != null; item = await cursor.next()) { 
                try {         
                    console.log('Facebook test')           
                    await page.goto(item.name);
                    await page.waitForTimeout(Math.random()*10000 + 5000); //Waits before continuing. (Trying not to get IP banned)

                    await page.$eval('div.x9f619.x78zum5.xdt5ytf.x1qughib.x1rdy4ex.xz9dl7a.xsag5q8.xh8yej3.xp0eagm.x1nrcals',
                    (res) => {
                        console.log('Test' + res);
                    })

                    console.log('Test 2')
                }
                catch (e) {
                    console.error(e);
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

export default Facebook;