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
      grid-no-confirm-delete
      :init-app-mode="data.appMode"
      :grid-mode="data.appMode"
      new-record-type="grid"
      grid-config="/bagong/accident_funddetail/gridconfig"
      form-config="/bagong/accident_funddetail/formconfig"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
    >
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch, inject } from "vue";
import { DataList, util } from "suimjs";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  AccidentFundID: { type: String, default: () => "" },
});

const emit = defineEmits({
  "update:modelValue": null,
  deletedLines: null,
});

const listControl = ref(null);

const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  deleted: [],
});

function newRecord(r) {
  r.Date = new Date().toISOString();
  r.Mutation = 0;
  r.Notes = "";

  data.records.push(r);
  updateItems();
}

function onGridRowDelete(_, index) {
  if (props.AccidentFundID.length > 0) {
    data.deleted.push(data.records[index]);
    emit("deletedLines", data.deleted);
  }

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

function getSdkLines(id) {
  data.records = [];
  if (id.length > 0) {
    const url = `/bagong/accident_funddetail/gets?AccidentFundID=${id}`;
    axios.post(url, { Sort: ["_id"] }).then(
      (r) => {
        data.records = r.data.data;
        updateItems();
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

onMounted(() => {
  setTimeout(() => {
    getSdkLines(props.AccidentFundID);
    updateItems();
  }, 800);
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
