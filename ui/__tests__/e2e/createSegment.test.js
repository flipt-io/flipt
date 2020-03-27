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

test("createSegment", async () => {
  await page.goto("localhost:8080");
  await page.click("[data-testid='segments']");
  await page.click("[data-testid='new-segment']");
  await page.click("[placeholder='Segment name']");
  await page.type("[placeholder='Segment name']", "Power users");
  await page.click("[placeholder='Segment description']");
  await page.type(
    "[placeholder='Segment description']",
    "Users that are willing to try out advanced functionality"
  );
  await page.click("[data-testid='create-segment']");
});
