const { capture } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'flags', async (page) => {});
})();
