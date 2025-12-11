<template>
  <div>
    <div class="title section_title">Configuration</div>
    <s-form ref="formConfig"
      v-if="!data.loading"
      v-model="data.records"
      :config="data.formCfg"
      keep-label
      only-icon-top
      hide-submit
      hide-cancel
      @field-change="onFieldChange"
    >
    </s-form>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch, inject } from "vue";
import { DataList, loadFormConfig, SForm, util } from "suimjs";

const props = defineProps({
  title: { type: String, default: "" },
  modelValue: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const axios = inject("axios");
const formConfig = ref(null);
const data = reactive({
  records: props.modelValue,
  formCfg: {},
});

function onFieldChange(name, v1, v2, oldValue){

}

function updateItems() {
 
}


onMounted(() => {
  loadFormConfig(axios, "/bagong/siteconfiguration/formconfig").then(
    (r) => {
      data.formCfg = r;
      util.nextTickN(2, () => {
        updateItems();
      });
    },
    (e) => util.showError(e)
  );
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
