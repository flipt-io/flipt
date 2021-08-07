<template>
  <div>
    <Nav />
    <RouterView />
  </div>
</template>

<script>
import Nav from "./components/Nav";

export default {
  name: "App",
  components: {
    Nav,
  },
};
</script>

<style></style>
