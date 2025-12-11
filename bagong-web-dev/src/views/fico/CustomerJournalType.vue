<template>
  <div class="w-full">
    <data-list class="card" ref="listControl" title="Customer Journal Type"
      grid-config="/fico/customerjournaltype/gridconfig" form-config="/fico/customerjournaltype/formconfig"
      grid-read="/fico/customerjournaltype/gets" form-read="/fico/customerjournaltype/get" grid-mode="grid"
      grid-delete="/fico/customerjournaltype/delete" form-keep-label form-insert="/fico/customerjournaltype/save"
      form-update="/fico/customerjournaltype/save" :grid-fields="['Enable']" :form-tabs-edit="['General', 'Configuration']"
      :form-fields="['Actions', 'Previews', 'DefaultOffset', 'Dimension', 'Customer']" :init-app-mode="data.appMode"
      @formNewData="newRecord" @formEditData="openForm" @postSave="saveConfig"
      :form-default-mode="data.formMode"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      >
      <template #form_tab_Configuration="{ item,mode}">
        <CustomerJournalTypeConfiguration ref="journalConfig" :id="item._id" :read-only="mode == 'view'"></CustomerJournalTypeConfiguration>
      </template>
      <template #form_input_Actions="{ item,mode }">
        <JournalTypeContext title="Action" v-model="item.Actions"  :read-only="mode == 'view'"></JournalTypeContext>
      </template>
      <template #form_input_Previews="{ item,mode}">
        <JournalTypeContext title="Previews" v-model="item.Previews"  :read-only="mode == 'view'"></JournalTypeContext>
      </template>
      <template #form_input_DefaultOffset="{ item,mode }">
        <AccountSelector v-model="item.DefaultOffset" row  :read-only="mode == 'view'"></AccountSelector>
      </template>
      <template #form_input_Dimension="{ item,mode }">
        <dimension-editor
          v-model="item.Dimension" 
          :default-list="profile.Dimension" 
          :read-only="mode == 'view'"
          ></dimension-editor>
      </template>
      <template #form_input_Customer="{ item,mode}">
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
          :read-only="mode == 'view'"
        ></s-input>
      </template>
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
import CustomerJournalTypeConfiguration from "./widget/CustomerJournalTypeConfiguration.vue";

layoutStore().name = "tenant";

const FEATUREID = 'CustomerJournalType'
const profile = authStore().getRBAC(FEATUREID)


const listControl = ref(null);
const journalConfig = ref(null);
const axios = inject('axios');

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
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
    axios.post('/bagong/customerjournaltypeconfiguration/save', dv).then(
      (r) => {},
      (e) => {
        data.loading = false;
      }
    );
  }
}
</script>