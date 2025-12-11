<template>
  <Menu as="div" class="relative px-4 py-2">
    <MenuButton
      class="flex gap-2 items-cente w-[180px] justify-between"
      :class="[
        props.lebelmenu == data.menu.label ? 'text-primary' : 'text-gray-600',
      ]"
      @click="OnMenuItem(data.menu.label)"
    >
      <div class="flex gap-2 w-full">
        <div class="grow text-left font-semibold">
          {{ data.menu.label }}
        </div>
      </div>
      <mdicon name="chevron-right" size="20" class="pr-2"></mdicon>
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
          'left-[230px]',
          `menu-items-${data.menu.label.toLowerCase().replaceAll(/ /g, '-')}`,
        ]"
        class="absolute top-[-6px] w-56 bg-white shadow-lg ring-1 ring-slate-200 focus:outline-none"
        @click="OnMenuItem(data.menu.label)"
      >
        <div class="relative divide-y divide-slate-200">
          <div
            class="py-1 text-center text-[1.4em] font-semibold sticky top-0 z-10 bg-white text-slate-400"
          >
            {{ data.menu.label }}

          </div>
          <div class="group-menu  divide-y divide-slate-200">
            <div class="py-1"  v-for="submenus in data.menu.submenu">
            
              <MenuItem v-slot="{ active }" v-for="menu in submenus">
                <div
                  v-if="menu.submenu && menu.submenu.length > 0"
                  :class="[viewType == 'full' ? 'text-left' : 'text-center']"
                >
                  <context-sub-menu
                    as="div"
                    :submenu="menu"
                    :active-menus="props.activeMenus"
                    :lebelmenu="activeMenus.at(-1)"
                  ></context-sub-menu>
                </div>
                <div v-else>
                  <a
                    @click="goto(menu.url)"
                    :class="[
                      active ? 'bg-slate-200' : ' text-gray-600',
                      activeMenus.at(-1) == menu.label
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
</template>
<script setup>
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { useRouter } from "vue-router";
import { reactive } from "vue";
import { util } from "suimjs";
import ContextSubMenu from "./ContextSubMenu.vue";

const props = defineProps({
  submenu: { type: Object, default: () => {} },
  lebelmenu: { type: String, default: "" },
  activeMenus: { type: Array, default: () => [] },
});

const router = useRouter();

const data = reactive({
  menu: props.submenu,
});

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
