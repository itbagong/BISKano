<template>
  <div v-if="submenu.length > 0">
    <Menu as="div" class="relative inline-block text-left">
      <MenuButton
        class="flex gap-2 items-center w-full"
        v-if="viewType == 'full'"
      >
        <mdicon
          :name="icon"
          size="16"
          class="w-[20px]"
          v-if="icon != ''"
        ></mdicon>
        <div v-else class="w-[20px]">&nbsp;</div>
        <div class="grow text-left">{{ label }}</div>
        <mdicon name="chevron-right" size="16" class="pr-2"></mdicon>
      </MenuButton>

      <MenuButton class="flex gap-2 items-center w-full" v-else>
        <mdicon
          :name="icon"
          size="16"
          class="w-full"
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
          :class="[viewType == 'full' ? 'left-[200px]' : 'left-[50px]']"
          class="absolute mt-2 top-0 w-56 bg-secondary shadow-lg ring-1 focus:outline-none divide-y divide-slate-200"
        >
          <div
            class="py-1 text-center text-slate-300 opacity-30 text-[1.4em] font-semibold"
          >
            {{ label }}
          </div>
          <div class="py-1" v-for="menuDiv in props.submenu">
            <MenuItem v-slot="{ active }" v-for="menu in menuDiv">
              <a
                @click="goto(menu.url)"
                :class="[
                  active ? 'bg-nav-title-bg text-slate-100' : 'text-white',
                  'cursor-pointer block px-4 py-2',
                ]"
                >{{ menu.label }}</a
              >
            </MenuItem>
          </div>
        </MenuItems>
      </transition>
    </Menu>
  </div>

  <div
    v-else-if="viewType == 'full'"
    class="flex gap-2 items-center w-full"
    @click="goto(url)"
  >
    <mdicon :name="icon" size="16" class="w-[20px]" v-if="icon != ''"></mdicon>
    <div v-else class="w-[20px]">&nbsp;</div>
    <div class="grow text-left">{{ label }}</div>
  </div>

  <div v-else class="flex gap-2 items-center w-full" @click="goto(url)">
    <mdicon :name="icon" size="16" class="w-full" v-if="icon != ''"></mdicon>
    <div v-else class="w-full">{{ label[0] }}</div>
  </div>
</template>

<script setup>
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { useRouter } from "vue-router";

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
  viewType: { type: String, default: "full" },
});

const router = useRouter();

function goto(destination) {
  if (typeof destination == "string") {
    router.push(destination);
  } else if (typeof destination == "function") {
    destination();
  }
}
</script>
