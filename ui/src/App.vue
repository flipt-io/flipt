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
import Nav from "./components/Nav";
import Footer from "./components/Footer";
import UpdateNotification from "./components/UpdateNotification";

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
