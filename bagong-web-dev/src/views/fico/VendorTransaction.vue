<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="cardTitle"
      v-show="data.appMode == 'grid'"
      grid-config="/fico/vendorjournal/gridconfig"
      form-config="/fico/vendorjournal/formconfig"
      grid-read="/fico/vendorjournal/gets"
      form-read="/fico/vendorjournal/get"
      grid-mode="grid"
      grid-delete="/fico/vendorjournal/delete"
      form-keep-label
      form-insert="/fico/vendorjournal/insert"
      form-update="/fico/vendorjournal/update"
      :grid-fields="['TransactionType', 'TotalAmount', 'Enable', 'Status']"
      :grid-custom-filter="data.customFilter"
      :form-tabs-edit="[
        'General',
        'Address',
        'Lines',
        'References',
        'Checklist',
        'Attachment',
        'Logs',
      ]"
      :form-tabs-view="[
        'General',
        'Address',
        'Lines',
        'References',
        'Checklist',
        'Attachment',
        'Logs',
      ]"
      grid-sort-field="TrxDate"
      grid-sort-direction="desc"
      :form-fields="[
        'VendorID',
        'DefaultOffset',
        'Dimension',
        'PostingProfileID',
        'Status',
        'ExpectedDate',
        'TrxDate',
        'DiscountAmount',
        'HeaderDiscountAmount',
        'PPHAmount',
      ]"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      stay-on-form-after-save
      @formNewData="newRecord"
      @formEditData="editRecord"
      @form-field-change="onFormFieldChange"
      @controlModeChanged="onControlModeChanged"
      :formHideSubmit="readOnly"
      @preSave="preSubmit"
      @gridResetCustomFilter="resetHeaderFilter"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @post-save="onPostSave"
      @alter-grid-config="alterGridConfig"
      @alter-form-config="alterFormConfig"
    >
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilter"
          v-model="data.customFilter"
          filterTextWithId
          @initNewItem="initNewItemFilter"
          @preChange="changeFilter"
          @change="refreshGrid"
        >
          <template #filter_1="{ item }">
            <s-input
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
          </template>
          <template #filter_2="{ item }">
            <div class="min-w-[300px]">
              <dimension-editor
                multiple
                :default-list="profile.Dimension"
                v-model="item.Dimension"
                :required-fields="[]"
                :dim-names="['Site', 'CC']"
                :custom-labels="{
                  CC: 'CC / Department'
                }"
              ></dimension-editor>
            </div>
            <s-input
              label="Transaction type"
              kind="text"
              class="w-[200px]"
              v-model="item.TransactionType"
              :allow-add="false"
              use-list
              :items="[
                'Vendor Purchase',
                'Credit Note',
                'Good Receive',
                'Site Entry Expense',
                'Employee Expense',
                'General Submission',
              ]"
            />
          </template>
        </grid-header-filter>
      </template>
      <template #grid_TransactionType="{ item }">
        <div v-if="item.TransactionType">{{ item.TransactionType }}</div>
        <div v-else>{{ getSubmissionType(item.References) }}</div>
      </template>
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
          :trx-type="item.TransactionType"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :auto-post="!waitTrxSubmit"
        />
        <template v-if="profile.canUpdate && mode !== 'new'">
          <s-button
            :disabled="
              inSubmission ||
              loading ||
              item.VendorID == undefined ||
              item.JournalTypeID != 'VJT-002'
            "
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Action"
            @click="data.modalAction = true"
            v-if="!readOnly"
          ></s-button>
          <s-button
            :disabled="inSubmission || loading"
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Preview"
            @click="data.appMode = 'preview'"
          ></s-button>
        </template>
      </template>
      <template #form_input_VendorID="{ item, config }">
        <s-input
          keep-label
          :label="config.label"
          v-model="item.VendorID"
          use-list
          lookup-url="/bagong/vendor/get-vendor-active"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookupSearchs="['_id', 'Name']"
          :field="config.field"
          @change="onChangeVendorID"
        ></s-input>
      </template>
      <template #form_input_DefaultOffset="{ item, mode }">
        <AccountSelector
          v-model="item.DefaultOffset"
          hide-account-type
          :read-only="readOnly || mode == 'view'"
        />
      </template>
      <template #form_input_DiscountAmount="{ item, config }">
        <s-input
          keep-label
          :label="config.label"
          v-model="item.DiscountAmountStr"
          read-only
        ></s-input>
      </template>
      <template #form_input_HeaderDiscountAmount="{ item, config }">
        <s-input
          keep-label
          :label="config.label"
          v-model="item.HeaderDiscountAmountStr"
          read-only
        ></s-input>
      </template>
      <template #form_input_PPHAmount="{ item, config }">
        <s-input
          keep-label
          :label="config.label"
          v-model="item.PPHAmountStr"
          read-only
        ></s-input>
      </template>
      <template #form_input_PostingProfileID="{ item, config }">
        {{ item.DefaultOffset }}
        <label class="input_label">{{ config.label }}</label>
        <div>
          {{ item.PostingProfileID === "" ? "-" : item.PostingProfileID }}
        </div>
      </template>
      <template #form_input_TrxDate="{ item, config, mode }">
        <s-input
          v-if="item.Status == 'READY'"
          :kind="config.kind"
          :label="config.label"
          v-model="item.TrxDate"
          keep-label
        />
      </template>
      <template #form_input_ExpectedDate="{ item, config }">
        <label class="input_label">{{ config.label }}</label>
        <div>
          {{
            item.ExpectedDate
              ? moment(item.ExpectedDate).format("DD/MM/YYYY")
              : "-"
          }}
        </div>
      </template>
      <template #grid_TotalAmount="{ item }">
        <div>{{ util.formatMoney(item.TotalAmount) }}</div>
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          ref="dimenstionCtl"
          v-model="item.Dimension"
          :read-only="readOnly || mode == 'view'"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #form_tab_Address="props">
        <Address
          v-model="props.item.AddressAndTax"
          :bank-detail="data.bankDetail"
          :items-bank="data.itemsBank"
          :readOnly="readOnly || props.mode == 'view'"
        ></Address>
      </template>
      <template #form_tab_Lines="{ item, mode }">
        <journal-line
          v-model="item.Lines"
          @calc="calcLineTotal"
          :read-only="readOnly || mode == 'view'"
          grid-config-url="/fico/vendorjournal/line/gridconfig"
          form-config-url="/fico/vendorjournal/line/formconfig"
          @new-record="newRecordLine"
          :key="item.Lines"
          :attch-kind="data.jType"
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.jType"
          @preOpenAttch="preOpenAttchLine"
          @closeAttch="closeAttchLine"
          @postDelete="postDeleteLine"
          show-references
          show-checklist
          :referenceTemplate="data.jurnalType?.ReferenceTemplateLineID"
          :checklistId="data.jurnalType?.ChecklistTemplateLineID"
        >
          <template #grid_Account="p">
            <AccountSelector
              v-model="p.item.Account"
              :items-type="['COA', 'EXP']"
              :read-only="
                item.Status !== 'READY' && (readOnly || mode == 'view')
              "
            ></AccountSelector>
          </template>
        </journal-line>
      </template>
      <template #form_tab_References="{ item, mode }">
        <References
          :ReferenceTemplate="data.jurnalType.ReferenceTemplateID"
          :readOnly="readOnly || mode == 'view'"
          v-model="item.References"
        />
      </template>
      <template #form_tab_Checklist="{ item, mode }">
        <Checklist
          v-model="item.ChecklistTemp"
          :checklist-id="data.jurnalType.ChecklistTemplateID"
          :readOnly="readOnly || mode == 'view'"
          :attch-kind="data.jType"
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.jType"
        />
      </template>

      <template #grid_header_buttons_2> </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #form_tab_Attachment="{ item, mode }">
        <attachment
          v-model="item.Attachment"
          :journalId="item._id"
          :journalType="data.jType"
          :tags="attachmentByTag(linesTag)"
          :read-only="readOnly || mode == view"
          ref="gridAttachment"
          single-save
        />
      </template>
      <template #grid_item_button_delete="{ item }">
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
              :tags="attachmentByTag(linesTag)"
              hide-button
              actionOnHeader
              read-only
            />
          </div>
        </div>
      </template>
    </PreviewReport>

    <s-modal
      :display="data.modalAction"
      hideButtons
      title="Action"
      @beforeHide="data.modalAction = false"
    >
      <action-vendor
        v-model="data.record"
        @close="onCloseAction"
        :jurnal-type="data.jType"
        :jurnal-id="data.record._id"
      />
    </s-modal>
  </div>
</template>

<script setup>
import { reactive, ref, computed, inject, watch, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SButton, SInput, SModal } from "suimjs";
import { authStore } from "@/stores/auth.js";
import helper from "@/scripts/helper.js";

import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";

import StatusText from "@/components/common/StatusText.vue";
import Address from "./widget/vendorjournal/Address.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import References from "@/components/common/References.vue";
import Checklist from "@/components/common/Checklist.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import JournalLine from "@/components/common/JournalLine.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import moment from "moment";
import ActionVendor from "./widget/vendorjournal/ActionVendor.vue";
import ActionAttachment from "@/components/common/ActionAttachment.vue";

layoutStore().name = "tenant";

const FEATUREID = "VendorJournal";
const profile = authStore().getRBAC(FEATUREID);

const dimenstionCtl = ref(null);
const customFilter = ref(null);
const listControl = ref(null);
const gridHeaderFilter = ref(null);
const gridAttachment = ref(null);

const axios = inject("axios");
const router = useRouter();
const route = useRoute();

const data = reactive({
  dataListMode: "grid",
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  record: {},
  preview: {},
  searchQuery: {
    VendorIDs: [],
    DateFrom: null,
    DateTo: null,
  },
  CustomFiltering: undefined,
  jurnalType: {},
  taxes: [],
  totalTaxIn: 0,
  totalTaxDec: 0,
  bankDetail: [],
  itemsBank: [],
  customFilter: null,
  modalAction: false,
  jType: "VENDOR",
});

function alterGridConfig(cfg) {
  cfg.fields.splice(
    7,
    0,
    helper.gridColumnConfig({
      field: "TotalAmount",
      label: "Amount",
      kind: "text",
    })
  );
}
function alterFormConfig(config) {
  if (route.query.id !== undefined) {
    let getUrlParam = route.query.id;
    listControl.value.selectData({ _id: getUrlParam }); //remark sementara tunggu suimjs update
    router.replace({ path: "/fico/VendorTransaction" });
  } else if (route.query.form == 1) {
    listControl.value.newData();
    router.replace({ path: "/fico/VendorTransaction" });
  }
}
const cardTitle = computed(() => {
  if (data.dataListMode == "grid") return "Vendor Journal";
  const formMode = listControl.value?.getFormMode();
  return formMode == "new"
    ? `Vendor Journal - ${formMode}`
    : "Vendor Journal - " +
        [
          data.record._id,
          data.record.AddressAndTax?.DeliveryName,
          util.formatMoney(data.record.TotalAmount),
        ].join(" | ");
});
const linesTag = computed({
  get() {
    return data.record.Lines?.map((e) => {
      return `${data.jType}_${data.record._id}_${e.LineNo}`;
    });
  },
});
const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "READY"].includes(data.record.Status);
  },
});

function attachmentByTag(linesTags) {
  return [`VENDOR_VENDOR ${data.record._id}`, ...linesTags];
}
function preOpenAttchLine(readOnly) {
  if (readOnly) return;
  preSubmit(data.record);
  axios.post("/fico/vendorjournal/update", data.record);
}
function closeAttchLine(readOnly) {
  if (readOnly) return;
  gridAttachment.value.refreshGrid();
}
function postDeleteLine() {
  gridAttachment.value.refreshGrid();
}
function initNewItemFilter(item) {
  item.VendorIDs = [];
  item.Dimension = [];
  item.TransactionType = "";
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
  if (item.TransactionType != "") {
    filters.push({
      Op: "$contains",
      Field: "TransactionType",
      Value: [item.TransactionType],
    });
  }
}
function refreshGrid() {
  listControl.value.refreshGrid();
}
function resetHeaderFilter() {
  gridHeaderFilter.value.reset();
}
function preSubmit(record) {
  record.Lines?.map((e) => {
    e.OffsetAccount = {
      AccountType: "VENDOR",
      AccountID: record.VendorID,
    };
    e.CurrencyID = "IDR";
    return e;
  });
}
function closePreview() {
  data.appMode = "grid";
}

async function GetVendorID(name) {
  try {
    const dataresposne = await axios.post("/tenant/vendor/find", {
      Where: {
        Op: "$contains",
        Field: "Name",
        Value: [name],
      },
    });

    return dataresposne.data[0];
  } catch (error) {
    util.showError(error);
  }
}

function calcLines() {
  const record = data.record;
  record.SubtotalAmount = record.Lines.reduce((a, b) => {
    return a + b.Amount;
  }, 0);
  record.TotalAmount = record.SubtotalAmount;
}

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Enable = true;
  // TODO: baca dari store, simpan as default dari masing2 tenant/company
  record.CurrencyID = "IDR";
  record.PaymentTermID = "";
  record.TrxDate = new Date();
  record.ExpectedDate = new Date();
  record.DeliveryDate = new Date();
  record.Actions = [];
  record.Previews = [];
  record.AddressAndTax = {};
  record.DefaultOffset = {};
  record.DiscountAmountStr = "(0)";
  record.HeaderDiscountAmountStr = "(0)";
  record.PPHAmountStr = "(0)";
  record.Status = "DRAFT";
  record.PostingProfileID = "";
  record.TaxInvoiceDate = new Date();
  openForm(record);
}
function editRecord(record) {
  openForm(record);
}

function getVendor(id, record) {
  if (id) {
    const url = "/bagong/vendor/get?_id=" + id;
    axios.post(url, [id]).then(
      (r) => {
        record.PaymentTermID = r.data.PaymentTermID;
        getPaymentTerms(record.PaymentTermID);
        conditionGetTaxCodes(record, r.data);
      },
      (e) => util.showError(e)
    );
  }
}

function getVendorBank(id) {
  if (id) {
    const url = "/bagong/vendor/get?_id=" + id;
    axios.post(url, [id]).then(
      (r) => {
        data.bankDetail = r.data.Detail.VendorBank;
        data.itemsBank = r.data.Detail.VendorBank?.map((item) => {
          return item.BankName;
        });
      },
      (e) => util.showError(e)
    );
  }
}

function getSubmissionType(items) {
  let find = items.find((v) => v.Key == "Submission Type");
  if (find) {
    return find.Value;
  } else {
    return "";
  }
}

function conditionGetTaxCodes(record, data) {
  if (data.Detail.Terms.Taxes1.length > 0) {
    record.TaxCodes = [data.Detail.Terms.Taxes1];
  } else if (data.Detail.Terms.Taxes2.length > 0) {
    record.TaxCodes = [data.Detail.Terms.Taxes2];
  } else {
    record.TaxCodes = [];
  }

  if (
    data.Detail.Terms.Taxes1.length > 0 &&
    data.Detail.Terms.Taxes2.length > 0
  ) {
    record.TaxCodes = [data.Detail.Terms.Taxes1, data.Detail.Terms.Taxes2];
  }
}

function getTaxCodes() {
  const url = "/fico/taxsetup/gets";
  axios.post(url, { Sort: ["_id"] }).then(
    (r) => {
      data.taxes = r.data.data.reduce((res, val) => {
        res[val._id] = val;
        return res;
      }, {});
    },
    (e) => {
      util.showError(e);
    }
  );
}

function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "VendorID":
      record.AddressAndTax.DeliveryName = v2;
      record.AddressAndTax.BillingName = v2;
      record.AddressAndTax.TaxName = v2;
      getVendor(v1, record);
      getVendorBank(v1);
      break;
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
    case "TaxCodes":
      record.Lines.map((e) => {
        e.TaxCodes = record.TaxCodes;
        return e;
      });
      // calcTax(v1);
      break;
    case "PaymentTermID":
      getPaymentTerms(v1);
      break;
    case "TrxDate":
      getPaymentTerms(data.record.PaymentTermID);
      break;
    case "HeaderDiscountType":
      data.record.HeaderDiscountValue = 0;
      calDiscHeader();
      break;
    case "HeaderDiscountValue":
      calDiscHeader();
      break;
  }
}

function newRecordLine(record) {
  record.Taxable = true;
  record.TaxCodes = data.record.TaxCodes;
}

function onControlModeChanged(mode) {
  data.dataListMode = mode;
}

function calcLineTotal(
  total = {
    PriceEach: 0,
    Qty: 0,
    Discount: 0,
    AmountPercent: 0,
    Subtotal: 0,
    Amount: 0,
  }
) {
  // calcTax(data.record.TaxCodes);
  // const amountTax = data.totalTaxIn - data.totalTaxDec;

  data.record.PriceTotalAmount = parseFloat(total.Subtotal);
  data.record.DiscountAmount = parseFloat(
    (total.Discount += total.AmountPercent)
  );
  data.record.SubtotalAmount =
    parseFloat(total.Subtotal) - data.record.DiscountAmount;
  const amountTax = data.record.Lines.reduce((count, val) => {
    val.PPN = genPPNPPH("PPN", val.TaxCodes, val);
    val.PPH = genPPNPPH("PPh", val.TaxCodes, val);
    count += val.PPH + val.PPN;
    return count;
  }, 0);
  const amountPPH = data.record.Lines.reduce((count, val) => {
    val.PPH = genPPNPPH("PPh", val.TaxCodes, val);
    count += val.PPH;
    return count;
  }, 0);
  const amountPPN = data.record.Lines.reduce((count, val) => {
    val.PPN = genPPNPPH("PPN", val.TaxCodes, val);
    count += val.PPN;
    return count;
  }, 0);

  data.record.TaxAmount = amountTax;
  data.record.PPNAmount = amountPPN;
  data.record.PPHAmount = amountPPH;
  data.record.PPHAmountStr = `(${util.formatMoney(
    Math.abs(data.record.PPHAmount)
  )})`;
  data.record.TotalAmount = total.Amount + amountTax;
  data.record.DiscountAmountStr = `(${util.formatMoney(
    data.record.DiscountAmount
  )})`;

  calDiscHeader();
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    if (
      [
        "CashPayment",
        "CashBankID",
        "TaxInvoiceNo",
        "HeaderDiscountType",
        "PaymentTermID",
        "TaxCodes",
        "InvoiceNo",
      ].includes(e.field) === true
    )
      return;
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
  closePreview()
  setModeGrid();

  let GRList = data.record.References.find((o) => o.Key.includes("GR Ref No"));
  if ((action == "Submit" || action == "Post") && GRList) {
    let param = GRList.Value.split(",");
    releaseGR(param);
  }
}
function trxErrorSubmit(e) {
  setLoadingForm(false);
}
function setModeGrid() {
  listControl.value.setControlMode("grid");
  // listControl.value.gridResetFilter();
  listControl.value.refreshList();
}

function openForm(record) {
  data.record = record;
  data.record.DiscountAmountStr = `(${
    record.DiscountAmount ? util.formatMoney(record.DiscountAmount) : 0
  })`;
  data.record.HeaderDiscountAmountStr = `(${
    record.HeaderDiscountAmount
      ? util.formatMoney(record.HeaderDiscountAmount)
      : 0
  })`;
  data.record.PPHAmountStr = `(${
    record.PPHAmount ? util.formatMoney(Math.abs(record.PPHAmount)) : 0
  })`;
  util.nextTickN(2, () => {
    if (record.JournalTypeID) getJurnalType(record.JournalTypeID);
    if (record.VendorID) getVendorBank(record.VendorID);
    const cfgTaxCodes = listControl.value.getFormField("TaxCodes");

    listControl.value.setFormFieldAttr(
      "TaxCodes",
      "lookupPayloadBuilder",
      (search) => {
        return helper.payloadBuilderTaxCodes(
          search,
          cfgTaxCodes,
          record.TaxCodes,
          "Purchase"
        );
      }
    );

    if (readOnly.value == true) {
      listControl.value.setFormMode("view");
    }
  });
}

function getJurnalType(id) {
  if (id === "" || id === null) {
    data.jurnalType = {};
    data.record.PostingProfileID = "";
    data.record.DefaultOffset = {};
    return;
  }
  listControl.value.setFormLoading(true);
  axios
    .post("/fico/vendorjournaltype/get", [id])
    .then(
      (r) => {
        data.jurnalType = r.data;
        data.jurnalType.ChecklistTemplateID = "TEST";
        data.record.PostingProfileID = r.data.PostingProfileID;
        data.record.DefaultOffset = r.data.DefaultOffset;
      },
      (e) => {
        data.jurnalType = {};
        data.record.PostingProfileID = "";
        data.record.DefaultOffset = {};
        util.showError(e);
      }
    )
    .finally(() => {
      listControl.value.setFormLoading(false);
    });
}

function onPostSave(r) {
  if (gridAttachment.value) gridAttachment.value.Save();
}

function getPaymentTerms(id) {
  if (id == "") {
    data.record.ExpectedDate = new Date();
    return;
  }
  const url = "/fico/paymentterm/find?_id=" + id;
  axios.post(url).then(
    (r) => {
      if (r.data.length > 0 && r.data[0].Days > 0) {
        const Days = r.data[0].Days;
        data.record.ExpectedDate = moment(
          new Date(data.record.TrxDate),
          "DD/MM/YYYY"
        ).add("days", Days);
      } else {
        data.record.ExpectedDate = new Date();
      }
    },
    (e) => {
      util.showError(e);
    }
  );
}

function releaseGR(grList) {
  if (grList.length == 0) return;
  const url = "/scm/inventory/receive/update-gr-used";
  let param = {
    GRNumber: grList,
  };
  axios.post(url, param).then(
    (r) => {},
    (e) => {
      util.showError(e);
    }
  );
}

function calDiscHeader() {
  let dr = data.record;
  if (!dr.HeaderDiscountType) {
    dr.HeaderDiscountAmount = 0;
    return;
  }
  const totalAmount = dr.PriceTotalAmount + dr.TaxAmount - dr.DiscountAmount;
  let discountAmount = 0;
  if (dr.HeaderDiscountType == "fixed") {
    discountAmount = dr.HeaderDiscountValue;
  } else {
    discountAmount = (totalAmount * parseFloat(dr.HeaderDiscountValue)) / 100;
  }
  dr.TotalAmount = totalAmount - discountAmount;
  dr.SubtotalAmount = dr.TotalAmount - dr.TaxAmount;
  dr.HeaderDiscountAmount = discountAmount;
  data.record.HeaderDiscountAmountStr = `(${util.formatMoney(discountAmount)})`;
}

function genPPNPPH(flag, taxcodes, dt) {
  if (!taxcodes) return 0;
  let res = [];
  let tax = taxcodes.filter((o) => o.includes(flag));
  for (let i in tax) {
    let ob = data.taxes[tax[i]];
    if (ob) {
      const decr = ob.InvoiceOperation == "Decrease" ? -1 : 1;
      res.push(dt.Amount * ob.Rate * decr);
    }
  }
  return res.length > 0 ? res.reduce((a, b) => a + b, 0) : 0;
}

function onCloseAction(v) {
  listControl.value.submitForm(
    data.record,
    () => {
      data.modalAction = false;
      listControl.value.setFormCurrentTab(2);
      gridAttachment.value.refreshGrid();
    },
    () => {},
    true
  );
}

function onChangeVendorID(name, v1, v2, old) {
  onFormFieldChange(name, v1, v2, old, data.record);
}

onMounted(() => {
  setTimeout(() => {
    getTaxCodes();
  }, 500);
});
</script>
<style>
.btn-cstm {
}
</style>
