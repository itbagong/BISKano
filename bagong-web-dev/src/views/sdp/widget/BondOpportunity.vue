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
		</data-list>
	</div>
</template>

<script setup>
import {DataList, util} from "suimjs";
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
		TypeBond: "",
		Amount: 0.0,
		SubmitDate: new Date(),
		StatusBond: "",
		ExpiredDate: new Date(),
		Guarantor: ""
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
