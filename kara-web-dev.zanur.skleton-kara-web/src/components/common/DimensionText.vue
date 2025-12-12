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
        if (props.dimension.Items==null || props.dimension.Items==undefined || props.dimension.Items.length==0) return "All Data"

        return props.dimension.Items.filter(el=>el.Value!="").map(el => {
            return el.Kind + "=" + (el.Value=="" ? "*" : el.Value)
        }).join(", ")
    }
})

</script>