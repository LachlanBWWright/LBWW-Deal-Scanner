import { describe, it, expect, vi } from "vitest";
import selectorRace from "../functions/selectorRace.js";

// Mock puppeteer Page
const createMockPage = () => ({
  waitForSelector: vi.fn(),
});

describe("selectorRace", () => {
  it("should return element when found selector appears first", async () => {
    const mockPage = createMockPage();
    const mockElement = { tagName: "div" };
    
    // Mock found selector to resolve quickly
    mockPage.waitForSelector.mockImplementation((selector) => {
      if (selector === ".found-item") {
        return Promise.resolve(mockElement);
      }
      return new Promise(() => {}); // Never resolves for noItem selector
    });

    const result = await selectorRace(mockPage as any, ".found-item", ".no-items");
    
    expect(result).toBe(mockElement);
  });

  it("should return null when no items selector appears first", async () => {
    const mockPage = createMockPage();
    
    // Mock no items selector to resolve quickly
    mockPage.waitForSelector.mockImplementation((selector) => {
      if (selector === ".no-items") {
        return Promise.resolve({});
      }
      return new Promise(() => {}); // Never resolves for found selector
    });

    const result = await selectorRace(mockPage as any, ".found-item", ".no-items");
    
    expect(result).toBe(null);
  });

  it("should return null when found selector fails", async () => {
    const mockPage = createMockPage();
    
    // Mock found selector to reject
    mockPage.waitForSelector.mockImplementation((selector) => {
      if (selector === ".found-item") {
        return Promise.reject(new Error("Selector not found"));
      }
      return new Promise(() => {}); // Never resolves for noItem selector
    });

    const result = await selectorRace(mockPage as any, ".found-item", ".no-items");
    
    expect(result).toBe(null);
  });

  it("should return null when no items selector fails", async () => {
    const mockPage = createMockPage();
    
    // Mock both selectors to fail
    mockPage.waitForSelector.mockRejectedValue(new Error("Selector not found"));

    const result = await selectorRace(mockPage as any, ".found-item", ".no-items");
    
    expect(result).toBe(null);
  });
});