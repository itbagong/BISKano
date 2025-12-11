<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      :grid-hide-select="true"
      grid-config="/scm/item/request/journal/type/gridconfig"
      form-config="/scm/item/request/journal/type/formconfig"
      grid-read="/scm/item/request/journal/type/gets"
      form-read="/scm/item/request/journal/type/get"
      grid-mode="grid"
      grid-delete="/scm/item/request/journal/type/delete"
      form-keep-label
      form-insert="/scm/item/request/journal/type/save"
      form-update="/scm/item/request/journal/type/save"
      :grid-fields="['Enable']"
      :form-fields="[]"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import { authStore } from "@/stores/auth.js";

layoutStore().name = "tenant";

const featureID = 'ItemRequestJournalType'

// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'ItemRequestJournalType'})

const profile = authStore().getRBAC(featureID)
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
  titleForm: "Item Request Journal Type",
  record: {
    _id: "",
  },
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Item Request Journal Type`;
  record._id = "";
  record.TrxType = "Item Request";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Item Request | ${record._id}`;
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
    data.titleForm = "Item Request Journal Type";
  }
}
onMounted(() => {});
</script>
