<template>
  <s-card class="w-full bg-white grid_card sitentry-expense" hide-footer no-gap>
    <template v-if="data.loadingGridCfg">
      <slot name="loader">
        <div class="loader"></div>
      </slot>
    </template>
    <s-grid
      class="sgrid-siteexpense"
      :class="[readOnly ? 'is-readonly' : '']"
      ref="gridCtl"
      no-gap
      auto-commit-line
      editor
      :hide-action="hideAction"
      :hide-control="hideControl || readOnly"
      hide-search
      hide-sort
      hide-refresh-button
      :hide-delete-button="hideDeleteButton || readOnly"
      hide-select
      hide-detail
      grid-auto-commit-line
      no-confirm-delete
      hide-footer
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
      <template #item_ID="{ item }">
        {{ item.ID }}
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
      <template #item_Value="{ item, header }">
        <s-input
          v-if="hasJournalID(item.JournalID)"
          read-only
          v-model="item.Value"
          class="text-right"
          kind="number"
        />
      </template>
      <template #item_UnitID="{ item, header }">
        <s-input
          :read-only="hasJournalID(item.JournalID)"
          read-only
          v-model="item.UnitID"
          use-list
          :lookup-url="header.input.lookupUrl"
          :lookup-key="header.input.lookupKey"
          :field="header.input.field"
          :lookup-labels="header.input.lookupLabels"
          :lookup-searchs="header.input.lookupSearchs"
        />
      </template>
      <template #item_ExpenseTypeID="{ item, header }">
        <s-input
          :read-only="hasJournalID(item.JournalID)"
          v-model="item.ExpenseTypeID"
          use-list
          :lookup-url="header.input.lookupUrl"
          :lookup-key="header.input.lookupKey"
          :lookup-payload-builder="
            props.groupIdValue.length > 0 ? expensePayload : undefined
          "
          :field="header.input.field"
          :lookup-labels="header.input.lookupLabels"
          :lookup-searchs="header.input.lookupSearchs"
        />
      </template>
      <template #item_JournalID="{ item }">
        <div class="bg-transparent" v-if="hasJournalID(item.JournalID)">
          <a
            href="#"
            class="text-blue-400 hover:text-blue-800"
            @click="redirect(item.JournalID)"
            >{{ item.JournalID }}
          </a>
        </div>
      </template>
      <template #item_ApprovalStatus="{ item }">
        <status-text :txt="item.ApprovalStatus" />
      </template>
      <template #item_buttons="{ item }">
        <template v-if="hasJournalID(item.JournalID)">&nbsp;</template>
      </template>
      <template #item_buttons_1="{ item }">
        <action-attachment
          :kind="`${attchKind}`"
          :ref-id="attchRefId"
          :tags="buildTags(item)"
          :tags-for-get="[`${attchTagPrefix}_EXPENSE_${attchRefId}_${item.ID}`]"
          :read-only="hasJournalID(item.JournalID)"
          @preOpen="emit('preOpenAttch', hasJournalID(item.JournalID))"
          @close="emit('closeAttch', hasJournalID(item.JournalID))"
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
import { SGrid, util, loadGridConfig, SInput, SCard, SButton } from "suimjs";
import ActionAttachment from "@/components/common/ActionAttachment.vue";

import helper from "@/scripts/helper.js";
import StatusText from "@/components/common/StatusText.vue";
import { useRouter } from "vue-router";

const router = useRouter();
const axios = inject("axios");
const props = defineProps({
  readOnly: { type: Boolean, default: false },
  modelValue: { type: Array, default: () => [] },
  gridConfigUrl: {
    type: String,
    default: "/bagong/site_expense/gridconfig",
  },
  hideAction: { type: Boolean, default: false },
  hideDeleteButton: { type: Boolean, default: false },
  hideControl: { type: Boolean, default: false },
  groupIdValue: { type: Array, default: () => [] },
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
  tagUpload: {
    type: String,
    default: "",
  },
});
const gridCtl = ref(null);

const emit = defineEmits({
  "update:modelValue": null,
  newRecord: null,
  rowFieldChanged: null,
  calc: null,
  preOpenAttch: null,
  closeAttch: null,
  reOpen: null,
});

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  gridCfg: {},
  loadingGridCfg: false,
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
  const r = {};
  r.LineNo = getMaxLineNo() + 1;
  r.ID = util.uuid();
  r.Name = "";
  r.ExpenseTypeID = "";
  r.Amount = 0;
  r.Value = 0;
  r.TotalAmount = 0;
  r.suimRecordChange = false;
  r.JournalID = "";

  emit("newRecord", r);
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
  const { Amount, Value, TotalAmount } = data.records.reduce(
    (total, e) => {
      total.Amount += e.Amount;
      total.Value += e.Value;
      total.TotalAmount += e.TotalAmount;
      return total;
    },
    {
      Amount: 0,
      Value: 0,
      TotalAmount: 0,
    }
  );
  emit("calc", { Amount, Value, TotalAmount });
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
  if (name == "Value") {
    record.TotalAmount = v1 * record.Amount;
  }

  if (name == "Amount") {
    record.TotalAmount = v1 * record.Value;
  }

  rowFieldChanged(name, record);

  emit("rowFieldChange", name, v1, v2, old, record, () => {
    rowFieldChanged(name, record);
  });
  updateItems();
}
function rowFieldChanged(name, record) {
  gridCtl.value.setRecord(gridCtl.value.getCurrentIndex(), record);

  if (["TotalAmount", "Value", "Amount"].includes(name)) {
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

function expensePayload(search) {
  const lookupKey = "_id";
  const lookupLabels = ["Name"];
  const lookupSearchs = ["_id", "Name"];

  const qp = {};

  if (search != "") qp.Take = 20;

  qp.Sort = [lookupLabels[0]];
  qp.Select = lookupSearchs;

  let idInSelect = false;
  const selectedFields = lookupLabels.map((x) => {
    if (x == lookupKey) {
      idInSelect = true;
    }
    return x;
  });

  if (!idInSelect) {
    selectedFields.push(lookupKey);
  }

  qp.Select = selectedFields;

  if (search.length > 0 && lookupSearchs.length > 0) {
    if (lookupSearchs.length == 1)
      qp.Where = {
        Field: lookupSearchs[0],
        Op: "$contains",
        Value: [search],
      };
    else
      qp.Where = {
        Op: "$or",
        items: lookupSearchs?.map((el) => {
          return { Field: el, Op: "$contains", Value: [search] };
        }),
      };
  }

  if (qp.Where != undefined) {
    const items = [
      { Op: "$contains", Field: `GroupID`, Value: props.groupIdValue },
    ];

    items.push(qp.Where);
    qp.Where = {
      Op: "$and",
      items: items,
    };
  } else {
    qp.Where = { Op: "$contains", Field: `GroupID`, Value: props.groupIdValue };
  }
  return qp;
}

function reOpen(record) {
  record.JournalID = "";
  record.ApprovalStatus = "";
  updateItems();
  emit("reOpen");
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
function buildTags(item) {
  let tags = [`${props.attchTagPrefix}_EXPENSE_${props.attchRefId}_${item.ID}`]
  if (props.tagUpload) {
    tags.push(props.tagUpload)
  }
  return tags
}
function redirect(JournalID) {
  const url = router.resolve({
    name: "fico-VendorTransaction",
    query: { id: JournalID },
  });
  window.open(url.href, "_blank");
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
.sgrid-siteexpense.suim_grid.is-readonly table.suim_table > tbody > tr:hover {
  background: transparent !important;
  cursor: auto;
}
</style>
