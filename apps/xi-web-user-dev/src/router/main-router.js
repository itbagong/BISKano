import { createRouter, createWebHistory } from 'vue-router';
import { authStore } from "@/stores/auth";
import { layoutStore } from "@/stores/layout";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
      meta: { guest: true, }
    },

    //-- iam
    { path: "/register", name: "register", component: () => import("@/views/iam/Register.vue"), meta: { guest: true, } },
    { path: "/login", name: "login", component: () => import("@/views/iam/Login.vue"), meta: { guest: true, } },
    { path: "/request-reset-password", name: "request-reset-password", component: () => import("@/views/iam/RequestToResetPassword.vue") },
    { path: "/public-reset-password", name: "public-reset-password", component: () => import("@/views/iam/PublicResetPassword.vue"), meta: { guest: true, } },
    { path: "/reset-password", name: "reset_password", component: () => import("@/views/iam/ResetPassword.vue"), },
    { path: "/account-profile", name: "account_profile", component: () => import("@/views/iam/AccountProfile.vue") },
    { path: "/activate-user", name: "public-activate", component: () => import("@/views/iam/ActivateUser.vue"),meta: { guest: true, } },

    //-- iam me
    { path: "/me/activate", name: "me-activate", component: () => import("@/views/iam/ActivateUser.vue") },
    { path: "/me/create-tenant", name: "me-create-tenant", component: () => import("@/views/iam/CreateTenant.vue") },
    { path: "/me/change-tenant", name: "me-change-tenant", component: () => import("@/views/iam/ChangeTenant.vue") },
    { path: "/me/join-tenant", name: "me-join-tenant", component: () => import("@/views/iam/JoinTenant.vue") },
    { path: "/me/reset-password", name: "reset_password", component: () => import("@/views/iam/ResetPassword.vue"),meta: { guest: true, } },
    { path: "/me/impersonate", name: "me_impersonate", component: () => import("@/views/iam/Impersonate.vue") },
    { path: "/me/deimpersonate", name: "me_deimpersonate", component: () => import("@/views/iam/Deimpersonate.vue") },

    //-- tenant
    { path: "/tenant/:fid", name: "tenant_fid", component: () => import("@/views/tenant/Index.vue") },


    //-- lab
    /*
    { path: "/lab/index", name: "lab_index", component: () => import("@/views/lab/Index.vue") },
    { path: "/lab/item", name: "lab_item", component: () => import("@/views/lab/Item.vue") },
    { path: "/lab/sales", name: "lab_sales", component: () => import("@/views/lab/Sales.vue") },
    { path: "/lab/form3", name: "lab_form3", component: () => import("@/views/lab/NewForm.vue") },
    { path: "/lab/editor", name: "lab_editor", component: () => import("@/views/lab/Editor.vue") },
*/

    //-- admin
    { path: "/admin/user", name: "admin_user", component: () => import("@/views/admin/User.vue") },
    { path: "/admin/role", name: "admin_role", component: () => import("@/views/admin/Role.vue") },
    { path: "/admin/featurecategory", name: "admin_featurecategory", component: () => import("@/views/admin/FeatureCategory.vue") },
    { path: "/admin/feature", name: "admin_feature", component: () => import("@/views/admin/Feature.vue") },
    { path: "/admin/tenant", name: "admin_tenant", component: () => import("@/views/admin/Tenant.vue") },
    { path: "/admin/tenantjoin", name: "admin_tenant_join", component: () => import("@/views/admin/TenantJoinReview.vue") },
    { path: "/admin/grant", name: "admin_grant", component: () => import("@/views/admin/Grant.vue") },
    { path: "/admin/app", name: "admin_app", component: () => import("@/views/admin/App.vue") },
    
    //-- msg
    { path: '/admin/msgtpl', name: 'admin_msgtpl', component: () => import('@/views/msg/MsgTpl.vue') },


    //-- general
    { path: "/general/noaccess", name: "general_noaccess", component: () => import("@/views/share/NoAccess.vue") },
    { path: "/general/tableview", name: "general_tableview", component: () => import("@/views/share/TableView.vue") },
    { path: "/msg/tpl", name: "msg_tpl", component: () => import("@/views/msg/MsgTpl.vue") },
  ]
})


router.beforeEach(async (to, from, next) => {
  const auth = authStore();
  if (to.matched.some((record) => record.meta.guest)) {
    next();
  } else {
    if (auth.appToken == "") {
      next({
        path: layoutStore().loginPage,
      });
    } else {
      next();
    }
  }
});

export default router

