<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      form-keep-label
      :title="data.title" 
      :grid-config="data.gridConfig"
      :grid-read="data.gridRead"
      :grid-delete="data.allowDelete ? data.gridDelete : ''"
      :form-config="data.formConfig"
      :form-read="data.formRead"
      grid-mode="grid"
      :form-insert="data.formInsert"
      :form-update="data.formUpdate"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      :form-tabs-edit="['General', 'Configuration']"
      @formNewData="newRecord"
      @formEditData="editRecord"
      :form-fields="['Dimension']"
      @formFieldChange="formFieldChange"

      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_Dimension="{ item }">
        <dimension-editor v-model="item.Dimension"></dimension-editor>
      </template>
      <template #form_tab_Configuration="{ item }">
        <BenefitDeductionConfiguration
          v-model="item.Detail"
          :page="
            route.query.objname.includes('payrollbenefit')
              ? 'payrollbenefit'
              : 'payrolldeduction'
          "
        ></BenefitDeductionConfiguration>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, watch, inject } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import { useRoute } from "vue-router";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import BenefitDeductionConfiguration from "./widget/BenefitDeductionConfiguration.vue";

layoutStore().name = "tenant";
const mapFeature = {
  '/fico/payrollbenefit':'Benefit',
  '/fico/payrolldeduction':'Deduction',
  '/fico/loansetup':'LoanMaster'
}

const FEATUREID = mapFeature[useRoute().query.objname]

const profile = FEATUREID === undefined ? 
                { 
                  canRead: true,
                  canCreate: true,
                  canUpdate: true,
                  canDelete: true,
                  canPosting: true,
                  canSpecial1: true,
                  canSpecial2: true,
                  canSpecial3: true,
                  Dimension:[]
                } : 
                authStore().getRBAC(FEATUREID)


const listControl = ref(null);
const route = useRoute();

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  gridConfig: route.query.objname + "/gridconfig",
  gridRead: route.query.objname + "/gets",
  gridDelete: route.query.objname + "/delete",
  formConfig: route.query.objname + "/formconfig",
  formRead: route.query.objname.includes("payrollbenefit")
    ? "/bagong/payrollbenefit/get"
    : "/bagong/payrolldeduction/get",
  formInsert: route.query.objname.includes("payrollbenefit")
    ? "/bagong/payrollbenefit/save"
    : "/bagong/payrolldeduction/save",
  formUpdate: route.query.objname.includes("payrollbenefit")
    ? "/bagong/payrollbenefit/save"
    : "/bagong/payrolldeduction/save",
  title: route.query.title,
});

function newRecord(record) {
  data.formMode = "new";
  record.IsActive = true;
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  openForm(record);
}

function formFieldChange(name, v1, v2, old, record) {
  const isManual = v1 == "Manual";
  const isCalcType = name == "CalcType";
  if (isCalcType) {
    record.Value = null;
    record.CustomEndPoint = "";
    setFormFieldAttr("CustomEndPoint", "hide", !isManual);
  }
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

    const isManual = record.CalcType == "Manual";
    setFormFieldAttr("CustomEndPoint", "hide", !isManual);
    setFormFieldAttr("Value", "hide", isManual);
  });
}

function setFormFieldAttr(name, attr, value) {
  listControl.value.setFormFieldAttr(name, attr, value);
}

watch(
  () => route.query.objname,
  (nv) => {
    data.gridConfig = route.query.objname + "/gridconfig";
    data.gridRead = route.query.objname + "/gets";
    data.gridDelete = route.query.objname + "/delete";
    data.formConfig = route.query.objname + "/formconfig";
    data.formRead = route.query.objname + "/get";
    data.formInsert = route.query.objname + "/insert";
    data.formUpdate = route.query.objname + "/update";

    util.nextTickN(2, () => {
      listControl.value.refreshList();
      listControl.value.refreshForm();
      window.location.reload();
    });
  }
);

watch(
  () => route.query.title,
  (nv) => {
    data.title = nv;
    listControl.value.setControlMode("grid");
  }
);
</script>
