<template>
  <div class="flex flex-col gap-2">
    <data-list
      ref="listLineControl"
      title="Item Request Line"
      class="grid-line-items"
      hide-title
      no-gap
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-no-confirm-delete
      gridAutoCommitLine
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      :grid-hide-control="!data.editable"
      :grid-editor-no-form="true"
      grid-hide-select
      :grid-hide-delete="!data.editable"
      grid-hide-detail
      :grid-editor="data.editable"
      grid-config="/scm/item/request/detail/line/gridconfig"
      form-config="/scm/item/request/detail/line/formconfig"
      :grid-fields="['FulfillmentType', 'QtyAvailable', 'WarehouseID', 'UoM']"
      @grid-row-add="newRecord"
      @form-field-change="onFormFieldChanged"
      @form-edit-data="openForm"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="onGridRowFieldChanged"
      @grid-refreshed="gridRefreshed"
      @grid-row-save="onGridRowSave"
      @alter-grid-config="alterGridConfig"
      form-focus
    >
      <template #grid_FulfillmentType="{ item, idx }">
        <s-input
          v-model="item.FulfillmentType"
          :disabled="!data.editable"
          use-list
          class="w-full"
          :items="data.fulfillmentType"
          @change="
            (field, v1, v2, old, ctlRef) => {
              item.WarehouseID = '';
              item.QtyAvailable = 0;
              item.InventDimFrom = {};
            }
          "
        ></s-input>
      </template>
      <template #grid_UoM="{ item }">
        <s-input
          ref="refUom"
          v-model="item.UoM"
          disabled
          use-list
          :lookup-url="`/tenant/unit/gets-filter?ItemID=${data.lineRecord.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>

      <template #grid_WarehouseID="{ item }">
        <s-input
          hide-label
          label="From warehouse"
          v-model="item.WarehouseID"
          class="w-full"
          :disabled="item.FulfillmentType != 'Item Transfer'"
          :use-list="item.FulfillmentType == 'Item Transfer'"
          :lookup-url="`/scm/item/balance/get-available-warehouse?ItemID=${data.lineRecord.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Text']"
          :lookup-searchs="['_id', 'Text']"
          @change="
            (field, v1, v2, old, ctlRef) => {
              const wh = data.listAvailableWarehouse.find(function (v) {
                return v._id == v1;
              });
              if (wh) {
                item.InventDimFrom = wh.InventDim;
                item.QtyAvailable = wh.Qty;
              }
            }
          "
        ></s-input>
      </template>
      <template #grid_QtyAvailable="{ item }">
        <s-input
          v-model="item.QtyAvailable"
          disabled
          kind="number"
          :class="
            item.QtyFulfilled <= item.QtyAvailable
              ? 'text-green-700 font-semibold'
              : 'text-red-700 font-semibold'
          "
        ></s-input>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, DataList } from "suimjs";
const axios = inject("axios");
const listLineControl = ref(null);
const refItemID = ref(null);

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  lineRecord: { type: Object, default: () => {} },
  editable: { type: Boolean, default: true },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  lineRecord: props.lineRecord,
  editable: props.editable,
  records: props.modelValue.map((dt) => {
    dt.suimRecordChange = false;
    return dt;
  }),
  fulfillmentType: [
    {
      text: "Item Transfer",
      value: "Item Transfer",
    },
    {
      text: "Purchase Request",
      value: "Purchase Request",
    },
    // {
    //   text: "Movement In",
    //   value: "Movement In"
    // },
    {
      text: "Movement Out",
      value: "Movement Out",
    },
    {
      text: "Assembly",
      value: "Assembly",
    },
  ],
  listAvailableWarehouse: [],
});
function newRecord() {
  if (data.lineRecord.QtyRequested === data.records.length) {
    return util.showError("Record can not bigger than Quantity Requested");
  }
  const record = {};
  record.FulfillmentType = "";
  record.QtyFulfilled = "";
  record.WarehouseID = "";
  record.QtyAvailable = 0;
  record.InventDimFrom = {};
  record.Remarks = "";
  record.UoM = data.lineRecord.UoM;
  data.records.push(record);
  listLineControl.value.setGridRecords(data.records);
  updateItems();
}
function openForm(record) {}

function onGridRowSave(record, index) {
  record.suimRecordChange = false;
  data.records[index] = record;
  listLineControl.value.setGridRecords(data.records);
  updateItems();
}

function onGridRowDelete(record, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listLineControl.value.setGridRecords(data.records);
  updateItems();
}

function updateItems() {
  const committedRecords = data.records.filter(
    (dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
  );
  emit("update:modelValue", committedRecords);
  emit("recalc");
}
function gridRefreshed() {
  listLineControl.value.setGridRecords(data.records);
}
function onGridRowFieldChanged(name, v1, v2, old, record) {
  console.log(name, v1, v2);
  if (name == "WarehouseID") {
    onGetStock(record, v1);
  }
}
function onGetStock(record, warehouseID) {
  const payload = [
    {
      ...record,
      ...data.lineRecord,
      InventDim: {
        WarehouseID: warehouseID,
      },
    },
  ];
  axios.post("/scm/item/balance/get-qty", payload).then(
    (r) => {
      console.log(r.data);
      record.QtyAvailable =
        r.data.find((item) => item.ItemID === data.lineRecord.ItemID)
          ?.QtyAvail ?? 0;
    },
    (e) => {
      return util.showError(e);
    }
  );
}
function alterGridConfig(cfg) {
  const newFields = [
    {
      field: "QtyAvailable",
      kind: "number",
      label: "Qty Available",
      readType: "show",
      disable: true,
      input: {
        field: "QtyAvailable",
        label: "Qty Available",
        hint: "",
        hide: false,
        placeHolder: "QtyAvailable",
        kind: "number",
        disable: true,
        required: false,
        multiple: false,
      },
    },
  ];
  const fields = [...cfg.fields, ...newFields];
  cfg.fields = fields;
}

function onGetsAvailableWarehouse() {
  axios
    .post(
      `/scm/item/balance/get-available-warehouse?ItemID=${data.lineRecord.ItemID}`
    )
    .then(
      (r) => {
        data.listAvailableWarehouse = r.data;
      },
      (e) => {
        return util.showError(e);
      }
    );
}
onMounted(() => {
  onGetsAvailableWarehouse();
});
</script>
