<template>
  <loader kind="skeleton" v-if="data.loading" />
  <div v-else>
    <div v-if="data.record._id">
    <img
      :src="helper.getAssetUrl(data.record._id)"
      :class="imgClass"
      alt="PasPhoto"
    />
  </div>
  <div v-else :class="noPhotoClass"></div>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, watch } from "vue";
import helper from "@/scripts/helper.js";
import Loader from "@/components/common/Loader.vue";

const axios = inject("axios");

const props = defineProps({
  modelValue: { type: String, default: "" },
  noPhotoClass: {
    type: String,
    default: "w-[50px] h-[40px] bg-cover bg-zinc-200",
  },
  imgClass: { type: String, default: "h-[40px] w-[50px] object-cover" },
});
const data = reactive({
  loading: false,
  record: {},
});
onMounted(() => {
  data.loading= true;
  axios
    .post("/asset/read-by-journal", {
      Tags: [`DOCUMENTS_CHECKLIST_${props.modelValue}_PasPhoto`],
    })
    .then((r) => {
      if (r.data.length > 0) {
        data.record = r.data[0];
      }
    })
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loading = false 
    });
});
</script>
<style></style>
