<template>
  <div
    class="grid md:grid-cols-12 grid-cols-1 gap-2 divide-x [&>*]:p-2 min-h-[calc(100vh-280px)]"
  >
    <div class="col-span-2">
      <Stage
        ref="stepCtl"
        :manPowerId="manPowerId"
        :stage-counts="data.applicantCount"
        v-model="data.selectedStage"
        @select="selectStage"
        @refresh="refresh"
      />
    </div>
    <div v-if="data.selectedPshycoTab === 'General'" class="col-span-3">
      <div
        v-if="data.selectedStage === 'PshycologicalTest'"
        class="flex tab_container grow"
      >
        <div
          v-for="tab in data.pshycoTab"
          :class="data.selectedPshycoTab === tab ? 'tab_selected' : 'tab'"
          @click="data.selectedPshycoTab = tab"
        >
          {{ tab }}
        </div>
      </div>
      <br />
      <Candidate
        ref="candidateCtl"
        :manPowerId="manPowerId"
        :stage="data.selectedStage"
        v-model="data.selectedCandidate"
        @select="selectedCandidate"
      />
    </div>
    <div v-if="data.selectedPshycoTab === 'General'" class="col-span-7">
      <Content
        ref="contentCtl"
        :manPowerId="manPowerId"
        :stage="data.selectedStage"
        :candidate-id="data.selectedCandidate.id"
        :stage-id="data.selectedCandidate.stageId"
        :item="data.selectedCandidate.item"
      />
    </div>
    <div v-if="data.selectedPshycoTab === 'Material'" class="col-span-10">
      <div class="flex justify-between">
        <div
          v-if="data.selectedStage === 'PshycologicalTest'"
          class="flex tab_container grow"
        >
          <div
            v-for="tab in data.pshycoTab"
            :class="data.selectedPshycoTab === tab ? 'tab_selected' : 'tab'"
            @click="data.selectedPshycoTab = tab"
          >
            {{ tab }}
          </div>
        </div>
        <s-button
          class="btn_primary submit_btn"
          label="Save"
          @click="onSaveTestSchedule()"
        />
      </div>
      <material
        ref="materialCtl"
        v-model="data.records"
        :test-id="props.manPowerId"
        test-schedule-type="PSYCHOLOGICAL"
      ></material>
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, watch, inject, onMounted } from "vue";
import { SButton, util } from "suimjs";
import Stage from "./TrackingStage.vue";
import Candidate from "./TrackingCandidate.vue";
import Content from "./TrackingContent.vue";
import Material from "../../widget/Material.vue";
import moment from 'moment';

const stepCtl = ref(null);
const candidateCtl = ref(null);
const contentCtl = ref(null);
const materialCtl = ref(null);

const axios = inject("axios");

const props = defineProps({
  manPowerId: { type: String, default: "" },
});

const data = reactive({
  selectedStage: "",
  selectedCandidate: { id: "", stageId: "", item: null },
  pshycoTab: ["General", "Material"],
  selectedPshycoTab: "General",
  records: [],
  applicantCount: [],
});

function selectStage(id) {
  data.selectedStage = id;
  data.selectedCandidate = { id: "", stageId: "", item: null };
}

function selectedCandidate(r) {
  data.selectedCandidate = r;
}

function refresh() {
  stepCtl.value?.refresh();
  candidateCtl.value?.refresh();
}

function onSaveTestSchedule() {
  const payload = data.records.map(item => {
    return {
      ...item,
      DateFrom: new Date(moment(item.DateFrom).format('YYYY-MM-DD')),
      DateTo: new Date(moment(item.DateTo).format('YYYY-MM-DD')),
      DateFromStr: moment(item.DateFrom).format('YYYY-MM-DD'),
      DateToStr: moment(item.DateTo).format('YYYY-MM-DD'),
    }
  })
  axios
    .post("/hcm/testschedule/save-psychological-schedule", payload)
    .then(
      async (r) => {
        materialCtl.value.getsMaterial()
        util.showInfo("Material has been successful save");
      },
      (e) => {
        util.showError(e);
      }
    );
}

function getApplicantCount() {
  axios
    .post("hcm/tracking/get-applicant-count", { JobID: props.manPowerId })
    .then(
      async (r) => {
        data.applicantCount = r.data.reduce((result, item) => {
          result[item.Stage] = item.Count;
          return result;
        }, {});
      },
      (e) => {
        util.showError(e);
      }
    );
}
watch(
  () => data.selectedStage,
  (nv) => {
    if (nv) {
      data.selectedPshycoTab = "General";
    }
  }
);
onMounted(() => {
  getApplicantCount();
});
</script>
