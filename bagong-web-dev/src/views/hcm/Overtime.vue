<template>
  <div class="w-full">
    <data-list
      v-if="data.appMode != 'preview'"
      class="card"
      ref="listControl"
      title="Overtime"
      grid-config="/hcm/overtime/gridconfig"
      form-config="/hcm/overtime/formconfig"
      grid-read="/hcm/overtime/gets"
      form-read="/hcm/overtime/get"
      grid-mode="grid"
      grid-delete="/hcm/overtime/delete"
      form-keep-label
      form-insert="/hcm/overtime/save"
      form-update="/hcm/overtime/save"
      :grid-fields="['RequestorDepartment', 'Status']"
      :form-tabs-edit="['General', 'Lines']"
      :form-tabs-view="['General', 'Lines']"
      :form-fields="[
        'RequestorID',
        'Dimension',
        'EstimatedStartTime',
        'EstimatedEndTime',
      ]"
      :init-app-mode="data.appMode"
      :formHideSubmit="readOnly"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      @formFieldChange="onFormFieldChange"
      @pre-Save="preSave"
      stay-on-form-after-save
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="customFilter"
      @alterGridConfig="alterGridConfig"
    >
      <!-- grid filter -->
      <template #grid_header_search="{ config }">
        <!-- <s-input
          v-model="data.query.OvertimeDate"
          kind="date"
          label="Overtime Date"
          class="w-full"
          @change="refreshData"
        ></s-input> -->
        <div class="flex gap-1">
          <s-input
            class="filter-date-from"
            label="Date From "
            kind="date"
            v-model="data.query.OvertimeDateFrom"
            @change="refreshData"
          />
          <s-input
            class="filter-date-to"
            label="Date To"
            kind="date"
            v-model="data.query.OvertimeDateTo"
            @change="refreshData"
          />
        </div>
        <s-input
          v-model="data.query.RequestorName"
          kind="text"
          label="Requestor Name"
          class="w-full"
          use-list
          lookup-url="/tenant/employee/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          v-model="data.query.RequestorDepartment"
          kind="text"
          label="Requestor Department"
          class="w-full"
          use-list
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=DME"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
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
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
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

      <!-- grid slot -->
      <template #grid_RequestorDepartment="{ item }">
        <s-input
          hide-label
          v-model="item.RequestorDepartment"
          use-list
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=DME"
          lookup-key="_id"
          :lookup-labels="['Name']"
          read-only
        />
      </template>

      <!-- form input -->
      <template #form_input_RequestorID="{ item }">
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
      <template #form_input_EstimatedStartTime="{ item, config }">
        <div class="mb-[27px]">
          <s-input
            label="Requestor Department"
            keep-label
            v-model="data.requestorDepartment"
            read-only
          />
        </div>
        <div>
          <s-input
            :label="config.label"
            v-model="item.EstimatedStartTime"
            :kind="config.kind"
            keep-label
          />
        </div>
      </template>
      <template #form_input_EstimatedEndTime="{ item, config }">
        <div class="mb-10">&nbsp;</div>
        <div>
          <s-input
            :label="config.label"
            v-model="item.EstimatedEndTime"
            :kind="config.kind"
            keep-label
          />
        </div>
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :read-only="readOnly || mode == 'view'"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>

      <!-- form tab -->
      <template #form_tab_Lines="{ item, mode }">
        <lines
          v-model="item.Lines"
          :key="item.Lines"
          :read-only="readOnly || mode == 'view'"
          grid-hide-detail
          grid-config-url="/hcm/overtime/line/gridconfig"
          @new-record="newRecordLine"
          @alter-grid-config="alterGridLinesConfig"
          @grid-row-field-changed="onGridLinesRowFieldChanged"
        >
          <template #grid_EmployeeName="p">
            <s-input
              hide-label
              v-if="p.item.EmployeeName"
              v-model="p.item.EmployeeName"
              read-only
            />
            <div v-else>&nbsp;</div>
          </template>
          <template #grid_Position="p">
            <s-input
              hide-label
              v-if="p.item.Position"
              v-model="p.item.Position"
              use-list
              lookup-key="_id"
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
              :lookup-labels="['Name']"
              read-only
            />
            <div v-else>&nbsp;</div>
          </template>
          <template #grid_EmployeeDepartment="p">
            <s-input
              hide-label
              v-if="p.item.EmployeeDepartment"
              v-model="p.item.EmployeeDepartment"
              use-list
              lookup-key="_id"
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=DME"
              :lookup-labels="['Name']"
              read-only
            />
            <div v-else>&nbsp;</div>
          </template>
          <template #grid_ActualOvertime="p">
            <s-input
              v-if="readOnly"
              v-model="p.item.ActualOvertime"
              kind="number"
              :read-only="readOnly"
              :rules="[maxLength]"
            />
            <div v-else class="suim_input">
              <input
                v-model="p.item.ActualOvertime"
                type="number"
                class="input_field text-right"
                max="4"
                step="0.1"
                @input="
                  (e) => {
                    const value = parseFloat(e.target.value);
                    if (value > 4) {
                      p.item.ActualOvertime = 4;
                    }
                  }
                "
              />
            </div>
          </template>
        </lines>
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
import { authStore } from "@/stores/auth";
import { reactive, ref, inject, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";

import Lines from "./widget/Lines.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import StatusText from "@/components/common/StatusText.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";

const FEATUREID = "Overtime";
const profile = authStore().getRBAC(FEATUREID);
const auth = authStore();

const axios = inject("axios");

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  query: {
    OvertimeDateFrom: null,
    OvertimeDateTo: null,
    RequestorName: "",
    RequestorDepartment: "",
    Status: "",
  },
  records: [],
  record: {},
  requestorDepartment: "",
  jType: "OVERTIME",
  journalType: {},
});
function maxLength(v) {
  if (parseFloat(v) > 4) {
    return "Max 4 allowed!";
  }
  return "";
}
function alterGridConfig(cfg) {
  cfg.fields.splice(
    1,
    0,
    helper.gridColumnConfig({
      field: "RequestorName",
      label: "Requestor Name",
      kind: "text",
    })
  );
  cfg.fields.splice(
    2,
    0,
    helper.gridColumnConfig({
      field: "RequestorDepartment",
      label: "Requestor Department",
      kind: "text",
    })
  );
}

function alterGridLinesConfig(cfg) {
  cfg.fields.splice(
    1,
    0,
    helper.gridColumnConfig({
      field: "EmployeeName",
      label: "Name",
      kind: "text",
    })
  );
  // cfg.fields.splice(
  //   2,
  //   0,
  //   helper.gridColumnConfig({
  //     field: "EmployeePosition",
  //     label: "Position",
  //     kind: "text",
  //   })
  // );
  cfg.fields.splice(
    3,
    0,
    helper.gridColumnConfig({
      field: "EmployeeDepartment",
      label: "Department",
      kind: "text",
    })
  );
}

let customFilter = computed(() => {
  const filters = [];
  if (
    data.query.OvertimeDateFrom != null &&
    data.query.OvertimeDateFrom != ""
  ) {
    filters.push({
      Field: "OvertimeDate",
      Op: "$gte",
      Value: helper.formatFilterDate(data.query.OvertimeDateFrom),
    });
  }
  if (data.query.OvertimeDateTo != null && data.query.OvertimeDateTo != "") {
    filters.push({
      Field: "OvertimeDate",
      Op: "$lte",
      Value: helper.formatFilterDate(data.query.OvertimeDateTo, true),
    });
  }
  if (data.query.RequestorName !== null && data.query.RequestorName !== "") {
    filters.push({
      Field: "RequestorID",
      Op: "$contains",
      Value: [data.query.RequestorName],
    });
  }
  if (
    data.query.RequestorDepartment !== null &&
    data.query.RequestorDepartment !== ""
  ) {
    filters.push({
      Field: "RequestorDepartment",
      Op: "$contains",
      Value: [data.query.RequestorDepartment],
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
  record.RequestorID = "";
  record.Status = "DRAFT";
  record.CompanyID = auth.appData.CompanyID;
  record.Lines = [];
  openForm(record);
}

function openForm(record) {
  data.record = record;
  data.requestorDepartment = "";
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("PostingProfileID", "readOnly", true);
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
  });
}

function preSave(record) {
  const mappingLines = record.Lines?.map((item) => ({
    EmployeeID: item.EmployeeID,
    Position: item.Position,
    Task: item.Task,
    ActualOvertime: item.ActualOvertime,
    OffDay: item.OffDay,
  }));
  record.Lines = mappingLines;
  // if (record.Lines.find(o => o.ActualOvertime > 4)) {
  //   util.showError('Max Actual Overtime is 4! Please check your lines')
  //   return
  // }
}

function getMasterDataById(id) {
  const url = "/tenant/masterdata/get";
  axios.post(url, [id]).then(
    (r) => {
      data.requestorDepartment = r.data.Name;
    },
    (e) => util.showError(e)
  );
}

function getEmployee(id) {
  if (id) {
    const url = "/bagong/employee/get";
    axios.post(url, [id]).then(
      (r) => {
        if (r.data.Detail.Department.length > 0) {
          getMasterDataById(r.data.Detail.Department);
        } else {
          data.requestorDepartment = "";
        }
        data.record.Dimension = r.data.Dimension;
      },
      (e) => util.showError(e)
    );
  } else {
    data.requestorDepartment = "";
  }
}

function getEmployeeForLines(id, record, onUpdatedRow) {
  if (id) {
    const url = "/bagong/employee/get";
    axios.post(url, [id]).then(
      (r) => {
        record.EmployeeName = r.data.Name;
        record.EmployeePosition = r.data.Detail.Position;
        record.Position = r.data.Detail.Position;
        record.EmployeeDepartment = r.data.Detail.Department;
        onUpdatedRow();
      },
      (e) => util.showError(e)
    );
  } else {
    record.EmployeeName = "";
    // record.EmployeePosition = "";
    record.Position = "";
    record.EmployeeDepartment = "";
    onUpdatedRow();
  }
}

function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
    case "RequestorID":
      getEmployee(v1);
      break;
  }
}

function newRecordLine(record) {
  record.EmployeeID = "";
  record.Position = "";
  record.Task = "";
  record.ActualOvertime = 0;
  record.OffDay = false;
}

function onGridLinesRowFieldChanged(name, v1, v2, old, record, onUpdatedRow) {
  if (name == "EmployeeID") {
    getEmployeeForLines(v1, record, onUpdatedRow);
  }
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
watch(
  () => data.record,
  (nv) => {
    if (nv) {
      getEmployee(nv.RequestorID);
    }
  }
);
</script>
