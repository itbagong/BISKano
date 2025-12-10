<template>
    <div class="w-full flex flex-col gap-5">
        <div class="card flex flex-col gap-2">
            <h1 class="text-primary">Join Tenant</h1>
            <s-form v-if="data.mode=='register'" ref="joinForm" 
                v-model="data.record" :config="data.config" keep-label
                hide-cancel submit-text="Create" submit-icon="domain"
                buttons-on-bottom :buttons-on-top="false"
                @submit-form="joinTenant"
            >
            </s-form>
            <div v-else-if="data.mode=='info'">
                Registration for {{  data.record.Email }} has been sucessfully done. Please check your inbox for further instruction.
            </div>
        </div>

        <div class="card flex flex-col gap-2" v-if="data.records.length > 0">
            <h1 class="text-primary">Pending Join Request</h1>
            <div class="flex flex-col gap-2">
                <div v-for="record in data.records" class="flex gap-2 even:bg-slate-100 py-1">
                    <div class="flex flex-col gap-1 grow">
                        {{ record.TenantName }}
                    </div>
                    <div>
                        {{ moment(record.Created).fromNow() }}
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>


<script setup>
import { layoutStore } from '@/stores/layout';
import { SForm, SList, createFormConfig, formInput, util } from 'suimjs';
import { inject } from 'vue';
import { onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import moment from 'moment';

layoutStore().change("tenant");

const router = useRouter();
const joinForm = ref(null);
const axios = inject('axios');
const data = reactive({
    mode: 'register',
    record: {},
    config: {},
    records: []
});

onMounted(() => {
    const cfg = createFormConfig("Reset Password",false);
    const tenantID_input = new(formInput);
    tenantID_input.field = "TenantID";
    tenantID_input.label = "Tenant ID or Code";
    tenantID_input.kind = "string";
    tenantID_input.required = true;

    cfg.addSection("General",false).addRowAuto(1, tenantID_input);
    data.config = cfg.generateConfig();

    loadTenantJoinRequest();
});

function joinTenant(_, cb) {
    axios.post("/iam/tenant/request-to-join",data.record).
        then(r => {
            util.showInfo('Request to join tenant has been sent');
            loadTenantJoinRequest();
            cb();
        }, e => {
            cb();
            util.showError(e);
        });
}

function loadTenantJoinRequest() {
    axios.post('/iam/tenantjoin/find?Status=PENDING',{'sort':['-_id']}).then(r => {
        data.records = r.data;
    }, e => util.showError(e))
}

</script>