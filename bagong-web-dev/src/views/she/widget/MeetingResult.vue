<template>
  {{ readOnly }}
  <data-list
    ref="resultControl"
    hide-title
    no-gap
    grid-config="/she/meeting/results/gridconfig"
    grid-mode="grid"
    new-record-type="grid"
    form-keep-label
    :init-app-mode="data.appMode"
    :init-form-mode="data.formMode"
    grid-auto-commit-line
    grid-hide-select
    grid-hide-detail
    grid-hide-search
    grid-hide-sort
    grid-hide-refresh
    :grid-hide-delete="readOnly"
    @grid-row-add="addNew"
    @grid-row-delete="onGridRowDelete"
    :grid-editor="!readOnly"
    :grid-fields="['Pica', 'Mom']"
    @alter-grid-config="onAlterConfig"
  >
    <template #grid_Pica="{ item }">
      <div class="w-full">
        <toggle
          :read-only="readOnly"
          v-model="item.UsePica"
          class="w-[50px] mb-2"
          @change="item.Pica = { DueDate: new Date(), Status: 'Open' }"
        />
        <pica
          :read-only="readOnly"
          v-model="item.Pica"
          v-if="item.UsePica"
          hide-title
          class="p-2"
        />
      </div>
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
import { DataList, SInput, util, SButton } from "suimjs";
import Uploader from "@/components/common/Uploader.vue";
import Toggle from "@/components/common/SButtonToggle.vue";
import Pica from "@/components/common/ItemPica.vue";

const resultControl = ref(null);

const props = defineProps({
  readOnly: { type: Boolean, default: false },
  modelValue: { type: Array, default: () => [] },
  jurnalId: { type: String, default: "" },
});

const data = reactive({
  record: props.modelValue == undefined ? [] : props.modelValue,
});

const emit = defineEmits({
  "update:modelValue": null,
  close: null,
});

function addNew(r) {
  r.Pica = {
    DueDate: new Date(),
  };
  r.ResultId = uuid();
  data.record.push(r);
  updateItems();
}

function updateItems() {
  nextTick(() => {
    resultControl.value.setGridRecords(data.record);
  });
}

function onGridRowDelete(_, index) {
  const newRecords = data.record.filter((_, idx) => {
    return idx != index;
  });
  data.record = newRecords;
  updateItems();
}

function uuid() {
  return Date.now().toString(36) + Math.random().toString(36).substr(2);
}

function onClose(val) {
  emit("close", val);
}

function onAlterConfig(config) {
  debounce(() => {
    updateItems();
  }, 500);
}

function createDebounce() {
  let timeout = null;
  return function (fnc, delayMs) {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      fnc();
    }, delayMs || 500);
  };
}
const debounce = createDebounce();

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>

<style>
.SHE_MEETING_RESULT
  .suim_area_table
  > .suim_table
  > thead
  > tr
  > th:nth-child(3) {
  width: 400px;
}
</style>
