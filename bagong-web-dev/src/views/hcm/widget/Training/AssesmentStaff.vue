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
                  @click="saveAssesmentStaff()"
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
                <!-- Essay -->
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
        <div class="w-full card  p-6">
          <s-input
              ref="refCustomer"
              v-model="data.practiceItemTemplate"
              lookup-key="_id"
              label="Item Template"
              class="w-full"
              use-list
              :lookup-url="`/she/mcuitemtemplate/find`"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              @change="filterPracticeChanged"
          ></s-input>
          <s-grid
              class="gridPracticeAssesmentStaff"
              ref="gridPracticeAssesmentStaff"
              hide-select
              hide-search
              hide-sort
              hide-refresh-button
              hide-new-button
              hide-delete-button
              hide-action
              hide-footer
              :config="data.cfgPractice"
              :grid-fields="['AssementName', 'MaxScore', 'AchievedScore', 'Note']"
          >
            <template #item_AssementName="{ item }">
              <div 
                  :class="{
                  total_score_title: item.AssementName == 'Total Score',
                  }"
              >{{ item.AssementName }}</div>
            </template>
            <template #item_MaxScore="{ item }">
              <div 
                  :class="{
                  total_score_title: item.AssementName == 'Total Score',
                  }"
              >{{ item.MaxScore }}</div>
            </template>
            <template #item_AchievedScore="{ item }">
              <s-input
                  v-if="item.AssementName !== 'Total Score'"
                  ref="refAchievedScore"
                  hide-label
                  v-model="item.AchievedScore"
                  class="w-100"
                  kind="number"
                  @change="
                  (field, v1, v2, old, ctlRef) => {
                      onGridRowFieldChanged('AchievedScore', v1, v2, old, item);
                  }
                  "
              ></s-input>
              <div v-else class="total_score">
                  {{ item.AchievedScore }}
              </div>
            </template>
            <template #item_Note="{ item }">
              <s-input
                  v-if="item.AssementName !== 'Total Score'"
                  ref="refNote"
                  hide-label
                  v-model="item.Note"
                  class="w-100"
                  kind="text"
                  @change="
                  (field, v1, v2, old, ctlRef) => {
                      onGridRowFieldChanged('Note', v1, v2, old, item);
                  }
                  "
              ></s-input>
              <div v-else class="total_score">
                  {{ item.Note }}
              </div>
            </template>
          </s-grid>
        </div>
      </div>
    </div>
</template>

<script setup>
import { reactive, ref, inject, computed } from "vue";
import { util, SButton, SGrid, SInput } from "suimjs";
import ImageCandidateTest from "../ImageCandidateTest.vue";

const axios = inject("axios");

const gridPracticeAssesmentStaff = ref(null);

const props = defineProps({
    view: {type: String, default: () => ""},
    selectedAssesmentStaff: {type: Object, default: () => {}},
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
    selectedPracticeItemTemplate: {},
    recordAssesmentStaffPractice: {},
})

function backAssesmentStaff() {
  emit("update:view", "tdc");
  data.activeTab = "Written Test";
  data.practiceItemTemplate = "";
  data.selectedPracticeItemTemplate = {};
  data.recordAssesmentStaffPractice = {};
  
  util.nextTickN(2, () => {
    emit("refreshGridAssesment");
  });
}

function handleClickTabs(params) {
  data.activeTab = params
  util.nextTickN(2, () => {
    if(params === 'Practice') {
      genFromCfgPractice()
    }
  });
}

function filterPracticeChanged(name, v1, v2) {
  const param = [v1]
  gridPracticeAssesmentStaff.value.setLoading(true);
  axios.post("/she/mcuitemtemplate/get", param).then(
    (r) => {
      data.selectedPracticeItemTemplate = r.data
      gridPracticeAssesmentStaff.value.setLoading(false);

      genFromCfgPractice()
    },
    (e) => {
      util.showError(e);
      gridPracticeAssesmentStaff.value.setLoading(false);
    }
  );
}

function onHandleClickStage(param) {
    emit("handleClickStage", param);
}

function onHandleClickTemplate(param) {
    emit("handleClickTemplate", param);
}

//Assesment Staff Practice 
function genFromCfgPractice() {
  data.cfgPractice = {
    fields: [
      {
        field: "AssementName",
        kind: "text",
        label: "Name",
        labelField: "",
        readType: "show",
        input: {
          field: "AssementName",
          kind: "text",
          label: "Name",
          lookupUrl: "",
          placeHolder: "Name",
        } 
      },
      {
        field: "MaxScore",
        kind: "text",
        label: "Max Score",
        labelField: "",
        readType: "show",
        input: {
          field: "MaxScore",
          kind: "text",
          label: "Max Score",
          lookupUrl: "",
          placeHolder: "Max Score",
        } 
      },
      {
        field: "AchievedScore",
        kind: "number",
        label: "Achieved Score",
        labelField: "",
        readType: "show",
        input: {
          field: "AchievedScore",
          kind: "number",
          label: "Achieved Score",
          lookupUrl: "",
          placeHolder: "Achieved Score",
        } 
      },
      {
        field: "Note",
        kind: "text",
        label: "Note",
        labelField: "",
        readType: "show",
        input: {
          field: "Note",
          kind: "text",
          label: "Note",
          lookupUrl: "",
          placeHolder: "Note",
        } 
      },
    ],
    setting: {
      idField: "",
      keywordFields: ['_id', 'Name'],
      sortable: ['_id']
    }
  }

  gridPracticeAssesmentStaff.value.setLoading(true); 
  
  const option = [
    {Op: "$eq", Field: "ParticipantID", Value: props.selectedAssesmentStaff._id},
  ]

  if (data.practiceItemTemplate !== "") {
    option.push({Op: "$eq", Field: "TemplateID", Value: data.practiceItemTemplate})
  }

  const param = {
    Select: ["_id", "Details", "ParticipantID", "TemplateID", "TotalScore"],
    Sort: ["-LastUpdate"],
    Take: 20,
    Where: {
      Op: "$and", items: option
    },
  }

  axios.post(`/hcm/tdcpracticestaff/find`, param).then(
    (r) => {
      const result = r.data.length > 0 ? r.data[0] : {}
      const mapData = []
      const dtLines = data.selectedPracticeItemTemplate.Lines || []

      let maxScoreTotal = 0;
      let scoreTotal = 0;
      // console.log("result", result)
      // console.log("dtLines", dtLines)
      if(dtLines.length > 0) {
        dtLines.forEach(e => {
          const dtDetails = result.Details?.find(o => {return o.AssementName == e.Description})
          maxScoreTotal = maxScoreTotal + e.AnswerValue
          scoreTotal = scoreTotal + (dtDetails?.AchievedScore || 0)

          if(dtDetails?.AssementName !== "Total Score"){
            mapData.push({
              AssementName: e.Description,
              MaxScore: e.AnswerValue,
              AchievedScore: dtDetails?.AchievedScore || 0,
              Note: dtDetails?.Note || ""
            })
          } else {
            mapData.push({
              AssementName: e.Description,
              MaxScore: e.AnswerValue,
              AchievedScore: 0,
              Note: ""
            })
          }
        });
      } else if(result._id) {
        result.Details.forEach(e => {
          maxScoreTotal = maxScoreTotal + e.MaxScore
          scoreTotal = scoreTotal + e.AchievedScore

          if(e.AssementName !== "Total Score"){
            mapData.push({
              AssementName: e.AssementName,
              MaxScore: e.MaxScore,
              AchievedScore: e.AchievedScore,
              Note: e.Note
            })
          }
        })
      }

      mapData.push({
        AssementName: "Total Score",
        MaxScore: maxScoreTotal,
        AchievedScore: scoreTotal,
        Note: ""
      })

      data.recordAssesmentStaffPractice = {
        ParticipantID: result.ParticipantID,
        TemplateID: result.TemplateID,
        Details: mapData,
        TotalScore: 0,
        _id: result._id,
      }

      gridPracticeAssesmentStaff.value.setLoading(false);
      gridPracticeAssesmentStaff.value.setRecords(mapData);
    },
    (e) => {
      util.showError(e);
      gridPracticeAssesmentStaff.value.setLoading(false);
    }
  );
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
  gridPracticeAssesmentStaff.value.setRecord(
		gridPracticeAssesmentStaff.value.getCurrentIndex(),
    record,
	);
}

function saveAssesmentStaff() {
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
      TrainingCenterID: props.selectedAssesmentStaff.TrainingCenterID,
      TemplateTestID: props.selectedTemplate.TemplateID,
      EmployeeID: props.selectedAssesmentStaff.EmployeeID,
      Details: answerDetails,
    }
    
    axios.post("/hcm/tdc/assign-essay-score", param).then(
      (r) => {
        const paramSubmit = {
          ParticipantID: props.selectedAssesmentStaff._id,
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
    const dtDetails = data.recordAssesmentStaffPractice?.Details || gridPracticeAssesmentStaff.value.getRecords()
    
    const filteredDetails = dtDetails.filter(e => e.AssementName !== "Total Score")
    // console.log("data.recordAssesmentStaffPractice: ", data.recordAssesmentStaffPractice)
    // console.log("filteredDetails: ", filteredDetails)

    let TotalScore = 0;
    filteredDetails.forEach(e => {
      TotalScore = TotalScore + e.AchievedScore;
    })
    // console.log("TotalScore: ", TotalScore)

    const param = {
      ParticipantID: data.recordAssesmentStaffPractice?.ParticipantID || props.selectedAssesmentStaff._id,
      TemplateID: data.recordAssesmentStaffPractice?.TemplateID || data.selectedPracticeItemTemplate._id,
      Details: filteredDetails,
      TotalScore: TotalScore || 0,
      _id: data.recordAssesmentStaffPractice?._id || null,
    }

    const url = data.recordAssesmentStaffPractice?._id ? "/hcm/tdcpracticestaff/update" :  "/hcm/tdcpracticestaff/save"
    axios.post(url, param).then(
      (r) => {
        util.nextTickN(2, () => {
          genFromCfgPractice()
        })
      },
      (e) => {}
    )
  }
}
//End Assesment Staff Practice
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