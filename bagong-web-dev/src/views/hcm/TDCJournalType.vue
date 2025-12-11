<template lang="">
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="TDC Journal Type"
      grid-config="/hcm/tdcjournaltype/gridconfig"
      form-config="/hcm/tdcjournaltype/formconfig"
      grid-read="/hcm/tdcjournaltype/gets"
      form-read="/hcm/tdcjournaltype/get"
      grid-mode="grid"
      grid-delete="/hcm/tdcjournaltype/delete"
      form-keep-label
      form-insert="/hcm/tdcjournaltype/save"
      form-update="/hcm/tdcjournaltype/save"
      :grid-fields="['Enable']"
      :form-fields="['Dimension']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      @postSave="saveConfig"
    >
      <template #form_input_Dimension="{ item }">
        <dimension-editor v-model="item.Dimension"></dimension-editor>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import { authStore } from "@/stores/auth.js";

import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";

layoutStore().name = "tenant";

const FEATUREID = "HCMtdcjournaltype";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const journalConfig = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Enable = true;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", [
      (v) => {
        let vLen = 0;
        let consistsInvalidChar = false;

        v.split("").forEach((ch) => {
          vLen++;
          const validCar =
            "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
              ch
            ) >= 0;
          if (!validCar) consistsInvalidChar = true;
          //console.log(ch,vLen,validCar)
        });

        if (vLen < 3 || consistsInvalidChar)
          return "minimal length is 3 and alphabet only";
        return "";
      },
    ]);
  });
}

function saveConfig() {
  if (journalConfig.value && journalConfig.value.getDataValue()) {
    let dv = journalConfig.value.getDataValue();
    axios.post("/bagong/vendorjournaltypeconfiguration/save", dv).then(
      (r) => {},
      (e) => {
        data.loading = false;
      }
    );
  }
}
</script>
<style lang=""></style>
