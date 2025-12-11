<template>
  <div class="w-full" style="margin-top: -4em">
    <s-form
      ref="inputOilUsage"
      v-model="data.record"
      :config="data.formCfg"
      keep-label
      :buttons-on-top="false"
    >
    </s-form>
  </div>
</template>
<script setup>
import { reactive, onMounted, inject, ref } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { SForm, loadFormConfig, util } from "suimjs";

layoutStore().name = "tenant";

const axios = inject("axios");

const props = defineProps({
  siteEntryAssetID: { type: String, default: "" },
  modelValue: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  formCfg: {},
  record:
    props.modelValue == null || props.modelValue == undefined
      ? {}
      : props.modelValue,
});

const inputOilUsage = ref(null);

function getFormRecord(id) {
  const url = "/bagong/siteentry_miningusage/find?SiteEntryAssetID=" + id;
  axios.post(url).then(
    (r) => {
      if (r.data.length > 0) {
        data.record = { ...r.data[0] };
      } else {
        data.record = {
          IsOilChange: false,
          OilUsage: 0,
          OilNotes: "",
        };
      }
      emit("update:modelValue", data.record);
    },
    (e) => util.showError(e)
  );
}

onMounted(() => {
  loadFormConfig(axios, "/bagong/siteentry_miningusage/formconfig").then(
    (r) => {
      data.formCfg = r;
      util.nextTickN(2, () => {
        const fieldRemoved = [
          "IsTireChange",
          "IsChangePosition",
          "TirePosition",
          "TireChangePlan",
          "TireType",
        ];
        fieldRemoved.forEach((field) => {
          inputOilUsage.value.removeField(field);
        });
      });
    },
    (e) => util.showError(e)
  );
  setTimeout(() => {
    getFormRecord(props.siteEntryAssetID);
  }, 500);
});
</script>
