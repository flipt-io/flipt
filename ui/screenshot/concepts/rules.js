const { capture } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'rules', async (page) => {
    await page.getByRole('link', { name: 'colorscheme' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
  });
})();
