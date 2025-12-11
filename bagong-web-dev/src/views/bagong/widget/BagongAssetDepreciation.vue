<template>
  <s-form
    v-if="!data.loading"
    v-model="data.record"
    :config="data.config"
    keep-label
    only-icon-top
    hide-submit
    hide-cancel
    class="form_detail_asset_bagong"
  >
  </s-form>
</template>

<script setup>
import { reactive, onMounted, inject, watch } from "vue";
import { SForm, loadFormConfig, util } from "suimjs";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  config: {},
  record: props.modelValue,
  loading: false,
});

const axios = inject("axios");

function getConfig() {
  let urlFormCfg = "/bagong/asset/detail/depreciation/formconfig";
  data.loading = true;
  loadFormConfig(axios, urlFormCfg).then(
    (r) => {
      data.config = r;
      data.loading = false;
    },
    (e) => util.showError(e)
  );
}

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

onMounted(() => {
  getConfig();
});
</script>
