import { expect, test } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/');
  await page.reload();
});

test('can create token', async ({ page }) => {
  await page.getByRole('link', { name: 'Settings' }).click();
  await page.getByRole('link', { name: 'API Tokens' }).click();
  await expect(
    page.getByRole('heading', { name: 'Static Tokens' })
  ).toBeVisible();
  await page.getByRole('button', { name: 'New Token' }).nth(0).click();
  await page.getByLabel('Name').fill('abcdef');
  await page.getByLabel('Description').fill('xyz');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: 'Copy' }).click();
  await expect(page.locator('pre')).toContainText('Copied to clipboard');
  await page.getByRole('button', { name: 'Close' }).click();
  await page.getByRole('cell', { name: 'abcdef', exact: true }).click();
});

test('can delete token', async ({ page }) => {
  await page.getByRole('link', { name: 'Settings' }).click();
  await page.getByRole('link', { name: 'API Tokens' }).click();
  await expect(
    page.getByRole('heading', { name: 'Static Tokens' })
  ).toBeVisible();
  await page.getByRole('link', { name: 'Delete , abcdef' }).click();
  await page.getByRole('button', { name: 'Delete' }).click();
});
