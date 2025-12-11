<template>
  <div
    v-if="submenu.length > 0"
    :class="[viewType == 'full' ? 'text-left' : 'text-center']"
  >
    <Menu as="div" class="relative inline-block">
      <MenuButton
        class="flex gap-2 items-cente w-[180px] justify-between"
        v-if="viewType == 'full'"
        :class="[activeMenus[0] == label ? 'text-primary' : 'text-gray-600']"
        @click="OnMenuItem(label)"
      >
        <div class="flex gap-2 w-full">
          <mdicon
            :name="icon"
            size="20"
            class="w-[20px]"
            v-if="icon != ''"
          ></mdicon>
          <div v-else class="w-[20px]">&nbsp;</div>
          <div class="grow text-left font-semibold">{{ label }}</div>
        </div>
        <mdicon name="chevron-right" size="20" class="pr-2"></mdicon>
      </MenuButton>

      <MenuButton
        @click="OnMenuItem(label)"
        class="flex gap-2 items-center w-full"
        v-else
      >
        <mdicon
          :name="icon"
          size="20"
          class="w-full"
          :class="[
            activeMenus.includes(label) ? 'text-primary' : 'text-gray-600',
          ]"
          v-if="icon != ''"
        ></mdicon>
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
          :class="[
            viewType == 'full' ? 'left-[190px]' : 'left-[40px]',
            `menu-items-${label.toLowerCase().replaceAll(/ /g, '-')}`,
          ]"
          class="absolute top-[-6px] w-56 bg-white shadow-lg ring-1 ring-slate-200 focus:outline-none "
        >
          <div class="relative divide-y divide-slate-200">
            <div
              class="py-1 text-center text-[1.4em] font-semibold sticky top-0 z-10 bg-white text-slate-400"
            >
              {{ label }}
            </div>
            <div class="group-menu  divide-y divide-slate-200">
              <div class="py-1 "  v-for="menus in props.submenu">
                <MenuItem v-slot="{ active }" v-for="menu in menus">
                  <div
                    v-if="menu.submenu && menu.submenu.length > 0"
                    :class="[viewType == 'full' ? 'text-left' : 'text-center']"
                    @click="OnMenuItem(menu.label)"
                  >
                    <context-sub-menu
                      as="div"
                      :submenu="menu"
                      :active-menus="activeMenus"
                      :lebelmenu="activeMenus[1]"
                    ></context-sub-menu>
                  </div>
                  <div v-else>
                    <a
                      @click="goto(menu.url)"
                      :class="[
                        active ? 'bg-slate-200' : ' text-gray-600',
                        activeMenus[1] == menu.label
                          ? 'text-primary'
                          : 'text-gray-600',
                        'text-gray-600 cursor-pointer block px-4 py-2 text-left',
                      ]"
                      >{{ menu.label }}</a
                    >
                  </div>
                </MenuItem>
              </div>
            </div>
          </div>
        </MenuItems>
      </transition>
    </Menu>
  </div>

  <div
    v-else-if="viewType == 'full'"
    class="flex gap-2 items-center w-full"
    :class="[activeMenus[0] == label ? 'text-primary' : 'text-gray-600']"
    @click="goto(url)"
  >
    <mdicon :name="icon" size="16" class="w-[20px]" v-if="icon != ''"></mdicon>
    <div v-else class="w-[20px]">&nbsp;</div>
    <div class="grow text-left font-semibold">{{ label }}</div>
  </div>

  <div v-else @click="goto(url)">
    <div class="flex justify-center">
      <mdicon
        :name="icon"
        size="20"
        :class="[activeMenus[0] == label ? 'text-primary' : 'text-gray-600']"
        class="text-center"
        v-if="icon != ''"
      ></mdicon>
      <div v-else class="w-full">{{ label[0] }}</div>
    </div>
  </div>
</template>

<script setup>
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { useRouter } from "vue-router";
import ContextSubMenu from "./ContextSubMenu.vue";
import { util } from "suimjs";

const props = defineProps({
  label: { type: String, default: "" },
  icon: { type: String, default: "" },
  url: {
    type: [String, Object],
    default: () => {
      return "";
    },
  },
  submenu: { type: Array, default: () => [] },
  activeMenus: { type: Array, default: () => [] },
  viewType: { type: String, default: "full" },
});

const router = useRouter();

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

function OnMenuItem(label) {
  setTimeout(function () {
    const menu = `.menu-items-${label.toLowerCase().replaceAll(/ /g, "-")}`;
    util.nextTickN(2, () => {
      const el = document.querySelector(menu);
      if (el) {
        // const button = el.getBoundingClientRect().bottom;
        // const layout = document.querySelector(".main-layout").clientHeight;
        // if (window.innerHeight < button) {
        //   el.style.top = `-${button - window.innerHeight + 10}px`;
        // }
        const menuHeight = document.querySelector(menu).clientHeight;
        const layout = document.querySelector(".nav_left").clientHeight;

        if (menuHeight > layout - 60) {
          el.style.maxHeight = `${layout - 60}px`;
          el.style.overflowX = `auto`;
        }

        const button = document
          .querySelector(menu)
          .getBoundingClientRect().bottom;
        if (window.innerHeight < button) {
          const el = document.querySelector(menu);
          el.style.top = `-${button - window.innerHeight + 10}px`;
        }
      }
    });
  }, 300);
}
</script>
