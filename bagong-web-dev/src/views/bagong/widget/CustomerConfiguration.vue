<template>
	<div>
		<s-form
			ref="form"
			:config="data.config"
			v-model="data.value"
			keep-label
			@field-change="onFieldChange"
      		:mode="readOnly?'view':'edit'"
		>
			<template #buttons>&nbsp;</template>
		</s-form>
	</div>
</template>

<script setup>
import {reactive, onMounted, inject, ref} from "vue";
import {SForm, loadFormConfig, util} from "suimjs";

const props = defineProps({
	modelValue: {type: Object, default: () => {}},
  readOnly: { type: Boolean, default: false },
});

const form = ref(null);

const emit = defineEmits({
	"update:modelValue": null,
});

const data = reactive({
	config: {},
	value: props.modelValue,
});

const axios = inject("axios");

onMounted(() => {
	loadFormConfig(axios, "/bagong/customerconfiguration/formconfig").then(
		(r) => {
			data.config = r;
		},
		(e) => util.showError(e)
	);
});

function onFieldChange(name, v1, v2, old) {
	if (name == "DividerType") {
		if (v1 == "Auto") {
			form.value.setFieldAttr("Divider", "hide", true);
			data.value.Divider = 0;
		} else {
			form.value.setFieldAttr("Divider", "hide", false);
		}
	}
}
</script>
