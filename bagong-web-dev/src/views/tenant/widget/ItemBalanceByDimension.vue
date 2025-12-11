<template>
  <div class="w-full">
    <s-grid
      v-model="data.records"
      ref="listControl"
      class="w-full"
      hide-sort
      sortField="Created"
      sortDirection="desc"
      :hide-new-button="true"
      :hide-detail="true"
      hide-refresh-button
      hide-delete-button
      auto-commit-line
      :config="data.gridCfg"
      form-keep-label
    >
      <template #header_search="{}">
        <div class="grow flex gap-3 justify-start">
          <s-input-builder
            v-model="data.filter.WarehouseID"
            useList
            label="Warehouse"
            lookup-url="/tenant/warehouse/find"
            lookup-key="_id"
            :lookup-labels="['Name', '_id']"
            :lookup-searchs="['_id', 'Name']"
            :query="data.queryParams"
            :multiple="false"
          ></s-input-builder>
          <s-input-builder
            v-model="data.filter.SectionID"
            useList
            label="Section"
            lookup-url="/tenant/section/find"
            lookup-key="_id"
            :lookup-labels="['Name', '_id']"
            :lookup-searchs="['_id', 'Name']"
            :query="data.queryParams"
            :multiple="false"
          ></s-input-builder>
          <s-input
            label="Balance date"
            kind="date"
            v-model="data.filter.BalanceDate"
          />
        </div>
      </template>
      <template #item_buttons_2="{ item }">
        <mdicon
          name="information"
          width="16"
          alt="history"
          class="cursor-pointer hover:text-primary"
          @click="
            historyTransaction(item);
            data.modalHistory = true;
          "
        />
      </template>
      <template #header_buttons_1="{}">
        <s-button
          icon="refresh"
          class="btn_primary refresh_btn"
          @click="refreshData"
        />
      </template>
    </s-grid>
    <s-modal
      :display="data.modalHistory"
      hideButtons
      title="Item Transaction"
      @beforeHide="data.modalHistory = false"
    >
      <s-card class="rounded-md w-full" hide-title no-gap>
        <div class="px-2 py-2">
          <loader v-if="data.loadingHistory" />
          <s-grid
            v-else
            ref="gridHistory"
            :config="data.gridCfgHistory"
            hide-detail
            hide-select
            hide-control
            hide-action
            hide-search
            hide-sort
            hide-refresh-button
            hide-save-button
            v-model="data.historyList"
          >
          </s-grid>
        </div>
      </s-card>
    </s-modal>
  </div>
</template>

<script setup>
import { reactive, ref, inject, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  SGrid,
  SInput,
  util,
  SButton,
  SModal,
  SCard,
  loadGridConfig,
} from "suimjs";
import SInputBuilder from "@/components/common/SInputBuilder.vue";
import Loader from "@/components/common/Loader.vue";

import moment from "moment";

const props = defineProps({
  item: { type: Object, default: () => {} },
});

layoutStore().name = "tenant";
const axios = inject("axios");

const listControl = ref(null);
const gridHistory = ref(null);

const data = reactive({
  filter: {
    ItemID: props.item._id,
    WarehouseID: "",
    SectionID: "",
    BalanceDate: "",
  },
  gridCfg: {},
  gridCfgHistory: {},
  modalHistory: false,
  historyList: [],
  records: [],
  queryParams: [],
  loadingHistory: false
});

function refreshData() {
  let param = data.filter;
  if (!data.filter.BalanceDate || data.filter.BalanceDate === 'Invalid date') {
    param = {
      WarehouseID: data.filter.WarehouseID,
      SectionID: data.filter.SectionID,
      ItemID: data.filter.ItemID,
    }
  }
  const url = "/scm/item/balance/gets-by-warehouse-and-section";
  axios.post(url, param).then(
    (r) => {
      listControl.value?.setRecords(r.data);
    },
    (e) => {
      util.showError(e);
    }
  );
}

function historyTransaction(item) {
  data.loadingHistory = true;
  let param = data.filter;
  param.WarehouseID =  item.WarehouseID;
  param.SectionID = item.SectionID;
  if (!data.filter.BalanceDate || data.filter.BalanceDate === 'Invalid date') {
    param = {
      WarehouseID: item.WarehouseID,
      SectionID: item.SectionID,
      ItemID: data.filter.ItemID,
    }
  }

  const url = "/scm/inventory/trx/gets-by-balance";
  axios.post(url, param).then(
    (r) => {
      data.historyList = r.data;
      data.loadingHistory = false;
    },
    (e) => {
      data.loadingHistory = false;
      util.showError(e);
    }
  );
}
onMounted(() => {
  loadGridConfig(axios, `/scm/item/balance/dimension/gridconfig`).then((r) => {
    data.gridCfg = r;
    refreshData();
  });
  loadGridConfig(axios, `/scm/inventory/trx/dimension/gridconfig`).then((r) => {
    data.gridCfgHistory = r;
  });
});
</script>
