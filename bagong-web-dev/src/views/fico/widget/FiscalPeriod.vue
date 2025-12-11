<template>
  <data-list ref="listControl" no-gap gridHideSearch gridHideSelect gridHideSort 
    gridHideRefresh grid-config="/fico/fiscalperiod/gridconfig" form-config="/fico/fiscalperiod/formconfig"
    :grid-read="'/fico/fiscalperiod/gets?FiscalYearID=' + param._id" form-read="/fico/fiscalperiod/get" grid-mode="grid"
    grid-delete="/fico/fiscalperiod/delete" form-insert="/fico/fiscalperiod/insert"
    form-update="/fico/fiscalperiod/update" :init-app-mode="data.appMode" :init-form-mode="data.formMode"
    @formNewData="newRecord" @formEditData="openForm" @preSave="onPreSave" form-focus>
  </data-list>
</template>
  
<script setup>
import { reactive, ref } from "vue";
import { DataList, util } from "suimjs";

const props = defineProps({
  param: { type: Object, default: () => {} },
});

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.FromDate = new Date();
  record.ToDate = new Date();
  record.Modules = {};
  record.Active = true;
  record.FiscalYearID = props.param._id;

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
  });

  record.Modules = record.Modules.Finance ?? "";
}

function onPreSave(record) {
  record.Modules = { Finance: record.Modules, Inventory: record.Modules };
}
</script>