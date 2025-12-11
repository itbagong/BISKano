<template>
  <div class="w-full">
    <data-list
      v-show="data.showApply === false && data.appMode == 'grid'"
      class="card"
      ref="listControl"
      :title="data.title"
      grid-config="/fico/cashjournal/gridconfig"
      form-config="/fico/cashjournal/formconfig"
      :grid-read="`/fico/cashjournal/gets?CashJournalType=${type}`"
      form-read="/fico/cashjournal/get"
      grid-mode="grid"
      grid-delete="/fico/cashjournal/delete"
      form-keep-label
      form-insert="/fico/cashjournal/save"
      form-update="/fico/cashjournal/save"
      :grid-fields="['Enable', 'Status']"
      grid-sort-field="TrxDate"
      grid-sort-direction="desc"
      :form-tabs-new="data.tabsNew"
      :form-tabs-edit="data.tabsEdit"
      :form-tabs-view="data.tabsView"
      :form-fields="['Lines', 'Dimension', 'JournalTypeID', 'PostingProfileID']"
      :grid-custom-filter="data.customFilter"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @formFieldChange="formFieldChange"
      :formHideSubmit="readOnly"
      grid-hide-select
      stay-on-form-after-save
      @preSave="preSubmit"
      @gridResetCustomFilter="resetGridHeaderFilter"
      @post-save="postSave"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @alterGridConfig="onAlterGridConfig"
      @alter-form-config="alterFormConfig"
    >
      <template #form_buttons_1="{ item, inSubmission, loading, mode }">
        <form-buttons-trx
          :disabled="inSubmission || loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="data.jType"
          :auto-post="!waitTrxSubmit"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
        />
        <template v-if="profile.canUpdate && mode !== 'new'">
          <s-button
            :disabled="inSubmission || loading"
            v-if="!readOnly"
            label="Apply"
            class="btn_primary"
            @click="openApply"
          ></s-button>
          <s-button
            :disabled="inSubmission || loading"
            v-if="!readOnly && isSubmission && mode == 'edit'"
            label="Action"
            class="btn_primary"
            @click="data.actionModal = true"
          />
          <s-button
            :disabled="inSubmission || loading"
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Preview"
            @click="data.appMode = 'preview'"
          ></s-button>
        </template>
      </template>
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilter"
          v-model="data.customFilter"
          filterTextWithId
          @initNewItem="initNewItemFilter"
          @preChange="changeFilter"
          @change="refreshGrid"
        >
          <!-- <template #filter_1="{ item }">
            <s-input
              v-if="['CASH OUT'].includes(props.journalType)"
              class="w-[200px]"
              label="Vendor"
              use-list
              lookup-url="/bagong/vendor/get-vendor-active"
              lookup-key="_id"
              :lookup-labels="['_id', 'Name']"
              :lookupSearchs="['_id', 'Name']"
              v-model="item.VendorIDs"
              multiple
            />
          </template> -->
          <template #filter_2="{ item }">
            <div
              class="min-w-[200px]"
              v-if="['CASH IN', 'CASH OUT'].includes(props.journalType)"
            >
              <dimension-editor
                multiple
                :default-list="profile.Dimension"
                v-model="item.Dimension"
                :required-fields="[]"
                :dim-names="['Site']"
              ></dimension-editor>
            </div>
          </template>
        </grid-header-filter>
      </template>

      <template #form_input_JournalTypeID_selected-option="{ option }">{{
        readOnly ? option.item : option.item?.Name
      }}</template>
      <template #form_input_JournalTypeID_option="{ option }">{{
        option.item?.Name
      }}</template>
      <template #form_input_PostingProfileID="{ item, config }">
        <div>
          <label class="input_label">
            <div>{{ config.label }}</div>
          </label>
          <div class="bg-transparent">{{ item.PostingProfileID }}</div>
        </div>
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_tab_Line="{ item, mode }">
        <journal-line
          ref="journalLine"
          grid-hide-detail
          v-model="item.Lines"
          @calc="calcLineTotal"
          @gridRowFieldChanged="onGridRowFieldChangedLine"
          @newRecord="newRecordLine"
          :read-only="readOnly || mode == 'view'"
          :grid-config-url="urlLineGridConfig"
          :attch-kind="data.jType"
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.jType"
          @preOpenAttch="preOpenAttchLine"
          @closeAttch="closeAttchLine"
          @postDelete="postDeleteLine"
        >
          <template #grid_item_buttons_1="p">
            <template v-if="props.type === 'CASH OUT'">
              <s-button
                icon="note-plus-outline"
                name="apply"
                alt="apply"
                class="cursor-pointer hover:text-success mt-[-3px]"
                width="16"
                @click="onCheque(p.item)"
                v-if="
                  p.item.PaymentType &&
                  ['PTY002', 'PTY003'].includes(p.item.PaymentType) &&
                  p.item.ChequeGiroID == ''
                "
              />
            </template>
          </template>
          <template #grid_ChequeGiroID="p">
            <template v-if="props.type === 'CASH OUT'">
              <div v-if="p.item.PaymentType === 'PTY004'">
                <s-input
                  v-model="p.item.ChequeGiroID"
                  use-list
                  :lookup-url="`/bagong/vendor/get-vendor-bank?VendorID=${p.item.Account.AccountID}`"
                  lookup-key="_id"
                  :lookup-labels="['BankName', 'BankAccountNo']"
                  :lookup-searchs="['_id', 'BankName']"
                />
              </div>
              <div v-else>
                {{ p.item.ChequeGiroID }}
              </div>
            </template>
          </template>
          <template #grid_Account="p" v-if="!isSubmission">
            <AccountSelector
              v-model="p.item.Account"
              :items-type="
                props.type === 'CASH IN'
                  ? ['COA', 'CST']
                  : ['COA', 'VND', 'EXP']
              "
              :read-only="
                (readOnly || mode == 'view') &&
                item.Status != 'DRAFT' &&
                !(ledgerEditor && item.Status == 'READY')
              "
            ></AccountSelector>
          </template>
          <template #grid_Ignore="p">
            <s-input
              v-if="enableIgnoreLine"
              :kind="p.header.input.kind"
              v-model="p.item.Ignore"
              @change="
                (name, v1, v2, old) => onIgnoreLine(name, v1, v2, old, p.item)
              "
            />
          </template>
        </journal-line>
      </template>
      <template #form_tab_Documents="{ item, mode }">
        <attachment
          :read-only="readOnly || mode == 'view'"
          ref="gridAttachment"
          v-model="item.Attachment"
          single-save
          :tags="linesTag"
          :journalId="item._id"
          :journalType="data.jType"
          @preSave="preSaveAttachment"
        />
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          ref="dimenstionCtl"
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :read-only="readOnly || mode == 'view'"
        ></dimension-editor-vertical>
      </template>
      <template v-for="tabName in formTabNames" v-slot:[tabName]="{ item }">
        <slot
          :name="tabName"
          :item="item"
          :config="{
            formMode: data.formMode,
            appMode: data.appMode,
          }"
          >{{ tabName }}</slot
        >
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #grid_item_button_delete="{ item }">
        <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
      </template>
    </data-list>
    <apply-list
      v-if="data.showApply"
      draft
      hide-from-filter
      hide-to-filter
      source-type="CashBank"
      :source-journal-id="data.record._id"
      @back="data.showApply = false"
      @save-apply="onSaveApply"
    />
    <cheque
      ref="cheque"
      :modal-apply="data.showCheque"
      :journal="data.record"
      :lines="data.record.Lines"
    />
    <PreviewReport
      v-if="data.appMode == 'preview'"
      class="card w-full"
      title="Preview"
      @close="closePreview"
      :disable-print="helper.isDisablePrintPreview(data.record.Status)"
      :SourceType="['CASH OUT', 'CASH IN'].includes(journalType) ? 'CASHBANK': journalType"
      :SourceJournalID="data.record._id"
      :hideSignature="false"
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
              :tags="linesTag"
              hide-button
              actionOnHeader
              read-only
            />
          </div>
        </div>
      </template>
    </PreviewReport>
    <s-modal
      :display="data.actionModal"
      hideButtons
      title="Submission"
      @beforeHide="data.actionModal = false"
    >
      <action-submission
        ref="actionSubmission"
        :type="type"
        @line-list="onSetLine"
        :site-id="helper.findDimension(data.record.Dimension, 'Site')"
      />
    </s-modal>
  </div>
</template>

<script setup>
import { reactive, ref, inject, computed, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { authStore } from "@/stores/auth";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SButton, SInput, SModal } from "suimjs";
import helper from "@/scripts/helper.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import JournalLine from "@/components/common/JournalLine.vue";
//import ApplyList from "../Apply/List.vue";
import ApplyList from "../TrxApply/Apply.vue";
import Cheque from "./ChequeGiro.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import ActionSubmission from "./ActionSubmition.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";

const props = defineProps({
  type: { type: String, default: "CASH IN" },
  journalType: { type: String, default: "CASH IN" },
  title: { type: String, default: "" },
  additionalFormTabsNew: { type: Array, default: [] },
  additionalFormTabsEdit: { type: Array, default: [] },
  additionalFormTabsView: { type: Array, default: [] },
  featureID: { type: String, default: "" },
  isSubmission: { type: Boolean, default: false },
  urlLineGridConfig: { type: String, default: "" },
});
layoutStore().name = "tenant";

const auth = authStore();
const FEATUREID = props.featureID;
const profile = auth.getRBAC(FEATUREID);
const ledgerEditor = auth.getRBAC("EditLedgerBeforePost").canSpecial1;

const gridHeaderFilter = ref(null);
const dimenstionCtl = ref(null);
const listControl = ref(null);
const journalLine = ref(null);
const cheque = ref(null);
const gridAttachment = ref(null);

const axios = inject("axios");
const router = useRouter();
const route = useRoute();

const data = reactive({
  showApply: false,
  showCheque: false,
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  title:
    props.title == "" || props.title == undefined ? props.type : props.title,
  record: {},
  tabsNew: Array.isArray(props.additionalFormTabsNew)
    ? [...["General"], ...props.additionalFormTabsNew]
    : ["General"],
  tabsEdit: Array.isArray(props.additionalFormTabsEdit)
    ? [...["General", "Line", "Documents"], ...props.additionalFormTabsEdit]
    : ["General", "Line", "Documents"],
  tabsView: Array.isArray(props.additionalFormTabsView)
    ? [...["General", "Line", "Documents"], ...props.additionalFormTabsView]
    : ["General", "Line", "Documents"],
  record: {},
  customFilter: null,
  filterSite: null,
  actionModal: false,
  jType: "CASHBANK",
  isLatestApprover: false,
  ignoredLineIds: [],
  deletedLineIds: [],
  dataApplyTo: [],
});

function initNewItemFilter(item) {
  item.VendorIDs = [];
  item.Dimension = [];
}
function changeFilter(item, filters) {
  if (item.VendorIDs.length > 0) {
    filters.push({
      Op: "$in",
      Field: "VendorID",
      Value: item.VendorIDs,
    });
  }
  helper.genFilterDimension(item.Dimension).forEach((e) => {
    filters.push(e);
  });
}

const formTabNames = computed({
  get() {
    let tabNames =
      data.formMode == "new"
        ? data.tabsNew
        : data.formMode == "edit"
        ? data.tabsEdit
        : data.tabsView;
    return tabNames.slice(3, tabNames.length).map((el) => {
      return "form_tab_" + el.replace(" ", "_");
    });
  },
});

const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});

const enableIgnoreLine = computed({
  get() {
    return (
      data.record.Status == "READY" ||
      (data.record.Status == "SUBMITTED" && data.isLatestApprover)
    );
  },
});
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "SUBMITTED", "READY"].includes(data.record.Status);
  },
});
const linesTag = computed({
  get() {
    const linetags = data.record.Lines?.map((e) => {
      return `${data.jType}_${data.record._id}_${e.LineNo}`;
    });
    const refTags = data.record.Lines?.reduce((acc, e) => {
      const headerid = e.References.find((o) => o.Key === "HeaderID");
      if (headerid?.Value) {
        acc.push(`${data.jType}_${headerid.Value}`);
      }
      return acc;
    }, []);
    const appytags = [];
    data.dataApplyTo?.forEach((o) => {
      appytags.push(`${o.Module}_${o.JournalID}`);
      if (o.Module === "VENDOR") {
        appytags.push(`VENDOR_VENDOR ${o.JournalID}`);
      }
    });
    return [
      `${data.jType}_${data.record._id}`,
      ...linetags,
      ...refTags,
      ...appytags,
    ];
  },
});

function preOpenAttchLine() {
  preSubmit(data.record);
  axios.post("/fico/cashjournal/save", data.record);
}
function closeAttchLine() {
  gridAttachment.value.refreshGrid();
}
function refreshLine() {
  journalLine.value?.refresh();
}

function onSaveApply() {
  // journalLine.value?.refresh();
  setLoadingForm(true);
  axios
    .post("/fico/cashjournal/get", [data.record._id])
    .then(
      (r) => {
        refreshFormRecord(r.data);
      },
      (err) => util.showError(e)
    )
    .finally(() => {
      setLoadingForm(false);
      util.nextTickN(2, () => {
        journalLine.value.refresh();
      });
    });
}

function onIgnoreLine(name, v1, v2, old, record) {
  const headerId = record.References.find((e) => e.Key === "HeaderID")?.Value;

  if (headerId != "" && headerId != data.record._id) {
    data.record.Lines.forEach((e) => {
      const r = e.References.filter(
        (e) => e.Key === "HeaderID" && e.Value === headerId
      );
      if (r.length > 0) e.Ignore = v1;
    });

    util.nextTickN(2, () => {
      refreshLine();
    });
  }
}
function postDeleteLine(deletedRecord) {
  const headerId = deletedRecord.References.find(
    (e) => e.Key === "HeaderID"
  )?.Value;

  if (headerId != "" && headerId != data.record._id) {
    data.deletedLineIds.push(headerId);
    data.record.Lines = data.record.Lines.filter((e) => {
      const r = e.References.filter(
        (e) => e.Key === "HeaderID" && e.Value === headerId
      );
      return r.length == 0;
    });

    util.nextTickN(2, () => {
      refreshLine();
    });
  }
  gridAttachment.value.refreshGrid();
}

function refreshGrid() {
  listControl.value.refreshGrid();
}
function resetGridHeaderFilter() {
  gridHeaderFilter.value.reset();
}
function fetchIsLastApprover() {
  data.isLatestApprover = false;
  axios
    .post("/fico/postingprofile/is-latest-approver", {
      SourceID: data.record._id,
      SourceType: data.jType,
    })
    .then(
      (r) => {
        data.isLatestApprover = r.data.LatestApprover;
      }
      // (err) => util.showError(err)
    );
}
function getApplyTo(record) {
  axios
    .post("/fico/apply/get-apply-to", {
      SourceJournalID: record._id,
    })
    .then(
      (r) => {
        data.dataApplyTo = r.data;
        util.nextTickN(2, () => {
          gridAttachment.value.refreshGrid();
        });
      }
      // (err) => util.showError(err)
    );
}
function openForm(record) {
  data.deletedLineIds = [];
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("PostingProfileID", "readOnly", true);
    listControl.value.setFormFieldAttr(
      "JournalTypeID",
      "lookupUrl",
      "/fico/cashjournaltype/find?TransactionType=" + props.journalType
    );
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
    if (data.record.Status == "SUBMITTED") {
      fetchIsLastApprover();
    }

    //   ["READY","SUBMITTED"].includes(data.record.Status)) {
    //   fetchIsLastApprover();
    // }
  });
}
function newRecord(record) {
  record._id = "";
  record.CashInDate = new Date();
  record.CurrencyID = "IDR";
  record.Status = "DRAFT";
  record.Lines = [];
  record.JournalTypeID = "";
  record.PostingProfileID = "";
  record.CompanyID = auth.companyId;
  record.TrxDate = new Date();
  record.CashJournalType = props.journalType;
  data.record = record;
  openForm(record);
}

function editRecord(record) {
  data.record = record;
  openForm(record);
  getApplyTo(record);
}

function preSubmit(record) {
  record.Lines?.map((e) => {
    if (props.journalType == "CASH OUT") {
      e.OffsetAccount = {
        AccountType: "CASHBANK",
        AccountID: record.CashBookID,
      };
    }
    e.Dimension = record.Dimension;
    return e;
  });
}

function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    listControl.value.setFormFieldAttr(e.field, "required", required);
  });
}

function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function openApply() {
  setLoadingForm(true);
  listControl.value.submitForm(
    data.record,
    () => {
      setLoadingForm(false);
      data.showApply = true;
    },
    () => {
      setLoadingForm(false);
    },
    true
  );
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
          setLoadingForm(false);
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
function trxPostSubmit(record) {
  closePreview()
  postSaveAttachment();

  if (data.record.Status == "SUBMITTED") {
    let ignoredLineIds = [];
    data.record.Lines.forEach((e) => {
      if (e.Ignore) {
        const headerId = e.References.find((e) => e.Key == "HeaderID")?.Value;
        if (
          headerId != "" &&
          headerId != data.record._id &&
          !ignoredLineIds.includes(headerId)
        )
          ignoredLineIds.push(headerId);
      }
    });
    if (ignoredLineIds.length > 0) updateReferences(ignoredLineIds);
  }
  setLoadingForm(false);
  setModeGrid();
}
function trxErrorSubmit(e) {
  setLoadingForm(false);
}
function refreshFormRecord(record) {
  listControl.value.setFormRecord(record);
  data.record = record;
  util.nextTickN(2, () => {
    getApplyTo(record);
  });
}
function closeApply(record) {
  data.showApply = false;
  setLoadingForm(true);
  axios
    .post("/fico/cashjournal/get", [data.record._id])
    .then(
      (r) => {
        refreshFormRecord(r.data);
      },
      (err) => util.showError(e)
    )
    .finally(() => {
      setLoadingForm(false);
      util.nextTickN(2, () => {
        journalLine.value.refresh();
      });
    });
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
function calcTotalAmount(totalAmount) {
  data.record.Amount = parseInt(totalAmount);
  data.record.ReportingAmount = parseInt(totalAmount);
}
function onCheque(record) {
  cheque.value.onApplyCheque(record);
}

function changeJournalTypeID(name, v1, v2, old, record) {
  let postingProfileID;
  let type = "";
  if (record[name] == "" || record[name] == null) {
    postingProfileID = "";
    type = "";
  } else {
    const arr = v2.split(" | ");
    const ppID = arr[1];
    type = arr[2];
    postingProfileID = ppID == undefined ? "" : ppID;
  }
  record["PostingProfileID"] = postingProfileID;
}

function formFieldChange(name, v1, v2, old, record) {
  if (name == "JournalTypeID") changeJournalTypeID(name, v1, v2, old, record);
}
function onGridRowFieldChangedLine(name, v1, v2, old, record, onUpdatedRow) {
  if (name == "PaymentType") {
    record.ChequeGiroID = "";
    onUpdatedRow();
  }
}

async function postSave(record) {
  for (let i in data.record.Lines) {
    let obj = data.record.Lines[i];
    if (obj._id) {
      obj.LineNo = parseInt(i);
      await reserveCashOut(obj);
    }
  }
  if (data.deletedLineIds.length > 0) {
    updateReferences(data.deletedLineIds);
  }
}
function newRecordLine(record) {
  record.Qty = 1;
  if (record.References == undefined) record.References = [];
  record.References.push({
    Key: "HeaderID",
    Value: data.record._id,
  });
}
async function reserveCashOut(param) {
  const url = "/fico/cg/reserve";
  axios.post(url, param).then(
    (r) => {},
    (e) => {}
  );
}

function onSetLine(dt) {
  data.actionModal = false;

  const newLines = dt.map((e, i) => {
    return { ...e, LineNo: data.record.Lines.length + i + 1 };
  });

  data.record.Lines = [...data.record.Lines, ...newLines];

  util.nextTickN(2, () => {
    refreshLine();

    setLoadingForm(true);

    const cbSuccess = () => {
      setLoadingForm(false);
      const ids = newLines.reduce((arr, e) => {
        const r = e.References.find((e) => e.Key === "HeaderID")?.Value;
        if (r != "" && !arr.includes(r)) arr.push(r);
        return arr;
      }, []);
      updateReferences(ids, [
        { key: "SubmissionJournalID", value: data.record._id },
      ]);
    };

    const cbError = () => {
      setLoadingForm(false);
    };
    util.nextTickN(2, () => {
      listControl.value.submitForm(data.record, cbSuccess, cbError);
    });
  });
}

function updateReferences(ids, references = []) {
  let param = {
    IDs: ids,
    References: references,
  };

  axios.post("/fico/cashjournal/update-references", param).then(
    (r) => {},
    (e) => {}
  );
}

function onAlterGridConfig(config) {
  config.setting.sortable = ["TrxDate", "Created"];
}
function alterFormConfig(config) {
  if (route.query.trxid !== undefined) {
    let currQuery = { ...route.query };
    listControl.value.selectData({ _id: currQuery.trxid }); //remark sementara tunggu suimjs update
    delete currQuery["trxid"];
    if (route.query.id) {
      router.replace({ path: route.path, query: currQuery });
    } else {
      router.replace({ path: route.path });
    }
  }
}
const addTags = computed({
  get() {
    return [`${data.jType}_${data.record._id}`];
  },
});
function preSaveAttachment(param) {
  param.Asset = {
    ...param.Asset,
    ...{ Tags: [`${data.jType}_${data.record._id}`] },
  };
}
function postSaveAttachment() {
  let tags = [...linesTag.value];
  data.record?.Attachment?.forEach((o) => {
    const tagAttc = o.Tags.filter((item) => {
      return tags.some((keyword) => item.includes(keyword));
    });
    tags = [...tags, ...tagAttc];
  });

  const payload = {
    Addtags: addTags.value,
    Tags: tags,
  };
  if (payload.Tags.length > 0) {
    helper.updateTags(axios, payload);
  }
}
function closePreview() {
  data.appMode = "grid";
}
</script>
