import { db } from "../globals/PrismaClient";
import { SCANNER } from "@prisma/client";

//Manages the time to live
export class TTLManager<T> {
  ttlMap = new Map<T, number>();
  maxTTL;
  newSize;
  retainSize;
  scannerType: SCANNER;

  constructor({
    maxTTL,
    newSize,
    retainSize,
    scannerType,
  }: {
    maxTTL?: number;
    newSize?: number;
    retainSize?: number;
    scannerType: SCANNER;
  }) {
    this.maxTTL = maxTTL ?? 20;
    this.newSize = newSize ?? 10;
    this.retainSize = retainSize ?? 20;
    this.scannerType = scannerType;
  }

  addItems() {
    db.ttlItem.create({
      data: {
        itemId: "2",
        ttl: this.maxTTL,
        scanner: this.scannerType,
      },
    });
  }
}
