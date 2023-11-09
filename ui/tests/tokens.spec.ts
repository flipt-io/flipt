import { expect, test } from '@playwright/test';

test.describe('API Tokens', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.reload();
  });

  test.describe('can create tokens', () => {
    test('default token', async ({ page }) => {
      await page.getByRole('link', { name: 'Settings' }).click();
      await page.getByRole('link', { name: 'API Tokens' }).click();
      await expect(
        page.getByRole('heading', { name: 'Static Tokens' })
      ).toBeVisible();
      await page.getByRole('button', { name: 'New Token' }).nth(0).click();
      await page.getByLabel('Name', { exact: true }).fill('abcdef');
      await page.getByLabel('Description').fill('xyz');
      await page.getByRole('button', { name: 'Create' }).click();
      await page.getByRole('button', { name: 'Copy' }).click();
      await expect(page.locator('pre')).toContainText('Copied to clipboard');
      await page.getByRole('button', { name: 'Close', exact: true }).click();
      await page.getByRole('cell', { name: 'abcdef', exact: true }).click();
    });

    test('expiring/namespaced token', async ({ page }) => {
      await page.getByRole('link', { name: 'Settings' }).click();
      await page.getByRole('link', { name: 'API Tokens' }).click();
      await expect(
        page.getByRole('heading', { name: 'Static Tokens' })
      ).toBeVisible();
      await page.getByRole('button', { name: 'New Token' }).nth(0).click();
      await page.getByLabel('Name', { exact: true }).fill('fooo');
      await page.getByLabel('Description').press('Tab');
      await page.getByLabel('Expires On').fill('2030-11-01');
      await page.getByLabel('Scope this token to a single namespace').check();
      await page.getByRole('button', { name: 'Default' }).click();
      await page.getByLabel('Default').getByText('Default').click();
      await page.getByRole('button', { name: 'Create' }).click();
      await page.getByRole('button', { name: 'Copy' }).click();
      await expect(page.locator('pre')).toContainText('Copied to clipboard');
      await page.getByRole('button', { name: 'Close', exact: true }).click();
      await page.getByRole('cell', { name: 'fooo', exact: true }).click();
    });
  });

  test('can delete token', async ({ page }) => {
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'API Tokens' }).click();
    await expect(
      page.getByRole('heading', { name: 'Static Tokens' })
    ).toBeVisible();
    await page.getByRole('link', { name: 'Delete , abcdef' }).click();
    await page.getByRole('button', { name: 'Delete', exact: true }).click();
  });
});
