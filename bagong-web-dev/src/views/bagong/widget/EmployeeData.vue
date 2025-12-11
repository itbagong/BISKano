<template>
  <div v-if="props.isResume" class="suim_form mb-10">
    <div class="mb-2 flex header justify-end">
      <s-button
        v-if="props.isResume"
        @click="onApplyForm"
        class="btn_primary"
        label="Apply"
      ></s-button>
    </div>
    <div class="form_inputs">
      <div class="section_group_container">
        <div class="section_group col flex-col grow">
          <div class="section grow">
            <div class="title section_title">Job Vacancy</div>
            <div class="flex flex-col gap-4">
              <div class="w-full items-start gap-2 grid gridCol3">
                <s-input
                  useList
                  label="Job Vacancy Title"
                  lookup-url="/hcm/manpowerrequest/get-available"
                  lookup-key="_id"
                  :lookup-labels="['Name']"
                  :lookup-searchs="['_id', 'Name']"
                  v-model="data.jobIDs"
                  class="min-w-[180px]"
                  multiple
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div>
    <s-form
      ref="formControl"
      :config="data.config"
      v-model="data.value"
      @fieldChange="onFieldChange"
      keep-label
      :hide-cancel="props.isResume"
      :tabs="props.isResume ? ['General', 'Documents'] : ['General']"
      @submitForm="onSaveForm"
    >
      <template #tab_Documents="props">
        <Checklist
          v-model="props.item.Checklists"
          :checklist-id="data.journalTye.ChecklistTemplate"
          :hide-fields="['Done', 'PIC', 'Expected', 'Actual', 'Notes']"
          is-key-secured
          :attch-kind="'DOCUMENTS'"
          :attch-ref-id="props.item._id"
          :attch-tag-prefix="'DOCUMENTS'"
        />
      </template>
      <template v-if="!props.isResume" #buttons>&nbsp;</template>
      <template v-if="props.isResume" #section_General_header="{}">
        <!-- <div class="flex justify-end mb-2">
          <s-button
            @click="onSaveFormControl"
            class="btn_primary"
            icon="content-save"
            label="Save"
          ></s-button>
        </div> -->
        <div class="title section_title">General</div>
      </template>
      <template #input_FamilyMembers="{ item }">
        <DataList
          ref="gridFamily"
          hide-title
          no-gap
          grid-editor
          grid-hide-search
          grid-hide-sort
          grid-hide-refresh
          grid-hide-detail
          grid-hide-select
          grid-no-confirm-delete
          init-app-mode="grid"
          grid-mode="grid"
          new-record-type="grid"
          grid-config="/bagong/employee/familymembers/gridconfig"
          :grid-fields="data.fields"
          grid-auto-commit-line
          @grid-row-add="newRecord"
          @grid-row-delete="onGridRowDelete"
          @gridRefreshed="onGridRefreshed"
          @alterGridConfig="alterGridConfig"
        ></DataList>
      </template>
      <template #input_Signature="{ item }">
        <label class="input_label">
          <div>Signature</div>
        </label>
        <div class="input_field">
          <input class="w-[100px]" type="file" ref="fileInputSignature" @change="handleFileUploadSignature" accept="image/png, image/jpeg"/>
          <span>{{ data.assetSignature.FileName ||  'No data' }}</span>
        </div>
      </template>
    </s-form>
  </div>
</template>

<script setup>
import { reactive, onMounted, ref, inject, watch } from "vue";
import { authStore } from "@/stores/auth";
import {
  DataList,
  SForm,
  SButton,
  SInput,
  loadFormConfig,
  util,
  createFormConfig,
} from "suimjs";
import Checklist from "@/components/common/Checklist.vue";

import moment from "moment";
const axios = inject("axios");
const auth = authStore();

const gridFamily = ref(null);
const formControl = ref(null);
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  employeeID: { type: String, default: "" },
  isResume: { type: Boolean, default: () => false },
  jobIDs: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
  updateJobIds: null,
  saveForm: null,
  submitForm: null,
});
const data = reactive({
  config: {},
  value: props.modelValue,
  records: props.modelValue.FamilyMembers,
  jobIDs: props.jobIDs,
  fields: ["Name", "PlaceOfBirth", "DateofBirth", "Gender"],
  journalTye: {},
  assetSignature: {},
});

function onSaveFormControl() {
  formControl.value.submit();
}
function onSaveForm(record, cbSuccess, cbError) {
  emit("saveForm", record, cbSuccess, cbError);
}

function onApplyForm() {
  emit("submitForm");
}
function getJurnalType() {
  axios.post( "/hcm/journaltype/get", ['CandidateResume']).then(
    (r) => {
      data.journalTye = r.data; 
    },
    (e) => {
      data.journalTye = {};
      util.showError(e)
    }
  );
}
onMounted(() => {
  getJurnalType();
  generateConfig();
  getDataAssets();
});

function generateConfig() {
  const cfg = createFormConfig("", true);
  if (!props.isResume) {
    // for employee
    cfg.addSection("", true).addRowAuto(
      3,
      {
        field: "EmployeeNo",
        label: "Employee No",
        kind: "text",
        required: true,
      },
      {
        field: "IdentityCardNo",
        label: "Identity Card No",
        kind: "text",
        required: true,
      },
      {
        field: "FamilyCardNo",
        label: "Family Card No",
        kind: "text",
        required: true,
      },
      {
        field: "PlaceOfBirth",
        label: "Place Of Birth",
        kind: "text",
        required: true,
      },
      {
        field: "DateOfBirth",
        label: "Date Of Birth",
        kind: "date",
        required: true,
      },
      {
        field: "Gender",
        label: "Gender",
        kind: "text",
        required: true,
        useList: true,
        lookupKey: "_id",
        lookupLabels: ["Name"],
        lookupSearch: ["Name"],
        lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=GEME",
      },
      {
        field: "Religion",
        label: "Religion",
        kind: "text",
        required: true,
        useList: true,
        lookupKey: "_id",
        lookupLabels: ["Name"],
        lookupSearch: ["Name"],
        lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=RME",
      },
      {
        field: "MaritalStatus",
        label: "Marital Status",
        kind: "text",
        required: true,
        useList: true,
        items: ["SINGLE", "MARRIED", "DIVORCE"],
      },
      {
        field: "Age",
        label: "Age",
        kind: "text",
      },
      { hide: true }
    );
    data.config = cfg.generateConfig();
  } else {
    // for resume
    // cfg.addSection("Job Vacancy", true).addRowAuto(
    //   3,
    //   {
    //     field: "JobVacancyTitle",
    //     label: "Job Vacancty Title",
    //     kind: "text",
    //     required: true,
    //     useList: true,
    //     multiple: true,
    //     lookupKey: "_id",
    //     lookupLabels: ["Name"],
    //     lookupSearch: ["_id", "Name"],
    //     lookupUrl: "/hcm/manpowerrequest/find",
    //   },
    //   { hide: true },
    //   { hide: true }
    // );
    // data.config = cfg.generateConfig();

    cfg.addSection("General", !props.isResume).addRowAuto(
      3,
      {
        field: "Name",
        label: "Name",
        kind: "text",
        required: true,
      },
      {
        field: "Email",
        label: "Email",
        kind: "text",
        required: true,
        disable: true,
        readOnly: true,
      },
      { hide: true }
    );
    data.config = cfg.generateConfig();

    cfg.addSection("Candidate Data", true).addRowAuto(
      3,
      {
        field: "IdentityCardNo",
        label: "Identity Card No",
        kind: "text",
        required: true,
      },
      {
        field: "FamilyCardNo",
        label: "Family Card No",
        kind: "text",
        required: true,
      },
      {
        field: "PlaceOfBirth",
        label: "Place Of Birth",
        kind: "text",
        required: true,
      },
      {
        field: "DateOfBirth",
        label: "Date Of Birth",
        kind: "date",
        required: true,
      },
      {
        field: "Gender",
        label: "Gender",
        kind: "text",
        required: true,
        useList: true,
        lookupKey: "_id",
        lookupLabels: ["Name"],
        lookupSearch: ["Name"],
        lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=GEME",
      },
      {
        field: "Religion",
        label: "Religion",
        kind: "text",
        required: true,
        useList: true,
        lookupKey: "_id",
        lookupLabels: ["Name"],
        lookupSearch: ["Name"],
        lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=RME",
      },
      {
        field: "MaritalStatus",
        label: "Marital Status",
        kind: "text",
        required: true,
        useList: true,
        items: ["SINGLE", "MARRIED", "DIVORCE"],
      },
      {
        field: "Age",
        label: "Age",
        kind: "text",
      },
      { hide: true }
    );
    data.config = cfg.generateConfig();
  }

  cfg.addSection("", true).addRowAuto(
    3,
    {
      field: "Address",
      label: "Address",
      kind: "text",
      required: true,
    },
    {
      field: "Province",
      label: "Province",
      kind: "text",
      required: true,
    },
    {
      field: "City",
      label: "City",
      kind: "text",
      required: true,
    },
    {
      field: "Subdistrict",
      label: "Subdistrict",
      kind: "text",
      required: true,
    },
    {
      field: "Village",
      label: "Village",
      kind: "text",
      required: true,
    },
    {
      field: "PostCode",
      label: "PostCode",
      kind: "text",
    },
    {
      field: "Domicile",
      label: "Domicile",
      kind: "text",
    },
    {
      field: "Phone",
      label: "Phone",
      kind: "text",
      required: true,
    },
    {
      field: "EmergencyPhone",
      label: "Emergency Phone",
      kind: "text",
    },
    { hide: true }
  );
  data.config = cfg.generateConfig();

  cfg.addSection("", true).addRowAuto(
    3,
    {
      field: "LastEducation",
      label: "Last Education",
      kind: "text",
      required: true,
    },
    {
      field: "Major",
      label: "Major",
      kind: "text",
      required: true,
    },
    {
      field: "SchoolOrUniversityName",
      label: "Name of SchoolOr/University",
      kind: "text",
      required: true,
    },
    {
      field: "WorkingExperience",
      label: "Working Experience (Years)",
      kind: "number",
      required: true,
    },
    { hide: true },
    { hide: true }
  );
  data.config = cfg.generateConfig();

  cfg.addSection("", true).addRowAuto(
    3,
    {
      field: "BankAccount",
      label: "Bank Account",
      kind: "text",
    },
    {
      field: "BankAccountName",
      label: "Bank Account Name",
      kind: "text",
    },
    {
      field: "BankAccountNo",
      label: "Bank Account No",
      kind: "text",
    },
    {
      field: "TaxIdentityNo",
      label: "NPWP No",
      kind: "text",
    },
    {
      field: "TaxIdentityName",
      label: "NPWP Name",
      kind: "text",
    },
    {
      field: "BPJSTK",
      label: "BPJS TK",
      kind: "text",
    },
    {
      field: "BPJSKES",
      label: "BPJS KES",
      kind: "text",
    },
    {
      field: "SIMType",
      label: "SIM Type",
      kind: "text",
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearch: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=SIMType",
    },
    {
      field: "SIMNo",
      label: "SIM No",
      kind: "text",
    },
    {
      field: "SIMIssueDate",
      label: "SIM Issue Date",
      kind: "date",
    },
    {
      field: "SIMExpirationDate",
      label: "SIM Expiration Date",
      kind: "date",
    },
    { hide: true }
  );
  data.config = cfg.generateConfig();

  cfg.addSection("", true).addRowAuto(
    3,
    {
      field: "BiologicalMotherName",
      label: "Biological Mother's Name",
      kind: "text",
      required: true,
    },
    {
      field: "Signature",
      label: "Signature",
      kind: "file",
      required: true,
    },
    // {
    //   field: "SpouseName",
    //   label: "Spouse Name",
    //   kind: "text",
    //   required: true,
    // },
    // {
    //   field: "SpousePlaceOfBirth",
    //   label: "Spouse Place of Birth",
    //   kind: "text",
    //   required: true,
    // },
    // {
    //   field: "SpouseDateOfBirth",
    //   label: "Spouse Date of Birth",
    //   kind: "date",
    //   required: true,
    // },
    // { hide: true },
    { hide: true }
  );
  data.config = cfg.generateConfig();

  cfg.addSection("FamilyMembers", true).addRowAuto(3, {
    field: "FamilyMembers",
    label: "FamilyMembers",
  });
  data.config = cfg.generateConfig();
  util.nextTickN(2, () => {
    if (props.isResume) {
      fetchData();
    }
  });
}

function fetchData() {
  axios.post("/bagong/employee/get-employee-resume").then(
    (r) => {
      data.value = r.data;
      data.value.Email = auth.appData?.Email;
      gridFamily?.value?.setGridRecords(data.value.FamilyMembers);
    },
    (e) => {
      util.showError(e);
    }
  );
}
function alterGridConfig(cfg) {}
function newRecord(r) {
  r.Name = "";
  // r.PlaceOfBirth = new Date();
  // r.DateofBirth = new Date();
  r.Gender = "";

  data.records.push(r);
  gridFamily.value.setGridRecords(data.records);

  data.value.FamilyMembers = data.records;
}

function onGridRowDelete(_, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  gridFamily.value.setGridRecords(data.records);
  data.value.FamilyMembers = data.records;
}

// function onGridRowFieldChanged(name, v1, v2, old, record) {
//     console.log("change:", record)
//     gridFamily.value.setGridRecord(
//         record,
//         gridFamily.value.getGridCurrentIndex()
//     );

//     gridFamily.value.setGridRecords(data.records);
// }
function onGridRefreshed() {
  if (data.records) {
    gridFamily.value.setGridRecords(data.records);
  }
}

function onFieldChange(name, v1, v2, old) {
  if (name == "DateOfBirth") {
    var birthDate = moment(v1);
    var nowDate = moment(new Date());
    var diff = nowDate.diff(birthDate, "years");
    data.value.Age = diff.toString();
  }
}

function handleFileUploadSignature(event) {
  const file = event.target.files[0];
  const reader = new FileReader();
  reader.readAsDataURL(file);

  reader.onloadend = () => {
    data.assetSignature["Content"] = reader.result.split(",")[1];
    data.assetSignature["OriginalFileName"] = file.name;
    data.assetSignature["FileName"] = file.name;
    data.assetSignature["ContentType"] = file.type;
    data.assetSignature["NewFileName"] = `signature/${data.assetSignature["RefID"]}.${file.name.split('.').pop().toLowerCase()}`;
  };
}

async function getDataAssets() {
  if(props.employeeID == ""){
    return
  }
  try {
    let param = {
      JournalType: "Employee Signature",
      JournalID: props.employeeID,
      Tags: [],
    };

    const resp = await axios.post("/asset/read-by-journal", param);
    
    if (resp.data.length > 0) {
      const respData = resp.data[0];
      const dtMap = {
        _id: respData._id,
        OriginalFileName: respData.OriginalFileName,
        Content: respData.Data.Content,
        ContentType: respData.ContentType,
        FileName: respData.OriginalFileName,
        UploadDate: respData.Data.UploadDate,
        Description: respData.Data.Description,
        URI: respData.URI,
        Kind: respData.Kind,
        Tags: respData.Tags,
        RefID: respData.RefID,
        ExpiredDate: respData.Data.ExpiredDate,
        Title: respData.Data.Title
      }
      data.assetSignature = dtMap;
    } else {
      setAssetSignature();
    }
  } catch (error) {
    util.showError(error);
  }
}

function setAssetSignature() {
  const r = {
    OriginalFileName: "",
    ContentType: "",
    Descriptions: "",
    UploadDate: new Date(),
    Tags: [],
    SameJournal: true,
    Kind: "Employee Signature",
    RefID: props.employeeID || "",
    NewFileName: ""
  }
  
  data.assetSignature = r;
}

function getDataAssestSignature() {
  return data.assetSignature
}

watch(
  () => data.value,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
watch(
  () => data.jobIDs,
  (nv) => {
    emit("updateJobIds", nv);
  },
  { deep: true }
);

defineExpose({
  getDataAssestSignature,
});
</script>

<style></style>
