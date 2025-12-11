<template>
  <div
    @mouseover="onHover"
    @mouseleave="onLeave"
    class="p-2 disabled-apply card-apply"
    :class="[
      selectedSource.MapApply[item._id]?.IsSettled === true
        ? 'active-apply'
        : '',
    ]"
  >
    <div class="flex justify-between mb-3 items-start">
      <div class="font-semibold">{{ item._id }} - {{ item.VoucherNo }}</div>
      <div class="text-[0.7rem] flex gap-3 items-center">
        {{ moment(item.Created).format("DD-MMM-yyyy hh:mm:ss") }}

        <button
          v-if="item.SourceRecID !== ''"
          @click="onUncheck"
          class="w-[50px] h-[24px] flex justify-center items-center bg-primary text-white"
        >
          Reset
        </button>

        <button
          v-else-if="item.CurrentSourceRecID !== ''"
          @click="onReset"
          class="w-[50px] h-[24px] flex justify-center items-center bg-primary text-white"
        >
          Reset
        </button>
        <template
          v-else-if="
            item.CurrentSourceRecID === '' && selectedSource._id !== ''
          "
        >
          <template v-if="kind == 'apply'">
            <button
              v-if="sourceOustanding >= oustanding"
              @click="onSettled"
              class="w-[50px] h-[24px] flex justify-center items-center outline outline-[1px] outline-primary text-primary"
            >
              Apply
            </button>
          </template>

          <button
            v-else
            @click="onSettled"
            class="w-[50px] h-[24px] flex justify-center items-center outline outline-[1px] outline-primary text-primary"
          >
            Apply
          </button>
        </template>
      </div>
    </div>
    <div class="grid grid-cols-5 gap-6 mb-2">
      <div class="col-span-2">
        {{ item.SourceType }}
        <br />
        <p class="text-[0.7rem]">{{ item.Text }}</p>
      </div>
      <div class="col-span-3">
        <div class="grid grid-cols-3 gap-6 mb-3">
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
              {{ util.formatMoney(item.Settled) }}
            </div>
          </div>
        </div>
      </div>
    </div>
    <slot name="adjustments" :item="item" />
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed } from "vue";
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
      SourceType: "",
      Text: "",
      Settled: 0,
      Amount: 0,
      Account: {},
      Adjustments: [],
      SourceRecID: "",
      CurrentSourceRecID: "",
    },
  },
  selectedSource: {
    type: Object,
    default: {
      _id: "",
      MapApply: {},
      Amount: 0,
      Settled: 0,
      CurrentSettled: 0,
    },
  },
  calcAdjustment: { type: Function },
  calcOustandingSource: { type: Function },
  calcOustanding: { type: Function },
});
const totalAdjustment = computed({
  get() {
    return props.calcAdjustment(props.item.Adjustments);
  },
});

const hasCurrentSourceRecId = computed({
  get() {
    return (
      props.item.CurrentSourceRecID !== "" &&
      props.item.CurrentSourceRecID !== undefined
    );
  },
});

const hasSourceId = computed({
  get() {
    return (
      props.selectedSource._id !== "" && props.selectedSource._id !== undefined
    );
  },
});

const oustanding = computed({
  get() {
    return props.calcOustanding(props.item, totalAdjustment.value);
  },
});
const sourceOustanding = computed({
  get() {
    return props.calcOustandingSource(props.selectedSource);
  },
});
const emit = defineEmits({
  onHover: null,
  onLeave: null,
  onReset: null,
  onSettled: null,
  onUncheck: null,
});

function onHover() {
  emit("onHover", props.item);
}
function onLeave() {
  emit("onLeave", props.item);
}
function onReset() {
  emit("onReset", props.item);
}
function onSettled() {
  emit("onSettled", props.item, oustanding.value);
}
function onUncheck() {
  emit("onUncheck", props.item);
}
</script>