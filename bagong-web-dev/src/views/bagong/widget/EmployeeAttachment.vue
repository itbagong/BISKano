<template>
    <div class="w-full">
        <GridAttachment :siteEntryAssetID="props.EmployeeID" gridConfig="/bagong/employee_attachment/gridconfig"
            :gridFields="['FileName', 'UploadDate', 'URI']" v-model="data.records"></GridAttachment>
    </div>
</template>
<script setup>
import { reactive, ref, onMounted, inject, watch } from 'vue';
import { DataList, SButton, loadFormConfig, util } from 'suimjs';
import GridAttachment from "@/components/common/GridAttachment.vue";

const props = defineProps({
    EmployeeID: { type: String, default: "" },
    modelValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
    "update:modelValue": null,
});

const data = reactive({
    records:
        props.modelValue == null || props.modelValue == undefined
            ? []
            : props.modelValue,
    fileRecords: [],
});

onMounted(() => {
    // console.log("attachment:", props.modelValue)
    // props.modelValue = [{ Description: "KTP" }, { Description: "KK" }]
})

watch(
    () => data.records,
    (nv) => {
        emit("update:modelValue", nv);
    },
    { deep: true }
);
</script>
<style></style>