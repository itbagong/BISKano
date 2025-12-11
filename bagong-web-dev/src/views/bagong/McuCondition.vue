<template>
  <div class="w-full card p-4">
    <data-list
      class="table-mcu"
      ref="gridMcu"
      title="MCU Condition"
      grid-config="/bagong/mcucondition/gridconfig"
      form-config="/bagong/mcucondition/formconfig"
      grid-read="/bagong/mcucondition/gets"
      init-app-mode="grid"
      grid-mode="grid"
      :gridEditor="data.isEditor"
      @alterGridConfig="alterGridConfig"
      @grid-refreshed="gridRefreshed"
      gridHideDelete
      gridHideDetail
      gridHideSelect
      grid-hide-refresh
      grid-hide-new
      grid-hide-search
      grid-hide-sort
    >
      <template #grid_header_buttons_1="{ config }">
        <s-button
          class="btn_primary"
          label="Save"
          @click="save"
          v-if="profile.canCreate"
        ></s-button>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, computed, inject } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SGrid, SInput, SButton, SModal, util } from "suimjs";
import { useRoute } from "vue-router";
import { authStore } from "@/stores/auth";

layoutStore().name = "tenant";
const FEATUREID = "MCUCondition";
const profile = authStore().getRBAC(FEATUREID);

const route = useRoute();
const axios = inject("axios");

const gridMcu = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  records: [],
  isEditor: true,
});

const items = [
  {
    _id: "01",
    Name: "Diastolic (mm Hg)",
    Gender: "Both",
  },
  {
    _id: "02",
    Name: "Systolic (mm Hg)",
    Gender: "Both",
  },
  {
    _id: "03",
    Name: "Cholesterol (mg/dL)",
    Gender: "Both",
  },
  {
    _id: "04",
    Name: "Uric Acid (mg/dl)",
    Gender: "Male",
  },
  {
    _id: "05",
    Name: "Uric Acid (mg/dl)",
    Gender: "Female",
  },
  {
    _id: "06",
    Name: "Random Blood Sugar (mg/dl)",
    Gender: "Both",
  },
  {
    _id: "07",
    Name: "BMI (Kg/m2)",
    Gender: "Male",
  },
  {
    _id: "08",
    Name: "BMI (Kg/m2)",
    Gender: "Female",
  },
  {
    _id: "09",
    Name: "Waist Circumference",
    Gender: "Male",
  },
  {
    _id: "10",
    Name: "Waist Circumference",
    Gender: "Female",
  },
  {
    _id: "11",
    Name: "Fasting Blood Sugar (mm Hg)",
    Gender: "Both",
  },
];

onMounted(() => {
  setTimeout(() => {}, 300);
});

function gridRefreshed() {
  let data = gridMcu.value.getGridRecords();
  var ids = new Set(data.map((d) => d._id));
  var grid = [...data, ...items.filter((d) => !ids.has(d._id))];
  grid.sort((a, b) => a._id - b._id);
  gridMcu.value.setGridRecords(grid);
}

function alterGridConfig(cfg) {
  cfg.fields = cfg.fields.filter((el) => ["_id"].indexOf(el.field) == -1);
  cfg.fields.forEach((f) => {
    if (f.field == "Name" || f.field == "_id") f.input.readOnly = true;
  });
}

function save() {
  let data = gridMcu.value.getGridRecords();
  for (let i in data) {
    let payload = data[i];
    axios.post("/bagong/mcucondition/save", payload).then(
      (r) => {
        if (data.length == parseInt(i) + 1)
          util.showInfo("Data has been saved");
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}
</script>
<style>
.table-mcu th:last-child,
.table-mcu td:last-child {
  display: none;
}
</style>
