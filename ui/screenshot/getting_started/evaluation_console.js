import { capture } from '../../screenshot.js';

(async () => {
  await capture('getting_started', 'evaluation_console', async (page) => {
    await page.getByRole('link', { name: 'Console' }).click();
    await page.locator('#flagKey-select-button').click();
    await page.getByRole('option', { name: 'new-login New Login' }).click();
    await page.getByRole('button', { name: 'Evaluate', exact: true }).click();
  });
})();
