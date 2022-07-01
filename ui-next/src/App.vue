<template>
  <div id="wrapper">
    <div id="main">
      <Nav />
      <RouterView />
    </div>
    <UpdateNotification />
    <Footer />
  </div>
</template>

<script>
import Nav from "./components/Nav.vue";
import Footer from "./components/Footer.vue";
import UpdateNotification from "./components/UpdateNotification.vue";

export default {
  name: "App",
  components: {
    Nav,
    Footer,
    UpdateNotification,
  },
};
</script>

<style scoped>
#wrapper {
  display: flex;
  min-height: 100vh;
  flex-direction: column;
}

#main {
  flex: 1;
}

footer {
  padding: 1rem;
}
</style>
