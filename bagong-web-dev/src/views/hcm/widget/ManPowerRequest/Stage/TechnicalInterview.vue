<template>
<Loader kind="skeleton" v-if="data.loading"/>
<div v-else>
    <status-text :txt="data.record.Status"/>
    <s-form ref="frmCtl" v-model="data.record" :config="data.frmCfg" keep-label
            hide-cancel  :buttons-on-bottom="false" buttons-on-top 
            @submit-form="submitForm" @field-change="formFieldChange">
        <template #input_Detail="{item}"> 
            <detail ref="detailCtl" :template-id="item.TemplateID" v-model="item.Detail" @calc-final-score="calcFinalScore"/>
        </template>
    </s-form>
</div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import moment from 'moment';
import StatusText from "@/components/common/StatusText.vue";
import Loader from "@/components/common/Loader.vue";
import Detail from "./TechnicalInterviewDetail.vue";

import { 
    util,
    SForm,
    SButton,
    loadFormConfig
} from "suimjs";

const props = defineProps({ 
    id: { type: String, default: '' },  
})

const axios = inject('axios');
const frmCtl = ref(null);
const detailCtl = ref(null);

const data = reactive({
    frmCfg:{},
    record:{},
    loading:false
})
watch(
    ()=>props.id,
    ()=>{
        fetchRecord()
    }
)
function calcGrade(){
    if(data.record.FinalScore  > 90) 
        data.record.Grade = "3C"
    else if(data.record.FinalScore  > 80)  
        data.record.Grade = "3B"
    else if(data.record.FinalScore  > 70)  
        data.record.Grade = "3A"
    else if(data.record.FinalScore  > 67) 
        data.record.Grade = "2C"
    else if(data.record.FinalScore  > 57) 
        data.record.Grade = "2B"
    else if(data.record.FinalScore  > 50)  
        data.record.Grade = "2A"
    else if(data.record.FinalScore  > 40) 
        data.record.Grade = "1D"
    else if(data.record.FinalScore  > 30) 
        data.record.Grade = "1C"
    else if(data.record.FinalScore  > 25) 
        data.record.Grade = "1B"
    else 
        data.record.Grade = "1B"
}

function calcFinalScore(total){
    data.record.FinalScore = total

       
    util.nextTickN(2, () => {
        calcGrade()
    })
}
function formFieldChange(name, v1,v2,old){
    if(name == "TemplateID"){
        data.record.FinalScore = 0
        data.record.Grade = "" 
    }
}
function fetchRecord(){
    data.loading = true
    axios.post("/hcm/techinal-interview/get", [props.id]).then((r)=>{
        const dt = r.data
        dt.Date = moment(dt.Date).local().year() >= 1901 ? dt.Date : null
        data.record = dt
    }).catch(e=>{
        util.showError(e)
    }).finally(()=>{
        data.loading = false
    })
}
function submitForm(record,cbOk, cbError){
    axios.post("/hcm/techinal-interview/update", record).then((r)=>{
        cbOk()
    }).catch(e=>{
        cbError()
        util.showError(e)
    })
}
onMounted(()=>{
    loadFormConfig(axios, "/hcm/techinal-interview/formconfig").
        then(r => {
            data.frmCfg = r 
            fetchRecord()
        }, e => util.showError(e))
})

</script>