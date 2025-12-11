<template>
  <Menu as="div" class="relative inline-block text-left">
    <MenuButton
      class="hover:text-primary cursor-pointer bg-slate-300 flex items-center justify-center"
    >
      <s-button class="btn_primary" :label="props.title" />
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
        class="origin-top-right absolute right-0 mt-2 w-56 bg-white shadow-lg ring-1 ring-slate-200 focus:outline-none divide-y divide-slate-500 z-10"
      >
        <div class="py-1" v-for="menuDiv in data.menus">
          <MenuItem v-slot="{ active }" v-for="menu in menuDiv">
            <a
              @click="postEmit(menu)"
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
import { reactive } from "vue";
import { SButton } from "suimjs";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  title: { type: String, default: () => "option " },
  menuDefault: { type: Boolean, default: () => true },
  menu: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const data = reactive({
  value: props.modelValue,
  menus: props.menuDefault
    ? [
        ...[
          [
            { label: "Purchase Request", emit: "PurchaseRequest" },
            { label: "Purchase Order", emit: "PurchaseOrder" },
          ],
        ],
        ...props.menu,
      ]
    : props.menu,
});

function postEmit(menu) {
  emit(menu.emit, menu);
}
defineExpose({});
</script>
