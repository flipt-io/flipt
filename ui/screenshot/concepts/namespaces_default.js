const { capture } = require('../../screenshot.js');

(async () => {
  await capture('concepts', 'namespaces_default', async (page) => {});
})();
