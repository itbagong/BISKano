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
      grid-config="/bagong/assetbooking/lines/gridconfig"
      form-config="/bagong/assetbooking/lines/formconfig"
      grid-auto-commit-line
      stay-on-form-after-save
      @grid-row-field-changed="onGridRowFieldChanged"
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
    >
      <template #grid_item_buttons_1="{ item }">
        <s-button
          v-if="props.allocationAction"
          class="btn_primary mx-1"
          label="Allocated"
          @click="onClickAllocatedLines(item)"
        />
        <div v-else>&nbsp;</div>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch } from "vue";
import { DataList, SButton } from "suimjs";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  AssetBookingID: { type: String, default: () => "" },
  allocationAction: { type: Boolean, default: () => false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
  clickAllocatedLines: null,
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
  r.index = data.records.length;
  r.AssetBookingID = props.AssetBookingID;
  r.FromDate = new Date().toISOString();
  r.ToDate = new Date().toISOString();
  r.UnitBooked = 0;
  r.UnitPrice = 0;
  r.Total = 0;
  r.UnitAllocated = 0;

  data.records.push(r);
  updateItems();
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
  if (name == "UnitBooked") {
    record.Total = v1 * record.UnitPrice;
  }
  if (name == "UnitPrice") {
    record.Total = v1 * record.UnitBooked;
  }
  listControl.value.setGridRecord(
    record,
    listControl.value.getGridCurrentIndex()
  );
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
  emit("recalc");
}

function onClickAllocatedLines(item) {
  emit("clickAllocatedLines", { from: "lines", data: item });
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
  },
  { deep: true }
);

watch(
  () => props.allocationAction,
  (nv) => {},
  { deep: true }
);
</script>
