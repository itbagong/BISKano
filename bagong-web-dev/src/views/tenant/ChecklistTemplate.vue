<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Checklist Template"
      grid-config="/tenant/checklisttemplate/gridconfig"
      form-config="/tenant/checklisttemplate/formconfig"
      grid-read="/tenant/checklisttemplate/gets"
      form-read="/tenant/checklisttemplate/get"
      grid-mode="grid"
      grid-delete="/tenant/checklisttemplate/delete"
      form-keep-label
      form-insert="/tenant/checklisttemplate/insert"
      form-update="/tenant/checklisttemplate/update"
      :form-fields="['Checklists', 'Dimension']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_Checklists="{ item }">
        <div class="input_label mb-2">Checklists Item</div>
        <s-list-editor
          no-gap
          :config="{}"
          v-model="item.Checklists"
          allow-add
          hide-header
          hide-select
          @validate-item="addChecklistItem"
          ref="Checklist"
        >
          <template #item="{ item }">
            <div class="w-full flex border gap-2">
              <div class="w-full border-r px-2 py-1">{{ item.Key }}</div>
              <div class="w-full border-r px-2 py-1" :id="`el_${item.PIC}`" />
            </div>
          </template>
          <template #editor>
            <s-form
              ref="inputEditor"
              v-model="data.objCheck"
              :config="data.formCfg"
              keep-label
              only-icon-top
              hide-submit
              hide-cancel
            >
            </s-form>
          </template>
        </s-list-editor>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor v-model="item.Dimension" :default-list="profile.Dimension"></dimension-editor>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, inject } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { SListEditor, DataList, util, createFormConfig, SForm } from "suimjs";
import { authStore } from "@/stores/auth.js";


import DimensionEditor from "@/components/common/DimensionEditor.vue";
import moment from "moment";

layoutStore().name = "tenant";


const FEATUREID = 'ChecklistTemplate'
const profile = authStore().getRBAC(FEATUREID)


const axios = inject("axios");

const listControl = ref(null);
const Checklist = ref(null);
const inputEditor = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  formCfg: {},
  objCheck: {},
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Checklists = [];
  record.IsActive = true;
  openForm(record);
}

function openForm(record) {
  // call get value pic for existing item
  if (record.Checklists.length > 0) {
    record.Checklists.forEach((v) => getValuePIC(v.PIC));
  }

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
  });
}

function addChecklistItem(record) {
  let isValid = inputEditor.value.validate();
  for (let ky in data.objCheck) {
    let val = data.objCheck[ky];
    record[ky] = val;
  }

  // call get value pic for new item
  getValuePIC(record.PIC);
  Checklist.value.setValidateItem(isValid);
  if (isValid) resetChecklistItem();
}

function resetChecklistItem() {
  data.objCheck = {};
}

function formatDate(params) {
  return moment(params).local().format("DD-MMM-YYYY");
}

// function for get value pic
function getValuePIC(id) {
  axios
    .post("/iam/user/find-by", {
      FindBy: "_id",
      FindID: id,
    })
    .then((r) => {
      const displayName = r.data[0].DisplayName;
      if (displayName) {
        document.getElementById(`el_${id}`).textContent = displayName;
      }
    });
}

function genCfgCheck() {
  const cfg = createFormConfig("", true);
  cfg.addSection("", true).addRow(
    {
      field: "Key",
      kind: "input",
      label: "Key",
      required: true,
    },
    {
      field: "PIC",
      kind: "input",
      label: "PIC",
      required: false,
      useList: true,
      allowAdd: false,
      lookupKey: "_id",
      lookupLabels: ["DisplayName"],
      lookupSearchs: ["_id", "DisplayName"],
      lookupUrl: "/iam/user/find-by",
    }
  );
  data.formCfg = cfg.generateConfig();
}

onMounted(() => {
  genCfgCheck();
});
</script>
