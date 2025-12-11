<template>
  <div class="grid grid-cols-1 gap-y-3">
    <div class="flex justify-between border-b pb-2 items-end">
      <div class="font-semibold text-[1.3em]">Stage</div>
      <s-button
        icon="refresh"
        class="btn_primary"
        tooltip="refresh"
        @click="emit('refresh')"
      />
    </div>
    <div class="grid grid-cols-1 gap-y-4 steps">
      <div
        v-for="(r, i) in data.records"
        :key="i"
        class="p-2 hover:bg-slate-200 font-semibold flex justify-between cursor-pointer rounded-md list"
        :class="[data.selected == r.kind ? 'bg-slate-200' : '']"
        @click="onSelect(r)"
      >
        <div class="flex gap-1 items-center">
          <mdicon name="checkbox-blank-circle" class="circle" size="10" />
          {{ r.text }}
        </div>
        <div class="text-gray-500">{{ stageCounts[r.kind] }}</div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, watch } from "vue";
import { SButton } from "suimjs";

const props = defineProps({
  modelValue: { type: String, default: "" },
  manPowerId: { type: String, default: "" },
  stageCounts: { type: String, default: "" },
});

const emit = defineEmits({
  select: null,
  refresh: null,
  "update:modelValue": null,
});

const data = reactive({
  records: [
    {
      kind: "Screening",
      text: "Screnning",
      total: null,
    },
    {
      kind: "PshycologicalTest",
      text: "Psikotes",
      total: null,
    },
    {
      kind: "Interview",
      text: "Interview",
      total: null,
    },
    {
      kind: "TechnicalInterview",
      text: "Technical Interview",
      total: null,
    },
    {
      kind: "MCU",
      text: "MCU",
      total: null,
    },
    {
      kind: "Training",
      text: "Training",
      total: null,
    },
    {
      kind: "OLPlotting",
      text: "OL & Ploting",
      total: null,
    },
    {
      kind: "PKWTT",
      text: "PKWT",
      total: null,
    },
    {
      kind: "OnBoarding",
      text: "Onboarding",
      total: null,
    },
  ],
  selected: props.modelValue,
});

function refresh() {}

function onSelect(r) {
  data.selected = r.kind;
}

defineExpose({
  refresh,
});

watch(
  () => data.selected,
  (nv) => {
    emit("select", data.selected);
  }
);
</script>

<style>
.list:nth-child(1) .circle {
  color: #e74c3c;
}
.list:nth-child(2) .circle {
  color: #9b59b6;
}
.list:nth-child(3) .circle {
  color: #2980b9;
}
.list:nth-child(4) .circle {
  color: #1abc9c;
}
.list:nth-child(5) .circle {
  color: #27ae60;
}
.list:nth-child(6) .circle {
  color: #f1c40f;
}
.list:nth-child(7) .circle {
  color: #e67e22;
}
.list:nth-child(8) .circle {
  color: #825e5c;
}
.list:nth-child(8) .circle {
  color: #bae4e6;
}
.list:nth-child(9) .circle {
  color: #127c2c;
}
</style>
