<template>
  <div class="w-full">
    <data-list 
      class="card"
      ref="dataCtl" title="Role" grid-config="/admin/role/gridconfig" form-config="/admin/role/formconfig"
      grid-read="/admin/role/gets" form-read="/admin/role/get" grid-mode="grid" grid-delete="/admin/role/delete"
      form-insert="/admin/role/insert" form-update="/admin/role/update" :grid-fields="['Enable']"
      :init-app-mode="data.appMode" :init-form-mode="data.formMode" :form-tabs-edit="['General', 'Feature']"
      stay-on-form-after-save @formNewData="newRole" @formEditData="formOpen" @alterGridConfig="alterGridConfig">
      <template #grid_Enable="{ item }">
        <mdicon v-if="item.Enable" class="text-primary" size="16" name="check-bold" />
        <mdicon v-else class="text-error" size="16" name="close-thick" />
      </template>
      <template #form_tab_Feature="{ item }">
        <RoleFeature :role="item"></RoleFeature>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, nextTick } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList } from "suimjs";
import RoleFeature from "./widget/RoleFeature.vue";
import { authStore } from "@/stores/auth.js";
authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})

layoutStore().name = "tenant";

const dataCtl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
});

function alterGridConfig (cfg) {
  //cfg.fields = cfg.fields.filter(f => f.field!="TenantName")
}

function newRole(record) {
  record.Name = "";
  record.Enable = true;

  formOpen(record)
}

function formOpen(record) {
  nextTick(() => {
    nextTick(() => {
      //dataCtl.value.setFormFieldAttr("TenantID", "hide", true)
      if (record._id!=undefined && record._id!=null && record._id!="") dataCtl.value.setFormFieldAttr("_id", "readOnly", true)
        else dataCtl.value.setFormFieldAttr("_id", "readOnly", false)
      dataCtl.value.setFormFieldAttr("_id", "rules", [
        (v) => {
          let vLen = 0;
          let consistsInvalidChar = false;

          v.split("").forEach((ch) => {
            vLen++;
            const validCar =
              "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstyuvxxyz_-".indexOf(ch) >= 0;
            if (!validCar) consistsInvalidChar = true;
          });

          if (vLen < 4 || consistsInvalidChar)
            return "minimal length is 4 and alphabet only";
          return "";
        },
      ]);
    })
  })
}
</script>