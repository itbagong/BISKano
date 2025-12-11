<template>
    <div>
        <data-list ref="lineCtl" title="Rule Line" no-gap hide-title
            grid-hide-search grid-hide-sort grid-mode="grid"
            grid-config="/kara/worklocationuser/gridconfig" 
            :grid-read="`/kara/admin/worklocationuser/gets?WorkLocationID=${workLocationId}`"
            grid-update="/kara/admin/worklocationuser/update"
            grid-delete="/kara/admin/worklocationuser/delete"
            form-config="/kara/worklocationuser/formconfig"
            form-read="/kara/admin/worklocationuser/get"
            form-insert="/kara/admin/worklocationuser/insert"
            form-update="/kara/admin/worklocationuser/update"
            @form-new-data="newRecord"
            >
        </data-list>
    </div>
</template>

<script setup>
import { reactive } from 'vue';
import { DataList } from 'suimjs';

const props = defineProps({
    workLocationId: {type: String, default: ""},
})

const data = reactive({
    appMode: "grid",
    formMode: "edit",
    lineDays: []
})

function newRecord (record) {
    record.WorkLocationID = props.workLocationId;
}


</script>