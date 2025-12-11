<template>
    <div class="w-full">
      <data-list 
        class="card"
        ref="listControl" 
        title="Sales Journal Type"
        grid-config="/sdp/salesorderjournaltype/gridconfig" 
        form-config="/sdp/salesorderjournaltype/formconfig"
        grid-read="/sdp/salesorderjournaltype/gets" 
        form-read="/sdp/salesorderjournaltype/get" 
        grid-mode="grid"
        grid-delete="/sdp/salesorderjournaltype/delete" 
        form-keep-label 
        form-insert="/sdp/salesorderjournaltype/save"
        form-update="/sdp/salesorderjournaltype/save" 
        :grid-fields="['Enable']" 
        :form-tabs-edit="['General']"
        :form-fields="['Actions', 'Previews', 'DefaultOffset', 'Dimension']" 
        :init-app-mode="data.appMode"
        :init-form-mode="data.formMode" 
        @formNewData="newRecord" 
        @formEditData="openForm" @postSave="saveConfig"
        :grid-hide-new="!profile.canCreate"
        :grid-hide-edit="!profile.canUpdate"
        :grid-hide-delete="!profile.canDelete"
      >
        <!-- <template #form_tab_Configuration="{ item }">
          <CustomerJournalTypeConfiguration ref="journalConfig" :id="item._id"></CustomerJournalTypeConfiguration>
        </template> -->
        <template #form_input_Actions="{ item }">
          <JournalTypeContext title="Action" v-model="item.Actions"></JournalTypeContext>
        </template>
        <template #form_input_Previews="{ item }">
          <JournalTypeContext title="Previews" v-model="item.Previews"></JournalTypeContext>
        </template>
        <template #form_input_DefaultOffset="{ item }">
          <AccountSelector v-model="item.DefaultOffset" row></AccountSelector>
        </template>
        <template #form_input_Dimension="{ item }">
          <dimension-editor v-model="item.Dimension" :default-list="profile.Dimension"></dimension-editor>
        </template>
        <!-- <template #form_input_Customer="{ item }">
          <s-input
            required
            :hide-label="false"
            label="Customer"
            v-model="item.Customer"
            class="w-full"
            use-list
            :disabled="item.CustomerGroup == ''"
            :lookup-url="`/tenant/customer/find?GroupID=${item.CustomerGroup}`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
          ></s-input>
        </template> -->
      </data-list>
    </div>
  </template>
  
  <script setup>
  import { reactive, ref, inject } from "vue";
  import { layoutStore } from "@/stores/layout.js";
  import { DataList, util, SInput } from "suimjs";
  import { authStore } from "@/stores/auth.js";
  
  import JournalTypeContext from './widget/JournalTypeContext.vue';
  import DimensionEditor from '@/components/common/DimensionEditorVertical.vue';
  import AccountSelector from '@/components/common/AccountSelector.vue';
//   import CustomerJournalTypeConfiguration from "./widget/CustomerJournalTypeConfiguration.vue";
  
  layoutStore().name = "tenant";

  const featureID = 'SalesJournalType';
  // authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
  // authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})
  const profile = authStore().getRBAC(featureID);
  
  const listControl = ref(null);
  const journalConfig = ref(null);
  const axios = inject('axios');
  
  const data = reactive({
    appMode: "grid",
    formMode: "edit",
  });
  
  function newRecord(record) {
    record._id = "";
    record.Name = "";
    record.Enable = true;
    record.Actions = [];
    record.Previews = [];
  
    openForm(record)
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
              "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(ch) >= 0;
            if (!validCar) consistsInvalidChar = true;
            //console.log(ch,vLen,validCar)
          });
  
          if (vLen < 3 || consistsInvalidChar)
            return "minimal length is 3 and alphabet only";
          return "";
        },
      ]);
    })
  }
  
  function saveConfig() {
    if (journalConfig.value && journalConfig.value.getDataValue()) {
      let dv = journalConfig.value.getDataValue()
      axios.post('/sdp/salesorderjournaltype/save', dv).then(
        (r) => {},
        (e) => {
          data.loading = false;
        }
      );
    }
  }
  </script>