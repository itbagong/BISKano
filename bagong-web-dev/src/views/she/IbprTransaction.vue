<template>
  <data-list
    class="card ibpr_transaction"
    ref="listControl"
    title="IBPR Transaction"
    grid-config="/she/ibprtrx/gridconfig"
    form-config="/she/ibprtrx/formconfig"
    grid-read="/she/ibprtrx/gets"
    form-read="/she/ibprtrx/get"
    grid-mode="grid"
    grid-delete="/she/ibprtrx/delete"
    form-keep-label
    form-insert="/she/ibprtrx/save"
    form-update="/she/ibprtrx/save"
    :init-app-mode="data.appMode"
    grid-hide-select
    @form-edit-data="openForm"
    @form-new-data="newData"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    stay-on-form-after-save
    @form-field-change="onFieldChange"
    :form-tabs-edit="[
      'General',
      'Resiko Awal',
      'Resiko Residual',
      'Penilaian Peluang',
    ]"
    :form-fields="['Dimension', 'Name']"
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
    <template #form_tab_Resiko_Awal="{ item }">
      <analisis
        v-model="item.InitialRisks"
        :master-ibpr="item.Lines"
        :format-row-span="formatRowSpan"
        :currentActionCfg="data.currentActionCfg"
        :risk-matrix="data.riskMatrix"
        :cal-matrix="calMatrix"
        :set-bg-matrix="setBgMatrix"
        @row-change="onRowChange"
        tab-id="Initial"
        @modal="onModal"
        :jurnal-id="item._id"
      />
    </template>
    <template #form_tab_Resiko_Residual="{ item }">
      <analisis
        v-model="item.ResidualRisks"
        :master-ibpr="item.Lines"
        :format-row-span="formatRowSpan"
        :currentActionCfg="data.currentActionCfg"
        :risk-matrix="data.riskMatrix"
        :cal-matrix="calMatrix"
        :set-bg-matrix="setBgMatrix"
        @row-change="onRowChange"
        tab-id="Residual"
        @modal="onModal"
        :jurnal-id="item._id"
      />
    </template>
    <template #form_tab_Penilaian_Peluang="{ item }">
      <analisis
        v-model="item.OpportunityAssessments"
        :master-ibpr="item.Lines"
        :format-row-span="formatRowSpan"
        :currentActionCfg="data.currentActionCfg"
        :risk-matrix="data.riskMatrix"
        :cal-matrix="calMatrix"
        :set-bg-matrix="setBgMatrix"
        @row-change="onRowChange"
        tab-id="Opportunity"
        @modal="onModal"
        :jurnal-id="item._id"
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
import Analisis from "./widget/IbprTrx/Analisis.vue";

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
      Field: "LocationID",
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

function onFieldChange(name, v1) {
  if (name == "TemplateID") {
    getIbrpTemplate(v1);
  }
}

function getIbrpTemplate(id) {
  if (!id) {
    mapField();
    return;
  }

  const url = "/she/masteribpr/get";
  axios.post(url, [id]).then(
    (r) => {
      mapField(r.data);
    },
    (e) => {
      util.showError(e);
    }
  );
}

function mapField(
  dt = { Location: "", LocationDetail: "", IBPRTeam: [], Lines: [] }
) {
  const fields = ["Location", "LocationDetail", "IBPRTeam", "Lines"];
  for (let i in fields) {
    let o = fields[i];
    data.record[o] = dt[o];
  }
}

function formatRowSpan(selector) {
  const myTable = document.querySelectorAll(
    `.${selector} .suim_table [name='grid_body']`
  );
  for (let i = 0, row; (row = myTable[0].rows[i]); i++) {
    const firstCell = row.cells[0];
    const secondCell = row.cells[1];
    firstCell.classList.remove("hidden");
    secondCell.classList.remove("hidden");
    if (data.record.Lines[i].ParentId == "") {
      let lengthRow = data.record.Lines.filter(
        (o) => o.ParentId == data.record.Lines[i].ID
      );
      firstCell.rowSpan = lengthRow.length + 1;
      secondCell.rowSpan = lengthRow.length + 1;
    } else {
      firstCell.classList.add("hidden");
      secondCell.classList.add("hidden");
    }
  }
}

function calMatrix(dt, selector) {
  if (!dt.Probability || !dt.Severity) {
    setBgMatrix(dt.ID, null, selector);
    return;
  }
  if (data.riskMatrix[`${dt.Severity}#${dt.Probability}`]) {
    let o = data.riskMatrix[`${dt.Severity}#${dt.Probability}`];
    dt.RiskRating = o._id;
    setBgMatrix(dt.ID, o.RiskID, selector);
  }
}

function setBgMatrix(id, riskID, selector) {
  util.nextTickN(3, () => {
    let child = document.querySelector(`.${selector}`);
    let parent = child.closest("td");
    parent.classList.remove(
      "BgMatrix-A",
      "BgMatrix-B",
      "BgMatrix-C",
      "BgMatrix-AA"
    );
    if (riskID) parent.classList.add(`BgMatrix-${riskID}`);
  });
}

function getMatrix() {
  const url = "/bagong/riskmatrix/find?Type=IBPR";
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

function onRowChange(name, v1, v2, current) {
  setEditabeRow(current.ID);
}

function setEditabeRow(id) {
  let idx = data.record.Lines.findIndex((o) => o.ID == id);
  if (idx > -1) data.record.Lines[idx].IsUpdated = true;
}

function onModal(nv) {
  if (nv) saveManual(data.record, true);
}

function saveManual(r, isNotif) {
  listControl.value.submitForm(
    r,
    () => {},
    () => {},
    isNotif
  );
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
onMounted(() => {
  loadFormConfig(axios, "/she/ibprtrx/currentaction/formconfig").then(
    (r) => {
      data.currentActionCfg = r;
    },
    (e) => util.showError(e)
  );

  getMatrix();
});
</script>

<style>
.ibpr_transaction .BgMatrix-C {
  background: #00ff00;
}
.ibpr_transaction .BgMatrix-B {
  background: #fac090;
}
.ibpr_transaction .BgMatrix-A {
  background: #ff9900;
}
.ibpr_transaction .BgMatrix-AA {
  background: #ff0000;
}
</style>
