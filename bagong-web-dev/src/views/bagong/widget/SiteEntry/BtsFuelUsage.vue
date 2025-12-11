<template>
  <div>
    <label class="input_label">Ritase</label>
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
      :grid-fields="['KMTempuh']"
      new-record-type="grid"
      grid-config="/bagong/siteentry_btsdetail/fuelusage/gridconfig"
      form-config="/bagong/siteentry_btsdetail/fuelusage/formconfig"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="gridRowFieldChange"
    >
      <template #grid_KMTempuh="{ item }">
        <s-input
          kind="number"
          v-model="item.KMTempuh"
          hide-label
          disabled
        ></s-input>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch } from "vue";
import { DataList, SInput } from "suimjs";

const props = defineProps({
  siteEntryAssetID: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
  getFuelUsageRecords: null,
});

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
});

function newRecord(r) {
  r.VendorID = "";
  r.Volume = "";
  r.KMStart = "";
  r.KMEnd = "";
  r.KMTempuh = "";
  r.Amount = "";
  r.TotalAmount = "";
  r.Notes = "";

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

function gridRowFieldChange(name, v1, v2, old, record) {
  if (name == "KMStart") {
    record.KMTempuh = record.KMEnd - v1;
  }
  if (name == "KMEnd") {
    record.KMTempuh = v1 - record.KMStart;
  }
  if (name == "Volume") {
    record.TotalAmount = v1 * record.Amount;
  }
  if (name == "Amount") {
    record.TotalAmount = v1 * record.Volume;
  }
}

onMounted(() => {
  setTimeout(() => {
    updateItems();
  }, 800);
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
    emit("getFuelUsageRecords", nv);
  },
  { deep: true }
);
</script>
