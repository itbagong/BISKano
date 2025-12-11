<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-line grid-line-items"
      editor
      hide-search
      hide-select
      hide-sort
      :hide-new-button="false"
      :hide-delete-button="false"
      hide-refresh-button
      :hide-detail="true"
      :hide-action="false"
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @delete-data="deleteRecord"
    >
      <template #item_ItemID="{ item, idx }">
        <s-input
          ref="refItemID"
          v-model="item.ItemID"
          :disabled="false"
          use-list
          :lookup-url="`/tenant/item/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeItem(v1, v2, item);
            }
          "
        ></s-input>
      </template>
      <template #item_SKU="{ item }">
        <s-input
          v-model="item.SKU"
          :disabled="
            ['SUBMITTED', 'POSTED', 'READY'].includes(props.item.Status)
          "
          use-list
          :lookup-url="`/tenant/itemspec/gets-info?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Description']"
          :lookup-searchs="['_id', 'SKU', 'Description']"
          class="w-full"
          @change="(...args) => handleChangeSKU(...args, item)"
        ></s-input>
      </template>
      <template #item_UoM="{ item }">
        <s-input
          ref="refUom"
          v-model="item.UoM"
          :disabled="false"
          use-list
          :lookup-url="`/tenant/unit/gets-filter?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid } from "suimjs";
const axios = inject("axios");
const refItemID = ref(null);
const refSKU = ref(null);
const refUom = ref(null);
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
  itemID: { type: String, default: () => "" },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  formMode: { type: String, default: () => "new" },
  activeFields: { type: Array, default: () => [] },
  readOnly: { type: Boolean, defaule: false },
  hideDetail: { type: Boolean, defaule: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  typeMoveIn: "line",
  gridCfg: {},
});

function newRecord() {
  const record = {};
  record.PurchaseRequestID = props.itemID;
  record.ItemID = "";
  record.SKU = "";
  record.Qty = 0;
  record.UoM = "";
  record.UnitPrice = 0;
  record.Tax = 0;
  record.Discount = 0;
  record.SubTotal = 0;
  record.Remarks = "";
  record.Item = {
    PhysicalDimension: {
      IsEnabledItemBatch: false,
      IsEnabledItemSerial: false,
    },
  };
  record.InventDim = props.item.Location;
  listControl.value.setRecords([...listControl.value.getRecords(), record]);
}

function deleteRecord(record, index) {
  const newRecords = record.items.filter((dt, idx) => {
    return idx != index;
  });
  listControl.value.setRecords(newRecords);
}

function onChangeItem(v1, v2, item) {
  if (typeof v1 != "string") {
    item.UoM = "";
    item.Item = {
      PhysicalDimension: {
        IsEnabledItemBatch: false,
        IsEnabledItemSerial: false,
      },
    };
  } else {
    axios.post("/tenant/item/get", [v1]).then(
      (r) => {
        item.UoM = r.data.DefaultUnitID;
        item.Item = r.data;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function handleChangeSKU(name, v1, v2, old, ctlRef, item, idx) {
  axios.post("/tenant/itemspec/gets-detail", [v1]).then((r) => {
    const res = r.data[0];
    item.Description = res.Description;
  });
}

function getDataValue() {
  return listControl.value.getRecords();
}
onMounted(() => {
  loadGridConfig(axios, props.gridConfig).then(
    (r) => {
      let tbLine = [
        "ItemID",
        "SKU",
        "Description",
        "UoM",
        "Qty",
        "UnitPrice",
        "Tax",
        "Discount",
        "SubTotal",
        "Remarks",
      ];
      const _fields = r.fields.filter((o) => {
        if (
          ["Qty", "UoM", "UnitPrice", "Tax", "Discount", "SubTotal"].includes(
            o.field
          )
        ) {
          o.width = "150px";
        } else {
          o.width = "300px";
        }
        o.idx = tbLine.indexOf(o.field);
        return tbLine.includes(o.field);
      });
      data.gridCfg = {
        ...r,
        fields: _fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
    },
    (e) => util.showError(e)
  );
  axios
    .post(props.gridRead, {
      Skip: 0,
      Take: 0,
      Sort: ["_id"],
    })
    .then(
      (r) => {
        if (listControl.value) {
          listControl.value.setRecords(r.data.data);
        }
      },
      (e) => util.showError(e)
    );
});
defineExpose({
  getDataValue,
});
</script>
<style>
/* .tb-line > div:nth-child(2) > div {
  overflow-x: auto;
  padding-bottom: 100px;
}
.tb-line > div:nth-child(2) > div > table {
  width: calc(100% + 40%) !important;
} */
</style>
