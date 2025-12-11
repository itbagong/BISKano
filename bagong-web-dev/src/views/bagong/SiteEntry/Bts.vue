<template>
  <grid-asset
    ref="grid"
    v-if="data.appMode == 'grid'"
    :title="data.title"
    :siteEntryId="data.id"
    :site="data.siteId"
    @back="router.push('/bagong/siteEntry')"
    @selectdata_detail="selectDataDetail"
    :purpose="data.category"
    :profile="profile"
    :filters="['SearchText', 'CustomerName']"
    :group-id-value="['EXG0001', 'EXG0003']"
  />
  <s-card
    :title="data.title"
    class="w-full bg-white suim_datalist"
    hide-footer
    v-else
  >
    <s-form
      ref="frmCtl"
      v-model="data.record"
      :config="data.formCfg"
      :mode="readOnly ? 'view' : data.formMode"
      keep-label
      :buttons-on-bottom="false"
      buttons-on-top
      :tabs="['Status', 'Fuel Usage', 'Expense', 'Attachment']"
      @submit-form="onSubmitForm"
      @cancelForm="onCancelFrom"
      :hideSubmit="readOnly"
      @post-submit-form="onPostSubmitForm"
    >
      <template #buttons_1="{ item }">
        <s-button
          :disabled="data.disabledFormButton"
          v-if="data.haveDraft"
          @click="onSubmit(item._id)"
          class="btn_primary"
          label="Submit"
        ></s-button>
      </template>
      <template #tab_Fuel_Usage="{ item, mode }">
        <header-info :item="data.record" />
        <fuel-usage
          :siteEntryAssetID="item.SiteEntryAssetID"
          v-model="item.RitaseFuelUsage"
          @get-fuel-usage-records="getFuelUsageRecords"
        />
      </template>
      <template #tab_Expense="{ item, mode }">
        <header-info :item="data.record" />
        <expense
          :grid-config-url="
            readOnly || mode == 'view'
              ? '/bagong/siteexpense-read/grid/gridconfig'
              : '/bagong/siteexpense/grid/gridconfig'
          "
          :group-id-value="['EXG0001', 'EXG0003']"
          v-model="item.Expense"
          :attch-kind="`${data.attchKind}_EXPENSE`"
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.attchKind"
          @preOpenAttch="preOpenAttch"
          @closeAttch="closeAttch"
          @reOpen="reOpen"
        />
      </template>
      <template #form_header="{}">
        <header-info :item="data.record" />
      </template>
      <template #input_KMStart="{ item, config }">
        <div class="flex gap-4">
          <s-input
            keep-label
            :label="config.label"
            v-model="item.KMStart"
            kind="number"
          ></s-input>
          <uploader
            ref="gridAttachmentKMStart"
            :journalId="item._id"
            :journalType="data.attchKind"
            :config="config"
            :tags="[`${data.attchKind}_STATUS_KM_START_${item._id}`]"
            :key="1"
            bytag
            hide-label
            single-save
            @close="closeAttch"
          />
        </div>
      </template>
      <template #input_KMEnd="{ item, config }">
        <div class="flex gap-4">
          <s-input
            keep-label
            :label="config.label"
            v-model="item.KMEnd"
            kind="number"
          ></s-input>
          <uploader
            ref="gridAttachmentKMEnd"
            :journalId="item._id"
            :journalType="data.attchKind"
            :config="config"
            :tags="[`${data.attchKind}_STATUS_KM_END_${data.record._id}`]"
            :key="1"
            bytag
            hide-label
            single-save
            @close="closeAttch"
          />
        </div>
      </template>
      <template #input_RitaseDetail="{ item, mode }">
        <ritase
          :read-only="readOnly || mode == 'view'"
          v-model="item.RitaseDetail"
          grid-config="bagong/siteentry_btsdetail/detail/ritasedetail/gridconfig"
          :siteId="data.siteId"
        ></ritase>
      </template>
      <template #input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          ref="dimenstionCtl"
          :read-only="readOnly || mode == 'view'"
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #tab_Attachment="{ item, mode }">
        <header-info :item="data.record" />
        <attachment
          :read-only="readOnly || mode == 'view'"
          ref="gridAttachment"
          :journalId="item._id"
          :tags="allTag"
          single-save
          :journalType="data.attchKind"
          :key="allTag"
        />
      </template>
    </s-form>
  </s-card>
</template>
<script setup>
import { ref, reactive, onMounted, inject, computed, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  SCard,
  SInput,
  SGrid,
  SForm,
  loadGridConfig,
  loadFormConfig,
  util,
  DataList,
  SListEditor,
  createFormConfig,
  SButton,
} from "suimjs";
import { authStore } from "@/stores/auth.js";
import helper from "@/scripts/helper.js";

import Expense from "./../widget/SiteEntry/Expense.vue";
import FuelUsage from "./../widget/SiteEntry/BtsFuelUsage.vue";
import HeaderInfo from "./../widget/SiteEntry/HeaderInfo.vue";
import GridAsset from "./../widget/SiteEntry/GridAsset.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import Ritase from "./../widget/SiteEntry/BtsRitase.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import Uploader from "@/components/common/Uploader.vue";
import { useRoute, useRouter } from "vue-router";

layoutStore().name = "tenant";

const FEATUREID = "SiteEntryBTS";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const route = useRoute();
const router = useRouter();

const grid = ref(null);
const dimenstionCtl = ref(null);
const frmCtl = ref(null);
const inputEditor = ref(null);
const editor = ref(null);
const gridAttachment = ref(null);
const gridAttachmentKMStart = ref(null);
const gridAttachmentKMEnd = ref(null);

const data = reactive({
  status: "Draft",
  id: route.query.id,
  siteId: route.query.siteId,
  appMode: "grid",
  formMode: "new",
  formCfg: {},
  record: {
    Expense: [],
  },
  title: "BTS",
  category: "BTS",
  cfgRitase: {},
  objRitase: {},
  disabledFormButton: false,
  attchKind: "SE_BTS",
  haveDraft: false,
  haveSubmitted: false,
});
const allTag = computed({
  get() {
    return [
      ...data.record.Expense?.map((e) => {
      return `${data.attchKind}_EXPENSE_${data.record._id}_${e.ID}`;
    }),
    `${data.attchKind}_STATUS_KM_START_${data.record._id}`,
    `${data.attchKind}_STATUS_KM_END_${data.record._id}`,
    ];
  },
});
const readOnly = computed({
  get() {
    return data.haveSubmitted;
  },
});
function preOpenAttch(readOnly) {
  if (readOnly) return;
  frmCtl.value.submit();
}
function closeAttch(readOnly) {
  if (readOnly) return;
  gridAttachment.value.refreshGrid();
}
function hasJournalID(journalID){
  return journalID != '' && journalID != undefined
}
watch(
  () => data.record.Expense,
  (nv) => {
    checkLineStatus(nv);
  },
  { deep: true }
);

function checkLineStatus(lines) {
  if (lines.length == 0) {
    data.haveDraft = true;
    data.haveSubmitted = false;
  } else {
    const drafts = lines.filter((e) => e.JournalID == "");
    data.haveDraft = drafts.length > 0;

    const submitteds = lines.filter((e) => e.JournalID != "");
    data.haveSubmitted = submitteds.length > 0;
  }
}
function selectDataDetail(record, siteEntryAsset) {
  data.record = record;

  checkLineStatus(data.record.Expense);

  setTimeout(() => {
    data.appMode = "form";

    util.nextTickN(2, () => {
      if (record.Status == "Partial" || record.Status == "Breakdown") {
        frmCtl.value.setFieldAttr("SpareAsset", "hide", false);
        frmCtl.value.setFieldAttr(
          "SpareAsset",
          "lookupUrl",
          `/tenant/asset/find?GroupID=UNT&Dimension.Key=Site&Dimension.Value=${data.siteId}`
        );
      } else {
        frmCtl.value.setFieldAttr("SpareAsset", "hide", true);
      }
    });
  }, 500);
}

function onCancelFrom() {
  data.appMode = "grid";
  util.nextTickN(2, () => {
    grid.value.refreshGrid();
  });
}
function setFormLoading(loading) {
  data.disabledFormButton = loading;
  frmCtl.value.setLoading(loading);
}
function onSubmitForm(record, cbSuccess, cbError) {
  save(cbSuccess, cbError);
}

function reOpen() {
  const cbSuccess = () => {
    checkLineStatus(data.record.Expense);
    setFormLoading(false);
  };
  const cbError = () => {
    setFormLoading(false);
  };
  setFormLoading(true);

  save(cbSuccess, cbError);
}

function save(cbSuccess = () => {}, cbError = () => {}) {
  const param = data.record;
  axios.post("/bagong/siteentry/save-asset-detail", param).then(
    (r) => {
      cbSuccess();
    },
    (e) => {
      util.showError(e);
      cbError();
    }
  );
}

function onSubmit() {
  frmCtl.value.setCurrentTab(0);

  const validDimension = dimenstionCtl.value.validate();
  const validExpense = helper.validateSiteEntryExpense(data.record.Expense);

  if (validDimension && validExpense) {
    setFormLoading(true);
    save(doSubmit);
  }
}
function doSubmit() {
  const url = "/bagong/postingprofile/post";
  const param = {
    JournalType: "SITEENTRY_BTS",
    JournalID: data.record._id,
    Op: "Submit",
    Text: "",
  };
  axios.post(url, param).then(
    (r) => {
      setFormLoading(false);
      onCancelFrom();
    },
    (e) => {
      util.showError(e);
      setFormLoading(false);
    }
  );
}

function getFuelUsageRecords(records) {
  // data.record = {
  //   ...data.record,
  //   KMStart: records[0].KMStart,
  //   KMEnd: records[records.length - 1].KMEnd,
  // };
}

function genCfgRitase() {
  const cfg = createFormConfig("", true);
  cfg.addSection("", true).addRow(
    {
      field: "Name",
      label: "Name",
      required: true,
      kind: "text",
    },
    {
      field: "KMStart",
      kind: "number",
      label: "KM Start",
      required: true,
    },
    {
      field: "KMEnd",
      kind: "number",
      label: "KM End",
      required: true,
    }
  );
  data.cfgRitase = cfg.generateConfig();
}

function addItemEditor(record) {
  let isValid = inputEditor.value.validate();
  for (let ky in data.objRitase) {
    let val = data.objRitase[ky];
    record[ky] = val;
  }
  editor.value.setValidateItem(isValid);
  if (isValid) data.objRitase = {};
}

function onPostSubmitForm(r) {}

function validateExpense() {
  const r = data.record.Expense.filter((o) => o.TotalAmount == 0);

  const valid = r.length == 0;
  if (!valid) {
    util.showError("Amount must be > 0");
  }
  return valid;
}

onMounted(() => {
  genCfgRitase();
  loadFormConfig(axios, "/bagong/siteentry_btsdetail/formconfig").then(
    (r) => {
      data.formCfg = r;
    },
    (e) => util.showError(e)
  );
});

watch(
  () => data.record,
  (nv) => {
    util.nextTickN(2, () => {
      util.nextTickN(2, () => {
        if (nv.Status == "Partial" || nv.Status == "Breakdown") {
          frmCtl.value.setFieldAttr("SpareAsset", "hide", false);
          frmCtl.value.setFieldAttr(
            "SpareAsset",
            "lookupUrl",
            `/tenant/asset/find?GroupID=UNT&Dimension.Key=Site&Dimension.Value=${data.siteId}`
          );
        } else {
          frmCtl.value.setFieldAttr("SpareAsset", "hide", true);
        }
      });
    });
  },
  { deep: true }
);
</script>
