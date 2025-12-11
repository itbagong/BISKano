<template>
    <div class="card">
        <s-form ref="formControl" title="Attendance Entry"
            v-if="data.formCfg!=null" 
            v-model="data.record" 
            :config="data.formCfg"
            keep-label submit-text="Submit"
            hide-cancel
            @submit-form="submit"
            @cancel-form="cancel"
            >
        </s-form>
    </div>
</template>

<script setup>
import { reactive, ref, onMounted, inject } from 'vue';
import { SForm, loadFormConfig, util } from 'suimjs';
import { layoutStore } from "@/stores/layout.js";

layoutStore().name = "tenant";

const axios = inject("axios");
const formControl = ref(null);

const data = reactive({
    record: {
        Op: "Checkin",
    },
    formCfg: null
});

function submit(_, fnOK, fnCancel) {
    axios.post("/kara/admin/trx/create", data.record).then(r => {
        util.showInfo("attendance has been submitted");
        fnOK();
    }, e => {
        util.showError(e);
        fnCancel();
    })
}

function cancel() {
}

onMounted(() => {
    loadFormConfig(axios, "/kara/trx/request/formconfig").then(r => {
        data.formCfg = r;
    }, e => util.showError(e))
});

</script>