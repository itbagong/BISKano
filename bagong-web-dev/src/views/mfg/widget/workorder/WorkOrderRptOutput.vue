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
        !['', 'DRAFT'].includes(props.item.WorkOrderPlanReportOutputStatus)
      "
      :hide-detail="true"
      :hide-select="true"
      :hide-delete-button="
        !['', 'DRAFT'].includes(props.item.WorkOrderPlanReportOutputStatus)
      "
      :hide-action="false"
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @delete-data="deleteRecord"
    >
      <template #item_buttons_1="{ item }">
        <log-trx
          :id="props.item.WorkOrderPlanReportOutputID"
          :hide-button="!['IN PROGRESS', 'END'].includes(props.item.Status)"
        />
      </template>
      <template #item_Type="{ item, idx }">
        <s-input
          ref="refType"
          v-model="item.Type"
          :disabled="
            !['', 'DRAFT'].includes(props.item.WorkOrderPlanReportOutputStatus)
          "
          useList
          class="min-w-[100px]"
          :items="[
            {
              key: 'WO Output',
              text: 'WO Output',
            },
            {
              key: 'Waste Item',
              text: 'Waste Item',
            },
            {
              key: 'Service',
              text: 'Service',
            },
          ]"
          @change="
            () => {
              item.ItemVarian = '';
              item.InventoryLedgerAccID = '';
              item.SKU = '';
              item.Description = '';
              item.Group = '';
              item.UnitID = '';
            }
          "
        ></s-input>
      </template>
      <template #item_InventoryLedgerAccID="{ item }">
        <div v-show="item.Type == 'Service'" class="w-full">
          <s-input-sku-item
            ref="refItemVarian"
            v-model="item.ItemVarian"
            :record="item"
            is-url
            :lookup-url="`/tenant/item/gets-detail?ItemGroupID=GRP0023`"
            :disabled="
              !['', 'DRAFT'].includes(
                props.item.WorkOrderPlanReportOutputStatus
              )
            "
            @afterOnChange="
              (val) => {
                onChangeItemVarian(item.ItemVarian, item.ItemVarian, item);
              }
            "
          ></s-input-sku-item>
        </div>
        <div v-show="item.Type == 'Waste Item'" class="w-full">
          <s-input-sku-item
            ref="refItemVarian"
            v-model="item.ItemVarian"
            :record="item"
            is-url
            :lookup-url="`/tenant/item/gets-detail?ItemGroupID=GRP0026`"
            :disabled="
              !['', 'DRAFT'].includes(
                props.item.WorkOrderPlanReportOutputStatus
              )
            "
            @afterOnChange="
              (val) => {
                onChangeItemVarian(item.ItemVarian, item.ItemVarian, item);
              }
            "
          ></s-input-sku-item>
        </div>
        <div v-show="item.Type == 'WO Output'" class="w-full">
          <s-input-sku-item
            ref="refItemVarian"
            v-model="item.ItemVarian"
            :record="item"
            :lookup-url="
              item.ItemVarian
                ? `/tenant/item/gets-detail?_id=${item.ItemVarian}`
                : `/tenant/item/gets-detail`
            "
            :disabled="
              !['', 'DRAFT'].includes(
                props.item.WorkOrderPlanReportOutputStatus
              )
            "
            @afterOnChange="
              (val) => {
                onChangeItemVarian(item.ItemVarian, item.ItemVarian, item);
              }
            "
            :lookup-payload-builder="
              (search) =>
                lookupPayloadBuilder(
                  search,
                  ['ID', 'Text'],
                  item.ItemVarian,
                  item
                )
            "
          ></s-input-sku-item>
        </div>
      </template>
      <template #item_Qty="{ item }">
        <s-input
          ref="refQty"
          v-model="item.Qty"
          :disabled="
            !['', 'DRAFT'].includes(props.item.WorkOrderPlanReportOutputStatus)
          "
          kind="number"
          class="w-full"
        ></s-input>
      </template>
      <template #item_GroupID="{ item }">
        {{ item.GroupID }}
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
          :autoReopen="false"
          :status="props.item.WorkOrderPlanReportOutputStatus"
          :journal-id="props.item.WorkOrderPlanReportOutputID"
          :posting-profile-id="props.item.WorkOrderPlanReportOutputPPID"
          journal-type-id="Work Order Report Output"
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
import { loadGridConfig, util, SInput, SGrid } from "suimjs";
import helper from "@/scripts/helper.js";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import SInputSkuItem from "../../../scm/widget/SInputSkuItem.vue";
const axios = inject("axios");
const refItemID = ref(null);
const props = defineProps({
  modelValue: { type: [Object, Array], default: () => [] },
  plan: { type: Array, default: () => [] },
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
const separatorID = "~~";
const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  btnTrxId: util.uuid(),
  gridCfg: {},
});
function newRecord() {
  const record = {};
  record.Type = "";
  record.ItemVarian = "";
  record.InventoryLedgerAccID = "";
  record.SKU = "";
  record.Description = "";
  record.Qty = 0;
  record.UnitID = "";
  record.UnitCost = "";
  listControl.value.setRecords([record, ...listControl.value.getRecords()]);
}
function deleteRecord(record, index) {
  const newRecords = record.items.filter((dt, idx) => {
    return idx != index;
  });
  listControl.value.setRecords(newRecords);
}
function onChangeItemVarian(v1, v2, item) {
  item.InventoryLedgerAccID = "";
  item.GroupID = "";
  if (typeof v1 == "string") {
    axios.post("/tenant/item/get", [v1.split("~~").at(0)]).then(
      (r) => {
        item.InventoryLedgerAccID = r.data._id;
        item.GroupID = r.data.ItemGroupID;
        item.Item = r.data;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function getDataValue() {
  return listControl.value.getRecords();
}
function onSelectDataLine() {
  return listControl.value.getRecords().filter((el) => el.isSelected == true);
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
        multiRow: colum[index].multiRow,
        readOnly: colum[index].readOnly
          ? colum[index].readOnly
          : !["", "DRAFT"].includes(props.item.Status) ||
            props.general.StatusOverall == "END",
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
      field: "Type",
      label: "Type",
      kind: "text",
      width: "100px",
      multiRow: 1,
      readOnly: false,
    },
    {
      field: "InventoryLedgerAccID",
      label: "Inventory Ledger Account",
      kind: "text",
      width: "300px",
      multiRow: 1,
      readOnly: false,
    },
    {
      field: "UnitID",
      label: "Unit",
      kind: "text",
      width: "100px",
      multiRow: 1,
      readOnly: true,
    },
    {
      field: "Qty",
      label: "Qty",
      kind: "number",
      width: "100px",
      multiRow: 1,
      readOnly: false,
    },
  ];
  data.gridCfg = generateGridCfg(colms);
  const output =
    props.item.WorkOrderPlanReportOutputLines == null
      ? []
      : props.item.WorkOrderPlanReportOutputLines;

  output.map((c) => {
    if (c.Type == "Waste Ledger") {
      c.ItemVarian = `${c.InventoryLedgerAccID}`;
    } else {
      c.ItemVarian = helper.ItemVarian(c.InventoryLedgerAccID, c.SKU);
    }
    return c;
  });
  data.value = output;
}
function lookupPayloadBuilder(search, select, value, item) {
  const qp = {};
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

  //setting search
  const query = [
    {
      Field: "ExcludeItemGroupID",
      Op: "$nin",
      Value: ["GRP0026", "GRP0023"],
    },
  ];
  qp.Where = {
    Op: "$and",
    items: query,
  };
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "Text", Op: "$contains", Value: [search] },
          { Field: "ID", Op: "$contains", Value: [search] },
        ],
      },
    ];
    items = [...items, ...query];
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
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
