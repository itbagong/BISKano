<template> 
    <table class="w-full"> 
        <thead>
            <tr class="bg-[#F7F8F9] [&>*]:p-2 [&>*]:border">
                <td>No</td>
                <td>Subject</td>
                <td class="text-center">Bobot%</td>
                <td>Description</td>
                <td class="text-center"> Score Mlc</td>
                <td class="text-center">Rata-rata nilai Mlc</td>
                <td class="text-center">Nilai</td>
            </tr>
        </thead>
        <tbody v-if="data.loading">
            <tr>
                <td colspan="7">
                    <div class="h-[300px] flex items-center justify-center" >
                        <Loader kind="circle"/>
                    </div>
                </td>
            </tr>
        </tbody>
        <tbody v-else>
            <tr v-for="(record, idx) in data.records" :key="idx" class="border [&>*]:p-2 [&>*]:border-r">
                <td valign="center" class="text-center  w-[40px] max-w-[70px]">{{idx + 1}}</td>
                <td valign="center" class="text-wrap  min-w-[200px " >{{record.Section}}</td>
                <td class="child-td text-wrap min-w-[200px]" valing="top" >
                    <div class="grid grid-cols-1   w-full  divide-y ">
                        <div v-for="(child, idx2) in record.Detail" :key="idx2" class="p-2" >
                            {{child.Weight}}
                        </div>
                    </div>
                </td>
                <!-- <td valign="center" class="text-right  w-[70px] max-w-[70px] font-semibold ">{{record.Weight}}</td> -->
                <td class="child-td text-wrap min-w-[200px]" valing="top" >
                    <div class="grid grid-cols-1   w-full  divide-y ">
                        <div v-for="(child, idx2) in record.Detail" :key="idx2" class="p-2" >
                            {{child.Description}}
                        </div>
                    </div>
                </td>
                <td class="child-td w-[90px] max-w-[90px]">
                    <div class="grid grid-cols-1  divide-y w-full ">
                        <div v-for="(child, idx2) in record.Detail" :key="idx2" class="px-2 py-1" > 
                            <s-input kind="number" class="w-[70px]" v-model="child.ScoreMlc" @change="(...args) => onChangeScoreMlc(idx, idx2, ...args)"  />   
                        </div>
                    </div>
                </td>
                <td class="child-td w-[70px] max-w-[70px]">
                    <div class="grid grid-cols-1 divide-y w-full ">
                        <div v-for="(child, idx2) in record.Detail" :key="idx2" class="p-2 text-right font-semibold" >
                            {{child.AverageMlc}}
                        </div>
                    </div>
                </td>

                <td valign="center "  class="text-right  w-[70px] max-w-[70px] font-semibold">{{record.Score}}</td>

            </tr>
        </tbody>
    </table>
</template>

<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import { 
    util, 
    SInput
} from "suimjs";
import Loader from "@/components/common/Loader.vue";

const props = defineProps({ 
    templateId: { type: String, default: '' },  
    modelValue: { type: Array, default: []},
});

const emit = defineEmits({ 
    "update:modelValue": null,
    calcFinalScore: null
});

const axios = inject('axios');

const data = reactive({
    records: props.modelValue,
    loading: false
})

function mappingMcuItemTemplateLines(lines){
    const groupLines = Object.groupBy(lines, (obj) => obj.Parent);
   
    function createLine(el){
        return el.reduce((r, e) => { 
            let Detail = buildLines(e.ID);
            if(e.Parent == ""){
                r.push({
                    ...e, 
                    Weight: e.AnswerValue,
                    Score: 0,
                    Section: e.Description,
                    Detail
                });
            }else{
                 r.push({
                    ...e,  
                    Weight: e.AnswerValue,
                    AverageMlc:0,
                    ScoreMlc:0,
                    Detail
                });
            }
            return r;
        }, [])
    }

    function buildLines(id) { 
        let arr = groupLines[id]
        if (arr === undefined) return [];
        return  createLine(arr)
    }

    const r = buildLines("");
    return r
}

function reload(){
    if(props.templateId == '' || props.templateId == undefined){
        data.records = []
        return
    }
    data.loading = true
    axios.post("/she/mcuitemtemplate/get", [props.templateId]).then(r=>{
        data.records = mappingMcuItemTemplateLines(r.data.Lines)
    }).catch(e=>{
        data.records = []
        util.showError(e)
    }).finally(()=>{
        data.loading = false
    })
}

function calcAverageMlc(idx, idxChild, v1){
    if (v1 !== null && v1 !== undefined) {
        data.records[idx].Detail[idxChild].AverageMlc = (v1 / 4)  * data.records[idx].Detail[idxChild].Weight
    }
}

function calcParentScore(idx){
    const total = data.records[idx].Detail.reduce((r,e)=>{
        r += e.AverageMlc
        return  r
    },0)
   data.records[idx].Score = total
}

function calcFinalScore(){
    const total = data.records.reduce((r,e)=>{
        r += e.Score
        return  r
    },0) / data.records.length
    emit("calcFinalScore", total)
    

}

function onChangeScoreMlc(idx, idxChild, field, v1, v2, old){
    let _v1 = v1 == '' ? 0 : v1
    if(_v1 > 4){
        util.nextTickN(2, () => {
            data.records[idx].Detail[idxChild].ScoreMlc = 4
        })
        _v1 = 4
    }
    calcAverageMlc(idx, idxChild, _v1)
    calcParentScore(idx)
    calcFinalScore()
} 

defineExpose({
    reload
})

watch(
    ()=>props.templateId,
    (nv)=>{
        reload()
    }
)

watch(
    ()=>data.records,
    (nv)=>{
        emit("update:modelValue",data.records)
    }
)

</script>

<style scoped>
.child-td {
    padding:  0 !important;
}
</style>