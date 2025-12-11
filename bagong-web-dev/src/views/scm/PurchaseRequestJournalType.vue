<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      :grid-hide-select="true"
      grid-config="/scm/purchase/request/journal/type/gridconfig"
      form-config="/scm/purchase/request/journal/type/formconfig"
      grid-read="/scm/purchase/request/journal/type/gets"
      form-read="/scm/purchase/request/journal/type/get"
      grid-mode="grid"
      grid-delete="/scm/purchase/request/journal/type/delete"
      form-keep-label
      form-insert="/scm/purchase/request/journal/type/save"
      form-update="/scm/purchase/request/journal/type/save"
      :grid-fields="['Enable']"
      :form-fields="[]"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
    >
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventory from "@/components/common/DimensionInventory.vue";
import { authStore } from "@/stores/auth";

layoutStore().name = "tenant";
const featureID = "PurchaseRequestJournalType";
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
  titleForm: "Purchase Request Journal Type",
  record: {
    _id: "",
  },
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Purchase Request Journal Type`;
  record._id = "";
  record.TrxType = "Purchase Request";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Purchase Request | ${record._id}`;
  data.record = record;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Purchase Request Journal Type";
  }
}
onMounted(() => {});
</script>
