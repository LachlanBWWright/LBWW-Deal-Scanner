import { describe, expect, it } from "vitest";
import { db } from "../globals/PrismaClient.js";
import {
  combineMatchResults,
  evaluateListingForQuery,
  matchPhrases,
  matchPriceRange,
  persistListingObservation,
  parsePhraseList,
  type DiscoveredListing,
} from "./listingState.js";

describe("listingState matching helpers", () => {
  it("parses comma-separated phrases deterministically", () => {
    expect(parsePhraseList(" xbox, Series X ,, Controller ")).toEqual([
      "xbox",
      "series x",
      "controller",
    ]);
  });

  it("matches required phrases across title and description text", () => {
    expect(
      matchPhrases(
        "Xbox console\nIncludes an elite controller",
        "xbox, elite controller",
        "",
      ),
    ).toEqual({ type: "Matched" });
  });

  it("rejects missing required phrases and excluded phrases", () => {
    expect(matchPhrases("Xbox console", "playstation", "")).toEqual({
      type: "Rejected",
      reason: "MissingRequiredPhrase",
    });
    expect(matchPhrases("Xbox console damaged", "xbox", "damaged")).toEqual({
      type: "Rejected",
      reason: "ExcludedPhrase",
    });
  });

  it("models price filters without blocking observation persistence", () => {
    expect(matchPriceRange(120, null, 100)).toEqual({
      type: "Rejected",
      reason: "AboveMaxPrice",
    });
    expect(matchPriceRange(80, 100, null)).toEqual({
      type: "Rejected",
      reason: "BelowMinPrice",
    });
    expect(matchPriceRange(100, 80, 120)).toEqual({ type: "Matched" });
  });

  it("combines phrase and price match results with first rejection winning", () => {
    expect(
      combineMatchResults(
        { type: "Rejected", reason: "ExcludedPhrase" },
        { type: "Rejected", reason: "AboveMaxPrice" },
      ),
    ).toEqual({ type: "Rejected", reason: "ExcludedPhrase" });
    expect(
      combineMatchResults({ type: "Matched" }, { type: "Matched" }),
    ).toEqual({ type: "Matched" });
  });

  it("persists rejected observations and notifies when price drops into range once", async () => {
    const query = await db.query.create({ data: {} });
    const canonicalUrl = `https://example.test/listing-${Date.now()}`;
    const firstListing = makeListing(canonicalUrl, 150);
    const secondListing = makeListing(canonicalUrl, 90);

    const firstPersisted = await persistListingObservation({
      listing: firstListing,
    });
    expect(firstPersisted.isOk()).toBe(true);
    if (firstPersisted.isErr()) return;

    const firstDecision = await evaluateListingForQuery({
      queryId: query.id,
      listingId: firstPersisted.value.listingId,
      source: "ebay",
      totalPrice: firstListing.totalPrice,
      match: matchPriceRange(firstListing.totalPrice, null, 100),
    });
    expect(firstDecision.isOk()).toBe(true);
    if (firstDecision.isErr()) return;
    expect(firstDecision.value).toEqual({
      type: "DoNotNotify",
      reason: "Rejected",
    });

    const secondPersisted = await persistListingObservation({
      listing: secondListing,
    });
    expect(secondPersisted.isOk()).toBe(true);
    if (secondPersisted.isErr()) return;

    const secondDecision = await evaluateListingForQuery({
      queryId: query.id,
      listingId: secondPersisted.value.listingId,
      source: "ebay",
      totalPrice: secondListing.totalPrice,
      match: matchPriceRange(secondListing.totalPrice, null, 100),
    });
    expect(secondDecision.isOk()).toBe(true);
    if (secondDecision.isErr()) return;
    expect(secondDecision.value).toEqual({
      type: "Notify",
      reason: "PriceDroppedIntoRange",
    });

    const thirdDecision = await evaluateListingForQuery({
      queryId: query.id,
      listingId: secondPersisted.value.listingId,
      source: "ebay",
      totalPrice: secondListing.totalPrice,
      match: matchPriceRange(secondListing.totalPrice, null, 100),
    });
    expect(thirdDecision.isOk()).toBe(true);
    if (thirdDecision.isErr()) return;
    expect(thirdDecision.value).toEqual({
      type: "DoNotNotify",
      reason: "AlreadyNotified",
    });

    const observations = await db.listingObservation.findMany({
      where: { listingId: secondPersisted.value.listingId },
    });
    expect(observations).toHaveLength(2);

    await db.query.delete({ where: { id: query.id } });
    await db.listing.delete({
      where: {
        source_canonicalUrl: {
          source: "ebay",
          canonicalUrl,
        },
      },
    });
  });

  it("notifies on first match and further configured price drops", async () => {
    const query = await db.query.create({ data: {} });
    const canonicalUrl = `https://example.test/further-drop-${Date.now()}`;
    const firstListing = makeListing(canonicalUrl, 100);
    const secondListing = makeListing(canonicalUrl, 94);

    const firstPersisted = await persistListingObservation({
      listing: firstListing,
    });
    expect(firstPersisted.isOk()).toBe(true);
    if (firstPersisted.isErr()) return;

    const firstDecision = await evaluateListingForQuery({
      queryId: query.id,
      listingId: firstPersisted.value.listingId,
      source: "ebay",
      totalPrice: firstListing.totalPrice,
      match: { type: "Matched" },
      furtherPriceDropRatio: 0.05,
    });
    expect(firstDecision.isOk()).toBe(true);
    if (firstDecision.isErr()) return;
    expect(firstDecision.value).toEqual({
      type: "Notify",
      reason: "FirstMatch",
    });

    const secondPersisted = await persistListingObservation({
      listing: secondListing,
    });
    expect(secondPersisted.isOk()).toBe(true);
    if (secondPersisted.isErr()) return;

    const secondDecision = await evaluateListingForQuery({
      queryId: query.id,
      listingId: secondPersisted.value.listingId,
      source: "ebay",
      totalPrice: secondListing.totalPrice,
      match: { type: "Matched" },
      furtherPriceDropRatio: 0.05,
    });
    expect(secondDecision.isOk()).toBe(true);
    if (secondDecision.isErr()) return;
    expect(secondDecision.value).toEqual({
      type: "Notify",
      reason: "FurtherPriceDrop",
    });

    await db.query.delete({ where: { id: query.id } });
    await db.listing.delete({
      where: {
        source_canonicalUrl: {
          source: "ebay",
          canonicalUrl,
        },
      },
    });
  });
});

function makeListing(canonicalUrl: string, totalPrice: number): DiscoveredListing {
  return {
    source: "ebay",
    externalId: null,
    canonicalUrl,
    title: "Price drop listing",
    price: totalPrice,
    shipping: null,
    totalPrice,
    currency: "AUD",
    imageUrl: null,
    description: null,
    availability: "available",
  };
}
