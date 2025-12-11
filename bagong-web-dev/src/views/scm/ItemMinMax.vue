<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/scm/item/min-max/gridconfig"
      form-config="/scm/item/min-max/formconfig"
      grid-read="/scm/item/min-max/gets"
      form-read="/scm/item/min-max/get"
      grid-mode="grid"
      grid-delete="/scm/item/min-max/delete"
      form-keep-label
      form-insert="/scm/item/min-max/save"
      form-update="/scm/item/min-max/save"
      form-hide-submit
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :grid-fields="[
        'Enable',
        'ItemID',
        'FinancialDimension',
        'WarehouseName',
        'AisleName',
        'SectionName',
        'BoxName',
      ]"
      :form-fields="[
        'ItemID',
        'SKU',
        'FinancialDimension',
        'InventoryDimension',
      ]"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :grid-custom-filter="customFilter"
      @controlModeChanged="onControlModeChanged"
      @formNewData="newRecord"
      @formEditData="editRecord"
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
          ref="refSite"
          v-model="data.search.Site"
          lookup-key="_id"
          label="Site"
          class="w-full"
          use-list
          :disabled="defaultList?.length === 1"
          :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
          :lookup-labels="['Label']"
          :lookup-searchs="['_id', 'Label']"
          :lookup-payload-builder="
            defaultList?.length > 0
              ? (...args) =>
                  helper.payloadBuilderDimension(
                    defaultList,
                    data.search.Site,
                    false,
                    ...args
                  )
              : undefined
          "
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
      <template #grid_FinancialDimension="{ item }">
        <DimensionText :dimension="item.FinancialDimension" />
      </template>
      <template #grid_ItemID="{ item }">
        {{ item.ItemName }}
      </template>
      <template #grid_WarehouseName="{ item }">
        {{ item.WarehouseName }}
      </template>
      <template #grid_AisleName="{ item }">
        {{ item.AisleName }}
      </template>
      <template #grid_SectionName="{ item }">
        {{ item.SectionName }}
      </template>
      <template #grid_BoxName="{ item }">
        {{ item.BoxName }}
      </template>
      <template #form_input_FinancialDimension="{ item }">
        <dimension-editor
          v-model="item.FinancialDimension"
          :readOnly="data.formMode == 'view'"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
      <template #form_input_InventoryDimension="{ item }">
        <dimension-invent-jurnal
          ref="RefDimensionInventory"
          v-model="item.InventoryDimension"
          :defaultList="profile.Dimension"
          :site="
            item.FinancialDimension &&
            item.FinancialDimension.find((_dim) => _dim.Key === 'Site') &&
            item.FinancialDimension.find((_dim) => _dim.Key === 'Site')[
              'Value'
            ] != ''
              ? item.FinancialDimension.find((_dim) => _dim.Key === 'Site')[
                  'Value'
                ]
              : undefined
          "
          title-header="Inventory Dimension"
          :hide-field="[
            'BatchID',
            'SerialNumber',
            'SpecID',
            'InventDimID',
            'VariantID',
            'Size',
            'Grade',
          ]"
          :mandatory="['WarehouseID']"
          @onFieldChanged="(v1, v2, _item) => {}"
        ></dimension-invent-jurnal>
      </template>
      <template #form_input_ItemID="{ item }">
        <s-input-sku-item
          ref="refItemVarian"
          label="Item Varian"
          v-model="item.ItemVarian"
          :record="item"
          :required="true"
          :keepErrorSection="true"
          :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
            item.ItemID,
            item.SKU
          )}`"
        ></s-input-sku-item>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          :disabled="data.loading"
          :icon="`content-save`"
          class="btn_primary submit_btn"
          label="Save"
          @click="onSave(item)"
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, computed, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
import { DataList, util, SInput, SButton } from "suimjs";
import helper from "@/scripts/helper.js";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import SInputSkuItem from "./widget/SInputSkuItem.vue";
import DimensionText from "@/components/common/DimensionText.vue";

layoutStore().name = "tenant";
const featureID = "ItemMinMax";
const profile = authStore().getRBAC(featureID);
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const listControl = ref(null);
const refItemVarian = ref(null);
const RefDimensionInventory = ref(null);
const axios = inject("axios");
const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];

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

  if (
    data.search.Site !== undefined &&
    data.search.Site !== null &&
    data.search.Site !== ""
  ) {
    filters.push(
      {
        Field: "Dimension.Key",
        Op: "$eq",
        Value: "Site",
      },
      {
        Field: "Dimension.Value",
        Op: "$eq",
        Value: data.search.Site,
      }
    );
  }

  if (data.search.WarehouseID !== null && data.search.WarehouseID !== "") {
    filters.push({
      Field: "InventoryDimension.WarehouseID",
      Op: "$eq",
      Value: data.search.WarehouseID,
    });
  }

  if (data.search.AisleID !== null && data.search.AisleID !== "") {
    filters.push({
      Field: "InventoryDimension.AisleID",
      Op: "$eq",
      Value: data.search.AisleID,
    });
  }

  if (data.search.SectionID !== null && data.search.SectionID !== "") {
    filters.push({
      Field: "InventoryDimension.SectionID",
      Op: "$eq",
      Value: data.search.SectionID,
    });
  }

  if (data.search.BoxID !== null && data.search.BoxID !== "") {
    filters.push({
      Field: "InventoryDimension.BoxID",
      Op: "$eq",
      Value: data.search.BoxID,
    });
  }
  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Item Min Max",
  loading: false,
  record: {
    _id: "",
  },
  search: {
    ItemIDs: "",
    SKU: "",
    Site: "",
    WarehouseID: "",
    AisleID: "",
    SectionID: "",
    BoxID: "",
  },
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Item Min Max`;
  record._id = "";
  record.ItemVarian = "";
  openForm(record);
}

function editRecord(record) {
  let ItemVarian = "";
  if (record.ItemID) {
    ItemVarian = helper.ItemVarian(record.ItemID, record.SKU);
  }
  record.ItemVarian = ItemVarian;
  data.formMode = "edit";
  data.titleForm = `Edit Item Min Max | ${record._id}`;
  data.record = record;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
    document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(1)"
    ).style.width = "33.33%";
    document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(2)"
    ).style.width = "33.33%";
    document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(3)"
    ).style.width = "33.33%";
  });
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
function onSave(record) {
  let valid = true;
  data.loading = true;
  const validInvDin = RefDimensionInventory.value.validate();
  if (
    !validInvDin ||
    !refItemVarian.value.validate() ||
    !listControl.value.formValidate()
  ) {
    valid = false;
  }
  if (valid) {
    listControl.value.submitForm(
      record,
      () => {
        data.loading = false;
      },
      (e) => {
        data.loading = false;
      }
    );
  } else {
    data.loading = false;
    return util.showError("field required");
  }
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
  let querySite = [
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

  if (data.search.Site) {
    let querySite = [
      {
        Field: "Dimension.Key",
        Op: "$eq",
        Value: "Site",
      },
      {
        Field: "Dimension.Value",
        Op: "$eq",
        Value: data.search.Site,
      },
    ];
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

    if (data.search.Site) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}
function alterGridConfig(cfg) {
  cfg.sortable = ["LastUpdate", "Created", "_id"];
  cfg.setting.idField = "LastUpdate";
  cfg.setting.sortable = ["LastUpdate", "Created", "_id"];

  const hideField = ["FinancialDimension"];
  const Dimension = ["WarehouseName", "AisleName", "SectionName", "BoxName"];
  let colmLine = [
    "_id",
    "ItemID",
    "MinStock",
    "MaxStock",
    "SafeStock",
    "WarehouseName",
    "AisleName",
    "SectionName",
    "BoxName",
  ];
  for (let index = 0; index < Dimension.length; index++) {
    cfg.fields.push({
      field: Dimension[index],
      kind: "Text",
      label: Dimension[index].replace(/([a-z])([A-Z])/g, "$1 $2"),
      readType: "show",
      input: {
        field: Dimension[index],
        label: Dimension[index].replace(/([a-z])([A-Z])/g, "$1 $2"),
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
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Item Min Max";
  }
}
onMounted(() => {
  if (defaultList.length === 1) {
    data.search.Site = defaultList[0];
  }
});
</script>
