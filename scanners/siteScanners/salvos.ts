import puppeteer from "puppeteer";
import axios from "axios";
import SalvosQuery from "../../schema/salvosQuery.js";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";

let cursor = SalvosQuery.find().cursor();

export async function scanSalvos() {
  if (!globals.SALVOS || !globals.SALVOS_CHANNEL_ID || !globals.SALVOS_ROLE_ID)
    return;
  setStatus("Scanning Salvos");

  if (cursor === null) cursor = SalvosQuery.find().cursor(); //Bug fix for cursor being null
  let item = await cursor.next();
  if (item === null) {
    cursor = SalvosQuery.find().cursor();
    item = await cursor.next();
  }

  let res = await axios.post(
    "https://jsonapi-au-valkyrie.sajari.com/sajari.api.pipeline.v1.Query/Search",
    {
      metadata: {
        project: ["1638755075284158580"],
        collection: ["test5"],
      },
      request: {
        pipeline: {
          name: "app",
        },
        tracking: {},
        values: {
          q: item.name,
          resultsPerPage: "1",
          filter: `(price >= ${item.minPrice ?? "0"} AND price <= ${
            item.maxPrice ?? "99999"
          })`,
          sort: "created", //created (New first ) -created (Old first)
        },
      },
    },
    {
      headers: {
        Origin: "https://www.salvosstores.com.au",
      },
    }
  );

  //Cancel if nothing found
  if (
    typeof res.data.searchResponse.results === "undefined" ||
    res.data.searchResponse.results.length <= 0
  )
    return;
  const data = res.data.searchResponse.results[0];
  const name = data.values.name.single;
  const image = data.values.image.single;
  const id = data.values.id.single;
  const slug = data.values.slug.single; // E.G. name = "Xbox One S 1TB Console" slug = "xbox-one-s-1tb-console"
  const price = data.values.price.single;

  if (!name || !image || !slug || !id || !price) return;

  if (name !== item.lastItemFound) {
    sendToChannel(
      globals.SALVOS_CHANNEL_ID,
      `<@&${
        globals.SALVOS_ROLE_ID
      }> Please know that a ${name} is available for $${price} at https://salvosstores.com.au/shop/p/${slug}/${getUrlCode(
        id
      )}`,
      { files: [image] }
    );
    item.lastItemFound = name;
    item.save();
  }
}

function getUrlCode(e: string) {
  const t = "Product";
  if (!e) return "";
  const n = D(e);
  const r = /(\w+):([\da-f-]+)/.exec(n);
  if (!r) return null;
  if (t && t !== r[1]) throw new Error("Schema is not correct");
  return r[2];
}

function A(e: string) {
  return Buffer.from(e, "base64").toString("utf8");
}
function D(e: string) {
  return A(M(e));
}
function M(e: string) {
  return m(e.replace(/[-_]/g, (e) => ("-" == e ? "+" : "/")));
}
function m(e: string) {
  return e.replace(/[^A-Za-z0-9\+\/]/g, "");
}
