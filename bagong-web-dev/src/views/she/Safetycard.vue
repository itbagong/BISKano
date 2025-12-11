<template>
  <div class="w-full Savety-Card">
    <data-list
      class="card"
      ref="listControl"
      title="Safety Card"
      grid-config="/she/safetycard/gridconfig"
      form-config="/she/safetycard/formconfig"
      grid-read="/she/safetycard/gets"
      form-read="/she/safetycard/get"
      grid-mode="grid"
      grid-delete="/she/safetycard/delete"
      form-keep-label
      form-insert="/she/safetycard/save"
      form-update="/she/safetycard/save"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      gridHideSelect
      @form-new-data="newRecord"
      :formHideSubmit="readOnly"
      @form-edit-data="editForm"
      :grid-custom-filter="customFilter"
      :form-fields="[
        'Dimension',
        'DetailsFinding',
        'FollowUp',
        'ActivityID',
        'Pica', 
      ]"
      :grid-fields="['DetailsFinding', 'FollowUp', 'Pica', 'Status']"
      :form-tabs-edit="['General', 'Detail Finding', 'Follow Up', 'Attachment']"
      :form-tabs-view="['General', 'Detail Finding', 'Follow Up', 'Attachment']"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @alterGridConfig="alterGridConfig"
      @preSave="onPreSave"
      stay-on-form-after-save
      @form-field-change="onFormFieldChange"
    >
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #grid_header_search="{ config }">
        <div
          class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter"
        >
          <s-input
            kind="date"
            label="Date From"
            class="w-[200px]"
            v-model="data.search.DateFrom"
            @change="refreshData"
          ></s-input>
          <s-input
            kind="date"
            label="Date To"
            class="w-[200px]"
            v-model="data.search.DateTo"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refCategory"
            v-model="data.search.Category"
            lookup-key="_id"
            label="Category"
            class="w-[200px]"
            use-list
            :lookup-url="`/tenant/masterdatatype/find?ParentID=SCC`"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refFinding"
            v-model="data.search.Finding"
            lookup-key="_id"
            label="Finding"
            class="w-[200px]"
            @keyup.enter="refreshData"
          ></s-input>
          <s-input
            ref="refResponse"
            v-model="data.search.Response"
            lookup-key="_id"
            label="Response"
            class="w-[200px]"
            @keyup.enter="refreshData"
          ></s-input>
          <s-input
            ref="refPosition"
            v-model="data.search.Position"
            lookup-key="_id"
            label="Position "
            class="w-[300px]"
            use-list
            :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=PTE`"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refSite"
            v-model="data.search.Site"
            lookup-key="_id"
            label="Site"
            class="w-[300px]"
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
          <s-input
            ref="refLocation"
            v-model="data.search.Location"
            lookup-key="_id"
            label="Location "
            class="w-[200px]"
            use-list
            :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=LOC`"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refPICA"
            v-model="data.search.PICA"
            lookup-key="_id"
            label="PICA"
            class="w-[200px]"
            use-list
            :items="['Yes', 'No']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refStatus"
            v-model="data.search.Status"
            lookup-key="_id"
            label="Status"
            class="w-[200px]"
            use-list
            :items="['SUBMITTED', 'READY', 'POSTED', 'REJECTED']"
            @change="refreshData"
          ></s-input>
        </div>
      </template>
      <template #grid_Pica="{ item }">
        <!-- <grid-pica :pica="item.Pica" v-if="item.Pica" /> -->
        {{ item.IsPica ? "Yes" : "No" }}
      </template>
      <template #grid_FollowUp="{ item }">
        {{ item.FollowUp.ResponseDescription }}
      </template>
      <template #grid_DetailsFinding="{ item }">
        {{ item.DetailsFinding.DetailsFinding }}
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          :read-only="readOnly || mode == 'view'"
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #form_input_DetailsFinding="{ item, config, mode }"
        >&nbsp;
        <!-- <div class="flex gap-4">
          <div class="basis-1/2">
            <s-input
              :read-only="readOnly || mode == 'view'"
              v-model="item.DetailsFinding.DetailsFinding"
              kind="text"
              :multi-row="3"
              class="mb-2"
              label="Finding Description"
            />
          </div>
          <div class="basis-1/2">
            <uploader
              :read-only="readOnly || mode == 'view'"
              ref="gridAttachmentFinding"
              :journalId="`${item._id}`"
              :config="config"
              :journalType="prefixUpload"
              single-save
              :tags="[`${prefixUpload}_FINDING_${item._id}`]"
              @pre-open="preOpenUploader"
            />
          </div>
        </div> -->
      </template>
      <template #form_input_FollowUp="{ item, config, mode }"
        >&nbsp;
        <div class="grid grid-cols-2 gap-4">
          <!-- <s-input
            :read-only="readOnly || mode == 'view'"
            v-model="item.FollowUp.ResponseDescription"
            kind="text"
            :multi-row="3"
            label="Response Description"
          />
          <uploader
            :read-only="readOnly || mode == 'view'"
            ref="gridAttachmentFollowUp"
            :journalId="`${item._id}`"
            :config="config"
            :journalType="prefixUpload"
            single-save
            :tags="[`${prefixUpload}_FOLLOWUP_${item._id}`]"
          /> -->
          <!-- <s-input
            v-model="item.DetailsFinding.Status"
            label="PICA"
            use-list
            :items="['Open', 'Completed']"
            @change="onChangeStatus"
          /> -->
          <div>
            <label class="input_label"><div>PICA</div></label>
            <s-toggle
              :read-only="readOnly || mode == 'view'"
              v-model="item.DetailsFinding.Status"
              class="w-[120px] mt-0.5"
              yes-label="Yes"
              no-label="No"
              @change="onChangeStatus"
            />
          </div>
        </div>
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #form_input_Pica="{ item }">
        <pica
          v-if="item.DetailsFinding.Status == true"
          v-model="item.Pica"
          :read-only="readOnly || mode == 'view'"
        />
        <div v-else></div>
      </template>
      <template #form_input_ActivityID="{ item, mode }">
        <s-input
          :read-only="!item.CategoryID || readOnly || mode == 'view'"
          label="Activity"
          v-model="item.ActivityID"
          use-list
          :lookup-url="
            '/tenant/masterdata/find?MasterDataTypeID=' + item.CategoryID
          "
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        />
      </template>
      <template #form_tab_Detail_Finding="{ item, mode }">
        <div class="grid grid-cols-2">
          <s-input
            :read-only="readOnly || mode == 'view'"
            v-model="item.DetailsFinding.DetailsFinding"
            kind="text"
            :multi-row="3"
            class="mb-2"
            label="Finding Description"
          />
        </div>

        <attachment
          :read-only="readOnly || mode == 'view'"
          :journal-id="item._id"
          :tags="[`${prefixUploadfinding}_${item._id}`]"
          :journalType="`${prefixUploadfinding}`"
          single-save
          @postSave="refreshAttch"
          @postDelete="refreshAttch"
        />

        <!--  // () => { // console.log(gridAttachmentCtl.value); //
        gridAttachmentCtl.value.refreshGrid(); // }  -->
      </template>
      <template #form_tab_Follow_Up="{ item, mode }">
        <div class="grid grid-cols-2">
          <s-input
            :read-only="readOnly || mode == 'view'"
            v-model="item.FollowUp.ResponseDescription"
            kind="text"
            :multi-row="3"
            label="Response Description"
          />
        </div>
        <attachment
          :read-only="readOnly || mode == 'view'"
          :journal-id="item._id"
          :tags="[`${prefixUploadfollowup}_${item._id}`]"
          :journalType="prefixUploadfollowup"
          single-save
          @postSave="refreshAttch"
          @postDelete="refreshAttch"
        />
      </template>
      <template #form_tab_Attachment="{ item, mode }">
      
        <attachment
          ref="gridAttachmentCtl"
          :read-only="readOnly || mode == 'view'"
          :journal-id="item._id"
          :tags="linesTag"
          :journalType="prefixUpload"
          single-save
        />
      </template>
      <template #form_buttons_1="{ item, inSubmission, loading }">
        <form-buttons-trx
          :disabled="loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="'SAFETYCARD'"
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
import { reactive, ref, inject, watch, computed, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SInput, util, SButton } from "suimjs";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";

import Pica from "@/components/common/ItemPica.vue";
import Uploader from "@/components/common/Uploader.vue";
import GridPica from "./widget/ViewGridPica.vue";
import { authStore } from "@/stores/auth.js";
import Attachment from "@/components/common/SGridAttachment.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import SToggle from "@/components/common/SButtonToggle.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";

import StatusText from "@/components/common/StatusText.vue";

import helper from "@/scripts/helper.js";
layoutStore().name = "tenant";
const FEATUREID = "Safetycard";
const profile = authStore().getRBAC(FEATUREID);
const gridAttachmentCtl = ref(null);
const gridAttachmentFinding = ref(null);
const gridAttachmentFollowUp = ref(null);
const listControl = ref(null);
const axios = inject("axios");

const prefixUpload = "SHE_SAFETYCARD";
const prefixUploadfinding = `FINDING_${prefixUpload}`;
const prefixUploadfollowup = `FOLLOWUP_${prefixUpload}`;
const linesTag = computed({
  get() {
    const tags = [
      `${prefixUpload}_${data.record._id}`,
      `${prefixUploadfinding}_${data.record._id}`,
      `${prefixUploadfollowup}_${data.record._id}`,
    ];
    return tags;
  },
});

function refreshAttch() {
  gridAttachmentCtl.value.refreshGrid();
}
let customFilter = computed(() => {
  const filters = [];
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "Created",
      Op: "$gte",
      Value: helper.formatFilterDate(data.search.DateFrom),
      // moment(data.search.DateFrom).utc().format("YYYY-MM-DDT00:mm:00Z"),
    });
  }
  if (
    data.search.DateTo !== null &&
    data.search.DateTo !== "" &&
    data.search.DateTo !== "Invalid date"
  ) {
    filters.push({
      Field: "Created",
      Op: "$lte",
      Value: helper.formatFilterDate(data.search.DateFrom, true),
    });
  }
  if (data.search.Category !== null && data.search.Category !== "") {
    filters.push({
      Field: "CategoryID",
      Op: "$eq",
      Value: data.search.Category,
    });
  }
  if (data.search.Finding !== null && data.search.Finding !== "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "DetailsFinding.DetailsFinding",
          Op: "$contains",
          Value: [data.search.Finding],
        },
      ],
    });
  }
  if (data.search.Response !== null && data.search.Response !== "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "FollowUp.ResponseDescription",
          Op: "$contains",
          Value: [data.search.Response],
        },
      ],
    });
  }
  if (data.search.Position !== null && data.search.Position !== "") {
    filters.push({
      Field: "PositionID",
      Op: "$eq",
      Value: data.search.Position,
    });
  }

  if (data.search.Location !== null && data.search.Location !== "") {
    filters.push({
      Field: "LocationID",
      Op: "$eq",
      Value: data.search.Location,
    });
  }

  if (data.search.Status !== null && data.search.Status !== "") {
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: data.search.Status,
    });
  }

  if (data.search.PICA !== null && data.search.PICA !== "") {
    if (data.search.PICA == "Yes") {
      filters.push({
        Field: "IsPica",
        Op: "$eq",
        Value: true,
      });
    } else {
      filters.push({
        Field: "IsPica",
        Op: "$ne",
        Value: false,
      });
    }
  }

  if (
    data.search.Site !== undefined &&
    data.search.Site !== null &&
    data.search.Site !== ""
  ) {
    filters.push(
      {
        Field: "Dimension.Key",
        Op: "$eq",
        Value: "Site",
      },
      {
        Field: "Dimension.Value",
        Op: "$eq",
        Value: data.search.Site,
      }
    );
  }

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  record: {},
  search: {
    Finding: "",
    Response: "",
    Category: "",
    Position: "",
    Location: "",
    DateFrom: null,
    DateTo: null,
    PICA: "",
    Status: "",
    Site: "",
  },
  listJurnalType: [],
});

const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
const modelUpload = {
  ID: "",
  Name: "",
};
function trxPreSubmit(status, action, doSubmit) {
  if (["DRAFT"].includes(status)) {
    listControl.value.submitForm(
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
  listControl.value.setControlMode("grid");
  listControl.value.refreshGrid();
}
function trxErrorSubmit() {
  setLoadingForm(false);
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function newRecord(r) {
  openForm(r);
  r.Status = "";
  data.record = r;
  data.record.Status = "";
  data.record.DetailsFinding = {
    DetailsFinding: "",
    Status: false,
    Attachments: [modelUpload],
  };
  if (data.listJurnalType.length > 0) {
    r.JournalTypeID = data.listJurnalType[0]._id;
    r.PostingProfileID = data.listJurnalType[0].PostingProfileID;
  }

  data.record.FollowUp = {
    ResponseDescription: "",
    Attachments: [modelUpload],
  };

  data.record.Pica = {};
}
function editForm(r) {
  openForm(r);
  r.DetailsFinding.Status = r.DetailsFinding.Status == "Open" ? true : false;
  data.record = r;
  listControl.value.setFormRecord(data.record);
  util.nextTickN(2, () => {
    if (readOnly.value) {
      listControl.value.setFormMode("view");
    }
  });
}
function openForm(r) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("DetailsFinding", "hide", true);
  });
}
function onChangeStatus() {
  util.nextTickN(2, () => {
    // if (data.record.DetailsFinding.Status == "Open") {
    if (data.record.DetailsFinding.Status == true) {
      data.record.IsPica = true;
      data.record.Pica = {
        FindingDescription: data.record.DetailsFinding.DetailsFinding,
        Status: "Open",
        DueDate: new Date(),
      };
    } else {
      data.record.IsPica = false;
      data.record.Pica = null;
    }
    console.log(data.record.IsPica);
  });
}
function preOpenUploader() {
  if (!data.record._id) {
    saveManual();
  }
}
function onPreSave(record) {
  console.log(record.IsPica);
  record.Pica = record.IsPica ? record.Pica : null;
  record.DetailsFinding.Status = record.DetailsFinding.Status
    ? "Open"
    : "Completed";
}
function saveManual() {
  listControl.value.setFormLoading(true);
  data.record.DetailsFinding.Status = data.record.DetailsFinding.Status
    ? "Open"
    : "Completed";
  listControl.value.submitForm(
    data.record,
    () => {
      listControl.value.setFormLoading(false);
    },
    () => {
      listControl.value.setFormLoading(false);
    }
  );
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}

function onFormFieldChange(field, v1, v2, old, record) {
  console.log(field);
  switch (field) {
    case "JournalTypeID":
      util.nextTickN(2, () => {
        getJournalType(v1, record);
      });
      break;
    case "CategoryID":
      record.ActivityID = null;
      break;
    default:
      break;
  }
}
function getJournalType(v1, record) {
  if (v1) {
    axios.post("/fico/shejournaltype/get", [v1]).then(
      (r) => {
        record.PostingProfileID = r.data.PostingProfileID;
      },
      (e) => util.showError(e)
    );
  }
}

function getsJournalType() {
  axios.post("/fico/shejournaltype/gets", {}).then(
    (r) => {
      data.listJurnalType = r.data.data;
    },
    (e) => util.showError(e)
  );
}

function alterGridConfig(cfg) {
  cfg.fields.map((f) => {
    if (["DetailsFinding", "FollowUp"].includes(f.field)) {
      f.width = "350px";
    }
  });
}
onMounted(() => {
  getsJournalType();
});
</script>

<style scoped>
.Savety-Card .box-upload {
  @apply border-2 border-dashed border-zinc-200 w-20 h-20 mr-2 rounded-md grid place-content-center relative;
}
</style>
