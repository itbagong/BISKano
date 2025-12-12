<template>
  <div class="w-full">
    <data-list class="card" ref="listControl" title="Feature Category"
      grid-config="/admin/featurecategory/gridconfig" form-config="/admin/featurecategory/formconfig"
      grid-read="/admin/featurecategory/gets" form-read="/admin/featurecategory/get" grid-mode="grid"
      grid-delete="/admin/featurecategory/delete" form-insert="/admin/featurecategory/insert"
      form-update="/admin/featurecategory/update" :grid-fields="['Enable']" :init-app-mode="data.appMode"
      :init-form-mode="data.formMode" @formNewData="newRecord" @formEditData="openForm">
      <template #grid_Enable="{ item }">
        <mdicon v-if="item.Enable" class="text-primary" size="16" name="check-bold" />
        <mdicon v-else class="text-error" size="16" name="close-thick" />
      </template>
      <template #form_tab_featurecategory="{ item }">
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, nextTick } from "vue";
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
  record.Enable = true;

  openForm(record)
}

function openForm() {
  util.nextTickN(2, () => {
    //nextTick(() => {
      listControl.value.setFormFieldAttr("_id", "rules", [
        (v) => {
          const errorStr = "minimal length is 5 and alphabet only"
          if (v==undefined) return errorStr
          let vLen = 0;
          let consistsInvalidChar = false;

          v.split("").forEach((ch) => {
            vLen++;
            const validCar =
              "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789".indexOf(ch) >= 0;
            if (!validCar) consistsInvalidChar = true;
          });

          if (vLen < 5 || consistsInvalidChar)
            return errorStr;
          return "";
        },
      ])
    //})
  })
}
</script>