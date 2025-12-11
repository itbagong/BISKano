<template>
  <s-modal title="Rejection" display hideTitle hideButtons hideClose>
    <template #title>
      <div class="p-4 pb-0">
        <h1 class="text-primary border-b-[1px] flex justify-between">
          Rejection
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
      <div>Are you sure you want to reject this request?</div>

      <div class="py-3 mb-3">
        <s-input
          ref="reasonReject"
          v-model="data.reasonForRejection"
          label="Reason for Rejection"
          class="w-full"
          multiRow="5"
          :required="true"
          :rules="reason"
          :keepErrorSection="true"
        ></s-input>
      </div>
      <template #footer>
        <div class="mt-5">
          <s-button
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Submit"
            @click="changeRejection"
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
const reasonReject = ref(null);
const props = defineProps({
  modelValue: { type: String, default: () => "" },
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
  reasonForRejection: props.modelValue,
});

function close() {
  emit("close");
}

function changeRejection() {
  if (reasonReject.value.validate()) {
    emit("changeRejection", data.reasonForRejection);
  }
}

onMounted(() => {});
</script>
