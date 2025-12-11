<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-material grid-line-items"
      :editor="!props.hideNewButton"
      v-model="data.value"
      hide-search
      hide-sort
      auto-commit-line
      no-confirm-delete
      hide-refresh-button
      hide-footer
      :hide-new-button="props.hideNewButton"
      :hide-detail="props.hideDetail"
      :hide-select="props.hideSelect"
      :hide-action="!['', 'DRAFT'].includes(props.item.Status)"
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @delete-data="deleteRecord"
    >
      <template #item_ExpenseType="{ item }">
        <s-input
          ref="refExpenseType"
          v-model="item.ExpenseType"
          use-list
          :lookup-url="`/tenant/expensetype/find`"
          lookup-key="_id"
          :lookup-labels="['LedgerAccountID', 'Name']"
          :lookup-searchs="['_id', 'LedgerAccountID', 'Name']"
          :lookup-payload-builder="
            props.groupIdValue.length > 0 ? expensePayloadWorkOrder : undefined
          "
          :disabled="props.hideNewButton"
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeExpenseType(v1, v2, item);
            }
          "
        ></s-input>
      </template>
      <template #header_buttons_1="{ item }"> </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid, SButton } from "suimjs";
const axios = inject("axios");
const props = defineProps({
  modelValue: { type: [Object, Array], default: () => [] },
  item: { type: Object, default: () => {} },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  readOnly: { type: Boolean, default: false },
  hideNewButton: { type: Boolean, default: true },
  hideDetail: { type: Boolean, default: true },
  hideSelect: { type: Boolean, default: true },
  groupIdValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
  setPlan: null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  gridCfg: {},
});

function newRecord() {
  const record = {};
  record.ExpenseType = "";
  record.TargetHour = 0;
  record.RatePerHour = 0;
  record.UsedHour = 0;
  listControl.value.setRecords([record, ...listControl.value.getRecords()]);
}
function deleteRecord(record, index) {
  const newRecords = record.items.filter((dt, idx) => {
    return idx != index;
  });
  listControl.value.setRecords(newRecords);
  data.value = newRecords;
}
function getDataValue() {
  return listControl.value.getRecords();
  // return data.value;
}
function setDataValue(records) {
  data.value = records;
}
function addDataValue(record) {
  data.value.push(record);
}
function onSelectDataLine() {
  return listControl.value.getRecords().filter((el) => el.isSelected == true);
}
function onChangeExpenseType(v1, v2, item) {
  item.RatePerHour = 0;
  if (typeof v1 == "string") {
    axios.post(`/tenant/expensetype/get`, [v1]).then(
      (r) => {
        item.RatePerHour = r.data.StandardCost;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}
function generateGridCfg(colum) {
  let addColm = [];
  for (let index = 0; index < colum.length; index++) {
    addColm.push({
      field: colum[index].field,
      kind: colum[index].kind,
      label: colum[index].label,
      readType: "show",
      labelField: "",
      width: colum[index].width,
      multiRow: colum[index].multiRow,
      input: {
        field: colum[index].field,
        label: colum[index].label,
        hint: "",
        hide: false,
        placeHolder: colum[index].label,
        kind: colum[index].kind,
        width: colum[index].width,
        multiRow: colum[index].multiRow ? colum[index].multiRow : 1,
        readOnly: colum[index].readOnly ? colum[index].readOnly : false,
        useList: colum[index].useList ? colum[index].useList : false,
        useLookup: colum[index].useLookup ? colum[index].useLookup : false,
        lookupKey: colum[index].lookupKey ? colum[index].lookupKey : "",
        lookupLabels: colum[index].lookupLabels
          ? colum[index].lookupLabels
          : null,
        lookupSearchs: colum[index].lookupSearchs
          ? colum[index].lookupSearchs
          : null,
        lookupUrl: colum[index].lookupUrl ? colum[index].lookupUrl : "",
      },
    });
  }
  return {
    fields: addColm,
    setting: {
      idField: "_id",
      keywordFields: ["_id", "Name"],
      sortable: ["_id"],
    },
  };
}
function createGridCfgMaterial() {
  const colum = [
    {
      field: "ExpenseType",
      label: "Expense Type",
      kind: "text",
      width: "200px",
      useList: true,
      useLookup: true,
      lookupKey: "_id",
      lookupLabels: ["LedgerAccountID", "Name"],
      lookupSearchs: ["_id", "LedgerAccountID", "Name"],
      lookupUrl: "/tenant/expensetype/find",
    },
    {
      field: "TargetHour",
      label: "Target Hour",
      kind: "number",
      width: "100px",
    },
    {
      field: "RatePerHour",
      label: "Rate Per Hour",
      kind: "number",
      width: "100px",
    },
    {
      field: "UsedHour",
      label: "Used Hour",
      kind: "number",
      width: "100px",
      readOnly: true,
    },
  ];
  data.gridCfg = generateGridCfg(colum);
  util.nextTickN(2, () => {
    getsPlanResource();
  });
}
function getsPlanResource() {
  if (listControl.value) {
    listControl.value.setLoading(true);
  }
  axios
    .post(
      `/mfg/workorderplan/summary/resource/gets?WorkOrderPlanID=${props.item._id}`,
      {}
    )
    .then(
      (r) => {
        util.nextTickN(2, () => {
          data.value = [...r.data.data, ...props.modelValue];
          emit("setPlan", data.value);
        });
      },
      (e) => {
        data.value = props.modelValue;
        emit("setPlan", data.value);
        util.showError(e);
      }
    )
    .finally(function () {
      if (listControl.value) {
        listControl.value.setLoading(false);
      }
    });
}

function expensePayloadWorkOrder() {
  return {
    Take: 20,
    Sort: ["LedgerAccountID"],
    Select: ["_id", "LedgerAccountID", "Name"],
    Where: {
      Op: "$contains",
      Field: "GroupID",
      Value: props.groupIdValue,
    },
  };
}

onMounted(() => {
  createGridCfgMaterial();
});
defineExpose({
  getDataValue,
  setDataValue,
  addDataValue,
  onSelectDataLine,
  getsPlanResource,
});
</script>
<style scoped>
.title-header {
  font-size: 14px;
  font-weight: 600;
}
.label-item {
  font-size: 14px;
  font-weight: 600;
}
.label-sku {
  font-size: 14px;
}
</style>
