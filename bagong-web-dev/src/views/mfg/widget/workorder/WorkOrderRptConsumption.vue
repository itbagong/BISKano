<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      :id="`rpt-${
        ['POSTED', 'READY'].includes(
          props.item.WorkOrderPlanReportConsumptionStatus
        )
          ? props.typeMaterail
          : 'item'
      }-material`"
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
        props.typeMaterail == 'plan' ||
        !['', 'DRAFT'].includes(props.item.WorkOrderPlanReportConsumptionStatus)
      "
      :hide-detail="true"
      :hide-select="true"
      :hide-delete-button="
        !['', 'DRAFT'].includes(props.item.WorkOrderPlanReportConsumptionStatus)
      "
      :hide-action="false"
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @delete-data="deleteRecord"
    >
      <template #header_search="{ config }">
        <div class="w-full flex gap-2 justify-left items-center header"></div>
      </template>
      <template #item_buttons_1="{ item }">
        <log-trx
          :id="props.item.WorkOrderPlanReportConsumptionID"
          :hide-button="!['IN PROGRESS', 'END'].includes(props.item.Status)"
        />
      </template>
      <template #nodata="{ config }">
        <div class="nodata">
          No {{ props.typeMaterail == "plan" ? "Planned" : "Additional" }} Item
        </div>
      </template>
      <template #item_ItemID="{ item }">
        <s-input-sku-item
          ref="refItemVarian"
          v-model="item.ItemVarian"
          :record="item"
          :lookup-url="`/tenant/item/gets-detail?_id=${item.ItemVarian}`"
          :disabled="
            props.typeMaterail == 'plan' ||
            !['', 'DRAFT'].includes(
              props.item.WorkOrderPlanReportConsumptionStatus
            )
          "
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilder(
                search,
                ['_id', 'Name'],
                item.ItemVarian,
                item
              )
          "
        ></s-input-sku-item>
      </template>
      <template #item_UnitID="{ item }">
        <s-input
          ref="refUom"
          v-model="item.UnitID"
          :disabled="
            props.typeMaterail == 'plan' ||
            !['', 'DRAFT'].includes(
              props.item.WorkOrderPlanReportConsumptionStatus
            )
          "
          use-list
          :lookup-url="`/tenant/unit/gets-filter?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
      <template #item_Qty="{ item }">
        <s-input
          ref="refQty"
          v-model="item.Qty"
          :disabled="
            !['', 'DRAFT'].includes(
              props.item.WorkOrderPlanReportConsumptionStatus
            )
          "
          kind="number"
          class="w-full"
        ></s-input>
      </template>
      <template #item_Requested="{ item }">
        <div style="text-align: right">
          <mdicon name="check-bold" size="16" v-if="item.Requested" />
        </div>
      </template>
      <template #item_RequestedBy="{ item }">
        <s-input
          ref="refrequestor"
          v-model="item.RequestedBy"
          lookup-key="_id"
          label=""
          class="w-full"
          :disabled="
            props.typeMaterail == 'plan' ||
            !['', 'DRAFT'].includes(
              props.item.WorkOrderPlanReportConsumptionStatus
            )
          "
          use-list
          :lookup-url="`/tenant/employee/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #item_WarehouseLocation="{ item }">
        <s-input
          v-if="props.item.WorkOrderPlanReportConsumptionStatus == 'READY'"
          hide-label
          label="From warehouse"
          v-model="item.WarehouseLocation"
          class="w-full"
          use-list
          :disabled="
            props.item.WorkOrderPlanReportConsumptionStatus == 'POSTED'
          "
          :lookup-url="`/scm/item/balance/get-available-warehouse?ItemID=${item.ItemID}&SKU=${item.SKU}&WarehouseID=${props.warehouseID}`"
          lookup-key="_id"
          :lookup-labels="['Text']"
          :lookup-searchs="['_id', 'Text']"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onGetsAvailableWarehouse(v1, item);
            }
          "
        ></s-input>
        <s-input
          v-else
          hide-label
          label="From warehouse"
          v-model="item.WarehouseLocation"
          class="w-full"
          :readOnly="
            ['POSTED'].includes(props.item.WorkOrderPlanReportConsumptionStatus)
          "
        ></s-input>
      </template>
      <template #item_Total="{ item }">
        <div style="text-align: right">
          {{
            util.formatMoney(item.Qty * item.CostUnit, {
              decimal: 0,
            })
          }}
        </div>
      </template>
      <template #header_buttons_1="{ item }">
        <s-tooltip v-if="props.typeMaterail != 'plan'" tooltip="Create Request">
          <template #content>
            <s-button
              v-if="false"
              :icon="`basket-plus-outline`"
              class="bg-blue-800 text-white submit_btn"
              label=""
              :no-tooltip="true"
              @click="createItemRequest"
            />
          </template>
        </s-tooltip>
        <s-tooltip
          v-if="props.typeMaterail != 'plan'"
          tooltip="Available Stock"
        >
          <template #content>
            <s-button
              :icon="`database`"
              class="btn_success submit_btn"
              label=""
              :no-tooltip="true"
              @click="getAvailableStock"
            />
          </template>
        </s-tooltip>
        <form-buttons-trx
          v-if="
            props.item.Status == 'IN PROGRESS' &&
            props.general.StatusOverall != 'END' &&
            props.typeMaterail == 'plan'
          "
          :key="data.btnTrxId"
          :moduleid="`mfg`"
          :autoPost="false"
          :autoReopen="false"
          :status="props.item.WorkOrderPlanReportConsumptionStatus"
          :journal-id="props.item.WorkOrderPlanReportConsumptionID"
          :posting-profile-id="props.item.WorkOrderPlanReportConsumptionPPID"
          journal-type-id="Work Order Report Consumption"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          @preReopen="preReopen"
        />
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid, SButton, STooltip } from "suimjs";
import helper from "@/scripts/helper.js";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import SInputSkuItem from "../../../scm/widget/SInputSkuItem.vue";
const axios = inject("axios");
const refItemID = ref(null);
const refSKU = ref(null);
const refUom = ref(null);
const props = defineProps({
  modelValue: { type: [Object, Array], default: () => [] },
  item: { type: Object, default: () => {} },
  plan: { type: Array, default: () => [] },
  general: { type: Object, default: () => {} },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  warehouseID: { type: String, default: () => "" },
  typeMaterail: { type: String, default: () => "additionl" },
  readOnly: { type: Boolean, default: false },
  filterItems: {
    type: Array,
    default() {
      return [];
    },
  },
});
const separatorID = "~~";
const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
  preReopen: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  btnTrxId: util.uuid(),
  typeItem: "Plan",
  gridCfg: {},
});
function newRecord() {
  const record = {};
  record.Type = props.typeMaterail;
  record.ItemID = "";
  record.SKU = "";
  record.UnitID = "";
  record.RequestedBy = "";
  record.Qty = 0;
  record.QtyAvailable = 0;
  record.Requested = false;
  record.Total = 0;
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
function setRecords(newDataSet) {
  data.value = newDataSet;
}
function onSelectDataLine() {
  return listControl.value.getRecords().filter((el) => el.isSelected == true);
}
function onGetsAvailableWarehouse(_id, item) {
  if (_id) {
    axios
      .post(
        `/scm/item/balance/get-available-warehouse?ItemID=${item.ItemID}&SKU=${item.SKU}&WarehouseID=${props.warehouseID}`
      )
      .then(
        (r) => {
          const wh = r.data.find(function (v) {
            return v._id == _id;
          });
          if (wh) {
            item.WarehouseLocation = wh.Text;
            item.InventDim = wh.InventDim;
          }
        },
        (e) => {
          return util.showError(e);
        }
      );
  } else {
    item.WarehouseLocation = "";
    delete item.InventDim;
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
      readOnly: colum[index].readOnly
        ? colum[index].readOnly
        : !["", "DRAFT"].includes(props.item.Status) ||
          props.general.StatusOverall == "END",
      input: {
        field: colum[index].field,
        label: colum[index].label,
        readOnly: colum[index].readOnly
          ? colum[index].readOnly
          : !["", "DRAFT"].includes(props.item.Status) ||
            props.general.StatusOverall == "END",
        hint: "",
        hide: false,
        placeHolder: colum[index].label,
        kind: colum[index].kind,
        width: colum[index].width,
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
function trxPreSubmit(status, action, doSubmit) {
  emit("preSubmit", status, action, doSubmit);
}
function trxPostSubmit(record) {
  emit("postSubmit");
}
function trxErrorSubmit(record) {
  emit("errorSubmit");
}
function preReopen(status, action, doSubmit) {
  emit("preReopen", status, action, doSubmit);
}

function createGridCfg() {
  const colms = [
    {
      field: "ItemID",
      label: `Item ${props.typeMaterail}`,
      kind: "text",
      width: "300px",
    },
    {
      field: "UnitID",
      label: "UoM",
      kind: "text",
      width: "100px",
    },
    {
      field: "Qty",
      label: "Qty",
      kind: "number",
      width: "100px",
      readOnly: false,
    },
    {
      field: "QtyAvailable",
      label:
        props.typeMaterail != "plan"
          ? "Qty Available"
          : "Remaining Planned Qty",
      kind: "number",
      width: "100px",
      readOnly: true,
    },
  ];
  if (props.typeMaterail != "plan") {
    colms.push({
      field: "RequestedBy",
      label: "Requested By",
      kind: "text",
      width: "200px",
    });
  }

  if (
    props.typeMaterail != "plan" &&
    ["READY", "POSTED"].includes(
      props.item.WorkOrderPlanReportConsumptionStatus
    )
  ) {
    colms.push({
      field: "WarehouseLocation",
      label: "Warehouse Location",
      kind: "text",
      width: "300px",
    });
  }

  data.gridCfg = generateGridCfg(colms);

  let itemPlan =
    props.item.WorkOrderPlanReportConsumptionLines == null
      ? []
      : props.item.WorkOrderPlanReportConsumptionLines;
  itemPlan.map((c) => {
    c.ItemVarian = helper.ItemVarian(c.ItemID, c.SKU);
    return c;
  });

  let itemAdditional =
    props.item.WorkOrderPlanReportConsumptionAdditionalLines == null
      ? []
      : props.item.WorkOrderPlanReportConsumptionAdditionalLines;
  itemAdditional.map((c) => {
    c.ItemVarian = helper.ItemVarian(c.ItemID, c.SKU);
    return c;
  });
  let plan = JSON.parse(JSON.stringify(props.plan)).map((p) => {
    const exists = itemPlan.find(function (i) {
      return `${i.ItemID}${i.SKU}` == `${p.ItemID}${p.SKU}`;
    });
    let Used = p.Used ? p.Used : 0;
    if (!exists) {
      const record = {
        Type: props.typeMaterail,
        Qty: 0,
        QtyAvailable: p.Required - Math.abs(Used),
        Requested: false,
        Total: 0,
      };
      return { ...p, ...record };
    } else {
      const rptPlan = { ...exists, ...p };
      rptPlan.QtyAvailable = p.Required - Math.abs(Used);
      return rptPlan;
    }
  });
  data.value = props.typeMaterail == "plan" ? plan : itemAdditional;
}
function createItemRequest() {
  emit("createItemRequest");
}
function getAvailableStock() {
  emit("getAvailableStock");
}
function lookupPayloadBuilder(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

  //setting search
  const queryItems = [
    {
      Field: "_id",
      Op: "$eq",
      Value: "CSM0000",
    },
  ];
  if (props.filterItems.length > 0) {
    qp.Where = {
      Op: "$and",
      items: queryItems,
    };
  }
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "ID", Op: "$contains", Value: [search] },
          { Field: "Text", Op: "$contains", Value: [search] },
        ],
      },
    ];
    if (props.filterItems.length > 0) {
      items = [...items, ...queryItems];
    }
    qp.Where =
      props.filterItems.length > 0
        ? {
            Op: "$and",
            items: items,
          }
        : {
            Op: "$or",
            items: [
              { Field: "ID", Op: "$contains", Value: [search] },
              { Field: "Text", Op: "$contains", Value: [search] },
            ],
          };
  }
  return qp;
}
onMounted(() => {
  createGridCfg();
});
defineExpose({
  getDataValue,
  setRecords,
  onSelectDataLine,
});
</script>
<style scoped>
.title-header {
  font-size: 14px;
  font-weight: 600;
}
</style>
<style>
#rpt-additional-material
  > div:nth-child(2)
  > div
  > table
  > thead
  > tr
  > th:nth-child(6) {
  background-color: #fd6e76 !important;
  color: white !important;
}
</style>
