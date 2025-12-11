<template>
  <div class="w-full">
    <div
      v-if="data.view === 'select'"
      class="flex flex-col justify-center items-center px-4 py-16"
    >
      <p class="text-lg font-semibold">Please Select Your Test</p>
      <div class="my-8 grid grid-cols-3 gap-3">
        <template v-for="test in data.candidateTest">
          <div
            class="px-4 py-8 bg-white border rounded-lg cursor-pointer"
            @click="chooseTest(test)"
          >
            <div class="flex items-center justify-center mb-2">
              <mdicon name="file" size="30"></mdicon>
            </div>
            <p class="text-md font-medium text-center">
              {{ test.TestName?.length > 0 ? test.TestName : "N/a" }}
            </p>
          </div>
        </template>
      </div>
    </div>

    <div v-if="data.view === 'instruction'">
      <div class="flex flex-col justify-center items-center my-16">
        <div class="w-[360px] border bg-white">
          <div class="bg-slate-800 p-2">
            <p class="text-white text-lg font-bold">Instructions</p>
          </div>
          <div class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p>Total Duration</p>
                <p class="font-bold">
                  {{ data.currentTest.Instruction.TotalDuration }} Min
                </p>
              </div>
              <div>
                <p>No. of Questions</p>
                <p class="font-bold">
                  {{ data.currentTest.Instruction.NoOfQuestions }} Question
                </p>
              </div>
            </div>
            <div class="mt-5">
              <div v-html="formattedInstruction"></div>
              <loader
                v-if="data.loadingAttach"
                kind="skeleton"
                skeleton-kind="circle"
              />

              <div
                class="flex flex-col gap-5 mt-5"
                v-else
                v-for="atc in data.instructionAttach"
              >
                <button @click="onPreviewImg(helper.getAssetUrl(atc._id))">
                  <img :src="helper.getAssetUrl(atc._id)" class="w-full" />
                </button>
              </div>
              <div class="flex items-center justify-center mt-4 mb-2">
                <s-button
                  class="bg-yellow-400 border-2 border-yellow-400 text-slate-800 rounded-md"
                  label="Start Test"
                  tooltip="Start Test"
                  @click="getQuestions"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="data.view === 'test'" class="p-4 bg-white border rounded-lg">
      <div class="flex items-center justify-between border-b-2 pb-5">
        <p class="font-bold">Finish Your Test</p>
        <div class="border-2 flex items-center gap-2 p-1">
          <mdicon name="clock-outline" size="18"></mdicon>
          <p class="text-red-500 font-semibold">{{ formattedTime }}</p>
        </div>
      </div>
      <div class="my-5">
        <loader v-if="data.loading" kind="skeleton" skeleton-kind="list" />
        <template
          v-else
          v-for="(question, idx) in data.dataQuestions.displayQuestions"
        >
          <div class="mb-6">
            <!-- Question -->
            <div class="flex items-start gap-1 mb-2">
              <p>{{ question.Number }}.</p>
              <p>{{ question.Description }}</p>
            </div>
            
            <!-- TDC -->
            <template v-if="data.currentTest.TestScheduleType == 'TDC'">
              <image-candidate-test
                v-model="question.ID"
                journal-type="MCU_ITEM_TEMPLATE_LINES"
              />
              <!-- Answer Multiple -->
              <div v-show="question.Condition?.length > 0" class="bg-slate-100 mb-2 p-1">
                <div v-for="condition in question.Condition" :key="condition.ID" class="flex items-center gap-2">
                  <input
                    :id="question.Number + '_' + question.Description + '_' + condition.ID"
                    type="radio"
                    class="bg-slate-800"
                    :name="'question_' + question.ID"
                    :value="condition.ID"
                    @change="() => multipleAnswerHandleChange(condition, question.ID)"
                    :checked="data.userAnswers.find(a => a.QuestionID === question.ID)?.AnswerID === condition.ID"
                  />
                  <div>
                    <p>{{ condition.Name }}</p>
                    <image-candidate-test
                      v-model="question.ID"
                      journal-type="MCU_ITEM_TEMPLATE_LINES_LIST"
                    />
                  </div>
                </div>
              </div>
              <!-- Answer Essay -->
              <div
                v-show="question.Condition?.length === 0"
              >
                <textarea 
                  rows="4" 
                  cols="100" 
                  class="border-black border p-3"
                  @input="(event) => essayAnswerHandleChange(event.target.value, question.ID)"
                ></textarea>
              </div>
            </template>
            <template v-else>
              <image-candidate-test
                v-model="question.ID"
                journal-type="MCU_ITEM_TEMPLATE_LINES"
              />
              <div
                v-for="condition in question.Condition"
                class="bg-slate-100 mb-2 p-1"
              >
                <div class="flex items-center gap-2">
                  <input
                    type="radio"
                    class="bg-slate-800"
                    :value="{
                      ...condition,
                      QuestionID: question.ID,
                      IsMostQuestionType: question.QuestionTypeIsMost,
                    }"
                    v-model="data.userAnswers[idx]"
                  />
                  <div>
                    <p>{{ condition.Name }}</p>
                    <image-candidate-test
                      v-model="question.ID"
                      journal-type="MCU_ITEM_TEMPLATE_LINES_LIST"
                    />
                  </div>
                </div>
              </div>
            </template>
          </div>
        </template>
      </div>
      <div class="flex items-center justify-between">
        <s-button
          class="bg-white border-2 border-red-500 text-red-500 rounded-md"
          label="Clear Answer"
          tooltip="Clear Answer"
          @click="clearAnswers"
        />
        <div class="flex items-center gap-2">
          <s-button
            :disabled="data.pagination.Skip === 0"
            class="bg-white border-2 border-slate-500 text-slate-500 rounded-md"
            label="Previous Question"
            tooltip="Previous Question"
            @click="previousQuestions"
          />
          <s-button
            :disabled="isLastBatch"
            class="bg-white border-2 border-slate-500 text-slate-500 rounded-md"
            label="Next Question"
            tooltip="Next Question"
            @click="nextQuestion"
          />
          <s-button
            class="bg-yellow-400 border-2 border-yellow-400 text-slate-800 rounded-md"
            label="Finish Test"
            tooltip="Finish Test"
            @click="finishTest"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// import { authStore } from "@/stores/auth";
import { reactive, inject, onMounted, computed, onUnmounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { util, SButton, SInput } from "suimjs";
import Loader from "@/components/common/Loader.vue";
import helper from "@/scripts/helper.js";
import { api as viewerApi } from "v-viewer";
import ImageCandidateTest from "./widget/ImageCandidateTest.vue";

layoutStore().name = "tenant";

// const FEATUREID = "TalentDevelopmentSubmission";
// const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");

const data = reactive({
  loading: false,
  view: "select",
  candidateTest: [],
  loadingAttach: false,
  selectedAttach: false,
  instructionAttach: [],
  currentTest: {},
  timerData: {
    totalTime: 0,
    timeRemaining: 0,
    intervalId: null,
    isRunning: false,
    isTimerSet: false,
  },
  dataQuestions: {
    total: 0,
    questions: [],
    displayQuestions: [],
  },
  userAnswers: [],
  pagination: {
    Skip: 0,
    Take: 2,
  },
});

const findCandidateTest = () => {
  const url = "/hcm/testschedule/find-candidate-test";
  const payload = {
    Date: new Date().toISOString(),
  };

  axios.post(url, payload).then(
    (r) => {
      data.candidateTest = r.data;
    },
    (e) => {
      util.showError(e);
    }
  );
};
const getInstructionAttach = (id) => {
  data.loadingAttach = true;
  axios
    .post("/asset/read-by-journal", {
      JournalType: "MCU_ITEM_TEMPLATE_INSTRUCTIONS",
      JournalID: id,
    })
    .then(
      (r) => {
        data.instructionAttach = r.data;
        data.loadingAttach = false;
      },
      (e) => {
        util.showError(e);
        data.loadingAttach = false;
      }
    )
    .catch((e) => {
      util.showError(e);
      data.previewLoading = false;
    });
};
const chooseTest = (test) => {
  data.currentTest = test;
  data.view = "instruction";
  getInstructionAttach(test.TemplateID);
};

function onPreviewImg(uri) {
  viewerApi({
    images: [uri],
  });
}
const formattedInstruction = computed(() => {
  return data.currentTest.Instruction.Instruction.replace(/\n/g, "<br>");
});

const getQuestions = () => {
  data.view = "test";
  data.loading = true;

  const test = data.currentTest;
  const url = "/hcm/testschedule/get-questions";
  const { TemplateID, TestID, TestScheduleType } = test;
  const payload = {
    TemplateID,
    TestID,
    TestScheduleType,
    ...data.pagination,
  };

  axios.post(url, payload).then(
    (r) => {
      const totalQuestions = r.data.count;
      const fetchedQuestions = r.data.data;

      data.dataQuestions.total = totalQuestions;

      if (data.pagination.Skip >= data.dataQuestions.questions.length) {
        data.dataQuestions.questions.push(...fetchedQuestions);
      }

      data.dataQuestions.displayQuestions = data.dataQuestions.questions.slice(
        data.pagination.Skip,
        data.pagination.Skip + data.pagination.Take
      );

      if (data.pagination.Skip === 0) {
        setTimer(test.Instruction.TotalDuration);
        startCountdown();
      }

      if(TestScheduleType == "TDC") {
        const userAnswersMap = fetchedQuestions.map(e => {
          const map = {
            QuestionID: e.ID,
            AnswerID: "",
            AnswerValue: "",
          }
          return map
        })
        data.userAnswers = userAnswersMap
      }

      setTimeout(() => {
        data.loading = false;
      }, 500);
    },
    (e) => {
      util.showError(e);
      data.loading = false;
    }
  );
};

const formattedTime = computed(() => {
  const hours = String(
    Math.floor(data.timerData.timeRemaining / 3600)
  ).padStart(2, "0");
  const minutes = String(
    Math.floor((data.timerData.timeRemaining % 3600) / 60)
  ).padStart(2, "0");
  const seconds = String(data.timerData.timeRemaining % 60).padStart(2, "0");
  return `${hours}:${minutes}:${seconds}`;
});

const setTimer = (duration) => {
  data.timerData.totalTime = duration * 60;
  data.timerData.timeRemaining = data.timerData.totalTime;
  data.timerData.isTimerSet = true;
};

const startCountdown = () => {
  if (data.timerData.intervalId || data.timerData.timeRemaining <= 0) return;

  data.timerData.isRunning = true;
  data.timerData.intervalId = setInterval(() => {
    if (data.timerData.timeRemaining > 0) {
      data.timerData.timeRemaining--;
    } else {
      clearInterval(data.timerData.intervalId);
      data.timerData.intervalId = null;
      data.timerData.isRunning = false;

      // save answer with submit
      saveAnswers(true);
    }
  }, 1000);
};

const clearTimer = () => {
  clearInterval(data.timerData.intervalId);
  data.timerData = {
    totalTime: 0,
    timeRemaining: 0,
    intervalId: null,
    isRunning: false,
    isTimerSet: false,
  };
};

const clearAnswers = () => {
  if(data.currentTest.TestScheduleType == "TDC") {
    const userAnswersMap = data.dataQuestions.questions?.map(e => {
      const map = {
        QuestionID: e.ID,
        AnswerID: "",
        AnswerValue: "",
      }
      return map
    })
    data.userAnswers = userAnswersMap
  } else {
    data.userAnswers = [];
  }
};

const saveAnswers = (withSubmit) => {
  const { TestID, TemplateID, TestScheduleType } = data.currentTest;
  let payload = {}
  if(TestScheduleType == "TDC") {
    payload = {
      TrainingCenterID: TestID,
      TemplateTestID: TemplateID,
      Details: data.userAnswers.map((answer) => ({
        QuestionID: answer.QuestionID,
        AnswerID: answer.AnswerID,
        AnswerValue: answer.AnswerValue,
      })),
    };
  } else {
    payload = {
      JobID: TestID,
      TemplateTestID: TemplateID,
      TestType: TestScheduleType,
      Details: data.userAnswers.map((answer) => ({
        QuestionID: answer.QuestionID,
        AnswerID: answer.ID,
        AnswerValue: answer.Letter,
        IsMostQuestionType: answer.IsMostQuestionType,
      })),
    };
  }

  const url = TestScheduleType == "TDC" ? "/hcm/tdc/save-answer" : "/hcm/psychologicaltest/save-answer";

  axios.post(url, payload).then(
    (r) => {
      if (!withSubmit) {
        // give loading
        data.loading = true;
        setTimeout(() => {
          data.loading = false;
        }, 500);

        data.pagination.Skip += data.pagination.Take;
        getQuestions(data.currentTest);
      } else {
        submitTest();
      }
      data.userAnswers = [];
    },
    (e) => {
      util.showError(e);
    }
  );
};

const submitTest = () => {
  const { TestID, TemplateID, TestScheduleType } = data.currentTest;
  let payload = {}

  if(TestScheduleType == "TDC") {
    payload = {
      TrainingCenterID: TestID,
      TemplateTestID: TemplateID
    }
  } else {
    payload = {
      JobID: TestID,
      TemplateTestID: TemplateID,
      TestType: TestScheduleType,
    };
  }
  
  const url = TestScheduleType == "TDC" ? "/hcm/tdc/submit" : "/hcm/psychologicaltest/submit";

  axios.post(url, payload).then(
    (r) => {
      clearTimer();
      findCandidateTest();
      data.view = "select";
      data.dataQuestions = {
        total: 0,
        questions: [],
        displayQuestions: [],
      };
      data.pagination = {
        Skip: 0,
        Take: 2,
      };
    },
    (e) => {
      util.showError(e);
    }
  );
};

const nextQuestion = () => {
  if (isLastBatch.value) return;
  // save answer
  saveAnswers();
};

const previousQuestions = () => {
  if (data.pagination.Skip > 0) {
    // give loading
    data.loading = true;
    setTimeout(() => {
      data.loading = false;
    }, 500);

    data.pagination.Skip -= data.pagination.Take;

    data.dataQuestions.displayQuestions = data.dataQuestions.questions.slice(
      data.pagination.Skip,
      data.pagination.Skip + data.pagination.Take
    );

    data.dataQuestions.questions.splice(
      data.pagination.Skip + data.pagination.Take
    );
  }
};

const isLastBatch = computed(
  () => data.pagination.Skip + data.pagination.Take >= data.dataQuestions.total
);

const multipleAnswerHandleChange = (answer, questionId) => {
  data.userAnswers = data.userAnswers.map(answerItem => {
    if (answerItem.QuestionID === questionId) {
      return { ...answerItem, AnswerID: answer.ID };
    }
    return answerItem;
  });
};

const essayAnswerHandleChange = (answerText, questionId) => {
  data.userAnswers = data.userAnswers.map(answerItem => {
    if (answerItem.QuestionID === questionId) {
      return { ...answerItem, AnswerValue: answerText };
    }
    return answerItem;
  });
};

const finishTest = () => {
  // save answer with submit
  saveAnswers(true);
};

onMounted(() => {
  findCandidateTest();
});

onUnmounted(() => {
  clearTimer();
});
</script>
