<template>
  <div class="w-full">
    <data-list
      ref="listControl"
      form-config="/she/mcutransaction/followup/formconfig"
      form-read="/she/mcutransaction/followup/get"
      form-keep-label
      init-app-mode="form"
      stay-on-form-after-save
      form-hide-buttons
      hide-title
      no-gap
      :form-fields="['DetailPemeriksaan', 'Document']"
      @alter-form-config="onAlterCfgForm"
    >
      <template #form_input_Document="{ item, config }">
        <uploader
          ref="gridAttachment"
          :journalId="jurnalId"
          :config="config"
          journalType="SHE_MCU_FOLLOW_UP"
          single-save
        />
      </template>
      <template #form_input_DetailPemeriksaan="{ item }">
        <div class="flex justify-between mb-2">
          <label class="input_label mt-4">
            <div>Result</div>
          </label>
        </div>

        <div class="card w-full shadow-none">
          <div
            class="grid grid-cols-5 text-sm font-semibold gap-2 border p-2 bg-slate-50"
          >
            <div v-for="(dt, idx) in data.headerDetail" :key="idx">
              {{ dt }}
            </div>
          </div>
          <div class="overflow-y-auto h-96 relative">
            <div v-for="(dt, idx) in item.DetailPemeriksaan" :key="idx">
              <detail-checkup v-model="item.DetailPemeriksaan[idx]" />
            </div>
          </div>
        </div>
      </template>
    </data-list>
  </div>
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

import { SInput, util, SButton, SGrid, SCard, DataList, SModal } from "suimjs";
import Uploader from "@/components/common/Uploader.vue";
import detailCheckup from "./McuCheckupDetail.vue";

const listControl = ref(null);
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  jurnalId: { type: String, default: () => "" },
});

const data = reactive({
  modal: {
    value: false,
    additionalItem: [],
    specialist: "",
  },
  record: props.modelValue,
  headerDetail: [
    "Detail Pemeriksaan",
    "Hasil",
    "Nilai Rujukan",
    "Unit",
    "Note",
  ],
});

function onAlterCfgForm(cfg) {
  listControl.value.setFormRecord(data.record);
}
</script>
