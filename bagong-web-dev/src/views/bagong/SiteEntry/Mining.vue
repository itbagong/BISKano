<template>
  <div>
    <grid-asset
      ref="gridCtl"
      v-if="data.appMode == 'grid'"
      :title="data.title"
      :siteEntryId="data.id"
      :site="data.siteId"
      @back="router.push('/bagong/siteEntry')"
      @selectdata_detail="selectDataDetail"
      :purpose="data.category"
      :profile="profile"
      :filters="['SearchText', 'CustomerName']"
      :group-id-value="['EXG0001', 'EXG0002']"
    />

    <s-card
      :title="data.title"
      class="w-full bg-white suim_datalist"
      hide-footer
      v-if="data.appMode == 'form' && data.formCfg && data.formCfg.setting"
    >
      <s-form
        ref="frmCtl"
        v-model="data.record"
        :config="data.formCfg"
        :mode="readOnly ? 'view' : data.formMode"
        keep-label
        :buttons-on-bottom="false"
        buttons-on-top
        :tabs="['Status', 'Expense', 'Attachment']"
        @submit-form="onSaveForm"
        @cancelForm="onCancelFrom"
        :hideSubmit="readOnly && !data.haveDraft"
        @post-submit-form="onPostSubmitForm"
        @fieldChange="onfieldChange"
      >
        <template #input__id="{ item, config }">
          <s-input
            v-model="item._id"
            keep-label
            :label="config.label"
            :disabled="true"
          ></s-input>
        </template>
        <template #input_DriverID="{ item, config, mode }">
          <s-input
            v-model="item.DriverID"
            :label="config.label"
            use-list
            :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${
              data.siteId
            }&Position=${data.record.DriverType != 'All' ? 'Driver' : ''}&Op=${
              data.record.DriverType == 'Driver' ? 'Contain' : 'Not Contain'
            }`"
            lookup-key="_id"
            :lookup-searchs="['_id', 'Name']"
            :lookup-labels="['Name']"
            :read-only="mode === 'view'"
          ></s-input>
        </template>
        <template #input_DriverIDReplacement="{ item, config, mode }">
          <s-input
            v-model="item.DriverIDReplacement"
            :label="config.label"
            use-list
            :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${
              data.siteId
            }&Position=${data.record.DriverType != 'All' ? 'Driver' : ''}&Op=${
              data.record.DriverType == 'Driver' ? 'Contain' : 'Not Contain'
            }`"
            lookup-key="_id"
            :lookup-searchs="['_id', 'Name']"
            :lookup-labels="['Name']"
            :read-only="mode === 'view'"
          ></s-input>
        </template>
        <template #input_DriverID2="{ item, config, mode }">
          <s-input
            v-model="item.DriverID2"
            :label="config.label"
            use-list
            :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${
              data.siteId
            }&Position=${data.record.DriverType != 'All' ? 'Driver' : ''}&Op=${
              data.record.DriverType == 'Driver' ? 'Contain' : 'Not Contain'
            }`"
            lookup-key="_id"
            :lookup-searchs="['_id', 'Name']"
            :lookup-labels="['Name']"
          ></s-input>
        </template>
        <template #input_DriverID2Replacement="{ item, config, mode }">
          <s-input
            v-model="item.DriverID2Replacement"
            :label="config.label"
            use-list
            :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${
              data.siteId
            }&Position=${data.record.DriverType != 'All' ? 'Driver' : ''}&Op=${
              data.record.DriverType == 'Driver' ? 'Contain' : 'Not Contain'
            }`"
            lookup-key="_id"
            :lookup-searchs="['_id', 'Name']"
            :lookup-labels="['Name']"
            :read-only="mode === 'view'"
          ></s-input>
        </template>
        <template #input_StartKM="{ item, config }">
        <div class="flex gap-4">
          <s-input
            keep-label
            :label="config.label"
            v-model="item.StartKM"
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
      <template #input_EndKM="{ item, config }">
        <div class="flex gap-4">
          <s-input
            keep-label
            :label="config.label"
            v-model="item.EndKM"
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
      <template #input_KM="{ item, config }">
        <GridKM 
          v-if="item._id !== undefined"
          :read-only="readOnly || mode == 'view'"
          v-model="item.KM"
          grid-config="bagong/siteentry_miningdetail/detail/kmdetail/gridconfig"
          :siteId="data.siteId"
          :recordID="item._id"
          :attchKind="data.attchKind"
          @close="closeAttch"
        />
      </template>
        <template #buttons_1="{ item }">
          <s-button
            :disabled="data.disabledFormButton"
            v-if="data.haveDraft"
            @click="onSubmit(item._id)"
            class="btn_primary"
            label="Submit"
          ></s-button>
        </template>
        <template #tab_Expense="{ item, mode }">
          <header-info :item="data.record" />
          <expense
            :grid-config-url="
              readOnly || mode == 'view'
                ? '/bagong/siteexpense-read/grid/gridconfig'
                : '/bagong/siteexpense/grid/gridconfig'
            "
            :group-id-value="['EXG0001', 'EXG0002']"
            v-model="item.Expense"
            kind="SE_MINING_EXPENSE"
            :attch-kind="`${data.attchKind}_EXPENSE`"
            :attch-ref-id="item._id"
            :attch-tag-prefix="data.attchKind"
            @preOpenAttch="preOpenAttch"
            @closeAttch="closeAttch"
            @reOpen="reOpen"
          />
        </template>
        <template #tab_Attachment="{ item, mode }">
          <header-info :item="data.record" />
          <attachment
            :read-only="readOnly || mode == 'view'"
            ref="gridAttachment"
            :journalId="item._id"
            :journalType="data.attchKind"
            :tags="allTag"
            single-save
          />
        </template>
        <template #input_Dimension="{ item, mode }">
          <dimension-editor-vertical
            ref="dimenstionCtl"
            :read-only="readOnly || mode == 'view'"
            v-model="item.Dimension"
            :default-list="profile.Dimension"
          ></dimension-editor-vertical>
        </template>
        <template #form_header="{}">
          <header-info :item="data.record" />
        </template>
      </s-form>
    </s-card>
  </div>
</template>
<script setup>
import { reactive, onMounted, inject, ref, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  SCard,
  SGrid,
  SForm,
  loadGridConfig,
  loadFormConfig,
  util,
  DataList,
  SButton,
  SModal,
  SInput,
} from "suimjs";
import { authStore } from "@/stores/auth.js";
import helper from "@/scripts/helper.js";

import Expense from "./../widget/SiteEntry/Expense.vue";
import GridKM from "./../widget/SiteEntry/GridKM.vue";
// import TireUsage from "./../widget/SiteEntry/TireUsage.vue";
// import OilUsage from "./../widget/SiteEntry/OilUsage.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import moment from "moment";
import HeaderInfo from "../widget/SiteEntry/HeaderInfo.vue";
import GridAsset from "./../widget/SiteEntry/GridAsset.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import Uploader from "@/components/common/Uploader.vue";
import { useRoute, useRouter } from "vue-router";

layoutStore().name = "tenant";

const FEATUREID = "SiteEntryMining";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const route = useRoute();
const router = useRouter();

const gridCtl = ref(null);
const dimenstionCtl = ref(null);
const tireUsageTab = ref(null);
const frmCtl = ref(null);
const gridAttachment = ref(null);
const data = reactive({
  status: "DRAFT",
  id: route.query.id,
  siteId: route.query.siteId,
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  formCfg: {},
  title: "Mining",
  category: "Mining",
  record: {
    Expense: [],
    Attachment: [],
    KM: [],
  },
  recTireUsage: {},
  recOilUsage: {},
  disabledFormButton: false,
  attchKind: "SE_MINING",
  haveDraft: false,
  haveSubmitted: false,
});

const allTag = computed({
  get() {
    return [
      ...data.record.Expense?.map((e) => {
      return `${data.attchKind}_EXPENSE_${data.record._id}_${e.ID}`;
    }),
    `${data.attchKind}_KM_START_${data.record._id}`,
    `${data.attchKind}_KM_END_${data.record._id}`,
    ];
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
function checkJournal(items) {
  if (Array.isArray(items) && items.length > 0)
    return items[0].JournalID !== "";
  return false;
}
function getStatus(record) {
  if (checkJournal(record.Expense)) return "SUBMITTED";
  return "DRAFT";
}

const readOnly = computed({
  get() {
    return data.haveSubmitted;
  },
});
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
  data.record = { ...record, DriverType: "Driver" };

  checkLineStatus(data.record.Expense);

  setTimeout(() => {
    data.appMode = "form";

    util.nextTickN(2, () => {
      let siteID = route.query.siteId;
      if (record.Status == "Partial" || record.Status == "Breakdown") {
        frmCtl.value.setFieldAttr("SpareAsset", "hide", false);
        frmCtl.value.setFieldAttr(
          "SpareAsset",
          "lookupUrl",
          `/tenant/asset/find?Dimension.Key=Site&Dimension.Value=${data.siteId}`
        );
        const cfgSpareAsset = frmCtl.value.getFormField("SpareAsset");
        frmCtl.value.setFormFieldAttr(
          "SpareAsset",
          "lookupPayloadBuilder",
          (search) => {
            return helper.payloadBuilderSpareAsset(
              search,
              cfgSpareAsset,
              record.SpareAsset,
              ["PRT", "ELC"]
            );
          }
        );
      } else {
        frmCtl.value.setFieldAttr("SpareAsset", "hide", true);
      }
    });
  }, 500);
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

function onCancelFrom() {
  data.appMode = "grid";
  util.nextTickN(2, () => {
    gridCtl.value.refreshGrid();
  });
}
function setFormLoading(loading) {
  data.disabledFormButton = loading;
  frmCtl.value.setLoading(loading);
}

function onSaveForm(record, cbSuccess, cbError) {
  save(cbSuccess, cbError);
}

function save(cbSuccess = () => {}, cbError = () => {}) {
  const param = data.record;
  axios.post("/bagong/siteentry/save-asset-detail", param).then(
    (r) => {
      cbSuccess();
    },
    (e) => {
      cbError();
      util.showError(e);
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
    JournalType: "SITEENTRY_MINING",
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

function onfieldChange(name, value1, value2, oldValue) {
  if (name == "Status") {
    if (value1 == "Partial" || value1 == "Breakdown") {
      frmCtl.value.setFieldAttr("SpareAsset", "hide", false);
      frmCtl.value.setFieldAttr(
        "SpareAsset",
        "lookupUrl",
        `/tenant/asset/find?Dimension.Key=Site&Dimension.Value=${data.siteId}`
      );
      const cfgSpareAsset = frmCtl.value.getField("SpareAsset");
      frmCtl.value.setFieldAttr(
        "SpareAsset",
        "lookupPayloadBuilder",
        (search) => {
          return helper.payloadBuilderSpareAsset(
            search,
            cfgSpareAsset,
            data.record.SpareAsset,
            ["PRT", "ELC"]
          );
        }
      );
    } else {
      frmCtl.value.setFieldAttr("SpareAsset", "hide", true);
    }
  }
}

onMounted(() => {
  loadFormConfig(axios, "/bagong/siteentry_miningdetail/formconfig").then(
    (r) => {
      data.formCfg = r;
    },
    (e) => util.showError(e)
  );
});
</script>
