<template>
	<s-card
		class="w-full bg-white suim_datalist"
		hide-footer
		:no-gap="true"
		:hide-title="true"
	>
		<s-form
			v-if="data.formCfg && data.formCfg.setting"
			ref="formCtl"
			v-model="value"
			:config="data.formCfg"
			:mode="data.formMode"
			class="pt-2"
			@cancelForm="cancelForm"
			hide-buttons
			hide-submit
			hide-cancel
		>
			<template #input_LetterHeadAsset="{config, item}">
				<span v-if="item[config.field]"
					><a
						@click="goto('/asset/view?id=' + item[config.field])"
						href="#"
						class="text-blue-600 hover:text-blue-800 visited:text-purple-600"
						>View File</a
					></span
				>
				<input
					:type="config.kind"
					:placeholder="config.caption || config.label"
					class="input_field"
					ref="control"
					@change="(file) => FileReadertoBase64(file, config)"
					:disabled="config.disabled"
				/>
			</template>
			<template #input_FooterAsset="{config, item}">
				<span v-if="item[config.field]"
					><a @click="goto('/asset/view?id=' + item[config.field])" href="#"
						>View File</a
					></span
				>
				<input
					:type="config.kind"
					:placeholder="config.caption || config.label"
					class="input_field"
					ref="control"
					@change="(file) => FileReadertoBase64(file, config)"
					:disabled="config.disabled"
				/>
			</template>
		</s-form>
	</s-card>
</template>

<script setup>
import {
	SCard,
	SForm,
	util,
	SInput,
	SModal,
	SButton,
	loadFormConfig,
} from "suimjs";
import {reactive, defineProps, ref, onMounted, inject, computed} from "vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";

const axios = inject("axios");

const props = defineProps({
	modelValue: {type: Object, default: () => {}},
	formConfig: {type: [String, Object], default: () => {}},
	formDefaultMode: {type: String, default: "edit"},
});

const data = reactive({
	formCfg: undefined,
	record: props.modelValue,
	formMode: props.formDefaultMode,
});

const emit = defineEmits({
	"update:modelValue": undefined,
	alterFormConfig: null,
});

function goto(destination) {
	window.location.href = import.meta.env.VITE_API_URL + destination;
}

const value = computed({
	get() {
		switch (props.kind) {
			default:
				return props.modelValue;
		}
	},

	set(v) {
		// handleChange(v, value2(v), props.modelValue, props.ctlRef);
		emit("update:modelValue", v);

		// nextTick(() => {
		// 	if (!props.disableValidateOnChange) validate();
		// });
	},
});

const formCtl = ref(null);

function refreshForm() {
	if (props.formConfig == undefined || props.formConfig == "") return;
	loadFormConfig(axios, props.formConfig).then(
		(r) => {
			emit("alterFormConfig", r);
			data.formCfg = r;
		},

		(e) => util.showError(e)
	);
}

function FileReadertoBase64(eventfile, config) {
	const value = props.modelValue;
	const file = eventfile.target.files[0];

	const reader = new FileReader();
	reader.readAsDataURL(file);

	reader.onloadend = () => {
		value["Upload" + config.field] = {
			Content: reader.result,
			Asset: {
				OriginalFileName: file.name,
				ContentType: file.type,
			},
		};

		emit("update:modelValue", value);
	};
}

onMounted(() => {
	refreshForm();
});
</script>
