const { capture } = require('../screenshot.js');

(async () => {
  await capture('create_segment', async (page) => {
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('button', { name: 'New Segment' }).click();
    await page.getByLabel('Name').fill('All Users');
    await page.getByLabel('Key').fill('all-users');
  });
})();
