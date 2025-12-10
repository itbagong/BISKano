<template>
  <div class="w-full">
    <data-list class="card" ref="listControl" title="Tenant"
      grid-config="/admin/tenant/gridconfig" form-config="/admin/tenant/formconfig" grid-read="/admin/tenant/gets"
      form-read="/admin/tenant/get" grid-mode="grid" grid-delete="/admin/tenant/delete"
      form-insert="/admin/tenant/insert" form-update="/admin/tenant/update"
      form-keep-label
      :grid-fields="['Dimensions']"
      :init-app-mode="data.appMode" :init-form-mode="data.formMode" @formNewData="newRecord" @formEditData="openForm">
      <template #grid_Dimensions="{item}">
        {{ item.Dimensions ? item.Dimensions.join(", ") : "" }}
      </template>
     </data-list>
  </div>
</template>

<script setup>
import { reactive, ref } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import { authStore } from "@/stores/auth.js";
authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})

layoutStore().name = "tenant";

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Dimensions = [];
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

        const minLength = 3
        if (vLen < minLength || consistsInvalidChar)
          return `minimal length is ${minLength} and alphabet only`
        return "";
      },
    ]);
  })
}
</script>