<template>
  <div>
    <loader v-if="data.loading" kind="skeleton" skeleton-kind="circle" />
    <div
      class="flex flex-col gap-5 mt-5"
      v-else
      v-for="atc in data.record"
    >
      <button @click="onPreviewImg(helper.getAssetUrl(atc._id))">
        <img :src="helper.getAssetUrl(atc._id)" class="" />
      </button>
    </div>
  </div>
</template>
<script setup>
import { reactive, inject, watch, onMounted } from "vue";
import { util } from "suimjs";
import Loader from "@/components/common/Loader.vue";
import helper from "@/scripts/helper.js";
import { api as viewerApi } from "v-viewer";

const axios = inject("axios");

const props = defineProps({
  modelValue: { type: String, default: () => "" },
  journalType: { type: String, default: () => "" },
});
const data = reactive({
  loading: false,
  id: props.modelValue,
  record: [],
});
const getAttachment = (id) => {
  data.loading = true;
  axios
    .post("/asset/read-by-journal", {
      JournalType: props.journalType,
      JournalID: id,
    })
    .then(
      (r) => {
        data.record = r.data;
        data.loading = false;
      },
      (e) => {
        util.showError(e);
        data.loading = false;
      }
    )
    .catch((e) => {
      util.showError(e);
      data.previewLoading = false;
    });
};

function onPreviewImg(uri) {
  viewerApi({
    images: [uri],
  });
}
watch(
  () => data.id,
  (nv) => {
    getAttachment(nv)
  },
  {deep: true}
);
onMounted(() => {
  getAttachment(props.modelValue)
})
</script>
