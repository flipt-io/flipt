const { capture } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'distributions', async (page) => {
    await page.getByRole('link', { name: 'colorscheme' }).click();
    await page.getByRole('link', { name: 'Evaluation' }).click(); 
  });
})();
