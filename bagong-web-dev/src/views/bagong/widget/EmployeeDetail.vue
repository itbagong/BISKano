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
  // console.log(data)
  // console.log(props)
  loadFormConfig(axios, "/bagong/employeedetail/formconfig").then(
    (r) => {
      data.config = r;
    },
    (e) => util.showError(e)
  );
});
</script>
