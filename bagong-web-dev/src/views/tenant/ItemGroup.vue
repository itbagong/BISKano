<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/tenant/itemgroup/gridconfig"
      form-config="/tenant/itemgroup/formconfig"
      grid-read="/tenant/itemgroup/gets"
      form-read="/tenant/itemgroup/get"
      grid-mode="grid"
      grid-delete="/tenant/itemgroup/delete"
      form-keep-label
      form-insert="/tenant/itemgroup/save"
      form-update="/tenant/itemgroup/save"
      :stay-on-form-after-save="false"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :grid-fields="['Enable']"
      :form-fields="['PhysicalDimension', 'FinanceDimension']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :form-tabs-new="data.formTabs"
      :form-tabs-edit="data.formTabs"
      :formInitialTab="data.formInitialTab"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @postSave="onPostSave"
      @formLoaded="onRowUpdate"
      @controlModeChanged="onControlModeChanged"
    >
      <template #form_tab_Specification="{ item }">
        <ItemSpecification
          ref="specificationConfig"
          typeSpec="itemgroup"
          :item="item"
        ></ItemSpecification>
      </template>
      <template #form_input_PhysicalDimension="{ item }">
        <div class="flex gap-2">
          <s-input
            v-for="(value, key, index) in item.PhysicalDimension"
            v-model="item.PhysicalDimension[key]"
            :label="data.dim.find((v) => v.key == key).label"
            kind="checkbox"
            class="w-full"
            :read-only="false"
          ></s-input>
        </div>
      </template>
      <template #form_input_FinanceDimension="{ item }">
        <div class="flex gap-2">
          <s-input
            v-for="(value, key, index) in item.FinanceDimension"
            v-model="item.FinanceDimension[key]"
            :label="data.dim.find((v) => v.key == key).label"
            kind="checkbox"
            class="w-full"
          ></s-input>
        </div>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, inject, computed } from "vue";
import { loadGridConfig, DataList, util, SInput } from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
import DimensionItem from "./widget/DimensionItem.vue";
import ItemSpecification from "./widget/ItemSpecification.vue";

layoutStore().name = "tenant";
const featureID = "ItemGroup";
const profile = authStore().getRBAC(featureID);
const listControl = ref(null);
const generalConfig = ref(null);
const specificationConfig = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Item group",
  formInitialTab: 0,
  formTabs: ["General"],
  record: {},
  tabspecKey: "spec",
  dim: [
    {
      key: "IsEnabledSpecVariant",
      label: "Variant",
    },
    {
      key: "IsEnabledSpecSize",
      label: "Size",
    },
    {
      key: "IsEnabledSpecGrade",
      label: "Grade",
    },
    {
      key: "IsEnabledItemBatch",
      label: "Batch",
    },
    {
      key: "IsEnabledItemSerial",
      label: "Serial",
    },
    {
      key: "IsEnabledLocationWarehouse",
      label: "Warehouse",
    },
    {
      key: "IsEnabledLocationAisle",
      label: "Aisle",
    },
    {
      key: "IsEnabledLocationSection",
      label: "Section",
    },
    {
      key: "IsEnabledLocationBox",
      label: "Box",
    },
  ],
  listSpec: [],
  gridCfgSpec: {},
});

function newRecord(record) {
  data.formMode = "new";
  data.formInitialTab = 0;
  data.formTabs = ["General"];
  record._id = "";
  record.Name = "";
  record.PhysicalDimension = {
    IsEnabledSpecVariant: false,
    IsEnabledSpecSize: false,
    IsEnabledSpecGrade: false,
    IsEnabledItemBatch: false,
    IsEnabledItemSerial: false,
    IsEnabledLocationWarehouse: false,
    IsEnabledLocationAisle: false,
    IsEnabledLocationSection: false,
    IsEnabledLocationBox: false,
  };
  record.FinanceDimension = {
    IsEnabledSpecVariant: false,
    IsEnabledSpecSize: false,
    IsEnabledSpecGrade: false,
    IsEnabledItemBatch: false,
    IsEnabledItemSerial: false,
    IsEnabledLocationWarehouse: false,
    IsEnabledLocationAisle: false,
    IsEnabledLocationSection: false,
    IsEnabledLocationBox: false,
  };
  data.titleForm = "Create Item Group";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.record = record;
  data.formTabs = ["General", "Specification"];
  data.titleForm = `Edit Item Group - ${record.Name}`;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", [
      (v) => {
        if (v == 0) return "required";
        return "";
      },
    ]);
  });
}

function onPostSave(record) {
  if (
    record.PhysicalDimension.IsEnabledSpecVariant ||
    record.PhysicalDimension.IsEnabledSpecSize ||
    record.PhysicalDimension.IsEnabledSpecGrade
  ) {
    data.formTabs = ["General", "Specification"];
    data.formInitialTab = 1;
  } else {
    data.formTabs = ["General"];
    data.formInitialTab = 0;
  }

  if (specificationConfig.value && specificationConfig.value.getDataValue()) {
    let dv = specificationConfig.value.getDataValue();
    dv.map(function (sp) {
      if (!record.PhysicalDimension.IsEnabledSpecVariant) {
        sp.SpecVariantID = "";
      }
      if (!record.PhysicalDimension.IsEnabledSpecSize) {
        sp.SpecSizeID = "";
      }
      if (!record.PhysicalDimension.IsEnabledSpecGrade) {
        sp.SpecGradeID = "";
      }
      return sp;
    });
    axios.post("/tenant/itemspec/save-multiple", dv).then(
      (r) => {},
      (e) => {
        data.loading = false;
      }
    );
  }
}

function onRowUpdate(record) {
  if (record.PhysicalDimension) {
    if (
      record.PhysicalDimension.IsEnabledSpecVariant ||
      record.PhysicalDimension.IsEnabledSpecSize ||
      record.PhysicalDimension.IsEnabledSpecGrade
    ) {
      data.formTabs = ["General", "Specification"];
      data.formInitialTab = 1;
    }
  } else {
    data.formTabs = ["General"];
  }
}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Item group";
  }
}
onMounted(() => {});
</script>
