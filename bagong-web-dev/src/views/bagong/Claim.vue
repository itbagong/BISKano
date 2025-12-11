<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Claim"
      grid-config="/bagong/claim/gridconfig"
      form-config="/bagong/claim/formconfig"
      grid-read="/bagong/claim/gets"
      form-read="/bagong/claim/get"
      grid-mode="grid"
      grid-delete="/bagong/claim/delete"
      form-insert="/bagong/claim/save"
      form-update="/bagong/claim/save"
      :form-fields="['Dimension', 'Lines', 'Summary', 'Position']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      :form-tabs-edit="['General', 'Line']"
      :form-tabs-new="['General', 'Line']"
      :form-tabs-view="['General', 'Line']"
      @formFieldChange="onChangeField"
      @formEditData="editRecord"
      @formNewData="newRecord"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_Dimension="{ item,mode }">
        <dimension-editor-vertical v-model="item.Dimension"  :default-list="profile.Dimension" :read-only="mode == 'view'"/>
      </template>
      <template #form_tab_Line="{ item,mode}">
        <s-grid
          ref="gridLines"
          :config="data.gridCfgLines"
          :hide-detail="mode ==' view'"
          :editor="mode !== 'view'"
          :hideNewButton="mode == 'view'"
          hide-select
          hide-search
          hide-sort
          hide-refresh-button
          hide-save-button
          @new-data="onNewDataLines"
          @delete-data="onDeleteGridLines" 
          v-model="data.record.Lines"
          @row-field-changed="onChangeGridLines"
        >
        </s-grid>
      </template>
      <template #form_input_Summary="{ item,mode }">
        <s-grid
          ref="gridSummary"
          :config="data.gridCfgSummary"
          hide-select
          hide-search
          hide-sort
          hide-refresh-button
          hide-new-button
          hide-detail
          hide-delete-button
          v-model="data.record.Summary"
        >
          <template #item_buttons="{ item }">
            <s-button
              class="btn_primary mx-1 my-2"
              label="Submit"
              @click="onSubmitSummary(item)"
              v-if="showHideSubmit(item)"
            />
          </template>
          <template #item_OffsetAccount="{ item }">
            <s-input
              v-model="item.OffsetAccount"
              use-list
              :lookup-url="`/tenant/ledgeraccount/find`"
              lookup-key="_id"
              :lookup-labels="['_id', 'Name']"
              :lookup-search="['_id', 'Name']"
              :read-only="mode == 'view'"
            />
          </template>
        </s-grid>
      </template>
      <template #form_input_Position="{ item }">
        <s-input
          v-model="item.Position"
          label="Position"
          use-list
          :lookup-url="`/tenant/masterdata/find?_id=${item.Position}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          read-only
          :key="item.Position"
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, watch } from "vue";
import {
  DataList,
  util,
  SInput,
  SButton,
  SModal,
  SGrid,
  loadGridConfig,
} from "suimjs";
import { authStore } from "@/stores/auth.js";

import { layoutStore } from "@/stores/layout.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";

layoutStore().name = "tenant";
const FEATUREID = 'Claim'
const profile = authStore().getRBAC(FEATUREID)

const listControl = ref(null);
const gridSummary = ref(null);
const gridLines = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  record: {},
  gridCfgSummary: {},
  gridCfgLines: {},
});

function onChangeField(name, v1, v2, old) {
  if (name == "EmployeeID") {
    getPosition(v1);
  }
}

function newRecord(record) {
  data.record = record;
  data.record.Summary = [];
  data.record.Lines = [];
  getTypeSummary();
}

function editRecord(record) {
  data.record = record;
  getTypeSummary();
}

function getPosition(id) {
  const url = "/bagong/employeedetail/find?EmployeeID=" + id;
  axios.post(url).then(
    (r) => {
      data.record.Position = r.data[0].Position;
    },
    (e) => {
      util.showError(e.error);
    }
  );
}

function getTypeSummary() {
  const url = "/tenant/masterdata/find?MasterDataTypeID=CLT";
  axios.post(url).then(
    (r) => {
      buildDataSummary(r.data);
    },
    (e) => {
      util.showError(e.error);
    }
  );
}

function buildDataSummary(src) {
  let res = [];
  let exist = data.record.Summary;
  for (let i in src) {
    let o = src[i];
    let objEmpty = {
      ClaimTypeID: o._id,
      ClaimSummaryAmount: 0,
      Balance: 0,
      OffsetAccount: "",
    };

    let f = exist.find((x) => x.ClaimTypeID == o._id);
    if (f) {
      objEmpty["ClaimSummaryAmount"] = f["ClaimSummaryAmount"];
      objEmpty["Balance"] = f["Balance"];
      objEmpty["OffsetAccount"] = f["OffsetAccount"];
    }

    res.push(objEmpty);
  }
  data.record.Summary = res;
  gridSummary.value.setRecords(res);
}

function onSubmitSummary(src) {
  console.log(src);
}

function onNewDataLines() {
  let obj = {
    Date: new Date(),
    Mutation: null,
    ClaimTypeID: "",
    isPlus: false,
  };
  gridLines.value.addData(obj);
}

function onDeleteGridLines(dt, index) {
  data.record.Lines.splice(index, 1);
}

function showHideSubmit(item) {
  return (
    item.OffsetAccount !== "" &&
    item.Balance == 0 &&
    item.ClaimSummaryAmount > 0
  );
}

function onChangeGridLines(name, v1, v2, current) {
  if (name == "Mutation") current.isPlus = v1 > 0;
}

function generateSummary() {
  if (data.record.Lines.length == 0) {
    data.record.Summary = [];
    getTypeSummary();
    return;
  }

  let obj = {};
  let src = data.record.Summary;
  for (let i in data.record.Lines) {
    let o = data.record.Lines[i];
    if (o.isPlus) {
      obj[o.ClaimTypeID + "#ClaimSummaryAmount"] =
        (obj[o.ClaimTypeID + "#ClaimSummaryAmount"] ?? 0) + o.Mutation;
    }
    obj[o.ClaimTypeID + "#Balance"] =
      (obj[o.ClaimTypeID + "#Balance"] ?? 0) + o.Mutation;
  }

  for (let ky in obj) {
    let f = src.find((o) => o.ClaimTypeID == ky.split("#")[0]);
    if (f) f[ky.split("#")[1]] = obj[ky];
  }
}

watch(
  () => listControl.value?.getFormCurrentTab(),
  (nv) => {
    if (nv == 0) generateSummary();
  }
);

onMounted(() => {
  loadGridConfig(axios, "/bagong/claim/line/gridconfig").then(
    (r) => {
      data.gridCfgLines = r;
    },
    (e) => util.showError(e)
  );
  loadGridConfig(axios, "/bagong/claim/summary_amount/gridconfig").then(
    (r) => {
      data.gridCfgSummary = r;
    },
    (e) => util.showError(e)
  );
});
</script>
