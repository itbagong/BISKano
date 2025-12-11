<template>
  <div class="w-full">
    <data-list
      v-show="data.appMode == 'grid'"
      class="card"
      ref="listControl"
      :title="data.title"
      grid-config="/fico/cashjournal/gridconfig"
      form-config="/fico/cashjournal/formconfig"
      form-read="/fico/cashjournal/get"
      grid-read="/fico/pettycash/gets"
      grid-mode="grid"
      grid-delete="/fico/cashjournal/delete"
      form-keep-label
      form-insert="/fico/cashjournal/save"
      form-update="/fico/cashjournal/save"
      :grid-fields="['Enable', 'Status']"
      :form-tabs-new="data.tabsView"
      :form-tabs-edit="data.tabsEdit"
      :form-tabs-view="data.tabsView"
      :form-fields="[
        'Lines',
        'Dimension',
        'JournalTypeID',
        'CashBookID',
        'PostingProfileID',
      ]"
      grid-sort-field="TrxDate"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @grid-refreshed="onGridRefreshed"
      @alterGridConfig="onAlterGridConfig"
      @alterFormConfig="onAlterFormConfig"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @formFieldChange="formFieldChange"
      :formHideSubmit="readOnly"
      grid-hide-select
      stay-on-form-after-save
      @preSave="preSubmit"
      :grid-custom-filter="data.customFilter"
      @gridResetCustomFilter="resetGridHeaderFilter"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilter"
          v-model="data.customFilter"
          @initNewItem="initNewItemFilter"
          @preChange="changeFilter"
          @change="refreshGrid"
        >
          <template #filter_2="{ item }">
            <div class="min-w-[260px]">
              <dimension-editor
                multiple
                :default-list="profile.Dimension"
                v-model="item.Dimension"
                :required-fields="[]"
                :dim-names="['PC', 'Site']"
              ></dimension-editor>
            </div>
          </template>
        </grid-header-filter>
      </template>
      <template #form_buttons_1="{ item, inSubmission, loading }">
        <form-buttons-trx
          :disabled="inSubmission || loading"
          :auto-post="!waitTrxSubmit"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="data.jType"
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
      </template>
      <template #form_input_JournalTypeID_selected-option="{ option }">
        {{ option.item?.Name }}
      </template>
      <template #form_input_JournalTypeID_option="{ option }">
        {{ option.item?.Name }}
      </template>
      <template #form_input_CashBookID="{ item, config, mode }">
        <div class="flex items-end gap-2">
          <s-input
            class="w-full"
            v-model="item.CashBookID"
            :label="config.label"
            :use-list="config.useList"
            lookup-url="/tenant/cashbank/find?CashBankGroupID=BANK"
            :lookup-key="config.lookupKey"
            :lookup-labels="config.lookupLabels"
            :lookup-searchs="config.lookupLabels"
            :read-only="readOnly || mode == 'view'"
          ></s-input>
          <a href="#" v-if="item.CashBookID" class="border-gray-400 border p-[7px] rounded-md hover:border-red-500" @click="redirect(item.CashBookID)" >
            <mdicon
              name="open-in-new"
              class="text-gray-600 hover:text-red-500"
              size="16"
            />
          </a>
        </div>
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_input_PostingProfileID="{ item, config }">
        <div>
          <label class="input_label">
            <div>{{ config.label }}</div>
          </label>
          <div class="bg-transparent">{{ item.PostingProfileID }}</div>
        </div>
      </template>
      <template #form_tab_Line="{ item, mode }">
        <journal-line
          grid-hide-detail
          v-model="item.Lines"
          @calc="calcLineTotal"
          @gridRowFieldChanged="onGridRowFieldChangedLine"
          :read-only="readOnly || mode == 'view'"
          :grid-config-url="
            item.CashJournalType === 'CASH IN'
              ? 'fico/cashin/line/gridconfig'
              : 'fico/cashout/line/gridconfig'
          "
          @new-record="newRecordLine"
          :attch-kind="data.jType"
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.jType"
          @preOpenAttch="preOpenAttchLine"
          @closeAttch="closeAttchLine"
          @postDelete="postDeleteLine"
        >
          <template #grid_Account="p">
            <AccountSelector
              v-model="p.item.Account"
              :items-type="
                item.CashJournalType === 'CASH IN'
                  ? ['COA', 'CST']
                  : ['EXP']
              "
              :group-id-value="['EXG0007']"
              :read-only="readOnly || mode == 'view'"
            ></AccountSelector>
          </template>
        </journal-line>
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          ref="dimenstionCtl"
          v-model="item.Dimension"
          :read-only="readOnly || mode == 'view'"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #form_tab_Submission_History="{ item, mode }">
        <submission-history
          :items="data.submissionHistory"
          :read-only="readOnly || mode == 'view'"
        />
      </template>
      <template #form_tab_References="{ item, mode }">
        <ref-template
          :ReferenceTemplate="data.jurnalType.ReferenceTemplateID"
          :read-only="readOnly || mode == 'view'"
          v-model="item.References"
          @get-items="getRefItems"
        />
      </template>
      <template #form_tab_Attachment="{ item }">
        <attachment
          ref="gridAttachment"
          v-model="item.Attachment"
          :journalId="item._id"
          :journalType="data.jType"
          single-save
      /></template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #grid_item_button_delete="{item}">  
        <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
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
      :hide-signature="false"
      :hide-section-title="true"
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
          <div>
            <action-attachment
              ref="PreviewAction"
              :kind="data.jType"
              :ref-id="data.record._id"
              hide-button
              actionOnHeader
              read-only
            />
          </div>
        </div>
      </template>
    </PreviewReport>
  </div>
</template>

<script setup>
import { reactive, ref, inject, computed, watch } from "vue";
import { authStore } from "@/stores/auth";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SButton, SInput } from "suimjs";
import helper from "@/scripts/helper.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import JournalLine from "@/components/common/JournalLine.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";
import References from "@/components/common/References.vue";
import RefTemplate from "./widget/RefTemplate.vue";
import SubmissionHistory from "./widget/SubmissionPettyCashHistory.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";

import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import Attachment from "@/components/common/SGridAttachment.vue";

import { useRouter, useRoute } from 'vue-router';

layoutStore().name = "tenant";

const FEATUREID = "PettyCashSubmission";
const profile = authStore().getRBAC(FEATUREID);

const auth = authStore();

const listControl = ref(null);
const dimenstionCtl = ref(null);
const gridAttachment = ref(null);
const gridHeaderFilter = ref(null);

const axios = inject("axios");
const route = useRoute()
const router = useRouter()

const data = reactive({
  trxType: "CASH IN",
  title: "Petty Cash Submission",
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  record: {},
  tabsNew: ["General"],
  tabsEdit: [
    "General",
    "Line",
    "Submission History",
    "References",
    "Attachment",
  ],
  tabsView: [
    "General",
    "Line",
    "Submission History",
    "References",
    "Attachment",
  ],
  search: {
    CompanyID: "",
    Text: "",
    StartDate: null,
    EndDate: null,
    Skip: 0,
    Take: 5,
  },
  referenceTemplateID: "",
  submissionHistory: [],
  customFilter: null,
  jurnalType: {},
  jType: "CASHBANK",
  jtId: "CJT-003",
});

const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});

const linesTag = computed({
  get() {
    return data.record.Lines?.map((e) => {
      return `${data.jType}_${data.record._id}_${e.LineNo}`;
    });
  },
});
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "SUBMITTED"].includes(data.record.Status);
  },
});

function preOpenAttchLine(readOnly) {
  if (readOnly) return;
  preSubmit(data.record);
  axios.post("/fico/cashjournal/update", data.record);
}
function closeAttchLine(readOnly) {
  if (readOnly) return;
  gridAttachment.value.refreshGrid();
}
function postDeleteLine() {
  gridAttachment.value.refreshGrid();
}
function initNewItemFilter(item) {
  item.Dimension = [];
}

function changeFilter(item, filters) {
  helper.genFilterDimension(item.Dimension).forEach((e) => {
    filters.push(e);
  });
}
function refreshGrid() {
  listControl.value.refreshGrid();
}
function resetGridHeaderFilter() {
  gridHeaderFilter.value.reset();
}

function getSubmissionHistory(record) {
  const url = "/fico/pettycash/get-submission-history";
  const payload = {
    CompanyID: "",
    JournalTypeID: record.JournalTypeID,
    SiteID: record.Dimension.reduce((val, obj) => {
      if (obj.Key === "Site") val = obj.Value;
      return val;
    }, ""),
    CashBankID: record.CashBookID,
    TrxDate: record.TrxDate,
  };
  axios.post(url, payload).then(
    (r) => {
      data.submissionHistory = r.data;
    },
    (e) => {
      util.showError(e);
    }
  );
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr(
      "JournalTypeID",
      "lookupUrl",
      "/fico/cashjournaltype/find?_id=" + data.jtId
    );

    if (readOnly.value === true) {
      data.formMode = "view";
      listControl.value.setFormMode("view");
    } else {
      listControl.value.setFormMode(data.formMode);
    }
  });
}

function newRecord(record) {
  data.formMode = "new";
  record._id = "";
  record.CashInDate = new Date();
  record.CurrencyID = "IDR";
  record.Status = "DRAFT";
  record.Lines = [];
  record.JournalTypeID = data.jtId;
  record.PostingProfileID = "";
  record.CompanyID = auth.companyId;
  record.CashJournalType = "CASH IN";
  record.TrxDate = new Date();
  data.record = record;
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.record = record;

  if (record.References.length > 0) {
    let isPettyCash = record.References.find(
      (o) => o.Key == "Submission Type" && o.Value == "Petty Cash"
    );
    data.tabsEdit = isPettyCash
      ? ["General", "Line", "Submission History", "References", "Attachment"]
      : ["General", "Line", "References", "Attachment"];
    data.tabsView = isPettyCash
      ? ["General", "Line", "Submission History", "References", "Attachment"]
      : ["General", "Line", "References", "Attachment"];
  }

  openForm(record);
  getSubmissionHistory(record);
}

function preSubmit(record) {
  record.Lines?.map((e) => {
    e.Dimension = record.Dimension;
    e.OffsetAccount = {
      AccountType: "CASHBANK",
      AccountID: record.CashBookID,
    };
    return e;
  });
}
function closePreview() {
  data.appMode = "grid";
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
function trxErrorSubmit(e) {
  setLoadingForm(false);
}

function trxPostSubmit(record) {
  setLoadingForm(false);
  closePreview()
  setModeGrid();
}

function setModeGrid() {
  listControl.value.setControlMode("grid");
  listControl.value.gridResetFilter();
  listControl.value.refreshList();
}

function calcLineTotal(total = { PriceEach: 0, Qty: 0, Amount: 0 }) {
  data.record.Amount = parseInt(total.Amount);
  data.record.ReportingAmount = parseInt(total.Amount);
}

function onGridRowFieldChangedLine(name, v1, v2, old, record, onUpdatedRow) {
  if (name == "PaymentType") {
    record.ChequeGiroID = "";
    onUpdatedRow();
  }
}

function getJurnalType(id) {
  if (id === "" || id === null) {
    data.jurnalType = {};
    data.record.PostingProfileID = "";
    return;
  }
  listControl.value.setFormLoading(true);
  axios
    .post("/fico/cashjournaltype/get", [id])
    .then(
      (r) => {
        data.jurnalType = r.data;
        data.record.PostingProfileID = r.data.PostingProfileID;
      },
      (e) => {
        data.jurnalType = {};
        data.record.PostingProfileID = "";
        util.showError(e);
      }
    )
    .finally(() => {
      listControl.value.setFormLoading(false);
    });
}
function formFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "JournalTypeID":
      getJurnalType(v1);
      break;
  }
}

function getRefItems(records) {
  const findIdx = records.findIndex((data) => data.ReferenceType == "items");
  if (findIdx > -1) {
    util.nextTickN(2, () => {
      const confValue = records[findIdx].ConfigValue.split("|");
      data.record.References[findIdx].Key = records[findIdx].Label;
      data.record.References[findIdx].Value = confValue[0];
    });
  }
}

function newRecordLine(r) {
  console.log(data.record)
  r.Qty = 1;
  r.Account = {
    AccountType: "EXPENSE",
    AccountID: "EX00047",
  };
  if (r.References == undefined) r.References = [];
  r.References.push({
    Key: "HeaderID",
    Value: data.record._id,
  });
}

function redirect(AccountID) {
  const url = router.resolve({ name: "fico-CashBank", query: { id: AccountID } });
  window.open(url.href, "_blank");
}
function onAlterFormConfig(cfg) {
  if (route.query.id !== undefined) {
    let currQuery = {...route.query};
    listControl.value.selectData({ _id: currQuery.id });
    delete currQuery['id'];
    router.replace({ path: route.path, query: currQuery });
  }
}
watch(
  () => data.record,
  (nv) => {
    if (nv) {
      util.nextTickN(2, () => {
        getJurnalType(nv.JournalTypeID);
      });
    }
  }
);
</script>
