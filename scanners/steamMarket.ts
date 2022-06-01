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
                await this.sleep(4000);
                console.log(item);
                await axios.get(`${item.name}`)
                    .then(res => {
                        let price: number = res.data.results[0].sell_price;
                        for(let instance in res.data.results) {
                            let thisPrice = parseFloat(res.data.results[instance].sell_price)/100.0;
                            if(thisPrice < price) price = thisPrice;
                            let lastPrice = <number>item.lastPrice;
                            if(price < item.maxPrice && price < (lastPrice * 0.95)) {
                                this.client.channels.fetch(this.queryChannelId)
                                    .then(channel => <TextChannel>channel)
                                    .then(channel => {
                                        if(channel) channel.send(`<@&${this.queryRoleId}> Please know that a ${res.data.results[instance].name} is available for $${price} USD at: ${item.name}`);
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
                    console.log(response.url());
                    new URL(query);
                    let steamQuery = new SteamQuery({
                        name: response.url().toString().concat("&norender=1"),
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
    async createCs(oldQuery: string, maxPrice: number): Promise<[boolean, string]> { 
        console.log("TEST");
        return[true, "TEST"];
    }
}

export default SteamMarket;

function sleep(arg0: number) {
    throw new Error("Function not implemented.");
}

