<template>
  <div class="w-full">
    <data-list class="card" ref="listControl" title="Feature"
      grid-config="/admin/feature/gridconfig" form-config="/admin/feature/formconfig" grid-read="/admin/feature/gets"
      form-read="/admin/feature/get" grid-mode="grid" grid-delete="/admin/feature/delete" form-keep-label
      form-insert="/admin/feature/insert" form-update="/admin/feature/update" :grid-fields="['Enable']"
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
import { reactive, ref } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import { authStore } from "../../stores/auth";

layoutStore().name = "tenant";
authStore().hasAccess({IsMemberOf:'Administrators'})

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

        if (vLen < 5 || consistsInvalidChar)
          return "minimal length is 5 and alphabet only";
        return "";
      },
    ]);
  })
}
</script>