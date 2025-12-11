<template>
  <div class="w-full">
    <s-card
      title="Training Report"
      class="w-full bg-white suim_datalist"
      hide-footer
    >
      <template #title_controls>
        <div class="flex gap-3">
          <s-input
            label="Search"
            kind="text"
            caption="enter search keyword"
            v-model="data.filter.search"
            @keyup.enter="refreshData"
            class="w-full min-w-[200px] text-black"
          />
          <s-input
            label="Employee"
            v-model="data.filter.employeeID"
            use-list
            :lookup-url="`/tenant/employee/find`"
            lookup-key="_id"
            :lookup-labels="['_id', 'Name']"
            :lookup-searchs="['_id', 'Name']"
            class="w-full min-w-[400px]"
            @change="
              (field, v1, v2, old, ctlRef) => {
                data.filter.employeeID = v1
                refreshData()
              }
            "
          />
          <s-button
            icon="refresh"
            class="btn_primary refresh_btn justify-end"
            tooltip="refresh"
            @click="refreshData"
            v-if="!hideRefreshButton"
          />
        </div>
      </template>
      <!-- tabs -->
      <div class="flex tab_container grow border-b border-gray-300 pb-2 mb-4">
        <button
          :class="{
            tab_selected: data.activeTab === 'internal',
            tab: data.activeTab !== 'internal',
          }"
          @click="handleChangeTab('internal')"
        >
          Internal
        </button>
        <button
          :class="{
            tab_selected: data.activeTab === 'external',
            tab: data.activeTab !== 'external',
          }"
          @click="handleChangeTab('external')"
        >
          External
        </button>
      </div>
      <!-- internal -->
      <div v-show="data.activeTab === 'internal'">
        <s-grid
          v-show="data.mode === 'grid'"
          ref="listControl"
          class=""
          hide-select
          hide-new-button
          hide-delete-button
          hide-edit
          hide-control
          :config="data.cfg"
          :grid-fields="['Date', 'TrainingDateFrom', 'TrainingDateTo']"
          @grid-refreshed="refresh"
        >
          <template #item_Date="{ item }">
            {{ moment(item.Date).format("YYYY-MM-DD").toString() }}
          </template>
          <template #item_TrainingDateFrom="{ item }">
            {{ moment(item.TrainingDateFrom).format("YYYY-MM-DD").toString() }}
          </template>
          <template #item_TrainingDateTo="{ item }">
            {{ moment(item.TrainingDateTo).format("YYYY-MM-DD").toString() }}
          </template>
          <template #item_buttons_1="{ item }">
            <s-button
              class="bg-transparent hover:text-green-500 m-2"
              icon="eye-outline"
              @click="onPreview(item)"
            ></s-button>
          </template>
        </s-grid>
        <div class="" v-show="data.mode === 'detail'">
          <div class="flex justify-end gap-1 mb-3">
            <s-button
              class="btn_warning back_btn"
              label="Back"
              icon="rewind"
              @click="
                () => (data.mode = 'grid')
              "
            />
          </div>
          <s-grid
            ref="gridDetail"
            class="w-full"
            hide-select
            hide-new-button
            hide-delete-button
            hide-edit
            hide-action
            hide-refresh-button
            hide-search
            hide-sort
            hide-control
            :config="data.cfgDetail"
          >
          </s-grid>
        </div>
      </div>
      <!-- external -->
      <div v-show="data.activeTab === 'external'" class="py-5">
        <div v-show="data.externalMode == 'list'">
          <div class="flex justify-end gap-2">
            <s-button
              icon="plus"
              class="btn_primary"
              label="Add"
              @click="externalNewRecord"
            />
          </div>
          <div class="grid grid-cols-4 gap-5 mt-5">
            <div 
              v-for="(asset, idx) in data.recordsExternal"
              :key="idx"
              class="w-full"
            >
              <div class="w-full h-[220px] flex justify-center items-center rounded-md border border-gray-400">
                <img :src="`/v1/asset/view?id=${asset._id}`" class="max-w-[300px] max-h-[200px]">
              </div>
              
              <div class="grid justify-center text-center w-full text-lg mt-3">
                <div>{{ (idx+1) + '. ' + asset.Title }}</div>
                <div class="text-gray-500">Exp. Date: {{ moment(asset.ExpiredDate).format("YYYY-MM-DD").toString() }}</div>
              </div>
            </div>
          </div>
        </div>
        <div v-show="data.externalMode == 'form'">
          <div class="flex justify-end gap-2">
            <s-button
              icon="content-save"
              class="btn_primary"
              label="Save"
              @click="saveAsset"
            />
            <s-button
              class="btn_warning back_btn"
              label="Back"
              icon="rewind"
              @click="handleBack"
            />
          </div>
          <div class="flex gap-5">
            <div class="box-uploader-style" @click="triggerFileInput">
              <mdicon name="upload" size="25" class="mx-auto" />
              <div class="text-lg">Upload File</div>
              <div v-show="data.asset.FileName !== undefined" class="">
                <span class="text-md">{{ data.asset.FileName }}</span><br />
                <span class="text-md">{{ data.asset.ContentType }}</span>
              </div>
              <input type="file" ref="fileInput" @change="handleFileUpload" hidden />
            </div>
            <div class="grid w-1/3">
              <s-input
                label="Title"
                kind="text"
                caption="Input title"
                v-model="data.asset.Title"
                class="w-full"
                keep-label
              />
              <s-input
                label="Expired Date"
                kind="date"
                v-model="data.asset.ExpiredDate"
                class="w-full"
              />
            </div>
          </div>
        </div>
      </div>
    </s-card>
  </div>
</template>

<script setup>
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
import { util, SCard, SGrid, SButton, SInput } from "suimjs";
import { reactive, ref, inject, onMounted } from "vue";
import moment from "moment";

layoutStore().name = "tenant";

const axios = inject("axios");
const auth = authStore();

const listControl = ref(null);
const gridDetail = ref(null);
const fileInput = ref(null);

const data = reactive({
  records: [],
  recordsExternal: [],
  recordEmployee: {},
  cfg: {},
  cfgDetail: {},
  mode: 'grid', // grid | detail
  externalMode: "list", // list | form
  activeTab: 'internal', // internal | external
  asset: {},
  filter: {
    search: "",
    employeeID: "",
  }
})

function refresh() {
  getDataReports(data.recordEmployee._id)
}

function handleChangeTab(params) {
  data.activeTab = params
}

function refreshData() {
  data.activeTab = 'internal'
  data.mode = 'grid'
  data.externalMode = 'list'
  data.asset = {}
  if(data.filter.employeeID == ""){
    return util.showError('Select filter employee!')
  }
  getDataReports(data.filter.employeeID)
  getDataAssets()
}

function getDataReports(employeeID) {
  listControl.value.setLoading(true);
  const payload = {
    EmployeeID: employeeID,
    TrainingType: "internal",
    Search: data.filter.search,
  }
  axios.post("/hcm/tdc/get-reports", payload).then(
      (res) => {
        const dt = res.data
        data.records = dt || []
        listControl.value.setRecords(dt);
        listControl.value.setLoading(false);
      },
      (err) => {
        listControl.value.setLoading(false);
        util.showError(err);
      }
  );
}

// async function getDataEmployee() {
//   const payload = {
//     Take: 1, 
//     Where: {Field: "Email", Op: "$eq", Value: auth.appData.Email}
//   }

//   await axios.post("/tenant/employee/find", payload).then(
//     (r) => {
//       const dt = r.data.length > 0 ? r.data[0] : {}
//       const employeeID = dt._id || "";
//       data.recordEmployee = dt;
//       getDataReports(employeeID);
//     },
//     (e) => {
//       util.showError(e);
//     }
//   );
// }

function genGrid() {
  data.cfg = {
    fields: [
      {
        field: "Date",
        kind: "text",
        label: "Date",
        labelField: "",
        readType: "show",
        input: {
          field: "Date",
          kind: "text",
          label: "Date",
          lookupUrl: "",
          placeHolder: "Date",
        },
      },
      {
        field: "TrainingName",
        kind: "text",
        label: "Training Name",
        labelField: "",
        readType: "show",
        input: {
          field: "TrainingName",
          kind: "text",
          label: "Training Name",
          lookupUrl: "",
          placeHolder: "Training Name",
        },
      },
      {
        field: "AssessmentType",
        kind: "text",
        label: "Assessment Type",
        labelField: "",
        readType: "show",
        input: {
          field: "AssessmentType",
          kind: "text",
          label: "Assessment Type",
          lookupUrl: "",
          placeHolder: "Assessment Type",
        },
      },
      {
        field: "TrainingDateFrom",
        kind: "text",
        label: "Training Date From",
        labelField: "",
        readType: "show",
        input: {
          field: "TrainingDateFrom",
          kind: "text",
          label: "Training Date From",
          lookupUrl: "",
          placeHolder: "Training Date From",
        },
      },
      {
        field: "TrainingDateTo",
        kind: "text",
        label: "Training Date To",
        labelField: "",
        readType: "show",
        input: {
          field: "TrainingDateTo",
          kind: "text",
          label: "Training Date To",
          lookupUrl: "",
          placeHolder: "Training Date To",
        },
      },
    ],
    setting: {
      idField: "",
      keywordFields: ["_id", "Date"],
      sortable: ["_id", "TrainingName"],
    },
  };

  data.cfgDetail = {
    fields: [
      {
        field: "TestType",
        kind: "text",
        label: "Test Type",
        labelField: "",
        readType: "show",
        input: {
          field: "TestType",
          kind: "text",
          label: "Test Type",
          lookupUrl: "",
          placeHolder: "Test Type",
        },
      },
      {
        field: "Score",
        kind: "text",
        label: "Score",
        labelField: "",
        readType: "show",
        input: {
          field: "Score",
          kind: "text",
          label: "Score",
          lookupUrl: "",
          placeHolder: "Score",
        },
      },
    ],
    setting: {
      idField: "",
      keywordFields: ["_id"],
      sortable: ["_id"],
    },
  };
}

function onPreview(dt) {
  data.mode = 'detail'
  let urlDetail =
    dt.AssessmentType === "Assessment Staff"
      ? "/hcm/tdc/assesment-staff"
      : dt.AssessmentType === "Assessment Driver"
      ? "/hcm/tdc/assesment-driver"
      : dt.AssessmentType === "Assessment Mechanic"
      ? "/hcm/tdc/assesment-mechanic"
      : "";

  if (urlDetail !== "") {
    gridDetail.value.setLoading(true);
    const paramDetail = {
      TrainingCenterID: dt.ID,
      Take: -1,
      Skip: 0,
    };

    axios.post(urlDetail, paramDetail).then(
      (res) => {
        const resDetail = res.data.data;
        const find = resDetail.find(e => {return e.EmployeeID == data.filter.employeeID})
        // console.log("find: ", find)
        let finalData = []
        if (dt.AssessmentType === "Assessment Staff") {
          finalData.push(
            {
              TestType: "Written Test",
              Score: find.WrittenTest
            },
            {
              TestType: "Practice Test",
              Score: find.PracticeTest
            },
          )
        } else if (dt.AssessmentType === "Assessment Driver") {
          finalData.push(
            {
              TestType: "Written Test",
              Score: find.WrittenTest
            },
            {
              TestType: "Practice Test",
              Score: find.PracticeTestScore
            },
            {
              TestType: "Practice Test (Duration)",
              Score: find.PracticeTestDuration
            },
          )
        } else if (dt.AssessmentType === "Assessment Mechanic") {
          finalData.push(
            {
              TestType: "Written Test",
              Score: find.WrittenTest
            },
            {
              TestType: "Practice Test",
              Score: find.PracticeTestScore
            },
          )
        }
        gridDetail.value.setRecords(finalData);
        gridDetail.value.setLoading(false);
      },
      (err) => {
        gridDetail.value.setLoading(false);
        util.showError(err);
      }
    );
  }
}


// External
async function getDataAssets() {
  if(data.filter.employeeID == ""){
    return util.showError('Select filter employee!')
  }
  try {
    let param = {
      JournalType: "TDC Training Report",
      JournalID: data.filter.employeeID,
      Tags: [],
    };

    const resp = await axios.post("/asset/read-by-journal", param);

    data.recordsExternal = resp.data.map((resp) => ({
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
      ExpiredDate: resp.Data.ExpiredDate,
      Title: resp.Data.Title
    }));
  } catch (error) {
    util.showError(error);
  }
}

function handleBack() {
  data.asset = {}
  data.externalMode = 'list'
}

function triggerFileInput() {
  fileInput.value.click();
}

function handleFileUpload(event) {
  const file = event.target.files[0];
  const reader = new FileReader();
  reader.readAsDataURL(file);

  reader.onloadend = () => {
    data.asset["Content"] = reader.result.split(",")[1];
    data.asset["OriginalFileName"] = file.name;
    data.asset["FileName"] = file.name;
    data.asset["ContentType"] = file.type;
  };
}

function externalNewRecord() {
  if(data.filter.employeeID == ""){
    return util.showError('Select filter employee!')
  }

  const r = {};
  r.OriginalFileName = "";
  r.ContentType = "";
  r.Descriptions = "";
  r.UploadDate = new Date();
  r.Tags = [];
  // if (props.singleSave) {
  //   r.isChange = true;
  // }
  r.SameJournal = true;
  r.Kind = "TDC Training Report";
  r.RefID = data.filter.employeeID;

  data.asset = r;
  data.externalMode = 'form';
}

async function saveAsset() {
  const record = data.asset
  if (!record.Title) {
    return util.showError('Title is required')
  }
  if (!record.ExpiredDate) {
    return util.showError('Expired Date is required')
  }
  const DataRecord = { ...record };
  delete DataRecord.OriginalFileName;
  delete DataRecord.ContentType;
  delete DataRecord.ReadOnly;
  delete DataRecord.Content;

  const param = {
    Content: record.Content,
    Asset: {
      _id: record._id,
      OriginalFileName: record.OriginalFileName,
      ContentType: record.ContentType,
      Kind: record.Kind,
      RefID: record.RefID,
      Data: DataRecord,
      URI: record.URI,
    },
  };
  console.log(param)
  
  try {
    await axios.post(`/asset/write-asset`, param);
    handleBack();
    getDataAssets();
  } catch (error) {
    util.showError(error);
  }
}

onMounted(() => {
  genGrid();
  // getDataEmployee();
})
</script>

<style scoped>
/* tab */
.tab_container {
  font-weight: 600;
  align-items: center;
  margin-bottom: 0.5rem;
}
.tab {
  text-align: center;
  padding: 0.5rem;
  --tw-border-opacity: 1;
  border-color: rgb(203 213 225 / var(--tw-border-opacity));
  border-bottom-width: 5px;
  cursor: pointer;
  min-width: 50px;
  padding-left: 20px !important;
  padding-right: 20px !important;
}
.tab:hover {
  --tw-text-opacity: 1;
  color: rgb(253 125 133 / var(--tw-text-opacity));
}
.tab_selected {
  text-align: center;
  padding: 0.5rem;
  --tw-border-opacity: 1;
  border-color: rgb(253 110 118 / var(--tw-border-opacity));
  border-bottom-width: 5px;
  cursor: pointer;
  min-width: 50px;
  padding-left: 20px !important;
  padding-right: 20px !important;
  --tw-text-opacity: 1;
  color: rgb(253 110 118 / var(--tw-text-opacity));
}

.tab-active {
  --tw-text-opacity: 1;
  color: rgb(253 110 118 / var(--tw-text-opacity));
}

/* box upload */
.box-uploader-style {
  @apply border-2 border-dashed border-zinc-200 w-12 h-12 mr-2 rounded-md grid place-content-center relative cursor-pointer;
  width: 200px;
  height: 200px;
  text-align: center;
}
.box-uploader-style:hover {
  @apply border-zinc-500 text-zinc-500;
}
</style>