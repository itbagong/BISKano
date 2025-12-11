<template>
  <s-modal
    :display="modelValue"
    title="Send Psikotest"
    @beforeHide="beforeHide"
    @submit="submit"
    :hideSubmit="hideSubmit"
  >
    <s-card class="rounded-md w-full" hide-title>
      <div class="px-2 w-[440px]">
        <s-input
          v-model="data.ids"
          class="w-full"
          label="Select Question"
          use-list
          lookup-url="/she/mcuitemtemplate/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['_id', 'Name']"
          multiple
        />
      </div>
    </s-card>
  </s-modal>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";

import { util, SInput, SModal } from "suimjs";
const props = defineProps({
  modelValue: { type: Boolean, default: false },
});
const data = reactive({
  ids: [],
});
const emit = defineEmits({
  submit: null,
  "update:modelValue": null,
});
const hideSubmit = computed({
  get() {
    return data.ids.length == 0;
  },
});
function beforeHide() {
  data.ids = [];
  emit("update:modelValue", false);
}
function submit() {
  emit("submit", data.ids);
  data.ids = [];
  emit("update:modelValue", false);
}
</script>
