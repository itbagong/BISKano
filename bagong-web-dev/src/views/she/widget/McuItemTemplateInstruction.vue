<template>
  <data-list
    ref="listControl"
    hide-title
    form-config="/she/mcuitemtemplate/instruction/formconfig"
    form-keep-label
    :init-app-mode="data.appMode"
    form-hide-submit
    form-hide-cancel
    :form-fields="['Attachment']"
  >
    <template #form_input_Attachment="{ item }">
      <uploader
        :journalId="jurnalId"
        :config="{ label: 'Attachment' }"
        journalType="MCU_ITEM_TEMPLATE_INSTRUCTIONS"
        is-single-upload
        single-save
      />
    </template>
  </data-list>
</template>
<script setup>
import {
  reactive,
  ref,
  inject,
  watch,
  computed,
  onMounted,
  nextTick,
} from "vue";
import {
  SInput,
  util,
  SButton,
  SGrid,
  SCard,
  DataList,
  loadGridConfig,
} from "suimjs";
import Uploader from "@/components/common/Uploader.vue";

const listControl = ref(null);

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  jurnalId: { type: String, default: () => "" },
});

const data = reactive({
  appMode: "form",
  record: props.modelValue ?? {},
});

onMounted(() => {
  listControl.value.setFormRecord(data.record);
});
</script>
