import type { ResultAsync } from "neverthrow";
import { db } from "../globals/PrismaClient.js";
import { resultAsync } from "../functions/neverthrowUtils.js";

export type ListingSource =
  | "cashConverters"
  | "ebay"
  | "gumtree"
  | "salvos"
  | "steamMarket"
  | "csTrade"
  | "lootFarm"
  | "tradeIt";

export interface DiscoveredListing {
  readonly source: ListingSource;
  readonly externalId: string | null;
  readonly canonicalUrl: string;
  readonly title: string;
  readonly price: number | null;
  readonly shipping: number | null;
  readonly totalPrice: number | null;
  readonly currency: string | null;
  readonly imageUrl: string | null;
  readonly description: string | null;
  readonly availability: string | null;
}

export type ListingStateError =
  | { readonly type: "PersistListingFailed"; readonly message: string }
  | { readonly type: "EvaluateListingFailed"; readonly message: string };

export type RejectionReason =
  | "MissingPrice"
  | "BelowMinPrice"
  | "AboveMaxPrice"
  | "MissingRequiredPhrase"
  | "ExcludedPhrase"
  | "Unavailable"
  | "NotMatched";

export type QueryMatchResult =
  | { readonly type: "Matched" }
  | { readonly type: "Rejected"; readonly reason: RejectionReason };

export type NotificationDecision =
  | {
      readonly type: "Notify";
      readonly reason:
        | "FirstMatch"
        | "PriceDroppedIntoRange"
        | "FurtherPriceDrop";
    }
  | {
      readonly type: "DoNotNotify";
      readonly reason:
        | "Rejected"
        | "AlreadyNotified"
        | "Unavailable"
        | "MissingPrice";
    };

export interface PersistedListingObservation {
  readonly listingId: string;
  readonly observationId: string;
}

interface PersistListingInput {
  readonly listing: DiscoveredListing;
  readonly observedAt?: Date;
  readonly detailFetchedAt?: Date | null;
}

interface EvaluateListingInput {
  readonly queryId: string;
  readonly listingId: string;
  readonly source: ListingSource;
  readonly totalPrice: number | null;
  readonly match: QueryMatchResult;
  readonly evaluatedAt?: Date;
  readonly furtherPriceDropRatio?: number;
}

export function stableListingId(source: ListingSource, canonicalUrl: string) {
  return `${source}:${canonicalUrl}`;
}

export function persistListingObservation({
  listing,
  observedAt = new Date(),
  detailFetchedAt = null,
}: PersistListingInput): ResultAsync<
  PersistedListingObservation,
  ListingStateError
> {
  return resultAsync(async () => {
    const listingId = stableListingId(listing.source, listing.canonicalUrl);
    const persistedListing = await db.listing.upsert({
      where: {
        source_canonicalUrl: {
          source: listing.source,
          canonicalUrl: listing.canonicalUrl,
        },
      },
      create: {
        id: listingId,
        source: listing.source,
        externalId: listing.externalId,
        canonicalUrl: listing.canonicalUrl,
        title: listing.title,
        imageUrl: listing.imageUrl,
        description: listing.description,
        availability: listing.availability,
        lastSeenAt: observedAt,
        lastDetailAt: detailFetchedAt,
      },
      update: {
        externalId: listing.externalId,
        title: listing.title,
        imageUrl: listing.imageUrl,
        description: listing.description,
        availability: listing.availability,
        lastSeenAt: observedAt,
        ...(detailFetchedAt ? { lastDetailAt: detailFetchedAt } : {}),
        unavailableAt:
          listing.availability === "unavailable" ? observedAt : null,
      },
    });

    const observation = await db.listingObservation.create({
      data: {
        listingId: persistedListing.id,
        source: listing.source,
        observedAt,
        price: listing.price,
        shipping: listing.shipping,
        totalPrice: listing.totalPrice,
        currency: listing.currency,
        title: listing.title,
        imageUrl: listing.imageUrl,
        description: listing.description,
        availability: listing.availability,
      },
    });

    return {
      listingId: persistedListing.id,
      observationId: observation.id,
    };
  }, "Failed to persist listing observation").mapErr((error) => ({
    type: "PersistListingFailed",
    message: error.message,
  }));
}

export function evaluateListingForQuery({
  queryId,
  listingId,
  source,
  totalPrice,
  match,
  evaluatedAt = new Date(),
  furtherPriceDropRatio = 0,
}: EvaluateListingInput): ResultAsync<NotificationDecision, ListingStateError> {
  return resultAsync(async () => {
    const previous = await db.queryListingState.findUnique({
      where: {
        queryId_listingId: {
          queryId,
          listingId,
        },
      },
    });

    if (match.type === "Rejected") {
      await db.queryListingState.upsert({
        where: {
          queryId_listingId: {
            queryId,
            listingId,
          },
        },
        create: {
          queryId,
          listingId,
          source,
          status: match.reason === "Unavailable" ? "unavailable" : "rejected",
          lastEvaluatedAt: evaluatedAt,
          lastRejectedReason: match.reason,
          lowestObservedPrice: totalPrice,
        },
        update: {
          status: match.reason === "Unavailable" ? "unavailable" : "rejected",
          lastEvaluatedAt: evaluatedAt,
          lastRejectedReason: match.reason,
          lowestObservedPrice: lowestPrice(previous?.lowestObservedPrice, totalPrice),
        },
      });
      return {
        type: "DoNotNotify",
        reason: match.reason === "Unavailable" ? "Unavailable" : "Rejected",
      } satisfies NotificationDecision;
    }

    if (totalPrice === null) {
      return {
        type: "DoNotNotify",
        reason: "MissingPrice",
      } satisfies NotificationDecision;
    }

    const decision = getMatchedDecision(
      previous?.status ?? null,
      previous?.lastNotifiedTotalPrice ?? null,
      totalPrice,
      furtherPriceDropRatio,
    );

    await db.queryListingState.upsert({
      where: {
        queryId_listingId: {
          queryId,
          listingId,
        },
      },
      create: {
        queryId,
        listingId,
        source,
        status: decision.type === "Notify" ? "notified" : "matched",
        lastEvaluatedAt: evaluatedAt,
        firstMatchedAt: evaluatedAt,
        lastMatchedAt: evaluatedAt,
        lastNotifiedAt: decision.type === "Notify" ? evaluatedAt : null,
        lastNotifiedTotalPrice:
          decision.type === "Notify" ? totalPrice : null,
        lowestObservedPrice: totalPrice,
      },
      update: {
        status: decision.type === "Notify" ? "notified" : "matched",
        lastEvaluatedAt: evaluatedAt,
        firstMatchedAt: previous?.firstMatchedAt ?? evaluatedAt,
        lastMatchedAt: evaluatedAt,
        lastRejectedReason: null,
        lastNotifiedAt:
          decision.type === "Notify"
            ? evaluatedAt
            : previous?.lastNotifiedAt,
        lastNotifiedTotalPrice:
          decision.type === "Notify"
            ? totalPrice
            : previous?.lastNotifiedTotalPrice,
        lowestObservedPrice: lowestPrice(previous?.lowestObservedPrice, totalPrice),
      },
    });

    return decision;
  }, "Failed to evaluate listing for query")
    .mapErr((error) => ({
      type: "EvaluateListingFailed",
      message: error.message,
    }));
}

function getMatchedDecision(
  previousStatus: string | null,
  lastNotifiedTotalPrice: number | null,
  totalPrice: number,
  furtherPriceDropRatio: number,
): NotificationDecision {
  if (!previousStatus || previousStatus === "unseen") {
    return { type: "Notify", reason: "FirstMatch" };
  }

  if (previousStatus === "rejected") {
    return { type: "Notify", reason: "PriceDroppedIntoRange" };
  }

  if (
    lastNotifiedTotalPrice !== null &&
    furtherPriceDropRatio > 0 &&
    totalPrice < lastNotifiedTotalPrice * (1 - furtherPriceDropRatio)
  ) {
    return { type: "Notify", reason: "FurtherPriceDrop" };
  }

  return { type: "DoNotNotify", reason: "AlreadyNotified" };
}

function lowestPrice(
  previousPrice: number | null | undefined,
  currentPrice: number | null,
): number | null {
  if (currentPrice === null) return previousPrice ?? null;
  if (previousPrice === null || previousPrice === undefined) return currentPrice;
  return Math.min(previousPrice, currentPrice);
}

export function matchPriceRange(
  totalPrice: number | null,
  minPrice: number | null,
  maxPrice: number | null,
): QueryMatchResult {
  if (totalPrice === null) return { type: "Rejected", reason: "MissingPrice" };
  if (minPrice !== null && totalPrice < minPrice) {
    return { type: "Rejected", reason: "BelowMinPrice" };
  }
  if (maxPrice !== null && totalPrice > maxPrice) {
    return { type: "Rejected", reason: "AboveMaxPrice" };
  }
  return { type: "Matched" };
}

export function parsePhraseList(input: string): string[] {
  return input
    .split(",")
    .map((phrase) => phrase.trim().toLocaleLowerCase())
    .filter((phrase) => phrase.length > 0);
}

export function matchPhrases(
  searchableText: string,
  requiredPhrases: string,
  excludePhrases: string,
): QueryMatchResult {
  const normalizedText = searchableText.toLocaleLowerCase();
  const required = parsePhraseList(requiredPhrases);
  const excluded = parsePhraseList(excludePhrases);

  if (required.some((phrase) => !normalizedText.includes(phrase))) {
    return { type: "Rejected", reason: "MissingRequiredPhrase" };
  }

  if (excluded.some((phrase) => normalizedText.includes(phrase))) {
    return { type: "Rejected", reason: "ExcludedPhrase" };
  }

  return { type: "Matched" };
}

export function combineMatchResults(
  first: QueryMatchResult,
  second: QueryMatchResult,
): QueryMatchResult {
  if (first.type === "Rejected") return first;
  if (second.type === "Rejected") return second;
  return { type: "Matched" };
}
