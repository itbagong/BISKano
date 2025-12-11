<template>
  <div class="w-full">
    <data-list
      class="card grid-line-items"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/mfg/physical/availability/gridconfig"
      form-config="/mfg/physical/availability/formconfig"
      grid-read="/mfg/physical/availability/gets"
      form-read="/mfg/physical/availability/get"
      grid-mode="grid"
      grid-delete="/mfg/physical/availability/delete"
      form-keep-label
      form-insert="/mfg/physical/availability/save"
      form-update="/mfg/physical/availability/save"
      :grid-fields="['Month', 'Year']"
      :form-fields="['Month', 'Year']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      grid-hide-select
      @controlModeChanged="onControlModeChanged"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @pre-save="onPreSave"
      @post-save="onPostSave"
      @alterGridConfig="onAlterGridConfig"
    >
      <template #grid_Month="{ item }">
        <div class="text-right">{{ data.listMonth[item.Month] }}</div>
      </template>
      <template #grid_Year="{ item }">
        <div class="text-right">{{ item.Year }}</div>
      </template>
      <template #form_input_Month="{ item }">
        <s-input
          ref="refMonth"
          label="Month"
          v-model="item.Month"
          class="w-full"
          use-list
          :items="data.listMonth"
        ></s-input>
      </template>
      <template #form_input_Year="{ item }">
        <s-input
          ref="refYear"
          label="Year"
          v-model="item.Year"
          class="w-full"
          use-list
          :items="data.listYear"
        ></s-input>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventory from "@/components/common/DimensionInventory.vue";
import { authStore } from "@/stores/auth";

layoutStore().name = "tenant";
const featureID = "PhysicalAvailability";
const profile = authStore().getRBAC(featureID);
const listControl = ref(null);
const axios = inject("axios");
const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Physical Availability",
  record: {
    _id: "",
  },
  listMonth: [
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December",
  ],
  listYear: [],
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Physical Availability`;
  record._id = "";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Physical Availability | ${record._id}`;
  record.Month = data.listMonth[record.Month];
  data.record = record;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}
function onAlterGridConfig(cfg) {
  util.nextTickN(2, () => {
    cfg.sortable = ["Created", "TrxDate", "_id"];
    cfg.setting.idField = "Created";
    cfg.setting.sortable = ["Created", "TrxDate", "_id"];
  });
}
function getYears() {
  var currentYear = new Date().getFullYear();
  var yearsToGoBack = 10;
  var yearsArray = [];
  for (var i = 0; i < yearsToGoBack; i++) {
    yearsArray.push(currentYear - i);
  }
  data.listYear = yearsArray;
}

function onPreSave(record) {
  record.Month = data.listMonth.indexOf(record.Month);
}
function onPostSave(record) {
  let payload = {
    IDs: [record._id],
  };
  axios.post("/mfg/physical/availability/calculate", payload).then(
    (r) => {},
    (e) => {
      return util.showError(e);
    }
  );
}

function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Physical Availability";
  }
}

onMounted(() => {
  getYears();
});
</script>
