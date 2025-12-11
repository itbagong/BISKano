<template>
  <div class="flex flex-col gap-2">
    <s-grid
      :id="['READY', 'POSTED'].includes(props.item.Status) ? 'tb-material' : ''"
      ref="listControl"
      class="w-full tb-material grid-line-items"
      v-model="data.value"
      hide-search
      hide-sort
      :editor="true"
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
      <template #item_ItemID="{ item, idx }">
        <s-input-sku-item
          ref="refItemVarian"
          v-model="item.ItemVarian"
          :record="item"
          :disabled="props.hideNewButton"
          :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
            item.ItemID,
            item.SKU
          )}`"
        ></s-input-sku-item>
      </template>
      <template #item_UnitID="{ item }">
        <s-input
          v-model="item.UnitID"
          :key="item.UnitID"
          :disabled="props.hideNewButton"
          use-list
          :lookup-url="`/tenant/unit/gets-filter?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>

      <template #item_Required="{ item, idx }">
        <s-input
          v-if="['', 'DRAFT'].includes(props.item.Status)"
          ref="refUom"
          v-model="item.Required"
          kind="number"
          class="w-full"
        ></s-input>
      </template>
      <template #item_WarehouseLocation="{ item }">
        <s-input
          v-if="props.item.Status == 'READY'"
          hide-label
          label="From warehouse"
          v-model="item.Location"
          class="w-full"
          use-list
          :lookup-url="`/scm/item/balance/get-available-warehouse?ItemID=${item.ItemID}&SKU=${item.SKU}&WarehouseID=${data.warehouseID}`"
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
          :readOnly="['POSTED'].includes(props.item.Status)"
        ></s-input>
      </template>
      <template #header_buttons_1="{ item }">
        <div class="w-full flex gap-2">
          <s-button
            v-if="false"
            :icon="`basket-plus-outline`"
            class="bg-blue-800 text-white submit_btn"
            label="Create Item Request"
            @click="createRequest"
          />
          <div>
            <div class="w-[145px] h-[30px]" v-if="data.loadingBtnStock">
              <loader kind="skeleton" skeleton-kind="input" />
            </div>
            <s-button
              v-else
              :icon="`database`"
              class="btn_success submit_btn"
              label="Get Available Stock"
              @click="getAvailableStock"
            />
          </div>
        </div>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid, SButton } from "suimjs";
import SInputSkuItem from "../../../scm/widget/SInputSkuItem.vue";
import Loader from "@/components/common/Loader.vue";
import helper from "@/scripts/helper.js";
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
});

const emit = defineEmits({
  "update:modelValue": null,
  setPlan: null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  loadingBtnStock: false,
  warehouseID: "",
  gridCfg: {},
});

function newRecord() {
  // emit("OpentFromMaterial", record);
  const record = {};
  record.InventoryLedgerAccID = "";
  record.ItemID = "";
  record.SKU = "";
  record.Required = 0;
  record.Reserved = 0;
  record.AvailableStock = 0;
  record.Used = 0;
  record.UnitID = "";
  record.ItemName = "";
  record.SKUName = "";
  listControl.value.setRecords([record, ...listControl.value.getRecords()]);
  updateItems();
}
function deleteRecord(record, index) {
  const newRecords = record.items.filter((dt, idx) => {
    return idx != index;
  });
  newRecords.map((dt) => {
    dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
    return dt;
  });
  listControl.value.setRecords(newRecords);
  data.value = newRecords;
}
function getDataValue() {
  return listControl.value.getRecords();
  // return data.value;
}
function setDataValue(records) {
  records.map((dt) => {
    dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
    return dt;
  });
  data.value = records;
}
function addDataValue(record) {
  record.ItemVarian = helper.ItemVarian(record.ItemID, record.SKU);
  data.value.push(record);
}

function updateItems() {
  emit("update:modelValue", listControl.value.getRecords());
  data.value = listControl.value.getRecords();
}

function onSelectDataLine() {
  return listControl.value.getRecords().filter((el) => el.isSelected == true);
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
        : !["", "DRAFT"].includes(props.item.Status),
      input: {
        field: colum[index].field,
        label: colum[index].label,
        hint: "",
        hide: false,
        placeHolder: colum[index].label,
        kind: colum[index].kind,
        width: colum[index].width,
        readOnly: colum[index].readOnly
          ? colum[index].readOnly
          : !["", "DRAFT"].includes(props.item.Status),
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
function createGridCfgMaterial(load = false) {
  const colum = [
    {
      field: "ItemID",
      kind: "text",
      label: "Item",
      width: "200px",
      readOnly: true,
    },
    {
      field: "UnitID",
      kind: "text",
      label: "Unit",
      width: "150px",
      readOnly: true,
    },

    {
      field: "Required",
      kind: "number",
      label: "Required",
      width: "100px",
      readOnly: false,
    },
    {
      field: "Reserved",
      kind: "number",
      label: "Reserved",
      width: "100px",
      readOnly: true,
    },
    {
      field: "AvailableStock",
      kind: "number",
      label: "Available Stock",
      width: "100px",
      readOnly: true,
    },
    {
      field: "Used",
      kind: "number",
      label: "Used",
      width: "100px",
      readOnly: true,
    },
  ];
  util.nextTickN(2, () => {
    if (["WOGeneral"].includes(props.item.JournalTypeID)) {
      colum.push({
        field: "Remarks",
        kind: "text",
        label: "Remarks",
        width: "300px",
        readOnly: false,
      });
    }

    if (["READY", "POSTED"].includes(props.item.Status)) {
      colum.push({
        field: "WarehouseLocation",
        kind: "text",
        label: "Warehouse Location",
        width: "300px",
        readOnly: false,
      });
    }
    data.gridCfg = generateGridCfg(colum);
    if (load) {
      getsPlanMaterial();
    }
  });
}

function getsPlanMaterial() {
  data.loadingBtnStock = true;
  if (listControl.value) {
    listControl.value.setLoading(true);
  }
  axios
    .post(
      `/mfg/workorderplan/summary/material/gets?WorkOrderPlanID=${props.item._id}`,
      {}
    )
    .then(
      (r) => {
        util.nextTickN(2, () => {
          data.value = [...r.data.data, ...props.modelValue].map((dt) => {
            dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
            return dt;
          });
          data.loadingBtnStock = false;
          util.nextTickN(2, () => {
            const el = document.querySelector(
              `#tb-material .suim_table > thead > tr > th:nth-child(${
                props.item.JournalTypeID == "WOGeneral" ? "8" : "7"
              })`
            );
            if (el) {
              el.style.backgroundColor = "#fd6e76";
              el.style.color = "white";
            }
          });
          emit("setPlan", data.value);
        });
      },
      (e) => {
        props.modelValue.map((dt) => {
          dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
          return dt;
        });
        data.value = props.modelValue;
        util.nextTickN(2, () => {
          const el = document.querySelector(
            `#tb-material .suim_table > thead > tr > th:nth-child(${
              props.item.JournalTypeID == "WOGeneral" ? "8" : "7"
            })`
          );
          el.style.backgroundColor = "#fd6e76";
          el.style.color = "white";
        });
        data.loadingBtnStock = false;
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
function createRequest() {}

function getWarehouseOnSite(siteId, cbOK, cbFalse) {
  let Site =
    props.item.Dimension &&
    props.item.Dimension.find((_dim) => _dim.Key === "Site") &&
    props.item.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? props.item.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : undefined;
  if (!siteId) {
    siteId = Site;
  }
  axios
    .post(`/tenant/warehouse/find`, {
      Where: {
        Op: "$and",
        Items: [
          {
            Field: "Dimension.Key",
            Op: "$eq",
            Value: "Site",
          },
          {
            Field: "Dimension.Value",
            Op: "$eq",
            Value: siteId,
          },
        ],
      },
    })
    .then(
      (r) => {
        if (r.data.length > 0) {
          data.warehouseID = r.data[0]._id;
        }
        if (cbOK) {
          cbOK(r.data);
        }
      },
      (e) => {
        util.showError(e);
        if (cbFalse) {
          cbFalse();
        }
      }
    );
}

function getAvailableStock(cbOK, cbFalse) {
  data.loadingBtnStock = true;
  getWarehouseOnSite(
    "",
    (site) => {
      let InventDim = props.item.InventDim;
      if (site.length > 0) {
        InventDim.WarehouseID = site[0]._id;
        axios
          .post(`/mfg/workorderplan/gets-available-stock`, {
            InventDim: props.item.InventDim,
            Items: data.value,
            GroupBy: ["WarehouseID"],
            BalanceFilter: { WarehouseIDs: [props.item.InventDim.WarehouseID] },
          })
          .then(
            (r) => {
              data.value.map(function (d) {
                if (!d.SKU) {
                  d["SKU"] = "";
                }
                const item = r.data.filter(function (v) {
                  return v.ItemID == d.ItemID && v.SKU == d.SKU;
                });

                d.AvailableStock = item.reduce((a, b) => {
                  return a + b.Qty;
                }, 0);
                return d;
              });
              data.loadingBtnStock = false;
              if (cbOK) {
                cbOK();
              }
            },
            (e) => {
              data.loadingBtnStock = false;
              if (cbFalse) {
                cbFalse();
              }
              util.showError(e);
            }
          );
      } else {
        data.value.map(function (d) {
          d.AvailableStock = 0;
          return d;
        });
        if (cbFalse) {
          cbFalse();
        }
        data.loadingBtnStock = false;
      }
    },
    null
  );
}
function onGetsAvailableWarehouse(_id, item) {
  if (_id) {
    axios
      .post(
        `/scm/item/balance/get-available-warehouse?ItemID=${item.ItemID}&SKU=${item.SKU}&WarehouseID=${data.warehouseID}`
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
onMounted(() => {
  createGridCfgMaterial(true);
  if (["READY"].includes(props.item.Status)) {
    getWarehouseOnSite();
  }
});
defineExpose({
  getsPlanMaterial,
  getDataValue,
  setDataValue,
  addDataValue,
  onSelectDataLine,
  getAvailableStock,
  createGridCfgMaterial,
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
<style>
/* #tb-material .suim_table > thead > tr > th:nth-child(7) {
  background-color: #fd6e76 !important;
  color: white !important;
} */
</style>
