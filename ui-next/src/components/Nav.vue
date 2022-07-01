<template>
  <nav
    class="navbar is-fixed-top is-primary"
    role="navigation"
    aria-label="main navigation"
  >
    <div class="navbar-brand">
      <a class="navbar-item" href="#">
        <h1 class="title has-text-white">Flipt</h1>
      </a>
      <a
        role="button"
        class="navbar-burger"
        :class="{ 'is-active': isActive }"
        aria-label="menu"
        aria-expanded="false"
        @click="isActive = !isActive"
      >
        <span aria-hidden="true" /><span aria-hidden="true" /><span
          aria-hidden="true"
        />
      </a>
    </div>
    <div class="navbar-menu" :class="{ 'is-active': isActive }">
      <div class="navbar-start">
        <RouterLink
          class="navbar-item has-text-weight-semibold"
          data-testid="flags"
          :to="{ name: 'flags' }"
          >Flags</RouterLink
        >
        <RouterLink
          class="navbar-item has-text-weight-semibold"
          data-testid="segments"
          :to="{ name: 'segments' }"
          >Segments</RouterLink
        >
      </div>
      <div class="navbar-end">
        <a
          class="navbar-item has-text-weight-semibold"
          target="_blank"
          href="https://flipt.io/docs/getting_started/?utm_source=app"
        >
          Documentation
        </a>
        <a
          class="navbar-item has-text-weight-semibold"
          target="_blank"
          href="/docs/"
          >API</a
        >
        <a
          v-if="ref"
          class="navbar-item is-size-7 has-text-weight-semibold"
          target="_blank"
          :href="refURL"
        >
          {{ ref }}
        </a>
      </div>
    </div>
  </nav>
</template>

<script>
import { mapGetters } from "vuex";

export default {
  Name: "nav",
  data() {
    return {
      isActive: false,
    };
  },
  computed: {
    ...mapGetters(["info"]),
    refURL() {
      if (this.info.isRelease && this.info.version) {
        return (
          "https://github.com/markphelps/flipt/releases/tag/" +
          this.info.version
        );
      } else if (this.info.commit) {
        return "https://github.com/markphelps/flipt/commit/" + this.info.commit;
      }
      return "https://github.com/markphelps/flipt";
    },
    ref() {
      if (this.info.isRelease && this.info.version) {
        return this.info.version;
      } else if (this.info.commit) {
        return this.info.commit.substring(0, 7);
      }
      return "";
    },
  },
  mounted() {
    this.$store.dispatch("getInfo");
  },
};
</script>
