<template>
  <div class="w-full">
    <!-- <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/tenant/itemspec/search/gridconfig"
      form-config="/tenant/itemspec/search/formconfig"
      grid-read="/tenant/itemspec/search"
      grid-mode="grid"
      form-keep-label
      form-hide-submit
      grid-hide-detail
      grid-hide-select
      grid-hide-delete
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :grid-fields="[]"
      :form-fields="[]"
      grid-hide-new
      grid-hide-edit
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
    >
    </data-list> -->
    <!-- readUrl="/tenant/itemspec/search" -->
    <s-card
      :title="data.titleForm"
      class="w-full bg-white suim_datalist card"
      :no-gap="false"
      :hide-title="false"
    >
      <s-grid
        v-model="data.value"
        ref="listControl"
        class="w-full"
        :config="data.gridCfg"
        hide-sort
        sortField="LastUpdate"
        sortDirection="desc"
        hide-select
        hide-new-button
        hide-action
        hide-paging
        auto-commit-line
        form-keep-label
      >
        <template #header_search="{ config }">
          <s-input
            ref="refItemID"
            v-model="data.search.Keyword"
            label="search keyword"
            class="w-full"
            hide-label
            @keyup.enter="onFilterRefresh"
          ></s-input>
        </template>
        <template #header_buttons="{ config }">
          <s-button
            icon="refresh"
            class="btn_primary refresh_btn"
            @click="onFilterRefresh"
          />
        </template>
        <template #paging>
          <s-pagination
            :recordCount="data.value.length"
            :pageCount="pageCount"
            :current-page="data.paging.currentPage"
            :page-size="data.paging.pageSize"
            @changePage="changePage"
            @changePageSize="changePageSize"
          ></s-pagination>
        </template>
      </s-grid>
    </s-card>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
import {
  DataList,
  SGrid,
  SCard,
  util,
  loadGridConfig,
  SInput,
  SButton,
  SPagination,
} from "suimjs";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";
const featureID = "SearchItem";
const profile = authStore().getRBAC(featureID);
const listControl = ref(null);
const axios = inject("axios");
const pageCount = computed({
  get() {
    return Math.ceil(data.count / data.paging.pageSize);
  },
});

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Search Item",
  loading: false,
  count: 0,
  value: [],
  gridCfg: {},
  search: {
    Keyword: "",
  },
  paging: {
    skip: 0,
    pageSize: 25,
    currentPage: 1,
  },
});

function onFilterRefresh(val) {
  data.paging = {
    skip: 0,
    pageSize: 25,
    currentPage: 1,
  };
  util.nextTickN(2, () => {
    refreshData();
  });
}

function changePageSize(pageSize) {
  data.paging.pageSize = pageSize;
  data.paging.currentPage = 1;
  refreshData();
}
function changePage(page) {
  data.paging.currentPage = page;
  refreshData();
}

function refreshData() {
  const payload = {
    ...data.search,
    Skip: (data.paging.currentPage - 1) * data.paging.pageSize,
    Take: data.paging.pageSize,
    Sort: ["-LastUpdate"],
  };
  listControl.value.setLoading(true);
  axios.post("/tenant/itemspec/search", payload).then(
    (r) => {
      util.nextTickN(2, () => {
        console.log(r.data);
        data.count = r.data.count;
        data.value = r.data.data;
        listControl.value.setLoading(false);
      });
    },
    (e) => {
      util.showError(e);
    }
  );
}

onMounted(() => {
  loadGridConfig(axios, `/tenant/itemspec/search/gridconfig`).then(
    (r) => {
      data.gridCfg = r;
      refreshData();
    },
    (e) => util.showError(e)
  );
});
</script>
