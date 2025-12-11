<template>
  <div class="w-full mt-4">
    <data-list
      ref="gridLoan"
      title="Loan"
      grid-config="/hcm/loan/gridconfig"
      form-config="/hcm/loan/formconfig"
      :grid-read="`/hcm/loan/gets?EmployeeID=${props.EmployeeID}`"
      grid-mode="grid"
      hideTitle
      gridHideDetail
      gridHideSearch
      gridHideNew
      gridHideSort
      @alterGridConfig="alterGridConfig"
    >
      <template #grid_item_buttons_1="{ item }">
        <s-button
          class="bg-blue-400 m-1"
          label="Generate"
          @click="generateLoan(item, this)"
          :disabled="item.Lines.length > 0"
        ></s-button>
        <s-button
          class="bg-blue-400 m-1"
          label="Detail"
          @click="showDetail(item)"
          :disabled="item.Lines.length == 0"
        ></s-button>
      </template>
    </data-list>
    <data-list
      v-if="data.isVisibleGrid"
      ref="gridLines"
      :title="data.gridTitle"
      grid-mode="grid"
      gridHideDetail
      gridHideDelete
      grid-hide-sort
      grid-hide-search
      grid-hide-new
      grid-hide-refresh
      grid-config="/hcm/loan/line/gridconfig"
      @gridRefreshed="onGridRefreshed"
      :grid-fields="['Status', 'Date']"
    >
      <template #grid_header_buttons_2>
        <s-button
          label="save"
          :disabled="data.loadingSave"
          icon="content-save"
          class="ml-2 btn_primary submit_btn"
          @click="handleSave"
        />
      </template>
      <template #grid_Date="{ item }">
        <s-input kind="date" v-model="item.Date" />
      </template>
      <template #grid_Status="{ item }">
        <span
          class="rounded-md p-2 m-2 inline-block"
          :class="
            item.Status == 'Unpaid'
              ? 'text-red-700, bg-red-400'
              : 'text-green-700, bg-green-400'
          "
          >{{ item.Status }}</span
        >
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, inject } from "vue";
import { DataList, SButton, SInput, loadFormConfig, util } from "suimjs";

const axios = inject("axios");
const props = defineProps({
  EmployeeID: { type: String, default: "" },
});

const gridLoan = ref(null);
const gridLines = ref(null);

const data = reactive({
  gridTitle: "",
  isVisibleGrid: false,
  record: {},
  gridRecords: [],
  loadingSave: false,
});

function alterGridConfig(cfg) {
  cfg.fields = cfg.fields.filter(
    (el) => ["EmployeeID"].indexOf(el.field) == -1
  );
}

function showDetail(d) {
    data.record = d
  data.isVisibleGrid = true;
  data.gridTitle = `ID: ${d._id}, Loan Date: ${d.RequestDate}`;
  util.nextTickN(2, () => {
    console.log("load data:", data.gridRecords);
    console.log("config:", gridLines.value.getGridConfig());
    data.gridRecords = d.Lines;
    gridLines.value.refreshGrid();
  });
}
function handleSave() {
  data.loadingSave = true;
  gridLines.value.setGridLoading(true);
  axios
    .post("/hcm/loan/save", data.record)
    .then((r) => {})
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loadingSave = false;
      gridLines.value.setGridLoading(false);
    });
}

function generateLoan(d, e) {
  let payload = d;
  axios
    .post("/hcm/loan/generate-loan", payload)
    .then(
      async (r) => {
        util.showInfo("Data has been generate");
        console.log(r);
      },
      (e) => {
        util.showError(e);
      }
    )
    .finally(() => {
      console.log("done");
      gridLoan.value.refreshList();
    });
}

function onGridRefreshed() {
  console.log("load:", data.gridRecords);
  gridLines.value.setGridRecords(data.gridRecords);
}
</script>
<style></style>
