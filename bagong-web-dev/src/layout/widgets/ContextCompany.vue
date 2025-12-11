<template>
<div class="text-black">
    <loader kind="skeleton" skeleton-kind="input" v-if="data.loading"/>
    <template v-else>
        <div v-if="viewType == 'full'" class="flex justify-between items-center hover:text-primary  cursor-pointer " @click="changeCompany">
            {{ data.companyName != '' ? data.companyName : 'No Company'}}
            <mdicon name="pencil" width="14" class="pr-2 " />
        </div>
        <div v-else class="hover:text-primary cursor-pointer" @click="changeCompany">
            <template v-if=" data.companyName != '' ">
                {{data.companyName.substring(0, 2)+'..'}}
            </template>
            <div v-else class="flex justify-center">
            <mdicon name="minus" width="20"/>
            </div> 
        </div>
    </template>
</div>
</template>
<script setup>
import {watch,onMounted,reactive,inject} from 'vue'
import { util,SButton } from 'suimjs';
import { authStore } from "@/stores/auth";
import Loader from "@/components/common/Loader.vue";

const auth = authStore();
const axios = inject("axios")
const props = defineProps({ 
  viewType: { type: String, default: "full" },
});
const emit = defineEmits({
    change: null,
})

const data = reactive({
    companyName:"", 
    loading: false
})
function getCompany(){
    if(auth.appData?.CompanyID == '' || auth.appData?.CompanyID == undefined){
        data.companyName = ""
        return
    }
    data.loading  = true
    axios.post("/tenant/company/get",[auth.appData.CompanyID]).then(co => {
       data.companyName = co.data.Name
    }, e => util.showError(e)).finally(()=>{ data.loading = false})
}
function changeCompany(){
    emit("change")
}

watch(() => auth.appData?.CompanyID, (nv) => { 
    getCompany()
}, { deep: true })
onMounted(() => {
    getCompany()
})
</script>