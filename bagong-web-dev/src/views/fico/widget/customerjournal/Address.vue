<template>
    <div>
        <s-form :config="data.config" v-model="data.value" keep-label  :mode="readOnly?'view':'edit'">
            <template #buttons>&nbsp;</template>
        </s-form>
    </div>
</template>

<script setup>
import { reactive, onMounted, inject } from 'vue';
import { SForm, loadFormConfig, util } from 'suimjs';

const props = defineProps({
    'modelValue': { type: Object, default: () => { } },
  readOnly: { type: Boolean, default: false },
})

const emit = defineEmits({
    'update:modelValue': null
})

const data = reactive({
    config: {},
    value: props.modelValue
})

const axios = inject('axios');

onMounted(() => {
    loadFormConfig(axios, '/fico/customerjournal/address/formconfig').then(r => {
        data.config = r
    }, e => util.showError(e));
})

</script>