<template>
  <div class="w-full">
    <data-list
      v-if="data.appMode != 'preview'"
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/hcm/leavecompensation/gridconfig"
      form-config="/hcm/leavecompensation/formconfig"
      grid-read="/hcm/leavecompensation/gets"
      form-read="/hcm/leavecompensation/get"
      grid-mode="grid"
      grid-delete="/hcm/leavecompensation/delete"
      form-keep-label
      form-insert="/hcm/leavecompensation/save"
      form-update="/hcm/leavecompensation/save"
      :grid-fields="['Status']"
      :form-fields="['Dimension', 'ApprovedAmount']"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
      stay-on-form-after-save
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
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
        <div>
          <s-button
            class="btn_primary"
            label="Close"
            icon="Close"
            @click="onClosingJob(item)"
          ></s-button>
        </div>
        <template v-if="mode !== 'new'">
          <s-button
            :disabled="inSubmission || loading"
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Preview"
            @click="data.appMode = 'preview'"
          ></s-button>
        </template>
      </template>
      <template #form_input_ApprovedAmount="{ item }">
        <s-input label="Approved amount" v-model="item.ApprovedAmount" kind="number" :read-only="!['READY'].includes(item.Status)" />
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
    </data-list>

    <PreviewReport
      v-if="data.appMode == 'preview'"
      class="card w-full"
      title="Preview"
      @close="closePreview"
      :disable-print="helper.isDisablePrintPreview(data.record.Status)"
      :SourceType="data.jType"
      :SourceJournalID="data.record._id"
      reload=1
    >
      <template #buttons="props">
        <div class="flex gap-[1px] mr-2">
          <form-buttons-trx
            :disabled="inSubmission || loading"
            :status="data.record.Status"
            :journal-id="data.record._id"
            :posting-profile-id="data.record.PostingProfileID"
            :journal-type-id="data.jType"
            :trx-type="data.record.TransactionType"
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
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import { DataList, util, SForm, SInput, SButton, loadFormConfig } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import StatusText from "@/components/common/StatusText.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";
const featureID = "LeaveCompensation";

const profile = authStore().getRBAC(featureID);
const headOffice = layoutStore().headOfficeID;
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const listControl = ref(null);
const axios = inject("axios");
const auth = authStore();

const data = reactive({
  isPreview: false,
  appMode: "grid",
  formMode: "edit",
  titleForm: "Leave Compensation",
  record: {
    _id: "",
    RequestDate: new Date(),
    Dimension: [],
    Status: "",
  },
  jType: "LEAVECOMPENSATION",
  journalType: {},
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Leave Compensation`;
  record._id = "";
  record.CompanyID = auth.appData.CompanyID;
  record.RequestDate = new Date();
  record.InventDimTo = {};
  record.Dimension = [];
  record.Status = "DRAFT";
  data.record = record;
  openForm(record);
}

function editRecord(record) {
  console.info
  data.formMode = "edit";
  data.titleForm = `Edit Leave Compensation | ${record._id}`;
  if (!record.CompanyID) {
    record.CompanyID = auth.appData.CompanyID;
  }
  data.record = record;
  getDetailEmployee(data.record.RequestorID, data.record);

  openForm(record);
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr(
      "_id",
      "hide",
      data.formMode == "new" ? true : false
    );
    const el = document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(1) > div > div > div:nth-child(1)"
    );
    if (record._id == "") {
      el.style.display = "none";
    } else {
      el.style.display = "block";
    }
  });
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("PostingProfileID", "readOnly", true);
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
  });
}
function getDetailEmployee(id, record) {
  axios.post("/bagong/employee/get", [id]).then(
    (r) => {
      util.nextTickN(2, () => {
        record.RequestorName = r.data.Name;
        record.Dimension = r.data.Dimension;
      });
    },
    (e) => {
      util.showError(e);
    }
  );
}
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
function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
    case "RequestorID":
      getDetailEmployee(v1, record);
      break;
  }
}
function onControlModeChanged(mode) {
  data.appMode = mode;
  if (mode === "grid") {
    data.titleForm = "Leave Compensation";
  }
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}

function closePreview() {
  data.appMode = "grid";
}
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
function setModeGrid() {
  listControl.value.setControlMode("grid");
  listControl.value.refreshList();
}
</script>
