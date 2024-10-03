import axios from "axios";
import globals from "../../globals/Globals.js";
import setStatus from "../../functions/setStatus.js";
import sendToChannel from "../../functions/sendToChannel.js";
import { getNotificationPrelude } from "../../functions/messagePreludes.js";
import { db, SCANNER } from "../../globals/PrismaClient.js";
import { checkIfNew } from "../../functions/handleItemUpdate.js";

export async function scanSalvos() {
  if (!globals.SALVOS || !globals.SALVOS_CHANNEL_ID || !globals.SALVOS_ROLE_ID)
    return;
  setStatus("Scanning Salvos");

  const item = await getSalvosQuery();

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
          resultsPerPage: "20",
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
    },
  );

  //Cancel if nothing found
  if (
    typeof res.data.searchResponse.results === "undefined" ||
    res.data.searchResponse.results.length <= 0
  )
    return;

  let notificationSent = false; //Don't spam notifications if multiple new items

  console.log("Salvos log");
  for (let i = 0; i < res.data.searchResponse.results.length; i++) {
    const data = res.data.searchResponse.results[i];
    const name = data.values.name.single;
    const image = data.values.image.single;
    const id = data.values.id.single;
    const slug = data.values.slug.single; // E.G. name = "Xbox One S 1TB Console" slug = "xbox-one-s-1tb-console"
    const price = data.values.price.single;
    console.log(i, name, image, price);

    if (!name || !image || !slug || !id || !price) return;

    if (price > item.maxPrice || price < item.minPrice) return;

    if (
      (await checkIfNew(id, SCANNER.SALVOS)) &&
      i <= 10 &&
      !notificationSent
    ) {
      sendToChannel(
        globals.SALVOS_CHANNEL_ID,
        `<@&${
          globals.SALVOS_ROLE_ID
        }> ${getNotificationPrelude()} a ${name} is available for $${price} at https://salvosstores.com.au/shop/p/${slug}/${getUrlCode(
          id,
        )}`,
        { files: [image] },
      );
      notificationSent = true;
    }
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

let index = 0;
async function getSalvosQuery() {
  let query = await db.salvos.findFirst({
    skip: index++,
  });
  if (query) {
    index++;
    return query;
  }
  index = 1; //Will find the first query in the line below
  return await db.salvos.findFirstOrThrow();
}
