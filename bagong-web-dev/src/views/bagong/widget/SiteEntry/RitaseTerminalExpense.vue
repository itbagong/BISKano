<template>  
    <table class="w-full">
      <thead>
        <tr class=" [&>*]:p-2 [&>*]:border-[1px] ">
          <td :style="{width:widthCell()+'%'}">&nbsp;</td>
          <td v-for="(item, idx) in terminals"   :key="idx"  :style="{width:widthCell()+'%'}"  class="font-semibold bg-[#FFF9E6] text-[#8C6C00]">{{item.TerminalName}}</td>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(item,idx) in value" :key="idx" class=" [&>*]:p-2 [&>*]:border-[1px] ">
          <td class="bg-[#FFF9E6] text-[#8C6C00] font-semibold">{{item.ExpenseName}}</td> 
          <td v-for="(detail, idx2) in item.List" :key="idx+idx2"  :class="[detail.Enable ? '' : 'bg-gray-200',]" >
            <div class=" flex gap-2 justify-end items-center w-full" v-if="detail.Enable" >
              <template v-if="detail.ExpenseCategory == 'Per Person'">
                <div class="max-w-[100px]">
                  <s-input v-model="detail.Value" kind="number" @change=" (...args) =>  onChangeCalcAmount(idx,idx2,detail.ExpenseValue, ...args,) "  />
                </div> 
                x
                <div class="font-semibold text-[0.75rem]"> {{util.formatMoney(detail.ExpenseValue)}}</div>
                =
                <div class="font-semibold text-[0.75rem]"> {{util.formatMoney(detail.Amount)}}</div>
              </template>
              <template v-else>
                 <s-input v-model="detail.Amount" read-only kind="number"/>
              </template>
            </div>
          </td>
        </tr>
      </tbody>
    </table> 
</template>

<script setup>
import { reactive, ref, onMounted, inject, computed } from "vue";
import { DataList, SInput, util } from "suimjs";

const axios = inject("axios");
const props = defineProps({
  terminals: { type: Array, default: () => [] },
  modelValue: { type: Array, default: () => [] },
});
const emit = defineEmits({
  "update:modelValue": null,
  calcTotalAmount: null,
});
function widthCell() {
  return 100 / parseInt(props.terminals.length + 1);
}
function onChangeCalcAmount(idx1, idx2, expenseValue, name, v1) {
  value.value[idx1].List[idx2].Amount = parseInt(v1) * parseInt(expenseValue);

  util.nextTickN(2, () => {
    calcTotalAmount();
  });
}
function calcTotalAmount() {
  const totalAmount = value.value.reduce((total, e) => {
    const r = e.List.reduce((total2, e2) => {
      return total2 + parseInt(e2.Amount ?? 0);
    }, 0);
    return total + r;
  }, 0);
  emit("calcTotalAmount", totalAmount);
}
const value = computed({
  get() {
    return props.modelValue;
  },
  set(v) {
    emit("update:modelValue", v);
  },
});
onMounted(() => {
  util.nextTickN(2, () => {
    calcTotalAmount();
  });
});
</script>