<template>
  <div class="w-full">
    <data-list
      v-if="data.appMode != 'preview'"
      class="card"
      ref="listControl"
      title="Business Trip"
      grid-config="/hcm/businesstrip/gridconfig"
      form-config="/hcm/businesstrip/form/formconfig"
      grid-read="/hcm/businesstrip/gets"
      form-read="/hcm/businesstrip/get"
      grid-mode="grid"
      grid-delete="/hcm/businesstrip/delete"
      form-keep-label
      form-insert="/hcm/businesstrip/save"
      form-update="/hcm/businesstrip/save"
      :grid-fields="['Status']"
      :form-tabs-edit="['General', 'Lines']"
      :form-tabs-view="['General', 'Lines']"
      :form-fields="['Dimension', 'RequestorPosition', 'RequestorSite']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      @pre-Save="preSave"
      stay-on-form-after-save
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="customFilter"
      @alterGridConfig="alterGridConfig"
      @form-field-change="onFormFieldChange"
    >
      <!-- grid filter -->
      <template #grid_header_search="{ config }">
        <s-input
          v-model="data.query.RequestDate"
          kind="date"
          label="Request Date"
          class="w-full"
          @change="refreshData"
        ></s-input>
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
      <template #form_buttons_1="{ item, config, loading, mode}">
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
      <!-- form input -->
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :read-only="readOnly || mode == 'view'"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #form_input_RequestorPosition="{ item, mode }">
        <s-input
          :key="`${item.RequestorID}_${item.RequestorPosition}`"
          label="Requestor Position"
          v-model="item.RequestorPosition"
          use-list
          lookup-key="_id"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
          :lookup-labels="['Name']"
          read-only
        />
      </template>
      <template #form_input_RequestorSite="{ item, mode }">
        <s-input
          :key="`${item.RequestorID}_${item.RequestorSite}`"
          label="Requestor Site"
          v-model="item.RequestorSite"
          use-list
          lookup-key="_id"
          lookup-url="/tenant/dimension/find?DimensionType=Site"
          :lookup-labels="['Label']"
          read-only
        />
      </template>

      <!-- form tab -->
      <template #form_tab_Lines="{ item, mode }">
        <lines
          v-model="item.Lines"
          :key="item.Lines"
          :read-only="readOnly || mode == 'view'"
          force-show-grid-detail
          grid-config-url="/hcm/businesstrip/line/gridconfig"
          form-config-url="/hcm/businesstrip/line/formconfig"
          :form-hide-cancel="false"
          @new-record="newRecordLine"
          @alter-grid-config="alterGridLinesConfig"
          @grid-row-field-changed="onGridLinesRowFieldChanged"
        >
          <template #form_EmployeeID="p">
            <s-input
              label="Employee Name"
              v-model="p.item.EmployeeID"
              use-list
              lookup-key="_id"
              lookup-url="/tenant/employee/find"
              :lookup-labels="['_id', 'Name']"
              read-only
            />
          </template>
          <template #grid_EmployeeNIK="p">
            <s-input
              hide-label
              v-if="p.item.EmployeeNIK"
              v-model="p.item.EmployeeNIK"
              read-only
            />
            <div v-else>&nbsp;</div>
          </template>
          <template #grid_EmployeePosition="p">
            <s-input
              hide-label
              v-if="p.item.EmployeePosition"
              v-model="p.item.EmployeePosition"
              use-list
              lookup-key="_id"
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
              :lookup-labels="['Name']"
              read-only
            />
            <div v-else>&nbsp;</div>
          </template>
          <template #grid_EmployeeLevel="p">
            <s-input
              hide-label
              v-if="p.item.EmployeeLevel"
              v-model="p.item.EmployeeLevel"
              use-list
              lookup-key="_id"
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=LME"
              :lookup-labels="['Name']"
              read-only
            />
            <div v-else>&nbsp;</div>
          </template>
          <template #grid_EmployeeDepartment="p">
            <s-input
              hide-label
              :key="`${item.EmployeeID}_${item.EmployeeDepartment}`"
              v-if="p.item.EmployeeDepartment"
              v-model="p.item.EmployeeDepartment"
              use-list
              lookup-key="_id"
              lookup-url="/tenant/dimension/find?DimensionType=CC"
              :lookup-labels="['Label']"
              read-only
            />
            <div v-else>&nbsp;</div>
          </template>
          <template #grid_EmployeeSite="{ item, mode }">
            <s-input
              hide-label
              :key="`${item.EmployeeID}_${item.EmployeeSite}`"
              v-model="item.EmployeeSite"
              use-list
              lookup-key="_id"
              lookup-url="/tenant/dimension/find?DimensionType=Site"
              :lookup-labels="['Label']"
              read-only
            />
          </template>
          <template #form_Details="{ item, mode }">
            <lines
              v-model="item.Details"
              :key="item.Details"
              :read-only="readOnly || mode == 'view'"
              grid-config-url="/hcm/businesstrip/line/detail/gridconfig"
              grid-hide-detail
              @new-record="newRecordLineDetail"
              @grid-row-field-changed="
                (name, v1, v2, old, record, onUpdatedRow) => {
                  onGridLineDetailRowFieldChanged(name, v1, record, item);
                }
              "
            >
            </lines>
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
import { authStore } from "@/stores/auth";
import { reactive, ref, inject, computed, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";

import Lines from "./widget/Lines.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import StatusText from "@/components/common/StatusText.vue";
import helper from "@/scripts/helper.js";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import LogTrx from "@/components/common/LogTrx.vue";

layoutStore().name = "tenant";

const FEATUREID = "B";
const profile = authStore().getRBAC(FEATUREID);
const auth = authStore();

const axios = inject("axios");

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  query: {
    RequestDate: null,
    RequestorName: "",
    Status: "",
  },
  records: [],
  record: {},
  jType: "BUSINESSTRIP",
  journalType: {},
});

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
}

function alterGridLinesConfig(cfg) {
  cfg.fields.splice(
    1,
    0,
    helper.gridColumnConfig({
      field: "EmployeeNIK",
      label: "NIK",
      kind: "text",
      readOnly: true,
      readOnlyOnEdit: true,
      readOnlyOnNew: true
    })
  );
  cfg.fields.splice(
    2,
    0,
    helper.gridColumnConfig({
      field: "EmployeePosition",
      label: "Position",
      kind: "text",
    })
  );
  cfg.fields.splice(
    3,
    0,
    helper.gridColumnConfig({
      field: "EmployeeLevel",
      label: "Level",
      kind: "text",
    })
  );
  cfg.fields.splice(
    4,
    0,
    helper.gridColumnConfig({
      field: "EmployeeDepartment",
      label: "Department",
      kind: "text",
    })
  );
  cfg.fields.splice(
    5,
    0,
    helper.gridColumnConfig({
      field: "EmployeeSite",
      label: "Site",
      kind: "text",
    })
  );
}

let customFilter = computed(() => {
  const filters = [];
  if (
    data.query.RequestDate !== null &&
    data.query.RequestDate !== "" &&
    data.query.RequestDate !== "Invalid date"
  ) {
    filters.push({
      Field: "RequestDate",
      Op: "$gte",
      Value: helper.formatFilterDate(data.query.RequestDate),
    });
  }
  if (data.query.RequestorName !== null && data.query.RequestorName !== "") {
    filters.push({
      Field: "RequestorID",
      Op: "$contains",
      Value: [data.query.RequestorName],
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
  if (record.RequestorID) {
    getEmployeeDetail(data.record.RequestorID, data.record);
  }
  if (!record.CompanyID) {
    record.CompanyID = auth.appData.CompanyID;
  }
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
    Location: item.Location,
    Task: item.Task,
    Details: item.Details,
    TotalCost: item.TotalCost,
  }));
  record.Lines = mappingLines;
}

function getEmployeeForLines(id, record, onUpdatedRow) {
  if (id) {
    const url = "/bagong/employee/get";
    axios.post(url, [id]).then(
      (r) => {
        record.EmployeeLevel = r.data.Detail.Level;
        record.EmployeeNIK = r.data.Detail.EmployeeNo;
        record.EmployeePosition = r.data.Detail.Position;
        record.EmployeeDepartment = r.data.Dimension.find(
          (o) => o.Key == "CC"
        ).Value;
        record.EmployeeSite = r.data.Dimension.find(
          (o) => o.Key == "Site"
        ).Value;
        onUpdatedRow();
      },
      (e) => util.showError(e)
    );
  } else {
    record.EmployeeLevel = "";
    record.EmployeeNIK = "";
    record.EmployeePosition = "";
    record.EmployeeDepartment = "";
    record.EmployeeSite = "";
    onUpdatedRow();
  }
  // console.log(record);
}
function getEmployeeDetail(id, record) {
  if (id) {
    const url = "/bagong/employee/get";
    axios.post(url, [id]).then(
      (r) => {
        record.RequestorEmail = r.data.Email;
        record.RequestorPosition = r.data.Detail.Position;
        record.RequestorSite = r.data.Dimension.find(
          (o) => o.Key == "Site"
        ).Value;
        record.Dimension = r.data.Dimension
      },
      (e) => util.showError(e)
    );
    const url2 = "/bagong/employeedetail/find?EmployeeID=" + id;
    axios.post(url2).then(
      (r) => {
        util.nextTickN(2, () => {
          if (r.data.length > 0) {
            record.RequestorNIK = r.data[0].EmployeeNo;
          }
        });
      },
      (e) => {
        util.showError(e);
      }
    );
  } else {
    record.RequestorEmail = "";
    record.RequestorPosition = "";
    record.RequestorSite = "";
  }
}

function newRecordLine(record) {
  record.EmployeeID = "";
  record.Location = "";
  record.Task = "";
}

function onGridLinesRowFieldChanged(name, v1, v2, old, record, onUpdatedRow) {
  if (name == "EmployeeID") {
    getEmployeeForLines(v1, record, onUpdatedRow);
  }
}
function newRecordLineDetail(record) {
  record.Description = "";
  record.Cost = 0;
}

function onGridLineDetailRowFieldChanged(name, v1, recordDetail, recordLine) {
  // if (name == "Cost") {
  //   recordLine.TotalCost = recordLine.Details.reduce((sum, item) => sum + item.Cost, 0);
  // }
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
      getEmployeeDetail(v1, record);
      break;
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
watch(
  () => data.record.Lines,
  (nv) => {
    nv.forEach((line) => {
      line.TotalCost = line.Details?.reduce(
        (sum, detail) => sum + detail.Cost,
        0
      );
    });
  },
  { deep: true }
);
</script>
