import { capture, scrollToBottom } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'constraints', async (page) => {
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('link', { name: 'new-users' }).click();
    await scrollToBottom(page);
  });
})();
