<template>
    <div class="flex flex-col gap-1">
        <div class="flex gap-1 items-center" v-for="(input, inputIdx) in inputCfg">
            <div class="w-[100px] text-xs">{{ input.Name }}</div>
            <s-input class="grow" hide-label :disabled="readOnly"
                :field="input.Name" :label="input.Name" :caption="input.Name" kind="text"
                v-model="data.location[inputIdx].Value" use-list :lookup-url="input.lookupUrl" lookup-key="_id"
                :lookup-labels="['Name']" @change="handleChange" ref="dimControl"></s-input>
        </div>
        <!--
            <div>{{ data.location }}</div>
            <div>{{ props.modelValue }}</div>
        -->
    </div>
</template>

<script setup>
import { SInput } from 'suimjs';
import { authStore } from '@/stores/auth';
import { reactive, ref, watch } from 'vue';

const auth = authStore()
const inventDimensions = auth.appData.InventDimensions
const dimControl = ref([])

const props = defineProps({
    modelValue: { type: Array, default: () => [] },
    readOnly: { type: Boolean, default: false }
})

const inputCfg = auth.appData.InventDimensions.map(d => {
    //const parentDimCfg = auth.appData.InventDimensions.filter(p => p.Name == d.Parent)
    if (d.Parent == '') d.lookupUrl = `/inventory/location/find?Kind=${d.Name}`
    else {
        if (props.modelValue==undefined) return d

        const parentDimData = props.modelValue.filter(p => p.Kind == d.Parent)
        if (parentDimData.length == 0)
            d.lookupUrl = `/inventory/location/find?Kind=${d.Name}&ParentID=DataThatWillNeverExists`
        else
            d.lookupUrl = `/inventory/location/find?Kind=${d.Name}&ParentID=${parentDimData[0].Value}`
    }
    return d
})

const emit = defineEmits({
    "update:modelValue": null,
})

const data = reactive({
    location: inventDimensions.map(d => {
        if (props.modelValue==undefined) return {Kind: d.Name, Value: ''}
        const pdim = props.modelValue.filter(el => el.Kind == d.Name)
        return { Kind: d.Name, Value: pdim.length == 0 ? d.DefaultValue : pdim[0].Value }
    })
})

/*
function buildLocationValues (nv) {
    const newLoc = nv.map(d => {
        return {Kind: d.Name, Value: d.Value}
    })
    return newLoc
}
*/

function handleChange(name, v1) {
    inputCfg.filter(d => d.Parent == name).forEach(d => {
        data.location = data.location.map(l => {
            if (l.Kind==d.Name) {
                l.Value=""
            }
            return l
        })
        if (v1 == undefined || v1 == "")
            d.lookupUrl = `/inventory/location/find?Kind=${d.Name}&ParentID=DataThatWillNeverExists`
        else
            d.lookupUrl = `/inventory/location/find?Kind=${d.Name}&ParentID=${v1}`
    })
}

watch(() => data.location, (nv) => {
    //const newLocation = buildLocationValues(nv)
    emit("update:modelValue", nv)
}, { deep: true })
</script>