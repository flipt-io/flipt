const { capture } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'rollouts', async (page) => {
    await page.getByRole('link', { name: 'new-contact-page' }).click();
    await page.evaluate(() =>
      // scroll to bottom of page to show all rollouts
      window.scrollTo(0, document.documentElement.scrollHeight)
    );
  });
})();
