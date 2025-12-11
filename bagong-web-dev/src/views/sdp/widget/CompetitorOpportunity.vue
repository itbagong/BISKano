<template>
	<div class="flex flex-col gap-2">
		<data-list
			ref="listControl"
			title="Sales Opportunity"
			hide-title
			no-gap
			grid-editor
			grid-hide-select
			grid-hide-search
			grid-hide-sort
			grid-hide-refresh
			grid-no-confirm-delete
			gridAutoCommitLine
			init-app-mode="grid"
			grid-mode="grid"
			form-keep-label
			:grid-fields="['Strength', 'Weakness', 'Threats', 'OpportunityPercentage']"
			:form-fields="['Strength', 'Weakness', 'Threats', 'OpportunityPercentage']"
			new-record-type="grid"
			:grid-config="props.gridConfig"
			:form-config="props.formConfig"
			@grid-row-add="newRecord"
			@form-field-change="onFormFieldChanged"
			@grid-row-delete="onGridRowDelete"
			@grid-row-field-changed="onGridRowFieldChanged"
			@grid-refreshed="gridRefreshed"
			@grid-row-save="onGridRowSave"
			@post-save="onFormPostSave"
			form-focus
		>
			<template #grid_Strength="{item}">
				<s-input
					ref="refStrength"
					label="Strength"
					hide-label
					v-model="item.Strength"
					:multi-row="2"
					class="w-100"
				></s-input>
			</template>
			<template #grid_Weakness="{item}">
				<s-input
					ref="refWeakness"
					label="Weakness"
					hide-label
					v-model="item.Weakness"
					:multi-row="2"
					class="w-100"
				></s-input>
			</template>
			<template #grid_OpportunityPercentage="{item}">
				<s-input
					ref="refOpportunityPercentage"
					label="Opportunity"
					hide-label
					v-model="item.OpportunityPercentage"
					:multi-row="2"
					class="w-100"
				></s-input>
			</template>
			<template #grid_Threats="{item}">
				<s-input
					ref="refThreats"
					label="Threats"
					hide-label
					v-model="item.Threats"
					:multi-row="2"
					class="w-100"
				></s-input>
			</template>

			<template #form_input_Strength="{item}">
				<s-input
					ref="refStrength"
					label="Strength"
					hide-label
					v-model="item.Strength"
					:multi-row="2"
					class="w-100"
				></s-input>
			</template>
			<template #form_input_Weakness="{item}">
				<s-input
					ref="refWeakness"
					label="Weakness"
					hide-label
					v-model="item.Weakness"
					:multi-row="2"
					class="w-100"
				></s-input>
			</template>
			<template #form_input_OpportunityPercentage="{item}">
				<s-input
					ref="refOpportunityPercentage"
					label="Opportunity"
					hide-label
					v-model="item.OpportunityPercentage"
					:multi-row="2"
					class="w-100"
				></s-input>
			</template>
			<template #form_input_Threats="{item}">
				<s-input
					ref="refThreats"
					label="Threats"
					hide-label
					v-model="item.Threats"
					:multi-row="2"
					class="w-100"
				></s-input>
			</template>
		</data-list>
	</div>
</template>

<script setup>
import {DataList, util, SInput} from "suimjs";
import {inject, onMounted, ref, reactive} from "vue";

const axios = inject("axios");

const props = defineProps({
	gridConfig: {type: String, default: () => ""},
	formConfig: {type: String, default: () => ""},
	modelValue: {type: Array, default: () => []},
});

const emit = defineEmits({
	"update:modelValue": null,
	recalc: null,
});

const data = reactive({
	records: props.modelValue.map((dt) => {
		dt.suimRecordChange = false;
		return dt;
	}),
});

const listControl = ref(null);

function updateItems() {
	const committedRecords = data.records.filter(
		(dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
	);
	console.log(committedRecords);
	emit("update:modelValue", committedRecords);
}

const newRecord = () => {
	const record = {
    Competitor: "",
    BiddingAmount: 0.0,
    Strength: "",
    Weakness: "",
    OpportunityPercentage: "",
    Threats: ""
	};
	data.records.push(record);
	listControl.value.setGridRecords(data.records);
	updateItems();
};


function onGridRowSave(record, index) {
	record.suimRecordChange = false;
	data.records[index] = record;
	listControl.value.setGridRecords(data.records);
	updateItems();
}

function onFormPostSave(record, index) {
	record.suimRecordChange = false;
	if (listControl.value.getFormMode() == "new") {
		data.records.push(record);
	} else {
		data.records[index] = record;
	}
	listControl.value.setGridRecords(data.records);
	updateItems();
}

function onGridRowDelete(record, index) {
	const newRecords = data.records.filter((dt, idx) => {
		return idx != index;
	});
	data.records = newRecords;
	listControl.value.setGridRecords(data.records);
	updateItems();
}

function onGridRowFieldChanged(name, v1, v2, old, record) {

	listControl.value.setGridRecord(
		record,
		listControl.value.getGridCurrentIndex()
	);
	updateItems();
}

function gridRefreshed() {
	listControl.value.setGridRecords(data.records);
}

</script>
