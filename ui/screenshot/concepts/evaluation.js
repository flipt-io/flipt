const { capture } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'evaluation', async (page) => {
    await page.getByRole('link', { name: 'Console' }).click();
    await page.locator('#flagKey-select-button').click();
    await page.getByRole('option', { name: 'colorscheme Color Scheme' }).click();
    await page.getByPlaceholder('{}').click();
    await page.getByPlaceholder('{}').fill('{\n\t"finished_onboarding":"false"\n}');
    await page.getByRole('button', { name: 'Evaluate', exact: true }).click();
  });
})();
