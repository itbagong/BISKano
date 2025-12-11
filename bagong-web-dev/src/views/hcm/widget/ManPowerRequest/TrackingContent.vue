<template>
  <div
    v-if="props.candidateId !== '' && props.stageId"
    class="p-2 tracking-content grid grid-cols-1 gap-y-2"
  >
    <div class="flex gap-4 items-start border-b pb-2">
      <loader
        kind="skeleton"
        v-if="data.loadingEmployee || data.loadingEmployeeDetail"
      />
      <template v-else>
        <!-- <div class="min-w-[100px] h-[100px] bg-cover bg-zinc-200"></div> -->
        <Photo
          no-photo-class="min-w-[100px] h-[100px] bg-cover bg-zinc-200"
          img-class="w-[120px] h-[100px] object-cover"
          v-model="item.CandidateID"
        ></Photo>

        <div class="flex-col">
          <div class="flex gap-4 items-center">
            <h1 class="text-[1.3em] font-bold">{{ data.employee?.Name }}</h1>
          </div>
          <div class="[&>*]:inline-block [&>*]:mr-4 text-gray-500">
            <div
              class="[&>*]:inline-block"
              v-if="data.employeeDetail?.Age != ''"
            >
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              {{ data.employeeDetail?.Age }} Years
            </div>
            <div
              class="[&>*]:inline-block"
              v-if="data.employeeDetail?.PlaceOfBirth != ''"
            >
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              {{ data.employeeDetail?.PlaceOfBirth }},
              {{
                moment(data.employeeDetail?.DateOfBirth).format("DD MMM YYYY")
              }}
            </div>
            <div
              class="[&>*]:inline-block"
              v-if="data.employeeDetail?.Gender != ''"
            >
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              <div>
                <s-input
                  read-only
                  class="ml-1"
                  v-model="data.employeeDetail.Gender"
                  use-list
                  lookup-url="/tenant/masterdata/find?MasterDataTypeID=GEME"
                  lookup-key="_id"
                  :lookup-labels="['Name']"
                  :lookup-searchs="['_id', 'Name']"
                >
                </s-input>
              </div>
            </div>
            <div class="[&>*]:inline-block" v-if="data.employee?.Email != ''">
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              {{ data.employee?.Email }}
            </div>
            <div
              class="[&>*]:inline-block"
              v-if="data.employeeDetail?.Phone != ''"
            >
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              {{ data.employeeDetail?.Phone }}
            </div>
            <div
              class="[&>*]:inline-block"
              v-if="data.employeeDetail?.Address != ''"
            >
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              {{ data.employeeDetail?.Address }}
            </div>
            <div
              class="[&>*]:inline-block"
              v-if="data.employeeDetail?.Domicile != ''"
            >
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              {{ data.employeeDetail?.Domicile }}
            </div>
            <div
              class="[&>*]:inline-block"
              v-if="data.employeeDetail?.LastEducation != ''"
            >
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              {{ data.employeeDetail?.LastEducation }}
            </div>
            <div
              class="[&>*]:inline-block"
              v-if="data.employeeDetail?.Major != ''"
            >
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              {{ data.employeeDetail?.Major }}
            </div>
            <div
              class="[&>*]:inline-block"
              v-if="data.employeeDetail?.WorkingExperience != 0"
            >
              <mdicon
                name="checkbox-blank-circle"
                class="circle text-gray-400"
                size="8"
              />
              {{ data.employeeDetail?.WorkingExperience }} years
            </div>
          </div>
        </div>
        <div>
          <action-attachment
            ref="PreviewAction"
            kind="DOCUMENTS"
            :ref-id="data.employee._id"
            read-only
            icon="file-download"
            :icon-width="30"
            buttonClass="border rounded-lg px-2 py-3"
            class=""
            tooltip="Download CV"
            :tags="[`DOCUMENTS_CHECKLIST_${data.employee._id}_CV`]"
          >
          </action-attachment>
        </div>
      </template>
    </div>
    <div class="w-full mt-1 stage">
      <Screening v-if="stage == 'Screening'" :id="stageId" />
      <Psikotest
        v-else-if="stage == 'PshycologicalTest'"
        :id="stageId"
        :item="props.item"
      />
      <interview v-else-if="stage == 'Interview'" :id="stageId" />
      <TechnicalInterview
        v-else-if="stage == 'TechnicalInterview'"
        :id="stageId"
      />
      <Mcu v-else-if="stage == 'MCU'" :id="stageId" />
      <Training v-else-if="stage == 'Training'" :id="stageId" />
      <ol-ploting v-else-if="stage == 'OLPlotting'" :id="stageId" />
      <Pkwtt v-else-if="stage == 'PKWTT'" :id="stageId" />
      <Onboarding v-else-if="stage == 'OnBoarding'" :id="stageId" />
    </div>
  </div>
  <div
    v-else
    class="min-h-full flex items-center justify-center text-[1.5em] font-bold text-gray-400 mb-4"
  >
    Please select candidate
  </div>
</template>
<script setup>
import { reactive, inject, onMounted, watch } from "vue";
import moment from "moment";

import Loader from "@/components/common/Loader.vue";
import ActionAttachment from "@/components/common/ActionAttachment.vue";

import Screening from "./Stage/Screening.vue";
import Psikotest from "./Stage/Psikotest.vue";
import Interview from "./Stage/Interview.vue";
import TechnicalInterview from "./Stage/TechnicalInterview.vue";
import Mcu from "./Stage/Mcu.vue";
import OlPloting from "./Stage/OlPloting.vue";
import Training from "./Stage/Training.vue";
import Pkwtt from "./Stage/Pkwtt.vue";
import Onboarding from "./Stage/Onboarding.vue";
import Photo from "../EmployeePhoto.vue";

import { util, SInput, SButton, STooltip } from "suimjs";

const props = defineProps({
  manPowerId: { type: String, default: "" },
  stage: { type: String, default: "" },
  candidateId: { type: String, default: "" },
  stageId: { type: String, default: "" },
  item: { type: Object, default: null },
});

const axios = inject("axios");

const data = reactive({
  employee: {},
  employeeDetail: {},
  loadingEmployee: false,
  loadingEmployeeDetail: false,
});
function fetchEmployee() {
  data.loadingEmployee = true;
  axios
    .post("/tenant/employee/get", [props.candidateId])
    .then((r) => {
      data.employee = r.data;
    })
    .catch((e) => {
      data.employee = {};
      util.showError(e);
    })
    .finally((e) => {
      data.loadingEmployee = false;
    });
}
function fetchEmployeeDetail() {
  data.loadingEmployeeDetail = true;
  axios
    .post(`/bagong/employeedetail/find?EmployeeID=${props.candidateId}`)
    .then((r) => {
      if (r.data.length > 0) {
        data.employeeDetail = r.data[0];
      } else {
        data.employeeDetail = {};
      }
    })
    .catch((e) => {
      data.employeeDetail = {};
      util.showError(e);
    })
    .finally((e) => {
      data.loadingEmployeeDetail = false;
    });
}
function refresh() {
  if (props.candidateId !== "") {
    fetchEmployee();
    fetchEmployeeDetail();
  } else {
    data.employee = {};
    data.employeeInfo = {};
  }
}

watch(
  () => props.candidateId,
  (nv) => {
    refresh();
  }
);

onMounted(() => {
  refresh();
});
</script>
<style>
.tracking-content > .stage .status-text {
  @apply text-[1.3em] p-1;
}
</style>
