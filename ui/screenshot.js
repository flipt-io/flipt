const { chromium } = require('playwright');
const { exec } = require('child_process');
const fs = require('fs');

const fliptAddr = process.env.FLIPT_ADDRESS ?? 'http://localhost:8080';

const screenshot = async (page, name) => {
  await page.screenshot({ path: 'screenshots/' + name });
};

const scrollToBottom = async (page) => {
  await sleep(1000);
  await page.evaluate(() => {
    window.scrollTo(0, document.body.scrollHeight);
  });
};

const sleep = (delay) => new Promise((resolve) => setTimeout(resolve, delay));

const capture = async function (folder, name, fn, opts = {}) {
  if (!opts.namespace) {
    opts.namespace = 'default';
  }

  try {
    const path = `${__dirname}/screenshot/${folder}/fixtures/${name}.yml`;
    if (fs.existsSync(path)) {
      exec(
        `flipt import --create-namespace --namespace=${opts.namespace} --address=${fliptAddr} ${path}`,
        (error, stdout, stderr) => {
          if (error) {
            console.error(`error: ${error.message}`);
            return;
          }

          if (stderr) {
            console.error(`stderr: ${stderr}`);
            return;
          }

          console.log(`stdout:\n${stdout}`);
        }
      );
    }
  } catch (err) {
    // ignore and we will just skip seeding
    console.debug(err);
  }

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 3
  });
  const page = await context.newPage();

  await page.goto(fliptAddr);
  await fn(page);
  await sleep(4000);
  await screenshot(page, `${folder}/${name}.png`);

  await context.close();
  await browser.close();
};

module.exports = { capture, sleep, screenshot, scrollToBottom };
