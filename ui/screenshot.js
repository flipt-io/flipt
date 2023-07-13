const { chromium } = require('playwright');

const fliptAddr = process.env.FLIPT_ADDRESS ?? 'http://localhost:8080';

const go = async (page, path) => {
  await page.goto(fliptAddr+path);
};

const screenshot = async (page, name) => {
  await page.screenshot({ path: 'screenshots/'+name });
};

const sleep = (delay) => new Promise((resolve) => setTimeout(resolve, delay));

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
  });
  const page = await context.newPage();

  // Create Flag
  await go(page, '/');
  await page.getByRole('button', { name: 'New Flag' }).click();
  await page.getByLabel('Name').fill('New Login');
  await page.getByLabel('Key').fill('new-login');
  await page.getByLabel('Description').fill('Enables the new login page for users');
  await screenshot(page, 'create_flag.png');

  await page.getByRole('button', { name: 'Create' }).click();

  // Create Variant
  await page.getByRole('button', { name: 'New Variant' }).click();
  await page
        .getByRole('dialog', { name: 'New Variant' })
        .locator('#key')
        .fill('big-blue-login-button');
  await page
        .getByRole('dialog', { name: 'New Variant' })
        .locator('#name')
        .fill('Big Blue Login Button');

  await sleep(2000);

  await screenshot(page, 'create_variant.png');
  await page.getByRole('button', { name: 'Create' }).click();

  await page.getByRole('link', { name: 'Segments' }).click();
  await page.getByRole('button', { name: 'New Segment' }).click();
  await page.getByLabel('Name').fill('All Users');
  await page.getByLabel('Key').fill('all-users');

  await sleep(2000);

  await screenshot(page, 'create_segment.png');

  await context.close();
  await browser.close();
})();
