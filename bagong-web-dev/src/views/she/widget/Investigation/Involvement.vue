<template>
  <div>
    <s-form
      ref="detailInvolvementFormCtl"
      v-model="data.record"
      :config="data.formCfg"
      keep-label
      only-icon-top
      :buttons-on-top="false"
      hide-cancel
      auto-focus
    >
      <template #input_Person="{ item, idx }">
        <s-grid
          class="PersonLine"
          ref="personLine"
          :config="data.cfgGridPersonLine"
          hide-search
          hide-sort
          hide-refresh-button
          hide-edit
          hide-select
          hide-paging
          editor
          auto-commit-line
          no-confirm-delete
          @new-data="newPerson"
          @row-field-changed="handleChanged"
        >
          <template #item_Employee="{ item, idx }">
            <s-input
              v-if="item.Role == 'ROLE-004'"
              ref="refEmployee"
              v-model="item.Employee"
              class="w-full"
            ></s-input>
          </template>
          <template #item_Company="{ item, idx }">
            <s-input
              v-if="item.Role == 'ROLE-004'"
              ref="refCompany"
              v-model="item.Company"
              class="w-full"
            ></s-input>
          </template>
          <template #item_Job="{ item, idx }">
            <s-input
              v-if="item.Role == 'ROLE-004'"
              ref="refJob"
              v-model="item.Job"
              class="w-full"
            ></s-input>
          </template>
          <template #item_Supervisor="{ item, idx }">
            <s-input
              v-if="item.Role == 'ROLE-004'"
              ref="refJob"
              v-model="item.Supervisor"
              class="w-full"
            ></s-input>
          </template>
          <template #item_Injured="{ item, idx }">
            <s-button
              icon="format-list-bulleted"
              class="btn_primary"
              label="Detail"
              no-tooltip
              @click="injuredPerson(item)"
            />
          </template>
          <template #item_button_delete="{ item, idx }">
            <a @click="deletePerson(item)" class="delete_action">
              <mdicon
                name="delete"
                width="16"
                alt="delete"
                class="cursor-pointer hover:text-primary"
              />
            </a>
          </template>
        </s-grid>
      </template>
      <template #input_MedicalTreatment="{ item, idx }">
        <s-grid
          class="medical-treatment-line"
          ref="medicalTreatmentLine"
          :config="data.cfgGridMedicalTreatmentLine"
          hide-search
          hide-sort
          hide-refresh-button
          hide-edit
          hide-select
          hide-paging
          editor
          auto-commit-line
          no-confirm-delete
          @new-data="newTreatment"
        >
          <template #item_button_delete="{ item, idx }">
            <a @click="deleteTreatment(item)" class="delete_action">
              <mdicon
                name="delete"
                width="16"
                alt="delete"
                class="cursor-pointer hover:text-primary"
              />
            </a>
          </template>
        </s-grid>
      </template>
      <template #input_Asset="{ item, idx }">
        <s-grid
          class="asset-line"
          ref="AssetLine"
          :config="data.cfgGridAssetLine"
          hide-search
          hide-sort
          hide-refresh-button
          hide-edit
          hide-select
          hide-paging
          editor
          auto-commit-line
          no-confirm-delete
          @new-data="newAsset"
          @row-field-changed="handleChanged"
        >
          <template #item_Type="{ item, idx }">
            <s-input
              read-only
              label=""
              v-model="item.Type"
              use-list
              lookup-url="/tenant/masterdata/find"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              :key="item.Type"
            />
          </template>
          <template #item_PartEquipment="{ item, idx }">
            <s-button
              icon="format-list-bulleted"
              class="btn_primary"
              label="Detail"
              no-tooltip
              @click="PartEquipmentAsset(item)"
            />
          </template>
          <template #item_button_delete="{ item, idx }">
            <a @click="deleteAsset(item)" class="delete_action">
              <mdicon
                name="delete"
                width="16"
                alt="delete"
                class="cursor-pointer hover:text-primary"
              />
            </a>
          </template>
        </s-grid>
      </template>
      <template #input_Environment="{ item, idx }">
        <s-grid
          class="asset-line"
          ref="EnvironmentLine"
          :config="data.cfgGridEnvironmentLine"
          hide-search
          hide-sort
          hide-refresh-button
          hide-edit
          hide-select
          hide-paging
          editor
          auto-commit-line
          no-confirm-delete
          @new-data="newEnvironment"
        >
          <template #item_button_delete="{ item, idx }">
            <a @click="deleteEnvironment(item)" class="delete_action">
              <mdicon
                name="delete"
                width="16"
                alt="delete"
                class="cursor-pointer hover:text-primary"
              />
            </a>
          </template>
        </s-grid>
      </template>
    </s-form>
  </div>
</template>
<script setup>
import { onMounted, inject, reactive, ref, computed } from "vue";
import {
  loadGridConfig,
  loadFormConfig,
  util,
  SForm,
  SGrid,
  SButton,
  SInput,
  SModal,
} from "suimjs";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import helper from "@/scripts/helper.js";
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
  showInjured: null,
  showPartEquipmentAsset: null,
});

const listControl = ref(null);
const personLine = ref(null);
const medicalTreatmentLine = ref(null);
const AssetLine = ref(null);
const EnvironmentLine = ref(null);

const data = reactive({
  record: props.modelValue,
  isExternal: false,
  formCfg: {},
  cfgGridPersonLine: {},
  cfgGridMedicalTreatmentLine: {},
  cfgGridAssetLine: {},
  cfgGridEnvironmentLine: {},
});

function loadFromInvolvement() {
  let url = `/she/investigasi/involvement/formconfig`;
  loadFormConfig(axios, url).then(
    (r) => {
      data.formCfg = r;
      loadGridPersonline();
      loadGridMedicalTreatmentLine();
      loadGridAssetLine();
      loadGridEnvironmentLine();
    },
    (e) => {}
  );
}
function loadGridPersonline() {
  let url = `/she/investigasi/personline/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridPersonLine = r;
      updateGridLine(data.record.Person, "Person");
    },
    (e) => {}
  );
}

function loadGridMedicalTreatmentLine() {
  let url = `/she/investigasi/medicaltreatmentline/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridMedicalTreatmentLine = r;
      updateGridLine(data.record.MedicalTreatment, "MedicalTreatment");
    },
    (e) => {}
  );
}

function loadGridAssetLine() {
  let url = `/she/investigasi/assetline/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridAssetLine = r;
      updateGridLine(data.record.Asset, "Asset");
    },
    (e) => {}
  );
}

function loadGridEnvironmentLine() {
  let url = `/she/investigasi/environmentline/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridEnvironmentLine = r;
      updateGridLine(data.record.Environment, "Environment");
    },
    (e) => {}
  );
}

function newPerson() {
  let r = {};
  const noLine = data.record.Person.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.Role = "";
  r.Employee = "";
  r.Company = "";
  r.Job = "";
  r.Gender = "";
  r.Supervisor = "";
  r.WorkingPeriod = "";
  r.Age = "";
  r.Injured = [];
  r.EstimationCost = 0;

  data.record.Person.push(r);
  updateGridLine(data.record.Person, "Person");
}
function deletePerson(r) {
  data.record.Person = data.record.Person.filter((obj) => obj._id !== r._id);
  updateGridLine(data.record.Person, "Person");
}
function newTreatment() {
  let r = {};
  const noLine = data.record.MedicalTreatment.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.Doctor = "";
  r.Hospital = "";
  r.Treatment = "";
  r.EstimationCost = 0;
  r.Remark = "";
  data.record.MedicalTreatment.push(r);
  updateGridLine(data.record.MedicalTreatment, "MedicalTreatment");
}
function deleteTreatment(r) {
  data.record.MedicalTreatment = data.record.MedicalTreatment.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.MedicalTreatment, "MedicalTreatment");
}
function newAsset() {
  let r = {};
  const noLine = data.record.Asset.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.Asset = "";
  r.Unit = "";
  r.Type = "";
  r.Damage = "";
  r.PartEquipment = [];
  r.TotalCost = 0;
  r.Remark = "";
  data.record.Asset.push(r);
  updateGridLine(data.record.Asset, "Asset");
}
function deleteAsset(r) {
  data.record.Asset = data.record.Asset.filter((obj) => obj._id !== r._id);
  updateGridLine(data.record.Asset, "Asset");
}

function newEnvironment() {
  let r = {};
  const noLine = data.record.Environment.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.Description = "";
  r.Damage = "";
  r.Remark = "";
  data.record.Environment.push(r);
  updateGridLine(data.record.Environment, "Environment");
}
function deleteEnvironment(r) {
  data.record.Environment = data.record.Environment.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.Environment, "Environment");
}
function updateGridLine(record, type) {
  record.map((obj, idx) => {
    obj.LineNo = parseInt(idx) + 1;
    return obj;
  });
  if (type == "Person") {
    personLine.value.setRecords(record);
  } else if (type == "MedicalTreatment") {
    medicalTreatmentLine.value.setRecords(record);
  } else if (type == "Asset") {
    AssetLine.value.setRecords(record);
  } else if (type == "Environment") {
    EnvironmentLine.value.setRecords(record);
  }
}
function handleChanged(field, v1, v2, old, record) {
  console.log("====================", field, v1, v2, old, record);
  switch (field) {
    case "Employee":
      record.Age = "";
      record.Gender = "";
      record.Job = "";
      record.WorkingPeriod = "";
      record.Company = "";
      if (v1) {
        axios.post("/bagong/employee/get", [v1]).then(
          (r) => {
            let emp = r.data;
            record.Age = emp.Detail.Age;
            record.Gender = emp.Detail.Gender;
            record.Job = emp.Detail.Position;
            record.WorkingPeriod = emp.Detail.WorkingPeriod;
            record.Company = emp.CompanyID;
          },
          (e) => util.showError(e)
        );
      }
      break;
    case "Role":
      record.Employee = "";
      record.Age = "";
      record.Gender = "";
      record.Job = "";
      record.WorkingPeriod = "";
      record.Company = "";
      break;
    case "Asset":
      record.Unit = "";
      record.Type = "";
      record.Damage = "";
      // record.PartEquipment = "";
      record.TotalCost = 0;
      record.Remark = "";
      if (v1) {
        axios.post("/tenant/asset/get", [v1]).then(
          (r) => {
            let asset = r.data;
            record.Unit = asset.AssetType;
            record.Type = asset.DriveType;
          },
          (e) => util.showError(e)
        );
      }
      break;
    default:
      break;
  }
}
function injuredPerson(record) {
  emit("showInjured", record.Injured);
}
function PartEquipmentAsset(record) {
  emit("showPartEquipmentAsset", record.PartEquipment);
}

onMounted(() => {
  loadFromInvolvement();
});
defineExpose({
  updateGridLine,
});
</script>
