import Vue from "vue";
import Router from "vue-router";

const Flags = () => import("@/components/Flags.vue");
const NewFlag = () => import("@/components/Flags/NewFlag.vue");
const Flag = () => import("@/components/Flags/Flag.vue");
const Targeting = () => import("@/components/Flags/Targeting.vue");
const Segments = () => import("@/components/Segments.vue");
const Segment = () => import("@/components/Segments/Segment.vue");
const NewSegment = () => import("@/components/Segments/NewSegment.vue");

Vue.use(Router);

export default new Router({
  routes: [
    {
      path: "/",
      redirect: "/flags",
    },
    {
      path: "/flags",
      name: "flags",
      component: Flags,
    },
    {
      path: "/flags/new",
      name: "new-flag",
      component: NewFlag,
    },
    {
      path: "/flags/:key",
      name: "flag",
      component: Flag,
    },
    {
      path: "/flags/:key/targeting",
      name: "targeting",
      component: Targeting,
    },
    {
      path: "/segments",
      name: "segments",
      component: Segments,
    },
    {
      path: "/segments/new",
      name: "new-segment",
      component: NewSegment,
    },
    {
      path: "/segments/:key",
      name: "segment",
      component: Segment,
    },
  ],
});
