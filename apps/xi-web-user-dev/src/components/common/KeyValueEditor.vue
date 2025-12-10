<template>
    <div>
        <data-list ref="list" no-gap hide-title
            v-model="data.dims" 
            grid-mode="grid" init-app-mode="grid"
            grid-hide-search grid-hide-refresh grid-editor new-record-type="grid" 
            grid-hide-sort grid-hide-select grid-hide-detail
            grid-no-confirm-delete grid-auto-commit-line
            grid-config="/admin/kv/gridconfig"
            @grid-row-add="addDim" @grid-row-delete="deleteDim" 
            @grid-refreshed="refresh"        
        >
        </data-list>
    </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue';
import { DataList, util } from 'suimjs';

const props = defineProps({
    modelValue: { type: Array, default: () => [] },
    mandatoryKeys: { type: Array, default: () => [] },
})

const list = ref(null);

const emit = defineEmits({
    "update:modelValue": null,
})

const data = reactive({
   dims: props.modelValue
})

function refresh() {
	if (list.value) list.value.setGridRecords(data.dims);
    emit("update:modelValue", data.dims);
}

function addDim() {
    const record = {
        Kind: "",
        Value: "",
        suimRecordChange: true,
    };
    if (data.dims==null) data.dims = [];
    data.dims.push(record);
    refresh()
}

function deleteDim(rcd, index) {
	const newRecords = data.dims.filter((dt, idx) => {
		return idx != index;
	});
	data.dims = newRecords;
	refresh();
}

</script>