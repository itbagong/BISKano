<template>
    <div autocomplete="off">
    <s-card hide-footer hide-title class="card">
       <s-form v-if="data.config"
        :buttons-on-top="false" buttons-on-bottom hide-cancel submit-text="Change Password"
        :config="data.config" v-model="data.record" 
        @submit-form="changePassword">
        <template #section_General_header>
            <div class="mb-5">
                To reset password, please enter token we sent to you and your new password on below form.
            </div>
        </template>
        <template #footer_2>
            <div class="mt-2 text-right text-primary cursor-pointer hover:text-secondary" @click="router.push('/login')">I remember my credentials, bring me back to Login Page</div>
        </template>
        </s-form>
    </s-card>
    </div>
</template>


<script setup>
import { layoutStore } from '@/stores/layout';
import { SCard, SForm, createFormConfig, formInput, util } from 'suimjs';
import { onMounted } from 'vue';
import { ref } from 'vue';
import { reactive } from 'vue';
import { inject } from 'vue';
import { useRoute, useRouter } from 'vue-router';

const route = useRoute();
const router = useRouter();
layoutStore().name = 'clear';

const axios = inject('axios');
const emailCtl = ref(null);
const data = reactive({
    config: null,
    record: {
        TokenID: route.query.uid,
        Token: '',
        Password: '',
        ConfirmPassword: ''
    }
});

onMounted(() => {
    const cfg = createFormConfig("Reset Password",true)
    const loginID_input = new(formInput)
    loginID_input.field = "Token"
    loginID_input.label = "Token"
    loginID_input.kind = "string"
    loginID_input.required = true

    const loginPassword_input = new(formInput)
    loginPassword_input.field = "Password"
    loginPassword_input.label = "Password"
    loginPassword_input.kind = "password"
    loginPassword_input.required = true

    const confirmPassword_input = new(formInput)
    confirmPassword_input.field = "ConfirmPassword"
    confirmPassword_input.label = "Confirm Password"
    confirmPassword_input.kind = "password"
    confirmPassword_input.required = true
    confirmPassword_input.rules = [(v) => {
        return v==data.record.Password ? '' : 'confirm password should be equal with the Password';
    }]

    cfg.addSection("General",false).addRowAuto(1, loginID_input, loginPassword_input, confirmPassword_input)
    data.config = cfg.generateConfig()  
})

function changePassword (_, cb) {
    axios.post('/iam/user/reset-password', data.record).
        then(r => {
            cb();
            router.push('/login');
        }, 
        e => {
            cb();
            util.showError(e)
        });
}

</script>