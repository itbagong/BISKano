<template>
  <div
    class="w-full flex pt-[70px] p-5 justify-center bg-main-bg main-layout"
    :class="[
      data.menuType == 'full'
        ? 'ml-[200px]'
        : data.menuType == 'mini'
        ? 'ml-[50px]'
        : '',
    ]"
  >
    <RouterView />
  </div>

  <div class="nav_top flex items-center">
    <div
      class="w-[200px] h-full flex items-center justify-center gap-2 px-2 bg-[#1841a6]"
      v-if="data.menuType == 'full'"
      @click="changeMenuType()"
    >
      <mdicon name="menu" class="cursor-pointer" width="18" />
      <h1 class="cursor-pointer text-[1em]">XiBar</h1>
    </div>
    <div
      v-else
      class="w-[50px] h-full flex items-center justify-center px-4 bg-[#1841a6] cursor-pointer"
      @click="changeMenuType()"
    >
      <h1 class="text-[1em]">Xi</h1>
    </div>
    <div class="grow h-full flex" v-if="auth.appToken != ''">
      <div class="flex flex-col pl-2">
        <div class="pt-2">{{  auth.appData.OriginalUserID == undefined ? auth.appData.DisplayName : `${auth.appData.DisplayName} | ${auth.appData.OriginalUserID}`  }}</div>
        <div class="text-[10px] text-slate-300">{{  auth.appData.TenantName  }}</div>
      </div>
    </div>
    <div v-else class="grow h-full flex flex-col px-2">
      &nbsp;
    </div>
    <div
      class="mr-4 flex gap-2 h-full items-center justify-center"
      v-if="auth.appToken == ''"
    >
      <mdicon
        size="18"
        name="login"
        class="nav_right_btn"
        @click="router.push('/login')"
      />
    </div>
    <div class="mr-4 flex gap-1 items-center justify-center" v-else>
      <user-context-menu :auth="auth" />
    </div>
  </div>

  <div
    v-if="data.menuType != 'hide'"
    class="nav_left flex-col gap-1 text-slate-100"
    :class="[data.menuType == 'full' ? 'w-[200px]' : 'w-[50px]']"
  >
    <div class="flex flex-col mt-4">
      <context-menu
        as="div"
        v-for="menu in appMenu"
        class="w-full pl-5 py-2 cursor-pointer hover:bg-nav-title-b"
        :icon="menu.icon"
        :label="menu.label"
        :url="menu.url"
        :submenu="menu.submenu"
        :view-type="data.menuType"
      />
    </div>
  </div>
</template>

<script setup>
import { useRouter } from "vue-router";
import { authStore } from "@/stores/auth";
import UserContextMenu from "./widgets/UserContextMenu.vue";
import { reactive, onMounted } from "vue";
import ContextMenu from "@/components/common/ContextMenu.vue";
import appMenu from "@/data/appmenu";

const router = useRouter();
const auth = authStore();

const data = reactive({
  menuType: "mini",
});

function changeMenuType() {
  switch (data.menuType) {
    case "full":
      data.menuType = "mini";
      break;

    case "mini":
      data.menuType = "hide";
      break;

    default:
      data.menuType = "full";
  }
}

function init() {
  if (auth.appToken == "") {
    //return router.push("/sign/in");
  }
}

onMounted(async () => {
  init();
});
</script>

<style>

.nav_top {
  @apply z-[999] fixed w-[100%] h-[50px] text-white bg-[#1b55e2];
  width: 100%;
}

.nav_left {
  @apply fixed h-full top-[50px] bg-[#1b4abf];
}

.nav_right_btn {
  @apply text-white hover:opacity-50 cursor-pointer;
}
</style>
