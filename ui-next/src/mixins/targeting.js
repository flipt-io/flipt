import reduce from "lodash/reduce";

export default {
  methods: {
    validRollout(distributions) {
      let sum = reduce(
        distributions,
        function (acc, d) {
          return acc + Number(d.rollout);
        },
        0
      );
      return sum <= 100;
    },

    computePercentages(n) {
      let sum = 100 * 100;

      let d = Math.floor(sum / n);
      let remainder = sum - d * n;

      let result = [];
      let i = 0;

      while (++i && i <= n) {
        result.push((i <= remainder ? d + 1 : d) / 100);
      }

      return result;
    },
  },
};
