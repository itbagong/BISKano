<template>
  <div class="flex flex-col gap-2">
    <data-list
      ref="listControl"
      title="Details"
      hide-title
      no-gap
      grid-editor
      grid-hide-select
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-no-confirm-delete
      grid-hide-detail
      gridAutoCommitLine
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      grid-config="/fico/fixedassetnumber/detail/gridconfig"
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @grid-refreshed="gridRefreshed"
      @grid-row-save="onGridRowSave"
      @post-save="onFormPostSave"
      @grid-row-deleted="onGridRowDeleted"
      @grid-row-field-changed="onGridRowFieldChanged"
      form-focus
    >
      <template #grid_paging>&nbsp;</template>
    </data-list>
  </div>
</template>

<script setup>
import { onMounted, watch } from "vue";
import { reactive, ref, inject } from "vue";
import { DataList, SButton, util } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
  calcTotalAsset: null,
});

const axios = inject("axios");
const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  records: props.modelValue.map((dt) => {
    return dt;
  }),
});

function newRecord() {
  const record = {};
  record.AssetName = "";
  record.FixedAssetGrup = "";
  record.NumberAsset = 0;
  record.InitialAssetNumber = 0;
  record.LastAssetNumber = 0;
  data.records.push(record);
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onGridRowDelete(record, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onFormPostSave(record, index) {
  record.suimRecordChange = false;
  if (listControl.value.getFormMode() == "new") {
    data.records.push(record);
  } else {
    data.records[index] = record;
  }
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onGridRowSave(record, index) {
  record.suimRecordChange = false;
  data.records[index] = record;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

// fixedassetgroup, prt, PRT | Property,
async function onGridRowFieldChanged(name, v1, v2, old, record) {
  if (name == "NumberAsset") {
    record.LastAssetNumber = record.InitialAssetNumber + v1 - 1;
    util.nextTickN(2, () => {
      calcTotalAsset();
    });
  }
  if (name == "FixedAssetGrup") {
    record.InitialAssetNumber = await getInitialNumber(v1);
    record.LastAssetNumber = record.InitialAssetNumber + record.NumberAsset;
  }
  listControl.value.setGridRecord(
    record,
    listControl.value.getGridCurrentIndex()
  );
  updateItems();
}

// fixedassetgroup, prt, PRT | Property,
async function onGridRowDeleted(record) {
  util.nextTickN(2, () => {
    calcTotalAsset();
  });
  updateItems();
}

async function getInitialNumber(id) {
  if (id == "") return;
  let payload = {
    AssetGroup: id,
  };
  let result = 0;
  try {
    const r = await axios.post(
      "/fico/fixedassetnumber/get-initial-number",
      payload
    );
    result = r.data.InitialAssetNumber;
  } catch (e) {
    util.showError(e);
  }
  return result;
}

function calcTotalAsset() {
  const totalAsset = data.records.reduce((total, e) => {
    return total + e.NumberAsset;
  }, 0);
  emit("calcTotalAsset", totalAsset);
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
}

function gridRefreshed() {
  listControl.value.setGridRecords(data.records);
}

onMounted(() => {
  setTimeout(() => {}, 500);
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
