<template>
  <div class="w-full">
    <label class="input_label" v-if="title != ''">{{ title }}</label>
    <data-list
      ref="listControl"
      :class="[readOnly ? 'is-readonly' : '']"
      hide-title
      :hide-action="readOnly"
      :hide-control="readOnly"
      no-gap
      grid-editor
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-hide-detail
      grid-hide-select
      grid-no-confirm-delete
      :init-app-mode="data.appMode"
      :grid-mode="data.appMode"
      new-record-type="grid"
      :grid-fields="['KMStart', 'KMEnd']"
      :grid-config="gridConfig"
      @alter-grid-config="onAlterGridConfig"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @gridRowFieldChanged="onHandleGridFieldChanged"
    >
      <template #grid_KMStart="{ item }">
        <div class="flex gap-4">
          <s-input
            keep-label
            hide-label
            v-model="item.KMStart"
            kind="number"
            :read-only="readOnly"
            @change="(field, v1, v2, old, ctlRef) => {
               item.KMTotal = item.KMEnd - v1;
            }"
          ></s-input>
          <uploader
            ref="gridAttachmentKMStart"
            :journalId="recordID"
            :journalType="attchKind"
            :config="{label: `KM Start : ${item.LineNo}`}"
            :tags="[`${attchKind}_KM_START_${recordID}`,`${attchKind}_KM_START_${recordID}_${item.ID}`]"
            :tags-for-get="[`${attchKind}_KM_START_${recordID}_${item.ID}`]"
            :key="1"
            bytag
            hide-label
            single-save
            @close="emit('close')"
          />
        </div>
      </template>
      <template #grid_KMEnd="{ item }">
        <div class="flex gap-4">
          <s-input
            keep-label
            hide-label
            v-model="item.KMEnd"
            :read-only="readOnly"
            kind="number"
            @change="(field, v1, v2, old, ctlRef) => {
               item.KMTotal = v1 - item.KMStart;
            }"
          ></s-input>
          <uploader
            ref="gridAttachmentKMEnd"
            :journalId="recordID"
            :journalType="attchKind"
            :config="{label: `KM End : ${item.LineNo}`}"
            :tags="[`${attchKind}_KM_END_${recordID}`,`${attchKind}_KM_END_${recordID}_${item.ID}`]"
            :tags-for-get="[`${attchKind}_KM_END_${recordID}_${item.ID}`]"
            :key="1"
            bytag
            hide-label
            single-save
            @close="emit('close')"
          />
        </div>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, inject, watch } from "vue";
import { DataList, SInput, loadGridConfig, util } from "suimjs";
import Uploader from "@/components/common/Uploader.vue";

const props = defineProps({
  readOnly: { type: Boolean, default: false },
  title: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
  gridConfig: { type: String, default: "" },
  siteId: { type: String, default: "" },
  recordID: { type: String, default: "" },
  attchKind: { type: String, default: "" },
});

const emit = defineEmits({
  "update:modelValue": null,
  close: null,
});

const listControl = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
});

function onGridRowDelete(_, index) {
  const newRecords = data.records.filter((_, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
  emit("update:modelValue", data.records);
}
function getMaxLineNo() {
  const obj = data.records.reduce(
    (acc, c) => {
      return acc.LineNo > c.LineNo ? acc : c;
    },
    { LineNo: 0 }
  );
  return obj.LineNo ?? 0;
}
function newRecord(r) {
  r.LineNo = getMaxLineNo() + 1;
  r.ID = util.uuid();
  r.KMStart = 0;
  r.KMEnd = 0;
  r.KMTotal = 0;

  data.records.push(r);
  updateItems();
}

function onHandleGridFieldChanged(name, v1, v2, old, record) {
  // if (["KMStart", "KMEnd"].includes(name)) {
  //   util.nextTickN(2, () => {
  //     record.KMTotal = record.KMEnd - record.KMStart;
  //   });
  // }
}
function onAlterGridConfig(config) {
  setTimeout(() => {
    updateItems();
  }, 500);
}
onMounted(() => {
  setTimeout(() => {
    // updateItems();
  }, 500);
});
</script>
