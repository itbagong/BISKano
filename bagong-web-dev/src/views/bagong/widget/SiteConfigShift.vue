<template>
  <div>
    <div class="title section_title">Shift</div>
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
      grid-config="/bagong/shift/gridconfig"
      form-config="/bagong/shift/formconfig"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
    >
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch, inject } from "vue";
import { DataList, loadGridConfig, util } from "suimjs";

const props = defineProps({
  title: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
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

function newRecord(r) {
  r.ShiftID = "";
  r.StartTime = "";
  r.EndTime = "";

  data.records.push(r);
  updateItems();
}

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

onMounted(() => {
  loadGridConfig(axios, "/bagong/shift/gridconfig").then(
    (r) => {
      setTimeout(() => {
        updateItems();
      }, 500);
    },
    (e) => util.showError(e)
  );
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
