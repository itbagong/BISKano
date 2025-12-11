<template>
  <div class="w-full bg-white p-3">
    <div class="grid grid-cols-4 gap-6 mb-3 pb-3 border-b">
      <s-input
        :disabled="data.loadingGenrate"
        label="No Faktur"
        v-model="data.param.prefix"
      />
      <s-input
        :disabled="data.loadingGenrate"
        label="From"
        v-model="data.param.from"
        kind="number"
      />
      <s-input
        :disabled="data.loadingGenrate"
        label="To"
        v-model="data.param.to"
        kind="number"
      />
      <div class="mt-2 flex">
        <s-button
          :disabled="data.loadingGenrate"
          class="btn_primary"
          @click="generateTax"
          label="Generate"
        />
      </div>
    </div>
    <div v-if="data.loadingGenrate">
      <loader kind="skeleton" skeleton-kind="list" />
      <loader kind="skeleton" skeleton-kind="list" />
    </div>
    <data-list
      v-else
      no-gap
      ref="listControl"
      title=""
      grid-config="/bagong/taxinvoice/gridconfig"
      form-config="/bagong/taxinvoice/formconfig"
      grid-read="/bagong/taxinvoice/gets"
      form-read="/bagong/taxinvoice/get"
      grid-mode="grid"
      grid-delete="/bagong/taxinvoice/delete"
      form-insert="/bagong/taxinvoice/insert"
      form-update="/bagong/taxinvoice/update"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :grid-custom-filter="customFilter"
      gridHideNew
      gridHideSelect
      gridHideDelete
    >
      <template #grid_header_search>
        <s-input
          class="w-full"
          hide-label
          label="enter search keyword"
          v-model="data.keyword"
          @keyup.enter="refreshGrid"
        ></s-input>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SInput, SButton, util } from "suimjs";
import { useRoute } from "vue-router";
import { authStore } from "@/stores/auth";

import Loader from "@/components/common/Loader.vue";

layoutStore().name = "tenant";

const FEATUREID = "TaxTransaction";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");

const listControl = ref(null);

const data = reactive({
  loadingGenrate: false,
  param: {
    prefix: "",
    from: 0,
    to: 0,
  },
  appMode: "grid",
  formMode: "edit",
  keyword: null,
});

function generateTax() {
  data.loadingGenrate = true;
  axios
    .post("/bagong/invoice/generate-tax-invoice", { ...data.param })
    .then(
      (r) => {},
      (err) => {
        data.showGrid = false;
        util.showError(err);
      }
    )
    .finally(() => {
      data.loadingGenrate = false;
    });
}

function refreshGrid() {
  listControl.value.refreshGrid();
}

const customFilter = computed(() => {
  const filters = [];
  if (data.keyword != null) {
    filters.push({
      Op: "$contains",
      Field: "FPNo",
      Value: [data.keyword],
    });
  }

  if (filters.length == 1) {
    return filters[0];
  } else if (filters.length > 1) {
    return { Op: "$and", Items: filters };
  } else {
    return null;
  }
});
</script>
