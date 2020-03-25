const qawolf = require("qawolf");

let browser;
let page;

beforeAll(async () => {
  browser = await qawolf.launch();
  const context = await browser.newContext();
  await qawolf.register(context);
  page = await context.newPage();
});

afterAll(async () => {
  await qawolf.stopVideos();
  await browser.close();
});

test("createFlag", async () => {
  await page.goto("localhost:8080");
  await page.click("[data-testid='new-flag']");
  await page.click("[placeholder='Flag name']");
  await page.type("[placeholder='Flag name']", "Awesome new feature");
  await page.click("[placeholder='Flag description']");
  await page.type(
    "[placeholder='Flag description']",
    "Our product manager cannot wait to ship this!"
  );
  await page.click("[data-testid='create-flag']");
});
