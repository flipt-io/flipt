const { test, expect } = require("@playwright/test");

test("createSegment", async ({ page }) => {
  await page.goto("/");
  await page.click("[data-testid='segments']");
  await page.click("[data-testid='new-segment']");
  await page.click("[placeholder='Segment name']");
  await page.type("[placeholder='Segment name']", "Power users");
  await page.click("[placeholder='Segment description']");
  await page.type(
    "[placeholder='Segment description']",
    "Users that are willing to try out advanced functionality"
  );
  await page.click("[data-testid='create-segment']");
  await page.click('[aria-label="breadcrumbs"] .router-link-active');
  const locator = page.locator("a", { hasText: "power-users" });
  await expect(locator).toBeVisible();
  await locator.click();
});

test("createSegmentDisallowSpecialChars", async ({ page }) => {
  await page.goto("/");
  await page.click("[data-testid='segments']");
  await page.click("[data-testid='new-segment']");
  await page.type("[placeholder='Segment name']", "My segment with colons");
  await page.type("[placeholder='Segment key']", "colons:are:not:allowed");
  await page.type(
    "[placeholder='Segment description']",
    "This should not be saved"
  );
  const locator = page.locator("p", {
    hasText: "Only letters, numbers, hypens and underscores allowed",
  });
  await expect(locator).toBeVisible();
});
