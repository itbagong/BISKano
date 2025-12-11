<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Fixed Asset Number"
      grid-config="/fico/fixedassetnumber/gridconfig"
      form-config="/fico/fixedassetnumber/formconfig"
      grid-read="/fico/fixedassetnumber/gets"
      form-read="/fico/fixedassetnumber/get"
      grid-mode="grid"
      grid-delete="/fico/fixedassetnumber/delete"
      form-update="/fico/fixedassetnumber/save"
      form-insert="/fico/fixedassetnumber/insert"
      :grid-fields="['Enable']"
      :form-fields="['Details', 'Dimension']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      form-hide-submit
    >
      <template #form_input_Details="{ item, mode }">
        <FixedAssetNumberDetail
          v-model="item.Details"
          @calcTotalAsset="calcTotalAsset"
          :read-only="mode == 'view'"
        ></FixedAssetNumberDetail>
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          :default-list="profile.Dimension"
          v-model="item.Dimension"
          :read-only="mode == 'view'"
        ></dimension-editor-vertical>
      </template>
      <template #form_buttons_1>
        <s-button
          icon="content-save"
          class="btn_primary submit_btn"
          label="Save"
          @click="onSaveForm"
        />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SButton } from "suimjs";
import FixedAssetNumberDetail from "./widget/FixedAssetNumberDetail.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";

layoutStore().name = "tenant";

const FEATUREID = "FixedAssetNumber";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  record: {},
  oldrecord: {},
});
const axios = inject("axios");

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.TotalAssetNumber = 0;
  record.Details = [];
  data.record = record;

  openForm(record);
}

function editRecord(record) {
  data.record = record;
  data.oldrecord = record;
  openForm(record);
}

function calcTotalAsset(totalAsset) {
  data.record.TotalAssetNumber = parseInt(totalAsset);
}

function onPostSave(record) {
  // save fixed asset number list
  const url = "/fico/fixedassetnumber/save-fixed-asset-number-list";
  let param = record;
  //alert(JSON.stringify(param))

  axios.post(url, param).then(
    (r) => {},
    (e) => {
      util.showError(e);
    }
  );
}

function openForm() {
  util.nextTickN(2, () => {});
}

function validateDetail(dt) {
  let resName = dt.filter((dt) => dt.AssetName == "" || dt.Name == "");
  if (resName.length > 0) {
    util.showError(`Asset name can't be empty`);
    return true
  }
  let resNumber = dt.filter((dt) => dt.NumberAsset === 0);
  if (resNumber.length > 0) {
    util.showError(`Number asset can't be empty`);
    return true
  }

  return false;
}

function onSaveForm() {
  if (validateDetail(data.record.Details)) {
    return;
  }

  listControl.value.submitForm(
    data.record,
    () => {},
    () => {}
  );
}
</script>
