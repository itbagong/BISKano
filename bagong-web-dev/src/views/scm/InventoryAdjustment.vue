<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Inventory Adjustment"
      form-hide-submit
      grid-config="/scm/inventoryadjustment/gridconfig"
      form-config="/scm/inventoryadjustment/formconfig"
      grid-read="/scm/inventoryadjustment/gets"
      form-read="/scm/inventoryadjustment/get"
      grid-mode="grid"
      grid-delete="/scm/inventoryadjustment/delete"
      form-keep-label
      form-insert="/scm/inventoryadjustment/draft"
      form-update="/scm/inventoryadjustment/draft"
      :grid-fields="['WarehouseName', 'AisleName', 'SectionName', 'BoxName']"
      :form-fields="[
        '_id',
        'StockOpnameID',
        'AdjustmentDate',
        'JournalType',
        'Company',
        'Note',
        'Status',
        'InputDate',
        'Dimension',
        'InventDim',
      ]"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :form-tabs-new="['General', 'Line']"
      :form-tabs-edit="['General', 'Line']"
      :form-tabs-view="['General', 'Line']"
      :formInitialTab="data.formInitialTab"
      form-default-mode="view"
      :grid-custom-filter="customFilter"
      @formEditData="editRecord"
      @gridRefreshed="onCancelForm"
      @alterGridConfig="alterGridConfig"
      :grid-hide-new="true"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #grid_header_search="{ config }">
        <s-input
          ref="refName"
          v-model="data.search.Name"
          lookup-key="_id"
          label="Text"
          class="w-full"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Adj Date From"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Adj Date To"
          v-model="data.search.DateTo"
          @change="refreshData"
        ></s-input>
        <!-- <s-input
          ref="refStatus"
          v-model="data.search.Status"
          lookup-key="_id"
          label="Status"
          class="w-[300px]"
          use-list
          :items="['NeedToReview', 'Done']"
          @change="refreshData"
        ></s-input> -->
        <s-input
          ref="refSite"
          v-model="data.search.Site"
          lookup-key="_id"
          label="Site"
          class="w-[450px]"
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
          @change="refreshData"
        ></s-input>
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
      <template #form_tab_Line="{ item }">
        <InventoryAdjustmentLine
          ref="lineConfig"
          v-model="data.record"
          :item="item"
          :itemID="item._id"
          :readOnly="data.formMode == 'edit'"
          :form-mode="data.formMode"
          :hide-detail="true"
          grid-config="/scm/inventoryadjustment/detail/gridconfig"
          :grid-read="
            '/scm/inventoryadjustment/detail/gets?InventoryAdjustmentID=' +
            item._id
          "
        ></InventoryAdjustmentLine>
      </template>
      <template #form_input_JournalType="{ item }">
        <s-input
          ref="refJournalType"
          label="Journal Type"
          v-model="item.JournalType"
          class="w-full"
          :required="true"
          :disabled="!['NeedToReview'].includes(item.Status)"
          :keepErrorSection="true"
          use-list
          :lookup-url="`/scm/inventorytransactionjournaltype/find?TransactionType=Inventory_Adjustment`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_input_Note="{ item }">
        <s-input
          ref="refNotes"
          label="Notes"
          v-model="item.Note"
          class="w-full"
          multiRow="5"
          :disabled="!['NeedToReview'].includes(item.Status)"
          :keepErrorSection="true"
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <div>
          <dimension-editor
            v-model="item.Dimension"
            sectionTitle="Financial Dimension"
            :default-list="profile.Dimension"
            :readOnly="true"
          ></dimension-editor>
        </div>
      </template>
      <template #form_input_InventDim="{ item }">
        <div>
          <dimension-invent-jurnal
            ref="RefDimensionInventory"
            v-model="item.InventDim"
            :disabled="true"
            :defaultList="profile.Dimension"
            :site="
              item.Dimension &&
              item.Dimension.find((_dim) => _dim.Key === 'Site') &&
              item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] != ''
                ? item.Dimension.find((_dim) => _dim.Key === 'Site')['Value']
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
          ></dimension-invent-jurnal>
        </div>
      </template>
      <template #form_buttons_2="{ item }">
        <div class="flex gap-[2px] ml-2">
          <s-button
            v-show="item.Status == 'NeedToReview'"
            class="btn_primary"
            label="Process"
            :disabled="data.isBtnProcess"
            @click="postProcess(item)"
          />
        </div>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, computed, onMounted, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";
import { useRouter, useRoute } from "vue-router";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import InventoryAdjustmentLine from "./widget/InventoryAdjustmentLine.vue";
import { authStore } from "@/stores/auth.js";
import moment from "moment";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";

const featureID = "InventoryAdjustment";
const route = useRoute();
const router = useRouter();
// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'InventoryAdjustment'})

const profile = authStore().getRBAC(featureID);
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const listControl = ref(null);
const lineConfig = ref(null);
const refJournalType = ref(null);

const axios = inject("axios");
const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];

let customFilter = computed(() => {
  const filters = [];
  if (data.search.Name !== null && data.search.Name !== "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "_id",
          Op: "$contains",
          Value: [data.search.Name],
        },
        {
          Field: "StockOpnameID",
          Op: "$contains",
          Value: [data.search.Name],
        },
      ],
    });
  }
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "AdjustmentDate",
      Op: "$gte",
      Value: moment(data.search.DateFrom).utc().format("YYYY-MM-DDT00:mm:00Z"),
    });
  }
  if (
    data.search.DateTo !== null &&
    data.search.DateTo !== "" &&
    data.search.DateTo !== "Invalid date"
  ) {
    filters.push({
      Field: "AdjustmentDate",
      Op: "$lte",
      Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
    });
  }
  if (data.search.Status !== null && data.search.Status !== "") {
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: data.search.Status,
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

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Inventory Adjustment",
  formInitialTab: 0,
  record: {
    _id: "",
    TrxDate: new Date(),
    Status: "",
  },
  search: {
    Name: "",
    DateFrom: null,
    DateTo: null,
    Status: "",
    Site: "",
  },
  stayOnForm: true,
  isBtnProcess: false,
  inventoryDimension: [
    { key: "warehouse", value: "" },
    { key: "aisle", value: "" },
    { key: "section", value: "" },
    { key: "box", value: "" },
  ],
});

function editRecord(record) {
  data.formMode = "edit";
  data.record = record;
  data.titleForm = `Edit Inventory Adjustment | ${record._id}`;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}

function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}

function alterGridConfig(cfg) {
  cfg.sortable = ["Created", "_id"];
  cfg.setting.idField = "Created";
  cfg.setting.sortable = ["Created", "_id"];
  const newFields = [
    {
      field: "WarehouseName",
      kind: "Text",
      label: "Warehouse Name",
      readType: "show",
      input: {
        field: "WarehouseName",
        label: "Warehouse Name",
        hint: "",
        hide: false,
        placeHolder: "Warehouse Name",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    },
    {
      field: "AisleName",
      kind: "Text",
      label: "Aisle Name",
      readType: "show",
      input: {
        field: "AisleName",
        label: "Aisle Name",
        hint: "",
        hide: false,
        placeHolder: "Aisle Name",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    },
    {
      field: "SectionName",
      kind: "Text",
      label: "Section Name",
      readType: "show",
      input: {
        field: "SectionName",
        label: "Section Name",
        hint: "",
        hide: false,
        placeHolder: "Section Name",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    },
    {
      field: "BoxName",
      kind: "Text",
      label: "Box Name",
      readType: "show",
      input: {
        field: "BoxName",
        label: "Box Name",
        hint: "",
        hide: false,
        placeHolder: "Box Name",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    },
  ];
  cfg.fields = [...cfg.fields, ...newFields];
  console.log(cfg);
}
function postProcess(record) {
  // if (!refJournalType?.value?.isValid() && !record.JournalType) {
  //   return util.showError("Journal type is required");
  // }
  let payload = {
    ID: record._id,
  };
  let payloadSave = record;
  let payloadLine = {
    InventoryAdjustmentID: record._id,
    InventoryAdjustmentDetails: [],
  };

  if (lineConfig.value && lineConfig.value.getDataValue()) {
    let dv = lineConfig.value.getDataValue();
    let WarehouseID = "";
    if (record.InventoryDimension) {
      WarehouseID = record.InventoryDimension.WarehouseID;
    }
    dv.map(function (d) {
      d.InventoryDimension = {
        WarehouseID: WarehouseID,
        AisleID: d.AisleID,
        SectionID: d.SectionID,
        BoxID: d.BoxID,
      };
      return d;
    });
    payload.Lines = dv;
    payloadSave.Lines = dv;
    payloadLine.InventoryAdjustmentDetails = dv;
  }
  data.isBtnProcess = true;
  listControl.value.setFormLoading(true);
  axios.post("/scm/inventoryadjustment/save", payloadSave).then(
    (r) => {
      axios
        .post("/scm/inventoryadjustment/detail/save-multiple", payloadLine)
        .then(
          (r) => {
            axios.post("/scm/inventoryadjustment/process", payload).then(
              (r) => {
                data.stayOnForm = false;
                data.isBtnProcess = false;
                listControl.value.setFormLoading(false);
                listControl.value.refreshForm();
                listControl.value.setControlMode("grid");
                listControl.value.refreshList();
              },
              (e) => {
                data.isBtnProcess = false;
                listControl.value.setFormLoading(false);
                util.showError(e);
              }
            );
          },
          (e) => {
            data.isBtnProcess = false;
            listControl.value.setFormLoading(false);
            util.showError(e);
          }
        );
    },
    (e) => {
      data.isBtnProcess = false;
      util.showError(e);
    }
  );
}
function onCancelForm(mode) {
  if (mode === "grid") {
    data.titleForm = "Inventory Adjustment";
  }
}
onMounted(() => {
  if (defaultList.length === 1) {
    data.search.Site = defaultList[0];
  }

  if (route.query.trxid !== undefined || route.query.id !== undefined) {
    let getUrlParam = route.query.trxid || route.query.id;
    axios
      .post(`/scm/inventoryadjustment/get`, [getUrlParam])
      .then(
        (r) => {
          listControl.value.setControlMode("form");
          listControl.value.setFormRecord(r.data);

          data.record = r.data;
        },
        (e) => util.showError(e)
      )
      .finally(() => {
        util.nextTickN(2, () => {
          router.replace({
            path: `/scm/InventoryAdjustment`,
          });
        });
      });
  }
});
</script>
