<template>
  <s-modal title="Rejection" display hideTitle hideButtons hideClose>
    <template #title>
      <div class="p-4 pb-0">
        <h1 class="text-primary border-b-[1px] flex justify-between">
          Add New Daily Report
          <a href="#" class="delete_close" @click="close()">
            <mdicon
              name="close"
              width="16"
              alt="close"
              class="cursor-pointer hover:text-primary"
            />
          </a>
        </h1>
      </div>
    </template>
    <s-card hide-title class="min-w-[400px]">
      <div>
        <s-input
          ref="refInputDate"
          v-model="data.report.WorkDate"
          kind="date"
          label="Work Date"
          class="w-full"
          :required="true"
          :keepErrorSection="true"
        ></s-input>
      </div>
      <template #footer>
        <div class="mt-5">
          <s-button
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Add New"
            @click="changeWork"
          ></s-button>
        </div>
      </template>
    </s-card>
  </s-modal>
</template>
<script setup>
import { onMounted, reactive, inject, ref } from "vue";
import { SCard, util, SModal, SButton, SInput } from "suimjs";
import { authStore } from "@/stores/auth";

const auth = authStore();
const axios = inject("axios");
const refInputTitle = ref(null);
const refInputDate = ref(null);
const props = defineProps({
  modelValue: {
    type: Object,
    defaule: {
      WorkDate: new Date(),
    },
  },
});
const emit = defineEmits({
  "update:modelValue": null,
  close: null,
});
const data = reactive({
  report: props.modelValue,
});

function close() {
  emit("close");
}

function changeWork() {
  emit("changeWork", data.report);
}

onMounted(() => {});
</script>
