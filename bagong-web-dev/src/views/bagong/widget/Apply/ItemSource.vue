<template>
  <div
    @click="onSelect"
    class="p-2"
    :class="[selectedSource._id == item._id ? 'acitve-item' : '']"
  >
    <div class="flex justify-between mb-3 items-start">
      <div v-if="kind == 'apply'" class="font-semibold">
        {{ item._id }} - {{ item.VoucherNo }}
      </div>
      <div v-else class="flex gap-2 items-center">
        <div class="font-semibold">{{ item._id }}</div>
        -
        <AccountSelector hide-label v-model="item.Account" read-only />
      </div>

      <div class="text-[0.7rem] flex gap-2 items-center">
        {{ moment(item.Created).format("DD-MMM-yyyy hh:mm:ss") }}
        <button
          @click="onReset"
          v-if="item.CurrentSettled > 0"
          class="bg-primary text-white w-[40px] h-[24px] flex justify-center items-center"
        >
          Reset
        </button>
      </div>
    </div>
    <div class="grid grid-cols-5 gap-6">
      <div class="col-span-2">
        <div v-if="kind == 'apply'">
          {{ item.SourceType }}
        </div>
        <br />
        <p class="text-[0.7rem]">{{ item.Text }}</p>
      </div>
      <div class="col-span-3">
        <div class="grid grid-cols-4 gap-6 mb-3">
          <div class="text-right">
            <label class="input_label">
              <div>Amount</div>
            </label>
            <div class="bg-transparent">
              {{ util.formatMoney(item.Amount) }}
            </div>
          </div>
          <div class="text-right">
            <label class="input_label">
              <div>Outstanding</div>
            </label>
            <div class="bg-transparent">
              {{ util.formatMoney(oustanding) }}
            </div>
          </div>

          <div class="text-right">
            <label class="input_label">
              <div>Settled</div>
            </label>
            <div class="bg-transparent">
              {{ util.formatMoney(item.Settled + item.CurrentSettled) }}
            </div>
          </div>

          <div class="text-right">
            <label class="input_label">
              <div>Current Settled</div>
            </label>
            <div class="bg-transparent">
              {{ util.formatMoney(item.CurrentSettled) }}
            </div>
            <div
              v-if="selectedInvoice.CurrentSourceRecID === item._id"
              class="text-primary text-[0.65rem] font-bold"
            >
              {{ util.formatMoney(selectedInvoice.Settled) }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed } from "vue";

import AccountSelector from "@/components/common/AccountSelector.vue";
import {
  SList,
  SForm,
  SInput,
  loadFormConfig,
  util,
  SButton,
  SModal,
} from "suimjs";
import moment from "moment";
const props = defineProps({
  kind: { type: String, default: "apply" },
  item: {
    type: Object,
    default: {
      _id: "",
      VoucherNo: "",
      SourceType: "",
      Text: "",
      Settled: 0,
      Amount: 0,
      CurrentSettled: 0,
      Account: {},
    },
  },
  selectedSource: { type: Object, default: { _id: "", MapApply: {} } },
  selectedInvoice: {
    type: Object,
    default: { _id: "", CurrentSourceRecID: "", Settled: 0 },
  },
  calcOustanding: { type: Function },
});
const emit = defineEmits({
  onSelect: null,
  onReset: null,
});

const oustanding = computed({
  get() {
    return props.calcOustanding(props.item);
  },
});

function onSelect() {
  emit("onSelect", props.item);
}
function onReset() {
  emit("onReset", props.item);
}
</script>
<style scoped>
 .label-number{
  text-align: right;
}
</style>