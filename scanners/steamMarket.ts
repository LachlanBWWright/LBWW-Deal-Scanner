import {Client, TextChannel} from "discord.js";
import axios from "axios";
import SteamQuery from '../schema/steamQuery.js';
import CsMarketItem from "../schema/csMarketItem.js";
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
    itemsFound: Map<string, number>; //<Url, TTL>
    //This is to store the names of CS items found,
    
    constructor(client: Client, queryChannelId: string, queryRoleId: string, csChannelId: string, csRoleId: string) {
        this.wasFound = false;
        this.client = client;
        this.queryChannelId = queryChannelId;
        this.queryRoleId = queryRoleId;
        this.csChannelId = csChannelId;
        this.csRoleId = csRoleId;

        this.itemsFound = new Map<string, number>();

        this.sleep = this.sleep.bind(this);
        this.scanQuery = this.scanQuery.bind(this);
        this.scanCs = this.scanCs.bind(this);
        this.createQuery = this.createQuery.bind(this);
        this.createCs = this.createCs.bind(this);
    }

    async scanQuery() {
        try {
            let cursor = SteamQuery.find().cursor()
            for(let item = await cursor.next(); item != null; item = await cursor.next()) {
                await this.sleep(15000);
                await axios.get(`${item.name}`)
                    .then(res => {
                        let price: number = res.data.results[0].sell_price;
                        for(let instance in res.data.results) {
                            let thisPrice = parseFloat(res.data.results[instance].sell_price)/100.0;
                            if(thisPrice < price) price = thisPrice;
                            let lastPrice = <number>item.lastPrice;
                            if(price < item.maxPrice && price * 1.04 < (item.lastPrice)) {
                                this.client.channels.fetch(this.queryChannelId)
                                    .then(channel => <TextChannel>channel)
                                    .then(channel => {
                                        if(channel) channel.send(`<@&${this.queryRoleId}> Please know that a ${res.data.results[instance].name} is available for $${price} USD at: ${item.displayUrl}`);
                                    })
                                    .catch(console.error)
                                break;
                            }
                        }
                        if(price != item.lastPrice) {
                            item.lastPrice = price;
                            item.save(err => console.log(err));
                        }
                }).catch(err => console.log(err));
            }
        }
        catch(e) {
            console.log(e);
        }
    }   

    async scanCs() {
        try {
            let cursor = CsMarketItem.find().cursor()
            for(let item = await cursor.next(); item != null; item = await cursor.next()) {
                await this.sleep(15000);
                await axios.get(`${item.name}`)
                    .then(async res => {
                        let i = 0;
                        for(let skin in res.data.listinginfo) {
                            // res.data.listinginfo[skin].asset.market_actions.link && https://api.csgofloat.com/?url=
                            //steam://rungame/730/76561202255233023/+csgo_econ_action_preview%20M%listingid%A%assetid%D4676111808990322346
                            //res.data.listinginfo[skin].listingid
                            //res.data.listinginfo[skin].asset.id

                            let query = "https://api.csgofloat.com/?url=".concat(res.data.listinginfo[skin].asset.market_actions[0].link)
                                .replace("%listingid%", res.data.listinginfo[skin].listingid)
                                .replace("%assetid%", res.data.listinginfo[skin].asset.id);
              
                            let price = (res.data.listinginfo[skin].converted_price_per_unit + res.data.listinginfo[skin].converted_fee_per_unit)/100.0;
                            
                            //Only calls the API if the skin isn't in the map, and the item is in the first 10
                            if(!this.itemsFound.has(query) && i < 10) await axios.get(query)
                                .then(response => {
                                    if(response.data.iteminfo.floatvalue < item.maxFloat && price <= item.maxPrice) {
                                        this.client.channels.fetch(this.csChannelId)
                                            .then(channel => <TextChannel>channel)
                                            .then(channel => {
                                                if(channel) channel.send(`<@&${this.csRoleId}> Please know that a ${response.data.iteminfo.full_item_name} with float ${response.data.iteminfo.floatvalue} is available for $${price} USD at: ${item.displayUrl}`);
                                            })
                                            .catch(console.error)
                                    }
                                }).catch(e => console.error);
                            if(i < 10) this.itemsFound.set(query, 20); //Puts the query into the map, or resets its TTL if the API was called for it
                            else if(this.itemsFound.has(query)) this.itemsFound.set(query, 20); //Resets its TTL if it's already been called once
                            i++;
                        }
                }).catch(err => console.log(err));
            }

            //Decrement the TTL in the map
            for(let [key, value] of this.itemsFound) {
                value--;
                if(value <= 0) this.itemsFound.delete(key);
            }
        }
        catch(e) {
            console.log(e);
        }
    }   

    sleep(ms: number) {
        return new Promise(
          res => setTimeout(res, ms)
        );
    }

    async createQuery(oldQuery: string, maxPrice: number): Promise<[boolean, string]> {
        let browser = await puppeteer.launch({headless: true, args: ['--no-sandbox']});
        const page = await browser.newPage();
        let response: [boolean, string];
        response = [false, ""];
        try {
            if(!oldQuery.includes("https://steamcommunity.com/market/search")) {
                response = [false, ""];
            }

            let responseUrl = "";
            let query = new URL(oldQuery);
            page.goto(query.toString());

            //New eventlistener replacement
            await page.waitForResponse(response => { 
                if(response.url().includes("https://steamcommunity.com/market/search/render/?query")) {
                    new URL(query);
                    let steamQuery = new SteamQuery({
                        name: response.url().toString().concat("&norender=1"),
                        displayUrl: oldQuery,
                        maxPrice: maxPrice
                    })
                    steamQuery.save(err => {console.log(err);});
                    console.log(steamQuery);
                    responseUrl = response.url().toString().concat("&norender=1"); //Makes it JSON instead of html
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
            console.log("If the app has just crashed, it\'s steamMarket.ts\'s fault!");
            await page.close();
            await browser.disconnect();
            await browser.close();
            return response;
        }
    }
    async createCs(oldQuery: string, maxPrice: number, maxFloat: number) { 
        //Init. Example: https://steamcommunity.com/market/listings/730/M4A1-S%20%7C%20Chantico%27s%20Fire%20%28Field-Tested%29
        //Conv. example: https://steamcommunity.com/market/listings/730/M4A1-S%20%7C%20Chantico%27s%20Fire%20%28Field-Tested%29/render/?query=&start=0&count=10&country=AU&language=english&currency=1
        try {
            if(oldQuery.includes("https://steamcommunity.com/market/listings/730/")) {
                let search = new URL(oldQuery.concat("/render/?query=&start=0&count=20&country=AU&language=english&currency=1").trim()).toString();
                let csMarketItem = new CsMarketItem({
                    name: search,
                    displayUrl: oldQuery,
                    maxPrice: maxPrice,
                    maxFloat: maxFloat
                });
                await csMarketItem.save(err => console.error);
                console.log("Search created test")
                console.log(csMarketItem);
                return search;
            }
            return "";
        }
        catch(e) {
            console.error;
            return "";
        }
    }
}

export default SteamMarket;


