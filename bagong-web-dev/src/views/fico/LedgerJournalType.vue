<template>
  <div class="w-full">
    <data-list class="card" ref="listControl" title="Ledger Journal Type"
      grid-config="/fico/ledgerjournaltype/gridconfig" form-config="/fico/ledgerjournaltype/formconfig"
      grid-read="/fico/ledgerjournaltype/gets" form-read="/fico/ledgerjournaltype/get" grid-mode="grid"
      grid-delete="/fico/ledgerjournaltype/delete" form-keep-label form-insert="/fico/ledgerjournaltype/insert"
      form-update="/fico/ledgerjournaltype/update" :grid-fields="['Enable']" :form-tabs-edit="['General']"
      :form-fields="['DefaultOffset', 'Actions', 'Previews', 'Dimension']" :init-app-mode="data.appMode" 
      :form-default-mode="data.formMode"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @formNewData="newRecord" @formEditData="openForm">
      <template #form_input_Actions="{ item,mode }">
        <JournalTypeContext title="Action" v-model="item.Actions" :read-only="mode =='view'"></JournalTypeContext>
      </template>
      <template #form_input_Previews="{ item,mode }">
        <JournalTypeContext title="Previews" v-model="item.Previews" :read-only="mode =='view'"></JournalTypeContext>
      </template>
      <template #form_input_DefaultOffset="{ item,mode }">
        <AccountSelector v-model="item.DefaultOffset" row  :read-only="mode =='view'"> </AccountSelector>
      </template>
      <template #form_input_Dimension="{ item,mode }">
        <dimension-editor v-model="item.Dimension" :default-list="profile.Dimension" :read-only="mode=='view'"></dimension-editor>
      </template>
    </data-list>
  </div>
</template>
  
<script setup>
import { reactive, ref } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import { authStore } from "@/stores/auth.js";

import JournalTypeContext from "./widget/JournalTypeContext.vue";
import DimensionEditor from '@/components/common/DimensionEditorVertical.vue';
import AccountSelector from '@/components/common/AccountSelector.vue';

layoutStore().name = "tenant";

const FEATUREID = 'AccountingJournalType'
const profile = authStore().getRBAC(FEATUREID)

const listControl = ref(null);

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
</script>