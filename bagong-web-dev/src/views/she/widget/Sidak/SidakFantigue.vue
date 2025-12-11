<template>
  <data-list
    class="card"
    ref="listControl"
    hide-title
    no-gap
    form-config="/she/sidak/fatigue/formconfig"
    form-keep-label
    :init-app-mode="data.appMode"
    form-hide-submit
    form-hide-cancel
    :form-fields="['WakeUp', 'Sign', 'Sleep', 'Fit', 'MedicineConsumption']"
    @alter-form-config="onAlterFormConfig"
  >
    <template #form_input_Sleep="{ item }">
      <s-input
        kind="time"
        v-model="item.Sleep"
        label="Sleep"
        @change="calculateDuration"
      />
    </template>
    <template #form_input_WakeUp="{ item }">
      <s-input
        kind="time"
        v-model="item.WakeUp"
        label="Wake Up"
        @change="calculateDuration"
      />
    </template>
    <template #form_input_Sign="{ item, config }">
      <div class="hidden">{{ (config.label = "Fatigue Sign") }}</div>
      <uploader
        ref="gridAttachmentSign"
        :journalId="jurnalId"
        :config="config"
        journalType="SHE_Fantigue_SIGN"
        :key="1"
        single-save
      />
    </template>
    <template #form_input_Fit="{ item }">
      <label class="input_label">
        <div>Fit</div>
      </label>
      <s-toggle
        v-model="item.Fit"
        class="w-[120px] mt-0.5 pointer-none"
        disabled
      />
    </template>

    <template #form_input_MedicineConsumption="{ item }">
      <label class="input_label">
        <div>Medicine Consumption</div>
      </label>
      <s-toggle
        v-model="item.MedicineConsumption"
        class="w-[120px] mt-0.5 pointer-none"
      />
    </template>
  </data-list>
</template>

<script setup>
import { reactive, ref, inject, watch, computed, onMounted } from "vue";
import { DataList, SInput, util, SButton } from "suimjs";
import moment from "moment";
import Uploader from "@/components/common/Uploader.vue";
import SToggle from "@/components/common/SButtonToggle.vue";

const gridAttachmentSign = ref(null);
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
  listControl.value.setFormRecord(data.record);
  data.record.Fit = data.record.Fit ?? false;
  data.record.Sign = data.record.Sign ?? [];
}

function calculateDuration() {
  util.nextTickN(2, () => {
    let start = formatTimeToDateTime(data.record.Sleep, false);
    let end = formatTimeToDateTime(data.record.WakeUp, true);
    let asHours = moment.duration(start.diff(end)).asHours();
    let duration = asHours ? Math.abs(asHours) : 0;
    data.record.SleepDuration = duration;
    data.record.Fit = duration >= 6;
  });
}

function formatTimeToDateTime(val, isWake) {
  let newDate = new Date();
  let hoursSleep = parseInt(data.record.Sleep.split(":")[0]);
  let hoursWake = parseInt(data.record.WakeUp.split(":")[0]);
  let isNextDay = hoursSleep > hoursWake;
  let customDate =
    isNextDay && isWake ? newDate.setDate(newDate.getDate() + 1) : new Date();

  let myDate = moment(customDate).format("YYYY-MM-DD");
  let myTime = val + ":00";
  let res = moment(myDate + " " + myTime).format();
  return moment(res);
}

function setAttr(field, attr, val) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr(field, attr, val);
  });
}

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

watch(
  () => data.record.MedicineConsumption,
  (nv) => {
    setAttr("MedicineDescription", "hide", !nv);
  },
  { deep: true }
);
</script>
