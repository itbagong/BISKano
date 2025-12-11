<template>
  <s-modal title="Rejection" display hideTitle hideButtons hideClose>
    <template #title>
      <div class="p-4 pb-0">
        <h1 class="text-primary border-b-[1px] flex justify-between">
          Mark As Completed
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
          ref="refInputTitle"
          v-model="data.wo.ComponentCategory"
          label="WO Component Category"
          class="w-full"
          useList
          :required="true"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=WOComponentCategory`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          :keepErrorSection="true"
        ></s-input>
      </div>
      <template #footer>
        <div class="mt-5">
          <s-button
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Completed"
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
    defaule: {},
  },
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
  WorkTitle: "",
});

function close() {
  emit("close");
}

function changeWork() {
  emit("changeWork", data.wo);
}

onMounted(() => {});
</script>
