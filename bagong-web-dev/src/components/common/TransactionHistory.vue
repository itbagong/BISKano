<template>
  <div class="w-full">
    <s-grid
      ref="listControl"
      hide-select
      hide-search
      hide-sort
      hide-refresh-button
      hide-new-button
      hide-action
      :config="data.cfg"
    >
      <template #header_search>
        <div class="w-full">
          <div class="flex gap-2 mb-4">
            <s-input
              v-if="!hideFilterDate"
              class="min-w-[200px]"
              kind="date"
              label="Date From"
              v-model="data.filter.DateFrom"
              @change="onRefreshData"
            />
            <s-input
              v-if="!hideFilterDate"
              class="min-w-[200px]"
              kind="date"
              label="Date To"
              v-model="data.filter.DateTo"
              @change="onRefreshData"
            />
            <s-input
              v-if="!hideFilterTrxType"
              label="Transaction type"
              kind="text"
              class="min-w-[200px]"
              v-model="data.filter.TrxType"
              :allow-add="false"
              use-list
              :items="data.itemListTrxType"
              @change="onRefreshData"
            />
            <slot
              name="filter_finance"
              :item="data.filter"
              v-if="!hideDimFinance"
            >
              <dim-editor
                multiple
                v-model="data.filter.Finance"
                :required-fields="[]"
                :dim-names="dimFinance"
                class="w-full"
                @change="onRefreshData"
              />
            </slot>
            <slot name="filter-inventory">
              <s-input
                v-if="!hideFilterWarehouse"
                class="w-full"
                label="WarehouseID"
                use-list
                multiple
                lookup-url="/tenant/warehouse/find"
                lookup-key="_id"
                :lookup-labels="['_id', 'Name']"
                :lookup-searchs="['_id', 'Name']"
                v-model="data.filter.DimInventory.WarehouseID"
                @change="onRefreshData"
              />
              <s-input
                v-if="!hideFilterSectionId"
                class="w-full"
                label="SectionID"
                multiple
                use-list
                lookup-url="/tenant/section/find"
                lookup-key="_id"
                :lookup-labels="['_id', 'Name']"
                :lookup-searchs="['_id', 'Name']"
                v-model="data.filter.DimInventory.SectionID"
                @change="onRefreshData"
              />
              <s-input
                class="w-full"
                v-if="!hideFilterSku"
                label="SKU"
                use-list
                multiple
                :lookup-url="`/tenant/item/gets-sku-detail-by-item?ItemID=${param._id}`"
                lookup-key="ID"
                :lookup-labels="['Text']"
                :lookup-searchs="['Text']"
                v-model="data.filter.DimInventory.SKU"
                @change="onRefreshData"
              />
            </slot>
            <slot name="filters" :item="data.filter"></slot>
            <div
              class="w-full"
              v-if="
                hideDimFinance &&
                hideFilterWarehouse &&
                hideFilterSectionId &&
                hideFilterSku
              "
            >
              &nbsp;
            </div>
            <div class="suim_input mt-[12px]">
              <s-button
                class="btn_primary"
                icon="refresh"
                @click="refreshData"
              />
            </div>
          </div>
          <div class="grid grid-cols-2 font-semibold border-solid border-[1px]">
            <div
              class="grid grid-cols-2 text-center border-solid border-r-[1px]"
            >
              <div class="col-span-2 py-1 border-b-[1px]">Opening</div>
              <div class="py-1">
                Balance : {{ util.formatMoney(data.record.Opening?.Balance) }}
              </div>
              <div class="py-1">
                Date :
                {{
                  data.record.Opening?.Date
                    ? moment(data.record.Opening.Date).format("DD/MM/YYYY")
                    : "-"
                }}
              </div>
            </div>
            <div class="grid grid-cols-2 text-center">
              <div class="col-span-2 py-1 border-b-[1px]">Closing</div>
              <div class="py-1">
                Balance : {{ util.formatMoney(data.record.Closing?.Balance) }}
              </div>
              <div class="py-1">
                Date :
                {{
                  data.record.Closing?.Date
                    ? moment(data.record.Closing.Date).format("DD/MM/YYYY")
                    : "-"
                }}
              </div>
            </div>
          </div>
        </div>
      </template>
      <template #item_SourceJournalID="{ item }">
        <a
          href="#"
          class="text-blue-400 hover:text-blue-800"
          @click="
            redirect(
              item.SourceJournalID,
              transactionType == 'Financial' ? item.TrxType : item.SourceTrxType
            )
          "
          >{{ item?.SourceJournalID }}
        </a>
      </template>
      <template #item_Qty="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.Qty) }}
        </div>
      </template>
      <template #item_TrxQty="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.TrxQty) }}
        </div>
      </template>
      <template #item_AmountPhysical="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.AmountPhysical) }}
        </div>
      </template>
      <template #item_AmountFinancial="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.AmountFinancial) }}
        </div>
      </template>
      <template #item_AmountAdjustment="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.AmountAdjustment) }}
        </div>
      </template>
      <template #item_Amount="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.Amount) }}
        </div>
      </template>
      <template #loading>
        <loader kind="skeleton" skeleton-kind="list" />
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, watch, nextTick } from "vue";
import { useRouter, useRoute } from "vue-router";
import { SGrid, SInput, SButton, util, loadGridConfig } from "suimjs";
import Loader from "@/components/common/Loader.vue";
import moment from "moment";
import helper from "@/scripts/helper.js";
import DimEditor from "@/components/common/DimensionEditor.vue";

const router = useRouter();
const axios = inject("axios");
const listControl = ref(null);

const props = defineProps({
  param: { type: Object, default: {} },
  transactionType: { type: String, default: "Financial" },
  configUrl: { type: String, default: "/fico/transaction_history/gridconfig" },
  url: { type: String, default: "" },
  hideFilterDate: { type: Boolean, default: false },
  hideFilterTrxType: { type: Boolean, default: false },
  hideFilterWarehouse: { type: Boolean, default: false },
  hideFilterSectionId: { type: Boolean, default: false },
  hideFilterSku: { type: Boolean, default: false },
  hideDimFinance: { type: Boolean, default: false },
  dimFinance: { type: Array, default: () => ["PC", "CC", "Site", "Asset"] },
  jurnalId: { type: String, default: "" },
});
const data = reactive({
  appMode: "grid",
  record: {},
  cfg: {},
  itemListTrxType:
    props.transactionType === "Financial"
      ? ["CASH IN", "SUBMISSION CASH IN", "CASH OUT", "SUBMISSION CASH OUT"]
      : ["Purchase Order", "Movement In", "Movement Out"],
  paramFilter: props.param,
  filter: {
    DateFrom: null,
    DateTo: null,
    TrxType: "",
    DimInventory: {
      WarehouseID: [],
      SectionID: [],
      SKU: [],
    },
  },
});

function refreshData() {
  listControl.value.setLoading(true);
  const url = props.url;
  let param = formatDim(
    JSON.parse(
      JSON.stringify({
        ...data.filter,
        DateFrom: helper.formatFilterDate(data.filter.DateFrom),
        DateTo: helper.formatFilterDate(data.filter.DateTo),
        ID: [props.jurnalId],
      })
    )
  );
  axios.post(url, param).then(
    (r) => {
      data.record = r.data;
      listControl.value.setLoading(false);
      listControl.value.setRecords(r.data.Transaction);
    },
    (e) => {
      listControl.value.setLoading(false);
      util.showError(e);
    }
  );
}
function formatDim(params) {
  let fin = {};
  for (let i in params.Finance) {
    let o = params.Finance[i];
    fin[o.Key] = o.Value;
  }
  params.DimFinance = fin;
  delete params.Finance;
  return params;
}

function onRefreshData() {
  util.nextTickN(2, () => {
    refreshData();
  });
}
function redirect(trxId, trxType) {
  let name = "";
  let query = {};
  switch (trxType) {
    case "CASH IN":
      name = "bagong-CashIn";
      query = { trxid: trxId };
      break;
    case "SUBMISSION CASH IN":
      name = "bagong-CashIn";
      query = { trxid: trxId, id: "SubmissionCashIn" };
      break;
    case "CASH OUT":
      name = "bagong-CashOut";
      query = { trxid: trxId };
      break;
    case "SUBMISSION CASH OUT":
      name = "bagong-CashOut";
      query = { trxid: trxId, id: "SubmissionCashOut" };
      break;
    case "Purchase Order":
      name = "scm-PurchaseOrder";
      query = { trxid: trxId };
      break;
    case "Movement In":
      name = "scm-InventoryJournal";
      query = { trxid: trxId, type: "Movement In", title: "Movement In" };
      break;
    case "Movement Out":
      name = "scm-InventoryJournal";
      query = { trxid: trxId, type: "Movement Out", title: "Movement Out" };
      break;
    case "Stock Opname":
      name = "scm-InventoryAdjustment";
      query = { trxid: trxId };
      break;
    default:
      break;
  }
  if (name !== "") {
    const url = router.resolve({ name: name, query: query });
    window.open(url.href, "_blank");
  }
}
defineExpose({
  refreshData,
});

onMounted(() => {
  loadGridConfig(axios, props.configUrl).then(
    (r) => {
      data.cfg = r;
    },
    (e) => util.showError(e)
  );
});
</script>
