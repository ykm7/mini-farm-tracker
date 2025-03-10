import AssetCollection from "@/components/AssetCollection.vue"
import SensorCollection from "@/components/SensorCollection.vue"
import Dashboard from "@/views/DasboardView.vue"
import { createRouter, createWebHistory } from "vue-router"

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: Dashboard,
      children: [
        {
          path: "",
          name: "Asset Collection",
          component: AssetCollection,
        },
        {
          path: "/sensor",
          name: "Sensor Collection",
          component: SensorCollection,
        },
      ],
    },
    {
      path: "/about",
      name: "about",
      // route level code-splitting
      // this generates a separate chunk (About.[hash].js) for this route
      // which is lazy-loaded when the route is visited.
      component: () => import("../views/AboutView.vue"),
    },
  ],
})

export default router
