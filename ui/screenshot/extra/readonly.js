import { capture } from '../../screenshot.js';

(async () => {
  await capture('extra', 'readonly', async (page) => {
    await page.route(/\/meta\/config/, async (route) => {
      const response = await route.fetch();
      const json = await response.json();
      json.storage = { type: 'git' };
      // Fulfill using the original response, while patching the
      // response body with our changes to mock git storage for read only mode
      await route.fulfill({ response, json });
    });

    await page.getByRole('link', { name: 'Flags' }).click();
  });
})();
