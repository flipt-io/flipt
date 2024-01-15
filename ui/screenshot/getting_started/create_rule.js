import { capture } from '../../screenshot.js';

(async () => {
  await capture('getting_started', 'create_rule', async (page) => {
    await page.getByRole('link', { name: 'new-login' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await page.getByRole('button', { name: 'New Rule' }).click();
    await page.locator('#segmentKey-0-select-button').click();
    await page.getByText('all-users').click();
    await page.getByLabel('Multi-Variate').check();
  });
})();
