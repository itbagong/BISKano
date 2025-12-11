<template>
  <s-modal title="Rejection" display hideTitle hideButtons hideClose>
    <template #title>
      <div class="p-4 pb-0">
        <h1 class="text-primary border-b-[1px] flex justify-between">
          Add New Report
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
        <!-- lookup-url="/mfg/bom/find"
          lookup-key="_id"
          :lookup-labels="['Title']"
          :lookup-search="['_id', 'Title']" -->
        {{ data.wo.WorkDescription }}
        <s-input
          ref="refInputTitle"
          v-model="data.WorkTitle"
          label="Work Title"
          class="w-full"
          useList
          :items="data.list"
          :required="true"
          :keepErrorSection="true"
        ></s-input>
        <s-input
          ref="refInputDate"
          v-model="data.wo.WorkDate"
          kind="date"
          label="Date"
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
      WorkOrderJournalID: "",
      WorkDescriptionNo: 0,
      WorkDescription: "",
      WorkDate: new Date(),
      ItemUsage: [],
      ManpowerUsage: [],
      Output: [],
    },
  },
  listWork: { type: Array, default: () => [] },
});
const reason = [
  (v) => {
    return v.length == 0 ? "reason for rejection is required" : "";
  },
];
const emit = defineEmits({
  "update:modelValue": null,
  close: null,
});
const data = reactive({
  wo: props.modelValue,
  list: props.listWork,
  WorkTitle: "",
});

function close() {
  emit("close");
}

function changeWork() {
  if (refInputTitle.value.validate() && refInputDate.value.validate()) {
    data.wo.WorkDescriptionNo = data.WorkTitle.split(" | ").at(0);
    data.wo.WorkDescription = data.WorkTitle.split(" | ").at(1);
    emit("changeWork", data.wo);
  }
}

onMounted(() => {});
</script>
