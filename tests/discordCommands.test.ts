import { describe, it, expect } from "vitest";
import createEbayQuery from "../commandManager/commandList/createEbayQuery.js";
import createCashQuery from "../commandManager/commandList/createCashQuery.js";
import createMultiSearchQuery from "../commandManager/commandList/createMultiSearchQuery.js";

describe("Discord Slash Commands", () => {
  describe("createEbayQuery", () => {
    it("should have correct command name and description", () => {
      const command = createEbayQuery.toJSON();
      
      expect(command.name).toBe("createebayquery");
      expect(command.description).toBe("Creates a search for an eBay Query");
    });

    it("should have required string option for query", () => {
      const command = createEbayQuery.toJSON();
      const queryOption = command.options?.find((opt: any) => opt.name === "query");
      
      expect(queryOption).toBeDefined();
      expect(queryOption.type).toBe(3); // STRING type
      expect(queryOption.description).toBe("The URL of the query. Sort by newest first.");
      expect(queryOption.required).toBe(true);
    });

    it("should have required number option for maxprice", () => {
      const command = createEbayQuery.toJSON();
      const maxPriceOption = command.options?.find((opt: any) => opt.name === "maxprice");
      
      expect(maxPriceOption).toBeDefined();
      expect(maxPriceOption.type).toBe(10); // NUMBER type
      expect(maxPriceOption.description).toBe("Enter the maximum price (in AUD) a notification.");
      expect(maxPriceOption.required).toBe(true);
    });
  });

  describe("createCashQuery", () => {
    it("should have correct command name and description", () => {
      const command = createCashQuery.toJSON();
      
      expect(command.name).toBe("createcashquery");
      expect(command.description).toBe("Creates a search for a Cash Converters Query");
    });

    it("should have required string option for query", () => {
      const command = createCashQuery.toJSON();
      const queryOption = command.options?.find((opt: any) => opt.name === "query");
      
      expect(queryOption).toBeDefined();
      expect(queryOption.type).toBe(3); // STRING type
      expect(queryOption.description).toBe("The URL of the query. Sort by price or newness.");
      expect(queryOption.required).toBe(true);
    });
  });

  describe("createMultiSearchQuery", () => {
    it("should have correct command name and description", () => {
      const command = createMultiSearchQuery.toJSON();
      
      expect(command.name).toBe("createmultisearch");
      expect(command.description).toBe("Creates a search for a CS:GO item on multiple trade bots");
    });

    it("should have required string option for skinname", () => {
      const command = createMultiSearchQuery.toJSON();
      const skinNameOption = command.options?.find((opt: any) => opt.name === "skinname");
      
      expect(skinNameOption).toBeDefined();
      expect(skinNameOption.type).toBe(3); // STRING type
      expect(skinNameOption.required).toBe(true);
    });

    it("should have required number options for float values", () => {
      const command = createMultiSearchQuery.toJSON();
      const minFloatOption = command.options?.find((opt: any) => opt.name === "minfloat");
      const maxFloatOption = command.options?.find((opt: any) => opt.name === "maxfloat");
      
      expect(minFloatOption).toBeDefined();
      expect(minFloatOption.type).toBe(10); // NUMBER type
      expect(minFloatOption.required).toBe(true);
      
      expect(maxFloatOption).toBeDefined();
      expect(maxFloatOption.type).toBe(10); // NUMBER type
      expect(maxFloatOption.required).toBe(true);
    });

    it("should have optional number option for maxprice", () => {
      const command = createMultiSearchQuery.toJSON();
      const maxPriceOption = command.options?.find((opt: any) => opt.name === "maxprice");
      
      expect(maxPriceOption).toBeDefined();
      expect(maxPriceOption.type).toBe(10); // NUMBER type
      expect(maxPriceOption.required).toBe(false);
    });
  });
});