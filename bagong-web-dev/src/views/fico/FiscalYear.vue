<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Fiscal Year"
      grid-config="/fico/fiscalyear/gridconfig"
      form-config="/fico/fiscalyear/formconfig"
      grid-read="/fico/fiscalyear/gets"
      form-read="/fico/fiscalyear/get"
      grid-mode="grid"
      grid-delete="/fico/fiscalyear/delete"
      form-keep-label
      form-insert="/fico/fiscalyear/save"
      form-update="/fico/fiscalyear/save"
      :grid-fields="['Enable']"
      :form-tabs-edit="['General', 'Period']"
      :form-fields="['Dimension']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_tab_Period="{ item }">
        <FiscalPeriod :param="item" />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import { DataList, util } from "suimjs";
import FiscalPeriod from "./widget/FiscalPeriod.vue";

layoutStore().name = "tenant";

const FEATUREID = "FiscalYear";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.From = new Date();
  record.To = new Date();
  record.CompanyID = authStore().companyID;
  record.IsActive = false;

  openForm(record);
}

function openForm(record) {
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
    listControl.value.setFormFieldAttr("IsActive", "hide", record._id === "");
  });
}
</script>
