<template>
    <div class="w-full">
        <div class="card" v-if="auth.appData && auth.appData.OriginalUserID==undefined">
            You are not in impersonation mode
        </div>
        <div v-else class="card">
            <s-form ref="loginFormCtl" v-model="data.record" :config="data.formCfg" keep-label only-icon-top
                buttons-on-bottom :buttons-on-top="false" hide-cancel submit-text="De-Impersonate" submit-icon="login-variant"
                @submitForm="deimpersonate" auto-focus>
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
    const cfg = createFormConfig("De-Impersonate", true);
    const selectUser_input = new (formInput);
    selectUser_input.field = "UserID";
    selectUser_input.label = "Enter your original userID";
    selectUser_input.kind = "string";
    selectUser_input.required = true;

    cfg.addSection("General", false).addRowAuto(1, selectUser_input)
    data.formCfg = cfg.generateConfig()
});


function deimpersonate(_, fnOK) {
    const origUserID = auth.appData.OriginalUserID;
    if (origUserID != data.record.UserID) {
        util.showError("invalid original userID");
    }

    axios.post("/iam/deimpersonate", data.record.UserID).then(r => {
        auth.updateJwt(r.data);
        fnOK();
    }, e => {
        util.showError(e);
        fnOK();
    })
}

</script>