<template>
	<div class="flex flex-col gap-2">
		<s-modal
			title="Ratio Factor"
			class=""
			:display="false"
			ref="ratioFactorModal"
			hideButtons
			@submit="closeModalRatioFactor"
		>
			<!-- slot input -->
			<div class="w-100 content-input">
				<s-input
					ref="totalDaysAYear"
					class="mb-4 min-w-[300px]"
					label="Total Days a Year ( TD )"
					v-model="data.ratioFactor.totalDaysAYear"
					placeholder="Total Days a Year ( TD )"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('totalDaysAYear', v1, v2)"
				></s-input>
				<s-input
					ref="totalDaysOffAYear"
					class="mb-4 min-w-[300px]"
					label="Total National Days Off a Year"
					v-model="data.ratioFactor.totalDaysOffAYear"
					placeholder="Total National Days Off a Year"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('totalDaysOffAYear', v1, v2)"
				></s-input>
				<s-input
					ref="totalActiveDaysAYear"
					class="mb-4 min-w-[300px]"
					label="Total Active Days a Year ( TAD )"
					:disabled="true"
					v-model="data.ratioFactor.totalActiveDaysAYear"
					placeholder="Total Active Days a Year ( TAD )"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('totalActiveDaysAYear', v1, v2)"
				></s-input>
			</div>
			<div class="w-100 content-input">
				<s-input
					ref="AnnualLeave"
					class="mb-4 min-w-[300px]"
					label="Annual Leave ( Cuti Tahunan ) + Day Cuti tanggal merah"
					v-model="data.ratioFactor.AnnualLeave"
					placeholder="Annual Leave ( Cuti Tahunan ) + Day Cuti tanggal merah"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('AnnualLeave', v1, v2)"
				></s-input>
				<s-input
					ref="workingDays"
					class="mb-4 min-w-[300px]"
					label="Hari Kerja Tersedia 1 Tahun"
					:disabled="true"
					v-model="data.ratioFactor.workingDays"
					placeholder="Hari Kerja Tersedia 1 Tahun"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('workingDays', v1, v2)"
				></s-input>
				<s-input
					ref="durationOfOneDutyPeriod"
					class="mb-4 min-w-[300px]"
					label="Durasi Satu Periode Tugas"
					v-model="data.ratioFactor.durationOfOneDutyPeriod"
					placeholder="Durasi Satu Periode Tugas"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('durationOfOneDutyPeriod', v1, v2)"
				></s-input>
			</div>
			<div class="w-100 content-input">
				<s-input
					ref="siteLeave"
					class="mb-4 min-w-[300px]"
					label="Site Leave + Travelling"
					v-model="data.ratioFactor.siteLeave"
					placeholder="Site Leave + Travelling"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('siteLeave', v1, v2)"
				></s-input>
				<s-input
					ref="oneDutyPeriodAndOneLeavPeriod"
					class="mb-4 min-w-[300px]"
					label="Satu Periode Tugas & Satu Periode Cuti"
					:disabled="true"
					v-model="data.ratioFactor.oneDutyPeriodAndOneLeavPeriod"
					placeholder="Satu Periode Tugas & Satu Periode Cuti"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('oneDutyPeriodAndOneLeavPeriod', v1, v2)"
				></s-input>
				<s-input
					ref="numberOfDutyAndLeavePeriodsInAYear"
					class="mb-4 min-w-[300px]"
					label="Jumlah Periode Tugas & Cuti dalam 1 Tahun"
					:disabled="true"
					v-model="data.ratioFactor.numberOfDutyAndLeavePeriodsInAYear"
					placeholder="Jumlah Periode Tugas & Cuti dalam 1 Tahun"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('numberOfDutyAndLeavePeriodsInAYear', v1, v2)"
				></s-input>
			</div>
			<div class="w-100 content-input">
				<s-input
					ref="numberOfDutyDays"
					class="mb-4 min-w-[300px]"
					label="Jumlah Hari Tugas"
					:disabled="true"
					v-model="data.ratioFactor.numberOfDutyDays"
					placeholder="Jumlah Hari Tugas"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('numberOfDutyDays', v1, v2)"
				></s-input>
				<s-input
					ref="numberOfLeaveDays"
					class="mb-4 min-w-[300px]"
					label="Jumlah Hari Cuti"
					:disabled="true"
					v-model="data.ratioFactor.numberOfLeaveDays"
					placeholder="Jumlah Hari Cuti"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('numberOfLeaveDays', v1, v2)"
				></s-input>
				<s-input
					ref="dayWork"
					class="mb-4 min-w-[300px]"
					label="Day Work"
					v-model="data.ratioFactor.dayWork"
					placeholder="Day Work"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('dayWork', v1, v2)"
				></s-input>
			</div>
			<div class="w-100 content-input">
				<s-input
					ref="dayOff"
					class="mb-4 min-w-[300px]"
					label="Day Off"
					v-model="data.ratioFactor.dayOff"
					placeholder="Day Off"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('dayOff', v1, v2)"
				></s-input>
				<s-input
					ref="dayWorkComparedDayWork"
					class="mb-4 min-w-[300px]"
					label="Day Work : Day Off"
					:disabled="true"
					v-model="data.ratioFactor.dayWorkComparedDayWork"
					placeholder="Day Work : Day Off"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('dayWorkComparedDayWork', v1, v2)"
				></s-input>
				<s-input
					ref="trainingAndOtherAYear"
					class="mb-4 min-w-[300px]"
					label="Training & other 1 Year"
					v-model="data.ratioFactor.trainingAndOtherAYear"
					placeholder="Training & other 1 Year"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('trainingAndOtherAYear', v1, v2)"
				></s-input>
			</div>
			<div class="w-100 content-input">
				<s-input
					ref="totalLossDays"
					class="mb-4 min-w-[300px]"
					label="Total Loss Days ( w/o other )"
					:disabled="true"
					v-model="data.ratioFactor.totalLossDays"
					placeholder="Total Loss Days ( w/o other )"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('totalLossDays', v1, v2)"
				></s-input>
				<s-input
					ref="effectiveWorkingDaysAYear"
					class="mb-4 min-w-[300px]"
					label="Effective Working Days a Year ( w/o other )"
					:disabled="true"
					v-model="data.ratioFactor.effectiveWorkingDaysAYear"
					placeholder="Effective Working Days a Year ( w/o other )"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('effectiveWorkingDaysAYear', v1, v2)"
				></s-input>
				<s-input
					ref="percentage"
					class="mb-4 min-w-[300px]"
					label="Percentage"
					v-model="data.ratioFactor.percentage"
					placeholder="Percentage"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('percentage', v1, v2)"
				></s-input>
			</div>
			<div class="w-100 content-input">
				<s-input
					ref="other"
					class="mb-4 min-w-[300px]"
					label="Other ( Sick/Permit/Alpha )"
					:disabled="true"
					v-model="data.ratioFactor.other"
					placeholder="Other ( Sick/Permit/Alpha )"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('other', v1, v2)"
				></s-input>
				<s-input
					ref="totalLossDaysWithOther"
					class="mb-4 min-w-[300px]"
					label="Total Loss Days ( with other )"
					:disabled="true"
					v-model="data.ratioFactor.totalLossDaysWithOther"
					placeholder="Total Loss Days ( with other )"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('totalLossDaysWithOther', v1, v2)"
				></s-input>
				<s-input
					ref="effectiveWorkingDaysAYearWithOther"
					class="mb-4 min-w-[300px]"
					label="Effective Working Days a Year ( EWD )"
					:disabled="true"
					v-model="data.ratioFactor.effectiveWorkingDaysAYearWithOther"
					placeholder="Effective Working Days a Year ( EWD )"
					kind="number"
					@change="(field, v1, v2, old, ctlRef) => onChangeRatioFactor('effectiveWorkingDaysAYearWithOther', v1, v2)"
				></s-input>
			</div>
			<!-- slot btn -->
			<div class="w-100 content-input">
				<s-input
					ref="ratioFactor"
					class="mb-4 min-w-[300px]"
					label="Ratio Factor"
					:disabled="true"
					read-only
					v-model="data.ratioFactor.ratioFactorView"
					placeholder="Ratio Factor"
				></s-input>
				
				<s-button
					class="btn_primary text-center"
					label="Calculate"
					@click="CalcRatioFactor"
				></s-button>
			</div>
		</s-modal>

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
				'EmployeeType',
				'Qty',
			]"
			:form-fields="[
				'EmployeeType',
				'Qty',
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
			<!-- slot btn grid -->
			<template #grid_header_buttons_1="{item}">
				<s-button
					class="btn_primary"
					label="Input Ratio Factor"
					@click="openModalRatioFactor"
				/>
			</template>
			<!-- slot grid -->
			<template #grid_EmployeeType="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'EmployeeType'
					)"
					:key="`gridemployeetypeinput ${hdrindex}`"
				>
					<s-input
						ref="refEmployeeType"
						hide-label
						use-list
						:lookup-url="`/tenant/masterdata/find?MasterDataTypeID=PTE`"
						lookup-key="_id"
						:lookup-labels="['_id', 'Name']"
						:lookup-searchs="['_id', 'Name']"
						@change="rowFieldChanged"
						v-model="item[hdr.input.field]"
					></s-input>
				</div>
			</template>

			<template #grid_Qty="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'Qty'
					)"
					:key="`gridQtyinput ${hdrindex}`"
				>
					<s-input
						ref="refQty"
						hide-label
						:ctl-ref="{rowIndex: item.index}"
						:field="hdr.input.field"
						:kind="hdr.input.kind"
						@change="rowFieldChanged"
						v-model="item[hdr.input.field]"
					></s-input>
				</div>
			</template>

			<!-- slot form -->
			<template #form_input_EmployeeType="{item}">
				<s-input
					ref="refEmployeeType"
					hide-label
					use-list
					:lookup-url="`/tenant/masterdata/find?MasterDataTypeID=PTE`"
					lookup-key="_id"
					:lookup-labels="['_id', 'Name']"
					:lookup-searchs="['_id', 'Name']"
					@change="rowFieldChanged"
					v-model="item.EmployeeType"
				></s-input>
			</template>

			<template #form_input_Qty="{item}">
				<s-input
					ref="refQty"
					hide-label
					kind="number"
					@change="rowFieldChanged"
					v-model="item.Qty"
				></s-input>
			</template>
		</data-list>
	</div>
</template>
<style>
	.content-input {
		display: flex;
    	gap: 12px;
	}
</style>
<script setup>
import {DataList, util, SInput, SButton, SModal} from "suimjs";
import {inject, ref, reactive, watch, onMounted} from "vue";

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
	ratioFactor: {
		totalDaysAYear: 366,
		totalDaysOffAYear: 6,
		totalActiveDaysAYear: 360, //calc
		AnnualLeave: 12,
		workingDays: 348, //calc
		durationOfOneDutyPeriod: 0, 
		siteLeave: 0, 
		oneDutyPeriodAndOneLeavPeriod: 0, //calc
		numberOfDutyAndLeavePeriodsInAYear: 0, //calc
		numberOfDutyDays: 0, //calc
		numberOfLeaveDays: 0, //calc
		dayWork: 0,
		dayOff: 0,
		dayWorkComparedDayWork: 0, //calc
		trainingAndOtherAYear: 0,
		totalLossDays: 0, //calc
		effectiveWorkingDaysAYear: 0, //calc
		percentage: 3,
		other: 0, //calc
		totalLossDaysWithOther: 0, //calc
		effectiveWorkingDaysAYearWithOther: 0, //calc
		ratioFactor: 0, //calc
		ratioFactorView: "0.00",
	},
	disableModal: false,
	listEmployeeType: [],
	employeeTypeDriver: "",
});

watch(
	() => props.modelValue,
	(nt) => {
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

// watch(
// 	() => data.records,
// 	(nt) => {
// 		console.log("data records: " , nt)
// 	}
// );

const listControl = ref(null);
const ratioFactorModal = ref(SModal);

const newRecord = () => {
	const records = listControl.value.getGridRecords();
	records.push({
		EmployeeType: "",
		Qty: 0,
		index: records.length,
	});

	listControl.value.setGridRecords(records);
	updateItems();
};

async function onFormFieldChanged(name, v1, v2, old, record) {
	updateItems();
}

function onGridRowSave(record, index) {
	record.suimRecordChange = false;

	const records = listControl.value.getGridRecords();
	records[index] = record;
	listControl.value.setGridRecords(records);
	updateItems();
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
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
	listControl.value.setGridRecord(
		record,
		listControl.value.getGridCurrentIndex()
	);

	updateItems();
}

function gridRefreshed() {
	//check
	listControl.value.setGridRecords(data.records);

	// disabledRatioFactor();
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

	// disabledRatioFactor();
}

function rowFieldChanged(name, v1, v2) {
	const currentIndex = data.currentIndex;
	const current = data.records[currentIndex];

	onGridRowFieldChanged(name, v1, v2, current, current);
	emit("rowFieldChanged", name, v1, v2, current, current);
	emit("update:modelValue", data.records);

	// if (v1 == data.employeeTypeDriver) {
	// 	openModalRatioFactor();
	// }
}


const AlterGridConfig = (gridconfigs) => {
	const records = data.records;

	const assetfield = gridconfigs.fields;

	data.defaultgridconfig = assetfield;
};

function closeModalRatioFactor() {
	ratioFactorModal.value.hide();
};
function openModalRatioFactor() {
	ratioFactorModal.value.show();
};

function onChangeRatioFactor(name, v1, v2) {
	// console.log(name,'-', v1, '-',v2);
	let ratioFactor = data.ratioFactor;
	ratioFactor[name] = v1;
	// console.log(ratioFactor[name]);
	ratioFactor.totalActiveDaysAYear = ratioFactor.totalDaysAYear - ratioFactor.totalDaysOffAYear; //Item 1 - Item 2
	ratioFactor.workingDays = ratioFactor.totalActiveDaysAYear - ratioFactor.AnnualLeave; //Item 3 - Item 4
	ratioFactor.oneDutyPeriodAndOneLeavPeriod = ratioFactor.durationOfOneDutyPeriod + ratioFactor.siteLeave; //Item 6 + Item 7
	ratioFactor.numberOfDutyAndLeavePeriodsInAYear = ratioFactor.workingDays / ratioFactor.oneDutyPeriodAndOneLeavPeriod; //Item 5 / Item 8
	ratioFactor.numberOfDutyDays = ratioFactor.durationOfOneDutyPeriod * ratioFactor.numberOfDutyAndLeavePeriodsInAYear; //Item 6 x Item 9
	ratioFactor.numberOfLeaveDays = ratioFactor.siteLeave * ratioFactor.numberOfDutyAndLeavePeriodsInAYear; //Item 7 x Item 9
	ratioFactor.dayWorkComparedDayWork = (ratioFactor.dayOff / (ratioFactor.dayOff + ratioFactor.dayWork )) * ratioFactor.numberOfDutyDays; //(item 13/(item13 +item12)) x item 10
	ratioFactor.totalLossDays = ratioFactor.totalDaysOffAYear + ratioFactor.AnnualLeave + ratioFactor.numberOfLeaveDays + ratioFactor.dayWorkComparedDayWork + ratioFactor.trainingAndOtherAYear; //Item ( 2+4+11+14+15 )
	ratioFactor.effectiveWorkingDaysAYear = ratioFactor.totalActiveDaysAYear - ratioFactor.totalLossDays; //Item 3 - Item 16
	ratioFactor.other = (ratioFactor.percentage / 100) * ratioFactor.effectiveWorkingDaysAYear;//Item 18 * 17
	ratioFactor.totalLossDaysWithOther = ratioFactor.totalLossDays + ratioFactor.other; //Item 16 + Item 19
	ratioFactor.effectiveWorkingDaysAYearWithOther = ratioFactor.totalDaysAYear - ratioFactor.totalLossDaysWithOther; //Item 1 - Item 20
	ratioFactor.ratioFactor = ratioFactor.totalActiveDaysAYear / ratioFactor.effectiveWorkingDaysAYearWithOther; //Item 3 / Item 21
	
	var num = ratioFactor.ratioFactor;
	ratioFactor.ratioFactorView = num.toLocaleString("sv-SE", { maximumFractionDigits: 2, minimumFractionDigits: 2 });
	data.ratioFactor = ratioFactor;
}

function CalcRatioFactor() {
	//check
	let records = listControl.value.getGridRecords();
	records.map(e => {
		// console.log("map", e)
		e.Qty = e.EmployeeType == data.employeeTypeDriver ? data.ratioFactor.ratioFactor : e.Qty;
		return e
	})
	data.records = [...records];
	// const committedRecords = records.filter(
	// 	(dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
	// );
	closeModalRatioFactor();
};

// function disabledRatioFactor() {
// 	// const records = data.records;
// 	const records = listControl.value.getGridRecords();
// 	data.disableModal = true;
// 	console.log("check disabled 1", records);
// 	console.log("check disabled 2", data.records);
// 	data.records.forEach(e => {
// 		console.log("check disabled 3", e.EmployeeType , data.employeeTypeDriver._id);
// 		if (e.EmployeeType == data.employeeTypeDriver._id) {
// 			data.disableModal = false;
// 		}
// 	});
// };

async function getsEmployeeType() {
	try {
		const dataresponse = await axios.post(
			`/tenant/masterdata/find?MasterDataTypeID=PTE`
		);
		
		data.listEmployeeType = dataresponse.data;
		dataresponse.data.forEach(e => {
			if(e.Name.trim() == "Driver") {
				data.employeeTypeDriver =  e._id;
			}
		});
	} catch (error) {
		util.showError(error);
	}
};

onMounted(() => {
	getsEmployeeType();
})
</script>
