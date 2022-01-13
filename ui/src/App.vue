<template>
  <div>
    <div id="wrapper">
      <Nav />
      <RouterView />
    </div>
    <Footer />
  </div>
</template>

<script>
import Footer from "./components/Footer.vue";
import Nav from "./components/Nav";

export default {
  name: "App",
  components: {
    Nav,
    Footer,
  },
};
</script>

<style></style>
