const { capture } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'segments', async (page) => {
    await page.getByRole('link', { name: 'Segments' }).click();
  });
})();
