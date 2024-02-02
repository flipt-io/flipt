import { expect, test } from '@playwright/test';

test.describe('Onboarding', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Support' }).click();
    await page.getByRole('heading', { name: 'Onboarding' }).click();
    await page.getByRole('link', { name: "Let's Go" }).click();
  });

  test('has expected tiles', async ({ page }) => {
    await expect(page.getByText('Get Started', { exact: true })).toBeVisible();
    await expect(page.getByText('Try the CLI', { exact: true })).toBeVisible();
    await expect(
      page.getByText('Checkout a Guide', { exact: true })
    ).toBeVisible();
    await expect(
      page.getByText('Integrate Your Application', { exact: true })
    ).toBeVisible();
    await expect(
      page.getByText('Join the Community', { exact: true })
    ).toBeVisible();
    await expect(
      page.getByText('View API Reference', { exact: true })
    ).toBeVisible();
    await expect(page.getByText('Support Us', { exact: true })).toBeVisible();
  });
});
