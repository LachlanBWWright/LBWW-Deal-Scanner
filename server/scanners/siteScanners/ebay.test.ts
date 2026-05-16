import { describe, expect, it } from "vitest";
import {
  type EbayCardElement,
  ebaySelectors,
  parseEbayItem,
  readRawEbayItem,
} from "./ebay.js";

function createFakeEbayCard(
  values: ReadonlyMap<string, { readonly textContent?: string; readonly src?: string }>,
): EbayCardElement {
  return {
    querySelector: (selectors: string) => values.get(selectors) ?? null,
  };
}

describe("eBay selector configuration", () => {
  it("uses resilient class selectors for current eBay result cards", () => {
    expect(ebaySelectors.resultsContainer).toBe("div.srp-river");
    expect(ebaySelectors.emptyResults).toBe(".srp-save-null-search__heading");
    expect(ebaySelectors.resultCard).toBe(
      "div.su-card-container.su-card-container--horizontal, li.s-item",
    );
    expect(ebaySelectors.title).toBe('div[role="heading"], .s-item__title');
    expect(ebaySelectors.price).toBe("span.s-card__price, .s-item__price");
    expect(ebaySelectors.image).toBe("img");
  });
});

describe("parseEbayItem", () => {
  it("reads current eBay card selectors from a card-like element", () => {
    const card = createFakeEbayCard(
      new Map([
        [ebaySelectors.title, { textContent: " New listing Borderlands 4 " }],
        [ebaySelectors.price, { textContent: " AU $69.95 " }],
        [ebaySelectors.image, { src: "https://i.ebayimg.com/images/g/example/s-l500.webp" }],
      ]),
    );

    expect(readRawEbayItem(card)).toEqual({
      title: "New listing Borderlands 4",
      priceText: "AU $69.95",
      imageUrl: "https://i.ebayimg.com/images/g/example/s-l500.webp",
    });
  });

  it("normalizes current eBay card text into scanner values", () => {
    const result = parseEbayItem({
      title: "New listing Borderlands 4 - PlayStation 5",
      priceText: "AU $69.95",
      imageUrl: "https://i.ebayimg.com/images/g/example/s-l500.webp",
    });

    expect(result).toEqual({
      foundName: "Borderlands 4 - PlayStation 5",
      foundPrice: 69.95,
      foundImage: "https://i.ebayimg.com/images/g/example/s-l500.webp",
    });
  });

  it("keeps older s-item fallback markup values usable", () => {
    const result = parseEbayItem({
      title: "Borderlands 4 Deluxe Edition",
      priceText: "$59.00",
      imageUrl: "https://i.ebayimg.com/images/g/fallback/s-l500.webp",
    });

    expect(result).toEqual({
      foundName: "Borderlands 4 Deluxe Edition",
      foundPrice: 91.45,
      foundImage: "https://i.ebayimg.com/images/g/fallback/s-l500.webp",
    });
  });

  it("returns empty values when required fields are missing", () => {
    expect(
      parseEbayItem({
        title: "Borderlands 4",
        priceText: null,
        imageUrl: "https://i.ebayimg.com/images/g/example/s-l500.webp",
      }),
    ).toEqual({ foundName: null, foundPrice: null, foundImage: null });

    expect(parseEbayItem(null)).toEqual({
      foundName: null,
      foundPrice: null,
      foundImage: null,
    });
  });
});
