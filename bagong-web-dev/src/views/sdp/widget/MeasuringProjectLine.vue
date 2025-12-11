<template>
	<div class="flex flex-col gap-2">
		<!-- <data-list
			ref="listControl"
			title="Measuring Project"
			hide-title
			no-gap
			grid-editor
			grid-hide-select
			grid-hide-search
			grid-hide-sort
			grid-hide-refresh
			grid-hide-new
			grid-no-confirm-delete
			gridAutoCommitLine
			init-app-mode="grid"
			grid-mode="grid"
			form-keep-label
			new-record-type="grid"
      		:grid-config="props.gridConfig"
			:form-config="props.formConfig"
			@form-field-change="onFormFieldChanged"
			@grid-row-field-changed="onGridRowFieldChanged"
			@grid-refreshed="gridRefreshed"
			@grid-row-save="onGridRowSave"
			@post-save="onFormPostSave"
			form-focus
		>
		</data-list> -->
		<div class="flex gap-2 filter-line-card">
			<div class="filter-line-box">
				<s-input
					ref="refBudget"
					v-model="data.filter.Budget"
					use-list
					:lookup-url="`/tenant/masterdata/find?MasterDataTypeID=ACT`"
					lookup-key="_id"
					:lookup-labels="['_id', 'Name']"
					:lookup-searchs="['_id', 'Name']"
					class="w-50"
					@change="
						(field, v1, v2, old, ctlRef) => {
							data.filter.Budget = v1
							onChangeItem(v1, v2, 'budget');
						}
					"
				></s-input>
			</div>
			<div class="filter-line-box">
				<s-input
					ref="refYear"
					v-model="data.filter.Year"
					use-list
					:items="props.listYears"
					class="w-50"
					@change="
						(field, v1, v2, old, ctlRef) => {
							data.filter.Year = v1
							onChangeItem(v1, v2, 'year');
						}
					"
				></s-input>
			</div>
		</div>
		<s-grid
			ref="listControl"
			class="w-full"
			hide-search
			hide-select
			hide-sort
			hide-new-button
			hide-delete-button
			hide-refresh-button
			hide-detail
			hide-action
			auto-commit-line
			no-confirm-delete
			:config="props.gridCfg"
			form-keep-label
			v-model="dtLines"
		>
			<template #item_LedgerAccount="{ item }">
				<s-input hide-label v-model="item.LedgerAccount" class="w-full" 
					read-only
					ref="refLedgerAccount"
                        :lookup-url="`/tenant/ledgeraccount/find?_id=${item.LedgerAccount}`" 
                        lookup-key="_id"
                        :lookup-labels="['_id', 'Name']"
                        :lookup-searchs="['_id','Name']"
                    ></s-input>
				<!-- <s-input
					ref="refLedgerAccount"
					v-model="item.LedgerAccount"
					class="w-full"
					read-only
				></s-input> -->
			</template>
			<template #item_January="{ item }">
				<div v-if="item.Month.January !== undefined">
					<s-input
						ref="refJanuary"
						v-model="item.Month.January"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_February="{ item }">
				<div v-if="item.Month.February !== undefined">
					<s-input
						ref="refFebruary"
						v-model="item.Month.February"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_March="{ item }">
				<div v-if="item.Month.March !== undefined">
					<s-input
						ref="refMarch"
						v-model="item.Month.March"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_April="{ item }">
				<div v-if="item.Month.April !== undefined">
					<s-input
						ref="refApril"
						v-model="item.Month.April"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_May="{ item }">
				<div v-if="item.Month.May !== undefined">
					<s-input
						ref="refMay"
						v-model="item.Month.May"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_June="{ item }">
				<div v-if="item.Month.June !== undefined">
					<s-input
						ref="refJune"
						v-model="item.Month.June"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_July="{ item }">
				<div v-if="item.Month.July !== undefined">
					<s-input
						ref="refJuly"
						v-model="item.Month.July"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_August="{ item }">
				<div v-if="item.Month.August !== undefined">
					<s-input
						ref="refAugust"
						v-model="item.Month.August"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_September="{ item }">
				<div v-if="item.Month.September !== undefined">
					<s-input
						ref="refSeptember"
						v-model="item.Month.September"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_October="{ item }">
				<div v-if="item.Month.October !== undefined">
					<s-input
						ref="refOctober"
						v-model="item.Month.October"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_November="{ item }">
				<div v-if="item.Month.November !== undefined">
					<s-input
						ref="refNovember"
						v-model="item.Month.November"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
			<template #item_December="{ item }">
				<div v-if="item.Month.December !== undefined">
					<s-input
						ref="refDecember"
						v-model="item.Month.December"
						class="w-full"
						kind="number"
					></s-input>
				</div>
				<span v-else>-</span>
			</template>
		</s-grid>
	</div>
</template>

<style scoped>
.filter-line-card {
	justify-content:end;
}
.filter-line-box {
	width: 25% !important;
}
</style>

<script setup>
import {DataList, util, loadGridConfig, SGrid, SInput} from "suimjs";
import {inject, onMounted, ref, reactive, computed} from "vue";
import moment from 'moment'

const axios = inject("axios");

const props = defineProps({
	gridConfig: {type: String, default: () => ""},
	formConfig: {type: String, default: () => ""},
	modelValue: {type: Array, default: () => []},
  	gridCfg: {type: Object, default: () => {}},
	listYears: {type: Array, default: () => []},
});

const emit = defineEmits({
	"update:modelValue": null,
	recalc: null,
});

const data = reactive({
	records: props.modelValue,
	// .map((dt) => {
	// 	dt.suimRecordChange = false;
	// 	return dt;
	// }),
	month: [ "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December" ],
	year: [],
	filter: {
		Budget: "",
		Year: "",
	},
});

const dtLines = computed({
  get() {
	let dataParentLines = props.modelValue;
	// console.log("dataParent",dataParentLines)
	let dtLines = []
	if (data.filter.Budget == "" || data.filter.Year == "") {
		return dtLines = []
	}

	dataParentLines.forEach(e => {
		if (e.Budget == data.filter.Budget && e.Year == data.filter.Year) {
			dtLines.push(e)
		}
	})
	// console.log(dtLines)
	return dtLines;
  },
  set(v) {
    emit("update:modelValue", v);
  },
});

const listControl = ref(null);

function onChangeItem(v1, v2, item) {
	//   console.log(v1, v2, item)
	if (item == "budget") {
		
	}else if (item == "year") {
		
	}
}
// function updateItems() {
// 	const committedRecords = data.records.filter(
// 		(dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
// 	);
// 	console.log(committedRecords);
// 	emit("update:modelValue", committedRecords);
// }

// function onGridRowSave(record, index) {
// 	record.suimRecordChange = false;
// 	data.records[index] = record;
// 	listControl.value.setGridRecords(data.records);
// 	updateItems();
// }

// function onFormPostSave(record, index) {
// 	record.suimRecordChange = false;
// 	if (listControl.value.getFormMode() == "new") {
// 		data.records.push(record);
// 	} else {
// 		data.records[index] = record;
// 	}
// 	listControl.value.setGridRecords(data.records);
// 	updateItems();
// }

// function onGridRowFieldChanged(name, v1, v2, old, record) {

// 	listControl.value.setGridRecord(
// 		record,
// 		listControl.value.getGridCurrentIndex()
// 	);
// 	updateItems();
// }

// function gridRefreshed() {
// 	listControl.value.setGridRecords(data.records);
// }
</script>
