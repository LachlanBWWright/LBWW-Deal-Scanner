import { db, SCANNER } from "../globals/PrismaClient.js";

/* 
Checks an item. If it doesn't, add it to the database, and return true.
If it already exists, extend the TTL
*/
export async function checkIfNew(itemId: string, scanner: SCANNER) {
  const item = await db.ttlItem.findUnique({ where: { itemId, scanner } });
  console.log("CHECKIFNEW");
  console.log(item);
  if (item) {
    db.ttlItem.update({
      where: { itemId, scanner },
      data: {
        lastUpdated: new Date(),
      },
    });
    return false;
  }

  await db.ttlItem.create({
    data: { itemId, scanner, lastUpdated: new Date() },
  });
  return true;
}
