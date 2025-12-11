<template>
	<div class="flex flex-col gap-2">
		<data-list
			ref="listControl"
			title="Customer Contact"
			hide-title
			no-gap
			:grid-hide-detail="readOnly"
			:grid-editor="!readOnly"
			:grid-hide-delete="readOnly"
			:grid-hide-control="readOnly" 
			grid-hide-select
			grid-hide-search
			grid-hide-sort
			grid-hide-refresh
			grid-no-confirm-delete
			
			gridAutoCommitLine
			init-app-mode="grid"
			grid-mode="grid"
			form-update="/tenant/contact/save"
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
  readOnly: { type: Boolean, default: false },
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
		Name: "",
		Company: "",
		Site: "",
		Role: "",
		PhoneNumber: "",
		Email: "",
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

async function onGridRowDelete(record, index) {
	// grid-delete="/bagong/contact/delete"
	try {
		const deletedData = data.records[index];
		const deleteRes = await axios.post(`/tenant/contact/delete`, deletedData);
		// console.log("deleteRes", deleteRes);
		// console.log("deletedData", deletedData);
		const newRecords = data.records.filter((dt, idx) => {
			return idx != index;
		});
		data.records = newRecords;
		listControl.value.setGridRecords(data.records);
		updateItems();
	} catch (error) {
		util.showError(error);
	}
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
