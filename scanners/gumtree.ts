import {Client, TextChannel} from "discord.js";
import puppeteer, { ElementHandle } from 'puppeteer';
import GumtreeQuery from "../schema/gumtreeQuery.js";

class Gumtree {
    client: Client;
    channelId: string;
    roleId: string;
    recentlyFound: Map<string, Map<string, number>>; //Date() - Returns current time, 
    cursor: any

    constructor(client: Client, channelId: string, roleId: string) {
        this.client = client;
        this.channelId = channelId;
        this.roleId = roleId;
        this.recentlyFound = new Map();
        this.cursor = GumtreeQuery.find().cursor();

        this.scan = this.scan.bind(this);
    }

    async scan(page: puppeteer.Page) {    
        try {
            let item = await this.cursor.next()
            if(item === null) {
                this.cursor = GumtreeQuery.find().cursor();
                item = await this.cursor.next();
            }
               
            await page.goto(item.name);
            await page.waitForTimeout(Math.random()*3000); //Waits before continuing. (Trying not to get IP banned)

            let foundName: string | undefined
            let foundPrice: number | undefined
            
            let results = await page.$$('.user-ad-collection-new-design'); //#react-root > div > div.page > div > div.search-results-page__content > main > section > div
            let result: puppeteer.ElementHandle<Element> | undefined;

            for(let i = 0; i < results.length; i++) {
                let tempRes = results[i];
                //#user-ad-1296285847 > div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span.user-ad-row-new-design__flag-top
                if(!(await tempRes.$('div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span.user-ad-row-new-design__flag-top'))) {
                    result = tempRes;
                    break;
                }
            }

            if(result) {
                let resName = await result.$eval('div.user-ad-row-new-design__main-content > p.user-ad-row-new-design__title > span', res => res.textContent);
                if(resName) foundName = resName;
                
                let resPrice = await result.$eval('div.user-ad-row-new-design__right-content > div:nth-child(1) > div > span.user-ad-price-new-design__price', res => res.textContent);
                if(resPrice) foundPrice = parseFloat(resPrice.replace(/[^0-9.-]+/g,""))
                if(resPrice && resPrice.includes('Free')) foundPrice = 0
            }
            
            //Map stuff
            let foundRecently = true;
            let map = this.recentlyFound.get(item.name) ?? this.recentlyFound.set(item.name, new Map()).get(item.name); //Expands map if item does not exist in it.
            if(map !== undefined && foundName !== undefined) {
                if(map.get(foundName) === undefined) foundRecently = false; //Block notification if date was already found
                map.set(foundName, Date.now()) //MS pased since 1970, lower = older

                if(map.size > 10) { //Purges the youngest from the map
                    let toDelete = "";
                    let oldest = 0;
                    map.forEach((val, key) => {
                        if(val < oldest || oldest === 0) {
                            toDelete = key;
                            oldest = val;
                        }
                    })
                    map.delete(toDelete);
                }
            }
            
            if(foundName !== undefined && foundPrice !== undefined && foundName != item.lastItemFound && foundPrice <= item.maxPrice && !foundRecently) {
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

export default Gumtree;