<template>
  <data-list
    v-if="data.appMode != 'preview'"
    class="card datalist-map-journal-type"
    ref="listControl"
    title="OL & Plotting"
    no-gap
    grid-editor
    grid-hide-search
    grid-hide-sort
    grid-hide-footer
    grid-hide-delete
    gridHideNew
    grid-no-confirm-delete
    init-app-mode="grid"
    grid-mode="grid"
    form-keep-label
    grid-auto-commit-line
    grid-config="/hcm/olplotting/gridconfig"
    form-config="/hcm/olplotting/formconfig"
    form-update="/hcm/olplotting/save"
    :grid-fields="['CandidateID', 'JobVacancyID', 'Plotting', 'Status', '_id']"
    :form-default-mode="data.formMode"
    :formHideSubmit="readOnly"
    @formEditData="editRecord"
    @gridRefreshed="gridRefreshed"
    @gridRowFieldChanged="onGridRowFieldChanged"
    @form-field-change="onFormFieldChange"
    @alter-form-config="onAlterFormConfig"
    stay-on-form-after-save
  >
    <template #grid_header_search>
      <div class="grow flex gap-3 justify-start grid-header-filter">
        <s-input
          class="w-[200px]"
          label="Job Vacancy"
          use-list
          lookup-url="/hcm/manpowerrequest/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="data.filter.jobVacancyID"
        />
      </div>
    </template>
    <template #grid_header_buttons_2>
      <s-button
        label="save"
        :disabled="data.loadingSave"
        icon="content-save"
        class="ml-2 btn_primary submit_btn"
        @click="handleSave"
      />
    </template>
    <!-- grid status -->
    <template #grid__id="{ item }">
      {{ item._id }}
    </template>
    <template #grid_Status="{ item }">
      <status-text :txt="item.Status" />
    </template>
    <template #grid_item_buttons_1="{ item }">
      <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
    </template>
    <template #grid_paging>&nbsp;</template>
    <template #grid_CandidateID="{ item, mode }">
      <s-input
        use-list
        lookup-url="/tenant/employee/find"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="item.CandidateID"
        read-only
      />
    </template>
    <template #grid_JobVacancyID="{ item, mode }">
      <s-input
        use-list
        lookup-url="/hcm/manpowerrequest/find"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="item.JobVacancyID"
        :read-only="!['', 'DRAFT'].includes(item.Status)"
      />
    </template>
    <template #grid_Plotting="{ item, mode }">
      <s-input
        use-list
        lookup-url="/tenant/dimension/find?DimensionType=Site"
        lookup-key="_id"
        :lookup-labels="['Label']"
        :lookupSearchs="['Label']"
        v-model="item.Plotting"
        :read-only="!['', 'DRAFT'].includes(item.Status)"
      />
    </template>
    <template #form_buttons_1="{ item, inSubmission, loading, mode }">
      <div class="flex flex-row gap-2">
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
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :auto-post="!waitTrxSubmit"
        />
      </div>
    </template>
  </PreviewReport>
</template>
<script setup>
import { reactive, ref, inject, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SButton, SInput } from "suimjs";
import { authStore } from "@/stores/auth";
import { useRoute } from "vue-router";
import LogTrx from "@/components/common/LogTrx.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import helper from "@/scripts/helper.js";
import PreviewReport from "@/components/common/PreviewReport.vue";

const axios = inject("axios");

layoutStore().name = "tenant";

const FEATUREID = "";
const profile = authStore().getRBAC(FEATUREID);
const auth = authStore();

const route = useRoute();
const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  records: [],
  record: {},
  filter: {
    jobVacancyID: route.query.JobVacancyID,
  },
  loadingSave: false,
  jType: "PLOTTING",
  journalType: {},
});

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
watch(
  () => JSON.stringify(data.filter),
  (nv) => {
    util.nextTickN(2, () => {
      gridRefreshed();
    });
  }
);

function onAlterFormConfig(cfg) {
  cfg.sectionGroups = cfg.sectionGroups.map((sectionGroup) => {
    sectionGroup.sections = sectionGroup.sections.map((section) => {
      section.rows.map((row) => {
        row.inputs = row.inputs.filter(
          (input) => !["OfferingLetter"].includes(input.field)
        );
        return row;
      });
      return section;
    });
    return sectionGroup;
  });
}
function handleSave() {
  const param = data.records
    .filter((e) => e.IsEdited == true)
    .map((e) => {
      return {
        ID: e._id,
        Site: e.Plotting,
      };
    });
  data.loadingSave = true;
  listControl.value.setGridLoading(true);
  axios
    .post("/hcm/tracking/plotting", param)
    .then((r) => {})
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loadingSave = false;
      gridRefreshed();
    });
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
  record.IsEdited = true;
}

function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
  }
}
function gridRefreshed() {
  listControl.value.setGridLoading(true);
  const items = [{ Op: "$eq", Field: "StageStatus", Value: "Not Selected" }];

  if (data.filter.jobVacancyID != "") {
    items.push({
      Op: "$eq",
      Field: "JobVacancyID",
      Value: data.filter.jobVacancyID,
    });
  }
  axios
    .post("/hcm/olplotting/gets", {
      Where: {
        Op: "$and",
        Items: items,
      },
    })
    .then(
      (r) => {
        listControl.value?.setGridLoading(false);
        data.records = r.data.data.map((e) => {
          e.IsEdited = false;
          return e;
        });
        listControl.value.setGridRecords(data.records);
      },
      (e) => {
        listControl.value?.setGridLoading(false);
        util.showError(e);
      }
    );
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
    if (!["EmployeeStatus", "THR", "BPJSTK", "BPJSHealth"].includes(e.field)) {
      listControl.value.setFormFieldAttr(e.field, "required", required);
    }
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
table td:last-child {
  text-align: center;
}
.datalist-map-journal-type .suim_area_table {
  position: relative;
  max-height: calc(100vh - 250px);
  overflow: auto;
}
.datalist-map-journal-type .suim_table > thead {
  position: sticky;
  top: 0;
}
</style>
