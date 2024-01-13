import { capture } from '../../screenshot.js';

(async () => {
  await capture('getting_started', 'create_variant', async (page) => {
    await page.getByRole('link', { name: 'new-login' }).click();
    await page.getByRole('button', { name: 'New Variant' }).click();
    await page
      .getByRole('dialog', { name: 'New Variant' })
      .locator('#key')
      .fill('big-blue-login-button');
    await page
      .getByRole('dialog', { name: 'New Variant' })
      .locator('#name')
      .fill('Big Blue Login Button');
  });
})();
