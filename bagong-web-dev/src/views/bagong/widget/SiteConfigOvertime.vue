<template>
  <div>
    <div class="title section_title">Overtime</div>
    <div class="mb-5">
      <s-input class="w-[300px]" label="Salary Used" v-model="data.salaryUsed" use-list keep-label :items="['UMK Site','Basic Salary']" />
    </div>
    <s-grid
      v-if="!data.loading"
      ref="gridOvertime"
      class="w-full"
      editor
      hide-search
      hide-sort
      :hide-new-button="false"
      hide-refresh-button
      :hide-detail="true"
      hide-select
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @row-field-changed="onRowFieldChanged"
      @delete-data="onDelete"
    >
      <template #item_Divider="{ item }">
        <s-input
          v-if="item.Method != 'Flat'"
          :disabled="item.Method == 'Flat'"
          v-model="item.Divider"
          kind="number"
        ></s-input>
        <div v-else>&nbsp;</div>
      </template>
      <template #item_TUL="{ item }">
        <s-input
          v-if="item.Method == 'Flat'"
          :disabled="item.Method != 'Flat'"
          v-model="item.TUL"
          kind="number"
        ></s-input>
        <div v-else>&nbsp;</div>
      </template>
      <template #item_TULHoliday="{ item }">
        <s-input
          v-if="item.Method == 'Flat'"
          :disabled="item.Method != 'Flat'"
          v-model="item.TULHoliday"
          kind="number"
        ></s-input>
        <div v-else>&nbsp;</div>
      </template>
    </s-grid>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch, inject } from "vue";
import { loadFormConfig, loadGridConfig, SInput, SGrid, util } from "suimjs";

const props = defineProps({
  title: { type: String, default: "" },
  salaryUsed: { type: String, default: "" },
  modelValue: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
  "update:salaryUsed": null,
});

const axios = inject("axios");
const gridOvertime = ref(null);
const data = reactive({
  records: props.modelValue,
  formCfg: {},
  gridCfg: {},
  salaryUsed: props.salaryUsed,
});
function newRecord() {
  let obj = {
    Position: "",
    Method: "",
    Divider: "",
    TUL: "",
    TULHoliday: "",
  };
  data.records = [...gridOvertime.value.getRecords(), obj]
  gridOvertime.value.setRecords(data.records);
}
function onDelete(record, index) {
  const newRecords = gridOvertime.value.getRecords().filter((dt, idx) => {
    return idx != index;
  });
  gridOvertime.value.setRecords(newRecords);
}
function onRowFieldChanged(name, v1, v2, old, record) {
  if (name === 'Method') {
    record.Divider = null
    record.TUL = null
    record.TULHoliday = null
  }
  // listControl.value.setGridRecord(
  //   record,
  //   listControl.value.getGridCurrentIndex()
  // );
  updateItems();
}
function updateItems() {
  gridOvertime.value.setRecords(data.records);
  emit("update:modelValue", data.records);
}

function hideAndShow(name, val) {}

onMounted(() => {
  loadGridConfig(axios, "/bagong/overtime/gridconfig").then(
    (r) => {
      data.gridCfg = r;
      util.nextTickN(2, () => {
        updateItems();
      });
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
watch(
  () => data.salaryUsed,
  (nv) => {
    console.log(nv)
    emit("update:salaryUsed", nv);
  },
  { deep: true }
);
</script>
