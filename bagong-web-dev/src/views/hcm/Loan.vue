<template>
  <div class="w-full">
    <data-list
      v-if="data.appMode != 'preview'"
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/hcm/loan/gridconfig"
      form-config="/hcm/loan/formconfig"
      grid-read="/hcm/loan/gets"
      form-read="/hcm/loan/get"
      grid-mode="grid"
      grid-delete="/hcm/loan/delete"
      form-keep-label
      form-insert="/hcm/loan/save"
      form-update="/hcm/loan/save"
      :grid-fields="['Status']"
      :form-fields="[
        'Dimension',
        'LoanPurpose',
        'LoanPurposeSpecify',
        'ApprovedLoan',
        'ApprovedInstallment',
        'ApprovedLoanPeriod',
        'Department',
        'Position',
        'WorkLocation',
        'EmployeeStatus'

      ]"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :formHideSubmit="!['', 'DRAFT', 'READY'].includes(data.record.Status)"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
      stay-on-form-after-save
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @alterFormConfig="alterFormConfig"
      @alterGridConfig="alterGridConfig"
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
        <template v-if="mode !== 'new'">
          <s-button
            :disabled="inSubmission || loading"
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Preview"
            @click="data.appMode = 'preview'"
          ></s-button>
        </template>
      </template>
      <template #form_input_Position="p">
        <s-input
          :key="p.item.Position + p.item.EmployeeID"
          label="Position"
          v-model="p.item.Position"
          use-list
          lookup-key="_id"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
          :lookup-labels="['Name']"
          read-only
        />
      </template>
      <template #form_input_Department="p">
        <s-input
          :key="p.item.Department + p.item.EmployeeID"
          label="Department"
          v-model="p.item.Department"
          use-list
          lookup-key="_id"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=DME"
          :lookup-labels="['Name']"
          read-only
        />
      </template>
      <template #form_input_WorkLocation="p">
        <s-input
          :key="p.item.WorkLocation + p.item.EmployeeID"
          label="Work Location"
          v-model="p.item.WorkLocation"
          use-list
          lookup-key="_id"
          lookup-url="/bagong/sitesetup/find"
          :lookup-labels="['Name']"
          read-only
        />
      </template>
      <template #form_input_EmployeeStatus="p">
        <s-input
          :key="p.item.EmployeeStatus + p.item.EmployeeID"
          label="Employee status"
          v-model="p.item.EmployeeStatus"
          use-list
          lookup-key="_id"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=ESM"
          :lookup-labels="['Name']"
          read-only
        />
      </template>
      <template #form_input_LoanPurpose="{ item, config }">
        <div>
          <s-input
            keep-label
            :label="config.label"
            v-model="item.LoanPurpose"
            :read-only="readOnly"
            use-list
            lookup-url="/tenant/masterdata/find?MasterDataTypeID=LoanPurpose"
            lookup-key="_id"
            :lookup-labels="['_id', 'Name']"
            :lookupSearchs="['_id', 'Name']"
          ></s-input>
          <s-input
            v-if="item.LoanPurpose === 'Others'"
            :read-only="readOnly"
            v-model="data.LoanPurposeSpecify"
            label="Please specify purpose"
            class="w-full mt-2"
          ></s-input>
        </div>
      </template>
      <template #form_input_ApprovedLoan="{ item, config }">
        <s-input
          v-if="['READY'].includes(item.Status)"
          keep-label
          :read-only="!['READY'].includes(data.record.Status)"
          v-model="item.ApprovedLoan"
          kind="number"
          :label="config.label"
          class="w-full"
          @change="
            (_, v1) => {
              if (item.ApprovedLoanPeriod === 0) {
                item.ApprovedInstallment = 0;
              } else {
                item.ApprovedInstallment = parseFloat(
                  (v1 / item.ApprovedLoanPeriod).toFixed(2)
                );
              }
            }
          "
        ></s-input>
        <template v-else>&nbsp;</template>
      </template>
      <template #form_input_ApprovedInstallment="{ item, config }">
        <s-input
          v-if="['READY'].includes(item.Status)"
          keep-label
          :read-only="!['READY'].includes(data.record.Status)"
          v-model="item.ApprovedInstallment"
          kind="number"
          :label="config.label"
          class="w-full"
        ></s-input>
        <template v-else>&nbsp;</template>
      </template>
      <template #form_input_ApprovedLoanPeriod="{ item, config }">
        <s-input
          v-if="['READY'].includes(item.Status)"
          keep-label
          :read-only="!['READY'].includes(data.record.Status)"
          v-model="item.ApprovedLoanPeriod"
          kind="number"
          :label="config.label"
          class="w-full"
          @change="
            (_, v1) => {
              if (v1 === 0) {
                item.ApprovedInstallment = 0;
              } else {
                item.ApprovedInstallment = parseFloat(
                  (item.ApprovedLoan / v1).toFixed(2)
                );
              }
            }
          "
        ></s-input>
        <template v-else>&nbsp;</template>
      </template>
      <!-- <template #form_input_LoanPurposeSpecify="{ item, config }">
        <div>
          <s-input
            v-if="item.LoanPurpose === 'Others'"
            :read-only="readOnly"
            v-model="data.LoanPurposeSpecify"
            label="Please specify purpose"
            class="w-full"
          ></s-input>
        </div>
      </template> -->
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
      reload="1"
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
import LogTrx from "@/components/common/LogTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import helper from "@/scripts/helper.js";
import moment from "moment";
import PreviewReport from "@/components/common/PreviewReport.vue";

layoutStore().name = "tenant";
const featureID = "LoanTransaction";

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
  titleForm: "Loan",
  record: {
    _id: "",
    RequestDate: new Date(),
    Dimension: [],
    Status: "",
  },
  jType: "LOAN",
  journalType: {},
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Loan`;
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
  data.titleForm = `Edit Loan | ${record._id}`;
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
function getDetailedFromNow(date) {
  const now = moment();
  const targetDate = moment(date);

  const years = now.diff(targetDate, "years");
  targetDate.add(years, "years");

  const months = now.diff(targetDate, "months");
  targetDate.add(months, "months");

  const days = now.diff(targetDate, "days");

  return `${years} years, ${months} months, ${days} days ago`;
}
function getDetailEmployee(id, record) {
  if (!id) {
    return 
  }
  axios.post("/tenant/employee/get", [id]).then(
    (r) => {
      util.nextTickN(2, () => {
        record.EmployeeName = r.data.Name;
        record.WorkLocation = r.data.Dimension.find(o => o.Key === 'Site').Value;
        record.Dimension = r.data.Dimension;

        const url = "/bagong/employeedetail/find?EmployeeID=" + id;
        axios.post(url).then(
          (rr) => {
            util.nextTickN(2, () => {
              if (rr.data.length > 0) {
                record.Department = rr.data[0].Department;
                record.Position = rr.data[0].Position;
                record.EmployeeStatus = rr.data[0].EmployeeStatus;
                record.NIK = rr.data[0].EmployeeNo;
                record.MobilePhoneNumber = rr.data[0].Phone;
                record.PeriodOfEmployement = getDetailedFromNow(
                  r.data.JoinDate
                );
                record.Salary = util.formatMoney(rr.data[0].BasicSalary);
                console.log(
                  moment(r.data.JoinDate).fromNow(),
                  rr.data[0].BasicSalary
                );
              }
            });
          },
          (e) => {
            util.showError(e);
          }
        );
      });
    },
    (e) => {
      util.showError(e);
    }
  );
}
function onFormFieldChange(name, v1, v2, old, record) {
  if (name === "EmployeeID") {
    getDetailEmployee(v1, record);
  }
  if (name === "LoanApplication") {
    if (record.LoanPeriod !== 0) {
      record.Installment = v1 / record.LoanPeriod;
    }
  }
  if (name === "LoanPeriod") {
    record.Installment = 0;
    if (v1 !== 0) {
      record.Installment = record.LoanApplication / v1;
    }
  }
  if (name === "JournalTypeID") {
    getJurnalType(v1, record);
  }
}
function onControlModeChanged(mode) {
  data.appMode = mode;
  if (mode === "grid") {
    data.titleForm = "Loan";
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
function alterFormConfig(cfg) {}
function alterGridConfig(cfg) {
  cfg.fields.splice(
    2,
    0,
    helper.gridColumnConfig({ field: "EmployeeNIK", label: "NIK" })
  );
  cfg.fields.splice(
    3,
    0,
    helper.gridColumnConfig({ field: "EmployeeName", label: "Employee name" })
  );
  return cfg;
}
</script>
