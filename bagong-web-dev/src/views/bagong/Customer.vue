<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Customer"
      grid-config="/tenant/customer/gridconfig"
      form-config="/tenant/customer/formconfig"
      grid-read="/fico/customerbalance/gets"
      form-read="/bagong/customer/get"
      grid-mode="grid"
      grid-delete="/tenant/customer/delete"
      form-keep-label
      form-insert="/bagong/customer/save"
      form-update="/bagong/customer/save"
      :grid-fields="['Enable']"
      :form-tabs-edit="[
        'General',
        'Detail',
        'Configuration',
        'Contact',
        'Balance',
        'Transaction',
      ]"
      :form-tabs-view="[
        'General',
        'Detail',
        'Configuration',
        'Contact',
        'Balance',
        'Transaction',
      ]"
      :form-fields="['Setting', 'Dimension', 'Contacts', 'IsActive']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @alterGridConfig="onAlterGridConfig"
      @formNewData="newRecord"
      @formEditData="editRecord"
      stay-on-form-after-save
      @form-field-change="onFormFieldChange"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @pre-save="onPreSave"
    >
      <template #form_input_IsActive="{ item, config,mode }">
        <s-input
          kind="text"
          v-model="item.Detail.DefaultPriceBook"
          class="w-full mb-5"
          keep-label
          label="Default Price Book"
          use-list
          multiple
          lookup-url="/sdp/salespricebook/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
           :read-only="mode == 'view'"
        />
        <s-input
          :kind="config.kind"
          v-model="item.IsActive"
          class="w-full"
          :label="config.label"
           :read-only="mode == 'view'"
        />
      </template>
      <template #form_input_Setting="{ item, config }">
        <setting-selector :label="config.label" v-model="item.Setting" />
      </template>
      <template #form_input_Dimension="{ item,mode }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :read-only="mode == 'view'"
        ></dimension-editor-vertical>
      </template>

      <template #form_tab_Detail="{ item,mode }">
        <CustomerDetail v-model="item.Detail" 
          :read-only="mode == 'view'"></CustomerDetail>
      </template>
      <template #form_tab_Configuration="{ item,mode }">
        <CustomerConfiguration v-model="item.Config" 
          :read-only="mode == 'view'"></CustomerConfiguration>
      </template>
      <template #form_tab_Contact="{ item,mode }">
        <CustomerContact
          :read-only="mode == 'view'"
          v-model="item.Contacts"
          :item="item"
          grid-config="/tenant/contact/gridconfig"
          form-config="/tenant/contact/formconfig"
        ></CustomerContact>
      </template>
      <template #form_tab_Balance="{ item }">
        <transaction-balance
          url-gets="/fico/customerbalance/gets"
          url-config="/tenant/customer/gridconfig"
          :dim-list="['PC', 'CC', 'Site', 'Asset']"
          :jurnal-id="item._id"
        />
      </template>
      <template #form_tab_Transaction="{ item }">
        <transaction-history
          ref="transactionHistory"
          url="/fico/customerbalance/get-transaction"
          :jurnal-id="item._id"
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

import CustomerDetail from "./widget/CustomerDetail.vue";
import CustomerConfiguration from "./widget/CustomerConfiguration.vue";
import CustomerContact from "./widget/CustomerContact.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import TransactionHistory from "@/components/common/TransactionHistory.vue";
import TransactionBalance from "@/components/common/TransactionBalance.vue";
import SettingSelector from "@/components/common/SettingSelector.vue";

layoutStore().name = "tenant";

const FEATUREID = "Customer";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const transactionHistory = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  record: {},
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.IsActive = true;
  record.Detail = {};
  record.Setting = {};
  data.record = record;
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
  const url = "/tenant/customergroup/get";
  axios.post(url, [id]).then(
    (r) => {
      data.record.Setting = r.data.Setting;
    },
    (e) => util.showError(e)
  );
}

function onPreSave(record) {
  if (record.Contacts && record.Contacts.length > 0) {
    for (let i = 0; i < record.Contacts.length; i++) {
      const contact = record.Contacts[i];
      if (contact.AsContactPerson) {
        record.Detail.PersonalContact = contact.Name;
        record.Detail.Email = contact.Email;
        record.Detail.MobilePhoneNo = contact.PhoneNumber;
        record.Detail.BusinessPhoneNo = contact.BusinessPhoneNo;
      }
    }
  }
  // console.log(record);
}
</script>
