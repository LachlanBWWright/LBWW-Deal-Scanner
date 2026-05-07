import type { TestingNotificationRequest } from "./types.js";

export const notificationPresets: Array<{
  label: string;
  preset: TestingNotificationRequest;
}> = [
  {
    label: "eBay deal",
    preset: {
      kind: "deal",
      source: "ebay",
      title: "Test eBay item listing at $49.99",
      url: "https://www.ebay.com.au/itm/test",
      price: 49.99,
    },
  },
  {
    label: "Cash Converters deal",
    preset: {
      kind: "deal",
      source: "cashConverters",
      title: "Test Cash Converters item for $29.99",
      url: "https://www.cashconverters.com.au/products/test",
      price: 29.99,
    },
  },
  {
    label: "Steam Market deal",
    preset: {
      kind: "deal",
      source: "steamMarket",
      title: "Test Steam item listed at $12.50",
      url: "https://steamcommunity.com/market/listings/730/Test%20Item",
      price: 12.5,
    },
  },
  {
    label: "Gumtree deal",
    preset: {
      kind: "deal",
      source: "gumtree",
      title: "Test Gumtree listing for $75.00",
      url: "https://www.gumtree.com.au/s-ad/test/test-item/1234567890",
      price: 75.0,
    },
  },
  {
    label: "Scanner error",
    preset: {
      kind: "error",
      source: "TestScanner",
      message: "Simulated scanner error for testing",
    },
  },
];
