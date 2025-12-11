<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="`Master Data Type`"
      grid-config="/tenant/masterdatatype/gridconfig"
      form-config="/tenant/masterdatatype/formconfig"
      grid-read="/tenant/masterdatatype/gets"
      form-read="/tenant/masterdatatype/get"
      grid-mode="grid"
      grid-delete="/tenant/masterdatatype/delete"
      form-keep-label
      form-insert="/tenant/masterdatatype/save"
      form-update="/tenant/masterdatatype/save"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      @postSave="onPostSave"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";

layoutStore().name = "tenant";

const FEATUREID = "MasterDataType";
const profile = authStore().getRBAC(FEATUREID);

const layout = layoutStore();
const listControl = ref(null);
const data = reactive({
  appMode: "grid",
  formMode: "edit",
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.IsActive = true;
  layout.setAddDataMaster(false);
  openForm(record);
}

function onPostSave() {
  layout.setAddDataMaster(true);
}
function openForm() {
  util.nextTickN(2, () => {
    // listControl.value.setFormFieldAttr("_id", "rules", [
    //   (v) => {
    //     let vLen = 0;
    //     let consistsInvalidChar = false;
    //     v.split("").forEach((ch) => {
    //       vLen++;
    //       const validCar =
    //         "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
    //           ch
    //         ) >= 0;
    //       if (!validCar) consistsInvalidChar = true;
    //       //console.log(ch,vLen,validCar)
    //     });
    //     if (vLen < 2 || consistsInvalidChar)
    //       return "minimal length is 2 and alphabet only";
    //     return "";
    //   },
    // ]);
  });
}
</script>
