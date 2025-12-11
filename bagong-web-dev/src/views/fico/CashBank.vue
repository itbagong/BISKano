<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Cash & Bank"
      grid-config="/tenant/cashbank/gridconfig"
      grid-read="/fico/cashbankbalance/gets"
      form-config="/tenant/cashbank/formconfig"
      form-read="/tenant/cashbank/get"
      grid-mode="grid"
      grid-delete="/tenant/cashbank/delete"
      form-keep-label
      form-insert="/tenant/cashbank/insert"
      form-update="/tenant/cashbank/update"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @form-new-data="newRecord"
      @form-edit-data="openForm"
      :form-tabs-edit="['General', 'Balance', 'Transaction']"
      :form-tabs-view="['General', 'Balance', 'Transaction']"
      :form-fields="['Dimension']"
      grid-hide-refresh
      @alterGridConfig="onAlterGridConfig"
      @alter-form-config="alterFormConfig"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="data.customFilter"
      @gridResetCustomFilter="resetGridHeaderFilter"
    >
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilter"
          v-model="data.customFilter"
          @initNewItem="initNewItemFilter"
          @preChange="changeFilter"
          @change="refreshGrid"
        >
          <template #filters="{ item }">
            <div class="w-full flex gap-3 items-center">
              <s-input
                class="w-full"
                keep-label
                label="Search"
                v-model="item.Keyword"
              />
              <s-input
                class="w-full"
                label="Cash bank group ID"
                use-list
                lookup-url="/tenant/cashbankgroup/find"
                lookup-key="_id"
                :lookup-labels="['Name']"
                :lookupSearchs="['_id', 'Name']"
                v-model="item.CashBankGroupIDs"
                multiple
              />
            </div>
          </template>
        </grid-header-filter>
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :read-only="mode == 'view'"
        ></dimension-editor>
      </template>

      <template #form_tab_Balance="{ item }">
        <transaction-balance
          url-gets="/fico/cashbankbalance/gets"
          url-config="/tenant/cashbank/gridconfig"
          :jurnal-id="item._id"
        />
      </template>
      <template #form_tab_Transaction="{ item }">
        <transaction-history
          ref="transactionHistory"
          url="/fico/cashbankbalance/get-transaction"
          config-url="/fico/transaction_history/cashbank/gridconfig"
          :jurnal-id="item._id"
          hide-filter-warehouse
          hide-filter-section-id
          hide-filter-sku
        />
      </template>
      <template #grid_header_buttons_1="{}">
        <s-button
          icon="refresh"
          class="btn_primary refresh_btn"
          @click="refreshData"
        />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SInput, util, SButton, SModal, SCard } from "suimjs";
import { authStore } from "@/stores/auth.js";
import { useRouter, useRoute } from "vue-router";

import DimensionEditor from "@/components/common/DimensionEditor.vue";
import moment from "moment";
import TransactionHistory from "@/components/common/TransactionHistory.vue";
import TransactionBalance from "@/components/common/TransactionBalance.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";

layoutStore().name = "tenant";

const FEATUREID = "CashAndBankModule";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const router = useRouter();
const route = useRoute();

const listControl = ref(null);
const transactionHistory = ref(null);
const gridHeaderFilter = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  search: {
    ID: [],
    Skip: 0,
    Take: 25,
  },
  gridCfg: {},
  customFilter: null,
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
  data.gridCfg = cfg;
}

function alterFormConfig(config) {
  if (route.query.id !== undefined) {
    let getUrlParam = route.query.id;
    listControl.value.selectData({ _id: getUrlParam }); //remark sementara tunggu suimjs update
    router.replace({ path: "/fico/CashBank" });
  } else if (route.query.form == 1) {
    listControl.value.newData();
    router.replace({ path: "/fico/CashBank" });
  }
}
function newRecord(record) {}

function openForm(record) {}

function initNewItemFilter(item) {
  item.Keyword = "";
  item.CashBankGroupIDs = [];
}

function changeFilter(item, filters) {
  if (item.Keyword.length > 0) {
    const filterKeyword = [];
    filterKeyword.push({
      Op: "$contains",
      Field: "_id",
      Value: [item.Keyword],
    });
    filterKeyword.push({
      Op: "$contains",
      Field: "Name",
      Value: [item.Keyword],
    });
    filters.push({
      Op: "$or",
      Items: filterKeyword,
    })
  }
  if (item.CashBankGroupIDs.length > 0) {
    filters.push({
      Op: "$in",
      Field: "CashBankGroupID",
      Value: item.CashBankGroupIDs,
    });
  }
}

function refreshGrid() {
  listControl.value.refreshGrid();
}

function resetGridHeaderFilter() {
  gridHeaderFilter.value.reset();
}
</script>

<style>
.flex.gap-\[1px\].header_button {
  margin-top: 12px;
}
</style>
