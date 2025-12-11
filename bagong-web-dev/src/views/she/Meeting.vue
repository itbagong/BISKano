<template>
  <div class="w-full Meeting">
    <data-list
      class="card SHE_MEETING"
      ref="listControl"
      title="Meeting"
      grid-config="/she/meeting/gridconfig"
      form-config="/she/meeting/formconfig"
      grid-read="/she/meeting/gets"
      form-read="/she/meeting/get"
      grid-mode="grid"
      grid-delete="/she/meeting/delete"
      form-keep-label
      form-insert="/she/meeting/save"
      form-update="/she/meeting/save"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      gridHideSelect
      @form-new-data="newRecord"
      @form-edit-data="openForm"
      :form-tabs-edit="['General', 'Results', 'Attachment']"
      :form-tabs-view="['General', 'Results', 'Attachment']"
      :form-fields="['Dimension', 'Attachments', 'MeetingDate']"
      :grid-fields="['Result']"
      :grid-custom-filter="customFilter"
      @pre-save="onPreSave"
      @postSave="onPostSave"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      stay-on-form-after-save
    >
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
            ref="refCoach"
            v-model="data.search.MeetingType"
            class="w-[300px]"
            use-list
            label="Meeting Type"
            lookup-url="/tenant/masterdata/find?MasterDataTypeID=MTY"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refCoach"
            v-model="data.search.Title"
            class="w-[300px]"
            label="Meeting Title"
            @keyup.enter="refreshData"
          ></s-input>
          <s-input
            ref="refStatus"
            v-model="data.search.Status"
            lookup-key="_id"
            label="Status"
            class="w-[200px]"
            use-list
            :items="['Open', 'Completed']"
            @change="refreshData"
          ></s-input>
        </div>
      </template>
      <template #grid_Result="{ item }">
        <div class="" v-for="(dt, idx) in item.Result" :key="idx">
          <grid-pica :pica="dt.Pica" hide-status />
        </div>
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #form_input_Dimension="{ item, mode }">
        {{ readOnly }}
        <dimension-editor-vertical
          :read-only="readOnly || mode == 'view'"
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #form_input_MeetingDate="{ item }">
        <label class="input_label">
          <div>Meeting date</div>
        </label>
        <div class="flex gap-2" v-if="mode == 'edit'">
          {{ moment(item.MeetingDate).format("DD MMM YYYY HH:mm") }}
        </div>
        <div class="flex gap-2" v-else>
          <s-input kind="date" v-model="data.record.bookDate" />
          <s-input kind="time" v-model="data.record.bookTime" />
        </div>
      </template>
      <template #form_input_Attachments="{ item, config, mode }">
        <div class="flex gap-4">
          <div>
            <div class="hidden">{{ (config.label = "Meeting") }}</div>
            <uploader
              :read-only="readOnly || mode == 'view'"
              ref="gridAttachment"
              :journalId="`MEETING_${item._id}`"
              :config="config"
              :journalType="prefixUpload"
              :tags="[`${prefixUpload}_MEETING_${item._id}`]"
              single-save
              @pre-open="preOpenUploader"
            />
          </div>
          <div>
            <uploader
              :read-only="readOnly || mode == 'view'"
              ref="gridAttachment"
              :journalId="`MOM_${item._id}`"
              :config="{ label: 'MoM' }"
              :journalType="prefixUpload"
              :tags="[`${prefixUpload}_MOM_${item._id}`]"
              single-save
              @pre-open="preOpenUploader"
            />
          </div>
        </div>
      </template>
      <template #form_tab_Results="{ item, mode }">
        <result
          :jurnalId="item._id"
          v-model="item.Result"
          @close="onSave"
          :read-only="readOnly || mode == 'view'"
        />
      </template>
      <template #form_tab_Attachment="{ item, mode }">
        <attachment
          :read-only="readOnly || mode == 'view'"
          :journal-id="item._id"
          :tags="linesTag"
          :journalType="prefixUpload"
          single-save
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SInput, util, SButton } from "suimjs";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import moment from "moment";
import Toggle from "@/components/common/SButtonToggle.vue";
import GridPica from "./widget/ViewGridPica.vue";
import { authStore } from "@/stores/auth.js";
import Uploader from "@/components/common/Uploader.vue";
import Result from "./widget/MeetingResult.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import LogTrx from "@/components/common/LogTrx.vue";

import helper from "@/scripts/helper.js";
const FEATUREID = "Meeting";
const profile = authStore().getRBAC(FEATUREID);
const prefixUpload = "SHE_MEETING";

const listControl = ref(null);
const gridAttachment = ref(null);
const linesTag = computed({
  get() {
    const tags = [
      `${prefixUpload}_${data.record._id}`,
      `MOM_${data.record._id}`,
      `MEETING_${data.record._id}`,
    ];
    return tags;
  },
});
layoutStore().name = "tenant";
let customFilter = computed(() => {
  const filters = [];
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "MeetingDate",
      Op: "$gte",
      Value: helper.formatFilterDate(data.search.DateFrom),
    });
  }
  if (
    data.search.DateTo !== null &&
    data.search.DateTo !== "" &&
    data.search.DateTo !== "Invalid date"
  ) {
    filters.push({
      Field: "MeetingDate",
      Op: "$lte",
      Value: helper.formatFilterDate(data.search.DateTo, true),
    });
  }

  if (data.search.MeetingType !== null && data.search.MeetingType !== "") {
    filters.push({
      Field: "MeetingType",
      Op: "$eq",
      Value: data.search.MeetingType,
    });
  }
  if (data.search.Title !== null && data.search.Title !== "") {
    filters.push({
      Field: "Title",
      Op: "$contains",
      Value: [data.search.Title],
    });
  }
  if (data.search.Status !== null && data.search.Status !== "") {
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: data.search.Status,
    });
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
    DateFrom: null,
    DateTo: null,
    Status: "",
    MeetingType: "",
    Title: "",
  },
});
const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record?.Status);
  },
});
function resetBookDate() {
  data.record.bookDate = new Date();
  data.record.bookTime = new Date();
}

const modelUpload = {
  ID: "",
  Name: "",
};

function newRecord(r) {
  resetBookDate();
  r.Status = "";
  r.Attachments = [modelUpload];
  r.MeetingDate = new Date();
  data.record = r;
  // openForm(r);
}

function openForm(r) {
  r.bookDate = moment(r.Date);
  r.bookTime = moment(r.Date).format("HH:mm");
  data.record = r;

  util.nextTickN(2, () => {
    if (readOnly.value) {
      listControl.value.setFormMode("view");
    }
  });
}

const mode = computed({
  get() {
    return listControl.value.getFormMode();
  },
});

function onPreSave(r) {
  let bookDate = moment(data.record.bookDate).format("YYYY-MM-DD");
  let bookTime = data.record.bookTime
    ? data.record.bookTime + ":00"
    : moment(data.record.bookDate).format("HH:mm") + ":00";
  let booking = moment(bookDate + " " + bookTime).format();
  r.MeetingDate = booking;
}

function onPostSave(r) {
  // gridAttachment.value.Save(r._id, "SHE_MEETING");

  if (gridAttachment.value) gridAttachment.value.Save();
  data.record = r;
}

function onSave(val) {
  if (!val) {
    listControl.value.submitForm(
      data.record,
      () => {},
      () => {},
      true
    );
  }
}
function preOpenUploader() {
  if (!data.record._id) {
    saveManual();
  }
}
function saveManual() {
  listControl.value.setFormLoading(true);
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
</script>
<style>
.SHE_MEETING
  .form_inputs
  > div.flex.section_group_container
  > div:nth-child(1) {
  width: 75%;
}

.SHE_MEETING
  .form_inputs
  > div.flex.section_group_container
  > div:nth-child(2) {
  width: 25%;
}
</style>
