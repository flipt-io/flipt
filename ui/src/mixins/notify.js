export default {
  methods: {
    notifySuccess(msg) {
      this.$toast.open({
        type: "is-primary",
        position: "is-bottom",
        message: msg
      });
    },
    notifyError(msg) {
      this.$toast.open({
        type: "is-danger",
        position: "is-bottom",
        message: msg,
        duration: 3000
      });
    }
  }
};
