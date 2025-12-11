<template>
    {{ dimensionText }}
</template>

<script setup>
import { computed } from 'vue';


const props = defineProps({
    dimension: { type: Object, default: () => { } }
})

const dimensionText = computed({
    get () {
        if (props.dimension==null || props.dimension==undefined || props.dimension.length==0) return ""

        return props.dimension.filter(el=>el.Value!="" && el.Value!=null).map(el => {
            return el.Key + "=" + (el.Value=="" ? "*" : el.Value)
        }).join(", ")
    }
})

</script>