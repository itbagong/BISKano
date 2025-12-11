<template>
  <div>
    <data-list
      ref="listControl"
      hide-title
      no-gap
      grid-editor
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-hide-detail
      grid-hide-select
      grid-no-confirm-delete
      grid-hide-paging
      init-app-mode="grid"
      grid-mode="grid"
      new-record-type="grid"
      grid-config="/hcm/testschedule/gridconfig"
      form-config="/hcm/testschedule/formconfig"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @alterGridConfig="alterGridConfig"
      :grid-fields="['TemplateID', 'TrainerType', 'TrainerName', 'TestType']"
    >
      <template #grid_TemplateID="{item}">
        <s-input
          ref="refTemplateID"
          hide-label
          v-model="item.TemplateID"
          use-list
          lookup-url="/she/mcuitemtemplate/find?Module=TDC&Menu=TRA"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-100"
        ></s-input>
      </template>
      <template #grid_TrainerType="{ item }">
        <s-input
          ref="refTrainerType"
          label="Trainer Type"
          v-model="item.TrainerType"
          use-list
          :items="['Internal', 'External']"
          class="w-100"
          @change="(field, v1, v2, old, ctlRef) => {
            item.TrainerName = ''
          }"
        ></s-input>
      </template>
      <template #grid_TrainerName="{ item }">
        <s-input
          v-if="item.TrainerType == 'Internal'"
          ref="refTrainerName"
          label="Trainer Name"
          v-model="item.TrainerName"
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-100"
        ></s-input>
        <s-input
          v-else
          ref="refTrainerName"
          label="Trainer Name"
          v-model="item.TrainerName"
          class="w-100"
        ></s-input>
      </template>
      <template #grid_TestType="{ item }">
        <s-input
          ref="refTestType"
          label="Test Type"
          v-model="item.TestType"
          use-list
          :items="['Pre-Test', 'Post-Test']"
          class="w-100"
          @change="(field, v1, v2, old, ctlRef) => {}"
        ></s-input>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, onMounted, watch } from "vue";
import { DataList, SInput, util } from "suimjs";

const listControl = ref(null);

const axios = inject("axios");

const props = defineProps({
  testId: { type: String, default: "" },
  testScheduleType: { type: String, default: "" },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
});

function alterGridConfig(cfg) {
  cfg.fields = cfg.fields.filter(
    (el) => ["TestScheduleType", "Status"].indexOf(el.field) == -1
  );
}

function newRecord(r) {
  r.TestID = props.testId;
  r.TestScheduleType = props.testScheduleType;
  r.TemplateID = "";
  r.Status = "OPEN";
  r.DateFrom = "";
  r.DateTo = "";
  r.TrainerName = "";
  r.TrainerType = "";
  r.TestType = "";

  data.records.push(r);
  updateItems();
}

function getsMaterial() {
  if (!props.testId) return;

  axios
    .post(`/hcm/testschedule/gets?TestID=${props.testId}`, {
      Sort: ["_id"],
    })
    .then(
      async (r) => {
        data.records = r.data.data;
        updateItems();
      },
      (e) => {
        util.showError(e);
      }
    );
}

function updateItems() {
  listControl.value?.setGridRecords(data.records);
}

function onGridRowDelete(records, index) {
  let payload = records;
  const dtRecord = payload.items[index];
  
  if(!dtRecord._id) {
    const newRecords = data.records.filter((dt, idx) => {
      return idx != index;
    });
    data.records = newRecords;
    updateItems();
    util.showInfo("Data has been delete");
  } else {
    axios.post("/hcm/testschedule/delete-schedule", payload.items[index]._id).then(
      async (r) => {
        const newRecords = data.records.filter((dt, idx) => {
          return idx != index;
        });
        data.records = newRecords;
        updateItems();
        util.showInfo("Data has been delete");
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

onMounted(() => {
  getsMaterial();
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
