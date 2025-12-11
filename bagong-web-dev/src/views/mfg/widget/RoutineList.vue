<template>
  <s-card :title="'Routine'" class="w-full bg-white suim_datalist" hide-footer>
    <s-grid
      class="w-full r-grid grid-line-items"
      ref="gridRoutine"
      hideNewButton
      hide-select
      :read-url="'/mfg/routine/gets'"
      delete-url="/mfg/routine/delete"
      v-if="data.gridCfg.setting"
      :config="data.gridCfg"
      form-keep-label
      @select-data="selectData"
      :custom-filter="customFilter"
    >
      <template #header_buttons_1>
        <div class="flex gap-2 mr-2 w-[600px] -mt-3">
          <!-- lookup-url="/bagong/sitesetup/find" -->
          <s-input
            label="Site"
            kind="input"
            class="w-[200px]"
            v-model="data.searchQuery.site"
            :allow-add="false"
            useList
            lookup-key="_id"
            :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
            :lookup-labels="['Label']"
            :lookup-searchs="['_id', 'Label']"
            @focus="onFocus"
            :caption="'Site'"
          />
          <s-input
            label="From"
            kind="date"
            class="grow"
            v-model="data.searchQuery.from"
          ></s-input>
          <s-input
            label="To"
            kind="date"
            class="grow"
            v-model="data.searchQuery.to"
          ></s-input>
        </div>
      </template>
      <template #header_buttons_2>
        <s-button
          icon="plus"
          class="btn_primary new_btn"
          @click="
            data.modalAddNew = true;
            data.objAddNew = {};
            data.objAddNew.SiteID = data.searchQuery.site;
            data.objAddNew.ExecutionDate = new Date();
          "
        />
      </template>
      <template #item_SiteID="{ item }">
        <s-input v-model="item.SiteName" read-only></s-input>
      </template>
    </s-grid>
    <s-modal
      :display="data.modalAddNew"
      hideButtons
      title="Add new"
      @beforeHide="data.modalAddNew = false"
    >
      <s-card class="rounded-md w-full" hide-title>
        <div class="px-2 w-[400px]">
          <s-form
            ref="inputAddNew"
            v-model="data.objAddNew"
            :config="data.formCfgAddNew"
            keep-label
            only-icon-top
            hide-submit
            hide-cancel
          >
            <template #input_SiteID="{ item }">
              <s-input
                ref="refItemSite"
                label="Site"
                kind="input"
                class="w-full"
                v-model="item.SiteID"
                :allow-add="false"
                useList
                lookup-key="_id"
                :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
                :lookup-labels="['Label']"
                :lookup-searchs="['_id', 'Label']"
                @focus="onFocus"
                :required="true"
                :keepErrorSection="true"
                :caption="'Site'"
              />
            </template>
            <template #footer_1>
              <div class="w-full h-[30px]" v-if="data.loadingAdd">
                <loader kind="skeleton" skeleton-kind="input" />
              </div>
              <s-button
                v-else
                icon="plus"
                label="Add New"
                class="w-full btn_primary flex justify-center"
                @click="addNew"
              />
            </template>
          </s-form>
        </div>
      </s-card>
    </s-modal>
  </s-card>
</template>

<script setup>
import {
  reactive,
  onMounted,
  inject,
  ref,
  nextTick,
  watch,
  computed,
} from "vue";
import {
  SCard,
  SGrid,
  SForm,
  loadGridConfig,
  loadFormConfig,
  util,
  SButton,
  SModal,
  SInput,
} from "suimjs";
import helper from "@/scripts/helper.js";
import { authStore } from "@/stores/auth";
import moment from "moment";
import Loader from "@/components/common/Loader.vue";
const featureID = "WorkRequest";
const profile = authStore().getRBAC(featureID);
const axios = inject("axios");
const inputAddNew = ref(null);
const gridRoutine = ref(null);
const refItemSite = ref(null);

const props = defineProps({
  title: { type: String, default: "" },
});

const data = reactive({
  modalAddNew: false,
  objAddNew: {},
  formCfg: {},
  gridCfg: {},
  searchQuery: {
    site: null,
    from: null,
    to: null,
  },
  loadingAdd: false,
});
const emit = defineEmits({
  selectdata: null,
  onPostAddNew: null,
});
function addNew() {
  const url = "/mfg/routine/add-new";
  const isValid = inputAddNew.value.validate();

  if (!isValid || !refItemSite.value.validate()) return;

  const param = data.objAddNew;
  param.ExecutionDate = moment(param.ExecutionDate).format(
    "YYYY-MM-DDT00:00:00Z"
  );
  data.loadingAdd = true;
  axios
    .post(url, param)
    .then(
      (r) => {
        nextTick(() => {
          data.modalAddNew = false;
          gridRoutine.value.refreshData();
          emit("onPostAddNew", r.data);
        });
      },
      (e) => {
        util.showError(e);
      }
    )
    .finally(() => {
      data.loadingAdd = false;
    });
}
function selectData(record) {
  emit("selectdata", record);
}
function genCfg() {
  loadFormConfig(axios, "/mfg/routine/formconfig").then(
    (r) => {
      data.formCfgAddNew = r;
    },
    (e) => util.showError(e)
  );
}
const customFilter = computed(() => {
  const filters = [];

  if (data.searchQuery.site != null) {
    filters.push({
      Op: "$in",
      Field: "SiteID",
      Value: [data.searchQuery.site],
    });
  }
  if (data.searchQuery.from != null) {
    filters.push({
      Op: "$gte",
      Field: "ExecutionDate",
      Value: moment(data.searchQuery.from).format("YYYY-MM-DDT00:00:00Z"),
    });
  }

  if (data.searchQuery.to != null) {
    filters.push({
      Op: "$lte",
      Field: "ExecutionDate",
      Value: moment(data.searchQuery.to).format("YYYY-MM-DDT00:00:00Z"),
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

function getDetailEmployee(_id) {
  let payload = [];
  if (_id) {
    payload = [_id];
  }
  axios
    .post("/tenant/employee/get-emp-warehouse", payload)
    .then(
      (r) => {
        if (r.data.Dimension != null) {
          data.searchQuery.site = r.data.Dimension.find(
            (_dim) => _dim.Key === "Site"
          )["Value"];
        }
      },
      (e) => util.showError(e)
    )
    .finally(function () {
      loadGridConfig(axios, "/mfg/routine/gridconfig").then(
        (r) => {
          data.gridCfg = r;
        },
        (e) => util.showError(e)
      );
      genCfg();
    });
}
function lookupPayloadBuilder(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

  //setting search
  const Site =
    profile.Dimension &&
    profile.Dimension.find((_dim) => _dim.Key === "Site") &&
    profile.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? profile.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : undefined;
  const querySite = [
    {
      Field: "Dimension.Key",
      Op: "$eq",
      Value: "Site",
    },
    {
      Field: "Dimension.Value",
      Op: "$eq",
      Value: Site,
    },
  ];
  if (Site) {
    qp.Where = {
      Op: "$and",
      items: querySite,
    };
  }
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "_id", Op: "$contains", Value: [search] },
          { Field: "Name", Op: "$contains", Value: [search] },
        ],
      },
    ];
    if (Site) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}
watch(
  () => data.filter,
  (nv) => {
    util.nextTickN(2, () => {
      gridSiteEntry.value.refreshData();
      if (nv.site) genCfg(nv.site);
    });
  },
  { deep: true }
);
onMounted(() => {
  const list = profile.Dimension.filter((v) => v.Key == "Site").map(
    (e) => e.Value
  );
  const v = list.length === 1 ? list[0] : "";
  data.searchQuery.site = v;
  loadGridConfig(axios, "/mfg/routine/gridconfig").then(
    (r) => {
      data.gridCfg = r;
    },
    (e) => util.showError(e)
  );
  genCfg();
});
</script>
