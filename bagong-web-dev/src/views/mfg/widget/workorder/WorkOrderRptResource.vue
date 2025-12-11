<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-material grid-line-items"
      v-model="data.value"
      editor
      hide-search
      hide-sort
      auto-commit-line
      no-confirm-delete
      hide-refresh-button
      hide-footer
      :hide-new-button="
        !['', 'DRAFT'].includes(props.item.WorkOrderPlanReportResourceStatus)
      "
      :hide-detail="true"
      :hide-select="true"
      :hide-delete-button="
        !['', 'DRAFT'].includes(props.item.WorkOrderPlanReportResourceStatus)
      "
      :hide-action="false"
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @delete-data="deleteRecord"
    >
      <template #item_buttons_1="{ item }">
        <log-trx
          :id="props.item.WorkOrderPlanReportResourceID"
          :hide-button="!['IN PROGRESS', 'END'].includes(props.item.Status)"
        />
      </template>
      <template #item_ExpenseType="{ item, idx }">
        <s-input
          ref="refExpenseType"
          v-model="item.ExpenseType"
          use-list
          :lookup-url="`/tenant/expensetype/find`"
          lookup-key="_id"
          :lookup-labels="['LedgerAccountID', 'Name']"
          :lookup-searchs="['_id', 'LedgerAccountID', 'Name']"
          class="w-full"
          :disabled="
            !['', 'DRAFT'].includes(
              props.item.WorkOrderPlanReportResourceStatus
            )
          "
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilder(
                search,
                ['_id', 'LedgerAccountID', 'Name'],
                item.ExpenseType,
                item
              )
          "
          @change="
            (field, v1) => {
              gridRowFieldChanged('ExpenseType', v1, item);
            }
          "
        ></s-input>
      </template>
      <template #item_Employee="{ item }">
        <s-input
          ref="refEmployee"
          v-model="item.Employee"
          use-list
          :lookup-url="`/bagong/employee/gets-filter?Department=DME002`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          :disabled="
            !['', 'DRAFT'].includes(
              props.item.WorkOrderPlanReportResourceStatus
            )
          "
        ></s-input>
      </template>
      <template #item_ActivityName="{ item }">
        <s-input
          ref="refActivityName"
          v-model="item.ActivityName"
          :disabled="
            !['', 'DRAFT'].includes(
              props.item.WorkOrderPlanReportResourceStatus
            )
          "
          multiRow="3"
          class="w-full"
        ></s-input>
      </template>
      <template #item_WorkingHour="{ item }">
        <s-input
          ref="refWorkingHour"
          v-model="item.WorkingHour"
          :disabled="
            !['', 'DRAFT'].includes(
              props.item.WorkOrderPlanReportResourceStatus
            )
          "
          kind="number"
          class="w-full"
        ></s-input>
      </template>
      <template #item_RatePerHour="{ item }">
        <s-input
          ref="refRatePerHour"
          v-model="item.RatePerHour"
          :disabled="
            !['', 'DRAFT'].includes(
              props.item.WorkOrderPlanReportResourceStatus
            )
          "
          kind="number"
          class="w-full"
        ></s-input>
      </template>
      <template #header_buttons_1="{ item }">
        <form-buttons-trx
          v-if="
            props.item.Status == 'IN PROGRESS' &&
            props.general.StatusOverall != 'END'
          "
          :key="data.btnTrxId"
          :moduleid="`mfg`"
          :autoPost="true"
          :status="props.item.WorkOrderPlanReportResourceStatus"
          :journal-id="props.item.WorkOrderPlanReportResourceID"
          :posting-profile-id="props.item.WorkOrderPlanReportResourcePPID"
          journal-type-id="Work Order Report Resource"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @preReopen="preReopen"
        />
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid, SButton } from "suimjs";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import moment from "moment";
const axios = inject("axios");
const props = defineProps({
  modelValue: { type: [Object, Array], default: () => [] },
  listActivity: { type: Array, default: () => [] },
  plan: { type: Array, default: () => [] },
  listPlan: { type: Array, default: () => [] },
  item: { type: Object, default: () => {} },
  general: { type: Object, default: () => {} },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  readOnly: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
  preReopen: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  btnTrxId: util.uuid(),
  gridCfg: {},
});
function newRecord() {
  const record = {};
  record.ActivityName = "";
  record.Employee = "";
  record.WorkingHour = 0;
  record.RatePerHour = 0;
  listControl.value.setRecords([record, ...listControl.value.getRecords()]);
}
function deleteRecord(record, index) {
  const newRecords = record.items.filter((dt, idx) => {
    return idx != index;
  });
  listControl.value.setRecords(newRecords);
}
function getDataValue() {
  return listControl.value.getRecords();
}
function onSelectDataLine() {
  return listControl.value.getRecords().filter((el) => el.isSelected == true);
}
function onChangeExpenseType(v1, v2, item) {
  item.WorkingHour = 0;
  if (typeof v1 == "string") {
    axios.post(`/tenant/expensetype/get`, [v1]).then(
      (r) => {
        item.WorkingHour = r.data.StandardCost;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}
function trxPreSubmit(status, action, doSubmit) {
  emit("preSubmit", status, action, doSubmit);
}
function trxPostSubmit(record) {
  emit("postSubmit");
}
function preReopen(status, action, doSubmit) {
  emit("preReopen", status, action, doSubmit);
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
      readOnly: colum[index].readOnly
        ? colum[index].readOnly
        : !["", "DRAFT"].includes(props.item.Status) ||
          props.general.StatusOverall == "END",
      input: {
        field: colum[index].field,
        label: colum[index].label,
        hint: "",
        hide: false,
        placeHolder: colum[index].label,
        kind: colum[index].kind,
        width: colum[index].width,
        multiRow: colum[index].multiRow ? colum[index].multiRow : 1,
        readOnly: colum[index].readOnly
          ? colum[index].readOnly
          : !["", "DRAFT"].includes(props.item.Status) ||
            props.general.StatusOverall == "END",
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
function createGridCfg() {
  const colms = [
    {
      field: "ExpenseType",
      label: "Expense Type",
      kind: "text",
      width: "200px",
    },
    {
      field: "ActivityName",
      label: "Activity",
      kind: "text",
      width: "200px",
      multiRow: 2,
    },
    {
      field: "Employee",
      label: "Employee",
      kind: "text",
      width: "200px",
    },
    {
      field: "WorkingHour",
      label: "Working Hour",
      kind: "number",
      width: "100px",
    },
    {
      field: "RatePerHour",
      label: "Rate Per Hour",
      kind: "number",
      width: "100px",
    },
  ];
  data.gridCfg = generateGridCfg(colms);
  data.value =
    props.item.WorkOrderPlanReportResourceLines == null
      ? []
      : props.item.WorkOrderPlanReportResourceLines;
}
function lookupPayloadBuilder(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

  //setting search
  // const queryItems = [
  //   {
  //     Field: "_id",
  //     Op: "$contains",
  //     Value: props.plan,
  //   },
  // ];
  // if (props.plan.length > 0) {
  //   qp.Where = {
  //     Op: "$and",
  //     items: queryItems,
  //   };
  // }
  if (search !== "" && search !== null) {
    qp.Where = {
      Op: "$or",
      items: [
        { Field: "_id", Op: "$contains", Value: [search] },
        { Field: "Name", Op: "$contains", Value: [search] },
      ],
    };
    // let items = [
    //   {
    //     Op: "$or",
    //     items: [
    //       { Field: "_id", Op: "$contains", Value: [search] },
    //       { Field: "Name", Op: "$contains", Value: [search] },
    //     ],
    //   },
    // ];
    // if (props.plan.length > 0) {
    //   items = [...items, ...queryItems];
    // }
    // qp.Where =
    //   props.plan.length > 0
    //     ? {
    //         Op: "$and",
    //         items: items,
    //       }
    //     : {
    //         Op: "$or",
    //         items: [
    //           { Field: "_id", Op: "$contains", Value: [search] },
    //           { Field: "Name", Op: "$contains", Value: [search] },
    //         ],
    //       };
  }
  return qp;
}
function gridRowFieldChanged(field, v1, item) {
  switch (field) {
    case "ExpenseType":
      let Expense = props.listPlan.find((e) => {
        return e.ExpenseType === v1;
      });
      if (Expense) {
        item.RatePerHour = Expense.RatePerHour;
      }
      break;
    default:
      break;
  }
}
onMounted(() => {
  createGridCfg();
});
defineExpose({
  getDataValue,
  onSelectDataLine,
});
</script>
<style scoped>
.title-header {
  font-size: 14px;
  font-weight: 600;
}
</style>
