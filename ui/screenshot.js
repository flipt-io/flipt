import { exec } from 'child_process';
import { existsSync } from 'fs';
import { chromium } from 'playwright';

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
  try {
    const path = `${__dirname}/screenshot/${folder}/fixtures/${name}.yml`;
    if (existsSync(path)) {
      exec(
        `flipt import --address=${fliptAddr} ${path}`,
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
  await sleep(5000);
  let random = Math.floor(Math.random() * 100000);
  await screenshot(page, `${folder}/${name}${random}.png`);

  await context.close();
  await browser.close();
};

export default { capture, sleep, screenshot, scrollToBottom };
