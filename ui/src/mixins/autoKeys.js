export default {
  methods: {
    formatStringAsKey(str) {
      let temp = str.toLowerCase().split(/\s+/).join("-");

      // Auto generated keys  should not begin or end in a hyphen
      if (temp.charAt(0) == "-") {
        temp = temp.slice(1);
      }

      if (temp.charAt(temp.length - 1) == "-") {
        temp = temp.slice(0, -1);
      }

      return temp;
    },
  },
};
