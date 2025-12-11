<template>
  <Menu as="div" class="relative inline-block text-left">
    <MenuButton
      class="hover:text-primary cursor-pointer rounded-full w-[30px] h-[30px] bg-slate-300 flex items-center justify-center"
    >
      <mdicon name="account"></mdicon>
    </MenuButton>

    <transition
      enter-active-class="transition ease-out duration-100"
      enter-from-class="transform opacity-0 scale-95"
      enter-to-class="transform opacity-100 scale-100"
      leave-active-class="transition ease-in duration-75"
      leave-from-class="transform opacity-100 scale-100"
      leave-to-class="transform opacity-0 scale-95"
    >
      <MenuItems
        class="origin-top-right absolute right-0 mt-2 w-56 bg-white shadow-lg ring-1 ring-slate-200 focus:outline-none divide-y divide-slate-200"
      >
        <div class="py-1" v-for="menuDiv in menus">
          <MenuItem v-slot="{ active }" v-for="menu in menuDiv">
            <a
              @click="goto(menu.url)"
              :class="[
                active ? 'bg-slate-200' : ' text-gray-600',
                'text-gray-600 cursor-pointer block px-4 py-2 text-sm',
              ]"
              >{{ menu.label }}</a
            >
          </MenuItem>
        </div>
      </MenuItems>
    </transition>
  </Menu>
</template>

<script setup>
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { inject, ref } from "vue";
import { useRouter } from "vue-router";
//import util from '@/assets/js/util';
import { authStore } from "@/stores/auth";

const router = useRouter();
const axios = inject("axios");
const auth = authStore();

const props = defineProps({
  auth: { type: Object, default: () => {} },
});

const homeUrl = import.meta.env.VITE_HOME_URL ?? "";

const menus = ref([
  [
    { label: "Profile", url: homeUrl + "/iam/accountprofile" },
    { label: "Change Tenant", url: homeUrl + "/iam/changetenant" },
  ],
  [
    { label: "Impersonate", url: homeUrl + "/iam/impersonate" },
    { label: "De-Impersonate", url: homeUrl + "/iam/deimpersonate" },
  ],
  [
    {
      label: "Logout",
      url: () => {
        axios.post(homeUrl + "/iam/logout").then(
          (r) => {
            auth.clear();
            router.push("/");
          },
          (err) => alert(err)
        );
      },
    },
  ],
]);

function goto(destination) {
  if (typeof destination == "string") {
    if (destination.startsWith("http")) {
      window.location.href = destination;
    } else {
      router.push(destination);
    }
  } else if (typeof destination == "function") {
    destination();
  }
}
</script>
