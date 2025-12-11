<template>
  <div>
    <div class="flex gap-4">
      <div
        v-for="(ritases, idx) in value"
        :key="idx"
        class="flex w-full items-end"
        style="max-width: 100%; overflow-x: auto; overflow-y: hidden"
      >
        <div
          class="flex-row w-full [&>*]:p-2 [&>*]:border-[1px] [&>*]:w-full] [&>*]:h-[40px]"
        >
          <div class="font-semibold bg-[#F3F4F7] min-w-[150px]">
            Total Penumpang
          </div>
          <div class="font-semibold bg-[#FBE9E8] text-error min-w-[150px]">
            Total Pendapatan
          </div>
        </div>
        <div
          v-for="(ritase, idx2) in ritases"
          :key="idx + idx2"
          class="flex-row w-full [&>*]:p-2 [&>*]:w-full] relative"
        >
          <div
            class="border-[1px] bg-[#FFF9E6] text-[#8C6C00] font-semibold h-[50px]"
          >
            {{ ritase.TerminalName }}
          </div>
          <div
            v-for="(Passenger, idx3) in ritase.Passengers"
            :key="idx + idx2 + idx3"
            class="border-[1px] flex gap-2 items-center h-[50px]"
          >
            <div class="min-w-[60px] max-w-[60px] text-right">
              <s-input
                :read-only="readOnly"
                v-model="Passenger.Total"
                kind="number"
                class="min-w-[60px]"
                @change="
                  (...args) => onInputPassenger(idx, idx2, idx3, ...args)
                "
              />
            </div>
            x
            <div class="font-semibold text-[0.75rem]">
              {{ util.formatMoney(Passenger.Tariff) }}
            </div>
          </div>
          <div class="border-[1px] text-right bg-[#F3F4F7] h-[40px]">
            {{ ritase.TotalPassenger }}
          </div>
          <div class="border-[1px] text-right bg-[#FBE9E8] text-error h-[40px]">
            {{ util.formatMoney(ritase.CombinedAmount)  }} 
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, computed, watch } from "vue";
import { SInput, util } from "suimjs";
import helper from "@/scripts/helper.js";

const props = defineProps({
  siteEntryAssetID: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false },
  trayekName: { type: String, default: "" },
  specialTrayekName: { type: String, default: "" },
  specialTerminalFrom: { type: String, default: "" },
  kurs: { type: Number, default: 0 },
});

const emit = defineEmits({
  "update:modelValue": null,
  calcTotalPenumpangA: null,
  calcTotalPenumpangB: null,
  calcTotalAmountA: null,
  calcTotalAmountB: null,
});

const value = computed({
  get() {
    return props.modelValue;
  },

  set(v) {
    emit("update:modelValue", v);
  },
});
function calcIncome(idxRitase, idxArr, idxPassenger) {
  const Passengers = JSON.parse(
    JSON.stringify(value.value[idxRitase][idxArr].Passengers)
  );
  const e = Passengers[idxPassenger]; 
 
  if(props.trayekName === props.specialTrayekName && e.From === props.specialTerminalFrom){
      e.Income = parseInt(e.Total) * parseInt(e.Tariff) * props.kurs
  }else{  
      e.Income = parseInt(e.Total) * parseInt(e.Tariff);
  } 
  value.value[idxRitase][idxArr].Passengers = [...Passengers];
}

function calcTotalIncome(idxRitase, idxArr) {
  const obj = JSON.parse(JSON.stringify(value.value[idxRitase][idxArr]));
  obj.Amount = obj.Passengers.reduce((accumulator, obj) => {
      return parseInt(accumulator) + parseInt(obj.Income ?? 0);
  }, 0);
  value.value[idxRitase][idxArr] = { ...obj };

}
function calcTotalCombinedAmount(idxRitase){
  const ritase = helper.cloneObject(value.value[idxRitase])

  value.value[idxRitase] = ritase.map((el,i)=>{
    el.CombinedAmount = i == 0 ? el.Amount : ritase[i-1].CombinedAmount + el.Amount 
    return el
  }) 

}
function calcTotalPassenger(idxRitase) {
  const ritase = value.value[idxRitase];
  ritase.forEach((e, i) => {
    const prevTotalPassenger = i == 0 ? 0 : ritase[i - 1].TotalPassenger;

    const dropPassenger = value.value[idxRitase].reduce((accumulator, obj) => {
      const sameTerminal = obj.Passengers.filter((d) => d.To == e.TerminalID);
      const Passenger =
        sameTerminal.length == 0 ? 0 : parseInt(sameTerminal[0].Total ?? 0);
      return parseInt(accumulator) + parseInt(Passenger);
    }, 0);

    e.TotalPassenger =
      e.Passengers.reduce((accumulator, obj) => {
        return parseInt(accumulator) + parseInt(obj.Total);
      }, 0) +
      parseInt(prevTotalPassenger) -
      parseInt(dropPassenger);
  });
}

function calcTotalAllPassenger(idxRitase) {
  const ritase = value.value[idxRitase] ?? [];
  let total = 0;
  ritase.forEach((item) => {
    item.Passengers.forEach((passengger) => {
      total += passengger.Total || 0;
    });
  });
  if (idxRitase == 0) {
    emit("calcTotalPenumpangA", total);
  } else {
    emit("calcTotalPenumpangB", total);
  }
}

function calcTotalAmount(idxRitase) {
  const ritase = value.value[idxRitase] ?? [];
  const total = ritase.reduce((total, e) => { 
  
    total += parseInt(e.Amount ?? 0);
   
    return parseInt(total);
  }, 0);

  if (idxRitase == 0) {
    emit("calcTotalAmountA", total);
  } else {
    emit("calcTotalAmountB", total);
  }
}


watch(
  () => props.kurs,
  (nv) => { 
    if(props.trayekName === props.specialTrayekName){
      value.value.forEach((e,i)=>{
        e.forEach((r,i2) => {
          if(r.TerminalID === props.specialTerminalFrom) { 
           r.Passengers.forEach((x,i3)=>{ 
              allCalc(i,i2,i3)
           })
          } 
        })
      })
    }  
  }, 
);
function allCalc(i,i2,i3){
  calcIncome(i, i2, i3);
  calcTotalIncome(i, i2);
  calcTotalCombinedAmount(i)
  calcTotalPassenger(i);
  calcTotalAmount(i);
  calcTotalCombinedAmount(i);
  calcTotalAllPassenger(i);
}
function onInputPassenger(i, i2, i3, name, val) {
  util.nextTickN(2, () => {
    allCalc(i,i2,i3)
  });
}

onMounted(() => {
  util.nextTickN(2, () => {
    calcTotalAmount(0);
    calcTotalAllPassenger(0);
    calcTotalAmount(1);
    calcTotalAllPassenger(1);
  });
});
</script>
