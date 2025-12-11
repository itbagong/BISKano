<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      form-hide-submit
      :title="data.titleForm"
      :grid-hide-new="true"
      :grid-hide-select="true"
      :grid-hide-detail="true"
      :grid-hide-delete="true"
      grid-config="/scm/item/balance/gridconfig"
      form-config=""
      grid-read="/scm/item/balance/gets"
      form-read="/scm/item/balance/get"
      grid-mode="grid"
      grid-delete="/scm/item/balance/delete"
      form-keep-label
      form-insert="/scm/item/balance/save"
      form-update="/scm/item/balance/save"
      :grid-fields="[
        'Enable',
        'ItemID',
        'SKU',
        'WarehouseID',
        'AisleID',
        'SectionID',
        'BoxID',
        'Qty',
        'QtyReserved',
        'QtyPlanned',
        'QtyAvail',
        'AmountPhysical',
        'AmountAdjustment',
        'AmountFinancial',
      ]"
      :form-fields="['InventDim', 'SKU']"
      form-default-mode="view"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :grid-custom-filter="customFilter"
      @alterGridConfig="alterGridConfig"
    >
      <template #grid_header_search="{ config }">
        <s-input
          ref="refItemID"
          v-model="data.search.ItemIDs"
          lookup-key="_id"
          label="Item"
          class="w-full"
          use-list
          :lookup-url="`/tenant/item/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
        <s-input
          ref="refSKU"
          v-model="data.search.SKU"
          label="SKU"
          class="w-[50%]"
          use-list
          :lookup-url="`/tenant/itemspec/gets-info?ItemID=${data.search.ItemIDs}`"
          lookup-key="_id"
          :lookup-labels="['Description']"
          :lookup-searchs="['_id', 'SKU', 'Description']"
        ></s-input>
        <s-input
          ref="refwarehouse"
          v-model="data.search.WarehouseID"
          lookup-key="_id"
          label="Warehouse"
          class="w-[50%]"
          use-list
          :lookup-url="`/tenant/warehouse/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilder(
                search,
                ['_id', 'Name'],
                data.search.WarehouseID,
                item
              )
          "
        ></s-input>
        <s-input
          ref="refAislle"
          v-model="data.search.AisleID"
          lookup-key="_id"
          label="Aisle"
          class="w-[50%]"
          use-list
          :lookup-url="`/tenant/aisle/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
        <s-input
          ref="refSection"
          v-model="data.search.SectionID"
          lookup-key="_id"
          label="Section"
          class="w-[50%]"
          use-list
          :lookup-url="`/tenant/section/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
        <s-input
          ref="refBoxID"
          v-model="data.search.BoxID"
          lookup-key="_id"
          label="Box"
          class="w-[50%]"
          use-list
          :lookup-url="`/tenant/box/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #grid_ItemID="{ item, idx }">
        {{ item.ItemName }}
      </template>
      <template #grid_SKU="{ item }">
        {{ item.SKU }}
      </template>
      <template #grid_Qty="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.Qty) }}
        </div>
      </template>
      <template #grid_QtyReserved="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.QtyReserved) }}
        </div>
      </template>
      <template #grid_QtyPlanned="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.QtyPlanned) }}
        </div>
      </template>
      <template #grid_QtyAvail="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.QtyAvail) }}
        </div>
      </template>
      <template #grid_AmountPhysical="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.AmountPhysical) }}
        </div>
      </template>
      <template #grid_AmountFinancial="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.AmountFinancial) }}
        </div>
      </template>
      <template #grid_AmountAdjustment="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.AmountAdjustment) }}
        </div>
      </template>
      <template #grid_WarehouseID="{ item }">
        {{ item.InventDim.WarehouseID }}
      </template>
      <template #grid_AisleID="{ item }">
        {{ item.InventDim.AisleID }}
      </template>
      <template #grid_SectionID="{ item }">
        {{ item.InventDim.SectionID }}
      </template>
      <template #grid_BoxID="{ item }">
        {{ item.InventDim.BoxID }}
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, computed, watch, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
import { DataList, util, SInput } from "suimjs";
import DimensionInventory from "@/components/common/DimensionInventJurnal.vue";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";
const featureID = "ItemBalance";
const profile = authStore().getRBAC(featureID);
const listControl = ref(null);
const axios = inject("axios");
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Item Balance",
  record: {
    _id: "",
  },
  search: {
    ItemIDs: "",
    SKU: "",
    WarehouseID: "",
    AisleID: "",
    SectionID: "",
    BoxID: "",
  },
});

let customFilter = computed(() => {
  const filters = [];

  if (data.search.ItemIDs !== null && data.search.ItemIDs !== "") {
    filters.push({
      Field: "ItemID",
      Op: "$eq",
      Value: data.search.ItemIDs,
    });
  }
  if (data.search.SKU !== null && data.search.SKU !== "") {
    filters.push({
      Field: "SKU",
      Op: "$eq",
      Value: data.search.SKU,
    });
  }

  if (data.search.WarehouseID !== null && data.search.WarehouseID !== "") {
    filters.push({
      Field: "InventDim.WarehouseID",
      Op: "$eq",
      Value: data.search.WarehouseID,
    });
  }

  if (data.search.AisleID !== null && data.search.AisleID !== "") {
    filters.push({
      Field: "InventDim.AisleID",
      Op: "$eq",
      Value: data.search.AisleID,
    });
  }

  if (data.search.SectionID !== null && data.search.SectionID !== "") {
    filters.push({
      Field: "InventDim.SectionID",
      Op: "$eq",
      Value: data.search.SectionID,
    });
  }

  if (data.search.BoxID !== null && data.search.BoxID !== "") {
    filters.push({
      Field: "InventDim.BoxID",
      Op: "$eq",
      Value: data.search.BoxID,
    });
  }
  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

function alterGridConfig(cfg) {
  const hideField = ["_id", "SKU", "CompanyID"];
  const Dimension = ["WarehouseID", "AisleID", "SectionID", "BoxID"];
  let colmLine = [
    "ItemID",
    "SKU",
    "BalanceDate",
    "WarehouseID",
    "AisleID",
    "SectionID",
    "BoxID",
    "UnitName",
    "Qty",
    "QtyReserved",
    "QtyPlanned",
    "QtyAvail",
    "AmountPhysical",
    "AmountFinancial",
    "AmountAdjustment",
  ];
  for (let index = 0; index < Dimension.length; index++) {
    cfg.fields.push({
      field: Dimension[index],
      kind: "Text",
      label: Dimension[index],
      readType: "show",
      input: {
        field: Dimension[index],
        label: Dimension[index],
        hint: "",
        hide: false,
        placeHolder: Dimension[index],
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    });
  }
  cfg.fields.map(function (el) {
    if (hideField.includes(el.field)) {
      el.readType = "hide";
    }
    el.idx = colmLine.indexOf(el.field);
    return el;
  });
  cfg.fields.sort((a, b) => (a.idx > b.idx ? 1 : -1));
}

function lookupPayloadBuilder(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

  //setting search
  const Site =
    profile.Dimension &&
    profile.Dimension.find((_dim) => _dim.Key === "Site") &&
    profile.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? profile.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : undefined;
  const querySite = [
    {
      Field: "Dimension.Key",
      Op: "$eq",
      Value: "Site",
    },
    {
      Field: "Dimension.Value",
      Op: "$eq",
      Value: Site,
    },
  ];
  if (Site) {
    qp.Where = {
      Op: "$and",
      items: querySite,
    };
  }
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "_id", Op: "$contains", Value: [search] },
          { Field: "Name", Op: "$contains", Value: [search] },
        ],
      },
    ];
    if (Site) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}

function getDefaultWarehouse() {
  const site =
    profile.Dimension &&
    profile.Dimension.find((_dim) => _dim.Key === "Site") &&
    profile.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? profile.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : undefined;
  axios
    .post("/tenant/warehouse/find", {
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
            Value: site,
          },
        ],
      },
    })
    .then((r) => {
      if (r.data.length > 0) {
        data.search.WarehouseID = r.data[0]._id;
      }
    });
}
watch(
  () => data.search,
  (nv) => {
    util.nextTickN(2, () => {
      listControl.value.refreshGrid();
    });
  },
  { deep: true }
);

onMounted(() => {
  getDefaultWarehouse();
});
</script>
