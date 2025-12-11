<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="cardTitle"
      v-show="data.appMode == 'grid'"
      grid-config="/fico/customerjournal/gridconfig"
      form-config="/fico/customerjournal/formconfig"
      grid-read="/fico/customerjournal/gets"
      form-read="/fico/customerjournal/get"
      grid-mode="grid"
      grid-delete="/fico/customerjournal/delete"
      form-keep-label
      form-insert="/fico/customerjournal/insert"
      form-update="/fico/customerjournal/update"
      :grid-custom-filter="data.customFilter"
      :form-tabs-new="['General', 'Address', 'Lines', 'References']"
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
        'DefaultOffset',
        'Dimension',
        'PostingProfileID',
        'ExpectedDate',
        'TrxDate',
        'DiscountAmount',
        'HeaderDiscountAmount',
      ]"
      stay-on-form-after-save
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @form-field-change="onFormFieldChange"
      @controlModeChanged="onControlModeChanged"
      @preSave="preSubmit"
      @gridResetCustomFilter="resetGridHeaderFilter"
      :formHideSubmit="readOnly"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-fields="['Status']"
      :form-hide-submit="
        data.record.JournalTypeID == 'CJT-002' && data.validateMiningRent
      "
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
              label="Customer"
              use-list
              lookup-url="/tenant/customer/find"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookupSearchs="['Name']"
              v-model="item.CustomerIDs"
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
                :dim-names="['PC', 'Site']"
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
          :trx-type="item.TransactionType"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :auto-post="!waitTrxSubmit"
        />
        <template v-if="profile.canUpdate">
          <s-button
            :disabled="
              inSubmission ||
              loading ||
              !data.actionList.includes(item.TransactionType)
            "
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Action"
            @click="data.action.modal = true"
          ></s-button>
          <s-button
            :disabled="inSubmission || loading"
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Preview"
            @click="data.appMode = 'preview'"
            v-if="mode !== 'new'"
          ></s-button>
          <s-button
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Sync"
            @click="onSyncLine"
            v-if="tabIndex == 2"
          ></s-button>
        </template>
      </template>
      <template #form_input_DefaultOffset="{ item, mode }">
        <AccountSelector
          v-model="item.DefaultOffset"
          hide-account-type
          :readOnly="readOnly || mode == 'view'"
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
      <template #form_input_PostingProfileID="{ item, config }">
        <label class="input_label">{{ config.label }}</label>
        <div>
          {{ item.PostingProfileID === "" ? "-" : item.PostingProfileID }}
        </div>
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          ref="dimenstionCtl"
          v-model="item.Dimension"
          :readOnly="readOnly || mode == 'view'"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #form_tab_Address="props">
        <Address
          v-model="props.item.AddressAndTax"
          :readOnly="readOnly || props.mode == 'view'"
        ></Address>
      </template>
      <template #form_tab_Lines="{ item, mode }">
        <journal-line
          ref="lineCtl"
          v-model="item.Lines"
          @calc="calcLineTotal"
          :read-only="readOnly || mode == 'view'"
          grid-config-url="/fico/customerjournal/line/gridconfig"
          form-config-url="/fico/customerjournal/line/formconfig"
          @new-record="newRecordLine"
          @alter-grid-config="onAlterGridConfig"
          :attch-kind="data.jType"
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.jType"
          @preOpenAttch="preOpenAttch"
          @closeAttch="closeAttch"
          show-references
          show-checklist
          :referenceTemplate="data.jurnalType?.ReferenceTemplateLineID"
          :checklistId="data.jurnalType?.ChecklistTemplateLineID"
          :key="item.Lines"
          references-hide-manual-input
        >
          <template #grid_Account="p">
            <AccountSelector
              v-model="p.item.Account"
              hide-account-type
              hide-label
              class="min-w-[120px]"
              :items-type="['COA']"
              :read-only="
                item.Status !== 'READY' && (readOnly || mode == 'view')
              "
            ></AccountSelector>
          </template>
        </journal-line>
        <!-- <Lines v-model="props.item.Lines" @recalc="calcLines" /> -->
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
          {{ moment(item.ExpectedDate).format("DD/MM/YYYY") }}
        </div>
      </template>

      <template #form_tab_References="{ item, mode }">
        <References
          :ReferenceTemplate="data.jurnalType.ReferenceTemplateID"
          :readOnly="readOnly || mode == 'view'"
          v-model="item.References"
          v-if="data.jurnalType.ReferenceTemplateID"
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

      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>

      <template #form_tab_Attachment="{ item }">
        <attachment
          v-model="item.Attachment"
          :journalId="item._id"
          :journalType="data.jType"
          :tags="linesTag"
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
      SourceType="CUSTOMER"
      :SourceJournalID="data.record._id"
      :disable-print="helper.isDisablePrintPreview(data.record.Status)"
      :Name="data.record.TransactionType"
      :reload="data.record.JournalTypeID == 'CJT-002' ? 2 : 1"
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
      <template #header_preview="{ preview }">
        <div class="w-full" v-if="preview.Header">
          <div class="grid grid-cols-4 gap-2">
            <div>
              <div
                v-for="(header, header_idx) in preview.Header.Data"
                :key="header_idx"
                class="p-1 grid grid-cols-3"
              >
                <div>{{ preview.Header.Data[header_idx][0] }}</div>
                <div class="col-span-2">
                  {{ preview.Header.Data[header_idx][1] }}
                </div>
              </div>
            </div>
            <div></div>
            <div></div>
            <div>
              <div
                v-for="(header, header_idx) in preview.Header.Data"
                :key="header_idx"
                class="p-1"
              >
                {{ preview.Header.Data[header_idx][5] }}
              </div>
            </div>
          </div>
        </div>
      </template>
      <template #footer_preview="{ preview }">
        <div
          v-for="(footer, footer_idx) in preview.Header.Footer"
          :key="footer_idx"
          class="gap-[1px] grid"
          :class="{
            gridCol1: footer.length == 1,
            gridCol2: footer.length == 2,
            gridCol3: footer.length == 3,
            gridCol4: footer.length == 4,
            gridCol5: footer.length == 5,
            gridCol6: footer.length == 6,
            gridCol7: footer.length == 7,
            gridCol8: footer.length == 8,
            gridCol9: footer.length == 9,
            gridCol10: footer.length == 10,
            gridCol11: footer.length == 11,
            gridCol12: footer.length == 12,
            gridCol13: footer.length == 13,
            gridCol14: footer.length == 14,
            gridCol15: footer.length == 15,
            gridCol16: footer.length == 16,
            gridCol17: footer.length == 17,
          }"
        >
          <div v-for="(v, v_idx) in footer" :key="v_idx">
            <div
              :class="[
                v == '' ? 'mb-4' : '',
                footer_idx == 0 || footer_idx == 3 ? 'font-semibold' : '',
              ]"
            >
              {{ v }}
            </div>
          </div>
        </div>
      </template>
    </PreviewReport>

    <s-modal
      :display="data.action.modal"
      hideButtons
      :title="'Action - ' + data.record.TransactionType"
      @beforeHide="data.action.modal = false"
    >
      <action-customer
        :data-item="data.record"
        @lines="onGenerateLines"
        @submit="onSubmitAction"
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

import References from "@/components/common/References.vue";

import Address from "./widget/customerjournal/Address.vue";
// import Lines from "./widget/customerjournal/Lines.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import Checklist from "@/components/common/Checklist.vue";

import StatusText from "@/components/common/StatusText.vue";

import JournalLine from "@/components/common/JournalLine.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import ActionCustomer from "./widget/customerjournal/ActionCustomer.vue";
import ActionAttachment from "@/components/common/ActionAttachment.vue";

import moment from "moment";

layoutStore().name = "tenant";

const FEATUREID = "CustomerJournal";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const lineCtl = ref(null);
const dimenstionCtl = ref(null);
const gridAttachment = ref(null);

const router = useRouter();
const route = useRoute();
const axios = inject("axios");
const gridHeaderFilter = ref(null);
const data = reactive({
  dataListMode: "grid",
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  record: {},
  preview: {},
  jurnalType: {},
  action: {
    modal: false,
  },
  taxes: [],
  totalTaxIn: 0,
  totalTaxDec: 0,
  customFilter: null,
  jType: "CUSTOMER",
  originTotalAmount: 0,
  actionList: [
    "Mining Invoice - Rent",
    "General Invoice",
    "General Invoice - Tourism",
    "General Invoice - Sparepart",
    "General Invoice - Unit Sales",
  ],
  validateMiningRent: true,
});

const cardTitle = computed(() => {
  if (data.dataListMode == "grid") return "Customer Transaction";
  const formMode = listControl.value.getFormMode();

  return formMode == "new"
    ? `Customer Transaction - ${formMode}`
    : "Customer Transaction - " +
        [
          data.record._id,
          data.record.AddressAndTax?.DeliveryName,
          util.formatMoney(data.record.TotalAmount),
        ].join(" | ");
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
    return ["DRAFT", "READY"].includes(data.record.Status);
  },
});
function preOpenAttch() {
  axios.post("/fico/customerjournal/update", data.record);
}
function closeAttch() {
  gridAttachment.value.refreshGrid();
}

function onAlterGridConfig(config) {
  // Dibawah ini kalau mau merubah grid config
  config.fields.forEach((element) => {
    switch (element.field) {
      case "Dimension":
        element.readType = "hide";
        break;

      case "UnitID":
        element.readType = "hide";
        break;

      case "DiscountType":
        //element.readType = "hide";
        break;
    }
  });
}
function alterFormConfig(config) {
  if (route.query.id !== undefined) {
    let getUrlParam = route.query.id;
    listControl.value.selectData({ _id: getUrlParam }); //remark sementara tunggu suimjs update
    router.replace({ path: "/fico/CustomerTransaction" });
  } else if (route.query.form == 1) {
    listControl.value.newData();
    router.replace({ path: "/fico/CustomerTransaction" });
  }
}
function initNewItemFilter(item) {
  item.CustomerIDs = [];
  item.Dimension = [];
}

function changeFilter(item, filters) {
  if (item.CustomerIDs.length > 0) {
    filters.push({
      Op: "$in",
      Field: "CustomerID",
      Value: item.CustomerIDs,
    });
  }
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
  data.record.TotalAmount = parseInt(total.Amount);
  data.originTotalAmount = parseInt(total.Amount);
  data.record.SubtotalAmount = parseInt(total.Subtotal);
  data.record.DiscountAmount = parseInt(
    (total.Discount += total.AmountPercent)
  );
  data.record.TaxAmount = 0;
  calcTaxAmount();
}
function calcTaxAmount(taxCodes) {
  data.totalTaxIn = 0;
  data.totalTaxDec = 0;
  const _taxCodes = taxCodes ?? data.record.TaxCodes;
  if (_taxCodes?.length > 0) {
    const filtered = data.taxes.filter((item) =>
      data.record.TaxCodes.includes(item._id)
    );
    filtered.map((v) => {
      if (v.InvoiceOperation == "Increase") {
        data.totalTaxIn = parseInt(
          (data.totalTaxIn += data.record.SubtotalAmount * v.Rate)
        );
      } else {
        data.totalTaxDec = parseInt(
          (data.totalTaxDec += data.record.SubtotalAmount * v.Rate)
        );
      }
    });
  } else {
    data.totalTaxIn = 0;
    data.totalTaxDec = 0;
  }

  const amountTax = data.totalTaxIn - data.totalTaxDec;

  data.record.TotalAmount = data.originTotalAmount + amountTax;
  data.record.TaxAmount = parseInt(amountTax);
  calDiscHeader();
}

function preSubmit(record) {
  record.Lines?.map((e) => {
    e.OffsetAccount = {
      AccountType: "CUSTOMER",
      AccountID: record.CustomerID,
    };
    return e;
  });
  data.totalAmount = record.TotalAmount;
}

function closePreview() {
  data.appMode = "grid";
}

function calcLines() {
  const record = data.record;
  record.SubtotalAmount = record.Lines.reduce((a, b) => {
    return a + b.Amount;
  }, 0);
  record.TotalAmount = record.SubtotalAmount;
}
function editRecord(record) {
  data.validateMiningRent = false;
  openForm(record);
}

function newRecord(record) {
  record._id = "";
  // // TODO: baca dari store, simpan as default dari masing2 tenant/company
  record.CurrencyID = "IDR";
  record.TrxDate = new Date();
  record.ExpectedDate = new Date();
  record.DeliveryDate = new Date();
  record.Actions = [];
  record.Previews = [];
  record.AddressAndTax = {};
  record.DefaultOffset = {};
  record.DiscountAmountStr = "(0)";
  record.HeaderDiscountAmountStr = "(0)";
  record.Status = "DRAFT";
  record.PostingProfileID = "";
  record.TaxInvoiceDate = new Date();
  record.References = [];
  data.validateMiningRent = true;
  util.nextTickN(2, () => {
    openForm(record);
  });
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
    getJurnalType(record.JournalTypeID);
    const cfgTaxCodes = listControl.value.getFormField("TaxCodes");

    listControl.value.setFormFieldAttr(
      "TaxCodes",
      "lookupPayloadBuilder",
      (search) => {
        return helper.payloadBuilderTaxCodes(
          search,
          cfgTaxCodes,
          record.TaxCodes,
          "Sales"
        );
      }
    );
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
  });
}

function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "CustomerID":
      record.AddressAndTax.DeliveryName = v2;
      record.AddressAndTax.BillingName = v2;
      record.AddressAndTax.TaxName = v2;
      getCustomer(v1, record);
      break;
    case "JournalTypeID":
      getJurnalType(v1);
      break;
    case "TaxCodes":
      record.Lines.map((e) => {
        e.TaxCodes = record.TaxCodes;
        return e;
      });

      calcTaxAmount();
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

function getJurnalType(id, record) {
  if (!id) {
    data.jurnalType = {};
    data.record.PostingProfileID = "";
    data.record.DefaultOffset = {};
    data.record.References = [];
    return;
  }
  listControl.value.setFormLoading(true);
  axios
    .post("/fico/customerjournaltype/get", [id])
    .then(
      (r) => {
        data.jurnalType = r.data;
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

function getCustomer(id, record) {
  if (id.length > 0) {
    const url = "/bagong/customer/get?_id=" + id;
    axios.post(url, [id]).then(
      (r) => {
        record.AddressAndTax.DeliveryAddress = `${r.data.Detail.DeliveryAddress}\n${r.data.Detail.DeliveryCity}\n${r.data.Detail.DeliveryProvince}\n${r.data.Detail.DeliveryCountry}\n${r.data.Detail.DeliveryZipcode}`;
        record.AddressAndTax.BillingAddress = `${r.data.Detail.Address} \n${r.data.Detail.City}\n${r.data.Detail.Province}\n${r.data.Detail.Country}\n${r.data.Detail.Zipcode}`;
        record.AddressAndTax.TaxAddress = r.data.Detail.TaxAddress;
        record.AddressAndTax.TaxType = r.data.Detail.TaxType;
        record.PaymentTermID = r.data.Detail.Termin;
        getPaymentTerms(record.PaymentTermID);
        conditionGetTaxCodes(record, r.data);
      },
      (e) => util.showError(e)
    );
  }
}

function conditionGetTaxCodes(record, data) {
  if (data.Detail.Tax1.length > 0) {
    record.TaxCodes = [data.Detail.Tax1];
  } else if (data.Detail.Tax2.length > 0) {
    record.TaxCodes = [data.Detail.Tax2];
  } else {
    record.TaxCodes = [];
  }

  if (data.Detail.Tax1.length > 0 && data.Detail.Tax2.length > 0) {
    record.TaxCodes = [data.Detail.Tax1, data.Detail.Tax2];
  }
}

function getTaxCodes() {
  const url = "/fico/taxsetup/gets";
  axios.post(url, { Sort: ["_id"] }).then(
    (r) => {
      data.taxes = r.data.data;
    },
    (e) => {
      util.showError(e);
    }
  );
}

function calcTax(taxCodes) {
  data.totalTaxIn = 0;
  data.totalTaxDec = 0;
  if (taxCodes?.length > 0) {
    const filtered = data.taxes.filter((item) => taxCodes.includes(item._id));
    filtered.map((v) => {
      if (v.InvoiceOperation == "Increase") {
        data.totalTaxIn = parseInt(
          (data.totalTaxIn += data.record.SubtotalAmount * v.Rate)
        );
      } else {
        data.totalTaxDec = parseInt(
          (data.totalTaxDec += data.record.SubtotalAmount * v.Rate)
        );
      }
    });
  } else {
    data.totalTaxIn = 0;
    data.totalTaxDec = 0;
  }
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
        "ExpectedDate",
        "DeliveryDate",
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

function calDiscHeader() {
  let dr = data.record;
  if (!dr.HeaderDiscountType) {
    dr.HeaderDiscountAmount = 0;
    return;
  }
  const totalAmount = dr.SubtotalAmount + dr.TaxAmount - dr.DiscountAmount;
  let discountAmount = 0;
  if (dr.HeaderDiscountType == "fixed") {
    discountAmount = dr.HeaderDiscountValue;
  } else {
    discountAmount = (totalAmount * parseFloat(dr.HeaderDiscountValue)) / 100;
  }
  dr.TotalAmount = totalAmount - discountAmount;
  dr.HeaderDiscountAmount = discountAmount;
}

function onGenerateLines(lines, ref) {
  data.action.modal = false;
  data.record.Lines = lines;
  listControl.value.setFormCurrentTab(2);
  data.validateMiningRent = false;

  let So = ref.find((x) => x.Key == "SO No.");
  for (let i in data.record.References) {
    let o = data.record.References[i];
    if (o.Key == "SO No.") o.Value = So.Value;
  }
}

function onSubmitAction(action, projectList) {
  let fProject = projectList.find((x) => x.key == action["ProjectID"]);
  let ref = data.record.References;
  if (ref.length == 0) return;

  let forAction = [
    { Field: "ProjectID", Name: "Action - Project" },
    { Field: "Start", Name: "Action - Bulan Timesheet Start" },
    { Field: "End", Name: "Action - Bulan Timesheet End" },
  ];

  for (let i in forAction) {
    let o = forAction[i];
    let f = ref.find((x) => x.Key == o.Name);
    if (f) {
      f.Value =
        o.Field == "ProjectID"
          ? fProject
            ? fProject.text
            : ""
          : moment(action[o.Field]).utc().format("YYYY-MM-DD");
    }
  }
}

function onSyncLine() {
  const url = "/bagong/invoice/sync-mining-invoice";

  axios.post(url, data.record).then(
    (r) => {},
    (e) => {
      util.showError(e);
    }
  );
}

const tabIndex = computed({
  get() {
    let idx = listControl.value.getFormCurrentTab();
    return idx ?? 0;
  },
});

function formatHeader(params) {}

onMounted(() => {
  setTimeout(() => {
    getTaxCodes();
  }, 500);
});
</script>
