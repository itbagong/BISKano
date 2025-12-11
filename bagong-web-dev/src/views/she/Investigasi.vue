<template>
  <div>
    <data-list
      class="card investigasi_transaction"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/she/investigasi/gridconfig"
      form-config="/she/investigasi/formconfig"
      grid-read="/she/investigasi/gets"
      form-read="/she/investigasi/get"
      grid-mode="grid"
      grid-delete="/she/investigasi/delete"
      form-keep-label
      form-insert="/she/investigasi/save"
      form-update="/she/investigasi/save"
      :init-app-mode="data.appMode"
      :form-fields="['Likehood', 'Severity', 'Dimension']"
      :grid-fields="['Dimension', 'Status']"
      :form-tabs-edit="[
        'General',
        'Details Accident',
        'Involvement',
        'Basic Cause',
        'Direct Cause',
        'Investigation',
        'Risk Reduction',
        'External Report',
        'PICA',
        'Attachments',
      ]"
      :form-tabs-view="[
        'General',
        'Details Accident',
        'Involvement',
        'Basic Cause',
        'Direct Cause',
        'Investigation',
        'Risk Reduction',
        'External Report',
        'PICA',
        'Attachments',
      ]"
      :grid-custom-filter="customFilter"
      grid-hide-select
      stay-on-form-after-save
      @form-edit-data="editRecord"
      @form-new-data="newData"
      @pre-save="onPreSave"
      @postSave="onPostSave"
      @alterGridConfig="alterGridConfig"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
    >
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #grid_header_search="{ config }">
        <div
          class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter"
        >
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
            ref="refContractor"
            v-model="data.search.Location"
            class="w-[300px]"
            use-list
            label="Location"
            lookup-url="/tenant/masterdata/find?MasterDataTypeID=LOC"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refLocationDetail"
            v-model="data.search.LocationDetail"
            class="w-[300px]"
            label="Location Detail"
            @keyup.enter="refreshData"
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
      <template #form_tab_Details_Accident="{ item }">
        <div class="relative">
          <DetailsAccident
            ref="refDetailsAccident"
            v-model="data.record"
          ></DetailsAccident>
        </div>
      </template>
      <template #form_tab_Involvement="{ item }">
        <div class="relative">
          <Involvement
            ref="refInvolvement"
            v-model="data.record"
            @showInjured="showInjured"
            @showPartEquipmentAsset="showPartEquipmentAsset"
          ></Involvement>
        </div>
      </template>
      <template #form_tab_Direct_Cause="{ item }">
        <div class="relative">
          <DirectCause v-model="data.record"></DirectCause>
        </div>
      </template>
      <template #form_tab_Basic_Cause="{ item }">
        <div class="relative">
          <BasicCause v-model="data.record"></BasicCause>
        </div>
      </template>
      <template #form_tab_Investigation="{ item }">
        <div class="relative">
          <InvestigationTeam v-model="data.record"></InvestigationTeam>
        </div>
      </template>
      <template #form_tab_External_Report="{ item }">
        <div class="relative">
          <ExternalReport v-model="data.record"></ExternalReport>
        </div>
      </template>
      <template #form_tab_Risk_Reduction="{ item }">
        <div class="relative">
          <RiskReduction
            ref="refRiskReduction"
            v-model="data.record"
          ></RiskReduction>
        </div>
      </template>
      <template #form_tab_PICA="{ item }">
        <div class="relative">
          <Pica ref="refPica" v-model="data.record"></Pica>
        </div>
      </template>
      <template #form_tab_Attachments="{ item }">
        <s-grid-attachment
          :key="data.record._id"
          :journal-id="data.record._id"
          :tags="linesTag"
          journal-type="Investigasi"
          ref="gridAttachmentCtl"
          @pre-Save="preSaveAttachment"
        ></s-grid-attachment>
      </template>
      <template #grid_Dimension="{ item }">
        <DimensionText :dimension="item.Dimension" />
      </template>
      <template #form_input_Likehood="{ item, mode }">
        <div class="grid grid-cols-4 gap-2 mt-2">
          <div>
            <s-input
              ref="refLikehood"
              v-model="item.Likehood"
              class="w-full"
              label="Likehood"
              use-list
              :items="data.listLikehood"
              @change="
                onFormFieldChange('Likehood', item.listLikehood, '', '', item)
              "
            ></s-input>
            <s-input
              ref="refSeverity"
              v-model="item.Severity"
              class="w-full"
              label="Severity"
              use-list
              :items="data.Severity"
              @change="
                onFormFieldChange('Severity', item.Severity, '', '', item)
              "
            ></s-input>
          </div>
          <div class="risk-management">
            <div class="matrix">
              <div class="value-box" :class="data.bgRisk">{{ item.Level }}</div>
            </div>
          </div>
        </div>
      </template>
      <template #form_input_Severity="{ item, mode }">
        <s-input
          ref="refSeverity"
          v-model="item.Severity"
          class="w-full"
          label="Severity"
          use-list
          :items="data.Severity"
        ></s-input>
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :read-only="mode == 'view'"
        ></dimension-editor-vertical>
      </template>
      <template #form_buttons_1="{ item, inSubmission, loading }">
        <form-buttons-trx
          :disabled="loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="'INVESTIGASI'"
          :moduleid="'she'"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :auto-post="false"
        />
      </template>
    </data-list>
    <s-modal
      :display="data.isDialogInjured"
      hide-buttons
      ref="refModalInjured"
      title="Injured"
      @before-hide="data.isDialogInjured = false"
    >
      <div :class="`min-w-[800px]`">
        <s-card hide-title>
          <s-grid
            v-model="data.listInjured"
            ref="refInjuredLine"
            class="w-full"
            hide-search
            hide-sort
            hide-refresh-button
            hide-edit
            hide-select
            hide-paging
            editor
            auto-commit-line
            no-confirm-delete
            form-keep-label
            :config="data.cfgGridCfgInjured"
            @new-data="addInjured"
          >
            <template #item_button_delete="{ item, idx }">
              <a @click="removeInjured(item)" class="delete_action">
                <mdicon
                  name="delete"
                  width="16"
                  alt="delete"
                  class="cursor-pointer hover:text-primary"
                />
              </a>
            </template>
          </s-grid>
        </s-card>
      </div>
      <template #buttons="{ item }"> </template>
    </s-modal>
    <s-modal
      :display="data.isDialogPartEquipment"
      hide-buttons
      ref="refModalPartEquipment"
      title="Part Equipment"
      @before-hide="data.isDialogPartEquipment = false"
    >
      <div :class="`min-w-[1000px]`">
        <s-card hide-title>
          <s-grid
            :key="data.key"
            v-model="data.listPartEquipment"
            ref="refPartEquipmentLine"
            class="w-full"
            hide-search
            hide-sort
            hide-refresh-button
            hide-edit
            hide-select
            hide-paging
            editor
            auto-commit-line
            no-confirm-delete
            form-keep-label
            :config="data.cfgGridCfgPartEquipment"
            @new-data="addPartEquipment"
          >
            <template #item_ItemID="{ item, idx }">
              <s-input-sku-item
                v-model="item.ItemVarian"
                :record="item"
                :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
                  item.ItemID,
                  item.SKU
                )}`"
              ></s-input-sku-item>
            </template>

            <template #item_CompanyAsset="{ item, idx }">
              <s-toggle
                v-model="item.CompanyAsset"
                class="w-[120px] mt-0.5"
                yes-label="Yes"
                no-label="No"
              />
            </template>
            <template #item_button_delete="{ item, idx }">
              <a @click="removePartEquipment(item)" class="delete_action">
                <mdicon
                  name="delete"
                  width="16"
                  alt="delete"
                  class="cursor-pointer hover:text-primary"
                />
              </a>
            </template>
          </s-grid>
        </s-card>
      </div>
      <template #buttons="{ item }"> </template>
    </s-modal>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import {
  DataList,
  SInput,
  SForm,
  SGrid,
  SButton,
  loadGridConfig,
  loadFormConfig,
  util,
  SModal,
} from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import StatusText from "@/components/common/StatusText.vue";
import DetailsAccident from "./widget/Investigation/DetailsAccident.vue";
import Involvement from "./widget/Investigation/Involvement.vue";
import DirectCause from "./widget/Investigation/DirectCause.vue";
import BasicCause from "./widget/Investigation/BasicCause.vue";
import InvestigationTeam from "./widget/Investigation/InvestigationTeam.vue";
import ExternalReport from "./widget/Investigation/ExternalReport.vue";
import RiskReduction from "./widget/Investigation/RiskReduction.vue";
import Pica from "./widget/Investigation/PICA.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import SInputSkuItem from "../scm/widget/SInputSkuItem.vue";
import SToggle from "@/components/common/SButtonToggle.vue";
import helper from "@/scripts/helper.js";
import moment from "moment";
import LogTrx from "@/components/common/LogTrx.vue";

layoutStore().name = "tenant";
const listControl = ref(null);
const refInvolvement = ref(null);
const refInjuredLine = ref(null);
const refPartEquipmentLine = ref(null);
const refDetailsAccident = ref(null);
const refRiskReduction = ref(null);
const refPica = ref(null);
const axios = inject("axios");

const gridAttachmentCtl = ref(null);
const linesTag = computed({
  get() {
    const tags = [`Ivestigasi-${data.record._id}`];
    return [...tags, `inv-dtl-accident-${data.record._id}`];
  },
});

let currentTab = computed(() => {
  if (listControl.value == null) {
    return 0;
  } else {
    return listControl.value.getFormCurrentTab();
  }
});
let customFilter = computed(() => {
  const filters = [];
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "Created",
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
      Field: "Created",
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
  if (data.search.Location !== null && data.search.Location !== "") {
    filters.push({
      Field: "Location",
      Op: "$eq",
      Value: data.search.Location,
    });
  }
  if (
    data.search.LocationDetail !== null &&
    data.search.LocationDetail !== ""
  ) {
    filters.push({
      Field: "LocationDetail",
      Op: "$eq",
      Value: data.search.LocationDetail,
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
  titleForm: "Investigasi",
  bgRisk: "bg-green",
  isLoadRisk: true,
  isDialogInjured: false,
  isDialogPartEquipment: false,
  record: {},
  listLikehood: [],
  Severity: [],
  cfgGridCfgInjured: {},
  cfgGridCfgPartEquipment: {},
  listInjured: [],
  listPartEquipment: [],
  listJurnalType: [],
  search: {
    DateFrom: null,
    DateTo: null,
    Status: "",
    No: "",
    Location: "",
    LocationDetail: "",
  },
  key: 0,
});

function newData(r) {
  r.AccidentDate = moment().format("YYYY-MM-DDTHH:mm");
  r.ReportingDate = moment().format("YYYY-MM-DDTHH:mm");
  r.Likehood = 1;
  r.Severity = 1;
  if (data.listJurnalType.length > 0) {
    r.JournalTypeID = data.listJurnalType[0]._id;
    r.PostingProfileID = data.listJurnalType[0].PostingProfileID;
  }
  data.titleForm = "Create New Investigasi";
  getsRiskmatrix(data.record);

  openForm(r);
}

function editRecord(r) {
  r.AccidentDate = moment(
    moment(r.AccidentDate ? r.AccidentDate : new Date()).format(
      "YYYY-MM-DDTHH:mm:00Z"
    )
  ).format("YYYY-MM-DDTHH:mm");
  r.ReportingDate = moment(
    moment(r.ReportingDate ? r.ReportingDate : new Date()).format(
      "YYYY-MM-DDTHH:mm:00Z"
    )
  ).format("YYYY-MM-DDTHH:mm");

  data.titleForm = `Edit Investigasi | ${r._id}`;
  util.nextTickN(2, () => {
    if (["SUBMITTED", "READY", "POSTED", "REJECTED"].includes(r.Status)) {
      listControl.value.setFormMode("view");
    } else {
      listControl.value.setFormMode("edit");
    }
  });
  openForm(r);
}
function openForm(r) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("Shift", "hide", true);
    listControl.value.setFormFieldAttr("Location", "required", true);
    listControl.value.setFormFieldAttr("JournalTypeID", "required", true);
  });
  data.record = r;
}

function onPreSave(record) {
  record.Likehood = record.Likehood.toString();
  record.Severity = record.Severity.toString();
  record.AccidentDate = moment(record.AccidentDate).format(
    "YYYY-MM-DDTHH:mm:00Z"
  );

  record.ReportingDate = moment(record.ReportingDate).format(
    "YYYY-MM-DDTHH:mm:00Z"
  );
}
function onPostSave(record) {
  record.Likehood = moment(
    moment(record.Likehood ? record.Likehood : new Date()).format(
      "YYYY-MM-DDTHH:mm:00Z"
    )
  ).format("YYYY-MM-DDTHH:mm");

  record.ReportingDate = moment(
    moment(record.ReportingDate ? record.ReportingDate : new Date()).format(
      "YYYY-MM-DDTHH:mm:00Z"
    )
  ).format("YYYY-MM-DDTHH:mm");

  if (refDetailsAccident.value) {
    refDetailsAccident.value.onSaveAttachment();
  }
  if (gridAttachmentCtl.value) {
    gridAttachmentCtl.value.Save();
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

function onFormFieldChange(field, v1, v2, old, record) {
  switch (field) {
    case "Likehood":
      getsRiskmatrix(record);
      break;
    case "Severity":
      getsRiskmatrix(record);
      break;
    case "JournalTypeID":
      util.nextTickN(2, () => {
        getJournalType(v1, record);
      });
      break;
    default:
      break;
  }
}

function preSaveAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [`Ivestigasi-${data.record._id}`];
    return asset;
  });
}

function alterGridConfig(cfg) {
  cfg.sortable = ["Created", "_id"];
  cfg.setting.idField = "Created";
  cfg.setting.sortable = ["Created", "_id"];
}

function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = `Investigasi`;
  }
}

function getsRiskmatrix(record) {
  axios.post("/bagong/riskmatrix/gets?Type=IBPR", {}).then(
    (r) => {
      const Matrix = r.data.data.find((m) => {
        return (
          m.SeverityName == record.Severity.toString() &&
          m.LikelihoodName == record.Likehood.toString()
        );
      });
      record.Level = `${Matrix.RiskID} (${Matrix.Value})`;
      if (Matrix.RiskID == "C") {
        data.bgRisk = "bg-green";
      } else if (Matrix.RiskID == "B") {
        data.bgRisk = "bg-yellow";
      } else if (Matrix.RiskID == "A") {
        data.bgRisk = "bg-orange";
      } else {
        data.bgRisk = "bg-red";
      }
    },
    (e) => util.showError(e)
  );
}

function getsSeverity() {
  axios.post("/bagong/severity/gets?Type=IBPR", {}).then(
    (r) => {
      data.Severity = r.data.data
        .map((v) => {
          return v.Value;
        })
        .sort((a, b) => a - b);
    },
    (e) => util.showError(e)
  );
}
function getslikelihood() {
  axios.post("/bagong/likelihood/gets?Type=IBPR", {}).then(
    (r) => {
      data.listLikehood = r.data.data
        .map((v) => {
          return v.Value;
        })
        .sort((a, b) => a - b);
    },
    (e) => util.showError(e)
  );
}

function showInjured(record) {
  data.isDialogInjured = true;
  data.listInjured = record;
}
function showPartEquipmentAsset(record) {
  data.isDialogPartEquipment = true;
  data.listPartEquipment = record;
  // data.listPartEquipment = [{}, {}];
}

function addInjured() {
  let r = {};
  const noLine = data.listInjured.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.BodyPart = "";
  r.Side = "";
  r.Remark = "";
  data.listInjured.push(r);
  // refInvolvement.value.updateGridLine(data.listInjured, "Injured");
  refInjuredLine.value.setRecords(data.listInjured);
}

function addPartEquipment() {
  let r = {};
  const noLine =
    data.listPartEquipment.length ??
    Math.max(...data.listPartEquipment.map((item) => item.LineNo));
  r._id = util.uuid();
  r.LineNo = noLine;
  r.ItemID = "";
  r.CompanyAsset = false;
  r.Estimation = "";
  r.Remark = "";
  data.listPartEquipment.push(r);
  refPartEquipmentLine.value.setRecords(data.listPartEquipment);
}

function removeInjured(r) {
  data.listInjured = data.listInjured.filter((obj) => obj._id !== r._id);
  refInvolvement.value.updateGridLine(data.listInjured, "Injured");
  refInjuredLine.value.setRecords(data.listInjured);
}

function loadGridInjuredline() {
  let url = `/she/investigasi/injuredline/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridCfgInjured = r;
    },
    (e) => {}
  );
}
function loadGridPartEquipmentline() {
  let url = `/she/investigasi/partequipmentline/gridconfig`;
  loadGridConfig(axios, url).then(
    (asset) => {
      loadGridConfig(axios, `scm/inventory/journal/line/gridconfig`).then(
        (journal) => {
          asset.fields = [
            ...asset.fields,
            ...journal.fields.filter((f) => f.field == "ItemID"),
          ];
          let sort = [
            "LineNo",
            "ItemID",
            "CompanyAsset",
            "Estimation",
            "Remark",
          ];
          asset.fields.map((f) => {
            if (f.field == "ItemID") {
              f.label = "Item";
              f.width = "350px";
            } else if (f.field == "CompanyAsset") {
              f.width = "50px";
            }
            f.idx = sort.indexOf(f.field);
            return f;
          });
          asset.fields = asset.fields.sort((a, b) => (a.idx > b.idx ? 1 : -1));
          data.cfgGridCfgPartEquipment = asset;
        },
        (e) => {}
      );
    },
    (e) => {}
  );
}

function onFieldChanged(record) {
  let RiskReduction = [];
  let SourceId = [...record.BasicCause, ...record.DirectCause].map((e) => {
    return e._id;
  });
  for (let i = 0; i < record.BasicCause.length; i++) {
    RiskReduction.push({
      _id: util.uuid(),
      Parent: true,
      LineNo: RiskReduction.length + 1,
      SourceId: record.BasicCause[i]._id,
      IdentifiedCause: record.BasicCause[i].SubBCDetail,
      SubParent: true,
      ControlNo: util.uuid(),
      KontrolPengendalian: "",
      SubKontrolPengendalian: "",
      Remark: "",
    });
  }

  for (let i = 0; i < record.DirectCause.length; i++) {
    RiskReduction.push({
      _id: util.uuid(),
      Parent: true,
      LineNo: RiskReduction.length + 1,
      SourceId: record.DirectCause[i]._id,
      IdentifiedCause: record.DirectCause[i].SubDCDetail,
      SubParent: true,
      ControlNo: util.uuid(),
      KontrolPengendalian: "",
      SubKontrolPengendalian: "",
      Remark: "",
    });
  }

  let datas = [];
  if (record.RiskReduction.length == 0) {
    record.RiskReduction = RiskReduction;
  } else {
    datas = JSON.parse(JSON.stringify(record.RiskReduction)).filter((r) => {
      return SourceId.includes(r.SourceId);
    });

    let RiskSource = JSON.parse(JSON.stringify(datas)).map((e) => {
      return e.SourceId;
    });

    let NewRiskReduction = RiskReduction.filter((r) => {
      return !RiskSource.includes(r.SourceId);
    }).map((m, i) => {
      let number = record.RiskReduction.at(-1)
        ? record.RiskReduction.at(-1).LineNo + (i + 1)
        : i + 1;
      m.LineNo = number;
      return m;
    });
    record.RiskReduction = [...datas, ...NewRiskReduction];
  }
  refRiskReduction.value.updateGridLines();
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
watch(
  () => currentTab.value,
  (nv) => {
    if (nv == 6) {
      onFieldChanged(data.record);
    } else if (nv == 8) {
      const grouped = data.record.RiskReduction.reduce((acc, item) => {
        const key = item.SourceId;
        if (!acc[key]) {
          acc[key] = [];
        }
        acc[key].push(item);
        return acc;
      }, {});

      let datas = [];
      for (const key in grouped) {
        let Action = [];
        for (let idx = 0; idx < grouped[key].length; idx++) {
          Action.push(grouped[key][idx].Remark);
        }
        datas.push({
          _id: util.uuid(),
          Cause: grouped[key][0].IdentifiedCause,
          Action: Action.join(),
          PIC: "",
          DueDate: new Date(),
          Status: "",
        });
      }
      data.record.PICA = datas;
      refPica.value.updateGridLines();
    }
  }
);

onMounted(() => {
  getsSeverity();
  getslikelihood();
  loadGridInjuredline();
  loadGridPartEquipmentline();
  getsJournalType();
});
</script>
<style lang="css" scoped>
.risk-management {
  display: flex;
  justify-content: center;
  align-items: center;
}
.value-box {
  padding: 10px;
  border-radius: 5px;
  color: white;
  display: inline-flex;
  justify-content: center;
  align-items: center;
  text-align: center;
  font-size: 20px;
  font-weight: 700;
}

.bg-green {
  background-color: #00b050; /* Warna latar belakang */
  border: 2px solid #03813c; /* Garis border */
}
.bg-yellow {
  background-color: #ffff00; /* Warna latar belakang */
  border: 2px solid #bebe03; /* Garis border */
}
.bg-orange {
  background-color: #ffc000; /* Warna latar belakang */
  border: 2px solid #bb8e07; /* Garis border */
}
.bg-red {
  background-color: red; /* Warna latar belakang */
  border: 2px solid darkred; /* Garis border */
}
</style>
