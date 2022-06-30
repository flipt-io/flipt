import targeting from "@/mixins/targeting";
import reduce from "lodash/reduce";

const validRollout = targeting.methods.validRollout;

test("validRollout returns true if rollouts sum to less than 100", () => {
  let distributions = [{ rollout: 20 }, { rollout: 30 }, { rollout: 20 }];

  expect(validRollout(distributions)).toBe(true);
});

test("validRollout returns false if rollouts sum to greater than 100", () => {
  let distributions = [{ rollout: 80 }, { rollout: 30 }];

  expect(validRollout(distributions)).toBe(false);
});

test("validRollout returns true if rollouts sum to equal to 100", () => {
  let distributions = [{ rollout: 80.43 }, { rollout: 19.57 }];

  expect(validRollout(distributions)).toBe(true);
});

const computePercentages = targeting.methods.computePercentages;

test("computePercentages returns an array of 1 percentage for n = 1", () => {
  let n = 1;
  let expected = [100.0];
  let got = computePercentages(n);

  expect(got).toStrictEqual(expected);
});

test("computePercentages returns an array of percentages evenly distributed if 100 % n = 0", () => {
  let n = 5;
  let expected = [20.0, 20.0, 20.0, 20.0, 20.0];
  let got = computePercentages(n);

  expect(got).toStrictEqual(expected);

  let sum = reduce(
    got,
    function (acc, d) {
      return acc + d;
    },
    0
  );

  expect(sum.toFixed(2)).toBe("100.00");
});

test("computePercentages returns any array of percentages that add up to 100 if 100 % n != 0", () => {
  let n = 3;
  let expected = [33.34, 33.33, 33.33];
  let got = computePercentages(n);

  expect(got).toStrictEqual(expected);

  let sum = reduce(
    got,
    function (acc, d) {
      return acc + d;
    },
    0
  );

  expect(sum.toFixed(2)).toBe("100.00");
});

test("computePercentages returns an array of percentages that add up to 100 for n > 100", () => {
  let n = 101;
  let expected = [1];
  let got = computePercentages(n);

  // expected = [1, 0.99, 0.99, ...]
  for (let i = 1; i < n; i++) {
    expected.push(0.99);
  }

  expect(got).toStrictEqual(expected);

  let sum = reduce(
    got,
    function (acc, d) {
      return acc + d;
    },
    0
  );

  expect(sum.toFixed(2)).toBe("100.00");
});
