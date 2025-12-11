<template>
    <div class="flex flex-col gap-2">
        Record: {{data.records}}
    </div>
</template>

<script setup>
    import { onMounted, reactive, ref, inject } from "vue";
import { SCard,
	SForm,
	util,
	SInput,
	SModal,
	SButton, 
loadFormConfig } from "suimjs";
import {layoutStore} from "../../../stores/layout.js";
import {useRoute} from "vue-router";

layoutStore().name = "tenant";

const axios = inject("axios");

const props = defineProps({
    modelValue: { type: Object, default: () => {} },
});

const data = reactive({
  records: props.modelValue,
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});



</script>