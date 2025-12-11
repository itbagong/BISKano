<template>
  <data-list
    class="card rsca_transaction"
    ref="listControl"
    :title="data.titleForm"
    grid-config="/she/p3k/gridconfig"
    form-config="/she/p3k/formconfig"
    grid-read="/she/p3k/gets"
    form-read="/she/p3k/get"
    grid-mode="grid"
    grid-delete="/she/p3k/delete"
    form-keep-label
    form-insert="/she/p3k/save"
    form-update="/she/p3k/save"
    :init-app-mode="data.appMode"
    :form-fields="['PatientType', 'PatientName', 'Dimension', 'Medicines']"
    :grid-fields="['Dimension']"
    :form-tabs-edit="['General', 'P3K Detail']"
    grid-hide-select
    stay-on-form-after-save
    @form-edit-data="editRecord"
    @form-new-data="newData"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    @alterGridConfig="alterGridConfig"
    @controlModeChanged="onControlModeChanged"
    @form-field-change="onFormFieldChange"
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
          ref="refPatientName"
          v-model="data.search.Name"
          lookup-key="_id"
          label="Patient Name"
          class="w-[400px]"
          use-list
          :lookup-url="`/tenant/employee/find`"
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
    <template #grid_Dimension="{ item }">
      <DimensionText :dimension="item.Dimension" />
    </template>
    <template #form_tab_P3K_Detail="{ item }">
      <div class="relative">
        <s-form
          ref="detailFormCtl"
          v-model="data.record"
          :config="data.formCfg"
          keep-label
          only-icon-top
          :buttons-on-top="false"
          hide-cancel
          auto-focus
        >
          <template #input_Medicines="{ item, mode }">
            <div class="relative">
              <s-grid
                class="medicines_lines"
                ref="gridLinesMedicines"
                :config="data.cfgGrid"
                hide-search
                hide-sort
                hide-refresh-button
                hide-edit
                hide-select
                hide-paging
                editor
                auto-commit-line
                no-confirm-delete
                @new-data="newMedicines"
              >
                <template #item_Name="{ item, idx }">
                  <s-input
                    ref="refName"
                    v-model="item.Name"
                    class="w-full"
                    use-list
                    allowAdd
                    lookup-url="/tenant/item/find?ItemGroupID=GRP0016"
                    lookup-key="_id"
                    :lookup-labels="['Name', 'CostUnit']"
                    :lookup-searchs="['_id', 'Name']"
                    @change="
                      (...args) => handleChangedMedicineName(item, ...args)
                    "
                  >
                    <template #selected-option="{ option }">
                      {{ option?.item?.Name }}
                    </template>
                  </s-input>
                </template>

                <template #item_button_delete="{ item, idx }">
                  <a @click="deleteMedicines(item)" class="delete_action">
                    <mdicon
                      name="delete"
                      width="16"
                      alt="delete"
                      class="cursor-pointer hover:text-primary"
                    />
                  </a>
                </template>
                <template #footer_1="{ items }">
                  <div
                    v-if="data.cfgGrid.fields.length > 0"
                    class="gap-2 p-4 ml-[60%]"
                  >
                    <div class="flex flex-row justify-between">
                      <span class="text-base font-semibold"
                        >Concultancy Price</span
                      >
                      <span class="text-base font-semibold">
                        <s-input
                          class="w-full"
                          kind="number"
                          hide-label
                          v-model="item.ConculPrice"
                        />
                      </span>
                    </div>
                    <div class="flex flex-row justify-between">
                      <span class="text-base font-semibold"
                        >Treatment/Medicine Prices</span
                      >
                      <span class="text-base font-semibold">
                        {{ helper.formatNumberWithDot(medicinePrices) }}
                      </span>
                    </div>
                    <div class="flex flex-row justify-between">
                      <span class="text-base font-semibold">Total Prices</span>
                      <span class="text-base font-semibold">
                        {{ helper.formatNumberWithDot(totalPrice) }}
                      </span>
                    </div>
                  </div>
                </template>
              </s-grid>
            </div>
          </template>
          <template #input_AlcoholTest="{ item }">
            <label class="input_label">
              <div>Breath Alcohol Test</div>
            </label>
            <s-toggle
              v-model="item.AlcoholTest"
              class="w-[120px] mt-0.5"
              yes-label="Positive"
              no-label="Negatif"
            />
          </template>
          <template #input_Diastolic="{ item }">
            <div class="grid grid-cols-2 gap-2 mt-2">
              <div>
                <s-input
                  kind="number"
                  ref="refDiastolic"
                  v-model="item.Diastolic"
                  label="Diastolic (mmHg)"
                  class="w-full"
                ></s-input>
              </div>
              <div>
                <s-input
                  kind="number"
                  ref="refDiastolic"
                  v-model="item.Systolic"
                  label="Systolic (mmHg)"
                  class="w-full"
                ></s-input>
              </div>
            </div>
          </template>
          <template #input_DrugTest="{ item }">
            <div class="section grow">
              <div class="title section_title">Drug Test (6 indicator)</div>
            </div>
            <div class="grid grid-cols-4 gap-2 mt-2">
              <div>
                <label class="input_label">
                  <div>Amphetamine</div>
                </label>
              </div>
              <div>
                <s-toggle
                  v-model="item.Amphetamine"
                  class="w-[120px] mt-0.5"
                  yes-label="Positive"
                  no-label="Negatif"
                />
              </div>
              <div>
                <label class="input_label">
                  <div>Morphin</div>
                </label>
              </div>
              <div>
                <s-toggle
                  v-model="item.Morphin"
                  class="w-[120px] mt-0.5"
                  yes-label="Positive"
                  no-label="Negatif"
                />
              </div>
              <div>
                <label class="input_label">
                  <div>Menthapet</div>
                </label>
              </div>
              <div>
                <s-toggle
                  v-model="item.Menthapet"
                  class="w-[120px] mt-0.5"
                  yes-label="Positive"
                  no-label="Negatif"
                />
              </div>
              <div>
                <label class="input_label">
                  <div>Cocain</div>
                </label>
              </div>
              <div>
                <s-toggle
                  v-model="item.Cocain"
                  class="w-[120px] mt-0.5"
                  yes-label="Positive"
                  no-label="Negatif"
                />
              </div>
              <div>
                <label class="input_label">
                  <div>Marijuana</div>
                </label>
              </div>
              <div>
                <s-toggle
                  v-model="item.Marijuana"
                  class="w-[120px] mt-0.5"
                  yes-label="Positive"
                  no-label="Negatif"
                />
              </div>
              <div>
                <label class="input_label">
                  <div>Benzodiaze</div>
                </label>
              </div>
              <div>
                <s-toggle
                  v-model="item.Benzodiaze"
                  class="w-[120px] mt-0.5"
                  yes-label="Positive"
                  no-label="Negatif"
                />
              </div>
            </div>
          </template>
        </s-form>
      </div>
    </template>
    <template #form_input_PatientType="{ item }">
      <label class="input_label">
        <div>Patient type</div>
      </label>
      <s-toggle
        v-model="item.PatientType"
        class="w-[120px] mt-0.5"
        yes-label="Internal"
        no-label="External"
        @change="
          onFormFieldChange('PatientType', item.PatientType, '', '', item)
        "
      />
    </template>
    <template #form_input_PatientName="{ item, mode }">
      <s-input
        v-if="!item.PatientType"
        ref="refPatientName"
        v-model="item.PatientName"
        label="Patient name"
        class="w-full"
      ></s-input>
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
import { reactive, ref, inject, onMounted, computed } from "vue";
import {
  DataList,
  SInput,
  SForm,
  SGrid,
  loadGridConfig,
  loadFormConfig,
  util,
} from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import SInputSkuItem from "../scm/widget/SInputSkuItem.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import SToggle from "@/components/common/SButtonToggle.vue";
import helper from "@/scripts/helper.js";
import moment from "moment";

layoutStore().name = "tenant";
const listControl = ref(null);
const gridLinesMedicines = ref(null);
const axios = inject("axios");

const medicinePrices = computed({
  get() {
    const total = data.record.Medicines.reduce((a, b) => {
      let val = 0;
      if (b.Price) {
        val = b.Price;
      }
      return a + val;
    }, 0);
    data.record.MedPRice = total;
    return total;
  },
});

const totalPrice = computed({
  get() {
    let TotalPrice = medicinePrices.value + data.record.ConculPrice;
    data.record.TotalPrice = TotalPrice;
    return TotalPrice;
  },
});
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
      Value: helper.formatFilterDate(data.search.DateFrom),
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
      Value: helper.formatFilterDate(data.search.DateTo, true),
    });
  }
  if (data.search.Name !== null && data.search.Name !== "") {
    filters.push({
      Field: "PatientName",
      Op: "$eq",
      Value: data.search.Name,
    });
  }
  if (data.search.Purpose !== null && data.search.Purpose !== "") {
    filters.push({
      Field: "Purpose",
      Op: "$eq",
      Value: data.search.Purpose,
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
  titleForm: "P3K",
  record: {},
  cfgGrid: {},
  formCfg: {},
  search: {
    DateFrom: null,
    DateTo: null,
    Name: "",
    Purpose: "",
    Site: "",
  },
});

function newData(r) {
  r.Medicines = [];
  r.PatientType = true;
  r.ConculPrice = 0;
  r.Date = new Date();
  r.TimeIn = moment().format("HH:mm");
  r.TimeOut = moment().format("HH:mm");
  data.record = r;
  data.titleForm = "Create New P3K";
}

function editRecord(r) {
  let TimeIn = new Date(r.TimeIn);
  let hoursIn = TimeIn.getUTCHours().toString().padStart(2, "0");
  let minutesIn = TimeIn.getUTCHours().toString().padStart(2, "0");

  let TimeOut = new Date(r.TimeOut);
  let hoursOut = TimeOut.getUTCHours().toString().padStart(2, "0");
  let minutesOut = TimeOut.getUTCMinutes().toString().padStart(2, "0");
  r.TimeIn = `${hoursIn}:${minutesIn}`;
  r.TimeOut = `${hoursOut}:${minutesOut}`;
  data.record = r;
  data.titleForm = `Edit P3K | ${r._id}`;
  util.nextTickN(2, () => {
    gridLinesMedicines.value.setRecords(r.Medicines);
  });
}

function openForm(r) {
  data.record = r;
}

function handleChangedMedicineName(record, field, v1, v2, old) {
  const arr = v2.split(" | ");
  const price = arr.length == 2 ? arr[1] : 0;
  record.Price = parseInt(price);
}
function onPreSave(record) {
  const d = record.Date;
  let TimeIn = record.TimeIn;
  let TimeOut = record.TimeOut;
  let dateIn = new Date(d);
  let dateOut = new Date(d);

  const [hoursIn, minutesIn] = TimeIn.split(":").map(Number);
  const [hoursOut, minutesOut] = TimeOut.split(":").map(Number);
  dateIn.setUTCHours(hoursIn, minutesIn, 0, 0);
  dateOut.setUTCHours(hoursOut, minutesOut, 0, 0);

  record.TimeIn = dateIn;
  record.TimeOut = dateOut;
  console.log(record);
}
function newMedicines() {
  let r = {};
  const noLine = data.record.Medicines.length + 1;
  r._id = util.uuid();
  r.MedNo = noLine;
  r.Name = "";
  r.Qty = 0;
  r.Price = 0;
  r.Remark = "";
  data.record.Medicines.push(r);
  updateGridMedicines();
}
function deleteMedicines(r) {
  data.record.Medicines = data.record.Medicines.filter(
    (obj) => obj._id !== r._id
  );
  updateGridMedicines();
}

function updateGridMedicines() {
  data.record.Medicines.map((obj, idx) => {
    obj.MedNo = parseInt(idx) + 1;
    return obj;
  });
  gridLinesMedicines.value.setRecords(data.record.Medicines);
}

function onFormFieldChange(field, v1, v2, old, record) {
  switch (field) {
    case "PatientType":
      record.PatientName = "";
      record.Age = "";
      record.Gender = "";
      break;
    case "PatientName":
      axios.post("/bagong/employee/get", [v1]).then(
        (r) => {
          let emp = r.data;
          record.Age = emp.Detail.Age;
          record.Gender = emp.Detail.Gender;
        },
        (e) => util.showError(e)
      );
      break;
    default:
      break;
  }
}

function alterGridConfig(cfg) {
  cfg.sortable = ["Created", "_id"];
  cfg.setting.idField = "Created";
  cfg.setting.sortable = ["Created", "_id"];
}
function loadGridMedicines() {
  let url = `/she/p3k/medicines/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGrid = r;
    },
    (e) => {}
  );
}

function loadFromDetail() {
  let url = `she/p3k/detail/formconfig`;
  loadFormConfig(axios, url).then(
    (r) => {
      data.formCfg = r;
    },
    (e) => {}
  );
}

function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = `P3K`;
  }
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
onMounted(() => {
  loadGridMedicines();
  loadFromDetail();
});
</script>
<style lang="css" scoped>
.switch-field {
  display: flex;
  margin-bottom: 0px;
  overflow: hidden;
}

.switch-field input {
  position: absolute !important;
  clip: rect(0, 0, 0, 0);
  height: 1px;
  width: 1px;
  border: 0;
  overflow: hidden;
}

.switch-field label {
  background-color: #e4e4e4;
  color: rgba(0, 0, 0, 0.6);
  font-size: 14px;
  line-height: 1;
  text-align: center;
  padding: 8px 16px;
  margin-right: -1px;
  border: 1px solid rgba(0, 0, 0, 0.2);
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.3), 0 1px rgba(255, 255, 255, 0.1);
  transition: all 0.1s ease-in-out;
}

.switch-field label:hover {
  cursor: pointer;
}

.switch-field input:checked + label {
  background-color: #fd6e76;
  box-shadow: none;
  color: white;
}

.switch-field label:first-of-type {
  border-radius: 4px 0 0 4px;
}

.switch-field label:last-of-type {
  border-radius: 0 4px 4px 0;
}
</style>
