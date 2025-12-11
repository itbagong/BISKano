<template>
  <s-card class="w-full bg-white grid_card" hide-footer no-gap>
    <template v-if="data.loadingGridCfg">
      <slot name="loader">
        <div class="loader"></div>
      </slot>
    </template>
    <s-grid
      class="sgrid-siteincome"
      :class="[readOnly ? 'is-readonly' : '']"
      ref="gridCtl"
      no-gap
      auto-commit-line
      editor
      :hide-action="hideAction"
      :hide-control="hideControl || readOnly"
      :hide-delete-button="hideDeleteButton || readOnly"
      hide-search
      hide-sort
      hide-refresh-button
      hide-detai
      hide-select
      hide-footer
      hide-detail
      grid-auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      @newData="onNewRecord"
      @deleteData="onRowDelete"
      @rowFieldChanged="onRowFieldChanged"
    >
      <template #item_Amount="{ item }">
        <slot name="item_Amount" :item="item">
          <s-input
            v-if="hasJournalID(item.JournalID)"
            read-only
            kind="number"
            class="text-right"
            v-model="item.Amount"
          />
        </slot>
      </template>
      <template #item_Name="{ item }">
        <s-input
          v-if="hasJournalID(item.JournalID)"
          read-only
          v-model="item.Name"
        />
      </template>
      <template #item_Notes="{ item }">
        <s-input
          v-if="hasJournalID(item.JournalID)"
          read-only
          v-model="item.Notes"
        />
      </template>
      <template #item_ApprovalStatus="{ item }">
        <status-text :txt="item.ApprovalStatus" />
      </template>
      <template #item_buttons_1="{ item }">
        <action-attachment
          :read-only="readOnly"
          :kind="`${attchKind}_INCOME`"
          :ref-id="attchRefId"
          :tags="[
            `${attchTagPrefix}_INCOME_${attchRefId}_${item.ID}`,
            tagUpload,
          ]"
          :tags-for-get="[`${attchTagPrefix}_INCOME_${attchRefId}_${item.ID}`]"
          :show-content="
            selectedAttchTag ==
            `${attchTagPrefix}_INCOME_${attchRefId}_${item.ID}`
          "
          @preOpen="emit('preOpenAttch', readOnly)"
        />
        <s-button
          v-if="
            item.ApprovalStatus == 'REJECTED' && hasJournalID(item.JournalID)
          "
          class="btn_reopen submit_btn text-xs"
          @click="reOpen(item)"
          label="RE-OPEN"
        />
      </template>
    </s-grid>
  </s-card>
</template>

<script setup>
import { reactive, ref, onMounted, inject, watch } from "vue";
import { SGrid, util, loadGridConfig, SCard, SInput, SButton } from "suimjs";
import StatusText from "@/components/common/StatusText.vue";
import ActionAttachment from "@/components/common/ActionAttachment.vue";

const axios = inject("axios");
const props = defineProps({
  readOnly: { type: Boolean, default: false },
  modelValue: { type: Array, default: () => [] },
  gridConfigUrl: {
    type: String,
    default: "/bagong/site_income/gridconfig",
  },
  hideAction: { type: Boolean, default: false },
  hideDeleteButton: { type: Boolean, default: false },
  hideControl: { type: Boolean, default: false },
  attchKind: {
    type: String,
    default: "",
  },
  attchRefId: {
    type: String,
    default: "",
  },
  attchTagPrefix: {
    type: String,
    default: "",
  },
  selectedAttchTag: {
    type: String,
    default: "",
  },
  tagUpload: { stype: String, default: "" },
});
const gridCtl = ref(null);

const emit = defineEmits({
  "update:modelValue": null,
  newRecord: null,
  rowFieldChanged: null,
  calc: null,
  preOpenAttch: null,
  reOpen: null,
});

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  gridCfg: {},
  loadingGridCfg: false,
  tag: props.tagAtt == "" ? props.kind : props.tagAtt,
});

function getMaxLineNo() {
  const obj = data.records.reduce(
    (acc, c) => {
      return acc.LineNo > c.LineNo ? acc : c;
    },
    { LineNo: 0 }
  );
  return obj.LineNo ?? 0;
}
function onNewRecord() {
  const r = {
    LineNo: getMaxLineNo() + 1,
    ID: util.uuid(),
    Name: "",
    Amount: 0,
    suimRecordChange: false,
    JournalID: "",
  };
  data.records.push(r);
  updateItems();
}
function updateItems() {
  util.nextTickN(2, () => {
    gridCtl.value.setRecords(data.records);
    emit("update:modelValue", data.records);
  });
}
function hasJournalID(journalID) {
  return journalID != "" && journalID != undefined;
}

function calc() {
  const Amount = data.records.reduce((total, e) => {
    total += e.Amount;
    return total;
  }, 0);
  emit("calc", Amount);
}
function onRowDelete(_, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  util.nextTickN(2, () => {
    calc();
    updateItems();
  });
}
function onRowFieldChanged(name, v1, v2, old, record) {
  rowFieldChanged(name, record);

  emit("rowFieldChange", name, v1, v2, old, record, () => {
    rowFieldChanged(name, record);
  });
  updateItems();
}
function rowFieldChanged(name, record) {
  gridCtl.value.setRecord(gridCtl.value.getCurrentIndex(), record);

  if (["Amount"].includes(name)) {
    util.nextTickN(2, () => {
      calc();
    });
  }

  updateItems();
}

function onAlterGridConfig(config) {
  updateItems();
}
function refresh() {
  data.records = props.modelValue;
  util.nextTickN(2, () => {
    updateItems();
    calc();
  });
}
onMounted(() => {
  data.loadingGridCfg = true;
  loadGridConfig(axios, props.gridConfigUrl).then(
    (r) => {
      data.loadingGridCfg = false;
      data.gridCfg = r;
      util.nextTickN(2, () => {
        updateItems();
        calc();
      });
    },
    (e) => {
      data.loadingGridCfg = false;
      util.showError(e);
    }
  );
});
function reOpen(record) {
  record.JournalID = "";
  record.ApprovalStatus = "";
  updateItems();
  emit("reOpen");
}
watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
defineExpose({
  calc,
  refresh,
});
</script>
<style>
.sgrid-siteincome.suim_grid.is-readonly table.suim_table > tbody > tr:hover {
  background: transparent !important;
}
</style>
