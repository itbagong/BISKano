<template>
  <div class="w-full man_power_req">
    <data-list
      v-if="data.appMode != 'preview'"
      class="card"
      ref="listControl"
      title="Man Power"
      grid-config="/hcm/manpowerrequest/gridconfig"
      form-config="/hcm/manpowerrequest/formconfig"
      grid-read="/hcm/manpowerrequest/gets"
      form-read="/hcm/manpowerrequest/get"
      grid-mode="grid"
      grid-delete="/hcm/manpowerrequest/delete"
      form-keep-label
      form-insert="/hcm/manpowerrequest/save"
      form-update="/hcm/manpowerrequest/save"
      :form-fields="['Dimension', 'Position', 'RequestorID', 'EmployementType', 'EmployeeSource', 'JobVacancyTitle']"
      :grid-fields="['IsClose', 'Status']"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :formHideSubmit="readOnly"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
      :form-tabs-edit="tabs"
      :form-tabs-view="tabs"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="data.customFilter"
      stay-on-form-after-save
    >
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilterCtl"
          v-model="data.customFilter"
          @initNewItem="initNewItemFilter"
          @preChange="changeFilter"
          hideFilterText
          hideFilterDate
          @change="refreshGrid"
        >
          <template #filter_1="{ item }">
            <s-input
              useList
              label="Job Title"
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              v-model="item.JobVacancyTitles"
              class="min-w-[180px]"
              multiple
            />
            <s-input
              useList
              label="Request"
              lookup-url="/tenant/employee/find"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              v-model="item.RequestorIDs"
              class="min-w-[180px]"
              multiple
            />
          </template>
          <template #filter_2="{ item }">
            <div class="flex gap-1">
              <s-input
                label="Request Date From "
                kind="date"
                v-model="item.ReqDateFrom"
              />
              <s-input
                label="Request Date To "
                kind="date"
                v-model="item.ReqDateTo"
              />
            </div>
            <div class="flex gap-1">
              <s-input
                label="Onsite Request Date From "
                kind="date"
                v-model="item.OnsiteDateFrom"
              />
              <s-input
                label="Onsite Request Date To "
                kind="date"
                v-model="item.OnsiteDateTo"
              />
            </div>
          </template>
        </grid-header-filter>
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <!-- grid status -->
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_buttons_1="{ item, inSubmission, loading, mode }">
        <form-buttons-trx
          :disabled="inSubmission || loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="data.jType"
          :auto-post="!waitTrxSubmit"
          moduleid="hcm"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
        />
        <s-button
          class="btn_primary"
          label="Close"
          icon="Close"
          @click="onClosingJob(item)"
        ></s-button>
        <template v-if="mode !== 'new'">
          <s-button
            :disabled="inSubmission || loading"
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Preview"
            @click="data.appMode = 'preview'"
          ></s-button>
        </template>
      </template>
      <template #grid_IsClose="{ item }">
        <div>{{ item.IsClose ? "Close" : "Open" }}</div>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          ref="Dimension"
          v-model="item.Dimension"
          :read-only="readOnly"
          sectionTitle="Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
      <template #form_input_RequestorID="{item}">
        <s-input
          v-if="readOnly"
          class="w-full"
          label="Requestor ID"
          use-list
          lookup-url="/tenant/employee/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.RequestorID"
          :read-only="readOnly || mode == 'view'"
        />
      </template>
      <template #form_input_EmployementType="{item}">
        <s-input
          v-if="readOnly"
          class="w-full"
          label="Employement type"
          use-list
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=EmploymentType"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.EmployementType"
          :read-only="readOnly || mode == 'view'"
        />
      </template>
      <template #form_input_EmployeeSource="{item}">
        <s-input
          v-if="readOnly"
          class="w-full"
          label="Employement source"
          use-list
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=EmployeeSource"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.EmployeeSource"
          :read-only="readOnly || mode == 'view'"
        />
      </template>
      <template #form_input_JobVacancyTitle="{item}">
        <s-input
          v-if="readOnly"
          class="w-full"
          label="JobVacancy title"
          use-list
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.JobVacancyTitle"
          :read-only="readOnly || mode == 'view'"
        />
      </template>
      <template #form_tab_Tracking="{ item }">
        <tracking :man-power-id="item._id" />
      </template>
      <template #form_input_Position="{ item }">
        <s-tab
          :tabs="['Replacement', 'Additional']"
          prefixClass="requestreason"
          @activeTab="getActiveTab"
          ref="customTab"
        >
          <template #tab_Replacement_body>
            <div class="grid grid-cols-4 gap-4">
              <s-input
                label="Position"
                v-model="item.Position"
                use-list
                lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
                lookup-key="_id"
                :lookup-labels="['Name']"
              />
              <s-input label="Class" v-model="item.Class" />
              <s-input
                label="Replaced employee name"
                v-model="item.ReplacedEmployeeName"
                use-list
                lookup-url="/tenant/employee/find"
                lookup-key="_id"
                :lookup-labels="['Name']"
              />
              <s-input
                label="Replacemnet reason"
                v-model="item.ReplacementSeason"
                multi-row="5"
              />
            </div>
          </template>
          <template #tab_Additional_body>
            <div class="grid grid-cols-4 gap-4">
              <s-input
                label="Additional number"
                v-model="item.AdditionalNumber"
                kind="number"
                @change="
                  (_, v1) => {
                    item.EmployeeNumberTotal = v1 + item.ExistingEmployeeNumber;
                  }
                "
              />
              <s-input
                label="Existing employee number"
                v-model="item.ExistingEmployeeNumber"
                kind="number"
                @change="
                  (_, v1) => {
                    item.EmployeeNumberTotal = item.AdditionalNumber + v1;
                  }
                "
              />
              <s-input
                label="Estimate cost per month"
                v-model="item.EstimateCostPerMonth"
                kind="number"
              />
              <s-input
                label="Employee number total"
                v-model="item.EmployeeNumberTotal"
                kind="number"
                read-only
              />
              <s-input
                label="Reason additional employee"
                v-model="item.ReasonAdditionalEmployee"
                multi-row="5"
              />
            </div>
          </template>
        </s-tab>
      </template>
    </data-list>

    <PreviewReport
      v-if="data.appMode == 'preview'"
      class="card w-full"
      title="Preview"
      @close="closePreview"
      :disable-print="helper.isDisablePrintPreview(data.record.Status)"
      :SourceType="data.jType"
      :SourceJournalID="data.record._id"
    >
      <template #buttons="props">
        <div class="flex gap-[1px] mr-2">
          <form-buttons-trx
            :disabled="inSubmission || loading"
            :status="data.record.Status"
            :journal-id="data.record._id"
            :posting-profile-id="data.record.PostingProfileID"
            :journal-type-id="data.jType"
            moduleid="hcm"
            @preSubmit="trxPreSubmit"
            @postSubmit="trxPostSubmit"
            @errorSubmit="trxErrorSubmit"
            :auto-post="!waitTrxSubmit"
          />
        </div>
      </template>
    </PreviewReport>
  </div>
</template>

<script setup>
import { reactive, ref, inject, computed } from "vue";
import { DataList, SInput, SButton, util } from "suimjs";
import STab from "@/components/common/STab.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import StatusText from "@/components/common/StatusText.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import helper from "@/scripts/helper.js";
import Tracking from "./widget/ManPowerRequest/Tracking.vue";

layoutStore().name = "tenant";

const featureID = "ManPower";
const axios = inject("axios");

const profile = authStore().getRBAC(featureID);
const auth = authStore();

const listControl = ref(null);
const customTab = ref(null);

const gridHeaderFilterCtl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  record: {},
  tabs: ["General"],
  customFilter: null,
  tabName: "",
  jType: "MANPOWER",
  journalType: {},
});

function getActiveTab(name) {
  data.tabName = name;
}

function refreshGrid() {
  listControl.value.refreshGrid();
}

function initNewItemFilter(item) {
  item.JobVacancyTitles = [];
  item.RequestorIDs = [];
  item.ReqDateFrom = null;
  item.ReqDateTo = null;
  item.OnsiteDateFrom = null;
  item.OnsiteDateTo = null;
}

function changeFilter(item, filters) {
  if (item.JobVacancyTitles.length > 0) {
    filters.push({
      Op: "$in",
      Field: "JobVacancyTitle",
      Value: item.JobVacancyTitles,
    });
  }
  if (item.RequestorIDs.length > 0) {
    filters.push({
      Op: "$in",
      Field: "RequestorID",
      Value: item.RequestorIDs,
    });
  }
  if (item.ReqDateFrom != null) {
    filters.push({
      Op: "$gte",
      Field: "RequestDate",
      Value: helper.formatFilterDate(item.ReqDateFrom),
    });
  }
  if (item.ReqDateTo != null) {
    filters.push({
      Op: "$lte",
      Field: "RequestDate",
      Value: helper.formatFilterDate(item.ReqDateTo),
    });
  }
  if (item.OnsiteDateFrom != null) {
    filters.push({
      Op: "$gte",
      Field: "OnsiteRequiredDate",
      Value: helper.formatFilterDate(item.OnsiteDateFrom),
    });
  }
  if (item.OnsiteDateTo != null) {
    filters.push({
      Op: "$lte",
      Field: "OnsiteRequiredDate",
      Value: helper.formatFilterDate(item.OnsiteDateTo),
    });
  }
}

function newRecord(record) {
  record._id = "";
  record.RequestDate = new Date();
  record.Status = "DRAFT";
  record.CompanyID = auth.appData.CompanyID;
  data.record = record;
  openForm(record);
}

function editRecord(record) {
  if (!record.CompanyID) {
    record.CompanyID = auth.appData.CompanyID;
  }
  if (record.Status == "") {
    record.Status = "DRAFT";
  }
  data.record = record;
  openForm(record);
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("PostingProfileID", "readOnly", true);
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
  });
}

function calcEmployeeNumberTotal(record, total) {
  record.EmployeeNumberTotal = total;
}

function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "AdditionalNumber":
      calcEmployeeNumberTotal(record, v1 + record.ExistingEmployeeNumber);
      break;
    case "ExistingEmployeeNumber":
      calcEmployeeNumberTotal(record, v1 + record.AdditionalNumber);
      break;
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
  }
}

function onControlModeChanged() {}
function onClosingJob(item) {
  listControl.value.setFormLoading(true);
  axios
    .post("/hcm/manpowerrequest/close-manpower-request", { JobID: item._id })
    .then(
      (r) => {
        listControl.value.setControlMode("grid");
      },
      (e) => {
        util.showError(e);
      }
    )
    .finally(() => {
      listControl.value.setFormLoading(false);
    });
}
const tabs = computed({
  get() {
    return ["POSTED"].includes(data.record.Status)
      ? ["General", "Tracking"]
      : ["General"];
  },
});
const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "SUBMITTED", "READY"].includes(data.record.Status);
  },
});
function getJurnalType(id) {
  if (id === "" || id === null) {
    data.journalType = {};
    data.record.PostingProfileID = "";
    return;
  }
  listControl.value.setFormLoading(true);
  axios
    .post("/hcm/journaltype/get", [id])
    .then(
      (r) => {
        data.journalType = r.data;
        data.record.PostingProfileID = r.data.PostingProfileID;
      },
      (e) => {
        data.journalType = {};
        data.record.PostingProfileID = "";
        util.showError(e);
      }
    )
    .finally(() => {
      listControl.value.setFormLoading(false);
    });
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    listControl.value.setFormFieldAttr(e.field, "required", required);
  });
}
function trxPreSubmit(status, action, doSubmit) {
  if (waitTrxSubmit.value) {
    listControl.value.setFormCurrentTab(0);
    trxSubmit(doSubmit);
  }
}
function trxSubmit(doSubmit) {
  setFormRequired(true);
  util.nextTickN(2, () => {
    const valid = listControl.value.formValidate();
    if (valid) {
      setLoadingForm(true);
      listControl.value.submitForm(
        data.record,
        () => {
          doSubmit();
        },
        () => {
          setLoadingForm(false);
        }
      );
    }
    setFormRequired(false);
  });
}
function trxPostSubmit(record, action) {
  setLoadingForm(false);
  closePreview();
  setModeGrid();
}
function trxErrorSubmit(e) {
  setLoadingForm(false);
}
function closePreview() {
  data.appMode = "grid";
}
function setModeGrid() {
  listControl.value.setControlMode("grid");
  listControl.value.refreshList();
}
</script>

<style>
.man_power_req .suim_form .section_group:nth-child(2) .section_title {
  border: none;
  margin: 0 !important;
}

.man_power_req .suim_form .section_group:nth-child(3) .section_title,
.man_power_req .suim_form .section_group:nth-child(5) .section_title {
  background: none;
  padding: 0;
  margin: 0 !important;
}
</style>
