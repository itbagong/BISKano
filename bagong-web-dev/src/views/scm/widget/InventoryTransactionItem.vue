<template>
  <div>
    <s-form
      :config="data.config"
      v-model="data.value"
      keep-label
      hide-submit
      hide-cancel
    >
      <template #input_PhysicalDimension="{ item }">
        <dimension-item
          v-model="item.PhysicalDimension"
          :readOnly="true"
        ></dimension-item>
      </template>
      <template #input_FinanceDimension="{ item }">
        <dimension-item
          v-model="item.FinanceDimension"
          :readOnly="true"
        ></dimension-item>
      </template>
    </s-form>
  </div>
</template>

<script setup>
import { reactive, onMounted, inject, watch } from "vue";
import { SForm, loadFormConfig, util } from "suimjs";
import DimensionItem from "../../tenant/widget/DimensionItem.vue";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  id: { type: String, default: "" },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  value:
    props.modelValue == null || props.modelValue == undefined
      ? {}
      : props.modelValue,
  config: {},
});

const axios = inject("axios");

function getDataValue() {
  return data.value;
}
watch(
  () => data.value,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

onMounted(() => {
  loadFormConfig(axios, "/tenant/item/formconfig").then(
    (r) => {
      data.config = r;
    },
    (e) => util.showError(e)
  );
});

defineExpose({
  getDataValue,
});
</script>
