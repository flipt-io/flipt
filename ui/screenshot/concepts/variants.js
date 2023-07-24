const { capture } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'variants', async (page) => {
    await page.getByRole('link', { name: 'colorscheme' }).click();
    await page.evaluate(() =>
      // scroll to bottom of page to show all variants
      window.scrollTo(0, document.documentElement.scrollHeight)
    );
  });
})();
