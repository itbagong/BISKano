<template>
    <div>
        <div class="card">
            <s-form ref="loginFormCtl" v-model="data.record" :config="data.formCfg" keep-label only-icon-top
                buttons-on-bottom :buttons-on-top="false" hide-cancel submit-text="Login" submit-icon="login-variant"
                @submitForm="login" auto-focus>
                <template #section_General_header>
                    <div>&nbsp;</div>
                </template>
                <template #footer_2>
                    <div class="flex flex-col gap-2 mt-4 items-center pb-5">
                        <a class="btn_text_primary" @click="router.push('public-reset-password')">I forget
                            my credentials, please help me</a>
                        <a class="btn_text_primary" @click="router.push('register')">I don't have
                            account and want to create new one</a>
                    </div>
                </template>
            </s-form>
        </div>

        <validate-twofa v-model="data"></validate-twofa>
    </div>
</template>

<script setup>
import { layoutStore } from '@/stores/layout'
import { onMounted, reactive, inject, ref } from 'vue'
import { useRouter } from 'vue-router';
import { authStore } from '@/stores/auth';
import { formInput, createFormConfig, SCard, SForm } from 'suimjs'
import { notifStore } from '@/stores/notif';
import ValidateTwofa from "./widget/ValidateTwofa.vue";

layoutStore().change("clear")

const loginFormCtl = ref(null)
const notif = notifStore()

const data = reactive({
    formCfg: {},
    record: {},
    validateTwofa: {
        modal: false,
    },
    modalTwofa: false,
})

const axios = inject("axios")
const router = useRouter()
const auth = authStore()

function login(record, cb1, cb2) {
    axios.post("iam/http-auth",
        { CheckName: "LoginID", SecondLifeTime: 60 * 60 * 6 },   //-- timeout: 6 hours
        { auth: { username: record.LoginID, password: record.Password } }).then(r => {
            auth.updateJwt(r.data)
            router.push("/")
            cb1()
        }, e => {
            cb2()
            notif.add({ kind: 'error', message: e })
            if (e == "Use2FA is required") {
                data.modalTwofa = true
            }
        })
}

onMounted(() => {
    const cfg = createFormConfig("Login Form", true)
    const loginID_input = new (formInput)
    loginID_input.field = "LoginID"
    loginID_input.label = "Login ID"
    loginID_input.kind = "string"
    loginID_input.required = true

    const loginPassword_input = new (formInput)
    loginPassword_input.field = "Password"
    loginPassword_input.label = "Password"
    loginPassword_input.kind = "password"
    loginPassword_input.required = true

    cfg.addSection("General", false).addRowAuto(1, loginID_input, loginPassword_input)
    data.formCfg = cfg.generateConfig()
})

</script>