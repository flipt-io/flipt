const qawolf = require("qawolf");
const selectors = require("./selectors/createSegment.json");

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
  await page.goto("http://0.0.0.0:8080/");
  // click segment nav bar
  await page.click(selectors["0_a"]);
  // new segment button
  await page.click(selectors["1_a"]);
  await page.click(selectors["2_segment_name_input"]);
  await page.type(selectors["5_segment_name_input"], "Power users");
  await page.press(selectors["6_segment_name_input"], "Tab");
  await page.fill(selectors["7_segment_key_input"], "power-users");
  await page.press(selectors["8_segment_key_input"], "Tab");
  await page.type(
    selectors["9_segment_descrip_input"],
    "Users that are willing to try out advanced functionality"
  );
  // create segment button
  await page.click(selectors["10_button"]);
  // click segment nav bar
  await page.click(selectors["11_a"]);
});
