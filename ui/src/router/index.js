import Vue from "vue";
import Router from "vue-router";

const Flags = () => import("@/components/Flags");
const NewFlag = () => import("@/components/Flags/NewFlag");
const Flag = () => import("@/components/Flags/Flag");
const Targeting = () => import("@/components/Flags/Targeting");
const Segments = () => import("@/components/Segments");
const Segment = () => import("@/components/Segments/Segment");
const NewSegment = () => import("@/components/Segments/NewSegment");

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
