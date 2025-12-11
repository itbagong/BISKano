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
              <div v-show="props.selectedTemplate.Status == 'DONE'">
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
                          <input 
                            type="radio" 
                            :id="'q1a' + idxAnswer + idxOption" 
                            :value="option.Name" 
                            disabled 
                            :checked="option.IsAnswer"
                          >

                          <div>
                            <label :for="'q1a' + idxAnswer + idxOption">{{option.Name}}</label>
                            <image-candidate-test
                              v-model="answer.QuestionID"
                              journal-type="MCU_ITEM_TEMPLATE_LINES_LIST"
                            />
                          </div>
                          <!-- status -->
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
        <!-- Sidebar -->
        <div class="w-1/4 bg-white border-gray-300 border-r p-4">
            <h3 class="font-semibold text-lg mb-4">Practice Tracking</h3>
            <ul class="space-y-2">
              <li>
                <button
                    :class="{
                      'tab-active font-semibold': data.activeTabPractice === 'basic-movement',
                    }"
                    class="w-full text-left py-2 px-4 hover:bg-gray-100 rounded"
                    @click="onHandleClickTabPractice('basic-movement')"
                >
                  Basic Movement
                </button>
              </li>
              <li>
                <button
                    :class="{
                      'tab-active font-semibold': data.activeTabPractice === 'main-road',
                    }"
                    class="w-full text-left py-2 px-4 hover:bg-gray-100 rounded"
                    @click="onHandleClickTabPractice('main-road')"
                >
                  Main Road & DDT
                </button>
              </li>
            </ul>
        </div>

        <div v-if="data.activeTabPractice === 'basic-movement'" class="w-3/4 p-6">
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
                Final Score: {{ finalScoreBasicMovement }}
              </div>
              <div class="w-2/4">
                Result: {{ finalScoreBasicMovement >= 75 ? "Lulus" : "Tidak Lulus" }}
              </div>
            </div>
          </div>

          <GridPracticeTest
            :templateId="data.practiceItemTemplateBasicMovement"
            v-model="data.recordPracticeTestBasicMovement"
          ></GridPracticeTest>
        </div>

        <div v-else class="w-3/4 p-6">
          <div class="flex gap-3">
            <s-input
              ref="refCustomer"
              v-model="data.practiceItemTemplateMainRoad"
              lookup-key="_id"
              label="Item Template"
              class="w-full mb-3"
              use-list
              :lookup-url="`/she/mcuitemtemplate/find`"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
            ></s-input>

            <div class="flex w-2/5  items-center justify-between">
              <div class="w-2/4">
                Final Score: {{ finalScoreMainRoad }}
              </div>
              <div class="w-2/4">
                Result: {{ finalScoreMainRoad >= 75 ? "Lulus" : "Tidak Lulus" }}
              </div>
            </div>
          </div>

          <GridPracticeTest
            :templateId="data.practiceItemTemplateMainRoad"
            v-model="data.recordPracticeTestMainRoad"
          ></GridPracticeTest>
        </div>
      </div>
    </div>
</template>

<script setup>
import { reactive, ref, inject, computed, onMounted, watch } from "vue";
import { util, SButton, SGrid, SInput } from "suimjs";
import GridPracticeTest from "./GridPracticeTest.vue";
import ImageCandidateTest from "../ImageCandidateTest.vue";

const axios = inject("axios");

// const gridPracticeAssesmentStaff = ref(null);

const props = defineProps({
    view: {type: String, default: () => ""},
    selectedAssesmentDriver: {type: Object, default: () => {}},
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
  practiceItemTemplateBasicMovement: "",
  practiceItemTemplateMainRoad: "",
  recordPracticeTestBasicMovement: [],
  recordPracticeTestMainRoad: [],
  activeTabPractice: "basic-movement",
  practiceTestBasicMovement: {},
  practiceTestMainRoad: {},
})

watch(
  () => props.selectedAssesmentDriver,
  (nv, old) => {
    getDataPractice();
  }
)

const finalScoreBasicMovement = computed({
  get () {
    let totalScore = 0;
    let finalScore = 0;

    data.recordPracticeTestBasicMovement?.forEach(e => {
      totalScore = totalScore + e.Score;
    })
    
    finalScore = totalScore > 0 ? totalScore / data.recordPracticeTestBasicMovement.length : 0;
    return finalScore;
  }
})

const finalScoreMainRoad = computed({
  get () {
    let totalScore = 0;
    let finalScore = 0;

    data.recordPracticeTestMainRoad?.forEach(e => {
      totalScore = totalScore + e.Score;
    })
    
    finalScore = totalScore > 0 ? totalScore / data.recordPracticeTestMainRoad.length : 0;
    return finalScore;
  }
})

function backAssesmentStaff() {
  emit("update:view", "tdc");
  data.activeTab = "Written Test";
  data.practiceItemTemplateBasicMovement = "";
  data.practiceItemTemplateMainRoad = "";
  data.activeTabPractice = "basic-movement";
  data.recordPracticeTestBasicMovement = [];
  data.recordPracticeTestMainRoad = [];
  
  util.nextTickN(2, () => {
    emit("refreshGridAssesment");
  });
}

function handleClickTabs(params) {
  data.activeTab = params

  // if (params == "Practice") {
  //   util.nextTickN(2, () => {
  //     getDataPractice();
  //   })
  // }
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
      TrainingCenterID: props.selectedAssesmentDriver.TrainingCenterID,
      TemplateTestID: props.selectedTemplate.TemplateID,
      EmployeeID: props.selectedAssesmentDriver.EmployeeID,
      Details: answerDetails,
    }
    
    axios.post("/hcm/tdc/assign-essay-score", param).then(
      (r) => {
        const paramSubmit = {
          ParticipantID: props.selectedAssesmentDriver._id,
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
    const dtBasicMovement = data.recordPracticeTestBasicMovement;
    const dtMainRoad = data.recordPracticeTestMainRoad;
    console.log("dtBasicMovement : ", dtBasicMovement)
    console.log("dtMainRoad : ", dtMainRoad)

    if (dtBasicMovement.length > 0) {
      console.log("Masuk dtBasicMovement")
      const params = {
        _id: data.practiceTestBasicMovement?._id || "",
        Date: data.practiceTestBasicMovement?.Date || new Date(),
        TrainingCenterID: data.practiceTestBasicMovement?.TrainingCenterID || props.selectedAssesmentDriver.TrainingCenterID,
        EmployeeID: data.practiceTestBasicMovement?.EmployeeID || props.selectedAssesmentDriver.EmployeeID,
        Type: "basic-movement",
        TemplateID: data.practiceTestBasicMovement?.TemplateID || data.practiceItemTemplateBasicMovement || "",
        FinalScore: finalScoreBasicMovement.value || 0,
        Detail: dtBasicMovement,
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
    if (dtMainRoad.length > 0) {
      console.log("Masuk dtMainRoad")
      const params = {
        _id: data.practiceTestMainRoad?._id || "",
        Date: data.practiceTestMainRoad?.Date || new Date(),
        TrainingCenterID: data.practiceTestMainRoad?.TrainingCenterID || props.selectedAssesmentDriver.TrainingCenterID,
        EmployeeID: data.practiceTestMainRoad?.EmployeeID || props.selectedAssesmentDriver.EmployeeID,
        Type: "main-road",
        TemplateID: data.practiceTestMainRoad?.TemplateID || data.practiceItemTemplateMainRoad || "",
        FinalScore: finalScoreMainRoad.value || 0,
        Detail: dtMainRoad,
      }
      
      const url = params?._id !== "" ? "/hcm/tdcpracticescore/update" :  "/hcm/tdcpracticescore/save"
      axios.post(url, params).then(
        (r) => {
          util.nextTickN(2, () => {
            getDataPractice();
          })
        },
        (e) => {
          util.showError(e);
        }
      )
    }
  }
}

function onHandleClickTabPractice(params) {
  data.activeTabPractice = params
}

function getDataPractice() {
  const params = {
    Select: ["_id", "Date", "TrainingCenterID", "EmployeeID", "Type", "TemplateID", "FinalScore", "Detail"],
    Where: {
      Op: "$and",
      items: [
        { Op: "$eq", Field: "TrainingCenterID", Value: props.selectedAssesmentDriver.TrainingCenterID },
        { Op: "$eq", Field: "EmployeeID", Value: props.selectedAssesmentDriver.EmployeeID },
      ]
    }
  }
  axios.post(`/hcm/tdcpracticescore/find`, params).then(
    (r) => {
      const dt = r.data
      const dtBasicMovement = dt.find(e => {return  e.Type == "basic-movement"});
      const dtMainRoad = dt.find(e => {return  e.Type == "main-road"});

      const dtTemplate = {
        _id: "",
        Date: new Date(),
        TrainingCenterID: props.selectedAssesmentDriver.TrainingCenterID,
        EmployeeID: props.selectedAssesmentDriver.EmployeeID,
        Type: "",
        TemplateID: "",
        FinalScore: 0,
        Detail: [],
      }
      //basic movement
      data.recordPracticeTestBasicMovement = dtBasicMovement ? dtBasicMovement.Detail : [];
      data.practiceTestBasicMovement = dtBasicMovement ? dtBasicMovement : dtTemplate;
      data.practiceTestBasicMovement.Type = "basic-movement";
      data.practiceTestBasicMovement.TemplateID = dtBasicMovement ? dtBasicMovement.TemplateID : "";

      //main road & ddt
      data.recordPracticeTestMainRoad = dtMainRoad ? dtMainRoad.Detail : [];
      data.practiceTestMainRoad = dtMainRoad ? dtMainRoad : dtTemplate;
      data.practiceTestMainRoad.Type = "main-road";
      data.practiceTestMainRoad.TemplateID = dtMainRoad ? dtMainRoad.TemplateID : "";
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