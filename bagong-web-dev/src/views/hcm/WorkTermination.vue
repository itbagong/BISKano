<template>
  <div class="w-full">
    <data-list
      v-show="data.appMode != 'preview'"
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/hcm/worktermination/gridconfig"
      form-config="/hcm/worktermination/formconfig"
      grid-read="/hcm/worktermination/gets"
      form-read="/hcm/worktermination/get-detail"
      grid-mode="grid"
      grid-delete="/hcm/worktermination/delete"
      form-keep-label
      form-insert="/hcm/worktermination/insert"
      form-update="/hcm/worktermination/save"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :form-tabs-new="['General']"
      :form-tabs-edit="generateTabs()"
      :form-tabs-view="generateTabs()"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :formHideSubmit="readOnly"
      :grid-fields="['Status']"
      :form-fields="[
        'EmployeeID',
        'EmployeeDetail',
        'Dimension',
        'JournalTypeID',
        'JoinedDate',
        'Site',
      ]"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :form-hide-submit="data.formMode === 'new'"
      @preSave="preSubmit"
      stay-on-form-after-save
    >
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #form_input_EmployeeDetail="{ item }">
        <div
          class="flex flex-col gap-2"
          v-if="
            ['Meninggal', 'PHK', 'Pensiun', 'SakitBerkepanjangan'].includes(
              item.Type
            ) && item.EmployeeID
          "
        >
          <s-input
            label="Grade"
            v-model="item.Grade"
            use-list
            lookup-key="_id"
            lookup-url="/tenant/masterdata/find?MasterDataTypeID=GDE"
            :lookup-labels="['Name']"
            read-only
          />
          <s-input
            label="Position"
            v-model="item.Postion"
            use-list
            lookup-key="_id"
            lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
            :lookup-labels="['Label']"
            read-only
          />
          <s-input
            label="Department"
            v-model="item.Department"
            use-list
            lookup-key="_id"
            lookup-url="/tenant/dimension/find?DimensionType=CC"
            :lookup-labels="['Label']"
            read-only
          />
          <s-input label="Education" v-model="item.Education" read-only />
          <s-input
            label="Working period"
            v-model="calculateWorkingDay"
            read-only
          />
          <s-input label="UMK Site" v-model="item.UMKSite" read-only />
        </div>
        <div v-else>&nbsp;</div>
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
      <template #form_input_Site="{ item, config }">
        <div>
          <s-input
            :key="item.EmployeeID + item.Site"
            keep-label
            useList
            read-only
            label="Site"
            lookup-url="/tenant/dimension/find?DimensionType=Site"
            lookup-key="_id"
            :lookup-labels="['Label']"
            :lookup-searchs="['_id', 'Label']"
            v-model="item.Site"
            class="min-w-[180px]"
          />
        </div>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          ref="dimenstionCtl"
          v-model="item.Dimension"
          :read-only="['SUBMITTED'].includes(item.Status)"
          sectionTitle="Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
      <template #form_tab_Exit_Interview="{ item, mode }">
        <References
          :ReferenceTemplate="data.refrenceTemplateID"
          :readOnly="readOnly || mode == 'view'"
          v-model="item.InterviewAnswers"
        />
      </template>
      <template #form_tab_Severance="{ item, mode }">
        <Severance v-model="data.record"></Severance>
      </template>
      <template #form_tab_Administrative="{ item, mode }">
        <Checklist
          v-model="item.Administrative"
          :checklist-id="data.checklistTemplateID"
          kind="WORK_TERMINATION"
          :attch-kind="data.attchKind"
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.attchKind"
          @preOpenAttch="preOpenAttch"
          :readOnly="readOnly || mode == 'view'"
        />
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
          <s-button
            v-if="data.formMode === 'new'"
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Save"
            icon="content-save"
            @click="onInsert(item)"
          ></s-button>
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
      <!-- grid status -->
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
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
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import { DataList, util, SForm, SInput, SButton, loadFormConfig } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import References from "@/components/common/References.vue";
import Checklist from "@/components/common/Checklist.vue";
import Severance from "./widget/Severance.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import helper from "@/scripts/helper.js";
import moment from "moment";

import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";

layoutStore().name = "tenant";
const featureID = "WorkTermination";

const profile = authStore().getRBAC(featureID);
const auth = authStore();

const headOffice = layoutStore().headOfficeID;
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const listControl = ref(null);
const dimenstionCtl = ref(null);
const axios = inject("axios");

const data = reactive({
  isPreview: false,
  appMode: "grid",
  formMode: "edit",
  titleForm: "Work Termination",
  record: {
    _id: "",
    RequestDate: new Date(),
    ResignDate: new Date(),
    Dimension: [],
    Status: "",
    InterviewAnswers: [],
  },
  tabs: ["General", "Exit Interview", "Administrative"],
  attchKind: "WORK_TERMINATION",
  jType: "WORKTERMINATION",
  refrenceTemplateID: "",
  checklistTemplateID: "",
  journalType: {},
  trxType: "",
});
function generateTabs() {
  switch (data.record.Type) {
    case "Resign":
      return ["General", "Exit Interview", "Administrative"];
    case null:
      return ["General", "Administrative"];
    default:
      return ["General", "Severance", "Administrative"];
  }
}
function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Work Termination`;
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
  data.formMode = "edit";
  data.titleForm = `Edit Work Termination | ${record._id}`;
  if (!record.CompanyID) {
    record.CompanyID = auth.appData.CompanyID;
  }

  data.record = record;
  setTransactionType(record.Type, data.record);
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
    util.nextTickN(2, () => {
      if (readOnly.value === true) {
        listControl.value.setFormMode("view");
      }
    });
  });
}
function setTransactionType(type, record) {
  switch (type) {
    case "Pensiun":
      data.trxType = "Work Termination - Pensiun";
      break;
    case "PHK":
      data.trxType = "Work Termination - PHK";
      break;
    case "Meninggal":
      data.trxType = "Work Termination - Meninggal";
      break;
    case "Resign":
      data.trxType = "Work Termination - Resign";
      break;
    case "SakitBerkepanjangan":
      data.trxType = "Work Termination - Sakit Berkepanjangan";
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
        Select: [
          "Name",
          "_id",
          "PostingProfileID",
          "ChecklistTemplate",
          "ReferenceTemplate",
        ],
      })
      .then(
        (r) => {
          if (r.data.length == 1) {
            data.journalType = r.data[0];
            record.JournalTypeID = r.data[0]._id;
            record.PostingProfileID = r.data[0].PostingProfileID;
            getJurnalType(record.JournalTypeID, record);
            // data.checklistTemplateID = r.data[0].ChecklistTemplate;
            // data.refrenceTemplateID = r.data[0].ReferenceTemplate;
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

function preSubmit(record) {}
function onInsert(record) {
  const validDimension = dimenstionCtl.value.validate();
  const valid = listControl.value.formValidate();

  if (valid && validDimension) {
    setLoadingForm(true);
    listControl.value.submitForm(
      data.record,
      () => {},
      () => {
        setLoadingForm(false);
      }
    );
  }
}
const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
function setFormMode(mode) {
  listControl.value.setFormMode(mode);
}
const calculateWorkingDay = computed({
  get() {
    const joinDate = moment(data.record.JoinedDate);
    const resignDate = moment(data.record.ResignDate);
    const duration = moment.duration(resignDate.diff(joinDate));

    const years = duration.years();
    const months = duration.months();
    const days = duration.days();

    let result;
    if (years === 0 && months === 0 && days > 0) {
      result = `${days} days`;
    } else if (years === 0 && months > 0) {
      result = `${months} months ${days > 0 ? `${days} days` : ""}`;
    } else if (months === 0) {
      result = `${years} years`;
    } else {
      result = `${years} years ${months} months`;
    }

    return result;
  },
});
function getDetailEmployee(id, record) {
  if (id == null) {
    return;
  }
  axios.post("/tenant/employee/get", [id]).then(
    (r) => {
      util.nextTickN(2, () => {
        record.EmployeeName = r.data.Name;
        record.JoinedDate = r.data.JoinDate;
        record.Grade = r.data.Grade;
        record.Department = r.data.Dimension.find((_dim) => _dim.Key === "CC")[
          "Value"
        ];
        record.Education = r.data.Education;
        record.Dimension = r.data.Dimension;
        const site = r.data.Dimension.find((_dim) => _dim.Key === "Site")[
          "Value"
        ];
        record.Site = site
        axios.post("/bagong/sitesetup/get", [site]).then(
          (rr) => {
            record.UMKSite = helper.formatNumberWithDot(
              rr.data.Configuration.UMK
            );
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
  const url = "/bagong/employeedetail/find?EmployeeID=" + id;
  axios.post(url).then(
    (r) => {
      util.nextTickN(2, () => {
        if (r.data.length > 0) {
          record.PointOfHire = r.data[0].POH;
          record.NIK = r.data[0].EmployeeNo;
        }
      });
    },
    (e) => {
      util.showError(e);
    }
  );
}
function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
    case "EmployeeID":
      getDetailEmployee(v1, record);
      break;
    case "Type":
      setTransactionType(v1, record);
      break;
  }
}
function onControlModeChanged(mode) {
  data.appMode = mode;
  if (mode === "grid") {
    data.titleForm = "Work Termination";
  }
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function onSubmit(status) {
  setLoadingForm(true);
  listControl.value.submitForm(
    { ...data.record, Status: status },
    () => {
      setLoadingForm(false);
    },
    () => {
      setLoadingForm(false);
    },
    true
  );
}
function preOpenAttch(readOnly) {
  // if (readOnly) return;
  // listControl.value.submitForm(data.record)
  return;
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
        data.refrenceTemplateID = r.data.ReferenceTemplate;
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
