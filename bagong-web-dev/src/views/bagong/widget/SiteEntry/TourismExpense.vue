<template>
  <div> 
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
      :grid-hide-control="props.noAction"
      :grid-hide-delete="props.noAction"
      grid-no-confirm-delete
      init-app-mode="grid"
      grid-mode="grid"
      grid-config="/bagong/site_expense/gridconfig"
      new-record-type="grid"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @grid-row-save="onGridRowSave"
      @alter-grid-config="onAlterGridConfig"
    >
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch } from "vue";
import { DataList, util } from "suimjs";

const props = defineProps({
  siteEntryAssetID: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
  isOther: { type: Boolean, default: false },
  noAction: { type: Boolean, default: false },
  hideFields: { type: Array, default: () => ["ID"] },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
});

const listControl = ref(null);

function newRecord(r) {
  r.ID = "";

  data.records.push(r);
  updateItems();
}

function onGridRowDelete(_, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  updateItems();
}

function onGridRowSave(record, index) {
  record.suimRecordChange = false;
  data.records[index] = record;
  updateItems();
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
  emit("update:modelValue", data.records);
}

function onAlterGridConfig(config) {
  for (let i in config.fields) {
    let o = config.fields[i];
    if (props.hideFields.includes(o.field)) o.readType = "hide";

    if (o.field == "Name" && props.isOther) {
      o.input.useList = true;
      o.input.useLookup = true;
      o.input.lookupKey = "_id";
      o.input.lookupLabels = ["Name"];
      o.input.lookupSearchs = ["Name"];
      o.input.lookupUrl = "/bagong/expense/find";
    }
  }
  setTimeout(() => {
    updateItems();
  }, 500);
}

onMounted(() => {
 
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
