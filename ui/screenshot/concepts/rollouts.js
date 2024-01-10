import { capture, scrollToBottom } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'rollouts', async (page) => {
    await page.getByRole('link', { name: 'new-contact-page' }).click();
    await scrollToBottom(page);
  });
})();
