<template>
  <data-list
    class="card w-full SHE_AUDIT"
    ref="listControl"
    title="Audit"
    grid-config="/she/audit/gridconfig"
    form-config="/she/audit/formconfig"
    grid-read="/she/audit/gets"
    form-read="/she/audit/get"
    grid-mode="grid"
    grid-delete="/she/audit/delete"
    form-insert="/she/audit/save"
    form-update="/she/audit/save"
    :init-app-mode="data.appMode"
    :form-tabs-edit="['General', 'Line']"
    :form-tabs-new="['General', 'Line']"
    :form-fields="['Dimension']"
    @formFieldChange="onFormFieldChange"
    @formEditData="editRecord"
    @formNewData="newData"
    form-keep-label
    stay-on-form-after-save
    :grid-custom-filter="customFilter"
  >
    <template #grid_header_search="{ config }">
      <div class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter">
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
          ref="refNoDoc"
          v-model="data.search.No"
          label="No Doc"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refName"
          v-model="data.search.Name"
          label="Name"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refCategory"
          v-model="data.search.Category"
          label="Category"
          class="w-[400px]"
          use-list
          :items="['SMK3', 'SMKPAU', 'SMKP']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refStatus"
          v-model="data.search.Status"
          label="Status"
          class="w-[200px]"
          use-list
          :items="['DRAFT', 'SUBMITTED', 'READY', 'POSTED', 'REJECTED']"
          @change="refreshData"
        ></s-input>
      </div>
    </template>
    <template #form_buttons_1="{ item, inSubmission, loading }">
      <form-buttons-trx
        :disabled="loading"
        :status="item.Status"
        :journal-id="item._id"
        :posting-profile-id="item.PostingProfileID"
        :journal-type-id="'audit'"
        :moduleid="'SHE'"
        @preSubmit="trxPreSubmit"
        @postSubmit="trxPostSubmit"
        @errorSubmit="trxErrorSubmit"
        :auto-post="!waitTrxSubmit"
      />
    </template>
    <template #form_input_Dimension="{ item }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :default-list="profile.Dimension"
      ></dimension-editor-vertical>
    </template>
    <template #form_tab_Line="{ item }">
      <Smk3
        :template-id="item.TemplateID"
        v-model="item.SMK3"
        v-if="item.Category == 'SMK3' && item.TemplateID !== ''"
        :formatNumDecimal="formatNumDecimal"
      />

      <Smkp
        :template-id="item.TemplateID"
        v-model="item.SMKP"
        v-if="item.Category == 'SMKP' && item.TemplateID !== ''"
        :formatNumDecimal="formatNumDecimal"
      />

      <Smkpau
        :template-id="item.TemplateID"
        v-model="item.SMKPAU"
        v-if="item.Category == 'SMKPAU' && item.TemplateID !== ''"
        :formatNumDecimal="formatNumDecimal"
        :jurnal-id="item._id"
      />
    </template>
  </data-list>
</template>
<script setup>
import { reactive, ref, watch, inject, onMounted, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton, SGrid } from "suimjs";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import helper from "@/scripts/helper.js";
import moment from "moment";
import Pica from "@/components/common/ItemPica.vue";
import { authStore } from "@/stores/auth.js";
import Smk3 from "./widget/Audit/AuditLinesSMK3.vue";
import Smkp from "./widget/Audit/AuditLinesSMKP.vue";
import Smkpau from "./widget/Audit/AuditLinesSMKPOU.vue";

const axios = inject("axios");
const listControl = ref(null);
const profile = authStore().getRBAC(FEATUREID);
layoutStore().name = "tenant";
const FEATUREID = "audit";
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
      Value: moment(data.search.DateFrom).utc().format("YYYY-MM-DDT00:mm:00Z"),
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
      Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
    });
  }
  if (data.search.No !== null && data.search.No !== "") {
    filters.push({
      Field: "_id",
      Op: "$contains",
      Value: [data.search.No],
    });
  }
  if (data.search.Name !== null && data.search.Name !== "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "Name",
          Op: "$contains",
          Value: [data.search.Name],
        },
      ],
    });
  }
  if (data.search.Category !== null && data.search.Category !== "") {
    filters.push({
      Field: "Category",
      Op: "$eq",
      Value: data.search.Category,
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
  record: {},
  search: {
    DateFrom: null,
    DateTo: null,
    No: "",
    Name: "",
    Category: "",
    Status: "",
  },
});

function newData(r) {
  openForm(r);
}

function editRecord(r) {
  openForm(r);
}

function openForm(r) {
  data.record = r;
}

function onFormFieldChange(name, v1, v2, old) {
  if (name == "Category") {
    data.record.TemplateID = "";
    const tmpLines = ["SMK3", "SMKP", "SMKPAU"];
    for (let i in tmpLines) {
      let o = tmpLines[i];
      if (o !== v1) data.record[o] = [];
    }
  }
}

function formatNumDecimal(params) {
  let splited = params.toFixed(2).split(".");
  let res = "";
  if (splited.includes("00") && splited[splited.length - 1] == "00") {
    res = splited[0];
  } else {
    res = params.toFixed(2);
  }
  return res;
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
</script>

<style>
.SHE_AUDIT .form_inputs > div.flex.section_group_container > div:nth-child(1) {
  width: 75%;
}

.SHE_AUDIT .form_inputs > div.flex.section_group_container > div:nth-child(2) {
  width: 25%;
}
</style>
