const { capture, scrollToBottom } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'variants', async (page) => {
    await page.getByRole('link', { name: 'colorscheme' }).click();
    await scrollToBottom(page);
  });
})();
