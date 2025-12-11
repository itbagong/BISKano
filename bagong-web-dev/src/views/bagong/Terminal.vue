<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Terminal"
      grid-config="/bagong/terminal/gridconfig"
      form-config="/bagong/terminal/formconfig"
      grid-read="/bagong/terminal/gets"
      form-read="/bagong/terminal/get"
      grid-mode="grid"
      grid-delete="/bagong/terminal/delete"
      form-keep-label
      form-insert="/bagong/terminal/insert"
      form-update="/bagong/terminal/update"
      :form-fields="['Corridor', 'Expenses', 'Dimension']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_Corridor="{ item,mode}">
        <corridor-grid-editor title="Corridor" v-model="item.Corridor" :read-only="mode=='view'" />
      </template>
      <template #form_input_Expenses="{ item,mode}">
        <expense-grid-editor title="Expense" v-model="item.Expenses" :read-only="mode=='view'" />
      </template>
      <template #form_input_Dimension="{ item ,mode}">
        <dimension-editor v-model="item.Dimension" :default-list="profile.Dimension" :read-only="mode=='view'" ></dimension-editor>
      </template>
      <template #grid_paging>&nbsp;</template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import { authStore } from "@/stores/auth.js";
import CorridorGridEditor from "./widget/CorridorGridEditor.vue";
import ExpenseGridEditor from "./widget/ExpenseGridEditor.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
layoutStore().name = "tenant";

const FEATUREID = 'Terminal'
const profile = authStore().getRBAC(FEATUREID)


const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Enable = true;
  record.IsActive = true;
  record.Expenses = [];

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
            "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
              ch
            ) >= 0;
          if (!validCar) consistsInvalidChar = true;
        });

        if (vLen < 3 || consistsInvalidChar)
          return "minimal length is 3 and alphabet only";
        return "";
      },
    ]);
  });
}
</script>
