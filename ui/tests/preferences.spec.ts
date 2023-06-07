import { test } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/');
  await page.reload();
});

test('can change timezone preferences', async ({ page }) => {
  await page.getByRole('link', { name: 'Settings' }).click();
  await page.getByRole('link', { name: 'General' }).click();
  await page.getByRole('heading', { name: 'Preferences' }).click();
  await page
    .getByRole('switch', {
      name: 'UTC Timezone Display dates and times in UTC timezone'
    })
    .click();
});
