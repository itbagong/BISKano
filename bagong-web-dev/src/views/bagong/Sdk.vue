<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="SDK"
      grid-config="/bagong/accident_fund/gridconfig"
      form-config="/bagong/accident_fund/formconfig"
      grid-read="/bagong/accident_fund/gets"
      form-read="/bagong/accident_fund/get"
      grid-mode="grid"
      grid-delete="/bagong/accident_fund/delete"
      form-keep-label
      form-insert="/bagong/accident_fund/insert"
      form-update="/bagong/accident_fund/update"
      :form-fields="['Balance', 'Dimension']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      :form-tabs-new="data.tabs"
      :form-tabs-edit="data.tabs"
      :form-tabs-view="data.tabs"
      @form-new-data="newRecord"
      @form-edit-data="openForm"
      @form-field-change="handleFieldChange"
      @pre-save="onPreSave"
      @post-save="onPostSave"
      @grid-row-deleted="onGridRowDeleted"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_Balance>
        <s-input
          kind="number"
          v-model="data.valBalance"
          label="Balance"
          class="w-full"
          readOnly
        />
      </template>

      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :read-only="mode =='view'"
        ></dimension-editor-vertical>
      </template>

      <template #form_tab_Line>
        <sdk-lines
          v-model="data.lines"
          :AccidentFundID="data.AccidentFundID"
          @deletedLines="dataDeletedLines"
        ></sdk-lines>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SInput, util } from "suimjs";
import { authStore } from "@/stores/auth.js";

import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import SdkLines from "./widget/SdkLines.vue";

layoutStore().name = "tenant";

const FEATUREID = 'SantunanDanaKecelakaan'
const profile = authStore().getRBAC(FEATUREID)

const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);

const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  tabs: ["General", "Line"],
  lines: [],
  valBalance: 0,
  AccidentFundID: "",
  delLines: [],
});

function newRecord(record) {
  record.EmployeeID = "";
  record.Position = "";
  record.Balance = 0;

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
        });

        if (vLen < 3 || consistsInvalidChar)
          return "minimal length is 3 and alphabet only";
        return "";
      },
    ]);

    data.valBalance = record.Balance;
    data.AccidentFundID = record._id;
    data.delLines = [];
  });
}

function handleFieldChange(name, v1, v2, old, record) {
  if (name == "EmployeeID") {
    axios.post("/bagong/employee/get", [v1]).then(
      (r) => {
        record.Position = r.data.Detail.Position;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function onPreSave(record) {
  // check if mutation is minus & greater than balance
  if (data.valBalance < 0) {
    if (-data.valBalance > record.Balance) {
      return;
    }
    return;
  }

  record.Balance = data.valBalance;
  if (data.delLines.length > 0) {
    data.delLines.forEach((item) => {
      const url = "/bagong/accident_fund/delete-accident-fund-detail";
      axios.post(url, item).then(
        (r) => {},
        (e) => {
          util.showError(e);
        }
      );
    });
  }
}

function onPostSave(record) {
  // check if mutation is minus & greater than balance
  if (data.valBalance < 0) {
    if (-data.valBalance > record.Balance) {
      util.showError("mutation is greater than balance");
      return;
    }
    util.showError("mutation is greater than balance");
    return;
  }

  const dataLines = data.lines.map((item) => ({
    AccidentFundID: record._id,
    ...item,
  }));
  dataLines.forEach((item) => {
    const url = "/bagong/accident_funddetail/save";
    axios.post(url, item).then(
      (r) => {},
      (e) => {
        util.showError(e);
      }
    );
  });
}

function onGridRowDeleted(record) {
  const url = "/bagong/accident_funddetail/find?AccidentFundID=" + record._id;
  axios.post(url).then(
    (r) => {
      if (r.data.length > 0) {
        r.data.forEach((item) => {
          onDeleteDetail(item);
        });
      }
    },
    (e) => {
      util.showError(e);
    }
  );
}

function onDeleteDetail(item) {
  const url = "/bagong/accident_funddetail/delete";
  axios.post(url, item).then(
    (r) => {},
    (e) => {
      util.showError(e);
    }
  );
}

function dataDeletedLines(datas) {
  data.delLines = datas;
}

watch(
  () => data.lines,
  (nv) => {
    const calcBalance = nv.reduce((a, b) => {
      return a + b.Mutation;
    }, 0);
    data.valBalance = calcBalance;
  },
  { deep: true }
);
</script>
