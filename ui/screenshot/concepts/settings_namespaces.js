import { capture } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'settings_namespaces', async (page) => {
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'Namespaces' }).click();
    await page.getByRole('button', { name: 'New Namespace' }).click();
    await page.getByLabel('Name', { exact: true }).fill('Production');
    await page.getByLabel('Description').fill('Production Environment');
    await page.getByRole('button', { name: 'Create' }).click();
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'Namespaces' }).click();
  });
})();
