import { capture } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'segments', async (page) => {
    await page.getByRole('link', { name: 'Segments' }).click();
  });
})();
