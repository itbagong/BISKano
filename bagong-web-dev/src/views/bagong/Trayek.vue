<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Trayek"
      grid-config="/bagong/trayek/gridconfig"
      form-config="/bagong/trayek/formconfig"
      grid-read="/bagong/trayek/gets"
      form-read="/bagong/trayek/get"
      grid-mode="grid"
      grid-delete="/bagong/trayek/delete"
      form-keep-label
      form-insert="/bagong/trayek/insert"
      form-update="/bagong/trayek/update"
      :form-fields="['ConfigDeposit', 'ConfigPremi', 'Dimension']"
      :form-tabs-new="data.tabsList"
      :form-tabs-edit="data.tabsList"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @preSave="onPreSave"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_ConfigDeposit="{ item,mode}">
        <div class="input_label mb-2">Config deposit</div>
        <s-form
          ref="configDeposit"
          v-model="item.ConfigDeposit"
          :config="data.formCfgDeposit"
          keep-label
          only-icon-top
          hide-submit
          hide-cancel
          :mode="mode == 'view'?'view':'edit'"
        >
        </s-form>
      </template>

      <template #form_input_ConfigPremi="{ item,mode }">
        <div class="input_label mb-2">Config premi</div>
        <s-form
          ref="configPremi"
          v-model="item.ConfigPremi"
          :config="data.formCfgPremi"
          keep-label
          only-icon-top
          hide-submit
          hide-cancel
          @field-change="fieldCfgPremiChange"
          :mode="mode == 'view'?'view':'edit'"
        >
        </s-form>
      </template>

      <template #form_input_Dimension="{ item,mode}">
        <dimension-editor
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :read-only="mode == 'view'"
        ></dimension-editor>
      </template>

      <template #form_tab_Tarifs="{ item }">
        <table
          class="w-full shadow rounded-md relative"
          v-if="item.Terminals && item.Terminals.length > 0"
        >
          <thead>
            <tr
              class="border-[#D8D8D8] bg-[#F8F8F9] [&>*]:px-2 [&>*]:py-1 [&>*]:border-[#D8D8D8] [&>*]:border-[1px] [&>*]:font-normal"
            >
              <th>From To</th>
              <th v-for="(aa, idx) in item.Terminals" :key="'headth_' + idx">
                {{ setTerminalName(aa) }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(aa, idx) in item.Terminals"
              :key="'bodytr_' + idx"
              class="[&>*]:px-2 [&>*]:py-1 [&>*]:border-[#D8D8D8] [&>*]:border-[1px] [&>*]:font-normal"
            >
              <td>
                {{ setTerminalName(aa) }}
              </td>
              <td
                class="text-center"
                v-for="(bb, idx2) in item.Terminals"
                :key="'bodytd_' + idx2"
              >
                <div
                  class="grid content-center"
                  v-if="data.tarifMatrix[aa + '#' + bb]"
                >
                  <s-input
                    kind="number"
                    v-model="data.tarifMatrix[aa + '#' + bb]['Rate']"
                  />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="font-medium">No data</div>
      </template>

      <template #form_tab_Expense="{ item,mode }">
        <expense-grid-editor
          title="Expense"
          page="Trayek"
          v-model="item.Expense"
          :group-id-value="['EXG0001', 'EXG0004']"
          :read-only="mode == 'view'"
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, watch, inject } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import { DataList, util, SInput, SForm, createFormConfig } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import ExpenseGridEditor from "./widget/ExpenseGridEditor.vue";

layoutStore().name = "tenant";

const FEATUREID = "Trayek";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const configDeposit = ref(null);
const configPremi = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  tabsList: ["General", "Tarifs", "Expense"],
  record: {},
  tarifMatrix: {},
  exitsMatrix: {},
  objTerminalList: {},
  formCfgDeposit: {},
  formCfgPremi: {},
});

const axios = inject("axios");

function newRecord(record) {
  data.formMode = "new";
  record.Tarifs = [];
  data.tarifMatrix = {};
  data.record = record;
  data.exitsMatrix = {};
  record.ConfigDeposit = {};
  record.ConfigPremi = {};
  record.IsActive = true;
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.record = record;
  data.exitsMatrix = record.Tarifs.reduce(
    (obj, item) => ((obj[item.From + "#" + item.To] = item), obj),
    {}
  );
  data.objTerminalList = {};
  for (let i in record.Terminals) {
    let ob = record.Terminals[i];
    getDataTerminals(ob);
  }
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

    // calc config premi
    data.record.ConfigPremi.Percent1 = record.ConfigPremi.Percent1 * 100;
    data.record.ConfigPremi.Percent2 = record.ConfigPremi.Percent2 * 100;

    let isInclude = [1, 3].includes(record.ConfigPremi.Method);
    configPremi.value.setFieldAttr("Target2", "hide", isInclude);
    configPremi.value.setFieldAttr("Percent2", "hide", isInclude);
  });
}

function mappingMatrix(source, flag) {
  if (flag == "obj") {
    let res = {};
    for (let i in source) {
      let from = source[i];
      for (let j in source) {
        let to = source[j];
        let nullObj = { From: from, To: to, Rate: 0 };
        if (data.exitsMatrix[from + "#" + to]) {
          res[from + "#" + to] = data.exitsMatrix[from + "#" + to];
        } else {
          res[from + "#" + to] = nullObj;
        }
      }
    }
    return res;
  }

  if (flag == "arr") {
    let res = [];
    for (let ky in data.tarifMatrix) {
      let val = data.tarifMatrix[ky];
      res.push(val);
    }
    return res;
  }
}

function onPreSave(record) {
  record.Tarifs = mappingMatrix(data.tarifMatrix, "arr");

  // calc config premi
  record.ConfigPremi.Percent1 = record.ConfigPremi.Percent1 / 100;
  record.ConfigPremi.Percent2 = record.ConfigPremi.Percent2 / 100;
  let isInclude = [1, 3].includes(record.ConfigPremi.Method);
  if (isInclude) {
    record.ConfigPremi.Target2 = 0;
    record.ConfigPremi.Percent2 = 0;
  }
}

function getDataTerminals(id) {
  const url = "/bagong/terminal/find?_id=" + id;
  axios.post(url).then(
    (r) => {
      if (r.data.length > 0) {
        let o = r.data[0];
        data.objTerminalList[o._id] = o;
      }
    },
    (e) => {}
  );
}

function setTerminalName(id) {
  return data.objTerminalList[id] ? data.objTerminalList[id].Name : "";
}

function fieldCfgPremiChange(name, v1, v2, old) {
  if (name == "Method") {
    let isInclude = [1, 3].includes(v1);
    configPremi.value.setFieldAttr("Target2", "hide", isInclude);
    configPremi.value.setFieldAttr("Percent2", "hide", isInclude);
  }
}

function genFormCfg() {
  // config deposit
  const cfgDeposit = createFormConfig("", true);
  cfgDeposit.addSection("", true).addRow(
    {
      field: "TargetFlat",
      label: "Target flat",
      kind: "number",
    },
    {
      field: "TargetNonFlat",
      label: "Target non flat",
      kind: "number",
    },
    {
      field: "IsToll",
      label: "Is toll",
      kind: "checkbox",
    }
  );
  data.formCfgDeposit = cfgDeposit.generateConfig();

  // config premi
  const cfgPremi = createFormConfig("", true);
  cfgPremi.addSection("", true).addRow(
    {
      field: "Method",
      label: "Method",
      kind: "input",
      useList: true,
      allowAdd: false,
      items: [1, 2, 3],
    },
    {
      field: "Target1",
      label: "Target 1",
      kind: "number",
    },
    {
      field: "Percent1",
      label: "Percent 1",
      kind: "number",
    },
    {
      field: "Target2",
      label: "Target 2",
      kind: "number",
    },
    {
      field: "Percent2",
      label: "Percent 2",
      kind: "number",
    }
  );
  data.formCfgPremi = cfgPremi.generateConfig();
}

watch(
  () => data.record.Terminals,
  (nv) => {
    data.tarifMatrix = mappingMatrix(data.record.Terminals, "obj");
  },
  { deep: true }
);

onMounted(() => {
  genFormCfg();
  getDataTerminals();
});
</script>
