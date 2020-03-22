const qawolf = require("qawolf");
const selectors = require("./selectors/createFlag.json");

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
  await page.goto("http://0.0.0.0:8080/");
  await page.click(selectors["0_a"]);
  await page.click(selectors["1_flag_name_input"]);
  await page.type(selectors["2_flag_name_input"], "Awesome feature");
  await page.click(selectors["3_flag_descriptio_input"]);
  await page.type(
    selectors["4_flag_descriptio_input"],
    "Our product manager cannot wait to ship this feature!"
  );
  await page.click(selectors["8_button"]);
  await page.click(selectors["9_button"]);
  await page.click(selectors["10_a"]);
});
