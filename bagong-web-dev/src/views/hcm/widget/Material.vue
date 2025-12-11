<template>
  <div>
    <data-list
      ref="listControl"
      hide-title
      no-gap
      :grid-editor="!readOnly"
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-hide-detail
      :grid-hide-delete="readOnly"
      :grid-hide-new="readOnly"
      :grid-hide-submit="readOnly"
      :form-hide-submit="readOnly"
      grid-hide-select
      grid-no-confirm-delete
      grid-hide-paging
      init-app-mode="grid"
      grid-mode="grid"
      new-record-type="grid"
      grid-config="/hcm/testschedule/gridconfig"
      form-config="/hcm/testschedule/formconfig"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @alterGridConfig="alterGridConfig"
      :form-default-mode="readOnly ? 'view' : 'edit'"
    >
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, onMounted, watch } from "vue";
import { DataList, util } from "suimjs";

const listControl = ref(null);

const axios = inject("axios");

const props = defineProps({
  testId: { type: String, default: "" },
  testScheduleType: { type: String, default: "" },
  readOnly: { type: Boolean, default: false },
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

function alterGridConfig(cfg) {
  cfg.fields = cfg.fields.filter(
    (el) =>
      [
        "TestScheduleType",
        "Status",
        "TestType",
        "TrainerName",
        "TrainerType",
      ].indexOf(el.field) == -1
  );
}

function newRecord(r) {
  r.TestID = props.testId;
  r.TestScheduleType = props.testScheduleType;
  r.TemplateID = "";
  r.Status = "OPEN";
  r.DateFrom = "";
  r.DateTo = "";
  r.TrainerName = "";
  r.TrainerType = "";

  data.records.push(r);
  updateItems();
}

function getsMaterial() {
  if (!props.testId) return;

  axios
    .post(`/hcm/testschedule/gets?TestID=${props.testId}`, {
      Sort: ["_id"],
    })
    .then(
      async (r) => {
        data.records = r.data.data;
        updateItems();
      },
      (e) => {
        util.showError(e);
      }
    );
}

function updateItems() {
  listControl.value?.setGridRecords(data.records);
}

function onGridRowDelete(records, index) {
  let payload = records;
  if (payload.items[index]._id) {
    axios.post("/hcm/testschedule/delete", payload.items[index]).then(
      async (r) => {
        const newRecords = data.records.filter((dt, idx) => {
          return idx != index;
        });
        data.records = newRecords;
        util.nextTickN(2, () => {
          updateItems();
        });
        util.showInfo("Data has been delete");
      },
      (e) => {
        util.showError(e);
      }
    );
  } else {
    const newRecords = data.records.filter((dt, idx) => {
      return idx != index;
    });
    data.records = newRecords;
    util.nextTickN(2, () => {
      updateItems();
    });
  }
  // let payload = records;
}

onMounted(() => {
  getsMaterial();
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
defineExpose({
  getsMaterial,
});
</script>
