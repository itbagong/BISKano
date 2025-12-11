<template>
  <div class="w-full">
    <data-list
      class="card SHE_SIDAK"
      ref="listControl"
      title="Sidak"
      grid-config="/she/sidak/gridconfig"
      form-config="/she/sidak/formconfig"
      grid-read="/she/sidak/gets"
      form-read="/she/sidak/get"
      grid-mode="grid"
      grid-delete="/she/sidak/delete"
      form-keep-label
      form-insert="/she/sidak/save"
      form-update="/she/sidak/save"
      :form-tabs-edit="data.tabsList"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      grid-hide-select
      @form-edit-data="openForm"
      @form-new-data="newRecord"
      :form-fields="['Dimension', 'Penalty', 'DateTime', 'Mess']"
      :grid-fields="[
        'EmployeeID',
        'Drug',
        'Fatigue',
        'SpeedGun',
        'Alcohol',
        'Position',
      ]"
      @pre-save="onPreSave"
      @form-field-change="onFieldChange"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="customFilter"
    >
      <template #grid_header_search="{ config }">
        <div
          class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter"
        >
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
            lookup-key="_id"
            label="Name"
            class="w-[400px]"
            use-list
            :lookup-url="`/tenant/employee/find`"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refFittoWork"
            v-model="data.search.FittoWork"
            label="Fit to Work"
            class="w-[200px]"
            use-list
            :items="['Yes', 'No']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refSpeedGun"
            v-model="data.search.SpeedGun"
            label="Speed Gun"
            class="w-[200px]"
            use-list
            :items="['Negative', 'Positive']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refMess"
            v-model="data.search.Mess"
            label="Mess"
            class="w-[200px]"
            use-list
            :items="['Yes', 'No']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refJabatan"
            v-model="data.search.Jabatan"
            lookup-key="_id"
            label="Jabatan"
            class="w-[200px]"
            use-list
            :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=PTE`"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refDrug"
            v-model="data.search.Drug"
            label="Drug"
            class="w-[200px]"
            use-list
            :items="['Negative', 'Positive']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refAlcohol"
            v-model="data.search.Alcohol"
            label="Alcohol"
            class="w-[200px]"
            use-list
            :items="['Negative', 'Positive']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refSite"
            v-model="data.search.Site"
            lookup-key="_id"
            label="Site"
            class="w-[200px]"
            use-list
            :disabled="defaultList?.length === 1"
            :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
            :lookup-labels="['Label']"
            :lookup-searchs="['_id', 'Label']"
            :lookup-payload-builder="
              defaultList?.length > 0
                ? (...args) =>
                    helper.payloadBuilderDimension(
                      defaultList,
                      data.search.Site,
                      false,
                      ...args
                    )
                : undefined
            "
            @change="refreshData"
          ></s-input>
        </div>
      </template>
      <template #grid_Alcohol="{ item }">
        <div
          class="p-2"
          :class="bgByParam('SSHEAlcohol', item.Alcohol.Alcohol)"
        >
          <div class="flex">
            Alcohol :
            {{ item.Alcohol.Alcohol }}
          </div>
          <div class="flex">
            Remark
            {{ item.Alcohol.Remark }}
          </div>
        </div>
      </template>
      <template #grid_SpeedGun="{ item }">
        <div
          class="p-2"
          :class="bgByParam('SSHESpeedGun', item.SpeedGun.Speed)"
        >
          <div class="flex">
            Speed Gun :
            {{ item.SpeedGun.Speed }}
          </div>
          <div class="flex">
            Location :
            {{ item.SpeedGun.LocationDetail }}
          </div>
        </div>
      </template>
      <template #grid_Fatigue="{ item }">
        <div
          class="p-2"
          :class="bgByParam('SSHEFatigue', item.Fatigue.SleepDuration)"
        >
          <div class="flex">
            Sleep Duration :
            {{ item.Fatigue.SleepDuration }}
          </div>
          <div class="flex">
            Remark :
            {{ item.Fatigue.Remark }}
          </div>
        </div>
      </template>
      <template #grid_Drug="{ item }">
        {{ viewGridSidak(item.Drug) }}
      </template>
      <template #grid_EmployeeID="{ item }">
        <s-input
          v-model="item.EmployeeID"
          use-list
          :lookup-url="'/tenant/employee/find'"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :key="item_id"
          read-only
        />
      </template>
      <template #grid_Position="{ item }">
        <s-input
          v-model="item.Position"
          use-list
          :lookup-url="'/tenant/masterdata/find?MasterDataTypeID=PTE'"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :key="item_id"
          read-only
        />
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #form_input_Penalty="{ item }">
        <div class="section grow" v-if="data.isNotice">
          <div class="title section_title">Penalty</div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <attachment :label="'SP'" v-model="item.Penalty.SP" />
              <attachment :label="'PHK'" v-model="item.Penalty.PHK" />
              <s-input
                v-model="item.Penalty.Reason"
                kind="text"
                :multi-row="5"
                class="mb-2"
                label="Reason"
              />
            </div>
          </div>
        </div>
        <div v-else></div>
      </template>
      <template #form_input_DateTime="{ item, mode }">
        <label class="input_label">
          <div>Date</div>
        </label>
        <div class="flex gap-2" v-if="mode == 'edit'">
          {{ moment(item.DateTime).format("DD MMM YYYY HH:mm") }}
        </div>
        <div class="flex gap-2" v-else>
          <s-input kind="date" v-model="data.record.bookDate" />
          <s-input kind="time" v-model="data.record.bookTime" />
        </div>
      </template>
      <template #form_input_Mess="{ item }">
        <label class="input_label">
          <div>Mess</div>
        </label>
        <s-toggle
          v-model="item.Mess"
          class="w-[120px] mt-0.5"
          yes-label="Mess"
          no-label="Not Mess"
        />
      </template>
      <template #form_tab_Fatigue="{ item }">
        <fantigue v-model="item.Fatigue" :jurnalId="item._id" />
      </template>
      <template #form_tab_Speed_Gun="{ item }">
        <speed-gun v-model="item.SpeedGun" :jurnalId="item._id" />
      </template>
      <template #form_tab_Alcohol="{ item }">
        <alcohol v-model="item.Alcohol" :jurnalId="item._id" />
      </template>
      <template #form_tab_Drug="{ item }">
        <drug v-model="item.Drug" :jurnalId="item._id" />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, watch, computed, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SInput, util, SButton, SForm } from "suimjs";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import Fantigue from "./widget/Sidak/SidakFantigue.vue";
import SpeedGun from "./widget/Sidak/SidakSpeedGun.vue";
import Alcohol from "./widget/Sidak/SidakAlcohol.vue";
import Drug from "./widget/Sidak/SidakDrug.vue";
import moment from "moment";
import { authStore } from "@/stores/auth.js";
import SToggle from "@/components/common/SButtonToggle.vue";

const FEATUREID = "Sidak";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const listControl = ref(null);

layoutStore().name = "tenant";
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
      Field: "EmployeeID",
      Op: "$eq",
      Value: data.search.Name,
    });
  }
  if (data.search.Jabatan !== null && data.search.Jabatan !== "") {
    filters.push({
      Field: "Position",
      Op: "$eq",
      Value: data.search.Jabatan,
    });
  }
  if (
    data.search.Site !== undefined &&
    data.search.Site !== null &&
    data.search.Site !== ""
  ) {
    filters.push(
      {
        Field: "Dimension.Key",
        Op: "$eq",
        Value: "Site",
      },
      {
        Field: "Dimension.Value",
        Op: "$eq",
        Value: data.search.Site,
      }
    );
  }

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  tabsList: ["General", "Fatigue", "Speed Gun", "Alcohol", "Drug"],
  record: {},
  isNotice: false,
  minSidakVal: {},
  search: {
    DateFrom: null,
    DateTo: null,
    Name: "",
    FittoWork: "",
    SpeedGun: "",
    Mess: "",
    Jabatan: "",
    Drug: "",
    Alcohol: "",
    Site: "",
  },
});

const modelUpload = {
  ID: "",
  Name: "",
};

function resetBookDate() {
  data.record.bookDate = new Date();
  data.record.bookTime = new Date();
}

function resetPenalty() {
  data.record.Penalty = {
    SP: [],
    PHK: [],
    Reason: "",
  };
}

function showNotice(r) {
  let res = false;
  const Drugs = [
    "Amphetam",
    "Morphin",
    "Marijuana",
    "Benzodiaze",
    "Menthapet",
    "Cocain",
  ];

  for (let ky in r.Drug) {
    let val = r.Drug[ky];
    if (Drugs.includes(ky) && val) res = true;
  }

  if (r.Alcohol && r.Alcohol.Alcohol > 0) res = true;

  if (r.SpeedGun && r.SpeedGun.Speed > 0) res = true;

  if (res) resetPenalty();
  data.isNotice = res;
}

function openForm(r) {
  let myDate =
    new Date(r.DateTime).getFullYear() < 1900 ? new Date() : r.DateTime;
  r.bookDate = moment(myDate).format("YYYY MM DD");
  r.bookTime = moment(myDate).format("HH:mm");
  data.record = r;
}

function getPosition(id) {
  if (!id) {
    setTimeout(() => {
      setFieldAttr("Position", "readOnly", false);
    }, 200);
    data.record.Position = "";
    setTimeout(() => {
      setFieldAttr("Position", "readOnly", true);
    }, 200);
    return;
  }
  const url = `/bagong/employeedetail/find?EmployeeID=${id}`;
  setFieldAttr("Position", "readOnly", false);
  axios.post(url).then(
    (r) => {
      if (r.data.length > 0) {
        data.record.Position = r.data[0].Position;
        setFieldAttr("Position", "readOnly", true);
      }
    },
    (e) => {
      util.showError(e.error);
    }
  );
}

function setFieldAttr(field, attr, val) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr(field, attr, val);
  });
}

function onPreSave(r) {
  let dt = moment(r.bookDate).format("YYYY MM DD");
  let time = r.bookTime + ":00";
  r.DateTime = moment(dt + " " + time).format();
}

function viewGridSidak(obj) {
  let res = [];
  for (let ky in obj) {
    if (ky[obj]) res.push(ky);
  }
  return res.length > 0 ? res.toString() : "-";
}

function getValSidak() {
  const url = "/tenant/masterdata/find?MasterDataTypeID=SSHE";
  axios.post(url).then(
    (r) => {
      for (let i in r.data) {
        let o = r.data[i];
        data.minSidakVal[o._id] = parseFloat(o.Name);
      }
    },
    (e) => util.showError(e)
  );
}

function bgByParam(field, val) {
  let res = val < data.minSidakVal[field] ? "red" : "green";
  return "bg" + res + "-100";
}

function newRecord(r) {
  r.DateTime = new Date();
  data.record = r;
  openForm(r);
}

function onFieldChange(name, v1, v2) {
  if (name == "EmployeeId") {
    getPosition(v1);
  }
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
watch(
  () => listControl.value?.getFormCurrentTab(),
  (nv) => {
    if (nv == 0) showNotice(data.record);
  },
  { deep: true }
);

onMounted(() => {
  getValSidak();
});
</script>

<style>
.SHE_SIDAK .box-upload {
  @apply border-2 border-dashed border-zinc-200 w-20 h-20 mr-2 rounded-md grid place-content-center relative;
}

.SHE_SIDAK td:has(> .bgred-100),
.SHE_SIDAK td:has(> .bggreen-100) {
  padding-left: 2px !important;
  padding-right: 2px !important;
}

.SHE_SIDAK td:has(> .bgred-100) {
  @apply bg-red-100;
}

.SHE_SIDAK td:has(> .bggreen-100) {
  @apply bg-green-100;
}

.SHE_SIDAK .form_inputs > div.flex.section_group_container > div:nth-child(1) {
  width: 70%;
}

.SHE_SIDAK .form_inputs > div.flex.section_group_container > div:nth-child(2) {
  width: 30%;
}
</style>
