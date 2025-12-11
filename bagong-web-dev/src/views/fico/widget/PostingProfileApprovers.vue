<template>
  <div class="flex flex-col gap-2">
    <h3>{{ title }}</h3>
    <div class="controls">
      <div class="flex"> 
      <SButton
      v-if="!readOnly"
        class="btn_primary"
        icon="account-multiple-plus"
        :label="'Add ' + title"
        @click="addRecord"
      />
      </div>
    </div>
    <data-list
      ref="listControl"
      title="Posting Profile"
      hide-title
      no-gap
			:grid-hide-detail="readOnly" 
			:grid-hide-delete="readOnly"
			grid-hide-control
      :grid-editor="!readOnly" 
      grid-hide-select
      grid-no-confirm-delete
      init-app-mode="grid"
      grid-mode="grid"
      grid-config="/fico/postingprofile/approver/gridconfig"
      form-config="/fico/postingprofile/approver/formconfig"
      @formNewData="newRecord"
      @grid-row-delete="onGridRowDelete"
      @grid-row-save="onGridRowSave"
      @alter-grid-config="alterGridConfig" 
    >
    
    </data-list>
  </div>
</template>

<script setup>
import { watch, onMounted } from "vue";
import { reactive, ref } from "vue";
import { DataList, SButton, util,SInput } from "suimjs";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  title: { type: String, default: "" },
    readOnly: {type: Boolean, default: false},
});

const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  records: props.modelValue.map((dt) => {
    dt.suimRecordChange = false;
    return dt;
  }),
});

function newRecord(record) {
  record.suimRecordChange = true;
  record.MinimalApproverCount = 1;
  return record;
}

function addRecord() {
  const record = newRecord({});
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
}
function alterGridConfig() {
  setTimeout(() => {  
    listControl.value.setGridRecords(data.records);
  }, 500);
}
onMounted(() => {
  setTimeout(() => {
    listControl.value.setGridRecords(data.records);
  }, 500);
});
</script>