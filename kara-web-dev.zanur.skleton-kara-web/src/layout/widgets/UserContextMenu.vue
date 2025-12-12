<template>
    <Menu as="div" class="relative inline-block text-left">
        <MenuButton class="hover:text-primary cursor-pointer">
            <mdicon name="account"></mdicon>
        </MenuButton>

        <transition enter-active-class="transition ease-out duration-100"
            enter-from-class="transform opacity-0 scale-95" enter-to-class="transform opacity-100 scale-100"
            leave-active-class="transition ease-in duration-75" leave-from-class="transform opacity-100 scale-100"
            leave-to-class="transform opacity-0 scale-95">
            <MenuItems
                class="origin-top-right absolute right-0 mt-2 w-56 rounded-md bg-secondary shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none divide-y divide-slate-500">
                <div class="py-1" v-for="menuDiv in menus">
                    <MenuItem v-slot="{ active }" v-for="menu in menuDiv">
                    <a @click="goto(menu.url)" 
                        :class="[active ? 'bg-gray-100 text-gray-900' : 'text-white', 'cursor-pointer block px-4 py-2 text-sm']">{{ menu.label }}</a>
                    </MenuItem>
                </div>
            </MenuItems>
        </transition>
    </Menu>
</template>

<script setup>
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { inject, ref } from 'vue';
import { useRouter } from 'vue-router';
//import util from '@/assets/js/util';
import { authStore } from '@/stores/auth';

const router = useRouter()
const axios = inject("axios")
const auth = authStore()

const props = defineProps({
    auth: { type: Object, default: () => { } },
})

const menus = ref([
    [
        {label:"Profile", url:"/account-profile"},
        {label:"Change Tenant", url:"/me/change-tenant"},
        {label:"Logout", url:()=>{
            axios.post("/iam/logout").then(r=>{
                auth.clear()
                router.push("/")
            }, err => alert(err))        
        }}]
])

function goto(destination) {
    if (typeof(destination)=="string") {
        router.push(destination)
    } else if (typeof(destination)=="function") {
        destination()
    }
}

</script>