<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Site"
      grid-config="/bagong/sitesetup/gridconfig"
      form-config="/bagong/sitesetup/formconfig"
      grid-read="/bagong/sitesetup/gets"
      form-read="/bagong/sitesetup/get"
      grid-mode="grid"
      grid-delete="/bagong/sitesetup/delete"
      form-keep-label
      form-insert="/bagong/sitesetup/insert"
      form-update="/bagong/sitesetup/update"
      :form-tabs-edit="data.tabsList"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      :form-fields="['Dimension']"
      @pre-save="onPreSave"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_Dimension="{ item,mode }">
        <dimension-editor v-model="item.Dimension"  :default-list="profile.Dimension" :read-only="mode == 'view'"></dimension-editor>
      </template>
      <template #form_tab_Benefit="{ item }">
        <div
          v-for="(dt, idx) in item.Benefits"
          :key="'benefit_' + idx"
          class="flex flex-row gap-2 mb-2"
        >
          <div class="basis-1/3">{{ idx + 1 }}. {{ dt.Name }}</div>
          <s-input class="basis-1/3" v-model="dt.Value" kind="number" />
          <div class="basis-1/3">
            <toggle
              v-model="dt.IsCash"
              class="w-[120px]"
              yes-label="Active"
              no-label="Inactive"
            />
          </div>
        </div>
      </template>

      <template #form_tab_Deduction="{ item }">
        <div
          v-for="(dt, idx) in item.Deductions"
          :key="'benefit_' + idx"
          class="flex flex-row gap-2 mb-2"
        >
          <div class="basis-1/3">{{ idx + 1 }}. {{ dt.Name }}</div>
          <s-input class="basis-1/3" v-model="dt.Value" kind="number" />
          <div class="basis-1/3">
            <toggle
              v-model="dt.IsCash"
              class="w-[120px]"
              yes-label="Active"
              no-label="Inactive"
            />
          </div>
        </div>
      </template>

      <template #form_tab_Config="{ item }">
        <site-config-shift
          title="Shift"
          v-model="item.Shift"
        ></site-config-shift>
        <site-configuration
          title="Overtime"
          v-model="item.Configuration"
        ></site-configuration>
        <site-config-overtime
          title="Overtime"
          v-model="item.Overtime"
          v-model:salaryUsed="item.SalaryUsed"
        ></site-config-overtime>
      </template>

      <template #form_tab_Expense="{ item }">
        <expense-grid-editor
          title="Expense"
          v-model="item.Expense"
        ></expense-grid-editor>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, watch, inject, onScopeDispose } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  DataList,
  util,
  SInput,
  SForm,
  SListEditor,
  createFormConfig,
} from "suimjs";
import { authStore } from "@/stores/auth.js";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import Toggle from "@/components/common/SButtonToggle.vue";
import SiteConfigShift from "./widget/SiteConfigShift.vue";
import SiteConfigOvertime from "./widget/SiteConfigOvertime.vue";
import SiteConfiguration from "./widget/SiteConfiguration.vue";
import ExpenseGridEditor from "./widget/ExpenseGridEditor.vue";

layoutStore().name = "tenant";


const FEATUREID = 'SiteMaster'
const profile = authStore().getRBAC(FEATUREID)

const axios = inject("axios");

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  tabsList: ["General", "Benefit", "Deduction", "Config", "Expense"],
  record: {},
});

function newRecord(record) {
  data.formMode = "new";
  record.IsActive = true;
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.record = record;

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
    getPayroll();
  });
}

function getPayroll() {
  const urlB = "/fico/payrollbenefit/find";
  const urlD = "/fico/payrolldeduction/find";

  axios.post(urlB).then(
    (r) => {
      let res = r.data.filter((o) => {
        return o.IsActive == true;
      });

      let obj = {};
      for (let i in res) {
        let ob = res[i];
        obj[ob._id] = ob;
      }
      if (data.formMode == "edit")
        buildBenefitsDeduction(data.record.Benefits, obj);
    },
    (e) => {
      util.showError(e);
    }
  );

  axios.post(urlD).then(
    (r) => {
      let res = r.data.filter((o) => {
        return o.IsActive == true;
      });

      let obj = {};
      for (let i in res) {
        let ob = res[i];
        obj[ob._id] = ob;
      }
      if (data.formMode == "edit")
        buildBenefitsDeduction(data.record.Deductions, obj);
    },
    (e) => {
      util.showError(e);
    }
  );
}

function buildBenefitsDeduction(source, obj) {
  for (let ky in obj) {
    let ob = obj[ky];
    if (source.length == 0) {
      source.push(ob);
    } else {
      let r = source.find((o) => o._id == ky);
      if (r == undefined) {
        source.push(ob);
      }
    }
  }
}

function onPreSave(r) {
  console.log(r);
}
</script>
