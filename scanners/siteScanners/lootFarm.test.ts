import puppeteer from "puppeteer";
import { test } from "vitest";
import { getLootFarmItems } from "./lootFarm";

test("loot farm scanner", async () => {
  const results = await getLootFarmItems();
  for (const result in results) {
    if (results[result].n === "M4A4 | Temukau") {
      //console.log(results[result]);
      console.log("Results: ", result);
      console.dir(results[result], { depth: 10 });
      //console.log(results[result].u);
      //console.log(results[result].u["3"]["s"]);
      //break;
    }
  }
});
