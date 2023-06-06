import { expect, test } from '@playwright/test';

test.describe('Flags - Read Only', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Flags' }).click();
  });

  test('can not create flag', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'New Flag' })).toBeDisabled();
  });

  test('can not update flag', async ({ page }) => {
    await page.getByRole('link', { name: 'test-flag' }).click();
    await page.getByLabel('Description').click();
    await page.getByLabel('Description').fill('Test flag description 2');
    await expect(page.getByRole('button', { name: 'Update' })).toBeDisabled();
  });

  test('can not add variants to flag', async ({ page }) => {
    await page.getByRole('link', { name: 'test-flag' }).click();

    await expect(
      page.getByRole('button', { name: 'New Variant' })
    ).toBeDisabled();
  });
});
