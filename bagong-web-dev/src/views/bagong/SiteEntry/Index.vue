<template>
  <s-card
    :title="'Site Entry'"
    class="w-full bg-white grid_card"
    hide-footer
  
  >
    <template v-if="data.loadingGridCfg || data.loadingSelectData" >
      <slot name="loader">
        <div class="loader"></div>
      </slot> 
    </template> 
    <s-grid
      class="w-full se-grid"
      hideNewButton
      :read-url="'/bagong/siteentry/gets'"
      delete-url="/bagong/siteentry/delete"
      total-url="/bagong/siteentry/get-total-site-entry"
      v-if="data.gridCfg.setting"
      :config="data.gridCfg"
      form-keep-label
      @select-data="selectData"
      ref="gridCtl"
      :custom-filter="customFilter" 
      :hide-edit="!profile.canUpdate"
      :hide-delete-button="!profile.canDelete"
    >
      <template #header_search>
        <div class="flex gap-2 w-full">
          <s-input
            label="Purpose"
            kind="text"
            class="w-[130px]"
            v-model="data.searchQuery.purpose"
            :allow-add="false"
            use-list
            :items="['Mining', 'BTS', 'Trayek', 'Tourism']"
          />
          <s-input
            label="Site"
            kind="text"
            class="w-[130px]"
            v-model="data.searchQuery.site"
            :allow-add="false"
            use-list
            lookup-url="/bagong/sitesetup/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['Name']"
          />
          <s-input
            label="From"
            kind="date"
            class="w-[130px]"
            v-model="data.searchQuery.from"
          ></s-input>
          <s-input
            label="To"
            kind="date"
            class="w-[130px]"
            v-model="data.searchQuery.to"
          ></s-input>
          <div>
            <s-button
              class="btn_primary new_btn mt-3"
              label="Clear"
              @click="
                data.searchQuery.purpose = null;
                data.searchQuery.site = null;
                data.searchQuery.from = null;
                data.searchQuery.to = null;
              "
            />
          </div>
        </div>
      </template>
      <template #header_buttons_2>
        <s-button
          v-if="profile.canCreate"
          icon="plus"
          class="btn_primary new_btn"
          @click="
            data.modalAddNew = true;
            data.objAddNew = {};
            data.objAddNew.TrxDate = new Date();
          "
        />
      </template>
      <template #item_SiteID="{ item }">
        {{ item.SiteName }}
      </template>
      <template #item_Created="{ item }">
        <div>{{ moment(item.Created).format("DD-MMM-YYYY") }}</div>
      </template>
      <template #grid_total="{ item }">
        <tr v-for="(dt, idx) in item" :key="idx" class="font-semibold">
          <td colspan="6" class="ml-4">Total</td>
          <td class="text-right">{{ util.formatMoney(dt.Income, {}) }}</td>
          <td class="text-right">{{ util.formatMoney(dt.Expense, {}) }}</td>
          <td class="text-right">{{ util.formatMoney(dt.Revenue, {}) }}</td>
          <td></td>
        </tr>
      </template>
    </s-grid>
    <s-modal
      :display="data.modalAddNew"
      hideButtons
      title="Add new"
      @beforeHide="data.modalAddNew = false"
    >
      <s-card class="rounded-md w-full relative" hide-title>
        <div class="px-2 w-[400px]">
          <s-form
            ref="frmCtl"
            v-model="data.formRecord"
            :config="data.formCfg"
            keep-label
            only-icon-top
            hide-submit
            hide-cancel
          >
            <template #footer_1>
              <s-button 
                v-if="!data.loadingAddNew"
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
import { reactive, onMounted, inject, ref, computed, watch } from "vue";
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
import moment from "moment";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { useRouter } from "vue-router";

layoutStore().name = "tenant";


const FEATUREID = 'SiteEntry'
const profile = authStore().getRBAC(FEATUREID) 
const siteIDs = profile.Dimension?.filter((e) => e.Key === "Site").map(
  (e) => e.Value
);



const axios = inject("axios");
const auth = authStore();
const router = useRouter();

const frmCtl = ref(null);
const gridCtl = ref(null);

const data = reactive({
  loadingGridCfg: false,
  loadingSelectData: false,
  modalAddNew: false,
  formRecord: {
    SiteID: siteIDs.length == 1 ? siteIDs[0] : null,
  },
  formCfg: {},
  gridCfg: {},
  total: 1,
  searchQuery: {
    purpose: null,
    site: siteIDs.length == 1 ? siteIDs[0] : null,
    from: null,
    to: null,
  },
  loadingAddNew: false
});

function addNew() {
  
  const isValid = frmCtl.value.validate();
  if (!isValid) return;

  const param = data.formRecord;
  param.CompanyID = auth.companyId;
  param.TrxDate = moment(param.TrxDate).format("YYYY-MM-DDT00:00:00Z");
  data.loadingSelectData = true
  data.loadingAddNew = true
  frmCtl.value.setLoading(true)
  axios.post("/bagong/siteentry/get-site-entry", param).then(
    (r) => {
      data.loadingAddNew = false
      frmCtl.value.setLoading(false)
      data.modalAddNew = false;  
      data.formRecord = {};
      axios
        .post("/bagong/siteentry/gets?SiteID=" + param.SiteID, {
          sort: ["-Created"],
        })
        .then(
          (r) => {
            data.loadingSelectData = false
            selectData(r.data.data[0]);
          },
          (e) => { 
            util.showError(e);
            data.loadingSelectData = false
          }
        );
      gridCtl.value.refreshData();
    },
    (e) => {
      frmCtl.value.setLoading(false)
      data.loadingAddNew = false
      util.showError(e);
       data.loadingSelectData = false
    }
  );
}

function selectData(record) {
  switch (record.Purpose) {
    case "Mining":
      router.push(
        "/bagong/siteEntry/Mining?id=" + record._id + "&siteId=" + record.SiteID
      );
      break;
    case "Trayek":
      router.push(
        "/bagong/siteEntry/Trayek?id=" + record._id + "&siteId=" + record.SiteID
      );
      break;
    case "BTS":
      router.push(
        "/bagong/siteEntry/Bts?id=" + record._id + "&siteId=" + record.SiteID
      );
      break;
    case "Tourism":
      router.push(
        "/bagong/siteEntry/Tourism?id=" +
          record._id +
          "&siteId=" +
          record.SiteID
      );
      break;
  }
}

const customFilter = computed(() => {
  const filters = [];
  if (data.searchQuery.purpose != null) {
    filters.push({
      Op: "$in",
      Field: "Purpose",
      Value: [data.searchQuery.purpose],
    });
  }

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
      Field: "TrxDate",
      Value: helper.formatFilterDate(data.searchQuery.from) //moment(data.searchQuery.from).format("YYYY-MM-DDT00:00:00Z"),
    });
  }

  if (data.searchQuery.to != null) {
    filters.push({
      Op: "$lte",
      Field: "TrxDate",
      Value: helper.formatFilterDate(data.searchQuery.to, true)//moment(data.searchQuery.to).format("YYYY-MM-DDT00:00:00Z"),
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

onMounted(() => {
  data.loadingGridCfg =true
  loadGridConfig(axios, "/bagong/siteentry/gridconfig").then(
    (r) => {
      r.setting.sortable = ["TrxDate"];
      data.gridCfg = r;
      data.loadingGridCfg = false
    },
    (e) => {
      data.loadingGridCfg = false
      util.showError(e)
    }
  );
  loadFormConfig(axios, "/bagong/siteentry/formconfig").then(
    (r) => {
      data.formCfg = r;
    },
    (e) => util.showError(e)
  );
});

watch(
  () => [
    data.searchQuery.purpose,
    data.searchQuery.site,
    data.searchQuery.from,
    data.searchQuery.to,
  ],
  (nv) => {
    if (nv[1] == "Invalid date") {
      data.searchQuery.from = null;
    }
    if (nv[2] == "Invalid date") {
      data.searchQuery.to = null;
    }
    util.nextTickN(2, () => {
      gridCtl.value.refreshData();
    });
  },
  { deep: true }
);
</script>
