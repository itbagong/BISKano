<template>
  <div class="w-full border flex mb-4">
    <div class="basis-1/2 p-2">
      <div class="font-semibold flex gap-2 mb-1">
        <div class="border-r pr-2 border-black">{{ item.PoliceNum }}</div>
        <div>{{ item.CustomerName }}</div>
      </div>
      <div class="text-xs">
        {{ moment(item.TrxDate).format("DD MMM YYYY") }}
      </div>
    </div>
    <div class="basis-1/2 grid grid-cols-3 p-2 gap-2 content-center">
      <div class="flex border-r">
        <div class="font-bold">Income :</div>
        <div class="pl-2">{{ util.formatMoney(item.SumIncome, {}) }}</div>
      </div>
      <div class="flex">
        <div class="font-bold">Expense :</div>
        <div class="pl-2">{{ util.formatMoney(item.SumExpense, {}) }}</div>
      </div>
      <div class="flex border-r">
        <div class="font-bold">Revenue :</div>
        <div class="pl-2">{{ util.formatMoney(item.SumRevenue, {}) }}</div>
      </div>
    </div>
  </div>
</template>
<script setup>
import moment from "moment";
import { util } from "suimjs";

const props = defineProps({
  item: { type: Object, default: {} },
});
</script>
