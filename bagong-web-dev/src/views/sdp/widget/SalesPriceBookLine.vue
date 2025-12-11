<template>
	<div class="flex flex-col gap-2">
		<data-list
			ref="listControl"
			:title="'Sales Price Book Line'"
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
			:form-fields="[
				'Dimension',
				'ProductionYear',
				'AssetType',
				'AssetID',
				'ItemID',
				'SalesPrice',
				'MaxPrice',
				'MinPrice',
				'MaxDiscount',
			]"
			:grid-fields="[
				'Dimension',
				'ProductionYear',
				'AssetType',
				'AssetID',
				'ItemID',
				'SalesPrice',
				'MaxPrice',
				'MinPrice',
				'MaxDiscount',
			]"
			grid-config="/sdp/salespricebook/line/gridconfig"
			form-config="/sdp/salespricebook/line/formconfig"
			@grid-row-add="newRecord"
			@form-field-change="onFormFieldChanged"
			@form-edit-data="openForm"
			@grid-row-delete="onGridRowDelete"
			@grid-row-field-changed="onGridRowFieldChanged"
			@grid-refreshed="gridRefreshed"
			@grid-row-save="onGridRowSave"
			@post-save="onFormPostSave"
			form-focus
		>
			<template #grid_MaxPrice="{item}">
				<s-input class="min-w-[100px]" kind="number" v-model="item.MaxPrice">
				</s-input>
			</template>
			<template #grid_MinPrice="{item}">
				<s-input class="min-w-[100px]" kind="number" v-model="item.MinPrice">
				</s-input>
			</template>
			<template #grid_SalesPrice="{item}">
				<s-input class="min-w-[100px]" kind="number" v-model="item.SalesPrice">
				</s-input>
			</template>
			<template #grid_ProductionYear="{item}">
				<s-input
					use-list
					class="min-w-[100px]"
					v-if="data.tabAssetLoading[item.Idx].productionYear == false"
					:disabled="data.tabAssetTrigger[item.Idx].productionYear == true"
					v-model="item.ProductionYear"
					:items="data.productionYears[item.Idx]"
					@change="
						(field, v1, v2, old, ctlRef) => {
							onCheckProductionYear(v1, v2, item.Idx);
						}
					"
				>
				</s-input>
				<div v-else>Loading...</div>
			</template>
			<template #grid_AssetType="{item}">
				<!-- {{data.tabAssetTrigger[item.Idx].assetType == true}} -->
				<s-input
					use-list
					class="min-w-[150px]"
					v-if="data.tabAssetLoading[item.Idx].assetType == false"
					:disabled="data.tabAssetTrigger[item.Idx].assetType == true"
					v-model="item.AssetType"
					:items="data.assetTypes"
					@change="
						(field, v1, v2, old, ctlRef) => {
							data.tabAssetLoading[item.Idx].assetID == true;
							data.tabAssetTrigger[item.Idx].assetID = false;
							data.tabAssetLoading[item.Idx].assetID == false;
						}
					"
				>
				</s-input>
				<div v-else>Loading...</div>
			</template>
			<template #grid_AssetID="{item}">
				<s-input
					v-model="item.AssetID"
					use-list
					v-if="data.tabAssetLoading[item.Idx].assetID == false"
					:lookup-url="`/tenant/asset/find?AssetType=${item.AssetType}`"
					:disabled="data.tabAssetTrigger[item.Idx].assetID == true"
					lookup-key="_id"
					:lookup-searchs="['Name']"
					:lookup-labels="['Name']"
					class="min-w-[150px]"
				></s-input>
				<div v-else>Loading...</div>
			</template>
			<template #grid_ItemID="{item}">
				<s-input
					v-model="item.ItemID"
					use-list
					:lookup-url="`/tenant/item/find`"
					lookup-key="_id"
					:disabled="data.tabAssetTrigger[item.Idx].itemID == true"
					:lookup-labels="['Name']"
					class="min-w-[100px]"
					@change="
						(field, v1, v2, old, ctlRef) => {
							data.tabAssetTrigger[item.Idx].productionYear = true;
							data.tabAssetTrigger[item.Idx].assetType = true;
							data.tabAssetTrigger[item.Idx].assetID = true;
							item.AssetType = '';
							item.ProductionYear = '';
							item.AssetType = '';
						}
					"
				></s-input>
			</template>
			<template #form_input_MaxPrice="{item}">
				<s-input
					class="min-w-[100px]"
					kind="number"
					label="Max Price"
					v-model="item.MaxPrice"
				>
				</s-input>
			</template>
			<template #form_input_MinPrice="{item}">
				<s-input
					class="min-w-[100px]"
					kind="number"
					label="Min Price"
					v-model="item.MinPrice"
				>
				</s-input>
			</template>
			<template #form_input_SalesPrice="{item}">
				<s-input
					class="min-w-[100px]"
					kind="number"
					label="Sales Price"
					v-model="item.SalesPrice"
				>
				</s-input>
			</template>
			<template #form_input_ProductionYear="{item}">
				<s-input
					use-list
					class="min-w-[100px]"
					label="Production Year"
					v-if="data.tabAssetLoading[item.Idx].productionYear == false"
					:disabled="data.tabAssetTrigger[item.Idx].productionYear == true"
					v-model="item.ProductionYear"
					:items="data.productionYears[item.Idx]"
					@change="
						(field, v1, v2, old, ctlRef) => {
							onCheckProductionYear(v1, v2, item.Idx);
						}
					"
				>
				</s-input>
				<div v-else>Loading...</div>
			</template>
			<template #form_input_AssetType="{item}">
				<s-input
					use-list
					class="min-w-[150px]"
					label="Asset Type"
					v-if="data.tabAssetLoading[item.Idx].assetType == false"
					:disabled="data.tabAssetTrigger[item.Idx].assetType == true"
					v-model="item.AssetType"
					:items="data.assetTypes"
					@change="
						(field, v1, v2, old, ctlRef) => {
							data.tabAssetLoading[item.Idx].assetID == true;
							data.tabAssetTrigger[item.Idx].assetID = false;
							data.tabAssetLoading[item.Idx].assetID == false;
						}
					"
				>
				</s-input>
				<div v-else>Loading...</div>
			</template>
			<template #form_input_AssetID="{item}">
				<s-input
					v-model="item.AssetID"
					use-list
					label="Asset ID"
					:lookup-url="`/tenant/asset/find?AssetType=${item.AssetType}`"
					v-if="data.tabAssetTrigger[item.Idx].assetID == false"
					lookup-key="_id"
					:lookup-searchs="['Name']"
					:lookup-labels="['_id', 'Name']"
					class="w-full"
				></s-input>
			</template>
			<template #form_input_ItemID="{item}">
				<s-input
					v-model="item.AssetID"
					use-list
					v-if="data.tabAssetLoading[item.Idx].assetID == false"
					:lookup-url="`/tenant/asset/find?AssetType=${item.AssetType}`"
					:disabled="data.tabAssetTrigger[item.Idx].assetID == true"
					lookup-key="_id"
					label="Item ID"
					:lookup-labels="['Name']"
					class="min-w-[150px]"
				></s-input>
				<div v-else>Loading...</div>
			</template>
			<template #grid_Dimension="{item}">
				<DimensionText :dimension="item.Dimension" />
			</template>
			<template #form_input_Dimension="{item}">
				<div class="title section_title">Physical Dimension</div>
				<dimension-item
					v-model="item.PhysicalDimension"
					:readOnly="true"
				></dimension-item>
				<div class="title section_title">Finance Dimension</div>
				<dimension-item
					v-model="item.FinanceDimension"
					:readOnly="true"
				></dimension-item>
			</template>
		</data-list>
	</div>
</template>

<script setup>
import {onMounted, inject} from "vue";
import {reactive, ref} from "vue";
import {DataList, SButton, SInput, util} from "suimjs";
import DimensionItem from "./DimensionItem.vue";
import DimensionText from "../../../components/common/DimensionText.vue";
const axios = inject("axios");

const props = defineProps({
	modelValue: {type: Array, default: () => []},
	items: {type: Array, default: () => []},
	dimension: {type: Array, default: () => []},
});

const emit = defineEmits({
	"update:modelValue": null,
	recalc: null,
});

const listControl = ref(null);

const data = reactive({
	appMode: "grid",
	formMode: "edit",
	isEdited: false,
	assetAll: [],
	productionYears: [],
	disableProductionYears: true,
	tabAssetTrigger: [],
	tabAssetLoading: [],
	assetTypes: [],
	records: props.modelValue.map((dt) => {
		dt.suimRecordChange = false;
		return dt;
	}),
});

function newRecord() {
	const record = {};
	record.Idx =
		listControl.value.getGridRecords().length > 0
			? listControl.value.getGridRecords().length + 1
			: 0;
	record.ProductionYear = null;
	record.AssetType = "";
	record.AssetID = "";
	record.ItemID = "";
	record.MinPrice = 0;
	record.MaxPrice = 0;
	record.Quantity = 0;
	record.SalesPrice = 0;
	record.MaxDiscount = 0;
	record.Unit = "";
	//openForm(record);
	//return record;
	data.tabAssetTrigger[record.Idx] = {
		productionYear: false,
		assetType: true,
		assetID: true,
		itemID: false,
	};

	data.tabAssetLoading[record.Idx] = {
		productionYear: false,
		assetType: false,
		assetID: false,
		itemID: false,
	};

	data.records.push(record);
	listControl.value.setGridRecords(data.records);
	updateItems();
	getLookupProductionYear(record.Idx);
}

function openForm(record) {
	updateJournalType(record.JournalTypeID);
}

function onGridRowDelete(record, index) {
	const newRecords = data.records.filter((dt, idx) => {
		return idx != index;
	});
	data.records = newRecords;
	data.productionYears = data.productionYears.filter((dt, idx) => {
		return idx != index;
	});
	data.tabAssetLoading = data.tabAssetLoading.filter((dt, idx) => {
		return (idx = index);
	});
	data.tabAssetTrigger = data.tabAssetTrigger.filter((dt, idx) => {
		return (idx = index);
	});
	listControl.value.setGridRecords(data.records);
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

function getDataValueBatch() {
	return [...data.listBatch, ...data.tempListBatch];
}
function getDataValueTempBatch() {
	return data.tempListBatch;
}

function getDataValue() {
	return listControl.value.getGridRecords();
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

async function getLookupProductionYear(idx = 0) {
	// productionYears
	try {
		// data.tabAssetLoading[idx].productionYear = true
		// const dataresponse = await axios.post(
		// 	`bagong/asset/find`
		// );
		// let productionYearArr = [];
		// let assetIDs = []
		// dataresponse.data.forEach((v, i) => {
		//   productionYearArr.push(parseInt(v.DetailUnit.ProductionYear))
		//   assetIDs.push(v._id)
		//   data.assetAll.push(v)
		// })
		// console.log("asset ids =>", data.assetAll)

		// let uniqueAssetIDs = [...new Set(assetIDs)]
		// let uniqueProductionYear = [...new Set(productionYearArr)]

		data.tabAssetLoading[idx].productionYear = true;
		data.tabAssetLoading[idx].assetType = true;
		data.tabAssetLoading[idx].assetID = true;
		await axios.post(`tenant/asset/find`).then((el) => {
			let assetID = [];
			el.data.forEach((v, i) => {
				if (props.dimension.length > 0 && v.Dimension.length > 0) {
					if (props.dimension[0].Value == v.Dimension[0].Value) {
						assetID.push(v._id);
					}
				} else {
					assetID.push(v._id);
				}
			});

			assetID = [...new Set(assetID)];
			let paramAsset =
				props.dimension.length > 0
					? {
							Where: {
								Field: "_id",
								Op: "$in",
								Value: assetID,
							},
					  }
					: {};
			axios
				.post(`bagong/asset/find`, paramAsset)
				.then((result) => {
					let productionYearArr = [];
					result.data.forEach((v, i) => {
						productionYearArr.push(parseInt(v.DetailUnit.ProductionYear));
						data.assetAll.push(v);
					});

					let uniqueProductionYear = [...new Set(productionYearArr)];

					data.productionYears[idx] = uniqueProductionYear;
					data.disableProductionYears = false;

					data.tabAssetLoading[idx].productionYear = false;
					data.tabAssetLoading[idx].assetType = false;
					data.tabAssetLoading[idx].assetID = false;
				})
				.catch((error) => {
					console.error(error);
					throw error;
				});
		});
	} catch (error) {
		util.showError(error);
	}
}

function onGridRowSave(record, index) {
	record.suimRecordChange = false;
	data.records[index] = record;
	listControl.value.setGridRecords(data.records);
	updateItems();
}

function onFormFieldChanged(name, v1, v2, old, record) {
	getItemTenant(record._id).then((item) => {
		record.PhysicalDimension = item.PhysicalDimension;
		record.FinanceDimension = item.FinanceDimension;
	});

	updateItems();
}

function updateJournalType(id) {
	// baca journaltypeid dan assisgn object tag

	// hide jia tag1==NONE

	// hide jika tag2==NONE
	util.nextTickN(1, () => {
		listControl.value.removeFormField("TagObjectID2");
	});
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
	listControl.value.setGridRecord(
		record,
		listControl.value.getGridCurrentIndex()
	);
	console.log(name, v1, v2, old, record);
	getItemTenant(v1).then((item) => {
		record.PhysicalDimension = item.PhysicalDimension;
		record.FinanceDimension = item.FinanceDimension;
	});
	updateItems();
}

function updateItems() {
	const committedRecords = data.records.filter(
		(dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
	);
	emit("update:modelValue", committedRecords);
	emit("recalc");
}

function gridRefreshed() {
	listControl.value.setGridRecords(data.records);
}

function onCheckIsEdit() {
	if (data.records.length == 0) {
		data.isEdited = true;
	}
}

function onCheckAssetType(v1, v2, old) {
	data.tabAssetTrigger.assetID = false;
	let value = v2;
}

function onCheckProductionYear(v1, v2, idx) {
	data.tabAssetLoading[idx].assetType = true;
	data.tabAssetLoading[idx].assetID = true;
	let getAssetType = data.assetAll.filter((el) => {
		return el.DetailUnit.ProductionYear == v1;
	});

	let assetIDs = [];
	getAssetType.forEach((v, i) => {
		assetIDs.push(v._id);
	});

	axios
		.post(`tenant/asset/find`, {
			Where: {
				Field: "_id",
				Op: "$in",
				Value: assetIDs,
			},
		})
		.then((el) => {
			let assetType = [];
			el.data.forEach((v, i) => {
				if (props.dimension.length > 0 && v.Dimension.length > 0) {
					if (props.dimension[0].Value == v.Dimension[0].Value) {
						assetType.push(v.AssetType);
					}
				} else {
					assetType.push(v.AssetType);
				}
			});
			let uniqueAssetType = [...new Set(assetType)];

			axios
				.post(`tenant/masterdata/find`, {
					Where: {
						Field: "_id",
						Op: "$in",
						Value: uniqueAssetType,
					},
				})
				.then((res) => {
					let result = [];
					res.data.forEach((val, idx) => {
						result.push({
							text: val.Name,
							key: val._id,
						});
					});
					data.assetTypes = result;
					data.tabAssetLoading[idx].assetType = false;
					data.tabAssetLoading[idx].assetID = false;

					data.tabAssetTrigger[idx].assetType = false;
					data.tabAssetTrigger[idx].assetID = false;
					data.tabAssetTrigger[idx].itemID = true;
				});
		});
}

const getDataUnit = () => {
	if (data.records.length > 0) {
		data.records = data.records.map((record) => {
			if (record.ItemID) {
				getItemTenant(record.ItemID).then((item) => {
					record.PhysicalDimension = item.PhysicalDimension;
					record.FinanceDimension = item.FinanceDimension;
				});
			}

			return record;
		});
	}
};

defineExpose({
	getDataValue,
	getDataValueBatch,
});

onMounted(() => {
	// console.log("dimension =>", props.dimension)
	if (props.items.length > 0) {
		// if (props.items.length > 0) {
		for (let i = 0; i < props.items.length; i++) {
			data.records[i].Idx = i;
			// console.log(data.records[i].ProductionYear == 0 && data.records[i].ItemID != "")
			if (data.records[i].ProductionYear == 0 && data.records[i].ItemID != "") {
				// data.disableProductionYears = true
				data.tabAssetLoading[i] = {
					productionYear: true,
					assetType: true,
					assetID: true,
					itemID: false,
				};

				data.tabAssetTrigger[i] = {
					productionYear: true,
					assetType: true,
					assetID: true,
					itemID: false,
				};
			} else if (
				data.records[i].ProductionYear != 0 &&
				data.records[i].ItemID == ""
			) {
				data.tabAssetLoading[i] = {
					productionYear: false,
					assetType: false,
					assetID: false,
					itemID: true,
				};

				data.tabAssetTrigger[i] = {
					productionYear: false,
					assetType: false,
					assetID: false,
					itemID: true,
				};
			} else {
				data.tabAssetLoading[i] = {
					productionYear: false,
					assetType: false,
					assetID: false,
					itemID: false,
				};

				data.tabAssetTrigger[i] = {
					productionYear: false,
					assetType: true,
					assetID: true,
					itemID: false,
				};
			}
		}
		getLookupProductionYear();
	}
	getDataUnit();
	onCheckIsEdit();
});
</script>
