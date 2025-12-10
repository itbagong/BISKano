<template>
  <div :class="{notifClearLayout:layout.name=='clear', notifTenantLayout:layout.name=='tenant' || layout.name=='admin'}">
    <div v-for="msg in messages" class="w-full text-center items-center justify-center flex gap-1 p-2 opacity-70" 
      :class="{'btn_secondary':msg.kind=='info','btn_warning':msg.kind=='warning', 'btn_error':msg.kind=='error'}">
      <div class="grow">{{ msg.message }}</div>
      <div class="cursor-pointer hover:bg-primary hover:text-white" @click="notif.remove(msg.msgid)">
        <mdicon size="16" name="close" />
      </div>
    </div>
  </div>

  <ClearLayout v-if="layout.name == 'clear'" />
  <TenantLayout v-if="layout.name == 'tenant'" />
</template>

<script setup>
import ClearLayout from "@/layout/ClearLayout.vue";
import TenantLayout from "@/layout/TenantLayout.vue";

import { layoutStore } from "@/stores/layout.js";
import { notifStore } from "@/stores/notif.js";
import { storeToRefs } from "pinia";

const layout = layoutStore()
const notif = notifStore()
const { messages } = storeToRefs(notif)

</script>

<style>
#app {
  /* min-width: 100vw; */
  min-height: 100vh;
  display: flex;
}

.notifClearLayout {
  @apply absolute flex flex-col top-0 left-0 w-full z-50 gap-[1px]
}

.notifTenantLayout {
  @apply absolute flex flex-col top-[80px] right-[5px] w-[40%] z-50 gap-[1px]
}
</style>

<style>
@import "@/assets/css/base.css";
@import "@/assets/css/app.css";
</style>
