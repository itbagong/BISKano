<template>
  <div class="w-full">
    <data-list class="card" ref="listControl" title="Reference Template"
      grid-config="/tenant/referencetemplate/gridconfig" form-config="/tenant/referencetemplate/formconfig"
      grid-read="/tenant/referencetemplate/gets" form-read="/tenant/referencetemplate/get" grid-mode="grid"
      grid-delete="/tenant/referencetemplate/delete" form-keep-label form-insert="/tenant/referencetemplate/insert"
      form-update="/tenant/referencetemplate/update" :form-fields="['Items', 'Dimension']" :init-app-mode="data.appMode"
      :init-form-mode="data.formMode" @formNewData="newRecord" @formEditData="openForm"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete">
      <template #form_input_Dimension="{ item }">
        <dimension-editor v-model="item.Dimension"  :default-list="profile.Dimension"></dimension-editor>
      </template>
      <template #form_input_Items="{ item }">
        <div class="input_label mb-2">Items</div>
        <s-list-editor no-gap :config="{}" v-model="item.Items" allow-add hide-header hide-select
          @validate-item="addItems" ref="refTemplate">
          <template #item="{ item }">
            <div class="w-full flex border">
              <div class="w-full border-r px-2 py-1">{{ item.Label }}</div>
              <div class="w-full border-r px-2 py-1">
                {{ item.ReferenceType }}
              </div>
              <div class="w-full px-2 py-1">{{ item.ConfigValue }}</div>
            </div>
          </template>

          <template #editor>
            <s-form ref="inputEditor" v-model="data.objRef" :config="data.formCfg" keep-label only-icon-top hide-submit
              hide-cancel @fieldChange="handleChange">
            </s-form>
          </template>
        </s-list-editor>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";

import DimensionEditor from '@/components/common/DimensionEditor.vue';
import {
  SListEditor,
  SInput,
  DataList,
  util,
  SForm,
  formInput,
  createFormConfig,
} from "suimjs";
import moment from "moment";

layoutStore().name = "tenant";

const FEATUREID = 'ReferenceTemplate'
const profile = authStore().getRBAC(FEATUREID)


const listControl = ref(null);
const refTemplate = ref(null);
const inputEditor = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  objRef: {},
  isDisabledCV: false,
  formCfg: {},
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Items = [];
  record.IsActive = true;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", [
      (v) => {
        let vLen = 0;
        let consistsInvalidChar = false;

        v.split("").forEach((ch) => {
          vLen++;
          const validCar =
            "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz".indexOf(
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
    resetItems();
  });
}

function addItems(record) {
  let isValid = inputEditor.value.validate();
  if (!isValid) return;

  for (let ky in data.objRef) {
    let val = data.objRef[ky];
    record[ky] = val;
  }
  refTemplate.value.setValidateItem(isValid);
  if (isValid) resetItems();
}

function resetItems() {
  data.objRef = {};
  inputEditor.value.setFieldAttr("ConfigValue", "required", false);
  inputEditor.value.setFieldAttr("ConfigValue", "hide", false);
}

function handleChange(name, v1, v2, old) {
  if (name == "ReferenceType") {
    data.objRef.ConfigValue = "";
    let isInclude = ["items", "lookup"].includes(v1);
    inputEditor.value.setFieldAttr("ConfigValue", "required", isInclude);
    inputEditor.value.setFieldAttr("ConfigValue", "hide", !isInclude);
  }
}

function genCfgRef() {
  const cfg = createFormConfig("", true);
  const label_input = new formInput();
  label_input.field = "Label";
  label_input.label = "Label";
  label_input.kind = "string";
  label_input.required = true;

  const type_input = new formInput();
  type_input.field = "ReferenceType";
  type_input.label = "Type";
  type_input.kind = "string";
  type_input.required = true;
  type_input.items = ["date", "items", "lookup", "number", "text", "textarea"];
  type_input.useList = true;
  type_input.allowAdd = false;

  const label_value = new formInput();
  label_value.field = "ConfigValue";
  label_value.label = "Config value";
  label_value.kind = "string";

  cfg
    .addSection("General", false)
    .addRowAuto(6, label_input, type_input, label_value);
  data.formCfg = cfg.generateConfig();
}

onMounted(() => {
  genCfgRef();
});
</script>
