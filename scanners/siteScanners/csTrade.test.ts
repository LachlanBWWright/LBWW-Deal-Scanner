import { test } from "vitest";
import { getCsTradeItems } from "./csTrade";

test("cs Trade scanner", async () => {
  const result = await getCsTradeItems();
  console.log(result[0].price);
  console.log(result[0].wear);
  console.log(result[0].market_hash_name);
});
