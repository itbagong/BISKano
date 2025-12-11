<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      :grid-hide-select="true"
      grid-config="/scm/purchase/order/journal/type/gridconfig"
      form-config="/scm/purchase/order/journal/type/formconfig"
      grid-read="/scm/purchase/order/journal/type/gets"
      form-read="/scm/purchase/order/journal/type/get"
      grid-mode="grid"
      grid-delete="/scm/purchase/order/journal/type/delete"
      form-keep-label
      form-insert="/scm/purchase/order/journal/type/save"
      form-update="/scm/purchase/order/journal/type/save"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-fields="['Enable']"
      :form-fields="[]"
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
const featureID = "PurchaseOrderJournalType";
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
  titleForm: "Purchase Order Journal Type",
  record: {
    _id: "",
  },
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Purchase Order Journal Type`;
  record._id = "";
  record.TrxType = "Purchase Order";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Purchase Order | ${record._id}`;
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
    data.titleForm = "Purchase Order Journal Type";
  }
}
onMounted(() => {});
</script>
