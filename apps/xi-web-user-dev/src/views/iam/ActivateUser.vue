<template>
    <div class="w-full h-screen flex justify-center items-center">
        <div class="card">
            <s-form ref="loginFormCtl" 
                v-model="data.record" :config="data.formCfg" keep-label
                only-icon-top buttons-on-bottom :buttons-on-top="false"
                hide-cancel submit-text="Activate" submit-icon="login-variant" 
                @submitForm="activateUser" 
            auto-focus>
                <template #section_General_header>
                    <div class="mb-5">
                        Enter activation code we have sent to your email and press <b>"Activate"</b> button.
                    </div>
                </template>
                <template #buttons_2>
                    <s-button
                        icon="email" label="Resend Activation Code"
                        class="btn_warning"
                        @click="resendActivationEmail"
                    />
                </template>
            </s-form>
        </div>
    </div>
</template>

<script setup>
import { layoutStore } from '@/stores/layout'
import { onMounted, reactive, inject, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router';
import { authStore } from '@/stores/auth';
import { formInput, createFormConfig, SCard, SForm, SButton, util } from 'suimjs'
import { notifStore } from '@/stores/notif';

layoutStore().change("clear")

const loginFormCtl = ref(null)
const notif = notifStore()
const route = useRoute()

const data = reactive({
    formCfg: {},
    record: {},
    email: route.query?.Email?? null,
})

const axios = inject("axios")
const router = useRouter()
const auth = authStore()

function login(record, cb1, cb2) {
    axios.post("/iam/http-auth",
        { CheckName: "LoginID", SecondLifeTime: 60 * 60 * 6},   //-- timeout: 6 hours
        { auth: { username: record.LoginID, password: record.Password } }).then(r => {
            auth.updateJwt(r.data)
            router.push("/")
            cb1()
        }, e => {
            cb2()
            notif.add({kind:'error', message:e})
        })
}

function activateUser(_, cb) {
    const payload = {
        Email: data.email,
        Token: data.record.ActivationCode
    }
    axios.post('/iam/user/activate', payload).then(r => {
        cb();
        util.showInfo('Your user has been activated');
    }, e => {
        cb();
        util.showError(e)
    });
}

function resendActivationEmail() {
    const payload = {
        Email: data.email
    }
    axios.post('/iam/user/resend-activation-email', payload).then(r => {
        util.showInfo('email has been sent, please check your inbox');
    }, e => util.showError(e))
}

onMounted(() => {
    const cfg = createFormConfig("User Activation",true)
    const activationCodeInput = new(formInput)
    activationCodeInput.field = "ActivationCode"
    activationCodeInput.label = "Activation code"
    activationCodeInput.kind = "string"
    activationCodeInput.required = true

    cfg.addSection("General",false).addRowAuto(1, activationCodeInput)
    data.formCfg = cfg.generateConfig()
})

</script>