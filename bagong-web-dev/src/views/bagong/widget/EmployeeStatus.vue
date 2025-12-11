<!-- <template>
    <div>
        <s-form :config="data.config" v-model="data.value" keep-label>
            <template #buttons>&nbsp;</template>
        </s-form>
    </div>
</template>
<script setup>
import { reactive, onMounted, inject } from "vue";
import { SForm, loadFormConfig, util } from "suimjs";

const props = defineProps({
    modelValue: { type: Object, default: () => { } },
});

const emit = defineEmits({
    "update:modelValue": null,
});

const data = reactive({
    config: {},
    value: props.modelValue,
});

const axios = inject("axios");

onMounted(() => {
    // console.log(data)
    // console.log(props)
    loadFormConfig(axios, "/bagong/employeedetail/formconfig").then(
        (r) => {
            console.log("config", r)
            data.config = r;
        },
        (e) => util.showError(e)
    );
});
</script>
<style></style> -->

<template>
  <div>
    <s-form
      :config="data.config"
      v-model="data.value"
      @postSubmitForm="postSubmitForm"
    >
      <template #input_BasicSalary="{ item, config }">
        <div class="flex gap-3 items-center">
          <div
            class="grow"
            :class="[data.editedSalary ? '' : 'disable-salary']"
          >
            <s-input
              :class="data.editedSalary ? '' : 'blured'"
              :disabled="data.editedSalary !== true"
              :kind="config.kind"
              :label="config.label"
              v-model="item.BasicSalary"
              :required="config.required"
            ></s-input>
          </div>
          <s-button
            v-if="!data.editedSalary"
            class="mt-2 btn_primary"
            @click="onEditSallary(item)"
            icon="pencil"
          />
        </div>
      </template>
      <template #buttons>&nbsp;</template>
    </s-form>
  </div>
</template>

<script setup>
import { reactive, onMounted, inject } from "vue";
import {
  SForm,
  SInput,
  SButton,
  loadFormConfig,
  util,
  createFormConfig,
} from "suimjs";
import moment from "moment";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  config: {},
  value: props.modelValue,
  editedSalary: false,
});

onMounted(() => {
  generateConfig();
  util.nextTickN(2, () => {
    if (data.value.JoinDate != "0001-01-01T00:00:00Z") {
      var joinDate = moment(data.value.JoinDate);
      var nowDate = moment(new Date());
      var diff = nowDate.diff(joinDate, "years");
      data.value.WorkingPeriod = diff;
    }
    bluredBasicSalary(data.value.BasicSalary, 2000);
  });
});

function generateConfig() {
  const cfg = createFormConfig("", true);
  cfg.addSection("", true).addRowAuto(
    3,
    {
      field: "Level",
      label: "Level",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=LME",
    },
    {
      field: "Position",
      label: "Position",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=PTE",
    },
    {
      field: "Grade",
      label: "Grade",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=GDE",
    },
    {
      field: "Department",
      label: "Department",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=DME",
    },
    {
      field: "POH",
      label: "POH",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=PME",
    },
    {
      field: "Group",
      label: "Group",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=GME",
    },
    {
      field: "SubGroup",
      label: "Sub Group",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=SGE",
    },
    {
      field: "UserCustomer",
      label: "User",
      kind: "text",
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/customer/find",
    },
    {
      field: "BPJSTKProgram",
      label: "BPJS TK Program",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=PBJTK",
    },
    // {
    //   field: "BPJSTKPercentage",
    //   label: "BPJS TK Percentage",
    //   kind: "number",
    //   required: true,
    // },
    // {
    //   field: "BPJSKESPercentage",
    //   label: "BPJS KES Percentage",
    //   kind: "number",
    //   required: true,
    // },
    {
      field: "DirectSupervisor",
      label: "Direct Supervisor",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/employee/find",
    },
    { hide: true },
    { hide: true }
  );
  data.config = cfg.generateConfig();

  cfg.addSection("", true).addRowAuto(
    3,
    {
      field: "WorkingPeriod",
      label: "Working Periode in Year",
      kind: "number",
      required: true,
    },
    {
      field: "EmployeeStatus",
      label: "Employee Status",
      kind: "text",
      required: true,
      useList: true,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["Name"],
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=ESM",
    },
    {
      field: "BasicSalary",
      label: "Basic Salary",
      kind: "number",
      required: true,
    },
    {
      field: "PermanentEmployeeDate",
      label: "Permanent Employee Date",
      kind: "date",
    },
    {
      field: "WorkerStatus",
      label: "Worker Status",
      kind: "text",
      required: true,
      useList: true,
      items: ["Aktif", "Non Aktif"],
    },
    {
      field: "ResignationDate",
      label: "Resignation Date",
      kind: "date",
    }
  );
  data.config = cfg.generateConfig();
}
function onEditSallary(record) {
  record.BasicSalary = 0;
  data.editedSalary = true;
}
// function bluredBasicSalary(value, time) {
//   // implement blur effect in component
//   const el = document.querySelector('input[placeholder="Basic Salary"]');
//   el.setAttribute("class", "input_field text-right");
//   if (value > 0 || value != "") {
//     setTimeout(() => {
//       el.setAttribute("class", "input_field text-right blured");
//     }, time);
//   } else {
//     setTimeout(() => {
//       el.setAttribute("class", "input_field text-right");
//     }, time);
//   }
// }
// function basicSalaryFocus() {
//   bluredBasicSalary(data.value.BasicSalary, 3000);
// }
// function basicSalaryChange(name, v1, v2, old) {
//   bluredBasicSalary(v1, 2500);
// }
</script>

<style>
.disable-salary .blured input {
  filter: blur(3.6px);
}
</style>
