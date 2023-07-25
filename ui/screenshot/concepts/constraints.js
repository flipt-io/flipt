const { capture } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'constraints', async (page) => {
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('link', { name: 'new-users' }).click();
    await page.evaluate(() =>
      // scroll to bottom of page to show all constraints
      window.scrollTo(0, document.documentElement.scrollHeight)
    );
  });
})();
