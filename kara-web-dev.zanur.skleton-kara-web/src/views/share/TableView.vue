<template>
    <div class="w-full">
      <data-list class="card" ref="listControl" 
        v-if="data.gridConfig!=''" form-keep-label
        :title="data.title" :grid-hide-delete="!data.allowDelete"
        :grid-config="data.gridConfig" :grid-read="data.gridRead" :grid-delete="data.allowDelete ? data.gridDelete : ''"
        :form-config="data.formConfig" 
        :form-read="data.formRead" grid-mode="grid" 
        :form-insert="data.formInsert" :form-update="data.formUpdate" :grid-fields="['Enable']"
        :init-app-mode="data.appMode" :init-form-mode="data.formMode" @formNewData="newRecord" @formEditData="openForm">
        <template #grid_Enable="{ item }">
          <mdicon v-if="item.Enable" class="text-primary" size="16" name="check-bold" />
          <mdicon v-else class="text-error" size="16" name="close-thick" />
        </template>
        <template #form_tab_Feature="{ item }">
        </template>
      </data-list>
    </div>
  </template>
  
  <script setup>
  import { reactive, ref, watch } from "vue";
  import { layoutStore } from "@/stores/layout.js";
  import { DataList, util } from "suimjs";
  import { useRoute } from "vue-router";
  
  layoutStore().name = "tenant";
  
  const listControl = ref(null);
  
  const route = useRoute()
  
  const data = reactive({
    appMode: "grid",
    formMode: "edit",
    title: route.query.title || route.query.objname,
    gridConfig: route.query.objname+"/gridconfig",
    gridRead: route.query.objname+"/gets",
    gridDelete: route.query.objname+"/delete",
    formConfig: route.query.objname+"/formconfig",
    formRead: route.query.objname+"/get",
    formInsert: route.query.objname+"/insert",
    formUpdate: route.query.objname+"/update",
    allowDelete: route.query.allowDelete==='true'
  });

  watch(() => route.query.objname, nv => {
    data.gridConfig = route.query.objname+"/gridconfig"
    data.gridRead = route.query.objname+"/gets"
    data.gridDelete = route.query.objname+"/delete"
    data.formConfig = route.query.objname+"/formconfig"
    data.formRead =  route.query.objname+"/get"
    data.formInsert = route.query.objname+"/insert"
    data.formUpdate =  route.query.objname+"/update"

    util.nextTickN(2, () => {
        listControl.value.refreshList()
        listControl.value.refreshForm()
    })
  })

  watch(() => route.query.title, nv => {
    data.title = nv
  })

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
              "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(ch) >= 0;
            if (!validCar) consistsInvalidChar = true;
            //console.log(ch,vLen,validCar)
          });
  
          if (vLen < 5 || consistsInvalidChar)
            return "minimal length is 5 and alphabet only";
          return "";
        },
      ]);
    })
  }
  </script>