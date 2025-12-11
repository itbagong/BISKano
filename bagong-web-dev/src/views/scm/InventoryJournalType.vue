<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      :grid-hide-select="true"
      grid-config="/scm/inventory/journal/type/gridconfig"
      form-config="/scm/inventory/journal/type/formconfig"
      grid-read="/scm/inventory/journal/type/gets"
      form-read="/scm/inventory/journal/type/get"
      grid-mode="grid"
      grid-delete="/scm/inventory/journal/type/delete"
      form-keep-label
      form-insert="/scm/inventory/journal/type/save"
      form-update="/scm/inventory/journal/type/save"
      :grid-fields="['Enable']"
      :form-fields="[
        'Actions',
        'Previews',
        'DefaultOffset',
        'InventoryDimension',
        'Dimension',
      ]"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @preSave="onPreSave"
      @postSave="saveConfig"
      @controlModeChanged="onControlModeChanged"
    >
      <template #form_input_Actions="{ item }">
        <JournalTypeContext
          ref="listAction"
          title="Action"
          v-model="item.Actions"
        ></JournalTypeContext>
      </template>
      <template #form_input_Previews="{ item }">
        <JournalTypeContext
          ref="listPreviews"
          title="Previews"
          v-model="item.Previews"
        ></JournalTypeContext>
      </template>
      <template #form_input_DefaultOffset="{ item }">
        <AccountSelector v-model="item.DefaultOffset" row></AccountSelector>
      </template>
      <template #form_input_InventoryDimension="{ item }">
        <dimension-invent-jurnal
          :key="data.keyDimension"
          v-model="item.InventoryDimension"
          :default-list="profile.Dimension"
        ></dimension-invent-jurnal>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          :key="data.keyDimension"
          :default-list="profile.Dimension"
          v-model="item.Dimension"
        ></dimension-editor>
      </template>
      <template #form_input_Vendor="{ item }">
        <s-input
          required
          :hide-label="false"
          label="Vendor"
          v-model="item.Vendor"
          class="w-full"
          use-list
          :disabled="item.VendorGroup == ''"
          :lookup-url="`/tenant/vendor/find?GroupID=${item.VendorGroup}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput } from "suimjs";

// import JournalTypeContext from "../fico/widget/JournalTypeContext.vue";
import JournalTypeContext from "./widget/JournalTypeContext.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import { authStore } from "@/stores/auth";

layoutStore().name = "tenant";
const featureID = "InventoryJournalType";
const profile = authStore().getRBAC(featureID);

const listControl = ref(null);
const listAction = ref(null);
const listPreviews = ref(null);
const journalConfig = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Inventory Journal Type",
  keyDimension: util.uuid(),
  record: {
    _id: "",
    Status: "",
  },
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Enable = true;
  record.Actions = [];
  record.Previews = [];
  data.titleForm = `Create Inventory Journal Type`;
  getDetailEmployee("", record);
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.record = record;
  data.titleForm = `Edit Inventory Journal Type | ${record._id}`;
  openForm(record);
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", [
      (v) => {
        let vLen = 0;
        let consistsInvalidChar = false;

        v.split("").forEach((ch) => {
          vLen++;
          const validCar =
            "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
              ch
            ) >= 0;
          if (!validCar) consistsInvalidChar = true;
          //console.log(ch,vLen,validCar)
        });

        if (vLen < 3 || consistsInvalidChar)
          return "minimal length is 3 and alphabet only";
        return "";
      },
    ]);
  });
}

function onPreSave(record) {
  if (listAction.value && listAction.value.getDataValue()) {
    let line = listAction.value.getDataValue();
    record.Actions = line;
  }
  if (listPreviews.value && listPreviews.value.getDataValue()) {
    let line = listPreviews.value.getDataValue();
    record.Previews = line;
  }
}

function saveConfig() {}

function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Inventory Journal Type";
  }
}

function getDetailEmployee(_id, record) {
  let payload = [];
  if (_id) {
    payload = [_id];
  }
  axios.post("/tenant/employee/get-emp-warehouse", payload).then(
    (r) => {
      // if (r.data.Dimension != null) {
      //   record.Dimension = r.data.Dimension;
      // }
      // record.InventoryDimension = {
      //   ...record.InventoryDimension,
      //   WarehouseID: r.data.Warehouse._id,
      // };
      // data.record = record;
      // data.keyDimension = util.uuid();
    },
    (e) => util.showError(e)
  );
}
</script>
