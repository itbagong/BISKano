<template>
    <div>
        <div class="w-full  mt-4 grid grid-cols-1 gap-y-2"> 
            <loader  v-if="data.loading" kind="skeleton" skeleton-kind="input"/>
            <template v-else>
                <template v-if="stage == 'PshycologicalTest'">
                    <s-button 
                        no-tooltip
                        :disabled="disabled" 
                        @click="handleSendPsikotest"
                        icon="send" 
                        class="btn_warning back_btn "
                        label="Send Psikotes" 
                    />
                </template>
                <template v-else-if="stage == 'MCU'">
                    <s-button 
                        no-tooltip
                        :disabled="disabled" 
                        @click="handleAddMCU"
                        icon="plus" 
                        class="btn_warning back_btn "
                        label="Create New MCU" 
                    />
                </template> 
                <template v-else-if="stage == 'OLPlotting'">
                    <s-button 
                        no-tooltip 
                        @click="handlePlotting"
                        icon="map-marker-account" 
                        class="btn_warning back_btn "
                        label="Plotting" 
                    />
                </template>
                <div class="flex gap-2">
                    <s-button  
                        no-tooltip
                        @click="handleAction('Failed')"
                        :disabled="disabled"
                        icon="close"
                        class="btn_error btn_failed w-full "
                        label="Mark as Failed" 
                    />
                    <s-button  
                        no-tooltip
                        @click="handleAction('Passed')"
                        :disabled="disabled"
                        icon="check"
                        class="btn_primary btn_passed w-full "
                        label="Mark as Passed" 
                    />
                </div>
            </template>
        </div>
        <!-- <modal-send-psikotest v-model="data.showModalSendPsikotest" @submit="sendPsikotest"/> -->
    </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import { useRouter } from "vue-router";
import {
    util,
    SInput, 
    SButton,
    SModal
} from "suimjs";

import Loader from "@/components/common/Loader.vue";
// import ModalSendPsikotest from "./ModalSendPsikotest.vue";

const props = defineProps({ 
    stage: { type: String, default: '' },
    manPowerId: { type: String, default: '' },
    mapIds: {type: String, default:[]},
})

const emit = defineEmits({ 
    refreshList:null
});

const axios = inject("axios");
const router = useRouter();

const data = reactive({ 
    loading: false,
    showModalSendPsikotest: false, 
})
const disabled = computed({
    get () {
        return Object.keys(props.mapIds).length == 0
    }
})
function handlePlotting(){
    const url = router.resolve({
        path: "/hcm/Plotting",
        query: { JobVacancyID: props.manPowerId },
    });
    window.open(url.href, "_blank");
}
function handleSendPsikotest(){
    // data.showModalSendPsikotest = true
    sendPsikotest()
}
function sendPsikotest(){
     const param  = {
        JobID: props.manPowerId,
        CandidateID: Object.keys(props.mapIds).map(e=>{ 
            return  props.mapIds[e].CandidateID
        }),
    }  
   
    data.loading = true
    axios.post("/hcm/tracking/send-psychological-test",param).then(r=>{
        emit("refreshList")
    }).catch(e=>{
        util.showError(e)
    }).finally(()=>{
        data.loading  = false
    })
}
function handleAddMCU(){
    const param  = {
        JobID: props.manPowerId,
        CandidateID: Object.keys(props.mapIds).map(e=>{ 
            return  props.mapIds[e].CandidateID
        }), 
    }  
   
    
    data.loading = true
    axios.post("/hcm/tracking/save-mcu",param).then(r=>{
        emit("refreshList")
    }).catch(e=>{
        util.showError(e)
    }).finally(()=>{
        data.loading  = false
    })

}
function handleAction(status){
    const param  = {
        JobID: props.manPowerId,
        StageID: Object.keys(props.mapIds).map(e=>{ 
            return  props.mapIds[e]._id
        }),
        Stage: props.stage,
        Status: status
    } 
   
    data.loading = true
    axios.post("/hcm/tracking/mark-status",param).then(r=>{
        emit("refreshList")
    }).catch(e=>{
        util.showError(e)
    }).finally(()=>{
        data.loading  = false
    })

}
 
</script>
<style>
.btn_passed button{
    @apply bg-green-500
}

</style>