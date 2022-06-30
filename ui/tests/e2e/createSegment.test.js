import { test, expect } from "@playwright/test";

const addr = "http://127.0.0.1:8080";

test("createSegment", async ({ page }) => {
  await page.goto(addr);
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
  await expect(page).toHaveText("power-users");
  await page.click('[href="#/segments/power-users"]');
});

test("createSegmentDisallowSpecialChars", async ({ page }) => {
  await page.goto(addr);
  await page.click("[data-testid='segments']");
  await page.click("[data-testid='new-segment']");
  await page.type("[placeholder='Segment name']", "My segment with colons");
  await page.type("[placeholder='Segment key']", "colons:are:not:allowed");
  await page.type(
    "[placeholder='Segment description']",
    "This should not be saved"
  );
  await expect(page).toHaveText(
    "Only letters, numbers, hypens and underscores allowed"
  );
});
