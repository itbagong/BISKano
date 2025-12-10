<template>
    <div>
        <div class="card w-[300px] flex flex-col gap-2">
            <h1>Register New User</h1>
            <s-form v-if="data.mode=='register'" ref="regForm"
                v-model="data.record" :config="data.config" keep-label
                hide-cancel submit-text="Register" submit-icon="account-plus"
                buttons-on-bottom :buttons-on-top="false"
                @submit-form="registerUser"
            >
                <template #footer_1>
                    <div class="flex flex-col gap-2">
                        <div>
                            By submitting this form, I declare that I am:
                        </div>
                        <ul class="list-outside list-disc ml-5">
                            <li>Already 17 years old or above</li>
                            <li>Complete this form without any force by others</li>
                            <li>Not a citizen of country in conflict</li>
                        </ul>
                        <div>
                            Has been read and fully understand "Terms of Conditions", 
                            and I agree with all points written there.
                        </div>
                    </div>
                </template>
            </s-form>
            <div v-else-if="data.mode=='info'">
                Registration for {{  data.record.Email }} has been sucessfully done. Please check your inbox for further instruction.
            </div>
        </div>
    </div>
</template>


<script setup>
import { layoutStore } from '@/stores/layout';
import { SForm, loadFormConfig, util } from 'suimjs';
import { inject } from 'vue';
import { onMounted, reactive, ref } from 'vue';

layoutStore().change("tenant");

const regForm = ref(null);
const axios = inject('axios');
const data = reactive({
    mode: 'register',
    record: {},
    config: {},
});

onMounted(() => {
    loadFormConfig(axios, "/iam/ui/register/formconfig").
        then(r => {
            data.config = r
            util.nextTickN(2, () => {
                regForm.value.setFieldAttr('Email','rules',[
                    (email) => {
                        const regex = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|'".+'")@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
                        return regex.test(email) ? '' : 'not valid email address';
                    }
                ])
            });
        }, e => util.showError(e))
});

function registerUser(_, cb) {
    axios.post("/iam/user/create",data.record).then(r => {
        data.mode = 'info';
        cb();
    }, e => {
        cb();
        util.showError(e);
    })
}

</script>