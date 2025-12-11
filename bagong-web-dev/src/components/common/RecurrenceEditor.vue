<template>
    <div class="flex gap-3">
        <s-input kind="date" label="Start Date" v-model="data.recurrenceData.StartDate" hideLabel></s-input>

        <div v-if="data.useFrequency" class="flex gap-2">
            <s-button class="btn_secondary" icon="minus-circle-outline" @click="minCount"></s-button>
            <s-input class="w-[60px]" kind="text" label="Count" v-model="data.recurrenceData.RecurenceCount"
                hideLabel></s-input>
            <s-button class="btn_secondary" icon="plus-circle-outline" @click="plusCount"></s-button>
        </div>

        <s-input useList :items="['DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY']" class="w-[140px]"
            v-model="data.recurrenceData.Frequency" hideLabel></s-input>
        <s-input kind="date" label="End Date" v-model="data.recurrenceData.EndDate" hideLabel></s-input>
    </div>
</template>
<script setup>
import { reactive, watch, onMounted } from 'vue';
import { SInput, SButton } from 'suimjs';

const props = defineProps({
    modelValue: { type: Object, Default: () => [] },
})

const emit = defineEmits({
    "update:modelValue": null,
})

const data = reactive({
    recurrenceData: {
        StartDate: "",
        EndDate: "",
        RecurenceCount: 0,
        RecurenceAmount: 0,
        Frequency: "DAY"
    },
    useFrequency: true
    // recurrenceData: props.modelValue,
})

function minCount() {
    let count = (data.recurrenceData.RecurenceCount - 1 < 0) ? 0 : (data.recurrenceData.RecurenceCount - 1)
    data.recurrenceData.RecurenceCount = count
    // data.recurrenceData.RecurenceCount = data.recurrenceData.RecurenceCount - 1
}

function plusCount() {
    let count = data.recurrenceData.RecurenceCount ? data.recurrenceData.RecurenceCount + 1 : 0 + 1
    data.recurrenceData.RecurenceCount = count
}

onMounted(() => {
    data.recurrenceData = props.modelValue ?? { RecurenceCount: 0, Frequency: "DAY" }
})

watch(() => props.modelValue, (nv) => {
    data.recurrenceData = nv
})

watch(() => data.recurrenceData, (nv) => {
    emit("update:modelValue", nv);
},
    { deep: true }
);
</script>
<style></style>