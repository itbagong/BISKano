x
<template>
  <data-list
    class="card rsca_transaction"
    ref="listControl"
    :title="data.titleForm"
    grid-config="/she/jsa/gridconfig"
    form-config="/she/jsa/formconfig"
    grid-read="/she/jsa/gets"
    form-read="/she/jsa/get"
    grid-mode="grid"
    grid-delete="/she/jsa/delete"
    form-keep-label
    form-insert="/she/jsa/save"
    form-update="/she/jsa/save"
    :init-app-mode="data.appMode"
    grid-hide-select
    @form-edit-data="editRecord"
    @form-new-data="newData"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    stay-on-form-after-save
    @alterGridConfig="alterGridConfig"
    @controlModeChanged="onControlModeChanged"
    @form-field-change="onFormFieldChange"
    :form-fields="['Dimension']"
    :form-tabs-edit="['General', 'Lines']"
    :form-tabs-view="['General', 'Lines']"
    :grid-fields="['Dimension', 'Status']"
    :grid-custom-filter="customFilter"
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
          ref="refType"
          v-model="data.search.Type"
          lookup-key="_id"
          label="Type"
          class="w-[200px]"
          use-list
          :items="['New', 'Revision']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refPosition"
          v-model="data.search.Position"
          class="w-[300px]"
          use-list
          label="Position"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
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
    <template #form_input_Dimension="{ item, mode }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :read-only="mode == 'view'"
      ></dimension-editor-vertical>
    </template>
    <template #form_tab_Lines="{ item }">
      <Lines
        ref="lineConfig"
        v-model="item.Lines"
        :item="item"
        :readOnly="data.formMode == 'edit'"
        :form-mode="data.formMode"
        :hide-detail="true"
        grid-config="/she/jsa/lines/gridconfig"
      />
    </template>
    <template #form_buttons_1="{ item, inSubmission, loading }">
      <form-buttons-trx
        :disabled="loading"
        :status="item.Status"
        :journal-id="item._id"
        :posting-profile-id="item.PostingProfileID"
        :journal-type-id="'JSA'"
        :moduleid="'she'"
        @preSubmit="trxPreSubmit"
        @postSubmit="trxPostSubmit"
        @errorSubmit="trxErrorSubmit"
        :auto-post="false"
      />
    </template>
    <template #grid_Status="{ item }">
      <status-text :txt="item.Status" />
    </template>
  </data-list>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed } from "vue";

import { DataList, util, SInput } from "suimjs";
import { layoutStore } from "@/stores/layout.js";
layoutStore().name = "tenant";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import Lines from "./widget/JsaLines.vue";
import helper from "@/scripts/helper.js";

const listControl = ref(null);
const axios = inject("axios");
let customFilter = computed(() => {
  const filters = [];
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "JsaDate",
      Op: "$gte",
      Value: helper.formatFilterDate(data.search.DateFrom),
    });
  }
  if (
    data.search.DateTo !== null &&
    data.search.DateTo !== "" &&
    data.search.DateTo !== "Invalid date"
  ) {
    filters.push({
      Field: "JsaDate",
      Op: "$lte",
      Value: helper.formatFilterDate(data.search.DateTo, true),
    });
  }

  if (data.search.No !== null && data.search.No !== "") {
    filters.push({
      Field: "_id",
      Op: "$contains",
      Value: [data.search.No],
    });
  }
  if (data.search.Type !== null && data.search.Type !== "") {
    filters.push({
      Field: "Type",
      Op: "$eq",
      Value: data.search.Type,
    });
  }
  if (data.search.Position !== null && data.search.Position !== "") {
    filters.push({
      Field: "PositionInvolved",
      Op: "$eq",
      Value: data.search.Position,
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
  record: {},
  currentActionCfg: {},
  riskMatrix: {},
  titleForm: "JSA",
  listJurnalType: [],
  search: {
    DateFrom: null,
    DateTo: null,
    Status: "",
    No: "",
    Type: "",
    Position: "",
  },
});
const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
function newData(r) {
  r.Status = "";
  if (data.listJurnalType.length > 0) {
    r.JournalTypeID = data.listJurnalType[0]._id;
    r.PostingProfileID = data.listJurnalType[0].PostingProfileID;
  }
  r.JsaDate = new Date();
  r.Lines = [];
  r.EquipmentInvolved = [];
  r.Apd = [];

  data.titleForm = "Create New JSA";
  openForm(r);
}

function editRecord(r) {
  openForm(r);
  // data.titleForm = `Edit JSA | ${r._id}`;
  util.nextTickN(2, () => {
    if (readOnly.value) {
      listControl.value.setFormMode("view");
    }
  });
}

function openForm(r) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("JournalTypeID", "required", true);
  });
  data.record = r;
}

function onPreSave(saveData) {
  let StepsTask = "";
  let Recommendation = "";
  saveData.Lines.map((l) => {
    if (l.Parent) {
      StepsTask = l.StepsTask;
      Recommendation = l.Recommendation;
    } else {
      l.StepsTask = StepsTask;
      l.Recommendation = Recommendation;
    }
    return l;
  });
}

function alterGridConfig(cfg) {
  cfg.sortable = ["Created", "_id"];
  cfg.setting.idField = "Created";
  cfg.setting.sortable = ["Created", "_id"];
}

function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = `JSA`;
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
onMounted(() => {
  getsJournalType();
});
</script>
