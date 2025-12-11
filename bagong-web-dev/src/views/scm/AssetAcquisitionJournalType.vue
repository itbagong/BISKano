<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      :grid-hide-select="true"
      grid-config="/scm/asset-acquisition/journal/type/gridconfig"
      form-config="/scm/asset-acquisition/journal/type/formconfig"
      grid-read="/scm/asset-acquisition/journal/type/gets"
      form-read="/scm/asset-acquisition/journal/type/get"
      grid-mode="grid"
      grid-delete="/scm/asset-acquisition/journal/type/delete"
      form-keep-label
      form-insert="/scm/asset-acquisition/journal/type/save"
      form-update="/scm/asset-acquisition/journal/type/save"
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

layoutStore().name = "tenant";
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
  titleForm: "Asset Acquisition Journal Type",
  record: {
    _id: "",
  },
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Asset Acquisition Journal Type`;
  record._id = "";
  record.TrxType = "Asset Acquisition";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Asset Acquisition | ${record._id}`;
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
    data.titleForm = "Asset Acquisition Journal Type";
  }
}
onMounted(() => {});
</script>
