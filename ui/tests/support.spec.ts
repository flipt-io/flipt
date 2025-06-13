import { expect, test } from '@playwright/test';

test.describe('Onboarding', () => {
  test.describe('on first run', () => {
    test('has expected content', async ({ page }) => {
      await page.goto('/');

      await expect(page.locator('h1')).toContainText('Onboarding');
      await expect(
        page.getByRole('heading', { name: 'Quick Start' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Checkout a Guide' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Integrate Your Application' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Changelog' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Chat With Us' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Support Us' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Join the Community' })
      ).toBeVisible();
      await expect(page.getByRole('heading', { name: 'Email' })).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Report an issue' })
      ).toBeVisible();
      await expect(
        page.getByRole('button', { name: 'Continue to Dashboard' })
      ).toBeVisible();
    });

    test('can continue to dashboard', async ({ page }) => {
      await page.goto('/');
      await page.getByRole('button', { name: 'Continue to Dashboard' }).click();
      await expect(page.locator('h1')).toContainText('Flags');
      expect(page.url()).toContain('/flags');
    });
  });

  test.describe('user navigates to the Support page', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/');
      await page.getByRole('link', { name: 'Get Help', exact: true }).click();
    });

    test('has expected content', async ({ page }) => {
      await expect(page.locator('h1')).toContainText('Support');
      await expect(
        page.getByRole('heading', { name: 'Quick Start' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Checkout a Guide' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Integrate Your Application' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Changelog' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Chat With Us' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Support Us' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Join the Community' })
      ).toBeVisible();
      await expect(page.getByRole('heading', { name: 'Email' })).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Report an issue' })
      ).toBeVisible();
      await expect(
        page.getByRole('button', { name: 'Continue to Dashboard' })
      ).toBeHidden();
    });
  });
});
