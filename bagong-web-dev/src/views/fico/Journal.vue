<template>
    <div class="w-full">
      <data-list class="card" ref="listControl" title="General Journal"
        grid-config="/tenant/ledger/journal/gridconfig" form-config="/tenant/ledger/journal/formconfig" grid-read="/tenant/ledger/journal/gets"
        form-read="/tenant/ledger/journal/get" grid-mode="grid" grid-delete="/tenant/ledger/journal/delete" form-keep-label
        form-insert="/tenant/ledger/journal/insert" form-update="/tenant/ledger/journal/update" :grid-fields="['Enable']"
        :form-tabs-edit="['General','Setup','Lines']"
        :init-app-mode="data.appMode" :init-form-mode="data.formMode" @formNewData="newRecord" @formEditData="openForm">
        <template #form_tab_Feature="{ item }">
        </template>
      </data-list>
    </div>
  </template>
  
  <script setup>
  import { reactive, ref } from "vue";
  import { layoutStore } from "@/stores/layout.js";
  import { DataList, util } from "suimjs";
  
  layoutStore().name = "clear";
  
  const listControl = ref(null);
  
  const data = reactive({
    appMode: "grid",
    formMode: "edit",
  });
  
  function newRecord(record) {
    record._id = "";
    record.Name = "";
    record.Enable = true;
  
    openForm(record)
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
              "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz".indexOf(ch) >= 0;
            if (!validCar) consistsInvalidChar = true;
            //console.log(ch,vLen,validCar)
          });
  
          if (vLen < 3 || consistsInvalidChar)
            return "minimal length is 3 and alphabet only";
          return "";
        },
      ]);
    })
  }
  </script>