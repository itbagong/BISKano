<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      :form-hide-submit="false"
      grid-config="/sdp/measuringproject/gridconfig"
      form-config="/sdp/measuringproject/formconfig"
      grid-read="/sdp/measuringproject/gets"
      form-read="/sdp/measuringproject/get"
      grid-mode="grid"
      grid-delete="/sdp/measuringproject/delete"
      form-keep-label
      form-insert="/sdp/measuringproject/save"
      form-update="/sdp/measuringproject/save"
      :grid-fields="['CustomerID']"
      :form-fields="[
        'Dimension',
        'StartPeriodMonth',
        'EndPeriodMonth',
        'ProjectPeriod',
        'CustomerID',
        'SoRefNo',
      ]"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :form-tabs-new="['General']"
      :form-tabs-edit="['General', 'Lines']"
      :formInitialTab="data.formInitialTab"
      form-default-mode="edit"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @preSave="onPreSave"
      @controlModeChanged="onControlModeChanged"
      @alterGridConfig="onAlterGridConfig"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <!-- slot tab -->
      <template #form_tab_Lines="{ item }">
        <MeasuringProjectLine
          v-model="item.Lines"
          :item="item"
          grid-config="/sdp/measuringproject/line/gridconfig"
          form-config="/sdp/measuringproject/line/formconfig"
          :gridCfg="data.gridCfg"
          :listYears="data.listYears"
        ></MeasuringProjectLine>
      </template>

      <!-- slot input form -->
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
      <template #form_input_StartPeriodMonth="{ item }">
        <s-input
          ref="refStartPeriodMonth"
          v-model="item.StartPeriodMonthView"
          label="Start Period Month"
          class="w-50"
          kind="month"
          @change="
            (field, v1, v2, old, ctlRef) => {
              item.StartPeriodMonth = moment(v1).format().toString();
              item.EndPeriodMonth = moment(v1)
                .add(item.ProjectPeriod, 'month')
                .format();
              item.EndPeriodMonthView = moment(item.EndPeriodMonth).format(
                'YYYY-MM'
              );
              // console.log(v1, item.StartPeriodMonth, item.EndPeriodMonth)
            }
          "
        ></s-input>
      </template>
      <template #form_input_EndPeriodMonth="{ item }">
        <s-input
          ref="refEndPeriodMonth"
          v-model="item.EndPeriodMonthView"
          label="End Period Month"
          class="w-50"
          kind="month"
          :disabled="true"
          read-only
        ></s-input>
      </template>
      <template #form_input_ProjectPeriod="{ item }">
        <s-input
          ref="refProjectPeriod"
          v-model="item.ProjectPeriod"
          label="Project Period"
          class="w-50"
          kind="number"
        ></s-input>
      </template>
      <template #form_input_CustomerID="{ item }">
        <s-input
          :key="data.CustomerID"
          v-model="item.CustomerID"
          label="Customer Name"
          class="w-50"
          disabled
          use-list
          :lookup-url="`/tenant/customer/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_input_SoRefNo="{ item }">
        <s-input
          ref="refSoRefNo"
          v-model="item.SoRefNo"
          label="SO Ref No."
          class="w-50"
          use-list
          :lookup-url="`/sdp/salesorder/find`"
          lookup-key="SalesOrderNo"
          :lookup-labels="['SalesOrderNo', 'Name']"
          :lookup-searchs="['SalesOrderNo', 'Name']"
          @change="
            (field, v1, v2, old, ctlRef) => {
              selectSO(v1, v2, item);
            }
          "
        ></s-input>
      </template>

      <!-- slot grid -->
      <template #grid_CustomerID="{ item }">
        <s-input
          ref="refCustomerIDGrid"
          v-model="item.CustomerID"
          class="w-50"
          :disabled="true"
          read-only
          use-list
          :lookup-url="`/tenant/customer/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>

      <!-- <template #form_buttons_2="{ item }">
        <div class="flex gap-[2px] ml-2">
          <s-button
            class="btn_primary"
            label="Save & Submit"
            @click="postSubmit(item)"
          />
          <s-button
            class="btn_primary"
            label="Process"
            @click="postProcess(item)"
          />
          <s-button
            class="btn_primary"
            label="Approve"
            @click="postApprove(item)"
          />
        </div>
      </template> -->
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, computed, watch, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  createFormConfig,
  DataList,
  util,
  SInput,
  SButton,
  SCard,
  SForm,
} from "suimjs";
import { authStore } from "@/stores/auth.js";
import MeasuringProjectLine from "./widget/MeasuringProjectLine.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import moment from "moment";

layoutStore().name = "tenant";

const featureID = "SalesMeasuringProject";
// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})
const profile = authStore().getRBAC(featureID);

const listControl = ref(null);
const lineConfig = ref(null);
const axios = inject("axios");
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
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Measuring Project",
  formInitialTab: 0,
  record: {
    _id: "",
    StartPeriodMonth: new Date(),
    EndPeriodMonth: new Date(),
  },
  stayOnForm: true,
  gridCfg: {},
  listYears: [],
  listSalesOrder: [],
  CustomerID: "",
});

function newRecord(record) {
  record._id = "";
  record.StartPeriodMonth = new Date();
  record.EndPeriodMonth = new Date();
  record.CustomerID = "";
  record.ProjectPeriod = 0;
  data.formMode = "new";
  data.titleForm = "Create New Measuring Project";
  openForm(record);
}

function editRecord(record) {
  record.StartPeriodMonthView = moment(record.StartPeriodMonth).format(
    "YYYY-MM"
  );
  record.EndPeriodMonthView = moment(record.EndPeriodMonth).format("YYYY-MM");

  data.formMode = "edit";
  data.record = record;
  // console.log(record)
  data.titleForm = `Edit Measuring Project | ${record.ProjectName}`;
  getDataYear(record);
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}

function onPreSave(record) {}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Measuring Project";
  }
}

function mapingLineFields() {
  let config = {
    setting: {
      idField: "",
      keywordFields: ["_id", "Name"],
      sortable: ["_id"],
    },
    fields: [],
  };
  const month = [
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December",
  ];
  let newFields = [
    {
      field: "LedgerAccount",
      kind: "Text",
      label: "Ledger Account",
      readType: "show",
      input: {
        field: "LedgerAccount",
        label: "LedgerAccount",
        hint: "",
        hide: false,
        placeHolder: "LedgerAccount",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    },
  ];
  month.forEach((e) => {
    newFields.push({
      field: e,
      kind: "Text",
      label: e,
      readType: "show",
      input: {
        field: e,
        label: e,
        hint: "",
        hide: false,
        placeHolder: e,
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    });
  });
  config.fields = newFields;
  data.gridCfg = config;
}

function getDataYear(record) {
  var startYear = new Date(record.StartPeriodMonth).getFullYear();
  var endYear = new Date(record.EndPeriodMonth).getFullYear();
  var years = [];
  while (startYear <= endYear) {
    years.push(startYear++);
  }
  years.sort(function (a, b) {
    return a - b;
  });

  data.listYears = years;
}

async function selectSO(v1, v2, item) {
  let SO = {};
  const salesOrder = await getSalesOrder(v1);
  salesOrder.forEach((e) => {
    if (e.SalesOrderNo == item.SoRefNo) {
      SO = e;
    }
  });
  // console.log("item", item)
  // console.log("listSalesOrder", data.listSalesOrder)
  // console.log("SO", SO)
  const ContractPeriod = SO.Lines?.length > 0 ? SO.Lines[0].ContractPeriod : 0;
  const ProjectPeriod =
    SO.Lines?.length > 0 && SO.Lines[0].UoM == "DAYS"
      ? 1
      : SO.Lines.length > 0 && SO.Lines[0].UoM == "MONTH"
      ? ContractPeriod
      : 0;
  // console.log("res", ContractPeriod, ProjectPeriod)
  let record = data.record;
  record.CustomerID = SO.CustomerID;
  record.ProjectPeriod = ProjectPeriod;
  record.StartPeriodMonth = moment(new Date()).format().toString();
  record.StartPeriodMonthView = moment(new Date()).format("YYYY-MM");
  record.EndPeriodMonth = moment(new Date())
    .add(record.ProjectPeriod, "months")
    .format();
  record.EndPeriodMonthView = moment(record.EndPeriodMonth).format("YYYY-MM");

  item.CustomerID = SO.CustomerID;
  // item.CustomerName = "";
  item.ProjectPeriod = ProjectPeriod;
  item.StartPeriodMonth = moment(new Date()).format().toString();
  item.StartPeriodMonthView = moment(new Date()).format("YYYY-MM");
  item.EndPeriodMonth = moment(new Date())
    .add(item.ProjectPeriod, "months")
    .format();
  item.EndPeriodMonthView = moment(item.EndPeriodMonth).format("YYYY-MM");

  data.record = record;
  data.CustomerID = SO.CustomerID;
  // console.log(item)
}

async function getSalesOrder(SoRefNo) {
  try {
    const dataresponse = await axios.post(
      `/sdp/salesorder/find?SalesOrderNo=${SoRefNo}`
    );
    const resData = dataresponse.data;
    // data.listSalesOrder = resData;
    return resData;
  } catch (error) {
    util.showError(error);
  }
}
function onAlterGridConfig(cfg) {
  cfg.setting.keywordFields = ["_id", "ProjectName", "ProjectAlias"];
}
// watch(
//   () => data.record,
//   (nv) => {
//     console.log(nv)
//   },
//   { deep: true }
// );

onMounted(() => {
  mapingLineFields();
  // getSalesOrder();
});
</script>
