<template>
    <s-modal :display="props.modelValue.modalTwofa" hideButtons title="Two Factor Authentication (2FA)">
        <s-card class="rounded-md w-full" hide-title>
            <div class="px-2 w-[440px]">
                <div class="mb-2">
                    Enter validation code we have sent to your email.
                </div>

                <s-input class="min-w-[300px]" ref="tenantSelector" v-model="data.validationCode" label="Validation Code" />

                <div class="flex gap-4 justify-center p-3">
                    <s-button class="bg-primary text-white px-4 justify-center" label="Validate" @click="validate()">
                    </s-button>
                </div>
            </div>
        </s-card>
    </s-modal>
</template>
  
<script setup>
import { computed, reactive, inject, onMounted } from 'vue';
import { SButton, SModal, SInput, util } from "suimjs";
import { notifStore } from '@/stores/notif';
import { useRouter } from 'vue-router';
import { authStore } from '@/stores/auth';


const axios = inject("axios")
const router = useRouter()
const auth = authStore()
const notif = notifStore()

const props = defineProps({
    modelValue: { type: Object, default: () => { } },
    modal: { type: Boolean, default: false },
});

const emit = defineEmits({
    "update:modelValue": null,
});

const value = computed({
    get() {
        return props.modelValue;
    },

    set(v) {
        emit("update:modelValue", v);
    },
});

const data = reactive({
    validationCode: "",
});


function validate() {
    axios.post('/iam/user/validate-twofa', data.validationCode).then(r => {
        var msg = 'Successfully'
        if (r.data == "LoginID") {
            axios.post("iam/http-auth",
                { CheckName: "LoginID", SecondLifeTime: 60 * 60 * 6 },   //-- timeout: 6 hours
                { auth: { username: props.modelValue.record.LoginID, password: props.modelValue.record.Password } }).then(r => {
                    auth.updateJwt(r.data)
                    router.push("/")
                    props.modelValue.modalTwofa = false
                }, e => {
                    notif.add({ kind: 'error', message: e })
                })
        } else {
            axios.post("/iam/user/change-tenant", props.modelValue.selectedTenant).
                then(r => {
                    props.modelValue.currentTenant = props.modelValue.selectedTenant._id;
                    auth.updateJwt(r.data);
                    props.modelValue.modalTwofa = false
                }, e => {
                    if (e == "Use2FA is required") {
                        data.modalTwofa = true
                    }
                    util.showError(e)
                })
        }
        util.showInfo(msg);
    }, e => {
        util.showError(e)
    });
}

onMounted(() => {

})
</script>