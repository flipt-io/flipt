import { expect, test } from '@playwright/test';

test.describe('Root', () => {
  test('has title', async ({ page }) => {
    await page.goto('/');

    // Expect a title "to contain" a substring.
    await expect(page).toHaveTitle(/Flipt/);
  });
});

test.describe('Root - Read Only', () => {
  // skip tests
  test.skip(
    true,
    'skip all read only tests until we can get the state/mock or API working'
  );

  test('has title and readonly message', async ({ page }) => {
    await page.goto('/');

    // Expect a title "to contain" a substring.
    await expect(page).toHaveTitle(/Flipt/);
    // Expect readonly message to be visible
    await expect(page.getByText('Read-Only')).toBeVisible();
  });
});
