import { expect, test } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/');
  await page.reload();
});

test('can change timezone preferences', async ({ page }) => {
  await page.getByRole('link', { name: 'Settings' }).click();
  await page.getByRole('link', { name: 'Preferences' }).click();
  await page.getByTestId('switch-timezone').click();
  await expect(page.getByText('Preferences saved')).toBeVisible();
});

test('can change theme preferences', async ({ page }) => {
  await page.getByRole('link', { name: 'Settings' }).click();
  await page.getByRole('link', { name: 'Preferences' }).click();
  await page.getByTestId('select-theme').selectOption('Dark');
  await expect(page.getByText('Preferences saved')).toBeVisible();
});
