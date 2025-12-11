<template>
  <div class="w-full">   
    <data-list
      :key="data.key"
      class="card"
      ref="listControl"
      v-if="data.gridConfig != ''"
      form-keep-label
      :title="data.title" 
      :grid-config="data.gridConfig"
      :grid-read="data.gridRead"
      :grid-delete="data.gridDelete"
      :form-config="data.formConfig"
      :form-read="data.formRead"
      grid-mode="grid"
      :form-insert="data.formInsert"
      :form-update="data.formUpdate"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      :form-fields="['Setting', 'Dimension']"
      @formNewData="newRecord"
      @formEditData="openForm"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_Setting="{ item, config,mode }">
        <setting-selector :label="config.label" v-model="item.Setting" :read-only="mode=='view'"/>
      </template>
      <template #form_input_Dimension="{ item,mode }">
        <dimension-editor v-model="item.Dimension" :default-list="profile.Dimension" :read-only="mode=='view'"></dimension-editor>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, watch,computed } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";;
import { useRoute } from "vue-router";


import { DataList, util, SInput } from "suimjs"
import DimensionEditor from "@/components/common/DimensionEditor.vue";

import SettingSelector from "@/components/common/SettingSelector.vue";

layoutStore().name = "tenant";

const mapFeature = {
  '/fico/paymentterm': 'PaymentTerm',
  '/tenant/cashbankgroup': 'CashAndBankGroup',
  '/tenant/customergroup': 'CustomerGroup',
  '/tenant/assetgroup': 'AssetGroup',
  '/kara/holiday': 'WorkHoliday',
  '/kara/holidayitem': 'WorkHolidayItem',
  '/fico/loansetup': 'LoanMaster',
  '/tenant/employeegroup': 'EmployeeGroup',
  '/tenant/company':'Company',
  '/tenant/numseq':'NumberSequence',
  '/tenant/numseqsetup':'NumberSequenceSetup',
  '/tenant/dimension':'Dimension',
  '/tenant/expensetype':'ExpenseType',
  '/tenant/expensetypegroup':'ExpenseTypeGroup',
  '/fico/taxsetup':'TaxCode',
  '/tenant/vendorgroup':'VendorGroup',
  '/tenant/itemserial':'ItemSerial',
  '/tenant/itembatch':'ItemBatch',
  '/tenant/warehouse/group':'WarehouseGroup',
  '/tenant/section':'Section', 
  '/tenant/aisle':'Aisle', 
  '/tenant/box':'BoxMaster', 
  '/tenant/specvariant':'SpecVariant', 
  '/tenant/specsize':'SpecSize', 
  '/tenant/specgrade':'SpecGrade', 
  '/tenant/contact': 'Contact',
} 

let profile = getProfile(mapFeature[useRoute().query.objname]) 

function getProfile(feature){
  return feature === undefined ? 
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
                authStore().getRBAC(feature)

}


const listControl = ref(null);

const route = useRoute();

const data = reactive({
  key: util.uuid(),
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  title: route.query.title || route.query.objname,
  gridConfig: route.query.objname + "/gridconfig",
  gridRead: route.query.objname + "/gets",
  gridDelete: route.query.objname + "/delete",
  formConfig: route.query.objname + "/formconfig",
  formRead: route.query.objname + "/get",
  formInsert: route.query.objname + "/insert",
  formUpdate: route.query.objname + "/update", 
});

watch(
  () => route.query.objname,
  (nv) => {


    profile = getProfile(mapFeature[route.query.objname])


    data.gridConfig = route.query.objname + "/gridconfig";
    data.gridRead = route.query.objname + "/gets";
    data.gridDelete = route.query.objname + "/delete";
    data.formConfig = route.query.objname + "/formconfig";
    data.formRead = route.query.objname + "/get";
    data.formInsert = route.query.objname + "/insert";
    data.formUpdate = route.query.objname + "/update";


    util.nextTickN(2, () => {
      data.key = util.uuid();
      listControl.value.refreshList();
      listControl.value.refreshForm();
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

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Enable = true;
  record.IsActive = true;
  record.Setting = {};
  record.Modules = []
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
          //console.log(ch,vLen,validCar)
        });

        if (vLen < 3 || consistsInvalidChar)
          return "minimal length is 3 and alphabet only";
        return "";
      },
    ]);


    if(route.query.objname == "/fico/taxsetup"){
      listControl.value.setFormFieldAttr("Modules", "rules", [
        (v) => { 
         
        if (!Array.isArray(v) || v.length == 0)
          return "Required";
        return "";
        },
     ]);
    }
  });
}
</script>
