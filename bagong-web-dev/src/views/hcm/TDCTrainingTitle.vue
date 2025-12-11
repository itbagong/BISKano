<template>
    <div class="w-full">
      <data-list
        class="card"
        ref="listControl"
        title="TDC Training Title"
        grid-config="/hcm/tdctitle/gridconfig"
        form-config="/hcm/tdctitle/formconfig"
        grid-read="/hcm/tdctitle/gets"
        form-read="/hcm/tdctitle/get"
        grid-mode="grid"
        grid-delete="/hcm/tdctitle/delete"
        form-keep-label
        form-insert="/hcm/tdctitle/save"
        form-update="/hcm/tdctitle/save"
        :init-app-mode="data.appMode"
        :init-form-mode="data.formMode"
        @formNewData="newRecord"
        @formEditData="openForm"
        @postSave="saveConfig"
      >
      </data-list>
    </div>
  </template>

  <script setup>
  import { reactive, ref } from "vue";
  import { layoutStore } from "@/stores/layout.js";
  import { DataList, util } from "suimjs";
  
  layoutStore().name = "tenant";
  
  const listControl = ref(null);
  
  const data = reactive({
    appMode: "grid",
    formMode: "edit",
  });
  
  function newRecord(record) {
    record._id = "";
    record.Name = "";
    record.Alias = "";
    openForm(record);
  }
  
  function openForm() {
    util.nextTickN(2, () => {});
  }
  
  function saveConfig() {}
  </script>
  <style lang=""></style>
  