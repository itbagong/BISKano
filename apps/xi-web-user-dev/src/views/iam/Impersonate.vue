<template>
    <div class="w-full">
        <div class="card" v-if="auth.appData.OriginalUserID && auth.appData.OriginalUserID!=''">
            You are doing impersonation already
        </div>
        <div v-else class="card">
            <s-form ref="loginFormCtl" v-model="data.record" :config="data.formCfg" keep-label only-icon-top
                buttons-on-bottom :buttons-on-top="false" hide-cancel submit-text="Impersonate" submit-icon="login-variant"
                @submitForm="impersonate" auto-focus>
            </s-form>
        </div>
    </div>
</template>

<script setup>
import { layoutStore } from '@/stores/layout';
import { SForm, createFormConfig, formInput, util } from 'suimjs';
import { inject } from 'vue';
import { onMounted, reactive, ref } from 'vue';
import { authStore } from '@/stores/auth';

const axios = inject("axios");
const auth = authStore();

layoutStore().change("tenant");

const data = reactive({
    formCfg: null,
    record: {
        UserID: null
    }
});

onMounted(() => {
    const cfg = createFormConfig("Impersonate", true);
    const selectUser_input = new (formInput);
    selectUser_input.field = "UserID";
    selectUser_input.label = "Select user to be impersonated";
    selectUser_input.kind = "string";
    selectUser_input.useList = true;
    selectUser_input.lookupUrl = "/iam/user/find";
    selectUser_input.lookupKey = "_id";
    selectUser_input.lookupLabels = ["DisplayName","Email"];
    selectUser_input.lookupSearchs = ["DisplayName","Email"];
    selectUser_input.required = true;

    cfg.addSection("General", false).addRowAuto(1, selectUser_input)
    data.formCfg = cfg.generateConfig()
});


function impersonate(_, fnOK) {
    axios.post("/iam/impersonate", data.record.UserID).then(r => {
        auth.updateJwt(r.data);
        util.showInfo(`Impersonate as ${data.record.UserID}`)
        fnOK();
    }, e => {
        util.showError(e);
        fnOK();
    })
}

</script>