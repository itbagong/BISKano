<template>
  <div class="flex flex-col gap-2">
    <data-list
      ref="listControl"
      title="Posting Contact"
      hide-title
      no-gap
      :grid-hide-detail="readOnly"
			:grid-editor="!readOnly"
			:grid-hide-delete="readOnly"
			:grid-hide-control="readOnly" 
      grid-hide-select
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-no-confirm-delete 
      gridAutoCommitLine
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      grid-config="/bagong/vendor/contact/gridconfig"
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @grid-refreshed="gridRefreshed"
      @grid-row-save="onGridRowSave"
      @post-save="onFormPostSave"
      form-focus
    >
    </data-list>
  </div>
</template>

<script setup>
import { onMounted } from "vue";
import { reactive, ref } from "vue";
import { DataList, SButton, util } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  records: props.modelValue ?? [],
});

function newRecord() {
  const record = {};
  record.Name = "";
  record.FirstName = "";
  record.JobTitle = "";
  record.Telephone = "";
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

function updateItems() {
  const committedRecords = data.records.filter(
    (dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
  );
  emit("update:modelValue", committedRecords);
  emit("recalc");
}

function gridRefreshed() {
  listControl.value.setGridRecords(data.records);
}

onMounted(() => {
  setTimeout(() => {}, 500);
});
</script>
