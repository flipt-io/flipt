import { launch, register, stopVideos, waitForPage } from "qawolf";
import expect from "expect-playwright";

let browser;
let page;
let context;

const addr = "http://127.0.0.1:8080";

beforeAll(async () => {
  browser = await launch();
  context = await browser.newContext();
  await register(context);
  page = await context.newPage();
});

afterAll(async () => {
  await stopVideos();
  await browser.close();
});

test("createFlag", async () => {
  await page.goto(addr);
  await page.click("[data-testid='new-flag']");
  await page.type("[placeholder='Flag name']", "Awesome new feature");
  await page.type("[placeholder='Flag description']", "Our product manager cannot wait to ship this!");
  await page.click("[data-testid='create-flag']");
  page = await waitForPage(context, 0, { waitUntil: "domcontentloaded" });
  await expect(page).toHaveText("awesome-new-feature");
  await page.click('[href="#/flags/awesome-new-feature"]');
});

test('createFlagDisallowSpecialChars', async () => {
  await page.goto(addr);
  await page.click("[data-testid='new-flag']");
  await page.type("[placeholder='Flag name']", "My flag with colons");
  await page.type("[placeholder='Flag key']", "colons:are:not:allowed");
  await page.type("[placeholder='Flag description']", "This should not be saved");
  await expect(page).toHaveText("Only letters, numbers, hypens and underscores allowed");
});
