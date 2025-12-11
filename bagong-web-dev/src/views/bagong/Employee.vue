<template>
  <div v-if="!isResume">
    <s-modal
      title="Create User"
      class="p-4"
      :display="false"
      ref="popupModal"
      hideButtons
      @submit="confirmDelete"
    >
      <s-input
        label="Password"
        class="mb-4 w-[240px]"
        kind="password"
        v-model="data.user.Password"
      ></s-input>
      <div class="flex flex-row-reverse gap-2">
        <s-button
          class="inline-block btn_primary text-white hover:bg-white hover:text-primary content-center"
          label="Cancel"
          @click="createCancel"
        ></s-button>
        <s-button
          class="inline-block bg-blue-500 text-white hover:bg-white hover:text-primary content-center"
          label="Create User"
          @click="createLogin"
        ></s-button>
      </div>
    </s-modal>
    <div class="w-full">
      <data-list
        class="card"
        ref="listControl"
        title="Employee"
        grid-config="/tenant/employee/gridconfig"
        form-config="/tenant/employee/formconfig"
        grid-read="/bagong/employee/get-employees"
        form-read="/bagong/employee/get"
        grid-mode="grid"
        grid-delete="/tenant/employee/delete"
        form-keep-label
        form-insert="/bagong/employee/save"
        form-update="/bagong/employee/save"
        :grid-fields="['Enable']"
        :form-tabs-edit="data.formTabs"
        :form-fields="['Dimension']"
        :init-app-mode="data.appMode"
        :form-default-mode="data.formMode"
        @formNewData="newRecord"
        @formEditData="editRecord"
        @preSave="onPreSave"
        @formRecordChange="onFormRecordChange"
        @formFieldChange="onFormFieldChange"
        @formModeChanged="onFormModeChanged"
        stay-on-form-after-save
        :grid-hide-new="!profile.canCreate"
        :grid-hide-edit="!profile.canUpdate"
        :grid-hide-delete="!profile.canDelete"
        :grid-custom-filter="data.customFilter"
        @alter-grid-config="alterGridConfig"
      >
        <template #grid_header_search>
          <grid-header-filter
            ref="gridHeaderFilter"
            v-model="data.customFilter"
            hideAll
            customTextLabel="Search"
            :fieldsText="['_id', 'Name', 'EmployeeNo', 'Email']"
            @initNewItem="initNewItemFilter"
            @preChange="changeFilter"
            @change="refreshGrid"
          >
            <template #filter_1="{ item }">
              <s-input
                class="w-[200px] filter-text"
                label="Search"
                v-model="item.Text"
              />
            </template>
            <template #filter_2="{ item }">
              <div class="min-w-[200px]">
                <s-input
                  label="Employment type"
                  kind="text"
                  class="w-[200px]"
                  v-model="item.EmploymentType"
                  :allow-add="false"
                  use-list
                  :items="[
                    'PERMANENT',
                    'CONTRACT',
                    'OUTSOURCE',
                    'EXTERNAL',
                    'PROBATION',
                    'CANDIDATE',
                  ]"
                />
              </div>
              <div class="min-w-[300px]">
                <dimension-editor
                  multiple
                  :default-list="profile.Dimension"
                  v-model="item.Dimension"
                  :required-fields="[]"
                  :dim-names="['Site', 'CC']"
                  :custom-labels="{
                    CC: 'CC / Department',
                  }"
                ></dimension-editor>
              </div>
            </template>
          </grid-header-filter>
        </template>
        <template #form_buttons_1="{ item }">
          <S-Button
            class="btn_success"
            label="create login"
            icon="account-plus"
            @click="showModal(item)"
            :disabled="item.IsLogin"
          ></S-Button>
        </template>
        <!-- <template #form_tab_Employee_Data="{ item }">
        <EmployeeDetail v-model="item.Detail"></EmployeeDetail>
      </template> -->
        <template #form_tab_Employee_Data="{ item }">
          <EmployeeData ref="employeeData" v-model="item.Detail" :employeeID="item._id"></EmployeeData>
        </template>
        <template #form_input_Dimension="{ item }">
          <dimension-editor
            v-model="item.Dimension"
            :default-list="profile.Dimension"
          ></dimension-editor>
        </template>
        <template #form_tab_Employee_Status="{ item }">
          <EmployeeStatus v-model="item.Detail"></EmployeeStatus>
        </template>
        <template #form_tab_Employee_Loan="{ item }">
          <EmployeeLoan :EmployeeID="item.Detail.EmployeeID"></EmployeeLoan>
        </template>
        <template #form_tab_Employee_Documents="{ item }">
          <Checklist
            v-model="item.Detail.Checklists"
            :checklist-id="data.journalTye.ChecklistTemplate"
            :hide-fields="['Done', 'PIC', 'Expected', 'Actual', 'Notes']"
            is-key-secured
            :attch-kind="'DOCUMENTS'"
            :attch-ref-id="item._id"
            :attch-tag-prefix="'DOCUMENTS'"
          />
        </template>
        <template #form_tab_Attachment="{ item }">
          <EmployeeAttachment
            :EmployeeID="item.Detail"
            v-model="item.Attachment"
          ></EmployeeAttachment>
          <!-- <grid-attachment :siteEntryAssetID="item._id" gridConfig="/bagong/employee_attachment/gridconfig"
          :gridFields="['FileName', 'UploadDate', 'URI']" v-model="item.Attachment"></grid-attachment> -->
        </template>
      </data-list>
    </div>
  </div>
  <div v-else>
    <s-card title="Resume" class="w-full bg-white suim_datalist" hide-footer>
      <EmployeeData
        v-model="data.resumeRecord"
        :jobIDs="data.jobIDs"
        :isResume="isResume"
        @saveForm="onSaveResumeForm"
        @submitForm="onSubmitResumeForm"
        @updateJobIds="(jobids) => (data.jobIDs = jobids)"
      ></EmployeeData>
    </s-card>
  </div>
</template>

<script setup>
import { authStore } from "@/stores/auth";
import { reactive, ref, inject, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton, SModal, SCard } from "suimjs";
import Checklist from "@/components/common/Checklist.vue";
import helper from "@/scripts/helper.js";

// import EmployeeDetail from "./widget/EmployeeDetail.vue";
import EmployeeData from "./widget/EmployeeData.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import EmployeeStatus from "./widget/EmployeeStatus.vue";
import EmployeeLoan from "./widget/EmployeeLoan.vue";
import EmployeeAttachment from "./widget/EmployeeAttachment.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";

// import GridAttachment from "@/components/common/GridAttachment.vue";
import moment from "moment";

const route = useRoute();

const isResume = route.query?.id == "Resume";

layoutStore().name = "tenant";

const FEATUREID = "Employee";
const profile = authStore().getRBAC(FEATUREID);
const auth = authStore();

const axios = inject("axios");

const listControl = ref(null);
const popupModal = ref(null);
const employeeData = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  user: {
    Password: "",
  },
  record: {},
  resumeRecord: {
    EmployeeID: "",
    Email: auth.appData?.Email,
    FamilyMembers: [],
  },
  jobIDs: [],
  formTabs: [
    "General",
    "Employee Data",
    "Employee Status",
    "Employee Loan",
    "Employee Documents",
    "Attachment",
  ],
  customFilter: null,
  journalTye: {},
  assetSignature: {},
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.IsActive = true;
  record.JoinDate = new Date().toISOString();
  record.IsLogin = false;
  data.record = record;

  openForm(record);
}

function editRecord(r) {
  data.record = r;
  openForm(r);
}

function openForm(record) {
  util.nextTickN(2, () => {
    if (record.Detail) {
      record.Detail.Age = "";
      record.Detail.JoinDate = record.JoinDate ?? "";
    }
    listControl.value.setFormFieldAttr("_id", "readOnly", true);
  });
}
function initNewItemFilter(item) {
  item.Dimension = [];
  item.EmploymentType = "";
  item.Department = "";
  item.Text = "";
}
function changeFilter(item, filters) {
  if (item.Text != "") {
    filters.push({
      Op: "$or",
      Items: ["_id", "Name", "EmployeeNo", "Email"].map((e) => {
        return {
          Op: "$contains",
          Field: e,
          Value: [item.Text],
        };
      }),
    });
  }
  if (item.EmploymentType) {
    filters.push({
      Op: "$contains",
      Field: "EmploymentType",
      Value: [item.EmploymentType],
    });
  }
  helper.genFilterDimension(item.Dimension).forEach((e) => {
    filters.push(e);
  });
}
function refreshGrid() {
  listControl.value.refreshGrid();
}
function showModal(d) {
  data.user = {
    _id: d._id,
    Email: d.Email,
    Name: d.Name,
    Password: d.Password,
  };
  popupModal.value.show();
}

function createLogin() {
  popupModal.value.hide();
  let payload = data.user;
  const url = "/bagong/employee/create-user";
  axios.post(url, payload).then(
    (r) => {
      util.showInfo("User has been created");
      data.record.UserID = payload._id;
      data.record.IsLogin = true;
      listControl.value.submitForm(
        data.record,
        () => {},
        () => {}
      );
    },
    (e) => {
      util.showError(e);
    }
  );
}

function createCancel() {
  popupModal.value.hide();
}

function onPreSave(record) {
  if (record.Detail) {
    record.Detail.PostCode = record.Detail.PostCode
      ? record.Detail.PostCode.toString()
      : "";
    record.Detail.WorkingPeriod = record.Detail.WorkingPeriod
      ? record.Detail.WorkingPeriod.toString()
      : "";
    record.Detail.Age = record.Detail.Age ? record.Detail.Age.toString() : "";
  }
  saveAsset();
}

function onFormFieldChange(name, v1, v2, old, record) {
  if (name == "JoinDate") {
    if (record.JoinDate != "0001-01-01T00:00:00Z") {
      var joinDate = moment(record.JoinDate);
      var nowDate = moment(new Date());
      var diff = nowDate.diff(joinDate, "years");
      record.Detail.WorkingPeriod = diff;
    }
  }
}

function onSaveResumeForm(record, cbSuccess, cbError) {
  const mapRecord = () => {
    const { Name, Email, ...rest } = record;
    return {
      Name,
      Email,
      Detail: { ...rest },
      Age: `${rest.Age}`,
    };
  };
  const params = mapRecord();
  axios.post("/bagong/employee/save-employee-resume", params).then(
    (r) => {
      cbSuccess();
      record.EmployeeID = r.data._id;
    },
    (e) => {
      util.showError(e);
      cbError();
    }
  );
}

function onSubmitResumeForm() {
  const payload = {
    EmployeeID: data.resumeRecord.EmployeeID,
    JobId: [...data.jobIDs],
  };
  axios.post("/hcm/tracking/apply-applicant", payload).then(
    (r) => {
      util.showInfo("Your application has been submitted");
    },
    (e) => {
      util.showError(e);
    }
  );
}
function getJurnalType() {
  axios.post("/hcm/journaltype/get", ["CandidateResume"]).then(
    (r) => {
      data.journalTye = r.data;
    },
    (e) => {
      data.journalTye = {};
      util.showError(e);
    }
  );
}
function alterGridConfig(cfg) {
  cfg.fields.splice(
    1,
    0,
    helper.gridColumnConfig({
      field: "EmployeeNo",
      label: "Employee No.",
      kind: "text",
    })
  );
}

async function saveAsset() {
  const record = employeeData.value.getDataAssestSignature();
  console.log("record: ",record)
  if (!record.OriginalFileName) {
    return 
  }

  const DataRecord = { ...record };
  delete DataRecord.OriginalFileName;
  delete DataRecord.ContentType;
  delete DataRecord.ReadOnly;
  delete DataRecord.Content;

  const param = {
    Content: record.Content,
    Asset: {
      _id: record._id,
      OriginalFileName: record.OriginalFileName,
      ContentType: record.ContentType,
      Kind: "Employee Signature",
      RefID: data.record._id,
      Data: DataRecord,
      URI: record.URI,
      NewFileName: record.NewFileName,
    },
  };

  // console.log("param signature: ",param)
  
  try {
    await axios.post(`/asset/write-asset`, param);
  } catch (error) {
    util.showError(error);
  }
}

watch(
  () => route.query.id,
  (nv) => {
    location.reload();
  }
);
onMounted(() => {
  getJurnalType();
});
</script>
