<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Ledger Account"
      grid-config="/tenant/ledgeraccount/gridconfig"
      form-config="/tenant/ledgeraccount/formconfig"
      form-read="/tenant/ledgeraccount/get"
      grid-read="/fico/ledgeraccountbalance/gets"
      grid-mode="grid"
      grid-delete="/tenant/ledgeraccount/delete"
      form-keep-label
      form-insert="/tenant/ledgeraccount/insert"
      form-update="/tenant/ledgeraccount/update"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @form-new-data="newRecord"
      @form-edit-data="openForm"
      :grid-custom-filter="customFilter"
      :grid-fields="['AccountType', 'Status']"
      :form-tabs-edit="['General', 'Balance', 'Transaction']"
      :form-tabs-view="['General', 'Balance', 'Transaction']"
      :form-fields="['Dimension']"
      @alterGridConfig="onAlterGridConfig"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          :default-list="profile.Dimension"
          v-model="item.Dimension"
          :read-only="mode == 'view'"
        ></dimension-editor-vertical>
      </template>
      <template #grid_header_search="{}">
        <s-input
          v-model="data.search.ID"
          label="Item"
          class="w-full"
          multiple
          use-list
          :lookup-url="`/tenant/ledgeraccount/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="onFilterRefresh"
        />
        <s-input
          v-model="data.search.AccountType"
          label="Account Type"
          class="w-full"
          multiple
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=ACT`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="onFilterRefresh"
        />
      </template>
      <template #form_tab_Balance="{ item }">
        <transaction-balance
          url-gets="/fico/ledgeraccountbalance/gets"
          url-config="/tenant/ledgeraccount/gridconfig"
          :dim-list="['PC', 'CC', 'Site', 'Asset']"
          :jurnal-id="item._id"
        />
      </template>
      <template #form_tab_Transaction="{ item }">
        <transaction-history
          ref="transactionHistory"
          url="/fico/ledgeraccountbalance/get-transaction"
          :jurnal-id="item._id"
          config-url="/fico/transaction_history/coa/gridconfig"
          hide-filter-warehouse
          hide-filter-section-id
          hide-filter-sku
        />
      </template>
      <!-- AccountType & Status -->
      <template #grid_AccountType="{ item }">
        <div v-if="item.AccountType">
          <s-input
            v-model="item.AccountType"
            hide-label
            use-list
            lookup-url="/tenant/masterdata/find?MasterDataTypeID=ACT"
            lookup-key="_id"
            :lookup-labels="['Name']"
            read-only
          />
        </div>
      </template>
      <template #grid_Status="{ item }">
        <div v-if="item.Status">
          <s-input
            v-model="item.Status"
            hide-label
            use-list
            lookup-url="/tenant/masterdata/find?MasterDataTypeID=SCOA"
            lookup-key="_id"
            :lookup-labels="['Name']"
            read-only
          />
        </div>
      </template>
      <!-- End Region -->
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import { DataList, SInput, util, SButton, SModal, SCard } from "suimjs";

import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import moment from "moment";
import TransactionHistory from "@/components/common/TransactionHistory.vue";
import TransactionBalance from "@/components/common/TransactionBalance.vue";

layoutStore().name = "tenant";

const FEATUREID = "ChartOfAccount";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");

const listControl = ref(null);
const transactionHistory = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  search: {
    ID: [],
    AccountType: [],
    Skip: 0,
    Take: 25,
  },
});

function onFilterRefresh() {
  util.nextTickN(2, () => {
    refreshData();
  });
}

function refreshData() {
  listControl.value.refreshList();
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

function newRecord(record) {
  resetFilterBalance();
}

function openForm(record) {
  record.Dimension = record.Dimension ? record.Dimension : [];
}

const customFilter = computed(() => {
  const filters = {
    ID: data.search.ID,
    AccountType: data.search.AccountType,
  };

  return filters;
});
</script>
