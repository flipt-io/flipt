import { expect, test } from '@playwright/test';

test.describe('Root - Read Only', () => {
  test('has title and readonly message', async ({ page }) => {
    await page.goto('/');

    // Expect a title "to contain" a substring.
    await expect(page).toHaveTitle(/Flipt/);
    // Expect readonly message to be visible
    await expect(page.getByText('Read-Only')).toBeVisible();
  });
});
