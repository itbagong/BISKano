<template>
<Loader kind="skeleton" v-if="data.loading"/>
<div v-else>
    <status-text :txt="data.record.Status"/>
    <s-form ref="frmCtl" v-model="data.record" :config="data.frmCfg" keep-label
            hide-cancel  :buttons-on-bottom="false" buttons-on-top 
            @submit-form="submitForm">
        <template #input_Notes="{item}">
            <div class="grid grid-cols-1 gap-x-2 border-t">
            <div class="font-semibold mb-3 tex-[1.2em]">
                Notes
            </div>
            <div class="w-full"> 
                <s-list-editor
                    no-gap
                    :config="{}"
                    v-model="item.Notes" 
                    allow-add
                    hide-select
                    @validate-item="addNote"
                    ref="notesEditorCtl"
                >
                    <template #header="">
                        <div class="w-full grid grid-cols-1 bg-[#F7F8F9] ">
                            <div class="grid grid-cols-2 gap-x-4 bg-transparent w-full border-b p-2">
                                <div>Interviewer</div>
                                <div>Note</div>
                            </div>
                        </div>
                    </template>
                    <template #item="{ item }">
                        <div class="grid grid-cols-2 gap-4 w-full ">
                            <s-input
                                v-model="item.EmployeeID"
                                kind="text" 
                                class="w-full mt-3"
                                label=""
                                use-list
                                keep-label
                                lookup-url="/tenant/employee/find"
                                lookup-key="_id"
                                :lookup-labels="['Name']"
                                :lookup-searchs="['_id', 'Name']" 
                            />
                            <s-input v-model="item.Note" multi-row="4" keep-label/>
                        </div>
                    </template>
                </s-list-editor>
            </div>
            </div>
        </template>
    </s-form>
</div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import moment from 'moment';
import StatusText from "@/components/common/StatusText.vue";
import Loader from "@/components/common/Loader.vue";

import { 
    util,
    SInput,
    SForm,
    SButton,
    loadFormConfig,
    SListEditor
} from "suimjs";

const props = defineProps({ 
    id: { type: String, default: '' },  
})

const axios = inject('axios');
const frmCtl = ref(null);
const notesEditorCtl  = ref(null)

const data = reactive({
    frmCfg:{},
    record:{},
    loading:"",
})
watch(
    ()=>props.id,
    ()=>{
        fetchRecord()
    }
)
function addNote(){
  notesEditorCtl.value.setValidateItem(true)
}
function fetchRecord(){
    data.loading = true

    axios.post("/hcm/interview/get", [props.id]).then((r)=>{
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
    axios.post("/hcm/interview/update", record).then((r)=>{
        cbOk()
    }).catch(e=>{
        cbError()
        util.showError(e)
    })
}
onMounted(()=>{
    loadFormConfig(axios, "/hcm/interview/formconfig").
        then(r => {
            data.frmCfg = r 
            fetchRecord()
        }, e => util.showError(e))
})

</script>