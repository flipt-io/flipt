import { capture } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'settings_tokens', async (page) => {
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'API Tokens' }).click();
    await page.getByRole('button', { name: 'New Token' }).click();
    await page.getByLabel('Name', { exact: true }).fill('Production');
    await page.getByLabel('Description').fill('Production API Token');
  });
})();
