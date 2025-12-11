<template>
  <div class="w-full">
    <data-list
      ref="listControl"
      class="card"
      title="General Submission"
      v-show="data.appMode == 'grid'"
      grid-config="/fico/vendorjournal/gridconfig"
      form-config="/fico/vendorjournal/formconfig"
      grid-read="/fico/vendorjournal/gets?TransactionType=General%20Submission"
      form-read="/fico/vendorjournal/get"
      form-insert="/fico/vendorjournal/insert"
      form-update="/fico/vendorjournal/update"
      grid-delete="/fico/vendorjournal/delete"
      grid-mode="grid"
      :grid-fields="['TotalAmount', 'Enable', 'Status']"
      :grid-custom-filter="data.customFilter"
      :form-tabs-edit="[
        'General',
        'Address',
        'Lines',
        'References',
        'Checklist',
        'Documents',
        'Logs',
      ]"
      :form-tabs-view="[
        'General',
        'Address',
        'Lines',
        'References',
        'Checklist',
        'Documents',
        'Logs',
      ]"
      :form-fields="[
        'VendorID',
        'DefaultOffset',
        'Dimension',
        'PostingProfileID',
        'TransactionType',
        'DiscountAmount',
        'HeaderDiscountAmount',
      ]"
      grid-sort-field="TrxDate"
      grid-sort-direction="desc"
      stay-on-form-after-save
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      @form-field-change="onFormFieldChange"
      :formHideSubmit="readOnly"
      @preSave="preSubmit"
      @post-save="onPostSave"
      @controlModeChanged="onControlModeChanged"
      form-keep-label
      gridHideSearch
      @gridResetCustomFilter="resetHeaderFilter"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @alterGridConfig="alterGridConfig"
      @alterFormConfig="onAlterFormConfig"
      :grid-page-size="10"
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
            <div class="min-w-[260px]">
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
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :trx-type="item.TransactionType"
          :auto-post="!waitTrxSubmit"
        />
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
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          ref="dimenstionCtl"
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :read-only="readOnly || mode == 'view'"
        ></dimension-editor-vertical>
      </template>
      <template #form_tab_Address="props">
        <Address
          v-model="props.item.AddressAndTax"
          :read-only="readOnly || props.mode == 'view'"
        ></Address>
      </template>
      <template #form_input_PostingProfileID="{ item, config }">
        <label class="input_label">{{ config.label }}</label>
        <div>
          {{ item.PostingProfileID === "" ? "-" : item.PostingProfileID }}
        </div>
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
          <template #grid_TagObjectID1="p">
            <AccountSelector
              v-model="p.item.TagObjectID1"
              :items-type="['AST', 'COA', 'EXP']"
              :read-only="
                item.Status !== 'READY' && (readOnly || mode == 'view')
              "
            ></AccountSelector>
          </template>
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
      <template #form_tab_References="props">
        <References
          :ReferenceTemplate="data.jurnalType.ReferenceTemplateID"
          :readOnly="readOnly || props.mode == 'view'"
          v-model="props.item.References"
        />
      </template>
      <template #form_tab_Checklist="{ item, mode }">
        <Checklist
          v-model="item.Checklists"
          :checklist-id="data.jurnalType.ChecklistTemplateID"
          :read-only="readOnly || mode == 'view'"
        />
      </template>

      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #grid_item_button_delete="{ item }">
        <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
      </template>
      <template #form_tab_Documents="{ item }">
        <s-grid-attachment
          ref="gridAttachment"
          v-model="item.Attachment"
          :journalId="item._id"
          :journalType="data.jType"
          :tags="linesTag"
          single-save
        />
      </template>
      <template #grid_TotalAmount="{ item }">
        <div>{{ util.formatMoney(item.TotalAmount) }}</div>
      </template>
      <template #form_input_TransactionType="{ item, config }">
        <s-input
          keep-label
          :label="config.label"
          kind="text"
          v-model="item.TransactionType"
          read-only
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, computed, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SGrid, SInput, SButton, SModal, util } from "suimjs";
import { useRoute, useRouter } from "vue-router";
import { authStore } from "@/stores/auth";
import helper from "@/scripts/helper.js";

import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";

import Address from "./widget/VendorJournal/Address.vue";
import Lines from "./widget/VendorJournal/Lines.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import References from "@/components/common/References.vue";
import Checklist from "@/components/common/Checklist.vue";

import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import JournalLine from "@/components/common/JournalLine.vue";

import StatusText from "@/components/common/StatusText.vue";

import LogTrx from "@/components/common/LogTrx.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";

layoutStore().name = "tenant";

const FEATUREID = "GeneralSubmission";
const profile = authStore().getRBAC(FEATUREID);

const route = useRoute();
const router = useRouter();

const axios = inject("axios");

const listControl = ref(null);
const dimenstionCtl = ref(null);
const gridHeaderFilter = ref(null);
const gridAttachment = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  record: null,
  preview: {},
  jurnalType: {},
  customFilter: null,
  jType: "VENDOR",
  taxes: [],
  totalTaxIn: 0,
  totalTaxDec: 0,
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

const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record?.Status);
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
    return ["DRAFT", "READY"].includes(data.record.Status);
  },
});
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
function refreshGrid() {
  listControl.value.refreshGrid();
}
function resetHeaderFilter() {
  gridHeaderFilter.value.reset();
}

function closePreview() {
  data.appMode = "grid";
}

function preSubmit(record) {
  record.Lines?.map((e) => {
    e.OffsetAccount = {
      AccountType: "VENDOR",
      AccountID: record.VendorID,
    };
    e.Dimension = record.Dimension;
    return e;
  });
}
function onPostSave(r) {
  if (gridAttachment.value) gridAttachment.value.Save();
  listControl.value.refreshForm();
}
function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Enable = true;
  record.CurrencyID = "IDR";
  record.PaymentTermID = "P30";
  record.Actions = [];
  record.Previews = [];
  record.AddressAndTax = {};
  record.TrxDate = new Date();
  record.ExpectedDate = null;
  record.DeliveryDate = null;
  record.TransactionType = "General Submission";
  record.Status = "DRAFT";
  record.DiscountAmountStr = "(0)";
  record.HeaderDiscountAmountStr = "(0)";
  openForm(record);
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

  data.preview = {};
  util.nextTickN(2, () => {
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
    if (record.JournalTypeID) getJurnalType(record.JournalTypeID);
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
  });
}

function getJurnalType(id) {
  if (id === "" || id === null) {
    data.jurnalType = {};
    data.record.PostingProfileID = "";
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

// function calcLineTotal(total = { PriceEach: 0, Qty: 0, Amount: 0 }) {
//   data.record.Amount = parseInt(total.Amount);
//   data.record.ReportingAmount = parseInt(total.Amount);
//   data.record.SubtotalAmount = parseInt(total.Amount);
//   data.record.TotalAmount = parseInt(total.Amount);
// }
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
function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "VendorID":
      record.AddressAndTax.DeliveryName = v2;
      record.AddressAndTax.BillingName = v2;
      record.AddressAndTax.TaxName = v2;
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
  }
}
function newRecordLine(record) {
  record.Taxable = true;
  record.TaxCodes = data.record.TaxCodes;
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
function trxPostSubmit(record) {
  setLoadingForm(false);
  setModeGrid();
}
function trxErrorSubmit(e) {
  setLoadingForm(false);
}
function setModeGrid() {
  listControl.value.setControlMode("grid");
  listControl.value.gridResetFilter();
  listControl.value.refreshList();
}
function onControlModeChanged(mode) {
  if (mode == "grid") data.record = null;
}

function onChangeVendorID(field, v1) {
  if (v1) {
    const url = "/bagong/vendor/get";
    axios
      .post(url, [v1])
      .then(
        (r) => {
          console.log(r.data);
          setFromVendorID(r.data);
        },
        (e) => {
          util.showError(e);
        }
      )
      .finally(() => {
        listControl.value.setFormLoading(false);
      });
  }
}

function setFromVendorID(dt) {
  let tc = [];
  if (dt.Detail.Terms.Taxes1) tc.push(dt.Detail.Terms.Taxes1);
  if (dt.Detail.Terms.Taxes2) tc.push(dt.Detail.Terms.Taxes2);
  // data.record.AddressAndTax.DeliveryName = v2;
  // data.record.AddressAndTax.BillingName = v2;
  // data.record.AddressAndTax.TaxName = v2;
  data.record.TaxCodes = tc;
}

function onAlterFormConfig(cfg) {
  if (route.query.id !== undefined) {
    let currQuery = {...route.query};
    listControl.value.selectData({ _id: currQuery.id });
    delete currQuery['id'];
    router.replace({ path: route.path, query: currQuery });
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
onMounted(() => {
  setTimeout(() => {
    getTaxCodes();
  }, 500);
});
</script>
<style></style>
