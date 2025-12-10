<template>
    <div class="w-full">
        <div class="card flex flex-col gap-2">
            <h1 class="text-primary">Select Tenant</h1>

            <div v-if="data.tenants.length == 0" class="flex flex-col gap-2">
                <div>
                    You don't have any tenant assigned to your account.
                </div>
                <div>
                    You can create nenw tenant <i>(if you have proper license to do so)</i>,
                    or you can request access to any tenant.
                </div>
                <div>
                    To request access you can either by go to <i>"tenant landing page"</i> or share your login ID to be
                    invited by tenant owner.
                </div>
                <div class="flex gap-2">
                    <s-button label="Create new tenant" icon="domain" class="btn_secondary"
                        @click="router.push('/me/create-tenant')" />
                    <s-button label="Request to join tenant" icon="plus" class="btn_secondary"
                        @click="router.push('/me/join-tenant')" />
                </div>
            </div>

            <div v-else class="flex flex-col gap-2">
                <div class="flex flex-col gap-1">
                    <div v-for="tenant in data.tenants" class="flex gap-2 items-center hover:bg-slate-100">
                        <div class="grow">{{ tenant.Name }}</div>
                        <s-button v-if="data.currentTenant != tenant._id" label="Select" class="btn_primary"
                            @click="selectTenant(tenant)"></s-button>
                    </div>
                </div>
                <div class="flex gap-2">
                    <s-button label="Create new tenant" icon="domain" class="btn_secondary"
                        @click="router.push('/me/create-tenant')" />
                    <s-button label="Request to join tenant" icon="plus" class="btn_secondary"
                        @click="router.push('/me/join-tenant')" />
                </div>
            </div>
        </div>
    </div>

    <validate-twofa v-model="data"></validate-twofa>
</template>

<script setup>
import { SButton, util } from 'suimjs';
import { layoutStore } from '@/stores/layout';
import { authStore } from '@/stores/auth';
import { reactive } from 'vue';
import { onMounted } from 'vue';
import { inject } from 'vue';
import { useRouter } from 'vue-router';
import ValidateTwofa from "./widget/ValidateTwofa.vue";

layoutStore().change('tenant');

const auth = authStore();
const axios = inject('axios');
const data = reactive({
    modalTwofa: false,
    tenants: [],
    currentTenant: auth.appData ? auth.appData.TenantID : '',
    selectedTenant: {},
})

const router = useRouter();

onMounted(() => {
    axios.post('/iam/tenant/my').then(r => {
        data.tenants = r.data
    });
})

function selectTenant(tenant) {
    data.selectedTenant = tenant
    axios.post("/iam/user/change-tenant", tenant).
        then(r => {
            data.currentTenant = tenant._id;
            auth.updateJwt(r.data);
        }, e => {
            if (e == "Use2FA is required") {
                data.modalTwofa = true
            }
            util.showError(e)
        })
}

</script>