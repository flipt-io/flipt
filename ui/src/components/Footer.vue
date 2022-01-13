<template>
  <footer class="footer">
    <div class="content has-text-centered">I am down here.</div>
  </footer>
</template>

<script>
export default {
  Name: "footer",
  data() {
    return {
      updateAvailable: false,
      version: "dev",
      latestVersion: "",
    };
  },
};
</script>
