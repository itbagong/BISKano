<template>
  <data-list
    ref="listControl"
    form-config="/she/mcutransaction/result/formconfig"
    form-read="/she/mcutransaction/result/get"
    form-keep-label
    init-app-mode="form"
    stay-on-form-after-save
    form-hide-buttons
    hide-title
    no-gap
    :form-fields="[
      'VisitResult',
      'DetailPemeriksaan',
      'AdditionalItem',
      'DocumentMCU',
    ]"
  >
    <template #form_input_VisitResult="{ item }">
      <s-input
        use-list
        :items="['Fit', 'UnFit']"
        v-model="item.VisitResult"
        label="Visit Result"
        class="w-[200px]"
      />
    </template>
    <template #form_input_DetailPemeriksaan="{ item }">
      <label class="input_label">
        <div>Result</div>
      </label>
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
            <div v-for="(v, index) in dt.Lines" :key="v.ID">
              <detail-checkup
                v-model="dt.Lines[index]"
                v-if="v.IsSelected"
                :name="dt.Name"
                :items="dt.Lines"
                :gender="gender"
              />
            </div>
          </div>
        </div>
      </div>
    </template>
    <template #form_input_AdditionalItem="{ item }">
      <label class="input_label">
        <div>Additional Item</div>
      </label>
      <div class="card w-full shadow-none">
        <div
          class="grid grid-cols-5 text-sm font-semibold gap-2 p-2 border bg-slate-50"
        >
          <div v-for="(dt, idx) in data.headerDetail" :key="idx">
            {{ dt }}
          </div>
        </div>
        <div class="overflow-y-auto h-96 relative">
          <div v-for="(v, index) in data.additionalItem" :key="v.ID">
            <detail-checkup
              v-model="data.additionalItem[index]"
              :items="data.additionalItem"
              :gender="gender"
            />
          </div>
        </div>
      </div>
    </template>
    <template #form_input_DocumentMCU="{ item, config }">
      <uploader
        ref="gridAttachment"
        :journalId="jurnalId"
        :config="config"
        journalType="SHE_MCU_RESULT"
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

import { SInput, util, SButton, SGrid, SCard, DataList } from "suimjs";
import detailCheckup from "./McuCheckupDetail.vue";
import Uploader from "@/components/common/Uploader.vue";
const axios = inject("axios");
const listControl = ref(null);

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  mcuPaket: { type: String, default: () => "" },
  mcuAddItem: { type: Array, default: () => [] },
  jurnalId: { type: String, default: () => "" },
  gender: { type: String, default: () => "" },
});

const emit = defineEmits({
  "update:modelValue": null,
});
const data = reactive({
  record: props.modelValue,
  additionalItem: [],
  headerDetail: [
    "Detail Pemeriksaan",
    "Hasil",
    "Nilai Rujukan",
    "Unit",
    "Note",
  ],
});

function onAddDetail() {
  let obj = {
    Name: "",
    Lines: [],
  };
  data.record.DetailPemeriksaan.push(obj);
}

function fecthPacakge() {
  if (!props.mcuPaket) {
    data.record.DetailPemeriksaan = [];
    return;
  }
  const url = "/she/mcumasterpackage/get";
  axios.post(url, [props.mcuPaket]).then(
    (r) => {
      data.record.DetailPemeriksaan = r.data.Lines;
    },
    (e) => {
      util.showError(e);
    }
  );
}

function fecthAdditionalItem() {
  const url = "/she/mcutransaction/get-mcu-item-last-child";
  axios
    .post(url, {
      Where: {
        Op: "$in",
        Field: "ID",
        Value: props.mcuAddItem,
      },
    })
    .then(
      (r) => {
        data.additionalItem = r.data;
      },
      (e) => {
        util.showError(e);
      }
    );
}

function filteredLines(items) {
  return items.filter((o) => o.IsSelected);
}

watch(
  () => props.mcuPaket,
  (nv) => {
    fecthPacakge();
  }
);

onMounted(() => {
  util.nextTickN(3, () => {
    listControl.value.setFormRecord(data.record);
    if (data.record.DetailPemeriksaan.length == 0) fecthPacakge();
    if (props.mcuAddItem) fecthAdditionalItem();
  });
});
</script>
