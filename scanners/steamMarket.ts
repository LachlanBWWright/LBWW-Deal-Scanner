import {Client, TextChannel} from "discord.js";
import axios from "axios";
import SteamQuery from '../schema/steamQuery.js';
import { JSONArray } from "puppeteer";
import { JSONEncodable } from "@discordjs/builders";
import puppeteer, { Browser, HTTPResponse } from 'puppeteer';

//For general market queries and CS Items

class SteamMarket {
    wasFound: boolean; //TODO: Consider Removing
    client: Client;
    queryChannelId: string;
    queryRoleId: string;
    csChannelId: string;
    csRoleId: string;

    constructor(client: Client, queryChannelId: string, queryRoleId: string, csChannelId: string, csRoleId: string) {
        this.wasFound = false;
        this.client = client;
        this.queryChannelId = queryChannelId;
        this.queryRoleId = queryRoleId;
        this.csChannelId = csChannelId;
        this.csRoleId = csRoleId;

        this.scan = this.scan.bind(this);
        this.createQuery = this.createQuery.bind(this);
    }

    async scan() {
       /*  let itemsArray: JSONArray = [];
        for(let i = 0; i < 20; i++) { //Has to make multiple searches due to a size limit.
            await axios.get(`https://tradeit.gg/api/v2/inventory/data?gameId=730&offset=${i*1000}&limit=1000&sortType=(CSGO)+Best+Float&searchValue=&minPrice=0&maxPrice=100000&minFloat=0&maxFloat=1&hideTradeLock=false&fresh=true`, {headers: {'User-Agent': 'Mozilla/5.0 (platform; rv:geckoversion) Gecko/geckotrail Firefox/firefoxversion'}})
                .then(res => {
                    itemsArray = [...itemsArray, ...res.data.items];
                    if(res.data.items.length < 750) i = 20; //Breaks the loop if it's reached the end of the item list
                })
                .catch(err => console.log('TradeIt error'));
        }
        try {
            let items = <any>itemsArray;
            let cursor = TradeItItem.find().cursor();
            for(let item = await cursor.next(); item != null; item = await cursor.next()) {
                let itemWasFound = false;
                let itemCount = items.length;
                for(let i = 0; i < itemCount; i++) {
                    if(items[i].price/100.0 <= item.maxPrice && items[i].name === item.name) {
                        let bestFloat = 1;
                        if(items[i].floatValue) bestFloat = items[i].floatValue;
                        else if(items[i].floatValues) {
                            for(let x = 0; x < items[i].floatValues.length; x++) {
                                if(items[i].floatValues[x] < bestFloat) bestFloat = items[i].floatValues[x];
                            }
                        }
                        if(bestFloat >= item.minFloat && bestFloat <= item.maxFloat) {
                            if(!item.found) { //This stops repeated notification messages; the skin must not appear in a search for another message to be sent
                                item.found = true;
                                item.save(err => console.log(err));

                                this.client.channels.fetch(this.channelId)
                                .then(channel => <TextChannel>channel)
                                .then(channel => {
                                    if(channel) channel.send(`<@&${this.roleId}> Please know that a ${items[i].name} with a float of ${bestFloat} is available for $${items[i].price/100.0} USD at: https://tradeit.gg/csgo/trade`);
                                })
                                .catch(console.error)
                            }
                            itemWasFound = true;
                        }
                    }
                }
                if(!itemWasFound) {
                    item.found = false;
                    item.save();
                }
            }
        }
        catch(err) {console.log(err);} */
    }   

    async createQuery(query: string, maxPrice: number): Promise<[boolean, string]> { //TODO: Implement
        //TODO: Implement
        const browser = await puppeteer.launch({headless: true, args: ['--no-sandbox']});
        let response: [boolean, string];
        response = [false, ""];
        try {
            if(!query.includes("https://steamcommunity.com/market/search")) {
                response = [false, ""];
            }

            const page = await browser.newPage();
            page.goto(query);
            let responseUrl = "";

            //New eventlistener replacement
            await page.waitForResponse(async response => { 
                if(response.url().includes("https://steamcommunity.com/market/search/render/?query")) {
                    console.log(response.url());
                    new URL(query);
                    const steamQuery = new SteamQuery({
                        name: response.url(),
                        maxPrice: maxPrice
                    })
                    steamQuery.save(err => {console.log(err);});
                    console.log(steamQuery);
                    responseUrl = await response.url().toString();
                    return true;
                }
                return false;
            })
            response = [true, responseUrl];
        }
        catch (e) {
            console.log(e);
            response = [false, ""];
        }
        finally {
            await browser.close();
            return response;
        }
    }
}

export default SteamMarket;

