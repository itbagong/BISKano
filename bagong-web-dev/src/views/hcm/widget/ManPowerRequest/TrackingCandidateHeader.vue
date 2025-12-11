<template> 
<div class="grid grid-cols-1  gap-y-2">
    
    <s-input kind="text" label="Search" class="input_search_candidate" v-model="data.item.Name" 
          @keyup.enter="apply"/>
    <div class="flex justify-between items-center">
        <div class="flex gap-1 items-center">
            <input type="checkbox" id="all" value="all"  @click="checkUncheck"/>
            <label for="all">Select All</label>
        </div>
        <s-button icon="filter" tooltip="Filter" @click="data.showDetail = !data.showDetail" :class="[data.showDetail ? 'bg-slate-100 text-slate-900':'']"/>
    </div>
    <div v-if="data.showDetail" class="grid grid-cols-1 gap-y-2 border-t">
        <div class="w-full grid grid-cols-1 gap-y-2 mb-2 border-b" v-if="stage == 'Screening'">
            <label class="input_label">
                Age
            </label> 
            <div class="flex gap-1">
                <label class="input_label  min-w-[30px]">
                    Min
                </label>
                <input type="range" min="1" max="100"   v-model="data.item.AgeFrom" class="w-full" />
                <div class=" min-w-[20px] text-right text-[0.8em]">{{data.item.AgeFrom == null ? '-' :data.item.AgeFrom}}</div>
            </div>
            <div class="flex gap-1">
                <label class="input_label min-w-[30px]">
                    Max
                </label>
                <input type="range" min="1" max="100"   v-model="data.item.AgeTo" class="w-full" />
                <div class=" min-w-[20px] text-right text-[0.8em]">{{ data.item.AgeTo == null ? '-' :data.item.AgeTo}}</div>
            </div>
            
        </div>
        <s-input  
            label="Status"
            use-list
            multiple
            :items="['Passed', 'Not Selected', 'Failed']"
            v-model="data.item.Status"
        />
        <s-input  
            label="Domicile"
            use-list
            multiple
            :items="[]"
            v-model="data.item.Domicile"
        />  
        <div class="grid grid-cols-2 gap-2">
            <s-button icon="delete" label="Clear" class="btn_error" @click="clear"/>
            <s-button icon="magnify" label="Apply" class="btn_apply" @click="apply"/>
        </div>  
    </div>
</div>
</template>
<script setup>

import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import {
    SInput, 
    SButton
} from "suimjs";
const props = defineProps({
    stage: { type: String, default: '' },
})
const emit = defineEmits({
    checkUncheck: null, 
    apply:null
});
const data = reactive({
    item: {Name:"", AgeFrom:null, AgeTo:null, Domicile:[], Status:[]},
    showDetail:false, 
})

function checkUncheck(){
    emit("checkUncheck")
}
function apply(){ 
    emit('apply',data.item)
}
function clear(){
    data.item = {Name:"", AgeFrom:null, AgeTo:null, Domicile:[], Status:[]}
    emit('apply',null)
}
// watch(
//   () => JSON.stringify(data.item),
//   (nv)=>{
//     const _nv = JSON.parse(nv)
//     data.isChange = true
//   }
// )
</script>
<style>
.input_search_candidate label.input_label{
    display: none !important;
}
</style>
<style>
.btn_apply button{
    @apply bg-green-600 text-white;
}
</style>