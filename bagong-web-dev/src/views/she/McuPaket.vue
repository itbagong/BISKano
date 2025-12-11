<template>
  <data-list
    class="card"
    ref="listControl"
    title="MCU Master Paket"
    grid-config="/she/mcumasterpackage/gridconfig"
    form-config="/she/mcumasterpackage/formconfig"
    grid-read="/she/mcumasterpackage/gets"
    form-read="/she/mcumasterpackage/get"
    grid-mode="grid"
    grid-delete="/she/mcumasterpackage/delete"
    form-keep-label
    form-insert="/she/mcumasterpackage/save"
    form-update="/she/mcumasterpackage/save"
    :init-app-mode="data.appMode"
    grid-hide-select
    @form-edit-data="openForm"
    @form-new-data="newData"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    stay-on-form-after-save
    :form-tabs-edit="['General', 'Line']"
    :grid-custom-filter="customFilter"
  >
    <template #grid_header_search="{ config }">
      <div class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter">
        <s-input
          kind="date"
          label="Date From"
          class="w-[200px]"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Date To"
          class="w-[200px]"
          v-model="data.search.DateTo"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refNo"
          v-model="data.search.No"
          label="ID"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refName"
          v-model="data.search.Name"
          label="Package name"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refProvider"
          v-model="data.search.Provider"
          lookup-key="_id"
          label="Hospital/Provider"
          class="w-[300px]"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=MPR`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refStatus"
          v-model="data.search.Status"
          lookup-key="_id"
          label="Status"
          class="w-[200px]"
          use-list
          :items="['DRAFT', 'SUBMITTED', 'READY', 'POSTED', 'REJECTED']"
          @change="refreshData"
        ></s-input>
      </div>
    </template>
    <template #form_tab_Line="{ item, mode }">
      <div class="w-full flex mb-4">
        <s-button
          class="bg-success text-white font-bold"
          label="+ Jenis Pemeriksaan"
          @click="item.Lines.push({ Name: '', Lines: [] })"
        ></s-button>
      </div>
      <lines
        v-for="(dt, idx) in item.Lines"
        :key="idx"
        class="mb-4"
        :data-items="item.Lines"
        v-model="item.Lines[idx]"
        :index="idx"
        :selected-id="selectedName"
      />
    </template>
  </data-list>
</template>
<script setup>
import {
  reactive,
  ref,
  inject,
  watch,
  computed,
  onMounted,
  nextTick,
} from "vue";
import {
  SInput,
  util,
  SButton,
  SGrid,
  SCard,
  DataList,
  loadGridConfig,
} from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import Lines from "./widget/McuPaketLines.vue";

layoutStore().name = "tenant";
const listControl = ref(null);
let customFilter = computed(() => {
  const filters = [];
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "Created",
      Op: "$gte",
      Value: moment(data.search.DateFrom).utc().format("YYYY-MM-DDT00:mm:00Z"),
    });
  }
  if (
    data.search.DateTo !== null &&
    data.search.DateTo !== "" &&
    data.search.DateTo !== "Invalid date"
  ) {
    filters.push({
      Field: "Created",
      Op: "$lte",
      Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
    });
  }

  if (data.search.No !== null && data.search.No !== "") {
    filters.push({
      Field: "_id",
      Op: "$contains",
      Value: [data.search.No],
    });
  }

  if (data.search.Name !== null && data.search.Name !== "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "PackageName",
          Op: "$contains",
          Value: [data.search.Name],
        },
      ],
    });
  }
  if (data.search.Provider !== null && data.search.Provider !== "") {
    filters.push({
      Field: "Provider",
      Op: "$eq",
      Value: data.search.Provider,
    });
  }
  if (data.search.Status !== null && data.search.Status !== "") {
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: data.search.Status,
    });
  }

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});
const data = reactive({
  appMode: "grid",
  record: {},
  search: {
    DateFrom: null,
    DateTo: null,
    No: "",
    Name: "",
    Provider: "",
    Status: "",
  },
});

function newData(r) {
  r.TrxDate = new Date();
}

function openForm(r) {
  data.record = r;
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
const selectedName = computed({
  get() {
    return data.record.Lines.map((o) => o.Name);
  },
});
</script>
