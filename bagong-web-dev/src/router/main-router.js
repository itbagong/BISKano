import { createRouter, createWebHistory } from 'vue-router';
import { authStore } from "@/stores/auth";
import { layoutStore } from "@/stores/layout";
import { routes } from 'vue-auto-route';
 
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes:[...routes, { path: "/:pathMatch(.*)*", name: "NotFound", component: () => import("@/views/404.vue") }],
})


router.beforeEach(async (to, from, next) => {
  const auth = authStore();
  const guestPath = ["/iam/login","/iam/publicresetpassword","/iam/register"];   
  if (guestPath.includes(to.path?.toLowerCase())) {
    next();
    return;
  } else {
    if (auth.appToken == '') {
      next({
        path: layoutStore().loginPage,
      });
      return;
    }

    auth.validate();
  }

  if (to.meta.rbac) {
    next({
      path: 'share/noacces'
    });
  }

  next();
});

export default router

