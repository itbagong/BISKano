<template>
  <div class="mb-2 flex flex-col header">
    <div class="flex tab_container grow">
      <div
        v-for="(tabTitle, tabIdx) in data.tabs"
        @click="data.currentTab = tabIdx"
        :class="{
          tab_selected: data.currentTab == tabIdx,
          tab: data.currentTab != tabIdx,
        }"
      >
        {{ tabTitle }}
      </div>
    </div>
    <div class="form_inputs" v-show="data.currentTab == 0">
      <div>
        <div class="flex tab_container grow">
          <div
            v-for="(r, idx) in data.generalRecords"
            :key="idx"
            :class="data.selectedTab === r.TemplateID ? 'tab_selected' : 'tab'"
            @click="data.selectedTab = r.TemplateID"
          >
            <s-input
              v-model="r.TemplateID"
              read-only
              class="w-full mb-2"
              use-list
              lookup-url="/she/mcuitemtemplate/find"
              lookup-key="_id"
              :lookup-labels="['Name']"
            />
          </div>
        </div>
      </div>
      <div class="mb-5"></div>
      <div
        class="grid grid-cols-1 gap-y-2 shadow-lg p-2 border-t-8 border-primary mb-4"
      >
        <div v-if="selectedItem?.Detail">
          <div v-if="!Array.isArray(selectedItem.Detail)">
            <div class="text-center border py-2 mb-2">
              <p class="text-md font-medium mb-1">IQScore</p>
              <p class="text-4xl font-medium">
                {{ selectedItem.Detail?.IQScore }}
              </p>
            </div>
            <div class="text-center border py-2">
              <p class="text-md font-medium mb-1">Score</p>
              <p class="text-4xl font-medium">
                {{ selectedItem.Detail?.Score }}
              </p>
            </div>
          </div>
          <div
            v-else
            class="w-full grid grid-cols-2 border-l border-t border-b"
          >
            <div
              v-for="(r, idx) in selectedItem.Detail"
              :key="idx"
              class="w-full border-r"
            >
              <p
                :class="`w-full text-center font-medium border-b uppercase ${
                  r.IsMostQuestionType && 'bg-slate-100'
                }`"
              >
                {{
                  r.IsMostQuestionType
                    ? "Menggambarkan Diri"
                    : "Tidak Menggambarkan Diri"
                }}
              </p>
              <div
                v-for="(pr, idx) in r.Detail"
                :key="idx"
                :class="`grid grid-cols-2 border-b ${
                  r.IsMostQuestionType && 'bg-slate-100'
                }`"
              >
                <p class="text-center border-r">{{ pr.Answer }}</p>
                <p class="text-center">{{ pr.Count }}</p>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="border p-2">
          <p class="text-center">No Data Score</p>
        </div>
      </div>
    </div>
    <div class="form_inputs" v-show="data.currentTab == 1">
      <div class="flex justify-end mb-2">
        <s-button
          v-if="!readOnly"
          class="btn_primary submit_btn"
          label="Save"
          @click="onSaveTestSchedule()"
        />
      </div>
      <material
        v-model="data.records"
        :test-id="talentDevelopmentAssesmentID"
        test-schedule-type="TALENTDEVELOPMENT"
        :read-only="readOnly"
      ></material>
    </div>
  </div>
</template>
<script setup>
import { computed, reactive, inject, onMounted } from "vue";
import { SButton, SModal, SInput, util, loadFormConfig } from "suimjs";
import Loader from "@/components/common/Loader.vue";
import Material from "../../../widget/Material.vue";

const axios = inject("axios");

const props = defineProps({
  generalRecords: { type: Array, default: () => [] },
  talentDevelopmentAssesmentID: { type: String, default: "" },
  readOnly: { type: Boolean, default: false },
});

const data = reactive({
  activeIndex: null,
  tabs: ["General", "Material"],
  currentTab: 0,
  generalRecords: props.generalRecords,
  frmCfg: {},
  records: [],
  selectedTab: "",
});

const toggle = (index) => {
  data.activeIndex = data.activeIndex === index ? null : index;
};
function onSaveTestSchedule() {
  axios
    .post("hcm/testschedule/save-talent-development-schedule", data.records)
    .then(
      async (r) => {
        util.showInfo("Material has been successful save");
      },
      (e) => {
        util.showError(e);
      }
    );
}

const selectedItem = computed(() => {
  return data.generalRecords.find((v) => v.TemplateID === data.selectedTab);
});

onMounted(() => {
  if (data.generalRecords.length > 0) {
    data.selectedTab = data.generalRecords[0].TemplateID
  }
  loadFormConfig(axios, "/hcm/psychologicaltest/formconfig").then(
    (r) => {
      data.frmCfg = r;
    },
    (e) => util.showError(e)
  );
});
</script>
<style scoped>
.score-card {
  @apply bg-white shadow-md rounded-md px-6 py-4 flex flex-col items-center;
}
</style>
