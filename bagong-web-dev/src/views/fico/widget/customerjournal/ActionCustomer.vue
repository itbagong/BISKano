<template>
  <s-card class="rounded-md w-full" hide-title>
    <div class="px-2 max-w-[400px] relative" v-if="data.mode == 'form'">
      <div name="pull invoice lines">
        <div class="flex gap-2 mb-2">
          <s-input
            label="Start"
            kind="date"
            class="w-full mb-2"
            v-model="data.action.Start"
          />
          <s-input
            label="End"
            kind="date"
            class="w-full mb-2"
            v-model="data.action.End"
          />
        </div>
        <s-input
          label="Project"
          class="w-full"
          use-list
          :items="data.projectList"
          v-model="data.action.ProjectID"
        />
        <s-button
          label="Submit"
          class="w-full btn_primary flex justify-center mt-6"
          @click="
            data.mode = 'grid';
            emit('submit', data.action, data.projectList);
          "
        />
      </div>
    </div>

    <data-list
      no-gap
      hide-title
      class="card grid-action-customer"
      ref="listControl"
      :grid-config="objType[kind].config"
      :grid-read="objType[kind].read"
      grid-mode="grid"
      grid-sort-direction="desc"
      @alter-grid-config="onAlterConfig"
      :init-app-mode="'grid'"
      grid-hide-new
      grid-hide-delete
      grid-hide-edit
      :grid-custom-filter="customFilter"
      grid-hide-search
      grid-hide-sort
      v-if="data.mode == 'grid'"
      @grid-check-uncheck="onCheckUncheck"
    >
      <template #grid_header_buttons_2="{ config }">
        <s-button
          class="bg-primary text-white font-bold w-full flex justify-center"
          label="Back"
          @click="data.mode = 'form'"
        ></s-button>
        <s-button
          class="bg-primary text-white font-bold w-full flex justify-center"
          label="Generate"
          @click="onGenerate"
        ></s-button>
      </template>
    </data-list>
  </s-card>
</template>
<script setup>
import { DataList, util, SButton, SInput } from "suimjs";
import { reactive, ref, computed, inject, watch, onMounted } from "vue";
import helper from "@/scripts/helper.js";
import moment from "moment";
const axios = inject("axios");
const listControl = ref(null);

const props = defineProps({
  dataItem: { type: Object, default: () => {} },
});

const emit = defineEmits({
  lines: null,
  submit: null,
});

const data = reactive({
  modal: {
    isShow: false,
  },
  action: {
    ProjectID: "",
    Start: new Date(),
    End: new Date(),
  },
  mode: "form",
  projectList: [],
  selectedRecord: {},
});

const objType = {
  mining: {
    read: "/bagong/invoice/get-site-entry",
    config: "/bagong/siteentry_asset/gridconfig",
    urlGenerate: "/bagong/invoice/generate-detail-mining",
  },
  general: {
    read: "/bagong/invoice/get-sales-order",
    config: "/bagong/siteentry_asset/gridconfig",
    urlGenerate: "/bagong/invoice/generate-general-invoice",
  },
};

function onGenerate() {
  let dt = selecteds.value;
  const url = objType[kind.value].urlGenerate;
  const o = helper.cloneObject(data.action);
  let SiteID = helper.findDimension(props.dataItem.Dimension, "Site");

  let param = {};

  if (kind.value == "general") {
    param.JournalID = props.dataItem._id;
    param.SalesOrderID = dt.map((o) => o._id);
  }
  if (kind.value == "mining") {
    param.SiteID = SiteID;
    param.CustomerID = props.dataItem.CustomerID;
    param.ProjectID = data.action.ProjectID;
    param.Start = moment(o.Start).format("YYYY-MM-DDT00:00:00Z");
    param.End = moment(o.End).format("YYYY-MM-DDT00:00:00Z");
    param.AssetID = dt.map((o) => o.AssetID);
    param.Journal = props.dataItem;
  }

  axios.post(url, param).then(
    (r) => {
      emit("lines", r.data.Lines, r.data.References);
    },
    (e) => util.showError(e)
  );
}

const customFilter = computed(() => {
  let SiteID = helper.findDimension(props.dataItem.Dimension, "Site");
  let res = {
    SiteID: SiteID,
    CustomerID: props.dataItem.CustomerID,
  };

  if (kind.value == "mining") {
    const param = helper.cloneObject(data.action);
    res = { ...res, ...param };
    res.Start = moment(res.Start).format("YYYY-MM-DDT00:00:00Z");
    res.End = moment(res.End).format("YYYY-MM-DDT00:00:00Z");
  }

  return res;
});

const kind = computed({
  get() {
    let res = "";
    if (props.dataItem.TransactionType == "Mining Invoice - Rent")
      res = "mining";
    if (props.dataItem.TransactionType.includes("General Invoice")) {
      res = "general";
    }
    return res;
  },
});

function onAlterConfig(config) {
  let arrFieldsMining = [
    helper.gridColumnConfig({ field: "AssetID", label: "Asset ID" }),
    helper.gridColumnConfig({ field: "PoliceNo", label: "Police No" }),
  ];

  let arrFieldsGeneral = [
    helper.gridColumnConfig({ field: "SalesOrderNo", label: "No SO" }),
    helper.gridColumnConfig({ field: "Name", label: "SO Name" }),
    helper.gridColumnConfig({ field: "CustomerID", label: "Customer" }),
    helper.gridColumnConfig({ field: "Status" }),
    helper.gridColumnConfig({
      field: "SalesOrderDate",
      label: "SO Date",
      kind: "date",
    }),
    helper.gridColumnConfig({
      field: "TotalAmount",
      label: "Total Amount",
      kind: "number",
    }),
  ];

  config.fields = kind.value == "general" ? arrFieldsGeneral : arrFieldsMining;
}

function resetModal() {
  data.action.ProjectID = "";
  data.action.Start = new Date();
  data.action.End = new Date();
}

function getProject() {
  const url = "/bagong/invoice/get-project";
  let da = helper.cloneObject(data.action);
  let siteId = helper.findDimension(props.dataItem.Dimension, "Site");
  let param = {
    Where: {
      SiteID: siteId,
      CustomerID: props.dataItem.CustomerID,
      Start: moment(da.Start).utc().format("YYYY-MM-DDT00:00:00Z"),
      End: moment(da.End).utc().format("YYYY-MM-DDT00:00:00Z"),
    },
  };
  axios.post(url, param).then(
    (r) => {
      data.projectList = r.data.map((d) => {
        return {
          key: d["_id"],
          text: d["Name"],
          item: { ...d },
        };
      });
    },
    (e) => {}
  );
}

const selecteds = computed(() => {
  return Object.keys(data.selectedRecord).map(
    (key) => data.selectedRecord[key]
  );
});

function onCheckUncheck(o) {
  if (o.isSelected) {
    data.selectedRecord[o._id] = o;
  } else {
    delete data.selectedRecord[o._id];
  }
}

watch(
  () => data.action.Start,
  (nv) => {
    getProject();
  },
  { deep: true }
);

watch(
  () => data.action.End,
  (nv) => {
    getProject();
  },
  { deep: true }
);

onMounted(() => {
  resetModal();
  if (kind.value == "general") data.mode = "grid";
});
</script>
<style>
.grid-action-customer .suim_grid .header {
  @apply mb-0;
}
</style>
