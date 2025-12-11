<template>
  <div class="w-full">
    <label class="input_label" v-if="title != ''">{{ title }}</label>
    <data-list
      ref="listControl"
      hide-title
      no-gap
      grid-editor
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-hide-detail
      grid-hide-select
      grid-no-confirm-delete
      :init-app-mode="data.appMode"
      :grid-mode="data.appMode"
      new-record-type="grid"
      :grid-fields="['DriverID', 'ReplacementAssetID']"
      :grid-config="gridConfig"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @gridRowFieldChanged="onHandleGridFieldChanged"
    >
      <template #grid_DriverID="{ item, config }">
        <s-input
          hide-label
          v-model="item.DriverID"
          use-list
          :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${props.siteId}&Position=Driver`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #grid_ReplacementAssetID="{ item }">
        <s-input
          class="min-w-[100px]"
          hide-label
          v-model="item.ReplacementAssetID"
          use-list
          lookup-key="_id"
          :lookup-url="`/tenant/asset/find?GroupID=UNT&Dimension.Key=Site&Dimension.Value=${props.siteId}`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, inject, watch } from "vue";
import { DataList, SInput, loadGridConfig, util } from "suimjs";

const props = defineProps({
  title: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
  gridConfig: { type: String, default: "" },
  siteId: { type: String, default: "" },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
});

function onGridRowDelete(_, index) {
  const newRecords = data.records.filter((_, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
  emit("update:modelValue", data.records);
}

function newRecord(r) {
  r.Name = "";
  r.DriverID = "";
  r.Status = "";
  r.ReplacementAssetID = "";
  r.KMStart = 0;
  r.KMEnd = 0;
  r.KMTotal = 0;

  data.records.push(r);
  updateItems();
}

function onHandleGridFieldChanged(name, v1, v2, old, record) {
  if (name == "Status") {
    record.ReplacementAssetID = "";
  }

  if (["KMStart", "KMEnd"].includes(name)) {
    util.nextTickN(2, () => {
      record.KMTotal = record.KMEnd - record.KMStart;
    });
  }
}

onMounted(() => {
  setTimeout(() => {
    updateItems();
  }, 800);
});
</script>
