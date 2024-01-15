import { capture } from '../../screenshot.js';

(async () => {
  await capture('getting_started', 'create_flag', async (page) => {
    await page.getByRole('button', { name: 'New Flag' }).click();
    await page.getByLabel('Name').fill('New Login');
    await page.getByLabel('Key').fill('new-login');
    await page
      .getByLabel('Description')
      .fill('Enables the new login page for users');
  });
})();
