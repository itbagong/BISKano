<template>
    <div class="w-full">
        <div class="card flex flex-col gap-2">
            <h1>Create New Tenant</h1>
            <s-form v-if="data.mode == 'register'" ref="regForm" v-model="data.record" :config="data.config" keep-label
                hide-cancel submit-text="Create" submit-icon="domain" buttons-on-bottom :buttons-on-top="false"
                @submit-form="createTenant">
            </s-form>
            <div v-else-if="data.mode == 'info'">
                Registration for {{ data.record.Email }} has been sucessfully done. Please check your inbox for further
                instruction.
            </div>
        </div>
    </div>
</template>


<script setup>
import { layoutStore } from '@/stores/layout';
import { SForm, loadFormConfig, util } from 'suimjs';
import { inject } from 'vue';
import { onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

layoutStore().change("tenant");

const router = useRouter();
const regForm = ref(null);
const axios = inject('axios');
const data = reactive({
    mode: 'register',
    record: {},
    config: {},
});

onMounted(() => {
    loadFormConfig(axios, "/iam/ui/create-tenant/formconfig").
        then(r => {
            data.config = r
            util.nextTickN(2, () => {
                regForm.value.setFieldAttr('FID', 'rules', [
                    (fid) => {
                        const regex = /^[a-zA-Z0-9(_\-)]*$/;
                        return regex.test(fid) ? '' : 'only accepts A-Z, a-z, 0-9 and character (- and _)';
                    }
                ])
            });
        }, e => util.showError(e))
});

function createTenant(_, cb) {
    axios.post("/iam/tenant/create", data.record).
        then(r => {
            router.push('/tenant/' + data.record.FID);
        }, e => {
            cb();
            util.showError(e);
        });
}

</script>