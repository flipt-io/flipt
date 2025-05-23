import { test, expect } from '@playwright/test';

test.describe('Branches', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('can create branch', async ({ page }) => {
    await page.getByTestId('create-branch-button').click();
    await page.getByRole('textbox', { name: 'New branch name' }).click();
    await page.getByRole('textbox', { name: 'New branch name' }).fill('foo');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Branch created successfully')).toBeVisible();
    await page.getByTestId('environment-namespace-switcher').getByRole('button').click();
    await page.getByTestId('environment-listbox').getByRole('button', { name: 'foo' }).click();
    await page.getByTestId('environment-listbox').getByRole('button', { name: 'default' }).click();
  });
});