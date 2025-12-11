<template>
  <div class="flex flex-col gap-2">
    <data-list
      ref="listControl"
      hide-title
      no-gap
      :grid-hide-detail="forceShowGridDetail ? false : (gridHideDetail || readOnly)"
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
      :grid-config="props.gridConfigUrl"
      :form-config="props.formConfigUrl"
      :form-fields="[
        'EmployeeID',
        'EmployeeNIK',
        'EmployeeName',
        'EmployeeLevel',
        'EmployeeSite',
        'EmployeePosition',
        'Position',
        'EmployeeDepartment',
        'Details'
      ]"
      :grid-fields="[
        'EmployeeNIK',
        'EmployeeName',
        'EmployeeLevel',
        'EmployeeSite',
        'EmployeePosition',
        'Position',
        'EmployeeDepartment',
        'ActualOvertime',
      ]"
      :init-form-mode="data.formMode"
      @grid-row-add="onGridNewRecord"
      @form-edit-data="openForm"
      @alter-grid-config="onAlterGridConfig"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="onGridRowFieldChanged"
      :form-hide-cancel="formHideCancel"
      :formHideSubmit="readOnly"
      :form-tabs-edit="data.formTabs"
      :form-tabs-new="data.formTabs"
      :form-tabs-view="data.formTabs"
      form-focus
      grid-hide-paging
    >
      <!-- slot for Overtime lines -->
      <template #grid_EmployeeNIK="{ item, header }">
        <slot name="grid_EmployeeNIK" :item="item" :header="header"></slot>
      </template>
      <template #grid_EmployeeName="{ item, header }">
        <slot name="grid_EmployeeName" :item="item" :header="header"></slot>
      </template>
      <template #grid_EmployeeLevel="{ item }">
        <slot name="grid_EmployeeLevel" :item="item"></slot>
      </template>
      <template #grid_EmployeeSite="{ item }">
        <slot name="grid_EmployeeSite" :item="item"></slot>
      </template>
      <template #grid_EmployeePosition="{ item, header }">
        <slot name="grid_EmployeePosition" :item="item" :header="header"></slot>
      </template>
      <template #grid_Position="{ item, header }">
        <slot name="grid_Position" :item="item" :header="header"></slot>
      </template>
      <template #grid_EmployeeDepartment="{ item, header }">
        <slot
          name="grid_EmployeeDepartment"
          :item="item"
          :header="header"
        ></slot>
      </template>
      <template #grid_ActualOvertime="{ item, mode }">
        <slot name="grid_ActualOvertime" :item="item" :mode="mode"></slot>
      </template>
      <template #form_input_EmployeeID="{ item }">
        <slot name="form_EmployeeID" :item="item"></slot>
      </template>
      <template #form_input_EmployeeNIK="{ item }">
        <slot name="grid_EmployeeNIK" :item="item"></slot>
      </template>
      <template #form_input_EmployeeName="{ item }">
        <slot name="grid_EmployeeName" :item="item"></slot>
      </template>
      <template #form_input_EmployeeLevel="{ item }">
        <slot name="grid_EmployeeLevel" :item="item"></slot>
      </template>
      <template #form_input_EmployeeSite="{ item }">
        <slot name="grid_EmployeeSite" :item="item"></slot>
      </template>
      <template #form_input_EmployeePosition="{ item }">
        <slot name="grid_EmployeePosition" :item="item"></slot>
      </template>
      <template #form_input_Position="{ item }">
        <slot name="grid_Position" :item="item"></slot>
      </template>
      <template #form_input_EmployeeDepartment="{ item }">
        <slot name="grid_EmployeeDepartment" :item="item"></slot>
      </template>
      <template #form_input_Details="{ item, mode }">
        <slot name="form_Details" :item="item" :mode="mode"></slot>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, watch } from "vue";
import { DataList, util } from "suimjs";

import helper from "@/scripts/helper.js";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false },
  gridConfigUrl: {
    type: String,
    default: "/fico/customerjournal/line/gridconfig",
  },
  formConfigUrl: {
    type: String,
    default: "/fico/customerjournal/line/formconfig",
  },
  gridHideDetail: { type: Boolean, default: false },
  formHideCancel: { type: Boolean, default: true },
  forceShowGridDetail: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  newRecord: null,
  gridRowFieldChanged: null,
  alterGridConfig: null,
  postDelete: null,
});

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  formTabs: buildFormTabs(),
  originRecord: {},
});

function buildFormTabs() {
  let r = ["General"];
  return r;
}

function openForm(r) {
  data.originRecord = helper.cloneObject(r);
  util.nextTickN(2, () => {
    if (props.readOnly === true) {
      data.formMode = "view";
      listControl.value.setFormMode("view");
    }
  });
}

function onGridNewRecord(r) {
  emit("newRecord", r);
  data.records.push(r);
  updateItems();
}

function onGridRowDelete(_, index) {
  const rowDelete = () => {
    const deletedRecord = data.records[index];
    const newRecords = data.records.filter((_, idx) => {
      return idx != index;
    });
    data.records = newRecords;
    util.nextTickN(2, () => {
      updateItems();
      emit("postDelete", deletedRecord);
    });
  };
  rowDelete();
}

function onAlterGridConfig(config) {
  emit("alterGridConfig", config);
  setTimeout(() => {
    updateItems();
  }, 500);
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
  gridFieldChange(name, record);
  emit("gridRowFieldChanged", name, v1, v2, old, record, () => {
    gridFieldChange(name, record);
  });
}

function gridFieldChange(name, record) {
  listControl.value.setGridRecord(
    record,
    listControl.value.getGridCurrentIndex()
  );
  updateItems();
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
}

function refresh() {
  data.records = props.modelValue;
  util.nextTickN(2, () => {
    updateItems();
  });
}

defineExpose({
  refresh,
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
