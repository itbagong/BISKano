<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.title"
      grid-config="/fico/submissionemployee/journal/gridconfig"
      form-config="/fico/submissionemployee/journal/formconfig"
      grid-read="/fico/employeeexpense/gets"
      form-read="/fico/vendorjournal/get"
      grid-mode="grid"
      grid-delete="/fico/vendorjournal/delete"
      form-keep-label
      form-insert="/fico/vendorjournal/save"
      form-update="/fico/vendorjournal/save"
      :grid-fields="['Enable', 'Status']"
      :form-tabs-new="data.tabsNew"
      :form-tabs-edit="data.tabsEdit"
      :form-tabs-view="data.tabsView"
      :form-fields="[
        '_id',
        'Lines',
        'Dimension',
        'JournalTypeID',
        'VendorID',
        'Status',
      ]"
      :grid-custom-filter="data.customFilter"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      grid-sort-field="TrxDate"
      grid-sort-direction="desc"
      @formNewData="newRecord"
      @formEditData="editRecord"
      :formHideSubmit="readOnly"
      grid-hide-select
      stay-on-form-after-save
      @preSave="preSubmit"
      @post-save="onPostSave"
      @gridResetCustomFilter="resetGridHeaderFilter"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @alterFormConfig="onAlterFormConfig"
      @form-field-change="onFormFieldChange"
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
              :dim-names="['Site']"
            ></dimension-editor>
          </div>
        </template> 
      </grid-header-filter>
      </template>
      <template #form_buttons_1="{ item, inSubmission, loading }">
        <form-buttons-trx
          :disabled="inSubmission || loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="data.jType"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :auto-post="!waitTrxSubmit"
        />
      </template>
      <template #form_input__id="{ item, config }">
        <label class="flex input_label">
          {{ config.label }}
        </label>
        <a
          href="#"
          class="text-blue-400 hover:text-blue-800"
          @click="redirect(item._id)"
          >{{ item._id }}
        </a>
      </template>
      <!-- <template #form_input_JournalTypeID_selected-option="{ option }">
        {{ option.item?.Name }}
      </template>
      <template #form_input_JournalTypeID_option="{ option }">
        {{ option.item?.Name }}
      </template> -->
      <!-- <template #form_input_VendorID_selected-option="{ option }">
        {{ option.item?.Name }}
      </template>
      <template #form_input_VendorID_option="{ option }">
        {{ option.item?.Name }}
      </template> -->
      <template #form_input_JournalTypeID="{ item, config }">
        <s-input
          :key="data.keyJournalTypeID"
          keep-label
          :label="config.label"
          v-model="item.JournalTypeID"
          use-list
          lookup-url="/fico/vendorjournaltype/find"
          read-only
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookupSearchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_input_VendorID="{ item, config }">
        <s-input
          :key="data.keyVendorID"
          keep-label
          :label="config.label"
          v-model="item.VendorID"
          use-list
          lookup-url="/tenant/vendor/find"
          read-only
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookupSearchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_tab_Line="{ item, mode}">
        <journal-line
          v-model="item.Lines"
          @calc="calcLineTotal"
          :read-only="readOnly|| mode == 'view'"
          grid-config-url="/fico/submissionemployee/line/gridconfig"
          @new-record="newRecordLine"
          :attch-kind="data.jType" 
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.jType" 
          @preOpenAttch="preOpenAttchLine"
          @closeAttch="closeAttchLine"
          @postDelete="postDeleteLine"
          show-references
          :referenceTemplate="data.record.ReferenceTemplateID"
        >
          <template #grid_Account="p">
            <AccountSelector
              v-model="p.item.Account"
              :items-type="['EXP']"
              :group-id-value="data.groupIdValue"  
              :read-only="
                item.Status !== 'READY' && (readOnly || mode == 'view')
              "
            ></AccountSelector>
          </template>
          <template #grid_TagObjectID1="p">
            <AccountSelector
              v-model="p.item.TagObjectID1"
              :items-type="['EMP']"
              :read-only="
                item.Status !== 'READY' && (readOnly || mode == 'view')
              "
            ></AccountSelector>
          </template>
        </journal-line>
      </template>
      <template #form_input_Dimension="{ item,mode}">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :read-only="readOnly || mode == 'view'"
          :default-list="profile.Dimension "
        ></dimension-editor-vertical>
      </template>
      <template #form_input_Status="{ item, config }">
        <div>
          <label class="input_label">{{ config.label }}</label>
        </div>
        <status-text :txt="item.Status" />
      </template>
      <template #form_tab_References="{ item,mode }">
        <ref-template
          :ReferenceTemplate="data.record.ReferenceTemplateID"
          :readOnly="readOnly || mode == 'view'"
          v-model="item.References "
          @get-items="getRefItems"
        >
          <template #ref_Employee="{ item, value }">
            <s-input
              ref="inputs"
              v-model="value.Value"
              :label="item.Label"
              class="w-full"
              use-list
              :lookup-url="
                '/bagong/employee/get-employee-by-position?SiteID=' +
                data.selectedSite.Value
              "
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
            ></s-input>
          </template>
        </ref-template>
      </template>
      <template #form_tab_Attachment="{ item }">
        <attachment
          ref="gridAttachment"
          v-model="item.Attachment"
          :journalId="item._id"
          :journalType="data.jType" 
          :tags="linesTag"
          single-save
        />
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #grid_item_button_delete="{item}">  
        <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, watch, computed, onMounted } from "vue";
import { authStore } from "@/stores/auth";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput } from "suimjs";
import { useRouter, useRoute } from "vue-router";
import helper from "@/scripts/helper.js";

import DimensionEditor from "@/components/common/DimensionEditor.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import JournalLine from "@/components/common/JournalLine.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";
import refTemplate from "./widget/RefTemplate.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import Attachment from "@/components/common/SGridAttachment.vue";

layoutStore().name = "tenant";

const FEATUREID = "EmployeeExpenseSubmission";
const profile = authStore().getRBAC(FEATUREID);

const auth = authStore();

const listControl = ref(null);
const gridHeaderFilter = ref(null);
const gridAttachment = ref(null);
const axios = inject("axios");
const route = useRoute();
const router = useRouter();

const data = reactive({
  title: "Employee Expense Submission",
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  record: {},
  tabsNew: ["General", "Line"],
  tabsEdit: ["General", "Line", "References", "Attachment"],
  tabsView: ["General", "Line", "References", "Attachment"],
  customFilter: null,
  selectedSite: "",
  groupIdValue: ["EXG0006"],
  jType: "VENDOR",
  taxes: [],
  keyJournalTypeID: 'JournalTypeID',
  keyVendorID: 'VendorID'
});

const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
const linesTag = computed({
  get(){
    return data.record.Lines?.map(e=>{
      return `${data.jType}_${data.record._id}_${e.LineNo}`
    })
  }
}) 
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "READY"].includes(data.record.Status);
  },
});
function preOpenAttchLine(readOnly){ 
  if(readOnly) return
  preSubmit(data.record)
  axios.post("/fico/vendorjournal/update", data.record)
}
function closeAttchLine(readOnly){ 
  if(readOnly) return 
  gridAttachment.value.refreshGrid()
}
function postDeleteLine(){
  gridAttachment.value.refreshGrid() 
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
function getLedgerJournalPost(siteID) {
  if (siteID == "" || siteID == null || siteID == undefined) return;
  const url = "/bagong/siteentry/get-detail-ledger-journal-post";
  const param = {
    SiteID: siteID,
    Type: "Employee Expense",
  };

  axios.post(url, param).then(
    (r) => {
      data.record.JournalTypeID = r.data._id;
      data.record.DefaultOffset = r.data.DefaultOffset;
      data.record.PostingProfileID = r.data.PostingProfileID;
      data.record.ReferenceTemplateID = r.data.ReferenceTemplateID;
      data.record.ChecklistTemplatelD = r.data.ChecklistTemplatelD;
      util.nextTickN(2, () => {
        getReferenceTemplate(r.data.ReferenceTemplateID);
      });
    },
    (e) => {}
  );
}

function getVendorID(siteID) {
  if (siteID == "" || siteID == null || siteID == undefined) return;
  const url = "/fico/employeeexpense/get-vendor";
  const param = {
    Site: siteID,
  };

  axios.post(url, param).then(
    (r) => {
      data.record.VendorID = r.data._id;
    },
    (e) => {}
  );
}

function getReferenceTemplate(id) {
  if (id == "") return;
  const url = "/tenant/referencetemplate/get";
  axios.post(url, [id]).then(
    (r) => {
      const references = r.data.Items.map((v, i) => {
        const confValue = v.ConfigValue.split("|");
        return {
          Key: v.Label,
          Value:
            v.ReferenceType == "items"
              ? confValue[0]
              : data.record.References.length > 0
              ? data.record.References[i].Value
              : "",
        };
      });
      data.record.References = references;
    },
    (e) => {}
  );
}


function openForm(record) {
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
    if (readOnly.value === true) { 
      listControl.value.setFormMode("view");
    }
  });
}

async function newRecord(record) { 
  record._id = "";
  record.Status = "DRAFT";
  record.CompanyID = auth.companyId;
  record.CurrencyID = "IDR";
  record.TrxDate = new Date();
  record.Checklists = [];
  record.References = [];
  record.Lines = [];
  record.TransactionType = "Employee Expense";
  data.record = record;

  openForm(record);
}

function editRecord(record) { 
  data.record = record;
  openForm(record);
}
function newRecordLine(record) {
  record.Taxable = true;
  record.TaxCodes = data.record.TaxCodes;
}
function preSubmit(record) {
  record.CurrencyID = "IDR";
  record.Lines?.map((e) => {
    e.Dimension = record.Dimension;
    e.OffsetAccount = record.DefaultOffset;
    return e;
  });
}
function onPostSave(r) {
  if (gridAttachment.value) gridAttachment.value.Save();
  listControl.value.refreshForm();
  data.keyJournalTypeID = `journalypeid_${r._id}`
  data.keyVendorID = r._id
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    if (
      [
        "TaxCodes",
        "HeaderDiscountAmount",
        "HeaderDiscountType",
        "HeaderDiscountValue",
      ].includes(e.field) === true
    )
      return;
    listControl.value.setFormFieldAttr(e.field, "required", required);
  });
}

function trxPreSubmit(status, action, doSubmit) {
  if (waitTrxSubmit.value) {
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

function calcLineTotal(
  total = {
    PriceEach: 0,
    Qty: 0,
    Discount: 0,
    AmountPercent: 0,
    Subtotal: 0,
    Amount: 0,
  }) {
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
    case "TaxCodes":
      record.Lines.map((e) => {
        e.TaxCodes = record.TaxCodes;
        return e;
      });
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

 

function getEmployeeExpTypeGroup(pcValue) {
  if (pcValue && pcValue != "") {
    let url = `/tenant/expensetypegroup/gets?Dimension.Key=PC&Dimension.Value=${pcValue}`;
    axios.post(url, { Sort: ["-_id"] }).then((r) => {
      const records = r.data.data;
      const find = records.find((v) => v.Name.includes("EMPLOYEE"));
      if (find) {
        data.groupIdValue = [data.groupIdValue[0], find._id];
      }
    });
  } else {
    data.groupIdValue = [data.groupIdValue[0]];
  }
}
function redirect(JournalID) {
  const url = router.resolve({
    name: "fico-VendorTransaction",
    query: { id: JournalID },
  });
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
  () => data.record.Dimension,
  (nv) => {
    const pc = nv?.find((e) => e.Key == "PC");
    if (pc != undefined) {
      getEmployeeExpTypeGroup(pc.Value);
    }

    const site = nv?.find((e) => e.Key == "Site");
    if (site != undefined) {
      getLedgerJournalPost(site.Value);
      getVendorID(site.Value);
      data.selectedSite = site;
    }
  },
  { deep: true }
);
onMounted(() => {
  setTimeout(() => {
    getTaxCodes();
  }, 500);
});
</script>
