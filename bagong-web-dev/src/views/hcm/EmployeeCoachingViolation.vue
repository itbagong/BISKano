<template>
  <div class="w-full">
    <data-list
      v-show="data.appMode != 'preview'"
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/hcm/coachingviolation/gridconfig"
      form-config="/hcm/coachingviolation/formconfig"
      grid-read="/hcm/coachingviolation/gets"
      form-read="/hcm/coachingviolation/get"
      grid-mode="grid"
      grid-delete="/hcm/coachingviolation/delete"
      form-keep-label
      form-insert="/hcm/coachingviolation/save"
      form-update="/hcm/coachingviolation/save"
      :grid-fields="['Status']"
      :form-fields="['Dimension', 'JournalTypeID', 'Position', 'Department', 'RequestorID', 'EmployeeID']"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :formHideSubmit="readOnly"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
      stay-on-form-after-save
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @alterGridConfig="alterGridConfig"
    >
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <!-- grid status -->
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_input_RequestorID="{item}">
        <s-input
          v-if="readOnly"
          class="w-full"
          label="Requestor name"
          use-list
          lookup-url="/tenant/employee/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.RequestorID"
          :read-only="readOnly || mode == 'view'"
        />
      </template>
      <template #form_input_EmployeeID="{item}">
        <s-input
          v-if="readOnly"
          class="w-full"
          label="Employee name"
          use-list
          lookup-url="/tenant/employee/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.EmployeeID"
          :read-only="readOnly || mode == 'view'"
        />
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
      <template #form_buttons_1="{ item, config, mode }">
        <div class="flex flex-row gap-2">
          <form-buttons-trx
            :disabled="inSubmission || loading"
            :status="item.Status"
            :journal-id="item._id"
            :posting-profile-id="item.PostingProfileID"
            :journal-type-id="data.jType"
            moduleid="hcm"
            :auto-post="!waitTrxSubmit"
            @preSubmit="trxPreSubmit"
            @postSubmit="trxPostSubmit"
            @errorSubmit="trxErrorSubmit"
          />
          <template v-if="profile.canUpdate && mode !== 'new'">
            <s-button
              :disabled="inSubmission || loading"
              class="bg-primary text-white font-bold w-full flex justify-center"
              label="Preview"
              @click="data.appMode = 'preview'"
            ></s-button>
          </template>
        </div>
      </template>
      <template #form_input_JournalTypeID="{ item, mode }">
        <s-input
          class="w-full"
          label="Journal type ID"
          use-list
          :lookup-url="renderURL(item.Type)"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.JournalTypeID"
          :read-only="readOnly || mode == 'view'"
          @change="(_, v1) => getJurnalType(v1)"
        />
      </template>
      <template #form_input_Position="{ item, mode }">
        <s-input
          :key="item.EmployeeID + item.Position"
          label="Position"
          v-model="item.Position"
          use-list
          lookup-key="_id"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
          :lookup-labels="['Name']"
          read-only
        />
      </template>
      <template #form_input_Department="{ item, mode }">
        <s-input
          :key="item.EmployeeID + item.Department"
          label="Department"
          v-model="item.Department"
          use-list
          lookup-key="_id"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=DME"
          :lookup-labels="['Name']"
          read-only
        />
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
      :VoucherNo="data.record.LedgerVoucherNo"
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
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import { DataList, util, SForm, SInput, SButton, loadFormConfig } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import LogTrx from "@/components/common/LogTrx.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import StatusText from "@/components/common/StatusText.vue";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";
const featureID = "CoachingViolation";

const profile = authStore().getRBAC(featureID);
const auth = authStore();

const headOffice = layoutStore().headOfficeID;
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const listControl = ref(null);
const axios = inject("axios");

const data = reactive({
  isPreview: false,
  appMode: "grid",
  formMode: "edit",
  titleForm: "Employee Coaching & Violation",
  record: {
    _id: "",
    RequestDate: new Date(),
    Dimension: [],
    Status: "",
  },
  jType: "COACHING",
  journalType: {},
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Employee Coaching & Violation`;
  record._id = "";
  record.RequestDate = new Date();
  record.InventDimTo = {};
  record.Dimension = [];
  record.Status = "DRAFT";
  record.CompanyID = auth.appData.CompanyID;
  data.record = record;
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Employee Coaching & Violation | ${record._id}`;
  if (!record.CompanyID) {
    record.CompanyID = auth.appData.CompanyID;
  }
  data.record = record;
  getDetailEmployee(data.record.EmployeeID, data.record);

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
  axios.post("/tenant/employee/get", [id]).then(
    (r) => {
      util.nextTickN(2, () => {
        record.EmployeeName = r.data.Name;
        record.Dimension = r.data.Dimension;
      });
    },
    (e) => {
      util.showError(e);
    }
  );
  const url = "/bagong/employeedetail/find?EmployeeID=" + id;
  axios.post(url).then(
    (r) => {
      util.nextTickN(2, () => {
        if (r.data.length > 0) {
          record.Department = r.data[0].Department;
          record.Position = r.data[0].Position;
          record.EmployeeNIK = r.data[0].EmployeeNo;
        }
      });
    },
    (e) => {
      util.showError(e);
    }
  );
}
const renderURL = (type) => {
  const params = new URLSearchParams({ TransactionType: type });
  const url = `/hcm/journaltype/find?${params.toString()}`;
  return url
}
function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "EmployeeID":
      getDetailEmployee(v1, record);
      break;
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
    case "Type":
    setTransactionType(v1, record);
      break;
  }
}
function setTransactionType(type, record) {
  util.nextTickN(2, () => {
    listControl.value.setFormLoading(true);
    axios
      .post(renderURL(type), {
        Take: 20,
        Sort: ["Name"],
        Select: ["Name", "_id", "PostingProfileID"],
      })
      .then(
        (r) => {
          if (r.data.length == 1) {
            data.journalType = r.data[0];
            record.JournalTypeID = r.data[0]._id;
            record.PostingProfileID = r.data[0].PostingProfileID;
          }
        },
        (e) => {
          util.showError(e);
        }
      )
      .finally(() => {
        listControl.value.setFormLoading(false);
      });
  });
}
function onControlModeChanged(mode) {
  data.appMode = mode;
  if (mode === "grid") {
    data.titleForm = "Employee Coaching & Violation";
  }
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
function alterGridConfig(cfg) {
  cfg.fields.splice(
    6,
    0,
    helper.gridColumnConfig({ field: "EmployeeNIK", label: "NIK" })
  );
  return cfg;
}
</script>
