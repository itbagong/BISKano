<template>
  <div class="w-full">
    <data-list
      class="card grid-line-items"
      ref="listControl"
      :title="data.titleForm"
      :grid-hide-select="true"
      grid-config="/mfg/workorder/journal/type/gridconfig"
      form-config="/mfg/workorder/journal/type/formconfig"
      grid-read="/mfg/workorder/journal/type/gets"
      form-read="/mfg/workorder/journal/type/get"
      grid-mode="grid"
      grid-delete="/mfg/workorder/journal/type/delete"
      form-keep-label
      form-insert="/mfg/workorder/journal/type/save"
      form-update="/mfg/workorder/journal/type/save"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :grid-fields="['Enable']"
      :form-fields="['DefaultOffsiteConsumption', 'DefaultOffsiteManPower']"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @alterGridConfig="onAlterGridConfig"
    >
      <template #form_input_DefaultOffsiteConsumption="{ item }">
        <div class="title section_title">Accrual</div>
        <AccountSelector
          v-model="item.DefaultOffsiteConsumption"
          row
        ></AccountSelector>
      </template>
      <template #form_input_DefaultOffsiteManPower="{ item }">
        <div class="title section_title">Default Offsite Manpower</div>
        <AccountSelector
          v-model="item.DefaultOffsiteManPower"
          row
        ></AccountSelector>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput } from "suimjs";
import AccountSelector from "@/components/common/AccountSelector.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventory from "@/components/common/DimensionInventory.vue";
import { authStore } from "@/stores/auth";

layoutStore().name = "tenant";
const featureID = "WorkOrderJournalType";
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
  titleForm: "Work Order Journal Type",
  record: {
    _id: "",
    DefaultOffset: {
      AccountID: "",
      AccountType: "",
    },
  },
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Work Order Journal Type`;
  record._id = "";
  record.TrxType = "Work Order";
  record.DefaultOffset = {
    AccountID: "",
    AccountType: "",
  };
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Work Order | ${record._id}`;
  data.record = record;
  openForm(record);
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
    data.record = record;
  });
}
function onAlterGridConfig(cfg) {
  util.nextTickN(2, () => {
    cfg.sortable = ["Created", "_id"];
    cfg.setting.idField = "Created";
    cfg.setting.sortable = ["Created", "_id"];
  });
}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Work Order Journal Type";
  }
}
onMounted(() => {});
</script>
