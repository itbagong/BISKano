<template>
  <div class="title section_title">Batch & Serial Numbers</div>
  <div class="flex flex-col gap-2">
    <data-list
      ref="listControl"
      title="Inventory Journal Line SN"
      class="grid-line-items"
      hide-title
      no-gap
      grid-hide-select
      grid-hide-detail
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-no-confirm-delete
      gridAutoCommitLine
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      grid-editor-no-form
      :form-hide-submit="
        ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(data.generalRecord.Status)
      "
      :grid-hide-control="
        ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(data.generalRecord.Status)
      "
      :grid-hide-delete="
        ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(data.generalRecord.Status)
      "
      :grid-editor="['', 'DRAFT'].includes(data.generalRecord.Status)"
      :grid-fields="['Qty']"
      grid-config="/scm/inventory/journal/batchserial/gridconfig"
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @grid-refreshed="gridRefreshed"
      @grid-row-save="onGridRowSave"
      @post-save="onFormPostSave"
      form-focus
    >
      <template #grid_Qty="{ item, idx }">
        <div>
          <s-input
            ref="qty"
            v-model="item.Qty"
            kind="number"
            class="w-full mb-0"
            :rules="rulesQty"
            :keepErrorSection="true"
            @change="
              (field, v1, v2, old, ctlRef) => {
                onGridFieldChanged(field, v1, v2, old, item);
              }
            "
          ></s-input>
        </div>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, inject } from "vue";
import { DataList, SInput, util } from "suimjs";

const listControl = ref(null);

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  transactionType: { type: String, default: () => "" },
  generalRecord: { type: Object, default: () => {} },
  lineRecord: { type: Object, default: () => {} },
});
const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});
const data = reactive({
  formMode: "edit",
  generalRecord: props.generalRecord,
  lineRecord: props.lineRecord,
  sumQty: 0,
  records: props.modelValue.map((dt) => {
    dt.suimRecordChange = false;
    return dt;
  }),
});
function newRecord() {
  if (data.records.length >= data.lineRecord.Qty) {
    return util.showError("Record can not bigger than Quantity");
  }
  if (data.sumQty >= data.lineRecord.Qty) {
    return util.showError("Maximal quantity is " + data.lineRecord.Qty);
  }
  const record = {};
  record.BatchID = "";
  record.SerialNumber = "";
  record.Qty = 1;
  data.records.push(record);
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
function onGridFieldChanged(name, v1, v2, old, record) {
  const tmpRecs = listControl.value.getGridRecords(); //[...data.records]\
  util.nextTickN(2, () => {
    const sum = tmpRecs.reduce((accumulator, object) => {
      return accumulator + object.Qty;
    }, 0);
    data.sumQty = sum;

    if (sum > data.lineRecord.Qty) {
      record.Qty = old
      util.showError("Maximal quantity is " + data.lineRecord.Qty);
    }
  });
}
function onGridRowSave(record, index) {
  record.suimRecordChange = false;
  data.records[index] = record;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onGridRowDelete(record, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  const sum = data.records.reduce((accumulator, object) => {
    return accumulator + object.Qty;
  }, 0);
  data.sumQty = sum;
  updateItems();
}
function updateItems() {
  const committedRecords = data.records.filter(
    (dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
  );
  emit("update:modelValue", committedRecords);
  emit("recalc");
}
const rulesQty = [
  (v) => {
    return v > data.lineRecord.Qty
      ? "Maximal quantity is " + data.lineRecord.Qty
      : "";
  },
];
function gridRefreshed() {
  listControl.value.setGridRecords(data.records);
}
</script>
