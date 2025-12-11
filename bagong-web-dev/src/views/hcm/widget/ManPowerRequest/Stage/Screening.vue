<template>
<Loader kind="skeleton" v-if="data.loading"/>
<div v-else>
    <status-text :txt="data.record.Status"/>
     <s-form ref="frmCtl" v-model="data.record" :config="data.frmCfg" keep-label
            hide-cancel  :buttons-on-bottom="false" buttons-on-top 
            @submit-form="submitForm">
        
    </s-form>
</div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";

import StatusText from "@/components/common/StatusText.vue";
import Loader from "@/components/common/Loader.vue";

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

const data = reactive({
    frmCfg:{},
    record:{},
    loading:""
})
watch(
    ()=>props.id,
    ()=>{
        fetchRecord()
    }
)
function fetchRecord(){
    data.loading = true
    axios.post("/hcm/screening/get", [props.id]).then((r)=>{
        data.record = r.data
    }).catch(e=>{
        util.showError(e)
    }).finally(()=>{
        data.loading = false
    })
}
function submitForm(record,cbOk, cbError){
    axios.post("/hcm/screening/update", record).then((r)=>{
        cbOk()
    }).catch(e=>{
        cbError()
        util.showError(e)
    })
}
onMounted(()=>{
    loadFormConfig(axios, "/hcm/screening/formconfig").
        then(r => {
            data.frmCfg = r 
            fetchRecord()
        }, e => util.showError(e))
})

</script>