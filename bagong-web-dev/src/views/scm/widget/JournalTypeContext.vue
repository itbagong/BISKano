<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full grid-line-items"
      :editor="true"
      hide-search
      hide-select
      hide-sort
      :hide-new-button="false"
      :hide-delete-button="false"
      hide-refresh-button
      :hide-detail="true"
      :hide-action="false"
      @new-data="newRecord"
      @delete-data="onDelete"
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
    >
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, reactive, ref } from "vue";
import { loadGridConfig, util, SGrid } from "suimjs";
const axios = inject("axios");
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  item: { type: Object, default: () => {} },
  itemID: { type: String, default: () => "" },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  formMode: { type: String, default: () => "new" },
  readOnly: { type: Boolean, defaule: false },
  hideDetail: { type: Boolean, defaule: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);
const data = reactive({
  value:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  gridCfg: {},
});

function newRecord() {
  const record = {};
  record.ID = "";
  record.Label = "";
  record.Address = "";
  listControl.value.setRecords([...listControl.value.getRecords(), record]);
}

function onDelete(record, index) {
  const newRecords = listControl.value.getRecords().filter((dt, idx) => {
    return idx != index;
  });
  listControl.value.setRecords(newRecords);
}

function getDataValue() {
  return listControl.value.getRecords();
}

onMounted(() => {
  loadGridConfig(axios, "/fico/journaltypecontext/gridconfig").then(
    (r) => {
      data.gridCfg = r;
    },
    (e) => util.showError(e)
  );
  listControl.value.setRecords(data.value);
});
defineExpose({
  getDataValue,
});
</script>
