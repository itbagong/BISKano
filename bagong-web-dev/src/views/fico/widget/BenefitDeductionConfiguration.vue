<template>
  <div>
    <s-form :config="data.config" v-model="data.value" keep-label>
      <template #buttons>&nbsp;</template>
    </s-form>
  </div>
</template>

<script setup>
import { reactive, onMounted, inject } from "vue";
import { SForm, loadFormConfig, util } from "suimjs";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  page: { type: String, default: () => "payrollbenefit" },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  config: {},
  value: props.modelValue,
});

const axios = inject("axios");

onMounted(() => {
  loadFormConfig(axios, `/bagong/${props.page}/detail/formconfig`).then(
    (r) => {
      data.config = r;
    },
    (e) => util.showError(e)
  );
});
</script>
