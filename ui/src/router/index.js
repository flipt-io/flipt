import Vue from "vue";
import Router from "vue-router";
import Flags from "@/components/Flags";
import NewFlag from "@/components/Flags/NewFlag";
import Flag from "@/components/Flags/Flag";
import Targeting from "@/components/Flags/Targeting";
import Segments from "@/components/Segments";
import Segment from "@/components/Segments/Segment";
import NewSegment from "@/components/Segments/NewSegment";

Vue.use(Router);

export default new Router({
  routes: [
    {
      path: "/",
      redirect: "/flags"
    },
    {
      path: "/flags",
      name: "flags",
      component: Flags
    },
    {
      path: "/flags/new",
      name: "new-flag",
      component: NewFlag
    },
    {
      path: "/flags/:key",
      name: "flag",
      component: Flag
    },
    {
      path: "/flags/:key/targeting",
      name: "targeting",
      component: Targeting
    },
    {
      path: "/segments",
      name: "segments",
      component: Segments
    },
    {
      path: "/segments/new",
      name: "new-segment",
      component: NewSegment
    },
    {
      path: "/segments/:key",
      name: "segment",
      component: Segment
    }
  ]
});
