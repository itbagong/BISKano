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
    :group-id-value="['EXG0001', 'EXG0004']"
    :filters="['SearchText', 'TrayekName']"
    @get-grid-records="getGridAssetRecords"
  >
  </grid-asset>

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
      :mode="data.formMode"
      keep-label
      :buttons-on-bottom="false"
      buttons-on-top
      :tabs="['Status', 'Ritase', 'Attachment']"
      :hide-submit="!profile.canUpdate"
      @submit-form="save"
      @cancelForm="onCancelFrom"
      @post-submit-form="onPostSubmitForm"
    >
      <template #loader>
        <loader />
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
          :read-only="mode == 'view'"
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
          :read-only="mode == 'view'"
        ></s-input>
      </template>
      <template #input_KondekturID="{ item, config, mode }">
        <s-input
          v-model="item.KondekturID"
          :label="config.label"
          use-list
          :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${data.siteId}&Position=Kondektur`"
          :lookup-searchs="['_id', 'Name']"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :read-only="mode == 'view'"
        ></s-input>
      </template>
      <template #input_KondekturID2="{ item, config, mode }">
        <s-input
          v-model="item.KondekturID2"
          :label="config.label"
          use-list
          :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${data.siteId}&Position=Kondektur`"
          :lookup-searchs="['_id', 'Name']"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :read-only="mode == 'view'"
        ></s-input>
      </template>
      <template #input_KernetID="{ item, config, mode }">
        <s-input
          v-model="item.KernetID"
          :label="config.label"
          use-list
          :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${data.siteId}&Position=Kernet`"
          :lookup-searchs="['_id', 'Name']"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :read-only="mode == 'view'"
        ></s-input>
      </template>
      <template #input_KernetID2="{ item, config, mode }">
        <s-input
          v-model="item.KernetID2"
          :label="config.label"
          use-list
          :lookup-url="`/bagong/employee/get-employee-by-position?SiteID=${data.siteId}&Position=Kernet`"
          :lookup-searchs="['_id', 'Name']"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :read-only="mode == 'view'"
        ></s-input>
      </template>
      <template #tab_Ritase="{ item }">
        <header-info :item="data.record" />
        <ritase
          ref="ritaseRef"
          :siteEntryAssetID="item._id"
          :trayek-name="data.record.TrayekName"
          :site="data.siteId"
          :assetID="data.assetId"
          :hide-edit="!profile.canUpdate"
          @refresh-attach="onRefreshAttach"
        />
      </template>
      <template #form_header="{}">
        <header-info :item="data.record" />
      </template>
      <template #input_Dimension="{ item }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #tab_Attachment="{ item }">
        <header-info :item="data.record" />
        <attachment
          ref="gridAttachment"
          v-model="item.Attachment"
          :journalId="item._id"
          :journalType="data.attchKind"
          :tags="allTag"
          :key="allTag"
        />
      </template>
    </s-form>
  </s-card>
</template>

<script setup>
import {
  ref,
  reactive,
  onMounted,
  inject,
  watch,
  nextTick,
  computed,
} from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  SCard,
  SGrid,
  SForm,
  SInput,
  SButton,
  loadGridConfig,
  loadFormConfig,
  util,
  DataList,
} from "suimjs";
import { authStore } from "@/stores/auth.js";

import HeaderInfo from "../widget/SiteEntry/HeaderInfo.vue";
import GridAsset from "./../widget/SiteEntry/GridAsset.vue";
import Ritase from "./../widget/SiteEntry/Ritase.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import Loader from "@/components/common/Loader.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import { useRoute, useRouter } from "vue-router";
layoutStore().name = "tenant";

const FEATUREID = "SiteEntryTrayek";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const route = useRoute();
const router = useRouter();

const grid = ref(null);
const frmCtl = ref(null);
const formButtonsTrx = ref(null);
const gridAttachment = ref(null);
const ritaseRef = ref(null);

const data = reactive({
  id: route.query.id,
  siteId: route.query.siteId,
  assetId: "",
  selectedSiteEntryAsset: null,
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  formCfg: {},
  record: {},
  title: "Trayek",
  category: "Trayek",
  trayekNames: [],
  policeNums: [],
  filterQuery: {
    trayekName: null,
    policeNo: null,
  },

  attchKind: "SE_TRAYEK",
  
});

function selectDataDetail(record, siteEntryAsset) {
  data.record = {};
  data.assetId = siteEntryAsset.AssetID;
  data.selectedSiteEntryAsset = siteEntryAsset
  util.nextTickN(2, () => {
    data.appMode = "form";
    data.record = record;
    util.nextTickN(2, () => {
      let siteID = route.query.siteId;
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
  });
}

function onCancelFrom() {
  data.appMode = "grid";
  util.nextTickN(2, () => {
    grid.value.refreshGrid();
  });
}
function onRefreshAttach() {
  gridAttachment.value.refreshGrid();
}
const allTag = computed({
  get() {
    return [
      `${data.attchKind}_${data.selectedSiteEntryAsset._id}`,
      `${data.attchKind}_RITASE_${data.selectedSiteEntryAsset._id}`
    ];
    // const records = ritaseRef?.value?.getRecords();
    // if (!Array.isArray(records)) {
    //   return [];
    // }
    // return [
    //   ...records.flatMap((e) => {
    //     return [
    //       `${data.attchKind}_RITASE_KM_START_${e._id}`,
    //       `${data.attchKind}_RITASE_KM_END_${e._id}`,
    //     ];
    //   }),
    // ];
  },
});
function save(record, cbSuccess, cbError) {
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
watch(
  () => frmCtl.value?.getCurrentTab(),
  (nv) => {}
);
// function onSubmit() {}
function preSubmitTrx(action, doSubmit) {
  if (action === "Submit") {
    frmCtl.value.setFieldAttr("DriverID", "required", true);
    frmCtl.value.setFieldAttr("Status", "required", true);

    nextTick(() => {
      const valid = frmCtl.value.validate();
      if (valid) {
        doSubmit();
      }
    });
  }
}

// function mappingDataFilter(records, field) {
//   const filterFlag = field == "TrayekName" ? "trayekNames" : "policeNums";

//   const datas = records
//     .filter((v) => v[field] != "")
//     .map((item) => item[field]);
//   const filterSet = new Set(datas);
//   const uniqueFilterSet = Array.from(filterSet);

//   data[filterFlag] = uniqueFilterSet;
// }

// function getGridAssetRecords(records) {
//   if (records.length > 0) {
//     // collect data trayek names
//     mappingDataFilter(records, "TrayekName");
//     // collect data police nums
//     mappingDataFilter(records, "PoliceNum");
//   }
// }

function onPostSubmitForm(r) {
  gridAttachment.value.Save();
}

onMounted(() => {
  loadFormConfig(axios, "/bagong/siteentry_trayekdetail/formconfig").then(
    (r) => {
      data.formCfg = r;
    },
    (e) => util.showError(e)
  );
});

watch(
  () => [data.filterQuery.trayekName, data.filterQuery.policeNo],
  (nv) => {
    util.nextTickN(2, () => {
      grid.value.filterRecords(nv[0], nv[1]);
    });
  },
  { deep: true }
);

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
