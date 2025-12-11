<template>
  <div>
    <s-card
      class="w-full bg-white grid_card"
      hide-footer
      v-if="!data.openNonAsset"
    >
      <template v-if="data.loadingGridCfg || data.loadingSelectData">
        <slot name="loader">
          <div class="loader"></div>
        </slot>
      </template>
      <template #title>
        <div class="card_title grow flex gap-2 items-center">
          <mdicon
            name="arrow-left"
            size="28"
            class="cursor-pointer"
            @click="emit('back')"
          />
          {{ props.title }}
        </div>
      </template>

      <!-- -->
      <s-grid
        ref="gridCtl"
        class="w-full"
        hideNewButton
        :hide-edit="!profile.canUpdate"
        hide-delete-button
        delete-url="/bagong/siteentry_asset/delete"
        total-url="/bagong/siteentry/get-total-site-entry-asset"
        v-if="data.gridDetailCfg.setting"
        :config="data.gridDetailCfg"
        form-keep-label
        @select-data="selectDataDetail"
        @grid-refreshed="gridRefreshed"
        hideSort
        hideRefreshButton
      >
        <template #header_search="{ props, item }">
          <div class="flex gap-2 w-full">
            <s-input
               v-if="filters.includes('SearchText')"
              class="w-[200px] filter-text"
              label="Police No | Hull No"
              v-model="data.query.SearchText"
              @change="changeFilter"
            />
            <s-input
              v-if="filters.includes('PoliceNum')"
              multiple
              label="Police No"
              class="w-[360px]"
              v-model="data.query.PoliceNum"
              use-list
              :items="filter.PoliceNum"
              @change="changeFilter"
            >
            </s-input>
            <s-input
              v-if="filters.includes('TrayekName')"
              multiple
              label="Trayek Name"
              class="w-[360px]"
              v-model="data.query.TrayekName"
              use-list
              :items="filter.TrayekName"
              @change="changeFilter"
            >
            </s-input>

            <s-input
              v-if="filters.includes('CustomerName')"
              multiple
              label="Customer Name"
              class="w-[360px]"
              v-model="data.query.CustomerName"
              use-list
              :items="filter.CustomerName"
              @change="changeFilter"
            >
            </s-input>
          </div>
        </template>
        <template #header_buttons>
          <s-button
            icon="refresh"
            class="btn_primary refresh_btn"
            tooltip="Sync"
            label="Sync"
            @click="syncData"
          />
        </template>
        <template #grid_total="{ item }">
          <tr class="font-semibold">
            <td colspan="8" class="ml-4">Sub Total</td>
            <td class="text-right">
              {{ util.formatMoney(data.totalAsset.Income, {}) }}
            </td>
            <td class="text-right">
              {{ util.formatMoney(data.totalAsset.Expense, {}) }}
            </td>
            <td class="text-right" v-if="props.purpose == 'Trayek'">
              {{ util.formatMoney(data.totalAsset.Bonus, {}) }}
            </td>
            <td class="text-right">
              {{ util.formatMoney(data.totalAsset.Revenue, {}) }}
            </td>
            <td></td>
            <td></td>
          </tr>
          <tr class="font-semibold" @click="onOpenNonAsset()">
            <td colspan="8" class="ml-4">Non Asset</td>
            <td class="text-right">
              <loader
                kind="skeleton"
                skeleton-kind="input"
                v-if="data.loadingTotalNonAsset"
              />
              <template v-else
                >{{ util.formatMoney(data.totalNonAsset.Income, {}) }}
              </template>
            </td>
            <td class="text-right">
              <loader
                kind="skeleton"
                skeleton-kind="input"
                v-if="data.loadingTotalNonAsset"
              />
              <template v-else
                >{{ util.formatMoney(data.totalNonAsset.Expense, {}) }}
              </template>
            </td>
            <td class="text-right" v-if="props.purpose == 'Trayek'">
              <loader
                kind="skeleton"
                skeleton-kind="input"
                v-if="data.loadingTotalNonAsset"
              />
              <template v-else
                >{{ util.formatMoney(data.totalNonAsset.Bonus, {}) }}
              </template>
            </td>
            <td class="text-right">
              <loader
                kind="skeleton"
                skeleton-kind="input"
                v-if="data.loadingTotalNonAsset"
              />
              <template v-else
                >{{ util.formatMoney(data.totalNonAsset.Revenue, {}) }}
              </template>
            </td>
            <td></td>
            <td></td>
          </tr>
          <tr class="font-semibold">
            <td colspan="8" class="ml-4">Total</td>
            <td class="text-right">
              <loader
                kind="skeleton"
                skeleton-kind="input"
                v-if="data.loadingTotalNonAsset"
              />
              <template v-else
                >{{ util.formatMoney(summaaryTotal.Income, {}) }}
              </template>
            </td>
            <td class="text-right">
              <loader
                kind="skeleton"
                skeleton-kind="input"
                v-if="data.loadingTotalNonAsset"
              />
              <template v-else
                >{{ util.formatMoney(summaaryTotal.Expense, {}) }}
              </template>
            </td>
            <td class="text-right" v-if="props.purpose == 'Trayek'">
              <loader
                kind="skeleton"
                skeleton-kind="input"
                v-if="data.loadingTotalNonAsset"
              />
              <template v-else
                >{{ util.formatMoney(summaaryTotal.Bonus, {}) }}
              </template>
            </td>
            <td class="text-right">
              <loader
                kind="skeleton"
                skeleton-kind="input"
                v-if="data.loadingTotalNonAsset"
              />
              <template v-else
                >{{ util.formatMoney(summaaryTotal.Revenue, {}) }}
              </template>
            </td>
            <td></td>
            <td></td>
          </tr>
        </template>
        <template #paging="{}">
          <s-pagination
            :recordCount="pagination.recordCount"
            :pageCount="pagination.pageCount"
            :current-page="pagination.currentPage"
            :page-size="pagination.pageSize"
            @changePage="changePage"
            @changePageSize="changePageSize"
          ></s-pagination>
        </template>
      </s-grid>
    </s-card>

    <non-asset
      v-else
      :non-asset-id="siteEntryId"
      :siteID="props.site"
      ref="pageNonAsset"
      :group-id-value="props.groupIdValue"
      @close="closeNonAsset"
    ></non-asset>
  </div>
</template>
<script setup>
import { reactive, onMounted, inject, ref, nextTick, computed } from "vue";
import {
  SCard,
  SGrid,
  SForm,
  loadGridConfig,
  loadFormConfig,
  util,
  SButton,
  SModal,
  SPagination,
  SInput,
} from "suimjs";

import helper from "@/scripts/helper.js";
import NonAsset from "./GridNonAsset.vue";
import Loader from "@/components/common/Loader.vue";

const axios = inject("axios");
const gridCtl = ref(null);
const pageNonAsset = ref(null);

const props = defineProps({
  title: { type: String, default: "" },
  siteEntryId: { type: String, default: "" },
  site: { type: String, default: "" },
  purpose: { type: String, default: "" },
  profile: { type: Object, default: {} },
  groupIdValue: { type: Array, default: () => [] },
  filters: {
    type: Array,
    default: ["SearchText", "PoliceNum", "TrayekName", "CustomerName"],
  },
});

const mapUrl = {
  Mining: "siteentry_miningdetail",
  Trayek: "siteentry_trayekdetail",
  BTS: "siteentry_btsdetail",
};

const labelTotal = {
  0: "SubTotal",
  1: "Non Asset",
  2: "Total",
};

const data = reactive({
  loadingGridCfg: false,
  loadingSelectData: false,
  gridDetailCfg: {},
  appMode: "grid",
  records: [],
  record: {},
  openNonAsset: false,
  disableNonAsset: false,
  originRecords: [],
  query: {
    SearchText: '',
    PoliceNum: [],
    TrayekName: [],
    CustomerName: [],
  },
  totalNonAsset: {
    Income: 0,
    Expense: 0,
    Bonus: 0,
    Revenue: 0,
  },
  totalAsset: {
    Income: 0,
    Expense: 0,
    Bonus: 0,
    Revenue: 0,
  },
  loadingTotalNonAsset: false,
});
const pagination = reactive({
  recordCount: 0,
  pageCount: 0,
  currentPage: 1,
  pageSize: 15,
});
const filter = reactive({
  PoliceNum: [],
  TrayekName: [],
  CustomerName: [],
});

const emit = defineEmits({
  selectdata_detail: null,
  back: null,
  getGridRecords: null,
});
const summaaryTotal = computed({
  get() {
    return {
      Income: data.totalNonAsset.Income + data.totalAsset.Income,
      Expense: data.totalNonAsset.Expense + data.totalAsset.Expense,
      Bonus: data.totalNonAsset.Bonus + data.totalAsset.Bonus,
      Revenue: data.totalNonAsset.Revenue + data.totalAsset.Revenue,
    };
  },
});
function selectDataDetail(record) {
  if (data.loadingSelectData) return;
  getFromAsset(record._id, record, props.site);
}

function getFromAsset(id, record, site) {
  const url = "/bagong/asset/get-asset-detail";
  const param = { _id: id };
  data.loadingSelectData = true;
  axios.post(url, param).then(
    (r) => {
      data.loadingSelectData = false;
      r.data.TrayekName = record.TrayekName;
      r.data.Dimension = r.data.Dimension == null ? [] : r.data.Dimension;
      emit("selectdata_detail", r.data, record, site);
    },
    (e) => {
      data.loadingSelectData = false;
      util.showError(e);
    }
  );
}

function refreshGrid() {
  if (gridCtl.value) gridCtl.value.refreshData();
}

async function onOpenNonAsset(idx) {
  data.openNonAsset = true;
}
function closeNonAsset() {
  data.openNonAsset = false;
  refreshDataGrid();
  fetchDataTotalNonAsset();
}
function disableActionNonAsset(disable) {
  data.disableNonAsset = disable;
}

function changeFilter() {
  util.nextTickN(2, () => {
    pagination.currentPage = 1;
    refreshDataGrid();
  });
}
function changePageSize(pageSize) {
  pagination.pageSize = pageSize;
  pagination.currentPage = 1;
  refreshDataGrid();
}
function changePage(page) {
  pagination.currentPage = page;
  refreshDataGrid();
}
function getFilteredRecord() {
  return helper.cloneObject(data.originRecords)?.filter((x) => {
    if (data.query.SearchText !== '') {
      if (!x.PoliceNum.toLowerCase().includes(data.query.SearchText.toLowerCase()) && !x?.HullNo?.toLowerCase().includes(data.query.SearchText.toLowerCase())) {
        return false;
      }
    }
    if (data.query.TrayekName.length > 0) {
      if (data.query.TrayekName.includes(x.TrayekName) == false) return false;
    }
    if (data.query.PoliceNum.length > 0) {
      if (data.query.PoliceNum.includes(x.PoliceNum) == false) return false;
    }
    if (data.query.CustomerName.length > 0) {
      if (data.query.CustomerName.includes(x.CustomerName) == false)
        return false;
    }

    return true;
  });
}
function calcTotalAsset(records) {
  data.totalAsset = records.reduce(
    (a, el) => {
      a.Income += el.Income || 0;
      a.Expense += el.Expense || 0;
      a.Bonus += el.Bonus || 0;
      a.Revenue += el.Revenue || 0;
      return a;
    },
    { Income: 0, Expense: 0, Bonus: 0, Revenue: 0 }
  );
}
function refreshDataGrid() {
  const dt = getFilteredRecord();

  calcTotalAsset(dt);

  pagination.recordCount = dt.length;
  pagination.pageCount = Math.ceil(
    pagination.recordCount / pagination.pageSize
  );

  const start = (pagination.currentPage - 1) * pagination.pageSize;
  const end =
    pagination.currentPage == pagination.pageCount
      ? pagination.recordCount
      : pagination.pageSize;

  const records = dt.splice(start, end);

  util.nextTickN(2, () => {
    gridCtl.value.setRecords(records);
  });
}

defineExpose({
  refreshGrid,
});

function mappingDataFilter(field) {
  let records = helper.cloneObject(data.originRecords);

  const datas = records
    .filter((v) => v[field] != "")
    .map((item) => item[field]);
  const filterSet = new Set(datas);

  const uniqueFilterSet = Array.from(filterSet);

  filter[field] = uniqueFilterSet;
}

function syncData() {
  axios
    .post("/bagong/siteentry/get-site-entry-asset", {
      SiteEntryID: props.siteEntryId,
    })
    .then(() => {
      try {
        fetchDataAsset();
      } catch (err) {
        console.log(err);
      }
    });
}

function fetchDataAsset() {
  gridCtl.value.setLoading(true);
  axios
    .post("/bagong/siteentry_asset/gets?SiteEntryID=" + props.siteEntryId, {})
    .then(
      (r) => {
        const sortedBy =
          props.purpose == "Trayek" ? "TrayekName" : "CustomerName";

        data.originRecords = r.data.data.sort((a, b) => {
          a[sortedBy] ? a[sortedBy] : (a[sortedBy] = "");
          b[sortedBy] ? b[sortedBy] : (b[sortedBy] = "");
          let x = a[sortedBy].toLowerCase();
          let y = b[sortedBy].toLowerCase();
          return x < y ? -1 : x > y ? 1 : 0;
        });

        if (props.filters.includes("PoliceNum")) mappingDataFilter("PoliceNum");

        if (props.filters.includes("TrayekName"))
          mappingDataFilter("TrayekName");

        if (props.filters.includes("CustomerName"))
          mappingDataFilter("CustomerName");

        refreshDataGrid();

        gridCtl.value.setLoading(false);
      },
      (e) => {
        gridCtl.value.setRecords([]);
        gridCtl.value.setLoading(false);
      }
    );
}
function fetchDataTotalNonAsset() {
  data.loadingTotalNonAsset = true;
  axios.post("/bagong/siteentry_nonasset/get", [props.siteEntryId]).then(
    (r) => {
      data.totalNonAsset = { ...data.totalNonAsset, ...r.data };
      data.loadingTotalNonAsset = false;
    },
    (e) => {
      data.loadingTotalNonAsset = false;
    }
  );
}
function initData() {
  fetchDataAsset();
  fetchDataTotalNonAsset();
}
onMounted(() => {
  data.loadingGridCfg = true;
  loadGridConfig(axios, "/bagong/siteentry_asset/gridconfig").then(
    (r) => {
      const arrGrid = [
        helper.gridColumnConfig({ field: "AssetID" }),
        helper.gridColumnConfig({ field: "PoliceNum", label: "Police No" }),
        helper.gridColumnConfig({ field: "ProjectID", label: "Project ID" }),
        helper.gridColumnConfig({ field: "HullNo", label: "Hull No" }),
        helper.gridColumnConfig({
          field: props.title != "Trayek" ? "CustomerName" : "TrayekName",
          label: props.title != "Trayek" ? "Customer Name" : "Trayek Name",
        }),
        helper.gridColumnConfig({ field: "AssetTypeName", label: "Type" }),
        helper.gridColumnConfig({ field: "Status" }),
      ];

      r.fields = [...arrGrid, ...r.fields];

      r.fields = r.fields.filter((x) => x.field !== "Availability");

      if (props.purpose == "Trayek") {
        r.fields.splice(
          7,
          0,
          helper.gridColumnConfig({
            field: "Bonus",
            label: "Bonus",
            kind: "number",
          })
        );
      }

      data.gridDetailCfg = r;
      data.loadingGridCfg = false;

      util.nextTickN(2, () => {
        initData();
      });
    },
    (e) => {
      data.loadingGridCfg = false;
      util.showError(e);
    }
  );
});
</script>
