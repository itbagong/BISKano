<template>
  <data-list
    class="card"
    ref="listControl"
    title="MCU"
    grid-config="/she/mcutransaction/gridconfig"
    form-config="/she/mcutransaction/formconfig"
    grid-read="/she/mcutransaction/gets"
    form-read="/she/mcutransaction/get"
    grid-mode="grid"
    grid-delete="/she/mcutransaction/delete"
    form-keep-label
    form-insert="/she/mcutransaction/save"
    form-update="/she/mcutransaction/save"
    :init-app-mode="data.appMode"
    grid-hide-select
    @form-edit-data="openForm"
    @form-new-data="newData"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    stay-on-form-after-save
    @form-field-change="onFieldChange"
    :form-tabs-edit="['General', 'Assesment', 'Result', 'Follow Up']"
    :form-fields="['Dimension', 'IsActive', 'Name']"
    :grid-fields="['Status']"
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
          ref="refCategory"
          v-model="data.search.Category"
          label="Category"
          class="w-[200px]"
          use-list
          :items="['Candidate', 'Employee']"
          @change="refreshData"
        ></s-input>
        <s-input
          :key="data.search.Category"
          ref="refName"
          v-model="data.search.Name"
          lookup-key="_id"
          label="Name"
          class="w-[350px] max-w-[350px]"
          use-list
          :lookup-url="
            data.search.Category == 'Candidate'
              ? '/she/mcutransaction/find?Category=' + data.search.Category
              : '/tenant/employee/find'
          "
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        >
        </s-input>
        <!-- <s-input
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
        ></s-input> -->
        <s-input
          ref="refPosition"
          v-model="data.search.Position"
          lookup-key="_id"
          label="Position"
          class="w-[200px]"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=PTE`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refCustomer"
          v-model="data.search.Customer"
          lookup-key="_id"
          label="Customer"
          class="w-[400px]"
          use-list
          :lookup-url="`/tenant/customer/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refPurpose"
          v-model="data.search.Purpose"
          lookup-key="_id"
          label="Purpose"
          class="w-[200px]"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=Purpose`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refProvider"
          v-model="data.search.Provider"
          lookup-key="_id"
          label="Hospital/Provider"
          class="w-[400px]"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=MPR`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refMCUPackage"
          v-model="data.search.MCUPackage"
          lookup-key="_id"
          label="MCU Package"
          class="w-[200px]"
          use-list
          :lookup-url="`/she/mcumasterpackage/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
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
    <template #form_buttons_1="{}">
      <s-button
        label="Generate Referrals"
        tooltip="Generate Referrals"
        class="btn_success"
        @click="onGenerateRef"
        v-if="tabIndex > 1"
        :disabled="data.loadingPdf"
      />
    </template>
    <template #form_input_Dimension="{ item, mode }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :read-only="mode == 'view'"
      ></dimension-editor-vertical>
    </template>
    <template #form_tab_Assesment="{ item, mode }">
      <assesment v-model="item.AssessmentResult" />
    </template>
    <template #form_tab_Result="{ item, mode }">
      <result
        v-model="item.MCUResult"
        :mcu-paket="item.MCUPackage"
        :mcu-add-item="item.AdditionalItem"
        :jurnal-id="item._id"
        :gender="data.objGender[item.Gender]"
      />
    </template>
    <template #form_input_Name="{ item, mode }">
      <s-input
        v-if="item.Category == 'Employee'"
        class="w-full"
        label="Name"
        use-list
        lookup-url="/tenant/employee/find"
        lookup-key="_id"
        :lookup-labels="['_id', 'Name']"
        :lookup-searchs="['_id', 'Name']"
        v-model="item.Name"
        keep-label
        @change="getEmployee"
      />
      <s-input
        v-if="item.Category == 'Candidate'"
        class="w-full"
        label="Name"
        keep-label
        v-model="item.Name"
      />
    </template>
    <template #form_tab_Follow_Up="{ item }">
      <div class="McuFollowUp">
        <div class="flex justify-end">
          <s-button
            icon="plus"
            tooltip="Add new follow up"
            class="btn_primary"
            @click="data.modalFU.value = true"
          />
        </div>
        <s-modal
          :display="data.modalFU.value"
          hideButtons
          title="Generate MCU Follow Up"
          @beforeHide="resetModalFU"
        >
          <div class="grid grid-cols-1 gap-2 w-[400px]">
            <s-input
              class="w-full mb-2"
              label="Additional Item/Follow Up"
              use-list
              lookup-url="/she/mcutransaction/get-mcu-item-last-child"
              lookup-key="ID"
              :lookup-labels="['Description']"
              :lookup-searchs="['ID', 'Description']"
              v-model="data.modalFU.additionalItem"
              keep-label
              multiple
            />
            <s-input
              v-model="data.modalFU.specialist"
              label="Specialist"
              multi-row="2"
            />
            <div class="flex justify-end">
              <s-button
                label="Generate"
                tooltip="Generate"
                class="btn_primary"
                @click="onGenerateFU"
              />
            </div>
          </div>
        </s-modal>
      </div>
      <div v-for="(dt, idx) in item.FollowUp" :key="idx">
        <follow-up :jurnal-id="dt.ID" v-model="item.FollowUp[idx]" />
      </div>
    </template>
    <template #grid_Status="{ item }">
      <status-text :txt="item.Status" />
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

import { SInput, util, SButton, SGrid, SCard, DataList, SModal } from "suimjs";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import StatusText from "@/components/common/StatusText.vue";
import Assesment from "./widget/McuAssesment.vue";
import Result from "./widget/McuResult.vue";
import FollowUp from "./widget/McuFollowUp.vue";
import { layoutStore } from "@/stores/layout.js";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";

const axios = inject("axios");
const listControl = ref(null);
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
  if (data.search.Category !== null && data.search.Category !== "") {
    filters.push({
      Field: "Category",
      Op: "$eq",
      Value: data.search.Category,
    });
  }
  if (data.search.Name !== null && data.search.Name !== "") {
    filters.push({
      Field: "Name",
      Op: "$eq",
      Value: data.search.Name,
    });
  }
  if (data.search.Position !== null && data.search.Position !== "") {
    filters.push({
      Field: "Position",
      Op: "$eq",
      Value: data.search.Position,
    });
  }
  if (data.search.Provider !== null && data.search.Provider !== "") {
    filters.push({
      Field: "Provider",
      Op: "$eq",
      Value: data.search.Provider,
    });
  }
  if (data.search.Customer !== null && data.search.Customer !== "") {
    filters.push({
      Field: "Customer",
      Op: "$eq",
      Value: data.search.Customer,
    });
  }
  if (data.search.Purpose !== null && data.search.Purpose !== "") {
    filters.push({
      Field: "Purpose",
      Op: "$eq",
      Value: data.search.Purpose,
    });
  }
  if (data.search.MCUPackage !== null && data.search.MCUPackage !== "") {
    filters.push({
      Field: "MCUPackage",
      Op: "$eq",
      Value: data.search.MCUPackage,
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
  record: {},
  modalFU: {
    value: false,
    additionalItem: [],
    specialist: "",
  },
  loadingPdf: false,
  objGender: {},
  search: {
    DateFrom: null,
    DateTo: null,
    Category: "",
    Name: "",
    Position: "",
    Customer: "",
    Provider: "",
    Purpose: "",
    MCUPackage: "",
    Site: "",
  },
});

function newData(r) {
  r.Date = new Date();
  data.record = r;
  data.record.AdditionalItem = [];
}

function openForm(r) {
  data.record = r;
  if (r.Category == "Employee") {
    getEmployee();
  }
}

function getEmployee() {
  util.nextTickN(3, () => {
    const url = "/bagong/employee/get";
    axios.post(url, [data.record.Name]).then(
      (r) => {
        data.record.Gender = r.data.Detail.Gender;
        data.record.Age = parseInt(r.data.Detail.Age);
        data.record.Position = r.data.Detail.Position;
        setAttrForm(["Age", "Position", "Gender"], "readOnly", true);
      },
      (e) => {
        util.showError(e);
      }
    );
  });
}

function onFieldChange(name, v1) {
  if (name == "Category") {
    resetEmployee();
    data.record.Name = "";
    setAttrForm(["Age", "Position", "Gender"], "readOnly", false);
  }
}

function setAttrForm(field, attr, val) {
  for (let i in field) {
    let o = field[i];
    listControl.value.setFormFieldAttr(o, "readOnly", val);
  }
}

function resetEmployee() {
  data.record.Name = "";
  data.record.Age = 0;
  data.record.Position = "";
  data.record.Gender = "";
}

function resetModalFU() {
  data.modalFU.value = false;
  data.modalFU.additionalItem = [];
  data.modalFU.specialist = "";
}

function onGenerateFU() {
  const url = "/she/mcutransaction/get-mcu-item-last-child";
  axios
    .post(url, {
      Where: {
        Op: "$in",
        Field: "ID",
        Value: data.modalFU.additionalItem,
      },
    })
    .then(
      (r) => {
        let obj = {
          ID: util.uuid(),
          DetailPemeriksaan: r.data,
          Notes: "",
          DoctorParamedic: data.modalFU.specialist,
          AdditionalItem: data.modalFU.additionalItem,
        };
        data.record.FollowUp.push(helper.cloneObject(obj));
      },
      (e) => {
        util.showError(e);
      }
    );
  resetModalFU();
}

const tabIndex = computed({
  get() {
    let idx = listControl.value.getFormCurrentTab();
    return idx ?? 0;
  },
});

function onGenerateRef() {
  const url = "/she/mcutransaction/save-preview";
  let p = {
    MCUTransaction: data.record,
    Site: helper.findDimension(data.record.Dimension, "Site"),
    IsResult: tabIndex.value == 2,
  };

  axios.post(url, p).then(
    (r) => {
      downloadPdf();
    },
    (e) => {
      util.showError(e);
    }
  );
}

function downloadPdf() {
  let type = tabIndex.value == 2 ? "MCU" : "MCU_FOLLOWUP";
  data.loadingPdf = true;
  const url = `/fico/postingprofile/preview-download-as-pdf?SourceType=${type}&SourceJournalID=${data.record._id}`;
  axios
    .get(url, {
      responseType: "blob",
    })
    .then(
      (r) => {
        const downloadUrl = window.URL.createObjectURL(r.data);
        window.open(downloadUrl, "__blank");
        window.URL.revokeObjectURL(url);
        data.loadingPdf = false;
      },
      (e) => {
        data.loadingPdf = false;
        util.showError(e);
      }
    );
}

function getListGender() {
  const url = "/tenant/masterdata/find?MasterDataTypeID=GEME";
  axios.post(url).then(
    (r) => {
      for (let i in r.data) {
        let o = r.data[i];
        data.objGender[o._id] = o.Name;
      }
    },
    (e) => {
      util.showError(e);
    }
  );
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
onMounted(() => {
  getListGender();
});
</script>
<style>
.McuFollowUp .modal_fullbg + div {
  top: -50px;
}
</style>
