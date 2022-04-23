import puppeteer from 'puppeteer';
import {PendingXHR} from 'pending-xhr-puppeteer'

class CsDeals {
    browser: Promise<puppeteer.Browser>;

    constructor() {
        this.browser = puppeteer.launch({
            headless: true
        });
    }

    async scan() {
        try {
            const page = await (await this.browser).newPage();
            await page.goto("https://cs.deals/trade-skins");
            await page.reload(); //This is necessary to get the network activity

            page.on('response', async res => {
                if(res.url().endsWith("botsinventory?appid=0")) {
                    let items = await res.json()
                    let csgoItemCount = items.response.items[730].length;
                    items = items.response.items[730];
                    let count = 0;
                    console.time("Test")
                    for(let i = 0; i < csgoItemCount; i++) count++
                    console.timeEnd("Test");
                    console.log("Count " + count);

                }
            })
        }
        catch(error) {
            console.log(error);
        }
    }
}

export default CsDeals;


/* 

JSON Example

{
  a: '68',
  c: 'Souvenir AUG | Radiation Hazard (Factory New)',
  e: 4,
  g: '3023755629',
  h: '188530139',
  i: 5.82,                                                                                  The actual price
  k: '5e98d9',
  q: 'OTg4ODg=',
  x: 5.43,                                                                                  Some sort of 'base price', lower than actual
  f: '22590511304',
  t: 8499961,
  f1: [
    [
      'Fnatic (Gold) | Cologne 2016',
      'cologne2016/fntc_gold.e622cb6a1885e75f2a1d068efdabba19a6e87d5a.png',
      '0.00000000000000000',
      '0'
    ],
    [
      'ESL (Gold) | Cologne 2016',
      'cologne2016/esl_gold.f2b63efe44b0a777411448aaf9604014ee414e02.png',
      '1.00000000000000000',
      '2'
    ]
  ],
  b1: '5621234145811519616',
  o: 'Rifle',
  h1: 'MTY4MzE4NQ==',
  c1: 'FN',
  d1: '0.05834920331836',                                                                Float Value
  e1: 160
}

*/