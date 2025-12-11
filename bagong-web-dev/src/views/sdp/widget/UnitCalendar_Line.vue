<template>
	<s-grid
		ref="gridCtl"
		class="w-full"
		hide-select
		editor
		:config="data.listCfg"
		hide-search
		:hide-control="props.IsSOref"
		hide-detail
		hide-sort
		:hide-delete-button="props.IsSOref"
		hide-refresh-button
		:hide-new-button="props.IsSOref"
		no-confirm-delete
		:hide-action="props.IsSOref"
		auto-commit-line
		@select-data="selectData"
		@new-data="newData"
		@get-data="getData"
		@delete-data="handleGridRowDelete"
		@row-updated="gridRowUpdated"
		@row-field-changed="handleGridFieldChanged"
		@save-row-data="handleGridRowSave"
		@row-deleted="onGridRowDelete"
		@grid-refreshed="handleGridRefreshed"
	>
		<template #item_AssetUnitID="{item, header}">
			<div v-if="!item['IsItem']">
				{{
					props.assets
						.filter((asset) => item[header.input.field].includes(asset._id))
						.map((asset) => asset.Name)
						.join(" | ")
				}}
			</div>
			<s-input
				v-if="item['IsItem']"
				hide-label
				:ctl-ref="{rowIndex: item.Index}"
				:field="header.input.field"
				:kind="header.input.kind"
				:label="
					header.input.kind == 'checkbox' || header.input.kind == 'bool'
						? ''
						: header.input.label
				"
				:disabled="header.input.readOnly"
				:caption="header.input.caption"
				:hint="header.input.hint"
				:multi-row="header.input.multiRow"
				:use-list="header.input.useList"
				:items="header.input.items"
				:rules="header.input.rules"
				:required="header.input.required"
				:read-only="header.input.readOnly"
				:lookup-url="header.input.lookupUrl"
				:lookup-key="header.input.lookupKey"
				:allow-add="header.input.allowAdd"
				:lookup-format1="header.input.lookupFormat1"
				:lookup-format2="header.input.lookupFormat2"
				:decimal="header.input.decimal"
				:date-format="header.input.dateFormat"
				:multiple="header.input.multiple"
				:lookup-labels="header.input.lookupLabels"
				:lookup-searchs="
					header.input.lookupSearchs && header.input.lookupSearchs.length == 0
						? header.input.lookupLabels
						: header.input.lookupSearchs
				"
				@focus="rowFieldFocus"
				@change="rowFieldChanged"
				v-model="item[header.input.field]"
				ref="inputs"
			/>
		</template>

		<template #item_Duration="{item, header}">
			<s-input
				hide-label
				:ctl-ref="{rowIndex: item.Index}"
				:field="header.input.field"
				:kind="header.input.kind"
				:label="
					header.input.kind == 'checkbox' || header.input.kind == 'bool'
						? ''
						: header.input.label
				"
				:disabled="header.input.readOnly || props.IsSOref"
				:caption="header.input.caption"
				:hint="header.input.hint"
				:multi-row="header.input.multiRow"
				:use-list="header.input.useList"
				:items="header.input.items"
				:rules="header.input.rules"
				:required="header.input.required"
				:read-only="header.input.readOnly"
				:lookup-url="header.input.lookupUrl"
				:lookup-key="header.input.lookupKey"
				:allow-add="header.input.allowAdd"
				:lookup-format1="header.input.lookupFormat1"
				:lookup-format2="header.input.lookupFormat2"
				:decimal="header.input.decimal"
				:date-format="header.input.dateFormat"
				:multiple="header.input.multiple"
				:lookup-labels="header.input.lookupLabels"
				:lookup-searchs="
					header.input.lookupSearchs && header.input.lookupSearchs.length == 0
						? header.input.lookupLabels
						: header.input.lookupSearchs
				"
				@focus="rowFieldFocus"
				@change="rowFieldChanged"
				v-model="item[header.input.field]"
				ref="inputs"
			/>
		</template>
		<template #item_Uom="{item, header}">
			<s-input
				hide-label
				:ctl-ref="{rowIndex: item.Index}"
				:field="header.input.field"
				:kind="header.input.kind"
				:label="
					header.input.kind == 'checkbox' || header.input.kind == 'bool'
						? ''
						: header.input.label
				"
				:disabled="header.input.readOnly || props.IsSOref"
				:caption="header.input.caption"
				:hint="header.input.hint"
				:multi-row="header.input.multiRow"
				:use-list="header.input.useList"
				:items="header.input.items"
				:rules="header.input.rules"
				:required="header.input.required"
				:read-only="header.input.readOnly"
				:lookup-url="header.input.lookupUrl"
				:lookup-key="header.input.lookupKey"
				:allow-add="header.input.allowAdd"
				:lookup-format1="header.input.lookupFormat1"
				:lookup-format2="header.input.lookupFormat2"
				:decimal="header.input.decimal"
				:date-format="header.input.dateFormat"
				:multiple="header.input.multiple"
				:lookup-labels="header.input.lookupLabels"
				:lookup-searchs="
					header.input.lookupSearchs && header.input.lookupSearchs.length == 0
						? header.input.lookupLabels
						: header.input.lookupSearchs
				"
				@focus="rowFieldFocus"
				@change="rowFieldChanged"
				v-model="item[header.input.field]"
				ref="inputs"
			/>
		</template>
	</s-grid>
</template>
<script setup>
import {SGrid, SInput, loadGridConfig, util} from "suimjs";
import moment from "moment";
import {reactive, onMounted, inject, ref, watch} from "vue";
import "moment/locale/id";

const gridCtl = ref(null);
const axios = inject("axios");

const props = defineProps({
	IsSOref: {type: Boolean, default: () => false},
	modelValue: {type: Array, default: () => []},
	assets: {type: Array, default: () => []},
	gridConfig: {type: String, default: ""},
});

const emit = defineEmits({
	"update:modelValue": null,
});

const data = reactive({
	controlMode: props.initAppMode,
	formMode: props.formDefaultMode,
	listCfg: {},
	records: [],
	rowIndex: 0,
});

watch(
	() => props.modelValue,
	(nt) => {
		const records = (nt ?? []).map((dt, index) => {
			dt.suimRecordChange = false;
			return dt;
		});
		data.records = records;
		gridCtl.value.setRecords(records);
	}
);

function rowFieldFocus(name, v1, v2, ctlRef) {
	data.rowIndex = ctlRef.rowIndex;
}

function onGridRowDelete(record, index) {
	const records = gridCtl.value.getRecords();
	const newRecords = records
		.filter((dt) => {
			return dt.Index != record.Index;
		})
		.map((record, index) => ({...record, Index: index}));

	gridCtl.value.setRecords(newRecords);
	emit("update:modelValue", newRecords);
}

function rowFieldChanged(name, v1, v2) {
	if (name === "AssetUnitID") {
		const records = gridCtl.value.getRecords().map((record) => {
			if (record.Index === data.rowIndex) {
				if (props.IsSOref) {
					return {...record, Qty: v1.length, AssetUnitID: [...v1]};
				} else {
					return {...record, Qty: v1.length, maxQty: v1.length, AssetUnitID: [...v1]};
				}
			}

			return record;
		});

		gridCtl.value.setRecords(records);
		emit("update:modelValue", records);
	}

	if (name === "Duration") {
		const records = gridCtl.value.getRecords().map((record) => {
			if (record.Index === data.rowIndex) {
				const dateend = new Date(record.StartDate);

				switch (String(record.Uom).toLowerCase()) {
					case "days":
						dateend.setDate(v1 + record.StartDate.getDate());
						break;
					case "day":
						dateend.setDate(v1 + record.StartDate.getDate());
						break;
					case "month":
						dateend.setMonth(v1 + record.StartDate.getMonth());
						break;
					case "months":
						dateend.setMonth(v1 + record.StartDate.getMonth());
						break;

					default:
						break;
				}

				return {
					...record,
					Duration: v1,
					EndDate: moment(dateend).local().format("YYYY-MM-DDTHH:mm:ssZ"),
				};
			}

			return record;
		});

		gridCtl.value.setRecords(records);
		emit("update:modelValue", records);
	}

	if (name === "Uom") {
		const records = gridCtl.value.getRecords().map((record) => {
			if (record.Index === data.rowIndex) {
				const dateend = new Date(record.StartDate);

				switch (String(v1).toLowerCase()) {
					case "days":
						dateend.setDate(record.Duration + record.StartDate.getDate());
						break;
					case "day":
						dateend.setDate(record.Duration + record.StartDate.getDate());
						break;
					case "month":
						dateend.setMonth(record.Duration + record.StartDate.getMonth());
						break;
					case "months":
						dateend.setMonth(record.Duration + record.StartDate.getMonth());
						break;

					default:
						break;
				}

				return {
					...record,
					Uom: v1,
					EndDate: moment(dateend).local().format("YYYY-MM-DDTHH:mm:ssZ"),
				};
			}

			return record;
		});

		gridCtl.value.setRecords(records);
		emit("update:modelValue", records);
	}
}

function handleGridFieldChanged(name, v1, v2, record, old) {
	if (name == "StartDate") {
		const datestart = new Date(v1);
		const dateend = new Date(v1);

		switch (String(record.Uom).toLowerCase()) {
			case "days":
				dateend.setDate(datestart.getDate() + record.Duration);
				break;
			case "day":
				dateend.setDate(datestart.getDate() + record.Duration);
				break;
			case "month":
				dateend.setMonth(datestart.getMonth() + record.Duration);
				break;
			case "months":
				dateend.setMonth(datestart.getMonth() + record.Duration);
				break;

			default:
				break;
		}
		record.EndDate = moment(dateend).local().format("YYYY-MM-DDTHH:mm:ssZ");
	}
}

function refreshList() {
	loadGridConfig(axios, props.gridConfig).then(
		(r) => {
			data.listCfg = r;
			util.nextTickN(2, () => {
				if (props.modelValue.length > 0) {
					gridCtl.value.setRecords(props.modelValue);
				}
      });
		},
		(e) => util.showError(e)
	);
}

function selectData(dt, index) {}

function newData(dt) {
	let records = gridCtl.value.getRecords();
	records.push({
		Index: 0,
		StartDate: new Date(),
		AssetUnitID: [],
		IsItem: true,
		EndDate: new Date(),
		Uom: "",
		Descriptions: "",
		maxQty: 0,
		Qty: 0,
		suimRecordChange: false,
	});

	records = records.map((record, index) => ({
		...record,
		Index: index,
	}));


	gridCtl.value.setRecords(records);
	emit("update:modelValue", records);
}

function getData(keyword) {}

onMounted(() => {
	refreshList();
});
</script>
