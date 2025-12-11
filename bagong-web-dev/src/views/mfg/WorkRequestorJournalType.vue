<template>
  <div class="w-full">
    <data-list
      class="card grid-line-items"
      ref="listControl"
      :title="data.titleForm"
      :grid-hide-select="true"
      grid-config="/mfg/workrequestor/journal/type/gridconfig"
      form-config="/mfg/workrequestor/journal/type/formconfig"
      grid-read="/mfg/workrequestor/journal/type/gets"
      form-read="/mfg/workrequestor/journal/type/get"
      grid-mode="grid"
      grid-delete="/mfg/workrequestor/journal/type/delete"
      form-keep-label
      form-insert="/mfg/workrequestor/journal/type/save"
      form-update="/mfg/workrequestor/journal/type/save"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-fields="['Enable']"
      :form-fields="[]"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @alterGridConfig="onAlterGridConfig"
    >
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import { authStore } from "@/stores/auth";

layoutStore().name = "tenant";
const featureID = "WorkRequestJournalType";
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
  titleForm: "Work Requestor Journal Type",
  record: {
    _id: "",
  },
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Work Requestor Journal Type`;
  record._id = "";
  record.TrxType = "Work Requestor";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Work Requestor Journal Type | ${record._id}`;
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
    cfg.sortable = ["Created", "_id"];
    cfg.setting.idField = "Created";
    cfg.setting.sortable = ["Created", "_id"];
  });
}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Work Requestor Journal Type";
  }
}
onMounted(() => {});
</script>
