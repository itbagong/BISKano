<template>
    <div class="flex flex-col gap-1">
        <div class="flex gap-1 items-center" v-for="v in data.dimValues">
            <div class="text-xs w-[100px]">{{ v.Kind }}</div>
            <s-input hide-label v-model="v.Value" :label="v.Kind" class="w-full" :disabled="readOnly"></s-input>
        </div>
    </div>
</template>

<script setup>
import { reactive, watch } from 'vue';
import { SInput } from 'suimjs';

const props = defineProps({
    modelValue: { type: Array, default: () => [] },
    dimNames: { type: Array, default: () => [] },
    column: { type: Number, default: 2 },
    readOnly: { type: Boolean, defaule: false }
})

const emit = defineEmits({
    "update:modelValue": null,
})

const data = reactive({
    //dimTypes: props.dimNames,
    dimValues: props.modelValue
})

function buildDimValues(dimNames) {
    const mv = props.modelValue
    const vs = dimNames.map(el => {
        const f = mv.filter(v => v.Kind == el)
        const v = f.length == 0 ? "" : f[0].Value
        return {
            Kind: el,
            Value: v
        }
    })
    return vs
}

watch(() => props.dimNames, (nv) => {
    data.dimValues = buildDimValues(nv)
    //data.dimTypes = nv
})

watch(() => data.dimValues, (nv) => {
    emit("update:modelValue", nv)    
}, { deep: true })

</script>