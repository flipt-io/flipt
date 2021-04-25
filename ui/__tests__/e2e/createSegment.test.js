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

test("createSegment", async () => {
  await page.goto(addr);
  await page.click("[data-testid='segments']");
  await page.click("[data-testid='new-segment']");
  await page.click("[placeholder='Segment name']");
  await page.type("[placeholder='Segment name']", "Power users");
  await page.click("[placeholder='Segment description']");
  await page.type("[placeholder='Segment description']", "Users that are willing to try out advanced functionality");
  await page.click("[data-testid='create-segment']");
  page = await waitForPage(context, 0, { waitUntil: "domcontentloaded" });
  await expect(page).toHaveText("power-users");
  await page.click('[href="#/segments/power-users]');
});

test('createSegmentDisallowSpecialChars', async () => {
  await page.goto(addr);
  await page.click("[data-testid='segments']");
  await page.click("[data-testid='new-segment']");
  await page.type("[placeholder='Segment name']", "My segment with colons");
  await page.type("[placeholder='Segment key']", "colons:are:not:allowed");
  await page.type("[placeholder='Segment description']", "This should not be saved");
  await expect(page).toHaveText("Only letters, numbers, hypens and underscores allowed");
});
