import { expect, test } from '@playwright/test';

test.describe('Root', () => {
  test('has title', async ({ page }) => {
    await page.goto('/');

    // Expect a title "to contain" a substring.
    await expect(page).toHaveTitle(/Flipt/);
  });
});
