<template>
  <footer class="footer">
    <div class="content has-text-centered">
      <p>
        <strong>Flipt</strong> version
        <a>{{ version }}</a>
      </p>
    </div>
  </footer>
</template>

<script>
export default {
  Name: "nav",
  data() {
    return {
      hasUpdate: false,
      version: "dev",
      latestVersion: "",
    };
  },
};
</script>
