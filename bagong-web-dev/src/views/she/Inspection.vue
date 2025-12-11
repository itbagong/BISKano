<template>
  <data-list
    class="card w-full"
    ref="listControl"
    title="Inspection"
    grid-config="/she/inspection/gridconfig"
    form-config="/she/inspection/formconfig"
    grid-read="/she/inspection/gets"
    form-read="/she/inspection/get"
    grid-mode="grid"
    grid-delete="/she/inspection/delete"
    form-insert="/she/inspection/save"
    form-update="/she/inspection/save"
    :init-app-mode="data.appMode"
    :form-tabs-edit="['General', 'Line']"
    :form-tabs-new="['General', 'Line']"
    :form-fields="['Dimension']"
    @formFieldChange="onFormFieldChange"
    @formEditData="editRecord"
    @formNewData="newData"
    @preSave="onPreSave"
    form-keep-label
    stay-on-form-after-save
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
          label="Template"
          class="w-[400px]"
          use-list
          multiple
          :lookup-url="`/she/mcuitemtemplate/find?Menu=SHE-0011`"
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
    <template #form_buttons_1="{ item, inSubmission, loading }">
      <form-buttons-trx
        :disabled="loading"
        :status="item.Status"
        :journal-id="item._id"
        :posting-profile-id="item.PostingProfileID"
        :journal-type-id="'INSPECTION'"
        :moduleid="'SHE'"
        @preSubmit="trxPreSubmit"
        @postSubmit="trxPostSubmit"
        @errorSubmit="trxErrorSubmit"
        :auto-post="!waitTrxSubmit"
      />
    </template>
    <template #form_input_Dimension="{ item }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :default-list="profile.Dimension"
      ></dimension-editor-vertical>
    </template>
    <template #form_tab_Line="{ item }">
      <div class="flex gap-4 mb-4 mt-2">
        <s-grid
          class="w-full"
          :config="data.legendcfg1"
          v-model="data.legend1"
          hide-search
          hide-sort
          hide-refresh-button
          hide-select
          hide-footer
          hide-new-button
          hide-action
          hide-control
          auto-commit-line
          no-confirm-delete
        />
        <s-grid
          class="w-full"
          :config="data.legendcfg2"
          v-model="data.legend2"
          hide-search
          hide-sort
          hide-refresh-button
          hide-select
          hide-footer
          hide-new-button
          hide-action
          hide-control
          auto-commit-line
          no-confirm-delete
        />
      </div>
      <lines
        v-model="item.Lines"
        :template-lines="item.TemplateID"
        :line-cfg="data.lineCfg"
        kind="inspection"
      />
    </template>
  </data-list>
</template>
<script setup>
import { reactive, ref, watch, inject, onMounted, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton, SGrid } from "suimjs";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import helper from "@/scripts/helper.js";
import moment from "moment";
import Pica from "@/components/common/ItemPica.vue";
import { authStore } from "@/stores/auth.js";
import Lines from "./widget/LinesInspectionCsms.vue";

const axios = inject("axios");
const listControl = ref(null);
const profile = authStore().getRBAC(FEATUREID);
layoutStore().name = "tenant";
const FEATUREID = "Inspection";
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
  if (data.search.Template !== null && data.search.Template !== "") {
    filters.push({
      Field: "TemplateID",
      Op: "$eq",
      Value: data.search.Template,
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
  legendcfg1: {},
  legendcfg2: {},
  lineCfg: [
    { field: "IsApplicable", label: "Not Applicable" },
    {
      field: "ValueDescription",
      label: "Value Description",
    },
    {
      field: "Value",
      label: "Value",
    },
    {
      field: "MaxValue",
      label: "Max Value",
    },
    {
      field: "Pica",
      label: "Pica",
    },
    {
      field: "Remark",
      label: "Remark",
    },
    {
      field: "Attachment",
      label: "Attachment",
    },
  ],
  legend1: [
    {
      Hazard: "AA",
      TingkatBahaya: "Bahaya kritis",
      Panduan: "Berhenti,isolasi,SEGERA laporkan keatasan dan perbaiki segera",
    },
    {
      Hazard: "A",
      TingkatBahaya: "Resiko Tinggi",
      Panduan:
        "Berhenti,SEGERA laporkan keatasn,putuskan:lanjutkan dengan catatan atau perbaiaki segera ",
    },
    {
      Hazard: "B",
      TingkatBahaya: "Resiko Sedang",
      Panduan: "Laporkan ke atasn dan perbaiki setelah bahaya AA-A selesai",
    },
    {
      Hazard: "C",
      TingkatBahaya: "Resiko rendah",
      Panduan: "Laporkan ke atasan dan perbaiki setelah bahaya AA-A-B selesai",
    },
  ],
  legend2: [
    { Value: "0", Ket: "Tidak ada/Tidak dilakukan sama sekali" },
    {
      Value: "1",
      Ket: "Telah dilakukan namun belum memenuhi kriteria/Beberapa item yang tidak sesuai/Tidak ada ",
    },
    {
      Value: "2",
      Ket: "Telah memenuhi kriteria dan sesuai pada item inspeksi",
    },
  ],
  search: {
    DateFrom: null,
    DateTo: null,
    No: "",
    Location: "",
    LocationDetail: "",
    Name: "",
    Template: "",
    Status: "",
  },
});

function newData(r) {
  openForm(r);
}

function editRecord(r) {
  openForm(r);
}

function openForm(r) {
  data.record = r;
}

function onPreSave(r) {
  for (let i in r.Lines) {
    let o = r.Lines[i];
    o.Value = o.Value == null ? 0 : parseInt(o.Value);
  }
}

function genCfg() {
  let cfg1 = [
    {
      field: "Hazard",
      label: "HAZARD",
    },
    {
      field: "TingkatBahaya",
      label: "TINGKAT BAHAYA ",
    },
    {
      field: "Panduan",
      label: "PANDUAN DALAM MELAKUKAN TINDAKAN PENGENDALIAN BAHAYA ",
    },
  ];
  let cfg2 = [
    {
      field: "Value",
      label: "HAZARD",
    },
    {
      field: "Ket",
      label: "TINGKAT BAHAYA ",
    },
  ];

  let tmp1 = [];
  for (let i in cfg1) {
    let o = cfg1[i];
    tmp1.push(helper.gridColumnConfig(o));
  }
  data.legendcfg1.fields = tmp1;

  let tmp2 = [];
  for (let i in cfg2) {
    let o = cfg2[i];
    tmp2.push(helper.gridColumnConfig(o));
  }
  data.legendcfg2.fields = tmp2;
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
onMounted(() => {
  genCfg();
});
</script>
