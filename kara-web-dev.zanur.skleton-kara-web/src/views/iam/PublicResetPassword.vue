<template>
    <div>
        <div class="card flex flex-col gap-2" v-if="!data.requested">
            <h2 class="text-primary">Change Password Request</h2>
            <div>
                Please enter email you used to register and click submit
            </div>
            <s-input v-model="data.email" kind="email" ref="emailCtl" hide-label caption="Please enter email"
                :rules="[isEmail]" />
            <div class="flex gap-2 items-center">
                <s-button class="bg-primary text-white" label="Remind me" @click="remindMe" />
                <div class="text-primary cursor-pointer hover:text-secondary" @click="router.push('/login')">I already
                    remember my credential</div>
            </div>
        </div>

        <div v-else class="bg-white p-2 rounded-md shadow border-slate-400 w-[450px] flex flex-col gap-2">
            <h2 class="text-primary">Request Sent !</h2>
            <div>
                Your password change request has been sent. If your email is valid user, we will send a message to recover
                your password.
            </div>
            <div class="flex gap-2 items-center">
                <div class="text-primary cursor-pointer hover:text-secondary" @click="router.push('/login')">Bring me back
                    to login page</div>
            </div>
        </div>
    </div>
</template>


<script setup>
import { layoutStore } from '@/stores/layout';
import { SInput, SButton, util } from 'suimjs';
import { ref } from 'vue';
import { reactive } from 'vue';
import { inject } from 'vue';
import { useRouter } from 'vue-router';

layoutStore().name = 'clear';

const router = useRouter();
const axios = inject('axios');
const emailCtl = ref(null);
const data = reactive({
    email: '',
    requested: false,
});

function isEmail(v) {
    if (v.match('.*@.*\..*')) {
        return '';
    } else {
        return 'is not valid email';
    }
}

function remindMe() {
    const valid = emailCtl.value.validate();
    if (!valid) {
        return;
    };

    axios.post('/iam/user/user-reset-password', { Email: data.email }).
        then(
            r => {
                data.requested = true;
            },
            e => util.showError(e));
}

</script>