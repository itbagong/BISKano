<template>
    <div class="w-full">
        <div class="card max-w-[600px] break-all">
            {{  auth.appData?.Email }}
            <!-- <br/><br/>
            {{  auth.appToken }}
            <br/><br/> -->
            <br/>
            <b>Expiry</b> : {{ moment(data.expiry).format("DD-MMM-yyyy hh:mm:ss")   }}
            <br/><br/> 
            <s-button icon='key-change' label='Change Password' @click="changePassword" class='btn_primary'/>
        </div>
    </div>
</template>

<script setup>
import { authStore } from '@/stores/auth';
import { layoutStore } from '@/stores/layout';
import { onMounted, reactive, inject } from 'vue';
import { SButton,util} from 'suimjs';
import moment from "moment"

layoutStore().name = 'tenant';
const auth = authStore();
const axios = inject("axios")
const data = reactive({
    appData: {},
    expiry: null
})
function changePassword(){
    axios.post('/iam/user/request-reset-password', {AppID:layoutStore().appID}).then(r => {
        // console.log(r)
        alert("Please check your email")
    },e=>util.showError(e))
}
onMounted(() => {
    data.expiry = auth.tokenExpiry;
})

</script>