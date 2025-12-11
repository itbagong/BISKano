<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-output grid-line-items"
      editor
      v-model="data.value"
      hide-search
      hide-sort
      auto-commit-line
      no-confirm-delete
      hide-footer
      hide-refresh-button
      :hide-new-button="props.hideNewButton"
      :hide-detail="props.hideDetail"
      :hide-select="props.hideSelect"
      :hide-action="!['', 'DRAFT'].includes(props.item.Status)"
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @delete-data="deleteRecord"
    >
      <template #item_Type="{ item, idx }">
        <s-input
          ref="refType"
          v-model="item.Type"
          :disabled="props.hideNewButton"
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
      <template #item_InventoryLedgerAccID="{ item, idx }">
        <s-input-sku-item
          v-if="item.Type == 'WO Output'"
          ref="refItemVarian"
          v-model="item.ItemVarian"
          :record="item"
          :disabled="props.hideNewButton"
          :lookup-url="`/tenant/item/gets-detail`"
          @afterOnChange="
            (val) => {
              onChangeItem(item.ItemID, item.ItemID, item);
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
        <s-input-sku-item
          v-else-if="item.Type == 'Service'"
          ref="refItemVarian"
          v-model="item.ItemVarian"
          :record="item"
          :disabled="props.hideNewButton"
          :lookup-url="`/tenant/item/gets-detail?ItemGroupID=GRP0023`"
          is-url
          @afterOnChange="
            (val) => {
              onChangeItem(item.ItemVarian, item.ItemVarian, item);
            }
          "
        ></s-input-sku-item>
        <s-input-sku-item
          v-else
          ref="refItemVarian"
          v-model="item.ItemVarian"
          :record="item"
          :disabled="props.hideNewButton"
          :lookup-url="`/tenant/item/gets-detail?ItemGroupID=GRP0026`"
          is-url
          @afterOnChange="
            (val) => {
              onChangeItem(item.ItemVarian, item.ItemVarian, item);
            }
          "
        ></s-input-sku-item>
      </template>
      <template #header_buttons_1="{ item }"> </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid, SButton } from "suimjs";
import helper from "@/scripts/helper.js";
import SInputSkuItem from "../../../scm/widget/SInputSkuItem.vue";
const axios = inject("axios");
const refItemID = ref(null);
const refSKU = ref(null);
const refUom = ref(null);
const props = defineProps({
  modelValue: { type: [Object, Array], default: () => [] },
  item: { type: Object, default: () => {} },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  readOnly: { type: Boolean, default: false },
  hideNewButton: { type: Boolean, default: true },
  hideDetail: { type: Boolean, default: true },
  hideSelect: { type: Boolean, default: true },
});

const emit = defineEmits({
  "update:modelValue": null,
  setPlan: null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  keyInvLedger: util.uuid(),
  gridCfg: {},
});

function newRecord() {
  const record = {};
  record.Type = "";
  record.ItemVarian = "";
  record.InventoryLedgerAccID = "";
  record.SKU = "";
  record.Description = "";
  record.Group = "";
  record.QtyAmount = 0;
  record.AchievedQtyAmount = 0;
  record.UnitID = "";
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
function onChangeItem(v1, v2, item) {
  item.InventoryLedgerAccID = "";
  item.Group = "";
  item.UnitID = "";
  if (typeof v1 == "string") {
    axios.post("/tenant/item/get", [v1.split("~~").at(0)]).then(
      (r) => {
        item.InventoryLedgerAccID = r.data._id;
        item.Group = r.data.ItemGroupID;
        item.UnitID = r.data.DefaultUnitID;
        item.Item = r.data;
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
        items: colum[index].items ? colum[index].items : [],
        multiRow: colum[index].multiRow ? colum[index].multiRow : 1,
        readOnly: colum[index].readOnly
          ? colum[index].readOnly
          : props.hideNewButton,
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
      field: "Type",
      label: "Type",
      kind: "text",
      width: "150px",
    },
    {
      field: "InventoryLedgerAccID",
      label: "Inv/Ledger",
      kind: "text",
      width: "350px",
      readOnly: false,
    },
    {
      field: "QtyAmount",
      label: "Qty Amount",
      kind: "number",
      width: "150px",
      readOnly: false,
    },
    {
      field: "AchievedQtyAmount",
      label: "Achieved Qty/Amount",
      kind: "number",
      width: "150px",
      readOnly: true,
    },
    {
      field: "UnitID",
      label: "Unit",
      kind: "text",
      width: "100px",
      readOnly: true,
    },
  ];
  data.gridCfg = generateGridCfg(colum);

  util.nextTickN(2, () => {
    getsPlanOutput();
  });
}
function getsPlanOutput() {
  if (listControl.value) {
    listControl.value.setLoading(true);
  }
  axios
    .post(
      `/mfg/workorderplan/summary/output/gets?WorkOrderPlanID=${props.item._id}`,
      {}
    )
    .then(
      (r) => {
        util.nextTickN(2, () => {
          r.data.data.map((c) => {
            c.ItemVarian = helper.ItemVarian(c.InventoryLedgerAccID, c.SKU);
            return c;
          });
          data.value = r.data.data;
          emit("setPlan", data.value);
        });
      },
      (e) => {
        props.modelValue.map((c) => {
          c.ItemVarian = helper.ItemVarian(c.InventoryLedgerAccID, c.SKU);
          return c;
        });
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
      // {
      //   Op: "$or",
      //   items: [
      //     { Field: "Text", Op: "$contains", Value: [search] },
      //     { Field: "_id", Op: "$contains", Value: [search] },
      //   ],
      // },
      { Field: "Text", Op: "$contains", Value: [search] },
      { Field: "_id", Op: "$contains", Value: [search] },
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
  createGridCfgMaterial();
});
defineExpose({
  getsPlanOutput,
  getDataValue,
  setDataValue,
  addDataValue,
  onSelectDataLine,
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
