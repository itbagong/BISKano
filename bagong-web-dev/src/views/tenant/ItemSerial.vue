<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/tenant/itemserial/gridconfig"
      form-config="/tenant/itemserial/formconfig"
      grid-read="/tenant/itemserial/gets"
      form-read="/tenant/itemserial/get"
      grid-mode="grid"
      grid-delete="/tenant/itemserial/delete"
      form-keep-label
      form-insert="/tenant/itemserial/save"
      form-update="/tenant/itemserial/save"
      :form-fields="['ItemID']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @preSave="onPreSave"
      @postSave="onPostSave"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_ItemID="{ item, config }">
        <s-input-sku-item
          ref="refItemVarian"
          label="Item Varian"
          v-model="item.ItemVarian"
          :record="item"
          :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
            item.ItemID,
            item.SKU
          )}`"
        ></s-input-sku-item>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import SInputSkuItem from "../scm/widget/SInputSkuItem.vue";
import helper from "@/scripts/helper.js";
layoutStore().name = "tenant";

const FEATUREID = "ItemSerial";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Item Serial",
});

function newRecord(record) {
  data.titleForm = `Create new Item Serial`;
  record._id = "";
  record.Name = "";
  record.ItemVarian = "";
  record.IsActive = true;
  openForm(record);
}
function editRecord(record) {
  data.titleForm = `Edit Item Serial | ${record._id}`;
  record.ItemVarian = helper.ItemVarian(record.ItemID, record.SKU);
}
function openForm() {
  util.nextTickN(2, () => {});
}
function onPreSave() {
  util.nextTickN(2, () => {});
}
function onPostSave() {
  util.nextTickN(2, () => {});
}
</script>
