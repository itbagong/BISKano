<template>
  <apply-list />
</template>
<script setup>
 
import { reactive, ref, inject, onMounted, watch } from "vue";
import { layoutStore } from "@/stores/layout.js"; 
import { authStore } from "@/stores/auth.js";

import ApplyList from "./widget/Apply/List.vue"

layoutStore().name = "tenant";

const FEATUREID = 'Administrator'
const profile = authStore().getRBAC(FEATUREID)


const listControl = ref(null);
const axios = inject("axios");

</script>