import { expect, test } from '@playwright/test';

test.describe('Onboarding', () => {
  test.describe('on first run', () => {
    test('has expected content', async ({ page }) => {
      await page.goto('/');

      await expect(page.locator('h1')).toContainText('Onboarding');
      await expect(
        page.getByRole('heading', { name: 'Get Started' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', {
          name: 'Introducing Flipt Cloud'
        })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Try the CLI' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Checkout a Guide' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Integrate Your Application' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Join the Community' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'View API Reference' })
      ).toBeVisible();
      await expect(
        page.getByRole('heading', { name: 'Support Us' })
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

  test.describe('user navigates to the onboarding page', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/');
      await page.getByRole('link', { name: 'Support', exact: true }).click();
      await page.getByRole('heading', { name: 'Onboarding' }).click();
      await page.getByRole('link', { name: "Let's Go" }).click();
    });

    test('has expected content', async ({ page }) => {
      await expect(page.locator('h1')).toContainText('Onboarding');
      await expect(
        page.getByText('Get Started', { exact: true })
      ).toBeVisible();
      await expect(
        page.getByText('Try the CLI', { exact: true })
      ).toBeVisible();
      await expect(
        page.getByText('Checkout a Guide', { exact: true })
      ).toBeVisible();
      await expect(
        page.getByText('Integrate Your Application', { exact: true })
      ).toBeVisible();
      await expect(
        page.getByText('Chat With Us', { exact: true })
      ).toBeVisible();
      await expect(
        page.getByText('View API Reference', { exact: true })
      ).toBeVisible();
      await expect(page.getByText('Support Us', { exact: true })).toBeVisible();
      await expect(
        page.getByText('Join the Community', { exact: true })
      ).toBeVisible();
      await expect(
        page.getByRole('button', { name: 'Continue to Dashboard' })
      ).toBeHidden();
    });
  });
});
