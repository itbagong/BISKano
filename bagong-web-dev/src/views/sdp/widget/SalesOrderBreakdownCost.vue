<template>
	<div class="flex flex-col gap-2">
		<data-list
			ref="listControl"
			title="Sales Order"
			hide-title
			no-gap
			grid-editor
			grid-hide-select
			grid-hide-search
			grid-hide-sort
			grid-hide-refresh
			grid-no-confirm-delete
			grid-auto-commit-line
			init-app-mode="grid"
			grid-mode="grid"
			form-keep-label
			new-record-type="grid"
			:grid-config="props.gridConfig"
			:grid-fields="[
				'BreakdownCostItem',
				'Amount',
			]"
			:form-config="props.formConfig"
			@grid-row-add="newRecord"
			@form-field-change="onFormFieldChanged"
			@grid-row-delete="onGridRowDelete"
			@grid-row-field-changed="onGridRowFieldChanged"
			@grid-refreshed="gridRefreshed"
			@grid-row-save="onGridRowSave"
			@post-save="onFormPostSave"
			@alter-grid-config="AlterGridConfig"
			form-focus
		>
			<template #grid_BreakdownCostItem="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'BreakdownCostItem'
					)"
					:key="`gridbreakdowncostiteminput ${hdrindex}`"
				>
					<s-input
						ref="refBreakdownCostItem"
						hide-label
						use-list
						:lookup-url="`/tenant/masterdata/find?MasterDataTypeID=SDBC`"
						lookup-key="_id"
						:lookup-labels="['_id', 'Name']"
						:lookup-searchs="['_id', 'Name']"
						@change="rowFieldChanged"
						v-model="item[hdr.input.field]"
					></s-input>
				</div>
			</template>

			<template #grid_Amount="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'Amount'
					)"
					:key="`gridamountinput ${hdrindex}`"
				>
					<s-input
						ref="refAmount"
						hide-label
						:ctl-ref="{rowIndex: item.index}"
						:field="hdr.input.field"
						:kind="hdr.input.kind"
						@change="rowFieldChanged"
						v-model="item[hdr.input.field]"
					></s-input>
				</div>
			</template>
		</data-list>
	</div>
</template>

<script setup>
import {DataList, util, SInput} from "suimjs";
import {inject, ref, reactive, watch} from "vue";

const axios = inject("axios");

const props = defineProps({
	gridConfig: {type: String, default: () => ""},
	formConfig: {type: String, default: () => ""},
	modelValue: {type: Array, default: () => []},
});

const emit = defineEmits({
	"update:modelValue": null,
	// recalc: null,
});

const data = reactive({
	records: (props.modelValue ?? []).map((dt, index) => {
		dt.suimRecordChange = false;
		dt.index = index;
		return dt;
	}),

	changed: false,
	defaultgridconfig: {},
	currentIndex: -1,
	recordChanged: false,
});

watch(
	() => props.modelValue,
	(nt) => {
		// console.log("watch", nt);
		if (data.changed == false) {
			const records = (nt ?? []).map((dt, index) => {
				dt.suimRecordChange = false;
				dt.index = index;
				return dt;
			});
			data.records = records;
			listControl.value.setGridRecords(records);
		}
		data.changed = false;
	}
);

const listControl = ref(null);

const newRecord = () => {
	const records = listControl.value.getGridRecords();
	records.push({
		BreakdownCostItem: "",
		Amount: 0,
		index: records.length,
	});

	listControl.value.setGridRecords(records);
	updateItems();
	// console.log("newRecord", records);
};

async function onFormFieldChanged(name, v1, v2, old, record) {
	updateItems();
	// console.log("onFormFieldChanged", record);
}

function onGridRowSave(record, index) {
	record.suimRecordChange = false;

	const records = listControl.value.getGridRecords();
	records[index] = record;
	listControl.value.setGridRecords(records);
	updateItems();
	// console.log("onGridRowSave", records);
}

async function onFormPostSave(record, index) {
	record.suimRecordChange = false;

	const records = listControl.value.getGridRecords();
	if (listControl.value.getFormMode() == "new") {
		records.push(record);
	} else {
		records[index] = record;
	}
	listControl.value.setGridRecords(records);
	updateItems();
	// console.log("onFormPostSave", records);
}

function onGridRowDelete(record, index) {
	const records = listControl.value.getGridRecords();
	const newRecords = records
		.filter((dt, idx) => {
			return idx != index;
		})
		.map((nr, index) => ({...nr, index: index}));

	listControl.value.setGridRecords(newRecords);
	updateItems();
	// console.log("onGridRowDelete", newRecords);
}

function onGridRowFieldChanged(name, v1, v2, old, record) {

	listControl.value.setGridRecord(
		record,
		listControl.value.getGridCurrentIndex()
	);

	updateItems();
	// console.log("onGridRowFieldChanged", record);
}

function gridRefreshed() {
	listControl.value.setGridRecords(data.records);
	// console.log("gridRefreshed", data.records);
}

function updateItems() {
	const records = listControl.value.getGridRecords();
	data.records = [...records];
	const committedRecords = records.filter(
		(dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
	);

	// emit("recalc", ReCalc(committedRecords));
	emit("update:modelValue", committedRecords);
	data.changed = true;
	// console.log("updateItems", committedRecords);
}

function rowFieldChanged(name, v1, v2) {
	const currentIndex = data.currentIndex;
	const current = data.records[currentIndex];

	onGridRowFieldChanged(name, v1, v2, current, current);
	emit("rowFieldChanged", name, v1, v2, current, current);
	emit("update:modelValue", data.records);
	// console.log("rowFieldChanged",data, data.records);
}


const AlterGridConfig = (gridconfigs) => {
	const records = data.records;

	const assetfield = gridconfigs.fields;

	data.defaultgridconfig = assetfield;
};
</script>
