<template>
  <data-list
    class="card rsca_transaction"
    ref="listControl"
    title="RSCA Transaction"
    grid-config="/she/rscatrx/gridconfig"
    form-config="/she/rscatrx/formconfig"
    grid-read="/she/rscatrx/gets"
    form-read="/she/rscatrx/get"
    grid-mode="grid"
    grid-delete="/she/rscatrx/delete"
    form-keep-label
    form-insert="/she/rscatrx/save"
    form-update="/she/rscatrx/save"
    :init-app-mode="data.appMode"
    grid-hide-select
    @form-edit-data="openForm"
    @form-new-data="newData"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    stay-on-form-after-save
    @form-field-change="onFieldChange"
    :form-fields="['Dimension']"
    :form-tabs-edit="['General', 'Lines']"
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
          label="No"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refLocation"
          v-model="data.search.Location"
          lookup-key="_id"
          label="Location "
          class="w-[200px]"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=LOC`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refLocationDetail"
          v-model="data.search.LocationDetail"
          label="Location Detail"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refName"
          v-model="data.search.Name"
          label="Name"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refTemplate"
          v-model="data.search.Template"
          lookup-key="_id"
          label="Template "
          class="w-[200px]"
          use-list
          :lookup-url="`/she/masteribpr/find`"
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
    <template #form_input_Dimension="{ item, mode }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :read-only="mode == 'view'"
      ></dimension-editor-vertical>
    </template>
    <template #form_tab_Lines="{ item }">
      <lines
        :template-id="item.TemplateID"
        v-model="item.Lines"
        :risk-matrix="data.riskMatrix"
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
  SModal,
  loadGridConfig,
  loadFormConfig,
} from "suimjs";
import { layoutStore } from "@/stores/layout.js";
layoutStore().name = "tenant";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import Lines from "./widget/RscaTrxLines.vue";

const listControl = ref(null);
const axios = inject("axios");
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
      Op: "$eq",
      Value: data.search.No,
    });
  }
  if (data.search.Template !== null && data.search.Template !== "") {
    filters.push({
      Field: "TemplateID",
      Op: "$eq",
      Value: data.search.Template,
    });
  }
  if (data.search.Name !== null && data.search.Name !== "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "Name",
          Op: "$contains",
          Value: [data.search.Name],
        },
      ],
    });
  }

  if (data.search.Location !== null && data.search.Location !== "") {
    filters.push({
      Field: "Location",
      Op: "$eq",
      Value: data.search.Location,
    });
  }

  if (
    data.search.LocationDetail !== null &&
    data.search.LocationDetail !== ""
  ) {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "LocationDetail",
          Op: "$contains",
          Value: [data.search.LocationDetail],
        },
      ],
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
  currentActionCfg: {},
  riskMatrix: {},
  search: {
    DateFrom: null,
    DateTo: null,
    No: "",
    Location: "",
    LocationDetail: "",
    Template: "",
    Name: "",
    Status: "",
  },
});

function newData(r) {
  data.record = r;
}

function openForm(r) {
  data.record = r;
}

function getMatrix() {
  const url = "/bagong/riskmatrix/find?Type=RSCA";
  axios.post(url).then(
    (r) => {
      for (let i in r.data) {
        let o = r.data[i];
        data.riskMatrix[o._id] = o;
        data.riskMatrix[`${o.SeverityID}#${o.LikelihoodID}`] = o;
      }
    },
    (e) => {}
  );
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
onMounted(() => {
  getMatrix();
});
</script>
