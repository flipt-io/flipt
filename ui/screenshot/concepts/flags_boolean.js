import { capture } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'flags_boolean', async (page) => {
    await page.getByRole('button', { name: 'New Flag' }).click();
    await page.getByLabel('Name').fill('New Contact Page');
    await page.getByLabel('Boolean').check();
    await page.getByRole('button', { name: 'Create' }).click();
  });
})();
