import { capture, scrollToBottom } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'distributions', async (page) => {
    await page.getByRole('link', { name: 'colorscheme' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await scrollToBottom(page);
  });
})();
