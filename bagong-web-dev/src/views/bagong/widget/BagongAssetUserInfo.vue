<template>
  <div>
    <div class="title section_title">{{ title }}</div>
    <data-list
      :key="data.keyGrid"
      ref="listControl"
      hide-title
      no-gap
      grid-editor
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-hide-detail
      grid-hide-select
      grid-hide-new
      grid-no-confirm-delete
      :init-app-mode="data.appMode"
      :grid-mode="data.appMode"
      new-record-type="grid"
      :grid-fields="[
        'SONumber',
        'SOStartDate',
        'SOEndDate',
        'SiteID',
        'CustomerID',
        'ProjectID',
      ]"
      grid-config="/bagong/user_info/gridconfig"
      form-config="/bagong/user_info/formconfig"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @alter-grid-config="onAlterGridConfig"
    >
      <template #grid_SiteID="{ item }">
        <s-input
          hide-label
          label="Site"
          v-model="item.SiteID"
          useList
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['_id', 'Name']"
          lookupUrl="/bagong/sitesetup/find"
          read-only
        />
      </template>
      <template #grid_ProjectID="{ item }">
        <s-input
          hide-label
          label="Project ID"
          v-model="item.ProjectID"
          lookup-key="_id"
          use-list
          :lookup-labels="['ProjectName']"
          :lookupSearchs="['ProjectName', '_id']"
          lookup-url="/sdp/measuringproject/find"
          read-only
        ></s-input>
        <!-- <div>{{ data.projects.find(o => {
          return o._id === item.ProjectID
        }).ProjectName}}</div> -->
      </template>
      <template #grid_CustomerID="{ item }">
        <s-input
          useList
          hide-label
          label="Customer"
          lookup-url="/tenant/customer/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          v-model="item.CustomerID"
          read-only
        />
      </template>
      <template #grid_SONumber="{ item }">
        <s-input
          kind="text"
          hide-label
          label="SO Number"
          v-model="item.SONumber"
          lookup-key="_id"
          use-list
          :lookup-labels="['SalesOrderNo', 'Name']"
          :lookupSearchs="['_id', 'Name', 'SalesOrderNo']"
          lookup-url="/sdp/salesorder/find"
          read-only
        ></s-input>
      </template>
      <template #grid_SOStartDate="{ item }">
        <s-input
          v-if="item.SOStartDate !== null"
          kind="date"
          hide-label
          label="SO Start Date"
          v-model="item.SOStartDate"
          read-only
        ></s-input>
        <span v-else></span>
      </template>
      <template #grid_SOEndDate="{ item }">
        <s-input
          v-if="item.SOEndDate !== null"
          kind="date"
          hide-label
          label="SO End Date"
          v-model="item.SOEndDate"
          read-only
        ></s-input>
        <span v-else></span>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch, inject } from "vue";
import { DataList, loadGridConfig, SInput, util } from "suimjs";

const props = defineProps({
  title: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);

const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  records: props.modelValue.map((dt) => {
    dt.suimRecordChange = false;
    // getProjectName(dt.ProjectID, dt);
    return dt;
  }),
  projects: [],
  keyGrid: util.uuid(),
});

function newRecord(r) {
  r.AssetDateFrom = new Date().toISOString();
  r.AssetDateTo = new Date().toISOString();
  r.SiteID = "";
  r.UserID = "";
  r.CustomerID = "";
  r.NoHullCustomer = "";
  r.Description = "";

  data.records.push(r);
  updateItems();
}

function onGridRowDelete(_, index) {
  const newRecords = data.records.filter((_, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  updateItems();
  util.nextTickN(2, () => {
    data.keyGrid = util.uuid()
  })
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
  emit("update:modelValue", data.records);
}
function onAlterGridConfig(config) {
  setTimeout(() => {
    updateItems();
  }, 500);
}

// ! This function was created because when using s-input lookup it got the bug when deleting row. its just happening on ProjectID field only
// function getProjectName(id, dt) {
//   axios
//     .post("/sdp/measuringproject/find", {
//       Take: 1,
//       Where: {
//         Field: "_id",
//         Op: "$eq",
//         Value: id,
//       },
//     })
//     .then((r) => {
//       if (r.data.length > 0) {
//         data.projects.push(r.data[0]);
//       }
//     });
// }
onMounted(() => {
  // loadGridConfig(axios, "/bagong/user_info/gridconfig").then(
  //   (r) => {
  //     setTimeout(() => {
  //       updateItems();
  //     }, 500);
  //   },
  //   (e) => util.showError(e)
  // );
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
