import { capture, scrollToBottom } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'variants', async (page) => {
    await page.getByRole('link', { name: 'colorscheme' }).click();
    await scrollToBottom(page);
  });
})();
