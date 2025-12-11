<template>
    <div class="w-full p-6 bg-white rounded-lg shadow">
      <!-- Tabs -->
      <div class="flex justify-between space-x-4 border-b border-gray-300 mb-4">
          <div class="flex tab_container grow">
              <div
                  :class="{
                  tab_selected: data.activeTab === 'Written Test',
                  tab: data.activeTab !== 'Written Test',
                  }"
                  @click="handleClickTabs('Written Test')"
              >
                  Written Test
              </div>
              <div
                  :class="{
                  tab_selected: data.activeTab === 'Practice',
                  tab: data.activeTab !== 'Practice',
                  }"
                  @click="handleClickTabs('Practice')"
              >
                  Practice
              </div>
          </div>
          
          <div class="flex gap-[2px] justify-end">
              <s-button
                  class="btn_primary"
                  label="Save"
                  icon="content-save"
                  @click="saveAssesment()"
              />
              <s-button
                  class="btn_warning back_btn"
                  label="Back"
                  icon="rewind"
                  @click="() => {backAssesmentStaff()}"
              />
          </div>
      </div>

      <!-- Content Written Test -->
      <div v-if="data.activeTab === 'Written Test'" class="flex space-x-6">
        <!-- Sidebar -->
        <div class="w-1/4 bg-white border-gray-300 border-r p-4">
            <h3 class="font-semibold text-lg mb-4">Assessment Tracking</h3>
            <ul class="space-y-2">
                <li v-for="(itemStage, index) in props.assesmentStage" :key="index">
                <button
                    :class="{
                    'tab-active font-semibold': props.activeStage === itemStage.name,
                    }"
                    class="w-full text-left py-2 px-4 hover:bg-gray-100 rounded"
                    @click="onHandleClickStage(itemStage)"
                >
                    {{ itemStage.name }}
                </button>
                </li>
            </ul>
        </div>

        <!-- Main Content -->
        <div v-if="props.selectedStage.templates?.length > 0" class="w-3/4 bg-white rounded-lg p-6">
          <!-- Tab List -->
          <div class="flex tab_container grow border-b border-gray-300 pb-2 mb-4">
            <button
              v-for="(template, idx) in props.selectedStage.templates" :key="idx"
              :class="{
                  tab_selected: props.activeQuestionTab === template.TemplateID,
                  tab: props.activeQuestionTab !== template.TemplateID,
              }"
              @click="onHandleClickTemplate(template)"
            >
              {{ template.TemplateName }}
            </button>
          </div>

          <!-- Score Cards -->
          <div class="grid grid-cols-3 gap-4 mb-6">
            <div class="bg-gray-50 p-4 rounded-lg shadow">
              <h4 class="text-lg font-semibold text-gray-600">Total Score</h4>
              <p class="text-2xl font-bold tab-active">{{props.selectedTemplate.Score || "0"}}</p>
            </div>
          </div>

          <!-- Question List -->
          <div class="space-y-4">
            <div>
              <h5 class="font-semibold text-gray-600">Answer</h5>
              <div v-show="props.selectedTemplate.Status === 'DONE'">
                <!-- pilihan ganda -->
                <div class="bg-gray-50 p-3 rounded-lg rounded-b-none shadow gap-2 grid grid-cols-2 gap-x-4">
                  <template v-for="(answer, idxAnswer) in props.selectedTemplate.answer" :key="idxAnswer">
                    <div v-if="answer.Details" class="question-multiple">
                      <p class="mb-2"><strong>{{ answer.No }}. {{ answer.Question }}</strong></p>
                      <image-candidate-test
                        v-model="answer.QuestionID"
                        journal-type="MCU_ITEM_TEMPLATE_LINES"
                      />
                      <ul class="choices">
                        <li v-for="(option, idxOption) in answer.Details" :key="idxOption">
                          <input type="radio" id="q1a" value="A" disabled :checked="option.IsAnswer">
                          
                          <div>
                            <label :for="'q1a' + idxAnswer + idxOption">{{option.Name}}</label>
                            <image-candidate-test
                              v-model="answer.QuestionID"
                              journal-type="MCU_ITEM_TEMPLATE_LINES_LIST"
                            />
                          </div>

                          <span class="correct" v-if="option.Value"> (Correct)</span>
                          <span class="incorrect" v-else-if="!option.Value && option.IsAnswer"> (Incorrect)</span>
                        </li>
                      </ul>
                    </div>
                  </template>
                </div>

                <!-- essay -->
                <div class="bg-gray-50 p-3 rounded-lg rounded-t-none shadow gap-2 grid grid-cols-1">
                  <template v-for="(answer, idxAnswer) in props.selectedTemplate.answer" :key="idxAnswer">
                    <div v-if="!answer.Details" class="question-essay">
                      <p class="mb-2"><strong>{{ answer.No }}. {{ answer.Question }}</strong></p>
                      <image-candidate-test
                        v-model="answer.QuestionID"
                        journal-type="MCU_ITEM_TEMPLATE_LINES"
                      />
                      <div class="grid grid-cols-5 gap-4">
                        <textarea rows="4" cols="100" readonly v-model="answer.Answer" class="col-span-3 p-3"></textarea>
                        <s-input
                          label="Max Score"
                          caption="Max Score"
                          kind="number"
                          v-model="answer.MaxScore"
                          disabled
                        ></s-input>
                        <s-input
                          label="Score"
                          caption="Score"
                          kind="number"
                          v-model="answer.Score"
                          @change="(name, v1, v2) => {answer.Score = v1}"
                        ></s-input>
                      </div>
                    </div>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="w-3/4 bg-white rounded-lg shadow p-6">
          <div class="w-full h-full flex justify-center items-center font-bold">
            No Data
          </div>
        </div>
      </div>

      <!-- Other Tabs -->
      <div v-if="data.activeTab === 'Practice'" class="flex space-x-6">

        <div class="w-full p-6">
          <div class="flex gap-3">
            <s-input
              ref="refCustomer"
              v-model="data.practiceItemTemplateBasicMovement"
              lookup-key="_id"
              label="Item Template"
              class="w-3/5 mb-3"
              use-list
              :lookup-url="`/she/mcuitemtemplate/find`"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
            ></s-input>

            <div class="flex w-2/5  items-center justify-between">
              <div class="w-2/4">
                Final Score: {{ finalScore }}
              </div>
              <div class="w-2/4">
                Result: {{ finalScore >= 75 ? "Lulus" : "Tidak Lulus" }}
              </div>
            </div>
          </div>

          <GridPracticeTest
            :templateId="data.practiceItemTemplateBasicMovement"
            v-model="data.recordPracticeTest"
          ></GridPracticeTest>
        </div>
      </div>
    </div>
</template>

<script setup>
import { reactive, ref, inject, computed, watch } from "vue";
import { util, SButton, SGrid, SInput } from "suimjs";
import GridPracticeTest from "./GridPracticeTest.vue";
import ImageCandidateTest from "../ImageCandidateTest.vue";

const axios = inject("axios");

const props = defineProps({
    view: {type: String, default: () => ""},
    selectedAssesmentMechanic: {type: Object, default: () => {}},
    assesmentStage: {type: Array, default: () => [
        { name: "Pre-test", templates: [] },
        { name: "Post-test", templates: [] },
    ]},
    selectedStage: {type: Object, default: () => {}},
    selectedTemplate: {type: Object, default: () => {}},
    activeStage: {type: String, default: () => ""},
    activeQuestionTab: {type: String, default: () => ""},
});

const emit = defineEmits({
	"update:view": null,
	refreshGridAssesment: null,
  handleClickStage: null,
  handleClickTemplate: null,
})

const data = reactive({
  activeTab: "Written Test",
  cfgPractice: {},
  practiceItemTemplate: "",
  recordPracticeTest: [],
  practiceTest: {},
})

watch(
  () => props.selectedAssesmentMechanic,
  (nv, old) => {
    getDataPractice();
  }
)

const finalScore = computed({
  get () {
    let totalScore = 0;
    let finalScore = 0;

    data.recordPracticeTest?.forEach(e => {
      totalScore = totalScore + e.Score;
    })
    
    finalScore = totalScore > 0 ? totalScore / data.recordPracticeTest.length : 0;
    return finalScore;
  }
})

function backAssesmentStaff() {
  emit("update:view", "tdc");
  data.activeTab = "Written Test";
  data.practiceItemTemplate = "";
  data.recordPracticeTest = [];
  
  util.nextTickN(2, () => {
    emit("refreshGridAssesment");
  });
}

function handleClickTabs(params) {
  data.activeTab = params
}

function onHandleClickStage(param) {
    emit("handleClickStage", param);
}

function onHandleClickTemplate(param) {
    emit("handleClickTemplate", param);
}

//Assesment Practice 
function saveAssesment() {
  if (data.activeTab === 'Written Test') {
    let multipleScore = 0;
    const answerDetails = props.selectedTemplate.answer.map(e => {
      if(e.Details) {
        const detail = e.Details.find(o => {return o.IsAnswer && o.Value})
        multipleScore += detail ? detail.Score : 0
      } else {
        const map = {
          QuestionID: e.QuestionID,
          Score: e.Score,
        }
        return map
      }
    })
    const param = {
      TrainingCenterID: props.selectedAssesmentMechanic.TrainingCenterID,
      TemplateTestID: props.selectedTemplate.TemplateID,
      EmployeeID: props.selectedAssesmentMechanic.EmployeeID,
      Details: answerDetails,
    }
    
    axios.post("/hcm/tdc/assign-essay-score", param).then(
      (r) => {
        const paramSubmit = {
          ParticipantID: props.selectedAssesmentMechanic._id,
          TemplateTestID: props.selectedTemplate.TemplateID,
          MultipleScore: multipleScore
        }
        axios.post("/hcm/tdc/submit-essay", paramSubmit).then(
          (res) => {
            util.nextTickN(2, () => {
              // getQuestionAnswer()
              backAssesmentStaff()
            });
          },
          (err) => {
            util.showError(err);
          }
        );

      },
      (e) => {
        util.showError(e);
      }
    );

  } else if (data.activeTab === 'Practice') {
    const dtPractice = data.recordPracticeTest;

    if (dtPractice.length > 0) {
      const params = {
        _id: data.practiceTest?._id || "",
        Date: data.practiceTest?.Date || new Date(),
        TrainingCenterID: data.practiceTest?.TrainingCenterID || props.selectedAssesmentMechanic.TrainingCenterID,
        EmployeeID: data.practiceTest?.EmployeeID || props.selectedAssesmentMechanic.EmployeeID,
        Type: "",
        TemplateID: data.practiceTest?.TemplateID || data.practiceItemTemplateBasicMovement || "",
        FinalScore: finalScore.value || 0,
        Detail: dtPractice,
      }
      
      const url = params?._id !== "" ? "/hcm/tdcpracticescore/update" :  "/hcm/tdcpracticescore/save"
      axios.post(url, params).then(
        (r) => {
          util.nextTickN(2, () => {
            getDataPractice();
          })
        },
        (e) => {
          util.showError(err);
        }
      )
    }
  }
}

function getDataPractice() {
  const params = {
    Select: ["_id", "Date", "TrainingCenterID", "EmployeeID", "Type", "TemplateID", "FinalScore", "Detail"],
    Where: {
      Op: "$and",
      items: [
        { Op: "$eq", Field: "TrainingCenterID", Value: props.selectedAssesmentMechanic.TrainingCenterID },
        { Op: "$eq", Field: "EmployeeID", Value: props.selectedAssesmentMechanic.EmployeeID },
      ]
    }
  }
  axios.post(`/hcm/tdcpracticescore/find`, params).then(
    (r) => {
      const dt = r.data
      const dtPractice = dt.find(e => {return  e.Type == ""});

      const dtTemplate = {
        _id: "",
        Date: new Date(),
        TrainingCenterID: props.selectedAssesmentMechanic.TrainingCenterID,
        EmployeeID: props.selectedAssesmentMechanic.EmployeeID,
        Type: "",
        TemplateID: "",
        FinalScore: 0,
        Detail: [],
      }
      
      data.recordPracticeTest = dtPractice ? dtPractice.Detail : [];
      data.practiceTest = dtPractice ? dtPractice : dtTemplate;
      data.practiceTest.Type = "";
      data.practiceTest.TemplateID = dtPractice ? dtPractice.TemplateID : "";
    },
    (e) => {
      util.showError(e);
    }
  )
}
//End Assesment Practice

</script>

<style scoped>
  .question-multiple {
      margin-bottom: 20px;
      /* width: 50%; */
  }
  .question-essay {
      margin-bottom: 20px;
      width: 100%;
  }
  .choices {
      list-style: none;
      padding: 0;
  }
  .choices li {
      margin: 5px 0;
      align-items: center;
      display: flex;
      gap: 10px;
  }
  .correct {
      color: #10A55F;
      font-weight: bold;
  }
  .incorrect {
      color: #E1514A;
      font-weight: bold;
  }

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

  .total_score {
    min-height: 35px;
    width: 100%;
    font-weight: bold;
    justify-content: flex-end;
    display: flex;
    align-items: center;
    font-size: 14px;
  }

  .total_score_title {
    min-height: 35px;
    width: 100%;
    font-weight: bold;
    display: flex;
    align-items: center;
    font-size: 14px;
  }
</style>