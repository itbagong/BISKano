<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Unit"
      grid-config="/tenant/unit/gridconfig"
      form-config="/tenant/unit/formconfig"
      grid-read="/tenant/unit/gets"
      form-read="/tenant/unit/get"
      grid-mode="grid"
      grid-delete="/tenant/unit/delete"
      form-keep-label
      form-insert="/tenant/unit/save"
      form-update="/tenant/unit/save"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :grid-fields="['Enable']"
      :form-fields="['Dimension']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :form-tabs-new="['General', 'Conversion']"
      :form-tabs-edit="['General', 'Conversion']"
      :form-hide-submit="true"
      :form-hide-cancel="true"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @formNewData="newRecord"
      @formEditData="openForm"
      @postSave="onPostSave"
      @alterGridConfig="onAlterGridConfig"
      @alterFormConfig="onAlterFormConfig"
    >
      <template #form_tab_Conversion="{ item }">
        <div v-if="item._id == ''" class="nodata">No data</div>
        <UnitConversion
          v-else
          ref="conversionConfig"
          :itemID="item._id"
          :form-mode="data.formMode"
          :hide-detail="false"
          :item="item"
          grid-config="/tenant/unit/conversion/gridconfig"
          :grid-read="'/tenant/unit/conversion/gets?FromUnit=' + item._id"
        ></UnitConversion>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          :disabled="data.loading"
          :icon="`content-save`"
          class="btn_primary submit_btn"
          label="Save"
          @click="onSave(item)"
        />
        <s-button
          :disabled="data.loading"
          :icon="`rewind`"
          class="btn_warning back_btn"
          :label="'Back'"
          @click="onBackForm"
        />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
import { DataList, util, SInput, SButton } from "suimjs";
import UnitConversion from "./widget/UnitConversion.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";

layoutStore().name = "tenant";
const featureID = "UnitMaster";
const profile = authStore().getRBAC(featureID);
const listControl = ref(null);
const conversionConfig = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  loading: false,
  record: {
    _id: "",
    Name: "",
  },
});

function newRecord(record) {
  data.formMode = "new";
  record._id = "";
  record.Name = "";
  openForm(record);
}

function openForm(record) {
  util.nextTickN(2, () => {
    let line = document.querySelector(".tab_container > div.tab");
    line.addEventListener("click", function (event) {
      util.nextTickN(2, () => {
        if (conversionConfig.value && conversionConfig.value.getDataValue()) {
          let dv = conversionConfig.value.getDataValue();
          dv.map(function (val) {
            val.FromUnit = record.Name;
            val.FromUnitID = record._id;
            return val;
          });
          conversionConfig.value.setDataValue(dv);
        }
      });
    });
  });
}

function onSave(record) {
  if (conversionConfig.value && conversionConfig.value.getDataValue()) {
    let dv = conversionConfig.value.getDataValue();
    dv.map(function (val) {
      val.FromUnit = record._id;
      return val;
    });
    let group = dv.reduce((result, currentObject) => {
      const key = `${currentObject.ToUnit}`;
      result[key] = result[key] || [];
      result[key].push(currentObject);
      return result;
    }, {});

    for (let i = 0; i < Object.keys(group).length; i++) {
      if (group[Object.keys(group)[i]].length > 1) {
        return util.showError("Unit duplicate conversion To unit");
      }
    }
  }
  let valid = true;
  if (listControl.value) {
    valid = listControl.value.formValidate() && record.Name != "";
  }
  data.loading = true;
  listControl.value.setFormLoading(true);
  if (valid) {
    listControl.value.submitForm(
      record,
      () => {
        listControl.value.setFormLoading(false);
      },
      () => {
        listControl.value.setLoadingForm(false);
      }
    );
  } else {
    listControl.value.setFormLoading(false);
    return util.showError("field is required");
  }
}

function onPostSave(record) {
  if (conversionConfig.value && conversionConfig.value.getDataValue()) {
    let dv = conversionConfig.value.getDataValue();
    dv.map(function (val) {
      val.FromUnit = record._id;
      return val;
    });
    let payload = {
      FromUnit: record._id,
      UnitConversions: dv,
    };
    axios
      .post("/tenant/unit/conversion/save-multiple", payload)
      .then(
        (r) => {},
        (e) => {
          data.loading = false;
        }
      )
      .finally(function () {
        data.loading = false;
      });
  } else {
    data.loading = false;
  }
}
function onBackForm() {
  listControl.value.cancelForm();
}

function onAlterGridConfig(cfg) {
  cfg.fields.map(function (el) {
    if (el.field == "UOMCategory") {
      el.label = "UOM Category";
    } else if (el.field == "UOMRatioType") {
      el.label = "UOM Ratio Type";
    }
    return el;
  });
}
function onAlterFormConfig(cfg) {
  cfg.sectionGroups = cfg.sectionGroups.map((sectionGroup) => {
    sectionGroup.sections = sectionGroup.sections.map((section) => {
      section.rows.map((row) => {
        row.inputs = row.inputs
          .filter((input) => ["Enable"].indexOf(input.field) == -1)
          .map((input) => {
            if (input.field === "UOMCategory") {
              input.label = "UOM Category";
            } else if (input.field === "UOMRatioType") {
              input.label = "UOM Ratio Type";
            }
            return input;
          });
        return row;
      });
      return section;
    });
    return sectionGroup;
  });
}
</script>
