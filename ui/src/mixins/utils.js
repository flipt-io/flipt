export default {
  methods: {
    isPresent(value) {
      return value !== undefined && value !== "";
    },
    isEmpty(value) {
      return value === undefined || value === "";
    },
  },
};
