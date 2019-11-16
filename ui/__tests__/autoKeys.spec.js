import autoKeys from "@/mixins/autoKeys";

const formatStringAsKey = autoKeys.methods.formatStringAsKey;

test("formatKeyAsString replace space with hyphen", () => {
  let s = "test key";
  expect(formatStringAsKey(s)).toBe("test-key");
});

test("formatKeyAsString removes beginning hyphens", () => {
  let s = "   test key";
  expect(formatStringAsKey(s)).toBe("test-key");
});

test("formatKeyAsString removes trailing hyphens", () => {
  let s = "test key   ";
  expect(formatStringAsKey(s)).toBe("test-key");
});

test("formatKeyAsString removes beginning and trailing hyphens", () => {
  let s = "   test key   ";
  expect(formatStringAsKey(s)).toBe("test-key");
});

test("formatKeyAsString uses one dash for multiple hyphens", () => {
  let s = "test    key";
  expect(formatStringAsKey(s)).toBe("test-key");
});
