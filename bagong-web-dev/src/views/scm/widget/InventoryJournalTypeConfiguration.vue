<template>
    <div>
        <s-form :config="data.config" v-model="data.value" keep-label>
            <template #buttons>&nbsp;</template>
        </s-form>
    </div>
</template>

<script setup>
import { reactive, onMounted, inject } from 'vue';
import { SForm, loadFormConfig, util } from 'suimjs';

const props = defineProps({
    'id': { type: String, default: '' }
})

const emit = defineEmits({
    'update:modelValue': null
})

const data = reactive({
    config: {},
    value: {
		VendorJournalTypeID: props.id
	}
})

const axios = inject('axios');

onMounted(() => {
    loadFormConfig(axios, '/bagong/inventorytransactionjournaltypeconfiguration/formconfig').then(r => {
        data.config = r
    }, e => util.showError(e));

    axios.post('/bagong/inventorytransactionjournaltypeconfiguration/find?VendorJournalTypeID='+props.id).then(
		(r) => {
			if (r.data && r.data.length > 0) {
				data.value = r.data[0];
			}
        },
            (e) => {
            data.loading = false;
        }
  	);
})

function getDataValue() {
	return data.value
}

defineExpose({
    getDataValue
})

</script>