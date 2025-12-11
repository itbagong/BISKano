<template>
  <data-list
    v-if="data.kind.urlGeneral"
    class="card master_IBPR"
    ref="listControl"
    :title="data.kind.title"
    :grid-config="`/she/${data.kind.urlGeneral}/gridconfig`"
    :form-config="`/she/${data.kind.urlGeneral}/formconfig`"
    :grid-read="`/she/${data.kind.urlGeneral}/gets`"
    :form-read="`/she/${data.kind.urlGeneral}/get`"
    grid-mode="grid"
    :grid-delete="`/she/${data.kind.urlGeneral}/delete`"
    form-keep-label
    :form-insert="`/she/${data.kind.urlGeneral}/save`"
    :form-update="`/she/${data.kind.urlGeneral}/save`"
    :init-app-mode="data.appMode"
    grid-hide-select
    @form-edit-data="openForm"
    @form-new-data="newData"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    stay-on-form-after-save
    @form-field-change="onFieldChange"
    :form-tabs-edit="['General', 'Lines']"
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
          ref="refLocation"
          v-model="data.search.IBPRTeam"
          lookup-key="_id"
          label="IBPR Team "
          class="w-[400px]"
          use-list
          multiple
          :lookup-url="`/tenant/employee/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
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
    <template #form_tab_Lines="{ item, mode }">
      <div class="relative h-[650px] overflow-y-scroll overflow-x-hidden">
        <s-grid
          class="master_IBPR_Lines"
          ref="gridLines"
          :config="data.cfgGrid"
          hide-search
          hide-sort
          hide-refresh-button
          hide-edit
          hide-select
          hide-footer
          editor
          auto-commit-line
          no-confirm-delete
          @new-data="newLine"
          @delete-data="onDeleteDetail"
        >
          <template #item_SituasiAktivitas="{ item }">
            <div class="grid grid-cols-1 gap-2" v-if="!item.ParentId">
              <s-input
                v-model="item.SituasiAktivitas"
                use-list
                lookup-url="/tenant/masterdata/find?MasterDataTypeID=IBPRActivity"
                lookup-key="_id"
                :lookup-labels="['Name']"
                :lookup-searchs="['_id', 'Name']"
              />
              <div class="grid grid-cols-2 gap-2 pb-4">
                <s-button
                  label="add detail"
                  tooltip="add detail"
                  class="btn_success"
                  @click="addDetail(item)"
                />
                <s-button
                  label="delete"
                  tooltip="delete"
                  class="btn_primary"
                  @click="onDeleteLine(item)"
                />
              </div>
            </div>
            <div v-else></div>
          </template>
          <template #item_Category="{ item }">
            <div class="grid grid-cols-1 gap-2" v-if="!item.ParentId">
              <s-input v-model="item.Category" />
              <div class="grid grid-cols-2 gap-2 pb-4">
                <s-button
                  label="add detail"
                  tooltip="add detail"
                  class="btn_success"
                  @click="addDetail(item)"
                />
                <s-button
                  label="delete"
                  tooltip="delete"
                  class="btn_primary"
                  @click="onDeleteLine(item)"
                />
              </div>
            </div>
            <div v-else></div>
          </template>
        </s-grid>
      </div>
    </template>
    <template #form_input_Dimension="{ item, mode }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :read-only="mode == 'view'"
      ></dimension-editor-vertical>
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
} from "suimjs";
import { layoutStore } from "@/stores/layout.js";
layoutStore().name = "tenant";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import { useRoute } from "vue-router";

const listControl = ref(null);
const gridLines = ref(null);
const axios = inject("axios");
const route = useRoute();
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
  if (data.search.IBPRTeam !== null && data.search.IBPRTeam.length > 0) {
    filters.push({
      Field: "IBPRTeam",
      Op: "$eq",
      Value: data.search.IBPRTeam,
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
  cfgGrid: {},
  kind: {},
  search: {
    DateFrom: null,
    DateTo: null,
    IBPRTeam: [],
    Location: "",
    LocationDetail: "",
    Name: "",
    Status: "",
  },
});

function openForm(r) {
  data.record = r;
  util.nextTickN(3, () => {
    updateGridLines();
  });
}

const mapMaster = {
  IBPR: {
    urlGeneral: "masteribpr",
    urlLines: "masteribpr/line",
    title: "Master IBPR",
  },
  RSCA: {
    urlGeneral: "masterrsca",
    urlLines: "masterrsca/lines",
    title: "Master RSCA",
  },
};

function newData(r) {
  data.record = r;
}

function newLine(r) {
  r = {};
  const noLine = data.record.Lines.length + 1;

  r.ID = util.uuid();
  r.ParentId = "";
  r.LineNo = noLine;

  data.record.Lines.push(r);
  updateGridLines();
}

function updateGridLines() {
  let filtered = data.record.Lines.filter((o) => o.ParentId == "");
  for (let i in filtered) {
    let o = filtered[i];
    o.LineNo = (parseInt(i) + 1).toString();
  }
  gridLines.value.setRecords(data.record.Lines);
  util.nextTickN(3, () => {
    formatRowSpan();
  });
}

function onDeleteLine(dt) {
  let childList = data.record.Lines.filter((o) => o.ParentId == dt.ID);
  for (let i in childList) {
    let c = childList[i];
    let findex = data.record.Lines.findIndex((o) => o.ID == c.ID);
    data.record.Lines.splice(findex, 1);
  }
  let findex = data.record.Lines.findIndex((o) => o.ID == dt.ID);
  data.record.Lines.splice(findex, 1);
  updateGridLines();
}

function onDeleteDetail(dt, idx) {
  data.record.Lines.splice(idx, 1);
  updateGridLines();
}

function addDetail(r) {
  let o = {
    LineNo: r.LineNo,
    ID: util.uuid(),
    ParentId: r.ID,
  };

  let idx = data.record.Lines.findIndex((o) => o.ID == r.ID);
  let lenChild = data.record.Lines.filter((o) => o.ParentId == r.ID);
  let idxSlice = idx + lenChild.length;
  data.record.Lines.splice(idxSlice + 1, 0, o);
  updateGridLines();
}

function formatRowSpan() {
  const myTable = document.querySelectorAll(
    ".master_IBPR_Lines .suim_table [name='grid_body']"
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

function loadGridLines(kind) {
  let url = `/she/${kind}/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      const hideField = ["CreatedBy", "UpdatedBy"];
      for (let i in hideField) {
        let o = hideField[i];
        let findex = r.fields.findIndex((x) => x.field == o);
        r.fields.splice(findex, 1);
      }
      data.cfgGrid = r;
    },
    (e) => {}
  );
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
watch(
  () => route.query.id,
  (nv) => {
    data.search = {
      DateFrom: null,
      DateTo: null,
      IBPRTeam: [],
      Location: "",
      LocationDetail: "",
      Name: "",
      Status: "",
    };
    data.kind = mapMaster[route.query.id];
    util.nextTickN(2, () => {
      if (["IBPR", "RSCA"].includes(nv)) {
        listControl.value.refreshList();
        listControl.value.refreshForm();
      }
    });
  }
);
onMounted(() => {
  data.kind = mapMaster[route.query.id];
  loadGridLines(data.kind.urlLines);
});
</script>

<style>
.master_IBPR .w-full.items-start.gap-2.grid.gridCol2 {
  @apply gap-4;
}

.master_IBPR .master_IBPR_Lines th {
  @apply sticky top-0 bg-gray-100;
}
</style>
