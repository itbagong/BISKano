<template>
<Loader kind="skeleton" v-if="data.loading"/>
<div v-else>
    <status-text :txt="data.record.Status"/>
     <s-form ref="frmCtl" v-model="data.record" :config="data.frmCfg" keep-label
            hide-cancel  :buttons-on-bottom="false" buttons-on-top 
            :tabs="['General','Attachment']"
            @submit-form="submitForm">
        <template #tab_Attachment="{item}">
            <s-grid-attachment 
                ref="attchCtl"
                :journalId="item._id"
                journalType="HCMPKWTTs" 
                is-single-upload
            />
        </template>
    </s-form>
</div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";

import moment from 'moment';

import SGridAttachment from "@/components/common/SGridAttachment.vue";
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
const attchCtl = ref(null)

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
    axios.post("/hcm/pkwtt/get", [props.id]).then((r)=>{
        const dt = r.data
        dt.JoinedDate = moment(dt.JoinedDate).local().year() >= 1901 ? dt.JoinedDate : null
        dt.ExpiredContractDate = moment(dt.ExpiredContractDate).local().year() >= 1901 ? dt.ExpiredContractDate : null
        data.record = dt
    }).catch(e=>{
        util.showError(e)
    }).finally(()=>{
        data.loading = false
    })
}
function submitForm(record,cbOk, cbError){
    axios.post("/hcm/pkwtt/update", record).then((r)=>{
        attchCtl.value?.Save();
        cbOk()
    }).catch(e=>{
        cbError()
        util.showError(e)
    })
}
onMounted(()=>{
    loadFormConfig(axios, "/hcm/pkwtt/formconfig").
        then(r => {
            data.frmCfg = r 
            fetchRecord()
        }, e => util.showError(e))
})

</script>