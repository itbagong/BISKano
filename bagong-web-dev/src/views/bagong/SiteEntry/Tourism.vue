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
    :group-id-value="['EXG0001', 'EXG0005']"
  />

  <s-card
    :title="data.title"
    class="w-full bg-white suim_datalist"
    hide-footer
    v-else-if="data.appMode == 'form'"
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
      <template #input_DriverID="{ item, config, mode }">
        <s-input
          v-model="item.DriverID"
          :label="config.label"
          use-list
          :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${data.siteId}&Position=Driver`"
          :lookup-searchs="['_id', 'Name']"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :read-only="readOnly || mode === 'view'"
        ></s-input>
      </template>
      <template #input_DriverID2="{ item, config, mode }">
        <s-input
          v-model="item.DriverID2"
          :label="config.label"
          use-list
          :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${data.siteId}&Position=Driver`"
          :lookup-searchs="['_id', 'Name']"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :read-only="readOnly || mode === 'view'"
        ></s-input>
      </template>
      <template #input_Helper="{ item, config, mode }">
        <s-input
          v-model="item.Helper"
          :label="config.label"
          use-list
          :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${data.siteId}&Position=Driver Helper`"
          :lookup-searchs="['_id', 'Name']"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :read-only="readOnly || mode === 'view'"
        ></s-input>
      </template>
      <template #tab_Expense="{ item }">
        <header-info :item="data.record" />
        <div class="w-full" v-if="data.loading">
          <loader kind="skeleton" skeleton-kind="list" />
        </div>
        <template v-else>
          <div class="w-full p-2 bg-slate-100 mb-2">
            <div class="font-semibold">Operational Expense</div>
          </div>
          <expense
            hide-control
            hide-delete-button
            :grid-config-url="
              readOnly || mode == 'view'
                ? '/bagong/siteexpense-read/grid/gridconfig'
                : '/bagong/siteexpense/grid/gridconfig'
            "
            :group-id-value="['EXG0001', 'EXG0005']"
            v-model="item.OperationalExpense"
            :attch-kind="`${data.attchKind}_EXPENSE`"
            :attch-refId="item._id"
            :attch-tag-prefix="data.attchKind + '_OPERATIONAL'"
            @preOpenAttch="preOpenAttch"
            @closeAttch="closeAttch"
            @reOpen="reOpen"
          />
          <div class="w-full p-2 bg-slate-100 mb-2">
            <div class="font-semibold">Other Expense</div>
          </div>
        </template>

        <expense
          :grid-config-url="
            readOnly
              ? '/bagong/siteexpense-read/grid/gridconfig'
              : '/bagong/siteexpense/grid/gridconfig'
          "
          :group-id-value="['EXG0001', 'EXG0005']"
          v-model="item.OtherExpense"
          :attch-kind="`${data.attchKind}_EXPENSE`"
          :attch-ref-id="item._id"
          :attch-tag-prefix="data.attchKind + '_OTHER'"
          @preOpenAttch="preOpenAttch"
          @closeAttch="closeAttch"
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
      <template #tab_Attachment="{ item }">
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
import { reactive, onMounted, inject, ref, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  SCard,
  SGrid,
  SForm,
  SInput,
  loadGridConfig,
  loadFormConfig,
  util,
  DataList,
  SButton,
  SModal,
} from "suimjs";
import { authStore } from "@/stores/auth.js";
import helper from "@/scripts/helper.js";

import Expense from "./../widget/SiteEntry/Expense.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import HeaderInfo from "./../widget/SiteEntry/HeaderInfo.vue";
import GridAsset from "./../widget/SiteEntry/GridAsset.vue";
// import Expense from "./../widget/SiteEntry/TourismExpense.vue";
import Loader from "@/components/common/Loader.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import { useRoute, useRouter } from "vue-router";

layoutStore().name = "tenant";

const FEATUREID = "SiteEntryTourism";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const route = useRoute();
const router = useRouter();

const grid = ref(null);
const dimenstionCtl = ref(null);
const tireUsageTab = ref(null);
const frmCtl = ref(null);
const gridAttachment = ref(null);

const data = reactive({
  status: "Draft",
  id: route.query.id,
  siteId: route.query.siteId,
  appMode: "grid",
  formMode: "",
  formCfg: {},
  title: "Tourism",
  category: "Tourism",
  record: {
    OperationalExpense: [],
    OtherExpense: [],
  },
  loading: false,
  disabledFormButton: false,
  attchKind: "SE_TOURISM",
  haveDraft: false,
  haveSubmitted: false,
});
const readOnly = computed({
  get() {
    return data.haveSubmitted;
  },
});
const allTag = computed({
  get() {
    return [
      ...data.record.OperationalExpense?.map((e) => {
        return `${data.attchKind}_OPERATIONAL_EXPENSE_${data.record._id}_${e.ID}`;
      }),
      ...data.record.OtherExpense?.map((e) => {
        return `${data.attchKind}_OTHER_EXPENSE_${data.record._id}_${e.ID}`;
      }),
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
watch(
  () => data.record.OperationalExpense,
  (nv) => {
    checkLineStatus([...nv, ...data.record.OtherExpense]);
  },
  { deep: true }
);

watch(
  () => data.record.OtherExpense,
  (nv) => {
    checkLineStatus([...nv, ...data.record.OperationalExpense]);
  },
  { deep: true }
);

function checkLineStatus(lines) {
  console.log(lines);
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

  checkLineStatus([
    ...data.record.OperationalExpense,
    ...data.record.OtherExpense,
  ]);

  setTimeout(() => {
    data.appMode = "form";

    util.nextTickN(2, () => {
      let siteID = route.query.siteId;
      if (!readOnly.value && data.record.OperationalExpense.length === 0) {
        getExpenseTypeBySite(data.siteId);
      }
    });
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

function save(cbSuccess = () => {}, cbError = () => {}) {
  const url = "/bagong/siteentry/save-asset-detail";

  const param = data.record;

  axios.post(url, param).then(
    (r) => {
      cbSuccess();
    },
    (e) => {
      cbError();
      util.showError(e);
    }
  );
}
function setLineExpense() {
  let lineNo = 1;
  data.record.OperationalExpense.map((e) => {
    e.LineNo = lineNo;
    lineNo++;
    return e;
  });
  data.record.OtherExpense.map((e) => {
    e.LineNo = lineNo;
    lineNo++;
    return e;
  });
}
function onSubmit() {
  frmCtl.value.setCurrentTab(0);

  const validDimension = dimenstionCtl.value.validate();
  const validOpearationalExpense = helper.validateSiteEntryExpense(
    data.record.OperationalExpense
  );
  const validOtherExpense = helper.validateSiteEntryExpense(
    data.record.OtherExpense
  );

  if (validDimension && validOpearationalExpense && validOtherExpense) {
    setLineExpense();
    setFormLoading(true);
    save(doSubmit);
  }
}
function doSubmit() {
  const url = "/bagong/postingprofile/post";
  const param = {
    JournalType: "SITEENTRY_TOURISM",
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
function reOpen() {
  const cbSuccess = () => {
    checkLineStatus([
      ...data.record.OperationalExpense,
      ...data.record.OtherExpense,
    ]);
    setFormLoading(false);
  };
  const cbError = () => {
    setFormLoading(false);
  };
  setFormLoading(true);

  save(cbSuccess, cbError);
}

function getExpenseTypeBySite(SiteID) {
  data.loading = true;
  axios
    .post("bagong/siteentry/get-expense-types-by-site", { SiteID })
    .then((r) => {
      try {
        const oprExpense = r.data.map((item, index) => {
          return {
            LineNo: index + 1,
            ID: item._id,
            Name: item.Name,
            ExpenseTypeID: item._id,
            Amount: item.Value,
            Value: 1,
            TotalAmount: item.Value,
            JournalID: "",
          };
        });
        data.record.OperationalExpense = oprExpense;
        // data.loading = false;
      } catch (err) {
        util.showError(err);
      }
    })
    .finally(() => {
      data.loading = false;
    });
}

function onPostSubmitForm(r) {}

function checkExpense(params, field) {
  let f = params[field].filter((o) => o.TotalAmount == 0);
  return f.length > 0;
}

onMounted(() => {
  loadFormConfig(axios, "/bagong/siteentry_tourismdetail/formconfig").then(
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
