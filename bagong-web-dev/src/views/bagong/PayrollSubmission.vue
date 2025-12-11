<template>
  <div class="w-full payroll-submission">
    <data-list
      class="card"
      ref="listControl"
      title="Payroll Submission"
      grid-config="/bagong/payroll_submission_custom/gridconfig"
      form-config="/bagong/payroll_submission_custom/formconfig"
      form-insert="/bagong/payroll/post-submission"
      form-update="/bagong/payroll/post-submission"
      grid-delete="/fico/ledgerjournal/delete"
      grid-mode="grid"
      form-keep-label
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @form-edit-data="onOpenForm"
      :form-fields="['Dimension']"
      :form-tabs-edit="['General', 'Line', 'References', 'Documents']"
      :form-tabs-new="['General', 'Line', 'Documents']"
      :grid-fields="['SiteID', 'Status', 'TotalAmount']"
      grid-hide-sort
      grid-hide-new
      @grid-refreshed="refreshGrid"
      @preSave="onPreSave"
      stay-on-form-after-save
      :form-hide-submit="hideFormSave"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @post-save="onPostSave"
    >
      <template #grid_header_search="{ config }">
        <div class="flex grow gap-2">
          <s-input
            kind="text"
            label="Searchhh"
            v-model="data.search.text"
            hideLabel
            class="grow mt-3"
            @keyup.enter="changeFilter"
          />
          <div class="min-w-[200px]">
            <dimension-editor
              multiple
              :default-list="profile.Dimension"
              v-model="data.search.Dimension"
              :required-fields="[]"
              :dim-names="['Site']"
              @change="changeFilter"
            ></dimension-editor>
          </div>
          <s-input
            kind="month"
            label="Period"
            class="mb-4 w-[150px]"
            v-model="data.search.period"
            @change="changeFilter"
          />
        </div>
      </template>
      <template #grid_header_buttons_2="{ config }">
        <s-button
          v-if="profile.canCreate"
          icon="plus"
          class="btn_primary new_btn"
          @click="openModal"
        />
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #form_tab_Line="{ item }">
        <s-grid
          hide-select
          hide-delete
          hide-title
          hide-action
          hide-control
          :config="data.lineCfg"
          v-model="data.record.Lines"
        >
        </s-grid>
      </template>
      <template #form_buttons_1="{ item }">
        <form-buttons-trx
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          post-url="/bagong/payroll/post"
          journal-type-id="LEDGERACCOUNT"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
        />
      </template>
      <template #form_tab_References="{ item }">
        <references
          :ReferenceTemplate="item.ReferenceTemplateID"
          :readOnly="false"
          v-model="item.References"
        />
      </template>
      <template #form_tab_Documents="{ item }">
        <attachment
          ref="gridAttachment"
          v-model="item.Attachment"
          :journalId="item._id"
          journalType="PAYROLL_SUBMISSION"
        />
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #grid_TotalAmount="{ item }">
        {{ calcAmount(item) }}
      </template>
      <template #grid_SiteID="{ item }">
        {{ item.SiteName }}
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #grid_item_button_delete="{ item }">
        <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
      </template>
    </data-list>
    <s-modal
      title="Payroll Site"
      class="p-4"
      :display="false"
      ref="generateModal"
      hideButtons
    >
      <s-input
        kind="month"
        label="Period"
        class="mb-4 w-[240px]"
        v-model="data.addNew.Period"
      ></s-input>
      <div class="mb-3">
        <dimension-editor
          v-model="data.addNew.Site"
          :default-list="profile.Dimension"
          :dim-names="['Site']"
        ></dimension-editor>
      </div>
      <div class="content-center">
        <s-button
          class="w-full btn_primary text-white justify-center"
          label="Submit"
          @click="getJournalType(data.addNew.Site)"
        ></s-button>
      </div>
    </s-modal>
  </div>
</template>
<script setup>
import { authStore } from "@/stores/auth";
import {
  reactive,
  ref,
  onMounted,
  inject,
  nextTick,
  watch,
  computed,
  readonly,
} from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  DataList,
  SGrid,
  SInput,
  SButton,
  SModal,
  util,
  loadGridConfig,
} from "suimjs";
import { useRoute } from "vue-router";
import helper from "@/scripts/helper.js";
import moment from "moment";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import References from "@/components/common/References.vue";
import StatusText from "@/components/common/StatusText.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import LogTrx from "@/components/common/LogTrx.vue";

layoutStore().name = "tenant";
const featureID = "PayrollSubmission";
const profile = authStore().getRBAC(featureID);
const dimensionSite = profile.Dimension?.filter((e) => e.Key === "Site").map(
  (e) => e.Value
);

const route = useRoute();
const axios = inject("axios");
const auth = authStore();

const gridSubmission = ref(null);
const generateModal = ref(null);
const listControl = ref(null);
const gridLine = ref(null);
const gridAttachment = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  config: {
    fields: [],
    setting: {},
  },
  records: [],
  record: {},
  isSelected: false,
  search: {
    text: "",
    siteID: [],
    period: null,
    Dimension: [],
  },
  addNew: {
    Period: "",
    SiteID: "",
  },
  lineCfg: {},
  journalPost: {},
});

onMounted(() => {
  refreshGrid();
  loadGridConfig(
    axios,
    "/bagong/payroll_submission_custom/submission_lines/gridconfig"
  ).then(
    (r) => {
      data.lineCfg = r;
    },
    (e) => util.showError(e)
  );
});

function generateSubmission(siteID) {
  listControl.value.setControlMode("form");
  listControl.value.setFormMode("new");
  data.record = {};
  generateModal.value.hide();
  data.record.Status = "DRAFT";
  data.record.PostingProfileID = data.journalPost.PostingProfileID;
  data.record.JournalType = data.journalPost._id;
  data.record.ReferenceTemplateID = data.journalPost.ReferenceTemplateID;
  data.record.TrxDate = new Date();
  data.record.Dimension = data.journalPost.Dimension;
  const findIdx = data.record.Dimension.findIndex((data) => data.Key == "Site");
  data.record.Dimension[findIdx].Value = siteID;
  getDimension(siteID)
  listControl.value.setFormRecord(data.record);
}
function getDimension(siteID) {
  axios.post('/bagong/sitesetup/get', [siteID]).then(
    (r) => {
      data.record.Dimension = r.data.Dimension
    },
  )
}
function openModal() {
  generateModal.value.show();
  data.addNew = {};
}
function changeFilter() {
  util.nextTickN(2, () => {
    refreshGrid();
  });
}

async function refreshGrid() {
  const url = "/bagong/payroll/get-journal-payroll";
  let period = data.search.period ?? new Date();
  const site = data.search.Dimension.find((e) => e.Key === "Site");

  let param = {
    CompanyID: "DEMO00",
    SiteID: site?.Value ?? [],
    Period: moment(period).format("YYYY-MM"),
    Skip: 0,
    Take: 25,
    Text: data.search.text,
  };
  listControl?.value?.setGridLoading(true);
  await axios.post(url, param).then(
    (r) => {
      listControl.value.setGridRecords(r.data.data);
      listControl.value?.setGridLoading(false);
    },
    (e) => {
      listControl.value.setGridLoading(false);
    }
  );
}

function generateLines(JournalTypeID, SiteID) {
  let dateStart = moment(data.addNew.Period).startOf("month");
  let dateEnd = moment(data.addNew.Period).endOf("month");
  const url = "/bagong/payroll/get-submission-detail";
  let param = {
    CompanyID: auth.appData.CompanyID,
    JournalTypeID: JournalTypeID,
    DateStart: dateStart,
    DateEnd: dateEnd,
    SiteID: SiteID,
  };

  axios.post(url, param).then(
    (r) => {
      data.record.Lines = r.data;
    },
    (e) => {}
  );
}

function getJournalType(site) {
  const url = "/bagong/siteentry/get-detail-ledger-journal-post";
  const siteID = site?.length == 0 ? "" : site[0].Value;
  const param = {
    SiteID: siteID,
    Type: "Payroll",
  };

  axios.post(url, param).then(
    (r) => {
      data.journalPost = r.data;
      generateSubmission(siteID);
      generateLines(r.data._id, siteID);
    },
    (e) => {
      util.showError(e);
    }
  );
}

function trxPreSubmit(status, action, doSubmit) {
  if (status == "DRAFT") {
    trxSubmit(doSubmit);
  }
}

function trxSubmit(doSubmit) {
  util.nextTickN(2, () => {
    listControl.value.submitForm(
      data.record,
      () => {
        doSubmit();
      },
      () => {}
    );
  });
}

function trxPostSubmit() {
  backToGrid();
}

function onPreSave(record) {
  record.CompanyID = auth.appData.CompanyID;
  record.SubmissionDate = record.TrxDate;
  if (data.addNew.Period && !record._id) {
    record.PayrollPeriod = data.addNew.Period
  }
  for (let i in record.Lines) {
    let obj = record.Lines[i];
    obj.Account.AccountType = "EXPENSE";
    obj.Account.AccountID = record.Expense;
  }
}

function onOpenForm(r) {
  if (r.Lines?.length > 0) r.Expense = r.Lines[0].Account.AccountID;
  r.JournalType = r.JournalTypeID;
  data.record = r;
}

const hideFormSave = computed(() => {
  return !["", "DRAFT"].includes(data.record.Status);
});

function backToGrid() {
  listControl.value.setControlMode("grid");
  listControl.value.refreshGrid();
}

function onPostSave(r) {
  if (gridAttachment.value) gridAttachment.value.Save();
}
function calcAmount(item) {
  const total = item.Lines.reduce((sum, item) => sum + item.Amount, 0);
  return helper.formatNumberWithDot(total);
}
// watch(
//   () => data.search,
//   (nv) => {
//     console.log(nv)
//     // refreshGrid();
//   },
//   { deep: true }
// );
</script>
<style scoped></style>
