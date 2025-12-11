<template>
  <div class="w-full">
    <s-grid
      ref="listControl"
      :read-url="urlGets"
      hide-select
      hide-search
      hide-sort
      hide-refresh-button
      hide-new-button
      hide-action
      :config="data.cfg"
      :custom-filter="customFiler"
    >
      <template #header_search>
        <div class="w-full flex gap-2 items-center justify-center">
          <s-input
            class="w-full"
            kind="date"
            label="Balance as of"
            v-model="data.filter.DateFrom"
            @change="onFilterRefresh"
          />
          <s-input
            class="w-full"
            kind="input"
            label="Dimension"
            use-list
            multiple
            :items="dimList"
            v-model="data.filter.Dimension"
            @change="onFilterRefresh"
          />
          <div class="suim_input mt-[12px]">
            <s-button class="btn_primary" icon="refresh" @click="refreshData" />
          </div>
        </div>
      </template>
      <!-- AccountType & Status -->
      <template #item_AccountType="{ item }">
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
      <template #item_Status="{ item }">
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
      <template #item_Qty="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.Qty) }}
        </div>
      </template>
      <template #item_QtyReserved="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.QtyReserved) }}
        </div>
      </template>
      <template #item_QtyPlanned="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.QtyPlanned) }}
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
      <template #item_TrxQty="{ item }">
        <div class="text-right">
          {{ helper.formatNumberWithDot(item.TrxQty) }}
        </div>
      </template>
      <!-- End Region -->
      <template #loading>
        <loader kind="skeleton" skeleton-kind="list" />
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed } from "vue";
import { SGrid, util, SInput, SButton, loadGridConfig } from "suimjs";
import Loader from "@/components/common/Loader.vue";
import helper from "@/scripts/helper.js";

const axios = inject("axios");

const listControl = ref(null);

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  transactionType: { type: String, default: "Financial" },
  urlGets: { type: String, default: "" },
  urlConfig: { type: String, default: "" },
  dimList: { type: Array, default: ["PC", "CC", "Site", "Asset"] },
  jurnalId: { type: String, default: "" },
});

const data = reactive({
  cfg: {},
  filter: {
    ID: [props.jurnalId],
    Dimension: [],
    DateFrom: null,
  },
});
const customFiler = computed({
  get() {
    return {
      ...data.filter,
      DateFrom: helper.formatFilterDate(data.filter.DateFrom),
    };
  },
});

function refreshData() {
  shoHideDimension();
  let param = JSON.parse(JSON.stringify(data.filter));
  if (param.ID.length == 0) return;
  listControl.value.refreshData();
}

function shoHideDimension() {
  let src = data.filter.Dimension;
  for (let i in props.dimList) {
    let v = props.dimList[i];
    let f = data.cfg.fields.find((o) => o.field == v);
    f.readType = src.includes(v) ? "show" : "hide";
  }
}

function onFilterRefresh() {
  util.nextTickN(2, () => {
    refreshData();
  });
}

function onAlterGridConfig(cfg) {
  if (props.transactionType === "Financial") {
    const CustomCol = [...props.dimList, "Balance"];
    const numFields = ["Balance"];
    let fieldShow = [...data.filter.Dimension, "Balance"];
    let dimColl = [];
    for (let i in CustomCol) {
      let dim = CustomCol[i];
      dimColl.push({
        field: dim,
        kind: numFields.includes(dim) ? "number" : "text",
        label: dim,
        readType: dim.includes(fieldShow) ? "show" : "hide",
        labelField: "",
        input: {
          field: dim,
          label: dim,
          hint: "",
          hide: false,
          placeHolder: dim,
          kind: numFields.includes(dim) ? "number" : "text",
          disable: false,
          required: false,
          multiple: false,
        },
      });
    }
    cfg.fields = [...cfg.fields, ...dimColl];
  } else {
    const CustomCol = [...props.dimList];
    let src = data.filter.Dimension;
    let dimColl = [];
    for (let i in CustomCol) {
      let dim = CustomCol[i];
      dimColl.push({
        field: dim,
        kind: "text",
        label: dim,
        readType: src.includes(CustomCol[i]) ? "show" : "hide",
        labelField: "",
        input: {
          field: dim,
          label: dim,
          hint: "",
          hide: false,
          placeHolder: dim,
          kind: "text",
          disable: false,
          required: false,
          multiple: false,
        },
      });
    }
    cfg.fields = [...dimColl, ...cfg.fields];
  }

  data.cfg = cfg;
}

onMounted(() => {
  loadGridConfig(axios, props.urlConfig).then(
    (r) => {
      onAlterGridConfig(r);
    },
    (e) => util.showError(e)
  );
});
</script>
