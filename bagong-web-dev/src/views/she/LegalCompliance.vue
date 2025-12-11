<template>
  <div>
    <s-card
      title="Legal Compliance"
      class="w-full bg-white suim_datalist card"
      hide-footer
      v-if="!data.showDetail"
    >
      <s-grid
        ref="listControl"
        no-gap
        auto-commit-line
        hide-search
        hide-sort
        hide-select
        hide-delete-button
        hide-new-button
        :config="data.gridCfg"
        read-url="/she/legalcompliance/gets"
        @selectData="onSelectData"
        :custom-filter="customFilter"
      >
        <template #header_search="{ config }">
          <s-input
            ref="refSite"
            v-model="data.search.Site"
            lookup-key="_id"
            label="Site"
            class="w-full"
            use-list
            :disabled="defaultList?.length === 1"
            :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
            :lookup-labels="['Label']"
            :lookup-searchs="['_id', 'Label']"
            :lookup-payload-builder="
              defaultList?.length > 0
                ? (...args) =>
                    helper.payloadBuilderDimension(
                      defaultList,
                      data.search.Site,
                      false,
                      ...args
                    )
                : undefined
            "
            @change="refreshData"
          ></s-input>
        </template>
        <template #item_SiteID="{ item }">
          {{ item.SiteName }}
        </template>
      </s-grid>
    </s-card>

    <data-list
      v-if="data.showDetail"
      class="card"
      ref="controlDetail"
      hide-title
      grid-config="/she/legalregisterdetail/gridconfig"
      form-config="/she/legalregisterdetail/formconfig"
      :grid-read="'/she/legalregisterdetail/gets?SiteID=' + data.record.SiteID"
      grid-mode="grid"
      grid-delete="/she/legalregisterdetail/delete"
      form-keep-label
      form-insert="/she/legalregisterdetail/save"
      form-update="/she/legalregisterdetail/save"
      :init-app-mode="data.appMode"
      grid-hide-select
      grid-hide-new
      grid-hide-sort
      @form-edit-data="openForm"
      :form-fields="['LegalDetails']"
      :form-tabs-edit="['General', 'Attachments']"
      :grid-custom-filter="customFilterDetail"
      @pre-save="onPreSave"
      @postSave="onPostSave"
      stayOnFormAfterSave
    >
      <template #grid_header_search>
        <div
          class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter"
        >
          <s-input
            useList
            label="Site"
            class="w-[300px]"
            lookup-url="/bagong/sitesetup/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            v-model="data.searchDetail.SiteID"
            disabled
          />
          <s-input
            ref="refID"
            v-model="data.searchDetail._id"
            class="w-[300px]"
            label="Ref No."
            keepLabel
            @keyup.enter="refreshDataDetail"
          ></s-input>
          <s-input
            ref="refLegalNo"
            v-model="data.searchDetail.LegalNo"
            class="w-[300px]"
            label="Legal No"
            keepLabel
            @keyup.enter="refreshDataDetail"
          ></s-input>
          <s-input
            ref="refStatus"
            v-model="data.searchDetail.Status"
            class="w-[300px]"
            label="Status"
            use-list
            :items="['Active', 'Inactive']"
            @change="refreshDataDetail"
          ></s-input>
          <button
            class="p-1 hover:bg-white hover:text-primary"
            @click="onCancel"
          >
            <mdicon name="close" size="16"></mdicon>
          </button>
        </div>
      </template>
      <template #form_input_LegalDetails="{ item, config }">
        <div class="grow" v-if="item.LegalDetails.length == 0">&nbsp;</div>
        <div v-for="(dt, idx) in item.LegalDetails" :key="idx" class="mb-4">
          <div class="grid grid-cols-2 gap-2 border-b">
            <div class="font-semibold">
              {{ dt.Subject }}
            </div>
            <div>
              <div
                v-for="(det, idx2) in dt.ActivityPoints"
                :key="idx2"
                class="grid grid-cols-3 mb-2"
              >
                <div>{{ det.Value }}</div>
                <div>
                  <s-toggle
                    v-model="det.IsComply"
                    class="w-[120px] mt-0.5"
                    yes-label="Comply"
                    no-label="Not Comply"
                    @change="onChangeToggle"
                  />
                </div>
                <uploader
                  ref="gridAttachment"
                  :journalId="det._id"
                  :config="config"
                  journalType="LEGAL_REGISTER"
                  singleSave
                  :tags="[`${item._id}_Activity_Points_${det._id}`]"
                  @pre-open="preOpenUploader"
                />
              </div>
            </div>
          </div>
        </div>
      </template>
      <template #form_header="{ item }">
        <h3 class="w-full mb-4">
          {{ item._id }}
        </h3>
      </template>
      <template #form_tab_Attachments="{ item }">
        <s-grid-attachment
          :key="data.recordDetail._id"
          :journal-id="data.recordDetail._id"
          :tags="linesTag"
          journal-type="LEGAL_COMPLIANCE"
          ref="gridLineAttachment"
          @pre-Save="preSaveAttachment"
        ></s-grid-attachment>
      </template>
      <template #form_buttons_1="{ item, inSubmission, loading }">
        <form-buttons-trx
          :disabled="loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="'JSA'"
          :moduleid="'she'"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :auto-post="false"
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import {
  reactive,
  ref,
  inject,
  watch,
  computed,
  onMounted,
  nextTick,
} from "vue";
import {
  SInput,
  util,
  SButton,
  SGrid,
  SCard,
  DataList,
  loadGridConfig,
} from "suimjs";
import helper from "@/scripts/helper.js";
import { layoutStore } from "@/stores/layout.js";
import SToggle from "@/components/common/SButtonToggle.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import Uploader from "@/components/common/Uploader.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";

layoutStore().name = "tenant";

const axios = inject("axios");
const listControl = ref(null);
const controlDetail = ref(null);
const gridAttachment = ref(null);
const linesTag = computed({
  get() {
    const tags = [`LEGAL_COMPLIANCE_${data.record._id}`];
    const idsLegal = JSON.parse(
      JSON.stringify(data.recordDetail?.LegalDetails)
    ).flatMap((item) =>
      item.ActivityPoints.map(
        (point) => `${data.recordDetail._id}_Activity_Points_${point._id}`
      )
    );
    return [...tags, ...idsLegal];
  },
});

const FormMode = computed({
  get() {
    return controlDetail.value.getFormMode();
  },
});

let customFilter = computed(() => {
  const filters = [];

  if (
    data.search.Site !== undefined &&
    data.search.Site !== null &&
    data.search.Site !== ""
  ) {
    filters.push({
      Field: "SiteID",
      Op: "$eq",
      Value: data.search.Site,
    });
  }

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

let customFilterDetail = computed(() => {
  const filters = [];
  if (data.searchDetail._id !== null && data.searchDetail._id !== "") {
    filters.push({
      Field: "_id",
      Op: "$contains",
      Value: [data.searchDetail._id],
    });
  }

  if (data.searchDetail.LegalNo !== null && data.searchDetail.LegalNo !== "") {
    filters.push({
      Field: "LegalNo",
      Op: "$eq",
      Value: data.searchDetail.LegalNo,
    });
  }

  if (data.searchDetail.Status !== null && data.searchDetail.Status !== "") {
    let Status = data.searchDetail.Status === "Active" ? true : false;
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: Status,
    });
  }

  if (
    data.searchDetail.SiteID !== undefined &&
    data.searchDetail.SiteID !== null &&
    data.searchDetail.SiteID !== ""
  ) {
    filters.push({
      Field: "SiteID",
      Op: "$eq",
      Value: data.searchDetail.SiteID,
    });
  }

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

const data = reactive({
  record: {
    _id: "",
    SiteID: "",
    LegalNo: "",
    Status: "",
  },
  searchDetail: {
    _id: "",
    SiteID: "",
    LegalNo: "",
    Status: "",
  },
  gridCfg: {},
  showDetail: false,
  recordDetail: {},
  search: {
    Site: "",
  },
});

function onSelectData(r) {
  data.record = r;
  data.searchDetail.SiteID = r.SiteID;
  data.showDetail = true;
}

function onCancel() {
  data.showDetail = false;
  data.record = {};
}

function openForm(r) {
  let PlantCompliance = r.LegalDetails.reduce(
    (sum, item) => sum + item.ActivityPoints.length,
    0
  );
  r.PlantCompliance = PlantCompliance;
  data.recordDetail = r;
}

function onChangeToggle() {
  let dt = data.recordDetail;
  let actual = dt.LegalDetails.reduce((acc, item) => {
    let countActual = item.ActivityPoints.filter((dt) => {
      return dt.IsComply;
    });
    return (acc += countActual.length);
  }, 0);

  dt.ActualCompliance = actual;
  dt.Achievement = (actual / dt.PlantCompliance) * 100;
}

function onPostSave(r) {}

function preSaveAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [`LEGAL_COMPLIANCE_${data.record._id}`];
    return asset;
  });
}

function trxPreSubmit(status, action, doSubmit) {
  if (["DRAFT"].includes(status)) {
    controlDetail.value.submitForm(
      data.record,
      () => {
        doSubmit();
      },
      () => {
        setLoadingForm(false);
      }
    );
  } else {
    doSubmit();
  }
}
function trxPostSubmit(data, action) {
  setLoadingForm(false);
}
function trxErrorSubmit() {
  setLoadingForm(false);
}
function setLoadingForm(loading) {
  controlDetail.value.setFormLoading(loading);
}
function preOpenUploader() {
  if (!data.recordDetail._id) {
    saveManual();
  }
}
function saveManual() {
  controlDetail.value.setFormLoading(true);
  controlDetail.value.submitForm(
    data.recordDetail,
    () => {
      controlDetail.value.setFormLoading(false);
    },
    () => {
      controlDetail.value.setFormLoading(false);
    }
  );
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshData();
  });
}
function refreshDataDetail() {
  util.nextTickN(2, () => {
    controlDetail.value.refreshGrid();
  });
}
onMounted(() => {
  loadGridConfig(axios, "/she/legalcompliance/gridconfig").then(
    (r) => {
      data.gridCfg = r;
    },
    (e) => util.showError(e)
  );
});
</script>
