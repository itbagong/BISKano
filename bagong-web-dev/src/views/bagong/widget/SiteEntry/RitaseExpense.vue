<template>
  <expense
    :read-only="readOnly"
    ref="gridCtl"
    v-model="data.records"
    @calc="calcLineTotal"
    hide-control
    hide-delete-button
    :grid-config-url="
      readOnly
        ? '/bagong/siteexpense-trayek-read/grid/gridconfig'
        : '/bagong/siteexpense-trayek/grid/gridconfig'
    "
    :attch-kind="attchKind"
    :attch-refId="attchRefID"
    :attch-tag-prefix="attchTagPrefix"
    :tag-upload="tagUpload"
    @preOpenAttch="emit('preOpenAttch', readOnly)"
    @reOpen="reOpen"
    @newRecord="newRecord"
  >
    <template #item_Amount="{ item }">
      <template v-if="item.ExpenseCategory === 'Per Person'">
        <div class="flex gap-2 justify-end items-center w-full">
          <div class="max-w-[100px]">
            <s-input
              :read-only="item.JournalID != ''"
              kind="number"
              field="Value"
              v-model="item.Value"
              @change="(...args) => onChangeInputVal(item, ...args)"
            ></s-input>
          </div>
          x
          <div class="font-semibold text-[0.75rem]">
            {{ util.formatMoney(item.Amount) }}
          </div>
          =
          <div class="font-semibold text-[0.75rem]">
            {{ util.formatMoney(item.TotalAmount) }}
          </div>
        </div>
      </template>
    </template>
  </expense>
</template>
<script setup>
import { reactive, ref, onMounted, watch } from "vue";
import { util, SInput } from "suimjs";
import Expense from "./Expense.vue";
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false },
  attchKind: {
    type: String,
    default: "",
  },
  attchRefID: {
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
});

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
});
watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
function onChangeInputVal(record, name, v1, v2) {
  record.TotalAmount = v1 * record.Amount;
  util.nextTickN(2, () => {
    gridCtl.value.calc();
  });
}
function calcLineTotal(total) {
  emit("calc", total.TotalAmount);
}
function newRecord(r) {
  r.Value = 1;
}
</script>
