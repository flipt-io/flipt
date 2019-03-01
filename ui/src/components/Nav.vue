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
          :to="{ name: 'flags' }"
          >Flags</RouterLink
        >
        <RouterLink
          class="navbar-item has-text-weight-semibold"
          :to="{ name: 'segments' }"
          >Segments</RouterLink
        >
      </div>
      <div class="navbar-end">
        <a
          class="navbar-item has-text-weight-semibold"
          href="https://flipt.dev/"
          >Documentation</a
        >
        <a class="navbar-item has-text-weight-semibold" href="/docs/">API</a>
      </div>
    </div>
  </nav>
</template>

<script>
export default {
  Name: "nav",
  data() {
    return {
      isActive: false
    };
  }
};
</script>
