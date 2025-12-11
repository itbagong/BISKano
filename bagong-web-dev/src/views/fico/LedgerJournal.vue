<template>
  <div class="w-full">
    <data-list
      class="card h-[650px]"
      ref="listControl"
      :title="cardTitle"
      v-show="data.appMode == 'grid'"
      grid-config="/fico/ledgerjournal/gridconfig"
      form-config="/fico/ledgerjournal/formconfig"
      grid-read="/fico/ledgerjournal/gets"
      form-read="/fico/ledgerjournal/get"
      grid-mode="grid"
      grid-delete="/fico/ledgerjournal/delete"
      form-keep-label
      form-insert="/fico/ledgerjournal/insert"
      form-update="/fico/ledgerjournal/update"
      :grid-fields="['Enable', 'Status']"
      :grid-custom-filter="data.customFilter"
      :form-tabs-edit="[
        'General',
        'Lines',
        'References',
        'Checklist',
        'Attachment',
      ]"
      :form-tabs-view="[
        'General',
        'Lines',
        'References',
        'Checklist',
        'Attachment',
      ]"
      :form-fields="['DefaultOffset', 'Dimension', 'Status']"
      stay-on-form-after-save
      grid-sort-field="TrxDate"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @form-new-data="newRecord"
      @form-edit-data="openForm"
      @form-field-change="onFormFieldChange"
      :form-hide-submit="readOnly"
      @control-mode-changed="onControlModeChanged"
      @alterFormConfig="onAlterFormConfig"
      @grid-reset-custom-filter="resetGridHeaderFilter"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete" 
    >
      <template #form_buttons_1="{ item, inSubmission, loading,mode}"> 
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
        <template v-if="profile.canUpdate && mode !== 'new' ">
          <s-button
            :disabled="inSubmission || loading"
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
        </template>
      </template>
      <template #grid_header_search>
        <grid-header-filter 
          ref="gridHeaderFilter"
          v-model="data.customFilter"
          @initNewItem="initNewItemFilter"
          @preChange="changeFilter"
          @change="refreshGrid"
        >
          <template #filter_1="{ item }">
            <s-input
              class="w-[200px]"
              label="Journal Type"
              use-list
              lookup-url="/fico/ledgerjournaltype/find"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookupSearchs="['Name']"
              v-model="item.JournalTypeIDs"
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
                :dim-names="['CC', 'Site']"
                
              ></dimension-editor>
            </div>
          </template>
        </grid-header-filter>
      </template>
      <template #form_input_DefaultOffset="{ item,mode }">
        <AccountSelector
          v-model="item.DefaultOffset"
          hide-account-type
          label-account="Offset Account"
          :disabled="readOnly || mode == 'view'"
        />
      </template>
      <template #form_input_Dimension="{ item, mode}">
        <dimension-editor-vertical
          ref="dimenstionCtl"
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :read-only="readOnly || mode == 'view'"
        ></dimension-editor-vertical>
      </template>
      <template #form_tab_References="props">
        <References
          :ReferenceTemplate="data.jurnalType.ReferenceTemplateID"
          :readOnly="readOnly || props.mode == 'view'"
          v-model="props.item.References "
        />
      </template>
      <template #form_tab_Checklist="props">
        <Checklist
          v-model="props.item.Checklists"
          :checklist-id="data.jurnalType.ChecklistTemplateID"
          :readOnly="readOnly || props.mode == 'view'"
          :attch-kind="data.jType" 
          :attch-ref-id="props.item._id"
          :attch-tag-prefix="data.jType" 
        />
      </template>
      <template #form_tab_Lines="{ item,mode }"> 
        <Lines v-model="item.Lines" @recalc="calcLines" :readOnly="readOnly || mode == 'view'" :status="item.Status" 
          :attch-kind="data.jType" 
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.jType" 
          @preOpenAttch="preOpenAttchLine"
          @closeAttch="closeAttchLine"
          @postDelete="postDeleteLine"
        />
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_input_Status="{ item }">
        <status-text :txt="item.Status" />
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
      <template #grid_item_button_delete="{item}">
        <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
      </template>
    </data-list>

    <PreviewReport
      v-if="data.appMode == 'preview'"
      class="card w-full"
      title="Preview"
      :preview="data.preview"
      :disable-print="helper.isDisablePrintPreview(data.record.Status)"
      @close="closePreview"
      SourceType="LEDGERACCOUNT"
      :SourceJournalID="data.record._id"
    >
      <template #buttons="props">
        <div class="flex gap-[1px] mr-2">
          <s-button class="btn_primary" label="Approve"></s-button>
          <s-button class="btn_warning" label="Reject"></s-button>
          <s-button class="btn_primary" label="Post"></s-button>
        </div>
      </template>
    </PreviewReport>
    <s-modal
      :display="data.action.modal"
      hideButtons
      title="Action"
      @beforeHide="resetAction"
    >
      <s-card class="rounded-md w-full" hide-title>
        <div class="px-2 w-[250px]">
          <div class="grid grid-cols-1">
            <s-input
              label="Action List"
              class="w-full mb-2"
              v-model="data.action.selectedAction"
              useList
              :items="[
                'Asset Depreciation',
                'Asset Disposal',
              ]"
              kind="text"
            />
            <s-input
              label="Period"
              class="w-full"
              v-model="data.action.period"
              kind="month"
            />
            <s-button
              label="Submit"
              class="w-full btn_primary flex justify-center mt-6"
              @click="generateLinesFromAction()"
              :disabled="data.action.selectedAction == ''"
            />
          </div>
        </div>
      </s-card>
    </s-modal>
  </div>
</template>

<script setup>
import { reactive, ref, computed, inject } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import { useRoute, useRouter } from "vue-router";
import { DataList, util, SButton, SInput, SModal } from "suimjs";

import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";
import helper from "@/scripts/helper.js";

import Lines from "./widget/ledgerjournal/Lines.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import References from "@/components/common/References.vue";
import Checklist from "@/components/common/Checklist.vue";

import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import moment from "moment";
import Attachment from "@/components/common/SGridAttachment.vue";


layoutStore().name = "tenant";

const FEATUREID = "LedgerJournal";

// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})

const profile = authStore().getRBAC(FEATUREID);


const listControl = ref(null);
const dimenstionCtl = ref(null)
const gridHeaderFilter = ref(null);
const gridAttachment = ref(null);
const axios = inject("axios");

const route = useRoute();
const router = useRouter();

const epFormUpdate =  "/fico/ledgerjournal/update"
const epFormInsert =  "/fico/ledgerjournal/insert"

const data = reactive({
  dataListMode: "grid",
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  record: {},
  preview: {},
  searchQuery: {
    JournalTypeIDs: [],
    Keyword: "",
    Status: "",
    DateFrom: null,
    DateTo: null,
  },
  action: {
    modal: false,
    selectedAction: "",
    period: new Date(),
  },
  customFilter: null,
  jurnalType:{},
  jType: "LEDGERACCOUNT"
});



const readOnly = computed(() => {
  return !["", "DRAFT"].includes(data.record.Status);
});
const linesTag = computed({
  get(){
    return data.record.Lines?.map(e=>{
      return `${data.jType}_${data.record._id}_${e.LineNo}`
    })
  }
})
const waitTrxSubmit = computed({
  get(){
    return ['DRAFT','READY'].includes(data.record.Status)
  }
})

const cardTitle = computed(() => {
  if (data.dataListMode == "grid") return "Ledger Journal";
  const formMode = listControl.value.getFormMode();
  return formMode == "new"
    ? `Ledger Journal - ${formMode}`
    : "Ledger Journal - " + data.record._id;
});

function preOpenAttchLine(readOnly){  
  if(readOnly) return
  axios.post("/fico/ledgerjournal/update", data.record)
}
function closeAttchLine(readOnly){ 
  if(readOnly) return 
  gridAttachment.value.refreshGrid()
}
function postDeleteLine(){
  gridAttachment.value.refreshGrid() 
}

function initNewItemFilter(item) {
  item.JournalTypeIDs = [];
  item.Dimension = []
}
function changeFilter(item, filters) {
  if (item.JournalTypeIDs.length > 0) {
    filters.push({
      Op: "$in",
      Field: "JournalTypeID",
      Value: item.JournalTypeIDs,
    });
  }
  helper.genFilterDimension(item.Dimension).forEach(e=>{
    filters.push(e)
  }) 
}
function refreshGrid() {
  listControl.value.refreshGrid();
}
function resetGridHeaderFilter() {
  gridHeaderFilter.value.reset();
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

function newRecord(record) {
  record._id = "";
  record.CurrencyID = "IDR";
  record.CompanyID = authStore().companyID;
  record.TrxDate = new Date();
  record.Name = "";
  record.Actions = [];
  record.Previews = [];
  record.Lines = [];
  record.Status = "";
  openForm(record);
}
function getJurnalType(id) {
  if (id === "" || id === null) {
      data.jurnalType = {};
      data.record.PostingProfileID = ""
      return
  }
  listControl.value.setFormLoading(true) 
  axios.post( "/fico/ledgerjournaltype/get", [id]).then(
    (r) => {
      data.jurnalType = r.data; 
      data.record.PostingProfileID = r.data.PostingProfileID;
    },
    (e) => {
      data.jurnalType = {};
      data.record.PostingProfileID = ""
      util.showError(e)
    }
  ).finally(()=>{
    listControl.value.setFormLoading(false)
  });
}
function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "VendorID":
      record.AddressAndTax.DeliveryName = v2;
      record.AddressAndTax.BillingName = v2;
      record.AddressAndTax.TaxName = v2;
    case "JournalTypeID":
      getJurnalType(v1);
      break;
  }
}

function onControlModeChanged(mode) {
  data.dataListMode = mode;
}

function postJournal() {
  data.appMode = "preview";
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    if (["ReferenceTemplateID", "ChecklistTemplateID"].includes(e.field))
      return;
    listControl.value.setFormFieldAttr(e.field, "required", required);
  });
}


function trxPreSubmit(status, action, doSubmit) {
  if (waitTrxSubmit.value) { 
    listControl.value.setFormCurrentTab(0)
    trxSubmit(doSubmit);
  }
}
function trxSubmit(doSubmit) {
  setFormRequired(true);
  util.nextTickN(2, () => {
    const valid = listControl.value.formValidate();
    const validDimension = dimenstionCtl.value.validate()
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
function openForm(record) {
  data.record = record;

  data.preview = {};
  util.nextTickN(2, () => {
    getJurnalType(record.JournalTypeID);
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
  });
}

function resetAction() {
  data.action.modal = false;
  data.action.selectedAction = "";
  data.action.period = new Date();
}

function generateLinesFromAction() {
  let f = data.record.Dimension.find((o) => o.Key == "Site");
  let siteID = f ? f.Value : "";
  const url = "/bagong/asset/get-depreciation";
  let param = {
    SiteID: siteID,
    Period: moment(data.action.period).format("YYYY-MM"),
    JournalTypeID: data.record.JournalTypeID,
  };
  axios.post(url, param).then(
    (r) => {
      data.record.Lines = r.data.map(item => ({
        ...item,
        Debit: Math.round(item.Debit),
        Credit: Math.round(item.Credit),
        Amount: Math.round(item.Debit),
        PriceEach: Math.round(item.PriceEach),
      }));
      listControl.value.setFormCurrentTab(1);
      resetAction();
    },
    (e) => {
      util.showError(e);
    }
  );
}

function onAlterFormConfig(cfg) {
  if (route.query.id !== undefined) {
    let currQuery = {...route.query};
    listControl.value.selectData({ _id: currQuery.id });
    delete currQuery['id'];
    router.replace({ path: route.path, query: currQuery });
  }
}
</script>
