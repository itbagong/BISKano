<template>
    <div>
        <label class="input_label">{{ title }}</label>
        <data-list 
            ref="listControl" 
            title="Posting Profile" 
            hide-title 
            no-gap 
            grid-editor 
            grid-hide-search 
            grid-hide-sort 
            grid-hide-refresh 
            grid-hide-detail 
            grid-hide-select 
            grid-no-confirm-delete
            init-app-mode="grid" 
            grid-mode="grid" 
            grid-config="/sdp/journaltypecontext/gridconfig" 
            new-record-type="grid" 
            grid-auto-commit-line
            @grid-row-add="newRecord" 
            @grid-row-delete="onGridRowDelete" 
            @grid-row-save="onGridRowSave"
        >
        </data-list>
    </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue';
import { DataList } from 'suimjs';

const props = defineProps({
    title: {type: String, default: ''},
    modelValue: {type: Array, default: () => []},
});

const emit = defineEmits({
    'update:modelValue': null
});

const data = reactive({
    records: props.modelValue==null || props.modelValue==undefined ? [] : props.modelValue
});

const listControl = ref(null);

function newRecord(r) {
    r.ID = '';
    r.Label = '';
    r.Addr = '';
    //r.suimRecordChange = true;

    data.records.push(r);
    updateItems();
}

function onGridRowDelete(_, index) {
    const newRecords = data.records.filter((dt, idx) => {
        return idx!=index;
    })
    data.records = newRecords;
    listControl.value.setGridRecords(data.records);
    updateItems();
}

function onGridRowSave (record, index) {
    record.suimRecordChange = false;
    data.records[index] = record;
    listControl.value.setGridRecords(data.records);
    updateItems();
}

onMounted(() => {
    setTimeout(() => {
        updateItems();
    }, 500);
})

function updateItems () {
    listControl.value.setGridRecords(data.records);
}

</script>