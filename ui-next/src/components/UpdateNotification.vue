<template>
  <div v-if="show" id="update" class="notification is-success">
    <button id="update-close" class="delete" @click="close"></button>
    <p class="content">
      <a target="_blank" :href="releaseURL">Flipt {{ info.latestVersion }}</a>
      is available
    </p>
  </div>
</template>

<script>
import { mapGetters } from "vuex";

export default {
  data() {
    return {
      isHidden: false,
    };
  },
  computed: {
    ...mapGetters(["info"]),
    releaseURL() {
      return (
        "https://github.com/markphelps/flipt/releases/tag/" +
        this.info.latestVersion
      );
    },
    show() {
      return this.info.isRelease && this.info.updateAvailable && !this.isHidden;
    },
  },
  methods: {
    close() {
      this.isHidden = true;
    },
  },
};
</script>

<style scoped>
#update {
  z-index: 99999;
  position: fixed;
  display: flex;
  flex-direction: column;
  right: 0;
  bottom: 0;
  align-items: flex-end;
}
</style>
