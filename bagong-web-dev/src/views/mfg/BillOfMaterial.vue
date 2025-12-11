<template>
  <div class="w-full">
    <data-list
      class="card grid-line-items"
      ref="listControl"
      :title="data.titleForm"
      form-hide-submit
      grid-config="/mfg/bom/gridconfig"
      form-config="/mfg/bom/formconfig"
      grid-read="/mfg/bom/gets"
      form-read="/mfg/bom/get"
      grid-mode="grid"
      grid-delete="/mfg/bom/delete"
      form-keep-label
      grid-hide-delete
      stay-on-form-after-save
      form-insert="/mfg/bom/draft"
      form-update="/mfg/bom/draft"
      :form-fields="['ItemID', 'SKU', 'OutputType', 'LedgerID']"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :form-tabs-new="data.tabs"
      :form-tabs-edit="data.tabs"
      :form-tabs-view="data.tabs"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onCancelForm"
      @alterGridConfig="onAlterGridConfig"
      @form-field-change="onFormFieldChange"
    >
      <template #form_tab_Material="{ item }">
        <BillOfMaterialItem
          ref="MaterialConfig"
          v-model="data.record"
          :item="item"
          :itemID="item._id"
          :hide-detail="true"
          grid-config="/mfg/bom/material/gridconfig"
          :grid-read="'/mfg/bom/material/gets?BoMID=' + item._id"
        ></BillOfMaterialItem>
      </template>
      <template #form_tab_Manpower="{ item }">
        <BillOfMaterialManpower
          ref="manpowerConfig"
          v-model="data.record"
          :item="item"
          :itemID="item._id"
          :hide-detail="true"
          grid-config="/mfg/manpower/gridconfig"
          :grid-read="'/mfg/bom/manpower/gets?BoMID=' + item._id"
        ></BillOfMaterialManpower>
      </template>
      <template #form_tab_Machinery="{ item }">
        <BillOfMaterialMachinery
          ref="machineryConfig"
          v-model="data.record"
          :item="item"
          :itemID="item._id"
          :hide-detail="true"
          grid-config="/mfg/bom/machinery/gridconfig"
          :grid-read="'/mfg/bom/machinery/gets?BoMID=' + item._id"
        ></BillOfMaterialMachinery>
      </template>
      <template #form_input_ItemID="{ item }">
        <div v-show="item.OutputType == 'Item'">
          <s-input-sku-item
            ref="refItemVarian"
            label="Item Variant"
            v-model="item.ItemVarian"
            :record="item"
            :required="true"
            :keepErrorSection="true"
            :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
              item.ItemID,
              item.SKU
            )}`"
          ></s-input-sku-item>
        </div>
      </template>
      <template #form_input_LedgerID="{ item }">
        <s-input
          v-show="item.OutputType == 'Ledger'"
          ref="refLedgerID"
          label="LedgerID"
          v-model="item.LedgerID"
          class="w-full"
          use-list
          lookup-key="_id"
          :keepLabel="true"
          :required="true"
          :lookup-url="`/tenant/ledgeraccount/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          :disabled="data.loading"
          icon="content-save"
          class="btn_primary"
          label="Save"
          @click="onSave(item)"
        />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
import { DataList, util, SForm, SInput, SButton, loadFormConfig } from "suimjs";
import helper from "@/scripts/helper.js";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import BillOfMaterialItem from "./widget/BillOfMaterialItem.vue";
import BillOfMaterialManpower from "./widget/BillOfMaterialManpower.vue";
import BillOfMaterialMachinery from "./widget/BillOfMaterialMachinery.vue";
import SInputSkuItem from "../scm/widget/SInputSkuItem.vue";

layoutStore().name = "tenant";
const featureID = "BOMMaster";
const profile = authStore().getRBAC(featureID);
const auth = authStore();
const listControl = ref(null);
const refItemVarian = ref(null);
const refLedgerID = ref(null);
const MaterialConfig = ref(null);
const manpowerConfig = ref(null);
const machineryConfig = ref(null);
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
  titleForm: "Bill Of Material",
  tabs: ["General", "Material", "Manpower"],
  loading: false,
  record: {
    _id: "",
  },
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Bill Of Material`;
  record._id = "";
  record.ItemVarian = "";
  record.LedgerID = "";
  record.OutputType = "Item";
  openForm(record);
}

function editRecord(record) {
  let ItemVarian = "";
  if (record.ItemID) {
    ItemVarian = helper.ItemVarian(record.ItemID, record.SKU);
  }
  record.ItemVarian = ItemVarian;
  data.formMode = "edit";
  data.titleForm = `Edit Bill Of Material | ${record._id}`;
  data.record = record;
  openForm(record);
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
    document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(1) > div > div > div:nth-child(1)"
    ).style.marginBottom = "11px";
    cssStyle(record.OutputType);
  });
}

function onAlterGridConfig(cfg) {
  util.nextTickN(2, () => {
    cfg.setting.idField = "Created";
    cfg.setting.sortable = ["Created", "Title", "BoMGroup", "_id"];
    cfg.setting.keywordFields = ["_id", "Title", "BoMGroup"];
  });
}
function onFormFieldChange(field, v1, v2, old, record) {
  switch (field) {
    case "OutputType":
      record.LedgerID = "";
      record.ItemVarian = "";
      record.ItemID = "";
      record.SKU = "";
      cssStyle(v1);
      break;
    default:
      break;
  }
}
function cssStyle(value) {
  if (value == "Ledger") {
    document
      .querySelector(
        ".form_inputs > div.flex.section_group_container > div:nth-child(2) > div > div"
      )
      .classList.remove("gap-4");
    document
      .querySelector(
        ".form_inputs > div.flex.section_group_container > div:nth-child(2) > div > div"
      )
      .classList.add("gap-3");
    document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(2) > div > div > div:nth-child(2)"
    ).style.display = "none";
  } else {
    document
      .querySelector(
        ".form_inputs > div.flex.section_group_container > div:nth-child(2) > div > div"
      )
      .classList.add("gap-4");
    document
      .querySelector(
        ".form_inputs > div.flex.section_group_container > div:nth-child(2) > div > div"
      )
      .classList.remove("gap-3");
    document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(2) > div > div > div:nth-child(2)"
    ).style.display = "block";
  }
}
function saveMaterial(record) {
  let line = [];
  if (MaterialConfig.value && MaterialConfig.value.getDataValue()) {
    line = MaterialConfig.value.getDataValue();
  }
  let payload = {
    BoMID: record._id,
    BoMMaterials: line,
  };
  axios.post("/mfg/bom/material/save-multiple", payload).then(
    (r) => {},
    (e) => {
      return util.showError(e);
    }
  );
}

function saveManpower(record) {
  let line = [];
  if (manpowerConfig.value && manpowerConfig.value.getDataValue()) {
    line = manpowerConfig.value.getDataValue();
  }
  let payload = {
    BoMID: record._id,
    BoMManpowers: line,
  };
  axios.post("/mfg/bom/manpower/save-multiple", payload).then(
    (r) => {},
    (e) => {
      return util.showError(e);
    }
  );
}

function saveMachinery(record) {
  let line = [];
  if (machineryConfig.value && machineryConfig.value.getDataValue()) {
    line = machineryConfig.value.getDataValue();
  }
  let payload = {
    BoMID: record._id,
    BoMMachinery: line,
  };
  axios.post("/mfg/bom/machinery/save-multiple", payload).then(
    (r) => {},
    (e) => {
      return util.showError(e);
    }
  );
}

function onSave(record) {
  if (record.OutputType == "Item") {
    record.LedgerID = "";
  } else {
    record.ItemID = "";
    record.SKU = "";
  }
  data.loading = true;
  const payload = JSON.parse(JSON.stringify(record));
  let validVarian = false;
  let validLedgerID = false;
  if (refItemVarian.value) {
    validVarian = !refItemVarian.value.validate();
  }
  if (refLedgerID.value) {
    validLedgerID = !refLedgerID.value.validate();
  }

  if (
    (record.OutputType == "Item" && validVarian) ||
    !listControl.value.formValidate()
  ) {
    data.loading = false;
    return util.showError("field general is required");
  }

  if (
    (record.OutputType == "Ledger" && validLedgerID) ||
    !listControl.value.formValidate()
  ) {
    data.loading = false;
    return util.showError("field general is required");
  }

  axios.post("/mfg/bom/save", payload).then(
    (r) => {
      record = r.data;
      Promise.all([
        saveMaterial(r.data),
        saveManpower(r.data),
        saveMachinery(r.data),
      ]).then(() => {
        data.loading = false;
        listControl.value.setControlMode("grid");
        listControl.value.refreshList();
        data.titleForm = `Bill Of Material`;
        return util.showInfo("BOM has been successful save");
      });
    },
    (e) => {
      data.loading = false;
      return util.showError(e);
    }
  );
}

function postSubmit(record) {
  record.Status = "SUBMITTED";
  if (record.OutputType == "Item") {
    record.LedgerID = "";
  } else {
    record.ItemID = "";
    record.SKU = "";
  }
  axios.post("/mfg/bom/save", record).then(
    (r) => {
      Promise.all([
        saveMaterial(r.data),
        saveManpower(r.data),
        saveMachinery(r.data),
      ]).then(() => {
        listControl.value.refreshForm();
        listControl.value.setControlMode("grid");
        listControl.value.refreshList();
        data.titleForm = `Bill Of Material`;
        return util.showInfo("BOM has been successful save");
      });
    },
    (e) => {
      return util.showError(e);
    }
  );
}
function onCancelForm(mode) {
  if (mode === "grid") {
    data.titleForm = "Bill Of Material";
  }
}
onMounted(() => {});
</script>
