<template>
  <data-list
    class="card"
    ref="listControl"
    hide-title
    no-gap
    form-config="/she/sidak/drug/formconfig"
    form-keep-label
    :init-app-mode="data.appMode"
    form-hide-submit
    form-hide-cancel
    @alter-form-config="onAlterFormConfig"
    :form-fields="['Sign', 'Evidance']"
  >
    <template #form_input_Sign="{ item, config }">
      <div class="hidden">{{ (config.label = "Drugs Sign") }}</div>
      <uploader
        ref="gridAttachmentSign"
        :journalId="jurnalId"
        :config="config"
        journalType="SHE_DRUGS_SIGN"
        :key="5"
        single-save
      />
    </template>
    <template #form_input_Evidance="{ item, config }">
      <div class="hidden">{{ (config.label = "Drugs Evidance") }}</div>
      <uploader
        ref="gridAttachmentEvidance"
        :journalId="jurnalId"
        :config="config"
        journalType="SHE_DRUGS_EVIDANCE"
        :key="6"
        single-save
      />
    </template>
  </data-list>
</template>

<script setup>
import { reactive, ref, inject, watch, computed, onMounted } from "vue";
import { DataList, SInput, util, SButton } from "suimjs";
import moment from "moment";
import Uploader from "@/components/common/Uploader.vue";

const gridAttachmentSign = ref(null);
const gridAttachmentEvidance = ref(null);
const listControl = ref(null);

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  jurnalId: { type: String, default: "" },
});

const data = reactive({
  appMode: "form",
  record: {},
});

const emit = defineEmits({
  "update:modelValue": null,
});

function onAlterFormConfig(config) {
  data.record = props.modelValue ?? {};
  data.record.Sign = data.record.Sign ?? [];
  data.record.Evidance = data.record.Evidance ?? [];
  listControl.value.setFormRecord(data.record);
}

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
