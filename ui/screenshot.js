const { chromium } = require('playwright');
const { exec } = require('child_process');
const fs = require('fs');

const fliptAddr = process.env.FLIPT_ADDRESS ?? 'http://localhost:8080';

const screenshot = async (page, name) => {
  await page.screenshot({ path: 'screenshots/'+name });
};

const sleep = (delay) => new Promise((resolve) => setTimeout(resolve, delay));

const capture = async function(name, fn) {
  try {
    const path = `${__dirname}/screenshot/fixtures/${name}.yml`;
    if (fs.existsSync(path)) {
      exec(`flipt import --address=${fliptAddr} ${path}`, (error, stdout, stderr) => {
        if (error) {
          console.error(`error: ${error.message}`);
          return;
        }

        if (stderr) {
          console.error(`stderr: ${stderr}`);
          return;
        }

        console.log(`stdout:\n${stdout}`);
      })
    }
  } catch (err) {
    // ignore and we will just skip seeding
    console.debug(err);
  }

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
  });
  const page = await context.newPage();

  await page.goto(fliptAddr);
  await fn(page);
  await sleep(2000);
  await screenshot(page, `${name}.png`);

  await context.close();
  await browser.close();
};

module.exports = { capture };
