export default {
  methods: {
    notifySuccess(msg) {
      this.$buefy.toast.open({
        type: "is-primary",
        position: "is-bottom",
        message: msg,
      });
    },
    notifyError(msg) {
      this.$buefy.toast.open({
        type: "is-danger",
        position: "is-bottom",
        message: msg,
        duration: 3000,
      });
    },
  },
};
