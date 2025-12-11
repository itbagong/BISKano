<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Vendor"
      grid-config="/tenant/vendor/gridconfig"
      form-config="/tenant/vendor/formconfig"
      grid-read="/fico/vendorbalance/gets"
      form-read="/bagong/vendor/get"
      grid-mode="grid"
      grid-delete="/tenant/vendor/delete"
      form-keep-label
      form-insert="/bagong/vendor/save"
      form-update="/bagong/vendor/save"
      :grid-fields="['Enable']"
      :form-fields="['Dimension', 'Setting']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      :form-tabs-edit="[
        'General',
        'Detail',
        'Terms',
        'Contacts',
        'Bank',
        'Balance',
        'Transaction',
      ]"
      :form-tabs-view="[
        'General',
        'Detail',
        'Terms',
        'Contacts',
        'Bank',
        'Balance',
        'Transaction',
      ]"     
      @alterGridConfig="onAlterGridConfig"
      @formNewData="newRecord"
      @formEditData="editRecord"
      stay-on-form-after-save
      @form-field-change="onFormFieldChange"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_Setting="{ item, config,mode }">
        <setting-selector :label="config.label" v-model="item.Setting" :read-only="readOnly || mode == 'view'"/>
      </template>
      <template #form_input_Dimension="{ item,mode}">
        <dimension-editor
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :read-only="readOnly || mode == 'view'"
        ></dimension-editor>
      </template>

      <template #form_tab_Detail="{ item,mode }">
        <VendorDetail v-model="item.Detail" :read-only="readOnly || mode == 'view'"></VendorDetail>
      </template>
      <template #form_tab_Terms="{ item,mode }">
        <VendorTerms v-model="item.Detail.Terms" :read-only="readOnly || mode == 'view'"></VendorTerms>
      </template>
      <template #form_tab_Contacts="{ item ,mode}">
        <VendorContacts v-model="item.Detail.VendorContacts" :read-only="readOnly || mode == 'view'"></VendorContacts>
      </template>
      <template #form_tab_Bank="{ item ,mode }">
        <VendorBank v-model="item.Detail.VendorBank" :read-only="readOnly || mode == 'view'"></VendorBank>
      </template>
      <template #form_tab_Balance="{ item }">
        <transaction-balance
          url-gets="/fico/vendorbalance/gets"
          url-config="/tenant/vendor/gridconfig"
          :dim-list="['PC', 'CC', 'Site', 'Asset']"
          :jurnal-id="item._id"
        />
      </template>
      <template #form_tab_Transaction="{ item }">
        <transaction-history
          ref="transactionHistory"
          :jurnal-id="item._id"
          url="/fico/vendorbalance/get-transaction"
          hide-filter-warehouse
          hide-filter-section-id
          hide-filter-sku
        />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, watch, inject } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput } from "suimjs";
import { authStore } from "@/stores/auth.js";
import VendorDetail from "./widget/VendorDetail.vue";
import VendorTerms from "./widget/VendorTerms.vue";
import VendorContacts from "./widget/VendorContacts.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import VendorBank from "./widget/VendorBank.vue";
import TransactionHistory from "@/components/common/TransactionHistory.vue";
import TransactionBalance from "@/components/common/TransactionBalance.vue";

import SettingSelector from "@/components/common/SettingSelector.vue";

layoutStore().name = "tenant";

const FEATUREID = "Vendor";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const transactionHistory = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Setting = {};
  data.record = record;
  record.Setting = {};
  openForm(record);
}

function editRecord(r) {
  data.record = r;
  openForm(r);
}

function openForm(record) {
  util.nextTickN(2, () => {});
}

function onAlterGridConfig(cfg) {
  const CustomCol = ["Balance"];
  let dimColl = [];
  for (let i in CustomCol) {
    let dim = CustomCol[i];
    dimColl.push({
      field: dim,
      kind: "number",
      label: dim,
      readType: "show",
      labelField: "",
      input: {
        field: dim,
        label: dim,
        hint: "",
        hide: false,
        placeHolder: dim,
        kind: "number",
        disable: false,
        required: false,
        multiple: false,
      },
    });
  }
  cfg.fields = [...cfg.fields, ...dimColl];
}

function onFormFieldChange(name, v1, v2, old, record) {
  if (name == "GroupID" && data.record.GroupID !== null) {
    setFromGroupID(v1);
  }
}

function setFromGroupID(id) {
  const url = "/tenant/vendorgroup/get";
  axios.post(url, [id]).then(
    (r) => {
      data.record.DepositAccount = r.data.DepositAccount;
      data.record.MainBalanceAccount = r.data.MainBalanceAccount;
      data.record.Setting = r.data.Setting;
    },
    (e) => util.showError(e)
  );
}
</script>
