<template>
  <data-list
    class="card w-full"
    ref="listControl"
    title="Observasi"
    grid-config="/she/observasi/gridconfig"
    form-config="/she/observasi/formconfig"
    grid-read="/she/observasi/gets"
    form-read="/she/observasi/get"
    grid-mode="grid"
    grid-delete="/she/observasi/delete"
    form-insert="/she/observasi/save"
    form-update="/she/observasi/save"
    :init-app-mode="data.appMode"
    :form-tabs-edit="['General', 'Line']"
    :form-tabs-new="['General', 'Line']"
    :form-fields="['Dimension']"
    @formFieldChange="onFormFieldChange"
    @formEditData="editRecord"
    @formNewData="newData"
    form-keep-label
    stay-on-form-after-save
    form-hide-submit
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
          ref="refName"
          v-model="data.search.Name"
          label="Name"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refObservee"
          v-model="data.search.Observee"
          lookup-key="_id"
          label="Observee"
          class="w-[400px]"
          use-list
          :lookup-url="`/tenant/customer/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refObserver"
          v-model="data.search.Observer"
          lookup-key="_id"
          label="Observer"
          class="w-[400px]"
          use-list
          :lookup-url="`/tenant/employee/find`"
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
        :journal-type-id="'Observasi'"
        :moduleid="'SHE'"
        @preSubmit="trxPreSubmit"
        @postSubmit="trxPostSubmit"
        @errorSubmit="trxErrorSubmit"
        :auto-post="!waitTrxSubmit"
      />
    </template>
    <template #form_buttons_2="">
      <s-button
        class="btn_primary submit_btn"
        label="save"
        icon="content-save"
        @click="onSave()"
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
      </div>
      <lines
        v-model="item.Lines"
        :template-lines="item.TemplateID"
        :line-cfg="data.lineCfg"
        kind="observasi"
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
  if (data.search.Observee !== null && data.search.Observee !== "") {
    filters.push({
      Field: "Observee",
      Op: "$eq",
      Value: data.search.Observee,
    });
  }
  if (data.search.Observer !== null && data.search.Observer !== "") {
    filters.push({
      Field: "Observer",
      Op: "$eq",
      Value: data.search.Observer,
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
  lineCfg: [
    { field: "IsApplicable", label: "Not Applicable" },
    {
      field: "Deviation",
      label: "Deviation Finding",
    },
    {
      field: "HazardCode",
      label: "Hazard Code",
    },
    {
      field: "HarzardLevel",
      label: "Hazard Level and Control",
    },
    {
      field: "Result",
      label: "Result",
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
      Possibility:
        "Kematian lebih dari 1 (multiple Fatality), Kerusakan Property > Rp 25.000.000",
      HazardCode: "AA",
      HazardLevel: "Bahaya Kritis (Stop dan Perbaiki Segera)",
    },
    {
      Possibility:
        "1 (Satu) kematian (Fatality), Kerusakan Property > Rp 10.000.000",
      HazardCode: "A",
      HazardLevel: "Resiko Tinggi (Perbaiki dalam 6 jam)",
    },
    {
      Possibility:
        "Perawatan ringan di lokasi namun tidak bisa langsung bekerja",
      HazardCode: "B",
      HazardLevel: "Resiko Sedang (Perbaiki dalam 3 hari)",
    },
    {
      Possibility:
        "Korban hanya memerlukan penanganan ringan di lokasi dan langsung dapat bekerja, Kerusakan Property Rp. < Rp. 1.000.000",
      HazardCode: "C",
      HazardLevel: "Resiko Rendah (Perbaiki, tidak prioritas)",
    },
  ],
  search: {
    DateFrom: null,
    DateTo: null,
    No: "",
    Name: "",
    Observee: "",
    Observer: "",
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

function genCfg() {
  let cfg1 = [
    {
      field: "Possibility",
      label: "POSSIBILITY CONSEQUENCE",
    },
    {
      field: "HazardCode",
      label: "HAZARD CODE",
    },
    {
      field: "HazardLevel",
      label: "HAZARD LEVEL AND CONTROL",
    },
  ];

  let tmp1 = [];
  for (let i in cfg1) {
    let o = cfg1[i];
    tmp1.push(helper.gridColumnConfig(o));
  }
  data.legendcfg1.fields = tmp1;
}

function onSave() {
  validateLines(data.record.Lines);
}

async function validateLines(r) {
  let isError = false;
  let maxLe = r.length;
  for (let i in r) {
    let o = r[i];
    if (!o.Result && !isError && o.TemplateLine.Parent !== "") {
      let fileAttach = await GetDataAssets(
        `SHE_observasi_LINES_${o.Attachment}`,
        "SHE_observasi",
        []
      );
      if (fileAttach.length == 0 || o.Remark == "") {
        isError = true;
        util.showError("if result is no remark & attachment must be required");
      }
    }
    if (parseInt(i) + 1 == maxLe && !isError) {
      listControl.value.submitForm(
        data.record,
        () => {},
        () => {}
      );
    }
  }
}

async function GetDataAssets(id, type, tags) {
  try {
    let param = {
      JournalType: type,
      JournalID: id,
      Tags: tags,
    };
    const resp = await axios.post("/asset/read-by-journal", param);
    return resp.data.map((resp) => ({
      _id: resp._id,
      OriginalFileName: resp.OriginalFileName,
      Content: resp.Data.Content,
      ContentType: resp.ContentType,
      FileName: resp.OriginalFileName,
      UploadDate: resp.Data.UploadDate,
      Description: resp.Data.Description,
      URI: resp.URI,
      Kind: resp.Kind,
      Tags: resp.Tags,
      RefID: resp.RefID,
    }));
  } catch (error) {
    util.showError(error);
  }
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
