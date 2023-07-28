const { capture } = require('../../screenshot.js');

(async () => {
  await capture('extra', 'darkmode', async (page) => {
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByLabel('Theme').selectOption('dark');
    await page.getByLabel('UTC TimezoneDisplay dates and times in UTC timezone').click();
    await page.getByRole('link', { name: 'Flags' }).click();
  });
})();
