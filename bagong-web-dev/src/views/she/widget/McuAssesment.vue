<template>
  <data-list
    ref="listControl"
    form-config="/she/mcutransaction/assesment/formconfig"
    form-read="/she/mcutransaction/assesment/get"
    form-keep-label
    init-app-mode="form"
    stay-on-form-after-save
    form-hide-buttons
    hide-title
  >
  </data-list>
</template>
<script setup>
import {
  reactive,
  ref,
  inject,
  watch,
  computed,
  onMounted,
  nextTick,
} from "vue";

import { SInput, util, SButton, SGrid, SCard, DataList } from "suimjs";
</script>
