<template>
  <div class="w-full">
    <data-list
      v-if="data.appMode != 'preview'"
      class="card"
      ref="listControl"
      title="Talent Development"
      grid-config="/hcm/talentdevelopment/gridconfig"
      form-config="/hcm/talentdevelopment/formconfig"
      grid-read="/hcm/talentdevelopment/gets"
      form-read="/hcm/talentdevelopment/get"
      grid-mode="grid"
      grid-delete="/hcm/talentdevelopment/delete"
      form-keep-label
      form-insert="/hcm/talentdevelopment/save"
      form-update="/hcm/talentdevelopment/save"
      :grid-fields="[
        'Status',
        'AssessmentStatus',
        'ActingSKStatus',
        'PermanentSKStatus',
      ]"
      :form-tabs-edit="
        !['POSTED'].includes(data.record.Status)
          ? ['General', 'Benefit Detail']
          : ['General', 'Benefit Detail', 'Tracking']
      "
      :form-tabs-view="
        !['POSTED'].includes(data.record.Status)
          ? ['General', 'Benefit Detail']
          : ['General', 'Benefit Detail', 'Tracking']
      "
      :form-fields="[
        '_id',
        'Dimension',
        'AssesmentResult',
        'JournalTypeID',
        'Assesment',
        'JoinedDate',
      ]"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @formFieldChange="onFormFieldChange"
      @post-save="onFormPostSave"
      stay-on-form-after-save
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="customFilter"
      @alterGridConfig="alterGridConfig"
    >
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <!-- grid filter -->
      <template #grid_header_search="{ config }">
        <s-input
          v-model="data.query.SubmissionType"
          kind="text"
          label="Submission Type"
          class="w-full"
          use-list
          :items="['Promotion', 'Rotation', 'Demotion', 'Salary Change', 'POH']"
          @change="refreshData"
        ></s-input>
        <s-input
          v-model="data.query.EmployeeName"
          kind="text"
          label="Employee Name"
          class="w-full"
          use-list
          lookup-url="/tenant/employee/find"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          v-model="data.query.Status"
          kind="text"
          label="Status"
          class="w-full"
          use-list
          :items="['DRAFT', 'SUBMITTED', 'READY', 'REJECTED', 'POSTED']"
          @change="refreshData"
        ></s-input>
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
          <template v-if="mode !== 'new'">
            <s-button
              :disabled="inSubmission || loading"
              class="bg-primary text-white font-bold w-full flex justify-center"
              label="Preview"
              @click="data.appMode = 'preview'"
            ></s-button>
          </template>
        </div>
      </template>
      <!-- form input -->
      <template #form_input__id="{ item, config }">
        <div>
          <s-input
            :label="config.label"
            keep-label
            v-model="item._id"
            read-only
          />
        </div>
      </template>
      <template v-if="readOnly" #form_input_Assesment="{ item, config }">
        <label class="input_label">
          <div>Tracking assesment</div>
        </label>
        <div>{{ item.Assesment }}</div>
      </template>
      <template #form_input_JournalTypeID="{ item, mode }">
        <s-input
          class="w-full"
          label="Journal type ID"
          use-list
          :lookup-url="`/hcm/journaltype/find?TransactionType=${data.trxType}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.JournalTypeID"
          :read-only="readOnly || mode == 'view'"
          @change="
            (_, v) => {
              getJurnalType(v, item);
            }
          "
        />
      </template>
      <template #form_input_AssesmentResult="{ item, config }">
        <div>
          <s-input
            :label="config.label"
            v-show="item.SubmissionType == 'POH'"
            keep-label
            v-model="item.AssesmentResult"
          />
        </div>
      </template>
      <template #form_input_JoinedDate="{ item, config }">
        <div>
          <s-input
            :label="config.label"
            keep-label
            v-model="item.JoinedDate"
            kind="date"
            read-only
          />
        </div>
      </template>

      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          ref="dimenstionCtl"
          v-model="item.Dimension"
          :read-only="readOnly || mode == 'view'"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>

      <!-- form tab -->
      <template #form_tab_Benefit_Detail>
        <div class="suim_area_table">
          <table class="w-full table-auto suim_table">
            <thead name="grid_header">
              <tr class="border-b-[1px] border-slate-500">
                <th class="text-left">Description</th>
                <th class="text-left">Existing</th>
                <th class="text-left">New Propose</th>
              </tr>
            </thead>
            <tbody name="grid_body">
              <template v-for="(value, key) in data.descBenefitDetail">
                <tr
                  class="cursor-pointer border-b-[1px] border-slate-200 last:border-non hover:bg-slate-200 even:bg-slate-100"
                >
                  <td class="py-2">
                    <div class="w-[300px]">
                      {{ key.split(/(?=[A-Z])/).join(" ") }}
                    </div>
                  </td>
                  <td class="py-2">
                    <div class="w-[300px]">
                      <s-input
                        hide-label
                        :label="key"
                        v-if="data.benefitDetail.Existing[key]"
                        v-model="data.benefitDetail.Existing[key]"
                        :kind="value.kind"
                        :use-list="value.useList"
                        :lookup-url="
                          value.useList ? value.lookupUrl : undefined
                        "
                        lookup-key="_id"
                        :lookup-labels="['Name']"
                        read-only
                      />
                      <div v-else>
                        <div v-if="value.kind != 'number'">-</div>
                        <div v-else>0</div>
                      </div>
                    </div>
                  </td>
                  <td class="py-2">
                    <div class="w-[300px]">
                      <s-input
                        hide-label
                        :label="key"
                        v-model="data.benefitDetail.NewPropose[key]"
                        :kind="value.kind"
                        :use-list="value.useList"
                        :lookup-url="
                          value.useList ? value.lookupUrl : undefined
                        "
                        lookup-key="_id"
                        :lookup-labels="['Name']"
                        :lookup-searchs="['Name', '_id']"
                      />
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </template>
      <template #form_tab_Tracking="{ item, mode }">
        <tracking-talent-development
          :record-talent="item"
        ></tracking-talent-development>
      </template>
      <!-- grid status -->
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #grid_AssessmentStatus="{ item }">
        <div class="flex items-center">
          <log-trx
            :id="item.AssessmentID"
            v-if="helper.isShowLog(item.AssessmentStatus)"
          />
          <status-text :txt="item.AssessmentStatus" />
        </div>
      </template>
      <template #grid_ActingSKStatus="{ item }">
        <div class="flex items-center">
          <log-trx
            :id="item.ActingSKID"
            v-if="helper.isShowLog(item.ActingSKStatus)"
          />
          <status-text :txt="item.ActingSKStatus" />
        </div>
      </template>
      <template #grid_PermanentSKStatus="{ item }">
        <div class="flex items-center">
          <log-trx
            :id="item.PermanentSKID"
            v-if="helper.isShowLog(item.PermanentSKStatus)"
          />
          <status-text :txt="item.PermanentSKStatus" />
        </div>
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
            moduleid="hcm"
            :auto-post="!waitTrxSubmit"
            @preSubmit="trxPreSubmit"
            @postSubmit="trxPostSubmit"
            @errorSubmit="trxErrorSubmit"
          />
        </div>
      </template>
    </PreviewReport>
  </div>
</template>

<script setup>
import { authStore } from "@/stores/auth";
import { reactive, ref, inject, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";

import TrackingTalentDevelopment from "./widget/TalentDevelopment/TrackingTalentDevelopment.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import StatusText from "@/components/common/StatusText.vue";
import helper from "@/scripts/helper.js";
import PreviewReport from "@/components/common/PreviewReport.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import LogTrx from "@/components/common/LogTrx.vue";

layoutStore().name = "tenant";

const FEATUREID = "TalentDevelopmentSubmission";
const profile = authStore().getRBAC(FEATUREID);
const auth = authStore();

const axios = inject("axios");

const listControl = ref(null);
const dimenstionCtl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  query: {
    SubmissionType: "",
    EmployeeName: "",
    Status: "",
  },
  // itemsNIK: [],
  records: [],
  record: {
    Status: "DRAFT",
  },
  jType: "TALENTDEVELOPMENT",
  journalType: {},
  trxType: "",
  descBenefitDetail: {
    Department: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=DME",
      kind: "text",
    },
    Position: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=PTE",
      kind: "text",
    },
    Grade: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=GDE",
    },
    Group: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=GME",
      kind: "text",
    },
    SubGroup: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=SGE",
      kind: "text",
    },
    Site: {
      useList: true,
      lookupUrl: "/bagong/sitesetup/find",
      kind: "text",
    },
    PointOfHire: {
      useList: true,
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=PME",
      kind: "text",
    },
    BasicSalary: {
      useList: false,
      lookupUrl: null,
      kind: "number",
    },
    Allowance: {
      useList: false,
      lookupUrl: null,
      kind: "text",
    },
  },
  benefitDetail: {
    Existing: {},
    NewPropose: {},
  },
});

function alterGridConfig(cfg) {
  cfg.fields.splice(
    2,
    0,
    helper.gridColumnConfig({
      field: "EmployeeName",
      label: "Employee Name",
      kind: "text",
    })
  );
  cfg.fields.splice(
    3,
    0,
    helper.gridColumnConfig({ field: "NIK", label: "NIK", kind: "text" })
  );
  cfg.fields.splice(
    4,
    0,
    helper.gridColumnConfig({
      field: "POH",
      label: "Point of Hire",
      kind: "text",
    })
  );
  cfg.fields.splice(
    5,
    0,
    helper.gridColumnConfig({
      field: "EmployeeJoinedDate",
      label: "Joined Date",
      kind: "date",
    })
  );
  cfg.fields.splice(
    11,
    0,
    helper.gridColumnConfig({
      field: "AssessmentStatus",
      label: "Status Assessment",
      kind: "text",
    })
  );
  cfg.fields.splice(
    12,
    0,
    helper.gridColumnConfig({
      field: "ActingSKStatus",
      label: "Status SK Acting",
      kind: "text",
    })
  );
  cfg.fields.splice(
    13,
    0,
    helper.gridColumnConfig({
      field: "PermanentSKStatus",
      label: "Status SK Permanent",
      kind: "text",
    })
  );
}

let customFilter = computed(() => {
  const filters = [];
  if (data.query.SubmissionType !== null && data.query.SubmissionType !== "") {
    filters.push({
      Field: "SubmissionType",
      Op: "$contains",
      Value: [data.query.SubmissionType],
    });
  }
  if (data.query.NIK !== null && data.query.NIK !== "") {
    const find = data.records.find((v) => v.NIK.includes(data.query.NIK));
    if (find) {
      filters.push({
        Field: "EmployeeID",
        Op: "$contains",
        Value: [find.EmployeeID],
      });
    }
  }
  if (data.query.EmployeeName !== null && data.query.EmployeeName !== "") {
    filters.push({
      Field: "EmployeeID",
      Op: "$contains",
      Value: [data.query.EmployeeName],
    });
  }
  if (data.query.Status !== null && data.query.Status !== "") {
    filters.push({
      Field: "Status",
      Op: "$contains",
      Value: [data.query.Status],
    });
  }
  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}

function newRecord(record) {
  record._id = "";
  record.CompanyID = auth.appData.CompanyID;
  record.EmployeeID = "";
  record.SubmissionType = "";
  record.Reason = "";
  record.Assesment = false;
  record.AssesmentResult = "";
  record.Status = "DRAFT";
  data.record = record;

  openForm(record);
}

function editRecord(record) {
  if (!record.CompanyID) {
    record.CompanyID = auth.appData.CompanyID;
  }

  data.record = record;
  getEmployee(data.record.EmployeeID, data.record);
  setTransactionType(record.SubmissionType, data.record);
  openForm(record);
}

function openForm(record) {
  data.benefitDetail = {
    Existing: {},
    NewPropose: {},
  };
  util.nextTickN(2, () => {
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
  });
  // util.nextTickN(2, () => {
  //   if (record.Assesment) {
  //     listControl.value.setFormFieldAttr("AssesmentResult", "hide", false);
  //   } else {
  //     listControl.value.setFormFieldAttr("AssesmentResult", "hide", true);
  //   }
  // });
}

function onFormPostSave(record) {
  util.nextTickN(2, () => {
    saveBenefitDetail(data.benefitDetail.NewPropose);
  });
}

function getEmployee(id, record) {
  const url = "/bagong/employee/get";
  axios.post(url, [id]).then(
    (r) => {
      record.EmployeeName = r.data.Name;
      record.PointOfHire = r.data.Detail.POH;
      record.JoinedDate = r.data.JoinDate;
      record.Dimension = r.data.Dimension;
    },
    (e) => util.showError(e)
  );
}

function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
    case "EmployeeID":
      getEmployee(v1, record);
      break;
    case "Assesment":
      if (v1) {
        listControl.value.setFormFieldAttr("AssesmentResult", "hide", false);
      } else {
        listControl.value.setFormFieldAttr("AssesmentResult", "hide", true);
      }
      break;
    case "SubmissionType":
      setTransactionType(v1, record);
      break;
  }
}

function getBenefitDetail(ID, EmployeeID) {
  data.benefitDetail = {
    Existing: {},
    NewPropose: {},
  };
  if (ID && EmployeeID) {
    const url = "/hcm/talentdevelopment/get-detail";
    axios.post(url, { ID, EmployeeID }).then(
      (r) => {
        data.benefitDetail = r.data;
      },
      (e) => util.showError(e)
    );
  }
}

function saveBenefitDetail(payload) {
  const url = "/hcm/talentdevelopmentdetail/save";
  axios.post(url, payload).then(
    (r) => {},
    (e) => util.showError(e)
  );
}
const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
function setTransactionType(type, record) {
  switch (type) {
    case "Promotion":
      data.trxType =
        "Talent%20Development%20-%20Promotion%20-%20General%20%26%20Benefit";
      break;
    case "Rotation":
      data.trxType = "Talent Development - Rotation";
      break;
    case "Demotion":
      data.trxType = "Talent Development - Demotion";
      break;
    case "Salary Change":
      data.trxType = "Talent Development - Salary Change";
      break;
    case "POH":
      data.trxType = "Talent Development - POH Change";
      break;
    default:
      break;
  }
  util.nextTickN(2, () => {
    listControl.value.setFormLoading(true);
    axios
      .post(`/hcm/journaltype/find?TransactionType=${data.trxType}`, {
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
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}

function closePreview() {
  data.appMode = "grid";
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
        data.checklistTemplateID = r.data.ChecklistTemplate;
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
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "SUBMITTED", "READY"].includes(data.record.Status);
  },
});

function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    if (["Assesment"].includes(e.field) === true) return;
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
    const validDimension = dimenstionCtl.value.validate();
    if (valid && validDimension) {
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
watch(
  () => listControl.value?.getFormCurrentTab(),
  (nv) => {
    if (nv && nv == 1) {
      getBenefitDetail(data.record._id, data.record.EmployeeID);
    }
  }
);
</script>
