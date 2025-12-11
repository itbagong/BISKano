<template>
  <s-card class="rounded-md w-full" hide-title>
    <div class="px-2 max-w-[400px] relative" v-if="data.mode == 'form'">
      <div name="manpowerrequest">
        <s-input
          label="Manpower Request"
          class="w-full"
          use-list
          :items="data.manpowerList"
          v-model="data.selectedManPower"
        />
        <s-button
          label="Submit"
          class="w-full btn_primary flex justify-center mt-6"
          @click="() => submitManpower()"
        />
      </div>
    </div>
    <s-grid
      v-else
      ref="gridParticipant"
      hide-select
      hide-search
      hide-sort
      hide-refresh-button
      hide-new-button
      hide-delete-button
      hide-action
      :config="data.cfg"
      :grid-fields="['IsCheck']"
    >
      <template #item_IsCheck="{ item }">
        <s-input
          ref="refIsCheck"
          kind="checkbox"
          v-model="item.IsCheck"
          class="w-100"
        ></s-input>
      </template>
      
      <template #header_buttons_1="{ config }">
        <s-button
          class="bg-primary text-white font-bold w-full flex justify-center"
          label="Back"
          @click="data.mode = 'form'"
        ></s-button>
        <s-button
          class="bg-primary text-white font-bold w-full flex justify-center"
          label="Generate"
          @click="onGenerate"
        ></s-button>
      </template>
    </s-grid>
  </s-card>
</template>
<script setup>
import { SGrid, util, SButton, SInput } from "suimjs";
import { reactive, ref, computed, inject, watch, onMounted } from "vue";
import helper from "@/scripts/helper.js";
import moment from "moment";
const axios = inject("axios");

const gridParticipant = ref(null);
const props = defineProps({
  value: {type:String, default: ""},
});

const emit = defineEmits({
  generate: null,
});

const data = reactive({
  modal: {
    isShow: false,
  },
  mode: "form",
  selectedManPower: '',
  manpowerList: [],
  participantList: [],
  cfg: {},
});

function getTrainingParticipant() {
  const param = {
    Where: { Stage: "Training", JobID: data.selectedManPower }
  }
  axios.post("hcm/tracking/get-applicant", param).then(
    (r) => {
      const dt = r.data.data
      dt.map(e => {
        e.IsCheck = true;
        return e
      })

      util.nextTickN(2, () => {
        gridParticipant.value.setRecords(dt || []);
      });
    },
    (e) => {}
  );
}
function resetModal() {
  data.selectedManPower = '';
  data.participantList = [];
}

function getProject() {
  const url = "/hcm/manpowerrequest/gets";
  let param = {
    Take: -1,
    Sort: ["-LastUpdate"],
    Select: ["_id", "Name"],
    Where: {
      Op: "$and",
      items: [
        { Op: "$eq", Field: "Status", Value: "POSTED" },
        { Op: "$eq", Field: "IsClose", Value: false },
      ]
    }
  }
  axios.post(url, param).then(
    (r) => {
      const dt = r.data.data
      data.manpowerList = dt.map((d) => {
        return {
          key: d["_id"],
          text: d["Name"],
          item: { ...d },
        };
      });
    },
    (e) => {}
  );
}

function submitManpower() {
  data.mode = 'grid';
  getTrainingParticipant();
}

async function onGenerate() {
  const dt = gridParticipant.value.getRecords();
  
  const dtMap = await Promise.all( 
    dt.filter(e => {return e.IsCheck !== false})
    .map(e => {return {
      EmployeeID: e.CandidateID,
      Name: e.Name,
      Position: "",
      Site: "",
      Department: "",
      TestDetails: [],
      TrainingCenterID: props.value,
      _id: "",
      ManpowerRequestID: e.JobVacancyID,
    }})
  );
  emit('generate', dtMap);
}

onMounted(() => {
  resetModal();
  data.mode = "form"
  getProject();
  data.cfg = {
    fields: [
      {
        field: "CandidateID",
        kind: "text",
        label: "Candidate ID",
        labelField: "",
        readType: "show",
        input: {
          field: "CandidateID",
          kind: "text",
          label: "Candidate ID",
          lookupUrl: "",
          placeHolder: "Candidate ID",
        } 
      },
      {
        field: "Name",
        kind: "text",
        label: "Name",
        labelField: "",
        readType: "show",
        input: {
          field: "Name",
          kind: "text",
          label: "Name",
          lookupUrl: "",
          placeHolder: "Name",
        } 
      },
      {
        field: "IsCheck",
        kind: "text",
        label: "",
        labelField: "",
        readType: "show",
        input: {
          field: "IsCheck",
          kind: "text",
          label: "",
          lookupUrl: "",
          placeHolder: "",
        } 
      },
    ],
    setting: {
      idField: "",
      keywordFields: ['_id', 'Name'],
      sortable: ['_id']
    }
  }
  // if (kind.value == "general") data.mode = "grid";
});
</script>
<style>
.grid-action-customer .suim_grid .header {
  @apply mb-0;
}
</style>
