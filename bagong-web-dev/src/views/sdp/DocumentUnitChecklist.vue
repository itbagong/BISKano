<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/sdp/documentunitchecklist/gridconfig"
      form-config="/sdp/documentunitchecklist/formconfig"
      grid-read="/sdp/documentunitchecklist/gets"
      form-read="/sdp/documentunitchecklist/get"
      grid-mode="grid"
      grid-delete="/sdp/documentunitchecklist/delete"
      form-keep-label
      grid-hide-sort
      form-insert="/sdp/documentunitchecklist/insert"
      form-update="/sdp/documentunitchecklist/update"
      :grid-fields="['WONo', 'SUNID', 'AssetID', 'ChasisNo', 'Status']"
      :form-fields="[
        'SUNID',
        'AssetID',
        'HullNo',
        'ChasisNo',
        'SRUTDate',
        'SKRBDate',
        'SubmissionDateSRUT',
        'SubmissionDatePolres',
        'SubmissionDateRekomPeruntukan',
        'SubmissionDateSamsat',
        'SubmissionDateUjiKIR',
        'RoutePermitDate',
        'Status',
        'Dimension',
      ]"
      :form-tabs-new="['General']"
      :form-tabs-edit="['General']"
      :form-tabs-view="['General']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :grid-custom-filter="customFilter"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @pre-save="preSave"
      @controlModeChanged="onCancelForm"
      @alter-form-config="onalterFormConfig"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <!-- @grid-refreshed="gridRefreshed" -->
      <template #grid_header_search="{ config }">
        <s-input
          kind="text"
          label="Search SUNID"
          class="w-full"
          v-model="data.searchData.SUNID"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          kind="text"
          label="Chasis No."
          class="w-[400px]"
          v-model="data.searchData.ChasisNo"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          kind="text"
          label="Status"
          class="w-[500px]"
          v-model="data.searchData.Status"
          use-list
          :items="[
            'Pengajuan SRUT',
            'Pengajuan Rekom Peruntukan',
            'Pengajuan Samsat',
            'Pengajuan Uji KIR',
            'Pengajuan Polres',
            'Need Action',
          ]"
          @change="refreshData"
        ></s-input>
      </template>
      <template #grid_AssetID="{item}">
        <span class="w-full">{{item.AssetName}}</span>
      </template>
      <template #form_input_SubmissionDateSRUT="{ item }">
        <s-input
          kind="date"
          label="Submission Date SRUT"
          keep-label
          class="w-full"
          v-model="item.SubmissionDateSRUT"
        ></s-input>
      </template>
      <template #form_input_SubmissionDatePolres="{ item }">
        <s-input
          kind="date"
          label="Submission Date Polres"
          keep-label
          class="w-full"
          v-model="item.SubmissionDatePolres"
        ></s-input>
      </template>
      <template #form_input_SubmissionDateRekomPeruntukan="{ item }">
        <s-input
          kind="date"
          label="Submission Date Rekom Peruntukan"
          keep-label
          class="w-full"
          v-model="item.SubmissionDateRekomPeruntukan"
        ></s-input>
      </template>
      <template #form_input_SubmissionDateSamsat="{ item }">
        <s-input
          kind="date"
          label="Submission Date Samsat"
          keep-label
          class="w-full"
          v-model="item.SubmissionDateSamsat"
        ></s-input>
      </template>
      <template #form_input_SubmissionDateUjiKIR="{ item }">
        <s-input
          kind="date"
          label="Submission Date Uji KIR"
          keep-label
          class="w-full"
          v-model="item.SubmissionDateUjiKIR"
        ></s-input>
      </template>
      <template #form_input_RoutePermitDate="{ item }">
        <s-input
          kind="date"
          label="Route Permit Date"
          keep-label
          class="w-full"
          v-model="item.RoutePermitDate"
        ></s-input>
      </template>
      <template #form_input_SRUTDate="{ item }">
        <s-input
          kind="date"
          label="SRUT Date"
          keep-label
          class="w-full"
          v-model="item.SRUTDate"
        ></s-input>
      </template>
      <template #form_input_SKRBDate="{ item }">
        <s-input
          kind="date"
          label="SKRB Date"
          keep-label
          class="w-full"
          v-model="item.SKRBDate"
        ></s-input>
      </template>
      <template #form_input_SUNID="{ item }">
        <s-input
          kind="text"
          label="SUN ID"
          keep-label
          class="w-full"
          v-model="item.SUNID"
        ></s-input>
      </template>
      <template #form_input_HullNo="{ item }">
        <s-input
          kind="text"
          keep-label
          label="Hull No"
          class="w-full"
          v-model="item.HullNo"
        ></s-input>
      </template>
      <template #form_input_AssetID="{ item }">
        <s-input
          use-list
          kind="text"
          label="Asset ID"
          class="w-full"
          v-model="item.AssetID"
          :lookup-url="'/tenant/asset/find'"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="
            (field, v1, v2, old, ctlRef) => {
              getAssetData(v1, v2, item);
            }
          "
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
      <template #grid_Status="{ item }">
        <p
          v-if="
            item.StatusSRUT == '' &&
            item.StatusRekomPeruntukan == '' &&
            item.StatusSamsat == '' &&
            item.StatusUjiKIR == '' &&
            item.StatusRoutePermit == '' &&
            item.StatusFinal == ''
          "
        >
          Need Action
        </p>
        <p v-else-if="item.StatusFinal != ''">{{ item.StatusFinal }}</p>
        <p v-else-if="item.StatusSRUT != ''">{{ item.StatusSRUT }}</p>
        <p v-else-if="item.StatusRekomPeruntukan != ''">
          {{ item.StatusRekomPeruntukan }}
        </p>
        <p v-else-if="item.StatusSamsat != ''">{{ item.StatusSamsat }}</p>
        <p v-else-if="item.StatusUjiKIR != ''">{{ item.StatusUjiKIR }}</p>
        <p v-else-if="item.StatusRoutePermit != ''">
          {{ item.StatusRoutePermit }}
        </p>
      </template>
    </data-list>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, inject, computed, onMounted } from "vue";
import { layoutStore } from "../../stores/layout.js";
import moment from "moment";
import { authStore } from "@/stores/auth.js";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import {
  DataList,
  util,
  SForm,
  SInput,
  createFormConfig,
  SButton,
} from "suimjs";
import documentunitchecklistLine from "./widget/documentunitchecklistLine.vue";
import { useRoute } from "vue-router";

layoutStore().name = "tenant";

const featureID = "DocumentUnitChecklist";
// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})
const profile = authStore().getRBAC(featureID);

const listControl = ref(null as any);
const lineConfig = ref(null);
const axios = inject("axios");
const route = useRoute();
let customFilter = computed(() => {
  const filters = [];
  if (data.searchData.SUNID !== null && data.searchData.SUNID !== "") {
    filters.push({
      Field: "SUNID",
      Op: "$contains",
      Value: [data.searchData.SUNID],
    });
  }
  if (data.searchData.ChasisNo !== null && data.searchData.ChasisNo !== "") {
    filters.push({
      Field: "ChasisNo",
      Op: "$contains",
      Value: [data.searchData.ChasisNo],
    });
  }
  if (data.searchData.Status !== null && data.searchData.Status !== "") {
    // filters.push({
    //   Field: "ChasisNo",
    //   Op: "$contains",
    //   Value: [data.searchData.Status],
    // });

    if (data.searchData.Status == "Pengajuan SRUT") {
      filters.push({
        Field: "StatusSRUT",
        Op: "$ne",
        Value: "",
      });
    } else if (data.searchData.Status == "Pengajuan Rekom Peruntukan") {
      filters.push({
        Field: "StatusRekomPeruntukan",
        Op: "$ne",
        Value: "",
      });
    } else if (data.searchData.Status == "Pengajuan Samsat") {
      filters.push({
        Field: "StatusSamsat",
        Op: "$ne",
        Value: "",
      });
    } else if (data.searchData.Status == "Pengajuan Uji KIR") {
      filters.push({
        Field: "StatusUjiKIR",
        Op: "$ne",
        Value: "",
      });
    } else if (data.searchData.Status == "Pengajuan Polres") {
      filters.push({
        Field: "StatusFinal",
        Op: "$ne",
        Value: "",
      });
    } else {
      filters.push({
        Field: "StatusSRUT",
        Op: "$eq",
        Value: "",
      });
      filters.push({
        Field: "StatusFinal",
        Op: "$eq",
        Value: "",
      });
      filters.push({
        Field: "StatusRekomPeruntukan",
        Op: "$eq",
        Value: "",
      });
      filters.push({
        Field: "StatusSamsat",
        Op: "$eq",
        Value: "",
      });
      filters.push({
        Field: "StatusUjiKIR",
        Op: "$eq",
        Value: "",
      });
      filters.push({
        Field: "StatusRoutePermit",
        Op: "$eq",
        Value: "",
      });
    }
  }
  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});
const roleID = [
  (v) => {
    let vLen = 0;
    let consistsInvalidChar = false;

    v.split("").forEach((ch) => {
      vLen++;
      const validCar =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
          ch
        ) >= 0;
      if (!validCar) consistsInvalidChar = true;
    });

    if (vLen < 3 || consistsInvalidChar)
      return "minimal length is 3 and alphabet only";
    return "";
  },
];

const data = reactive({
  title: null as any,
  appMode: "grid",
  formMode: "edit",
  titleForm: "Document Unit Checklist",
  record: [],
  searchData: {
    SUNID: "",
    ChasisNo: "",
    Status: "",
  },
  allowDelete: route.query.allowdelete === "true",
  formAssets: {},
  isSelected: false,
});

watch(
  () => route.query.objname,
  (nv) => {
    util.nextTickN(2, () => {
      listControl.value.refreshList();
      listControl.value.refreshForm();
    });
  }
);

watch(
  () => route.query.title,
  (nv) => {
    data.title = nv;
    listControl.value.setControlMode("grid");
  }
);

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}

function preSave(record) {
  if (record.StartPeriod == null) {
    record.StartPeriod = moment().format("YYYY-MM-DDThh:mm:ssZ");
  }

  if (record.EndPeriod == null) {
    record.EndPeriod = moment().format("YYYY-MM-DDThh:mm:ssZ");
  }

  // if (record.AssetID != "") {
  //   axios.post("/tenant/asset/find?_id="+record.AssetID).then((resAsset) => {
  //     var data = resAsset.data[0]

  //     console.log("data =>", data)
  //     record.AssetName = record.AssetID + " | " + data.Name
  //   })
  // }
}

async function getWorkOrder(value, item) {
  const res = await axios.post("/mfg/workorder/find?SunID=" + value);

  const response = res.data[0];

  item.AssetID = response.EquipmentNo;
  item.ChasisNo = "";
  getAssetData(response.EquipmentNo, "", item);
}

async function getAssetData(id, val, item) {
  if (id != "") {
    const res = await axios.post("/bagong/asset/find?_id=" + id);

    const response = res.data[0];

    item.AssetName = val
    item.EngineNo = response.DetailUnit.MachineNum
    item.ChasisNo = response.DetailUnit.ChassisNum;
  } else {
    util.showError("Asset ID is empty!!");
  }
}

function newRecord() {
  data.titleForm = "Create New Document Unit Checklist";
  openForm();
}

function lookupPayloadBuilder(search, config, value) {
  const qp = {};

  return qp;
}

function editRecord(record) {
  data.titleForm = `Edit Document Unit Checklist | ${record._id}`;
  openForm();
}

function gridRefreshed() {
  if (
    data.searchData.SUNID != "" ||
    data.searchData.ChasisNo != "" ||
    data.searchData.Status != ""
  ) {
    axios.post("/sdp/documentunitchecklist/gets-filter", data.searchData).then(
      async (r) => {
        listControl.value.setGridRecords(r.data.data);
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
function onAlterGridConfig(cfg) {
  cfg.setting.idField = "Created";
  cfg.setting.sortable = ["_id", "Created"];
}

function onPostSave(record) {
  // let lines = lineConfig.value.getDataValue();
  // let payloadBatch = {
  //   documentunitchecklistID: record._id,
  //   Lines: lines.filter(function (b) {
  //     return b.documentunitchecklistID != "";
  //   }),
  // };
  // console.log("payloadBatch =>", lines)
  // axios
  //   .post("/sdp/documentunitchecklistline/save-multiple", payloadBatch)
  //   .then(
  //     (r) => {
  //       console.log("rrr =>", r)
  //     },
  //     (e) => {}
  //   );
}

function onCancelForm(mode) {
  if (mode === "grid") {
    data.titleForm = "Document Unit Checklist";
  }
}

const addCancel = () => {
  data.formMode = "new";
  // record._id = "";
  // record.TrxDate = new Date();
  // record.Status = "";
  data.titleForm = "Create New Document Unit Checklist";
  // openForm(record);
};

const onsubmit = () => {};

function onalterFormConfig(r) {
  console.log(r);
}
</script>
