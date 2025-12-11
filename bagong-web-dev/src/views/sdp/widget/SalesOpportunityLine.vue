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
			:grid-fields="[
				'Asset',
				'Item',
				'Spesifications',
			]"
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
			<template #grid_Asset="{item}">
				<s-input
					ref="refAsset"
					hide-label
					v-model="item.Asset"
					use-list
					:lookup-url="`/tenant/asset/find`"
					lookup-key="_id"
					:lookup-labels="['Name']"
					:lookup-searchs="['_id', 'Name']"
					class="w-100"
					:disabled="item.Item !== '' && item.Item !== null"
					@change="
						(field, v1, v2, old, ctlRef) => {
							onGridRowFieldChanged('Asset', v1, v2, old, item);
						}
					"
				></s-input>
			</template>

			<template #grid_Item="{item}">
				<s-input
					ref="refItem"
					hide-label
					v-model="item.Item"
					use-list
					:lookup-url="`/tenant/item/find`"
					lookup-key="_id"
					:lookup-labels="['Name']"
					:lookup-searchs="['_id', 'Name']"
					class="w-100"
					:disabled="item.Asset !== '' && item.Asset !== null"
					@change="
						(field, v1, v2, old, ctlRef) => {
							onGridRowFieldChanged('Item', v1, v2, old, item);
						}
					"
				></s-input>
			</template>

			<template #grid_Spesifications="{item}">
				<s-input
					:ref="data.spesification"
					hide-label
					v-model="item.Spesifications"
					use-list
					:lookup-url="
						item.Asset !== '' && item.Asset !== undefined
							? `/tenant/masterdata/find?MasterDataTypeID=SPC`
							: `/tenant/specvariant/find`
					"
					lookup-key="_id"
					:lookup-labels="['Name']"
					:lookup-searchs="['_id', 'Name']"
					class="w-100"
					multiple
					@change="
						(field, v1, v2, old, ctlRef) => {
							onGridRowFieldChanged('Spesifications', v1, v2, old, item);
						}
					"
					:lookup-payload-builder="
						(search) =>
						lookupPayloadBuilder(search, item.Spesifications, item)
					"
				></s-input>
			</template>
		</data-list>
	</div>
</template>

<script setup>
import {DataList, util, SInput} from "suimjs";
import {inject, ref, reactive, watch, onMounted} from "vue";
const axios = inject("axios");

const props = defineProps({
	gridConfig: {type: String, default: () => ""},
	formConfig: {type: String, default: () => ""},
	modelValue: {type: Array, default: () => []},
	trxType: {type: String, default: () => ""},
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
	changed: false,
	defaultgridconfig: {},
	currentIndex: -1,
	recordChanged: false,
	spesification: "",
	dataSpesification: [],
	// SalesPriceBook: {},
	cfg: {},
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
	const records = listControl.value.getGridRecords();
	records.push({
		Item: "",
		Asset: "",
		Description: "",
		Qty: 0,
		ContractPeriod: 0,
		index: records.length,
	});

	listControl.value.setGridRecords(records);
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

function rowFieldChanged(name, v1, v2) {
	const currentIndex = data.currentIndex;
	const current = data.records[currentIndex];

	onGridRowFieldChanged(name, v1, v2, current, current);
	// emit("rowFieldChanged", name, v1, v2, current, current);
	// emit("update:modelValue", data.items);
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
	const typeVal = typeof v1 === "string";
	if (name == "Item") {
		if (v1 && typeVal) {
			const itemtenant = getItemTenant(v1);
			const itemspec = getItemSpecs(v1);

			Promise.all([itemtenant, itemspec]).then(
				([respitemtenant, resitemspec]) => {
					record.Uom = respitemtenant.DefaultUnitID;
					record.Spesifications = [];
					data.dataSpesification = resitemspec.map((item) => item.SpecVariantID);
					// record.Spesifications = resitemspec.map((item) => item.SpecVariantID);
					data.spesification = record.Item;
				}
			);
		} else {
			if (record) {
				record.Uom = undefined;
				record.Spesifications = [];
				data.spesification = "";
				data.dataSpesification = [];
			} else {
				record = {
					PhysicalDimension: {},
					FinanceDimension: {},
				};
			}
		}

		if (record) {
			record.Asset = "";
		}
		if (!typeVal) {
			record.Item = "";
			record.Asset = "";
		}
	}

	if (name == "Asset") {
		if (record) {
			record.Item = "";
			record.Spesifications = [];
			data.spesification = record.Asset;
			data.dataSpesification = [];
		}
		if (!typeVal) {
			record.Asset = "";
			record.Item = "";
			record.Spesifications = [];
			data.spesification = "";
			data.dataSpesification = [];
		}
	}

	listControl.value.setGridRecord(
		record,
		listControl.value.getGridCurrentIndex()
	);
	updateItems();
}

async function getItemTenant(idItemTenant) {
	try {
		const dataresponse = await axios.post(
			`/tenant/item/find?_id=${idItemTenant}`
		);
		const item = dataresponse.data[0];

		return item;
	} catch (error) {
		util.showError(error);
	}
}

async function getItemSpecs(ItemID) {
	try {
		const dataresponse = await axios.post(
			`/tenant/itemspec/find?ItemID=${ItemID}`
		);
		return dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
}

function gridRefreshed() {
	listControl.value.setGridRecords(data.records);
}

const AlterGridConfig = (gridconfigs) => {
	data.cfg = gridconfigs;

	const records = data.records;

	gridconfigs.fields?.map((fields) => {
		if (props.trxType == 'Asset') {
			// hide in asset
			["Item"]?.map((flds) => {
				if (fields.field == flds) {
					fields.readType = "hide";
				}
			});

			// show in asset
			[
				"Asset",
				"Spesifications",
				"Description",
				"ContractPeriod",
				"Uom",
				"Qty",
			]?.map((flds) => {
				if (fields.field == flds) {
					fields.readType = "show";
				}
			})
		} else {
			// hide in item
			["Asset", "ContractPeriod"]?.map((flds) => {
				if (fields.field == flds) {
					fields.readType = "hide";
				}
			});

			// show in item
			[
				"Item",
				"Spesifications",
				"Description",
				"Uom",
				"Qty",
			]?.map((flds) => {
				if (fields.field == flds) {
					fields.readType = "show";
				}
			})
		}
		
		return fields;
	});

	const assetfield = gridconfigs.fields;

	data.defaultgridconfig = assetfield;
	if (records.length > 0) {
		records.map((record) => {
			const gridcfg = assetfield;

			return gridcfg;
		});
	}
};

function rowFieldFocus(name, v1, v2, ctlRef) {
	const prevIndex = data.currentIndex;
	const currentRowIndex = ctlRef.rowIndex;

	if (prevIndex == currentRowIndex) {
		return;
	}

	data.currentIndex = currentRowIndex;
	data.recordChanged = false;
}

function lookupPayloadBuilder(search, value, item) {
	const qp = {};
	if (search != "") data.filterTxt = search;
	qp.Take = 20;
	qp.Sort = ["Name"];
	qp.Select = ["_id", "Name"];

	//setting search
	if (search.length > 0) {
		qp.Where = {
			Op: "$or",
			items: qp.Select.map((el) => {
				return {Field: el, Op: "$contains", Value: [search]};
			}),
		};
	}
	
	if (item.Item !== '' && item.Item !== null && data.dataSpesification.length > 0) {
		const whereExisting = {
			Op: "$or",
			items: data.dataSpesification.map((el) => {
				return {Field: "_id", Op: "$eq", Value: el};
			}),
		};
		const items = [];
		if (data.dataSpesification.length > 0) {
			items.push(whereExisting);
		}
		if (qp.Where != undefined) {
			items.push(qp.Where);
		}

		if (items.length > 0) {
			qp.Where = {Op: "$or", items: items};
		}
	} else if (item.Item !== '' && item.Item !== null && item.Spesifications.length > 0 && data.dataSpesification.length == 0) {
		const whereExisting = {
			Op: "$or",
			items: item.Spesifications.map((el) => {
				return {Field: "_id", Op: "$eq", Value: el};
			}),
		};
		const items = [whereExisting];
		if (qp.Where != undefined) {
			items.push(qp.Where);
		}

		if (items.length > 0) {
			qp.Where = {Op: "$or", items: items};
		}
	}

	return qp;
}

watch(() => props.trxType, (nv) => {
	util.nextTickN(2, () => {
		AlterGridConfig(data.cfg);
	});
})
</script>
