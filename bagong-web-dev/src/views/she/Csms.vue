<template>
  <data-list
    class="card w-full"
    ref="listControl"
    title="CSMS"
    grid-config="/she/csms/gridconfig"
    form-config="/she/csms/formconfig"
    grid-read="/she/csms/gets"
    form-read="/she/csms/get"
    grid-mode="grid"
    grid-delete="/she/csms/delete"
    form-insert="/she/csms/save"
    form-update="/she/csms/save"
    :init-app-mode="data.appMode"
    :form-tabs-edit="['General', 'Line']"
    :form-tabs-view="['General', 'Line']"
    :form-tabs-new="['General', 'Line']"
    :form-fields="['Dimension', 'Status']"
    :grid-fields="['Dimension', 'Status', 'CSMSStatus']"
    :grid-custom-filter="customFilter"
    @formFieldChange="onFormFieldChange"
    @formEditData="editRecord"
    @formNewData="newData"
    @alterGridConfig="alterGridConfig"
    @preSave="onPreSave"
    form-keep-label
    stay-on-form-after-save
  >
    <template #grid_header_search="{ config }">
      <div class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter">
        <s-input
          kind="date"
          label="Date From"
          class="w-[200px]"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Date To"
          class="w-[200px]"
          v-model="data.search.DateTo"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refName"
          v-model="data.search.No"
          class="w-[300px]"
          label="No"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refCustomer"
          v-model="data.search.Customer"
          class="w-[300px]"
          use-list
          label="Customer"
          lookup-url="/tenant/customer/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refTemplate"
          v-model="data.search.Template"
          class="w-[300px]"
          use-list
          label="Template"
          lookup-url="/she/mcuitemtemplate/find?Menu=SHE-0010"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refStatus"
          v-model="data.search.Status"
          lookup-key="_id"
          label="Status"
          class="w-[200px]"
          use-list
          :items="['DRAFT', 'SUBMITTED', 'READY', 'POSTED', 'REJECTED']"
          @change="refreshData"
        ></s-input>
      </div>
    </template>
    <template #form_buttons_1="{ item, inSubmission, loading }">
      <form-buttons-trx
        :disabled="loading"
        :status="item.Status"
        :journal-id="item._id"
        :posting-profile-id="item.PostingProfileID"
        :journal-type-id="'CSMS'"
        :moduleid="'she'"
        @preSubmit="trxPreSubmit"
        @postSubmit="trxPostSubmit"
        @errorSubmit="trxErrorSubmit"
        :auto-post="!waitTrxSubmit"
      />
    </template>
    <template #form_input_Dimension="{ item, mode }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :default-list="profile.Dimension"
        :read-only="mode == 'view'"
      ></dimension-editor-vertical>
    </template>
    <template #form_tab_Line="{ item }">
      <lines
        v-model="item.Lines"
        :template-lines="item.TemplateID"
        :line-cfg="data.lineCfg"
        kind="csms"
      />
    </template>
    <template #grid_CSMSStatus="{ item }">
      <status-text :txt="item.CSMSStatus" />
    </template>
    <template #grid_Status="{ item }">
      <status-text :txt="item.Status" />
    </template>
    <template #grid_Dimension="{ item }">
      <DimensionText :dimension="item.Dimension" />
    </template>
  </data-list>
</template>
<script setup>
import { reactive, ref, watch, inject, onMounted, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import StatusText from "@/components/common/StatusText.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import moment from "moment";
import Pica from "@/components/common/ItemPica.vue";
import { authStore } from "@/stores/auth.js";
import Lines from "./widget/LinesInspectionCsms.vue";

const axios = inject("axios");
const listControl = ref(null);
const profile = authStore().getRBAC(FEATUREID);
layoutStore().name = "tenant";
const FEATUREID = "Inspection";
let customFilter = computed(() => {
  const filters = [];
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "CsmsDate",
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
      Field: "CsmsDate",
      Op: "$lte",
      Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
    });
  }

  if (data.search.No !== null && data.search.No !== "") {
    filters.push({
      Field: "_id",
      Op: "$contains",
      Value: [data.search.No],
    });
  }
  if (data.search.Customer !== null && data.search.Customer !== "") {
    filters.push({
      Field: "Customer",
      Op: "$eq",
      Value: data.search.Customer,
    });
  }
  if (data.search.Template !== null && data.search.Template !== "") {
    filters.push({
      Field: "TemplateID",
      Op: "$eq",
      Value: data.search.Template,
    });
  }
  if (data.search.Status !== null && data.search.Status !== "") {
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: data.search.Status,
    });
  }

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});
const data = reactive({
  appMode: "grid",
  waitTrxSubmit: true,
  record: {},
  cfg: {},
  lineCfg: [
    { field: "IsApplicable", label: "Not Applicable" },
    {
      field: "ValueDescription",
      label: "Work Description",
    },
    {
      field: "Metode",
      label: "Metode",
    },
    {
      field: "Bobot",
      label: "Bobot",
    },
    {
      field: "Result",
      label: "Result",
    },
    {
      field: "ACH",
      label: "ACH (%)",
    },
    {
      field: "Attachment",
      label: "Attachment",
    },
    {
      field: "Validation",
      label: "Validation",
    },
    {
      field: "ExpDate",
      label: "Exp Date",
    },
    {
      field: "Remark",
      label: "Remark",
    },
  ],
  listJurnalType: [],
  search: {
    DateFrom: null,
    DateTo: null,
    Status: "",
    No: "",
    Customer: "",
    Template: "",
  },
});

function newData(r) {
  r.CsmsDate = new Date();
  r.StartDate = new Date();
  r.EndDate = new Date();
  if (data.listJurnalType.length > 0) {
    r.JournalTypeID = data.listJurnalType[0]._id;
    r.PostingProfileID = data.listJurnalType[0].PostingProfileID;
  }
  openForm(r);
}

function editRecord(r) {
  openForm(r);
}

function openForm(r) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("JournalTypeID", "required", true);

    if (["SUBMITTED", "READY", "POSTED", "REJECTED"].includes(r.Status)) {
      listControl.value.setFormMode("view");
    } else {
      listControl.value.setFormMode("edit");
    }
  });
  data.record = r;
}

function onPreSave(r) {
  for (let i in r.Lines) {
    let o = r.Lines[i];
    o.Result = o.Result == null ? 0 : parseInt(o.Result);
  }
}

function onFormFieldChange(field, v1, v2, old, record) {
  switch (field) {
    case "JournalTypeID":
      util.nextTickN(2, () => {
        getJournalType(v1, record);
      });
      break;
    default:
      break;
  }
}

function trxPreSubmit(status, action, doSubmit) {
  if (["DRAFT"].includes(status)) {
    listControl.value.submitForm(
      data.record,
      () => {
        doSubmit();
      },
      () => {
        setLoadingForm(false);
      }
    );
  } else {
    doSubmit();
  }
}
function trxPostSubmit(data, action) {
  setLoadingForm(false);
  listControl.value.setControlMode("grid");
  listControl.value.refreshGrid();
}
function trxErrorSubmit() {
  setLoadingForm(false);
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function getJournalType(v1, record) {
  if (v1) {
    axios.post("/fico/shejournaltype/get", [v1]).then(
      (r) => {
        record.PostingProfileID = r.data.PostingProfileID;
      },
      (e) => util.showError(e)
    );
  }
}
function getsJournalType() {
  axios.post("/fico/shejournaltype/gets", {}).then(
    (r) => {
      data.listJurnalType = r.data.data;
    },
    (e) => util.showError(e)
  );
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
function alterGridConfig(cfg) {
  cfg.fields.push({
    field: "CSMSStatus",
    kind: "Text",
    label: "CSMS Status",
    readType: "show",
    input: {
      field: "CSMSStatus",
      label: "CSMS Status",
      hint: "",
      hide: false,
      placeHolder: "CSMS Status",
      kind: "text",
      disable: false,
      required: false,
      multiple: false,
    },
  });
  let sortColm = [
    "_id",
    "CsmsDate",
    "Customer",
    "TemplateID",
    "CSMSStatus",
    "Status",
  ];
  cfg.fields.map((f) => {
    f.idx = sortColm.indexOf(f.field);
    return f;
  });
  cfg.fields.sort((a, b) => (a.idx > b.idx ? 1 : -1));
}
onMounted(() => {
  getsJournalType();
});
</script>
