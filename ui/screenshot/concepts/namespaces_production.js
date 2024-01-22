import { capture } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'namespaces_production', async (page) => {
    await page.getByRole('button', { name: 'Default' }).click();
    await page.getByText('production').click();
  });
})();
