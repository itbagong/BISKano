<template>
  <div class="w-full">
    <data-list
      class="card grid-line-items"
      ref="listControl"
      :title="data.titleForm"
      :grid-hide-select="true"
      grid-config="/mfg/routine/template/gridconfig"
      form-config="/mfg/routine/template/formconfig"
      grid-read="/mfg/routine/template/gets"
      form-read="/mfg/routine/template/get"
      grid-mode="grid"
      grid-delete="/mfg/routine/template/delete"
      form-keep-label
      form-insert="/mfg/routine/template/save"
      form-update="/mfg/routine/template/save"
      :grid-fields="['Enable', 'Dimension']"
      :form-fields="['Items', 'Dimension']"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @alterGridConfig="onAlterGridConfig"
    >
      <template #grid_Dimension="{ item }">
        <DimensionText :dimension="item.Dimension" />
      </template>
      <template #form_input_Items="{ item }">
        <s-grid
          v-model="item.Items"
          ref="itemsControlLine"
          class="w-full grid-line-items"
          hide-search
          hide-sort
          :hide-detail="true"
          hide-footer
          editor
          :config="data.gridItemsConfig"
          hide-refresh-button
          hide-select
          auto-commit-line
          no-confirm-delete
          form-keep-label
          @new-data="onNewItem"
          @delete-data="onDeleteItem"
        >
        </s-grid>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          :key="data.keyFinancialDimension"
          ref="FinancialDimension"
          sectionTitle="Financial Dimension"
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
import { DataList, util, SGrid, loadGridConfig } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionText from "@/components/common/DimensionText.vue";

layoutStore().name = "tenant";
const featureID = "P2HTemplate";
const profile = authStore().getRBAC(featureID);
const headOffice = layoutStore().headOfficeID;

const listControl = ref(null);
const itemsControlLine = ref(null);
const axios = inject("axios");
const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Routine template",
  record: {
    _id: "",
    Items: [],
  },
  gridItemsConfig: {
    fields: [],
    setting: {},
  },
  keyFinancialDimension: util.uuid(),
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Routine template`;
  record._id = "";
  record.TrxType = "Routine template";
  record.Items = [];
  data.record = record;
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Routine template | ${record._id}`;
  data.record = record;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}
function onAlterGridConfig(cfg) {
  util.nextTickN(2, () => {
    cfg.sortable = ["Created", "_id"];
    cfg.setting.idField = "Created";
    cfg.setting.sortable = ["Created", "_id"];
  });
}

function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Routine template";
  }
}
function onNewItem(record) {
  data.record.Items.push({ ...record, ItemNo: data.record.Items.length + 1 });
  itemsControlLine.value.setRecords(data.record.Items);
}
function onDeleteItem(record, index) {
  const newRecords = data.record.Items.filter((dt, idx) => {
    return idx != index;
  });
  data.record.Items = newRecords;
}
onMounted(() => {
  loadGridConfig(axios, "/mfg/routine/template/items/gridconfig").then(
    (r) => {
      data.gridItemsConfig = r;
    },
    (e) => util.showError(e)
  );
});
</script>
