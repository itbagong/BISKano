<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/scm/vendor/pricelist/gridconfig"
      form-config="/scm/vendor/pricelist/formconfig"
      grid-read="/scm/vendor/pricelist/gets"
      form-read="/scm/vendor/pricelist/get"
      grid-mode="grid"
      grid-delete="/scm/vendor/pricelist/delete"
      form-keep-label
      form-insert="/scm/vendor/pricelist/save"
      form-update="/scm/vendor/pricelist/save"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      form-hide-submit
      :grid-fields="['Enable', 'SKU']"
      :form-fields="['ItemID', 'FinancialDimension', 'InventoryDimension']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="fieldchange"
      @formNewData="newRecord"
      @formEditData="editRecord"
    >
      <template #form_input_ItemID="{ item }">
        <s-input-sku-item
          ref="refItemVarian"
          label="Item Varian"
          v-model="item.ItemVarian"
          :record="item"
          :required="true"
          :keepErrorSection="true"
          :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
            item.ItemID,
            item.SKU
          )}`"
        ></s-input-sku-item>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          :disabled="data.loading"
          :icon="`content-save`"
          class="btn_primary submit_btn"
          label="Save"
          @click="onSave(item)"
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";
import SInputSkuItem from "./widget/SInputSkuItem.vue";
import { authStore } from "@/stores/auth";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";
const featureID = "VendorPriceList";
const profile = authStore().getRBAC(featureID);
const listControl = ref(null);
const refItemVarian = ref(null);
const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Vendor Pricelist",
  record: {
    _id: "",
    ItemVarian: "",
  },
  loading: false,
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Vendor Pricelist`;
  record._id = "";
  record.ItemVarian = "";
  openForm(record);
}

function editRecord(record) {
  let ItemVarian = "";
  if (record.ItemID) {
    ItemVarian = helper.ItemVarian(record.ItemID, record.SKU);
  }
  record.ItemVarian = ItemVarian;
  data.formMode = "edit";
  data.titleForm = `Edit Vendor Pricelist | ${record._id}`;
  data.record = record;
  openForm(record);
}

const fieldchange = (name, v1, v2, old, record) => {
  if (name == "ItemID") {
    record.SKU = "";
  }
};

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}
function onSave(record) {
  let valid = true;
  data.loading = true;
  if (!refItemVarian.value.validate() || !listControl.value.formValidate()) {
    valid = false;
  }
  if (valid) {
    listControl.value.submitForm(
      record,
      () => {
        data.loading = false;
      },
      () => {
        data.loading = false;
      }
    );
  } else {
    data.loading = false;
  }
}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Vendor Pricelist";
  }
}
onMounted(() => {});
</script>
