<template>
     <div class="border-t w-full mt-4 grid grid-cols-4 pt-4 gap-3">
        <div class="col-span-3 grid grid-cols-3 gap-3">
            <template v-if="revenueType == 'Premi'"> 
                <div class="flex flex-col" v-if="method == 1"> 
                    <b class="pb-1">Perhitungan Bersih 1 </b> 
                    <div class="flex gap-2 items-center">
                        <div class="border-r pr-2">{{util.formatMoney(totalMethod)}}</div>
                        <div class="flex gap-1 items-center" >
                        <b>Bonus: </b>{{util.formatMoney(bonus1)}} 
                        </div>
                        
                    </div>
                    <div class="text-[0.75em]">
                        <div class="flex gap-4">
                        <div><b>Bonus Target</b>: {{util.formatMoney(target1)}}</div>
                        <div><b>Bonus Persen</b>: {{percent1 * 100}}%</div>
                        </div>
                    </div>
                </div>
                <div class="flex flex-col" v-else-if="method == 2">
                    <b class="pb-1">Perhitungan 2 Setengah Kotor Bonus 2 Target </b>   
                    <div class="flex gap-2 items-center">
                        <div class="border-r pr-2">{{util.formatMoney(totalMethod)}}</div>
                        <div class="flex gap-1  items-center">
                        <b>Bonus-1 : </b>{{util.formatMoney(bonus1)}},
                        <b>Bonus-2 : </b>{{util.formatMoney(bonus2)}}
                        
                    </div>
                </div>
                <div class="text-[0.75em]">
                    <div class="flex gap-4">
                    <div>
                        <b>Bonus-1 Target</b>: {{util.formatMoney(target1)}} <br/>
                        <b>Bonus-2 Target</b>: {{util.formatMoney(target2)}}
                    </div>
                    <div>
                        <b>Bonus-1 Persen</b>: {{percent1 * 100}}%<br/>
                        <b>Bonus-2 Persen</b>: {{percent2 * 100}}%
                    </div>
                    </div> 
                </div> 
                </div>
                <div class="flex flex-col" v-else-if="method == 3">
                    <b class="pb-1">Perhitungan 3 Setengah Kotor Bonus 1 Target  </b> 
                    <div class="flex gap-2 items-center">
                        <div class="border-r pr-2">{{util.formatMoney(totalMethod)}}</div>
                        <div class="flex gap-1 items-center">
                        <b>Bonus: </b>{{util.formatMoney(bonus1)}}  
                    </div>
                        
                </div>
                <div class="text-[0.75em]">
                    <div class="flex gap-4">
                        <div><b>Bonus Target</b>: {{util.formatMoney(target1)}}</div>
                        <div><b>Bonus Persen</b>: {{percent1 * 100}}%</div>
                    </div>
                </div>
                </div>
            </template>
        </div>
        <div class="grid grid-cols-3 gap-3  ">
            <div class="flex flex-col">
                <b class="pb-1">Revenue</b>
                {{util.formatMoney(revenue)}} 
            </div>
            <div class="flex flex-col">
                <b class="pb-1">Income</b> {{util.formatMoney(income)}}
            </div>
            <div class="flex flex-col">
                <b class="pb-1">Expense</b> {{util.formatMoney(expense)}}
            </div>
        </div>
     </div>
</template>
<script setup>
import { util } from "suimjs";
const props = defineProps({
  revenueType: { type: String, default: "Premi" },
  income: { type: Number, default: 0 },
  revenue: { type: Number, default: 0 },
  expense: { type: Number, default: 0 },
  method: { type: Number, default: 0 },
  target1: { type: Number, default: 0 },
  target2: { type: Number, default: 0 },
  percent1: { type: Number, default: 0 },
  percent2: { type: Number, default: 0 },
  bonus1: { type: Number, default: 0 },
  bonus2: { type: Number, default: 0 },
  totalBonus: { type: Number, default: 0 },
  totalMethod: { type: Number, default: 0 },
});
</script>