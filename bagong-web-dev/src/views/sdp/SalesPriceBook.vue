<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/sdp/salespricebook/gridconfig"
      form-config="/sdp/salespricebook/formconfig"
      grid-read="/sdp/salespricebook/gets"
      form-read="/sdp/salespricebook/get"
      grid-mode="grid"
      grid-delete="/sdp/salespricebook/delete"
      form-keep-label
      form-insert="/sdp/salespricebook/insert"
      form-update="/sdp/salespricebook/update"
      :form-fields="['Name', 'StartPeriod', 'EndPeriod', 'Dimension']"
      :form-tabs-new="['General', 'Line']"
      :form-tabs-edit="['General', 'Line']"
      :form-tabs-view="['General', 'Line']"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :grid-custom-filter="customFilter"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @pre-save="preSave"
      @controlModeChanged="onCancelForm"
      @alter-form-config="onalterFormConfig"
      @alterGridConfig="onAlterGridConfig"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <!-- @grid-refreshed="gridRefreshed" -->
      <template #grid_header_search="{ config }">
        <s-input
          v-model="data.searchData.SalesPriceBookName"
          kind="text"
          label="Search Price Book Name"
          class="w-full"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          v-model="data.searchData.DateFrom"
          kind="date"
          label="Period From"
          class="w-full"
          @change="refreshData"
        ></s-input>
        <s-input
          v-model="data.searchData.DateTo"
          kind="date"
          label="Period To"
          class="w-full"
          @change="refreshData"
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
      <template #form_tab_Line="{ item, config }">
        <SalesPriceBookLine
          ref="lineConfig"
          v-model="item.Lines"
          :dimension="item.Dimension"
          :itemID="item._id"
          :items="item.Lines"
          :form-mode="data.formMode"
        ></SalesPriceBookLine>
      </template>
    </data-list>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, inject, computed, onMounted } from "vue";
import { layoutStore } from "../../stores/layout.js";
import { authStore } from "@/stores/auth.js";
import moment from "moment";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import {
  DataList,
  util,
  SForm,
  SInput,
  createFormConfig,
  SButton,
} from "suimjs";
import SalesPriceBookLine from "./widget/SalesPriceBookLine.vue";
import { useRoute } from "vue-router";

layoutStore().name = "tenant";

const featureID = "SalesPriceBook";
// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})
const profile = authStore().getRBAC(featureID);

const listControl = ref(null as any);
const lineConfig = ref(null);
const axios = inject("axios");
const route = useRoute();

const roleID = [
  (v) => {
    let vLen = 0;
    let consistsInvalidChar = false;

    v.split("").forEach((ch) => {
      vLen++;
      const validCar =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
          ch
        ) >= 0;
      if (!validCar) consistsInvalidChar = true;
    });

    if (vLen < 3 || consistsInvalidChar)
      return "minimal length is 3 and alphabet only";
    return "";
  },
];
let customFilter = computed(() => {
  const filters = [];
  if (
    data.searchData.SalesPriceBookName !== null &&
    data.searchData.SalesPriceBookName !== ""
  ) {
    filters.push({
      Field: "Name",
      Op: "$contains",
      Value: [data.searchData.SalesPriceBookName],
    });
  }
  if (
    data.searchData.DateFrom !== null &&
    data.searchData.DateFrom !== "" &&
    data.searchData.DateTo !== "Invalid date"
  ) {
    filters.push({
      Field: "StartPeriod",
      Op: "$gte",
      Value: moment(data.searchData.DateFrom)
        .utc()
        .format("YYYY-MM-DDT00:mm:00Z"),
    });
  }
  if (
    data.searchData.DateTo !== null &&
    data.searchData.DateTo !== "" &&
    data.searchData.DateTo !== "Invalid date"
  ) {
    filters.push({
      Field: "EndPeriod",
      Op: "$lte",
      Value: moment(data.searchData.DateTo)
        .utc()
        .format("YYYY-MM-DDT23:59:00Z"),
    });
  }
  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});
const data = reactive({
  title: null as any,
  appMode: "grid",
  formMode: "edit",
  titleForm: "Sales Price Book",
  record: [],
  searchData: {
    SalesPriceBookName: "",
    DateFrom: "",
    DateTo: "",
  },
  allowDelete: route.query.allowdelete === "true",
  formAssets: {},
  isSelected: false,
});

watch(
  () => route.query.objname,
  (nv) => {
    util.nextTickN(2, () => {
      listControl.value.refreshList();
      listControl.value.refreshForm();
    });
  }
);

watch(
  () => route.query.title,
  (nv) => {
    data.title = nv;
    listControl.value.setControlMode("grid");
  }
);

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}

function preSave(record) {
  if (record.StartPeriod == null) {
    record.StartPeriod = moment().format("YYYY-MM-DDThh:mm:ssZ");
  }

  if (record.EndPeriod == null) {
    record.EndPeriod = moment().format("YYYY-MM-DDThh:mm:ssZ");
  }

  if (record.Lines.length > 0) {
    record.Lines.forEach((item, idx) => {
      delete record.Lines[idx].Idx;
      if (record.Lines[idx].ProductionYear == "") {
        record.Lines[idx].ProductionYear = 0;
      }

      if (record.Lines[idx].MinPrice > record.Lines[idx].MaxPrice) {
        util.showError("Minumum Price can't more higher than Maximum Price");
        return true;
      }
    });
  }
}

function newRecord() {
  data.titleForm = "Create New Sales Price Book";
  openForm();
}

function editRecord(record) {
  data.titleForm = `Edit Sales Price Book | ${record._id}`;
  openForm();
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
function gridRefreshed() {
  if (
    data.searchData.SalesPriceBookName != "" ||
    data.searchData.DateFrom != "" ||
    data.searchData.DateTo != ""
  ) {
    axios.post("/sdp/salespricebook/gets-filter", data.searchData).then(
      async (r) => {
        listControl.value.setGridRecords(r.data.data);
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function onPostSave(record) {
  // let lines = lineConfig.value.getDataValue();
  // let payloadBatch = {
  //   SalesPriceBookID: record._id,
  //   Lines: lines.filter(function (b) {
  //     return b.SalesPriceBookID != "";
  //   }),
  // };
  // console.log("payloadBatch =>", lines)
  // axios
  //   .post("/sdp/salespricebookline/save-multiple", payloadBatch)
  //   .then(
  //     (r) => {
  //       console.log("rrr =>", r)
  //     },
  //     (e) => {}
  //   );
}

function onCancelForm(mode) {
  if (mode === "grid") {
    data.titleForm = "Sales Price Book";
  }
}

const addCancel = () => {
  data.formMode = "new";
  // record._id = "";
  // record.TrxDate = new Date();
  // record.Status = "";
  data.titleForm = "Create New Sales Price Book";
  // openForm(record);
};

const onsubmit = () => {};

function onalterFormConfig(r) {}
function onAlterGridConfig(cfg: any) {
  cfg.setting.idField = "Created";
  cfg.setting.sortable = ["_id", "Created"];
}
</script>
