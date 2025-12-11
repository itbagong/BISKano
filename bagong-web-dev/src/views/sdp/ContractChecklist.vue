<template>
	<div class="w-full">
		<!-- {{data.itemSalesOrder}} -->
		<data-list
			v-if="data.openSection == 'contract-checklist'"
			v-show="data.openSection == 'contract-checklist'"
			class="card"
			ref="listControl"
			:title="data.titleForm"
			grid-config="/sdp/contract-checklist/gridconfig"
			form-config="/sdp/contract-checklist/formconfig"
			grid-read="/sdp/contract-checklist/gets"
			form-read="/sdp/contract-checklist/get"
			grid-mode="grid"
			grid-delete="/sdp/contract-checklist/delete"
			form-keep-label
			grid-hide-sort
			form-insert="/sdp/contract-checklist/insert"
			form-update="/sdp/contract-checklist/update"
			:form-fields="[
				'SalesOrderRefNo',
				'SalesOrderDate',
				'SalesOrderName',
				'SPKNo',
				'CustomerID',
				'CustomerName',
				'CustomerAddress',
				'CustomerCity',
				'CustomerProvince',
				'CustomerCountry',
				'CustomerZipcode',
				'Dimension',
			]"
			:form-tabs-new="['General', 'Checklist', 'Attachment']"
			:form-tabs-edit="['General', 'Checklist', 'Attachment']"
			:form-tabs-view="['General', 'Checklist', 'Attachment']"
			:init-app-mode="data.appMode"
			:init-form-mode="data.formMode"
			:form-initial-tab="data.initialTab"
			@formNewData="newRecord"
			@formEditData="editRecord"
			@pre-save="preSave"
			@grid-refreshed="gridRefreshed"
			@controlModeChanged="onCancelForm"
			@alter-form-config="onalterFormConfig"
			@post-save="onPostSave"
		>
			<template #form_input_SalesOrderRefNo="{item}">
				<s-input
					use-list
					ref="refInput"
					label="Sales Order Ref No."
					class="w-full"
					v-model="data.itemSalesOrder.SalesOrderRefNo"
					:lookup-url="`/sdp/salesorder/find`"
					lookup-key="SalesOrderNo"
					:lookup-labels="['SalesOrderNo', 'Name']"
					:lookup-searchs="['SalesOrderNo', 'Name']"
					@change="
						(field, v1, v2, old, ctlRef) => {
							getSalesOrderData(v1, item);
						}
					"
				></s-input>
			</template>
			<template #form_input_SalesOrderDate="{item}">
				<s-input
					use-list
					ref="refInput"
					label="Sales Order Date"
					class="w-full"
					v-model="data.itemSalesOrder.SalesOrderDate"
					disabled
				></s-input>
			</template>
			<template #form_input_SalesOrderName="{item}">
				<s-input
					use-list
					ref="refInput"
					label="Sales Order Name"
					class="w-full"
					v-model="data.itemSalesOrder.SalesOrderName"
					disabled
				></s-input>
			</template>
			<template #form_input_SPKNo="{item}">
				<s-input
					use-list
					ref="refInput"
					label="SPK/PO/Contract No"
					class="w-full"
					v-model="data.itemSalesOrder.SPKNo"
					disabled
				></s-input>
			</template>
			<template #form_input_CustomerID="{item}">
				<s-input
					use-list
					ref="refInput"
					label="Customer"
					class="w-full"
					v-model="data.customerSelectedItem.CustomerID"
					lookup-url="/tenant/customer/find"
					:lookup-payload-builder="payloadBuilderCustomer"
					lookup-key="_id"
					:lookup-labels="['_id', 'Name']"
					:lookup-search="['_id', 'Name']"
					@change="
						(field, v1, v2, old, ctlRef) => {
							getCustomerData(v1, item);
						}
					"
				></s-input>
			</template>

			<template #form_tab_Attachment="{item}">
				<s-grid-attachment
					:journal-id="item._id"
					journal-type="Contract Checklist"
					ref="gridAttachment"
					v-model="item.Attachment"
				></s-grid-attachment>
			</template>
			<template #form_input_CustomerName="{item}">
				<s-input
					use-list
					ref="refInput"
					label="Name"
					class="w-full"
					v-model="data.customerSelectedItem.CustomerName"
					disabled
				></s-input>
			</template>
			<template #form_input_CustomerAddress="{item}">
				<s-input
					use-list
					ref="refInput"
					label="Address"
					class="w-full"
					v-model="data.customerSelectedItem.CustomerAddress"
					disabled
				></s-input>
			</template>
			<template #form_input_CustomerCity="{item}">
				<s-input
					use-list
					ref="refInput"
					label="City"
					class="w-full"
					v-model="data.customerSelectedItem.CustomerCity"
					disabled
				></s-input>
			</template>
			<template #form_input_CustomerProvince="{item}">
				<s-input
					use-list
					ref="refInput"
					label="Province"
					class="w-full"
					v-model="data.customerSelectedItem.CustomerProvince"
					disabled
				></s-input>
			</template>
			<template #form_input_CustomerCountry="{item}">
				<s-input
					use-list
					ref="refInput"
					label="Country"
					class="w-full"
					v-model="data.customerSelectedItem.CustomerCountry"
					disabled
				></s-input>
			</template>
			<template #form_input_CustomerZipcode="{item}">
				<s-input
					use-list
					ref="refInput"
					kind="number"
					label="Zipcode"
					class="w-full"
					v-model="data.customerSelectedItem.CustomerZipcode"
					disabled
				></s-input>
			</template>
			 <template #form_input_Dimension="{ item }">
				<dimension-editor
				v-model="item.Dimension"
				:default-list="profile.Dimension"
				></dimension-editor>
			 </template>
			<template #form_tab_Checklist="{item}">
				<ContractStatusGrid
					ref="ContractStatusConfig"
					v-if="data.checklistShow == true"
					v-show="data.checklistShow == true"
					:model-value="data.woSelectedItem"
					:is-edited="data.isEdited"
					:open-section="data.openSection"
					:salesOrderRefNo="data.itemSalesOrder.SalesOrderRefNo"
					:salesOrderId="data.itemSalesOrder.Id"
					@update:open-section="(newValue) => (data.openSection = newValue)"
					@update:model-value="updateModelValue"
					@update:item-selected="itemSelectChecklist"
					:form-mode="data.formMode"
				/>
				<div v-else>
					<div class="nodata">
						Please Choose field "Sales Order Ref No" first!!
					</div>
				</div>
			</template>
		</data-list>
		<ContractChecklistBast
			v-else-if="data.openSection == 'bast-form'"
			v-show="data.openSection == 'bast-form'"
			ref="formContractChecklistBast"
			:is-edited="data.isEdited"
			:item-sales-order="data.itemSalesOrder"
			:item-customer="data.customerSelectedItem"
			:asset-id="data.assetIDSelected"
			v-model="data.woSelectedItem.Bast"
			@update:cancelBast="cancelBast"
			@update:isEdited="isEdited"
			form-config1="/sdp/contract-checklist/bast1/formconfig"
			form-config2="/sdp/contract-checklist/bast2/formconfig"
			form-config3="/sdp/contract-checklist/bast3/formconfig"
			form-config5="/sdp/contract-checklist/bast5/formconfig"
		>
		</ContractChecklistBast>
		<!-- {{data.checklistRecord}} -->
	</div>
</template>

<script setup lang="ts">
import {reactive, ref, watch, inject, onMounted} from "vue";
import {layoutStore} from "../../stores/layout.js";
import moment from "moment";
import {authStore} from "@/stores/auth.js";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import {DataList, util, SForm, SInput, createFormConfig, SButton} from "suimjs";
import ContractStatusGrid from "./widget/ContractChecklistStatusGrid.vue";
import ContractChecklistBast from "./widget/ContractChecklistBast.vue";
import Checklist from "@/components/common/Checklist.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";

import {useRoute} from "vue-router";

layoutStore().name = "tenant";

const featureID = "ContractChecklist";
// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})
const profile = authStore().getRBAC(featureID);

const refInput = ref(null);
const ContractStatusConfig = ref(null)
const listControl = ref(null as any);
const checkListConfig = ref(null as any);
const gridAttachment = ref(SGridAttachment);
const formContractChecklistBast = ref(null as any);
const lineConfig = ref(null);
const axios = inject("axios");
const route = useRoute();

const roleID = [
	(v) => {
		let vLen = 0;
		let consistsInvalidChar = false;

		v.split("").forEach((ch) => {
			vLen++;
			const validCar =
				"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
					ch
				) >= 0;
			if (!validCar) consistsInvalidChar = true;
		});

		if (vLen < 3 || consistsInvalidChar)
			return "minimal length is 3 and alphabet only";
		return "";
	},
];

const data = reactive({
	title: null as any,
	appMode: "grid",
	formMode: "edit",
	titleForm: "Contract Checklist",
	bastRecord: {},
	isEdited: false,
	checklistRecord: [],
	woSelectedItem: [],
	checklistShow: false,
	assetIDSelected: "",
	customerSelectedItem: {
		CustomerID: "",
		CustomerName: "",
		CustomerAddress: "",
		CustomerCity: "",
		CustomerProvince: "",
		CustomerCountry: "",
		CustomerPhone: "",
		CustomerZipcode: 0,
	},
	jurnalType: {},
	openSection: "contract-checklist",
	record: [],
	initialTab: 0,
	itemSalesOrder: {
		Id: "",
		SalesOrderRefNo: "",
		SalesOrderDate: "",
		SalesOrderName: "",
		SPKNo: ""
	},
	form: {
		SalesOrderRefNo: "",
		SalesOrderDate: "",
		SalesOrderName: "",
		CustomerID: "",
		CustomerName: "",
		CustomerAddress: "",
		CustomerCity: "",
		CustomerProvince: "",
		CustomerCountry: "",
		CustomerZipcode: "",
	},
	searchData: {
		name: "",
		DateFrom: "",
		DateTo: "",
	},
	collection: {
		section1: [
			{
				key: "O - Ring",
				value: false,
			},
			{
				key: "FSI",
				value: false,
			},
			{
				key: "AutoLUB",
				value: false,
			},
			{
				key: "Blade",
				value: false,
			},
			{
				key: "Ripper",
				value: false,
			},
			{
				key: "Pontoon",
				value: false,
			},
		],
		section2: [
			{
				key: "Bucket",
				value: false,
			},
			{
				key: "AC",
				value: false,
			},
			{
				key: "Radio",
				value: false,
			},
			{
				key: "Kaca Spion",
				value: false,
			},
			{
				key: "Kaca Spion Kiri",
				value: false,
			},
			{
				key: "Suction Hose",
				value: false,
			},
		],
		section3: [
			{
				key: "Kaca Spion Dalam",
				value: false,
			},
			{
				key: "Lampu Sein Kiri",
				value: false,
			},
			{
				key: "Lampu Sein kanan",
				value: false,
			},
			{
				key: "Kunci Roda",
				value: false,
			},
			{
				key: "Part Manual / Literatur",
				value: false,
			},
			{
				key: "Winch",
				value: false,
			},
		],
		section4: [
			{
				key: "Ban Cadangan",
				value: false,
			},
			{
				key: "Baterai",
				value: false,
			},
			{
				key: "Wiper",
				value: false,
			},
			{
				key: "Kunci Kontak",
				value: false,
			},
			{
				key: "Fire Extuingiser",
				value: false,
			},
			{
				key: "Lain-lain",
				value: false,
			},
		],
	},
	allowDelete: route.query.allowdelete === "true",
	formAssets: {},
	isSelected: false,
});

watch(
	() => route.query.objname,
	(nv) => {
		util.nextTickN(2, () => {
			listControl.value.refreshList();
			listControl.value.refreshForm();
		});
	}
);

watch(
	() => route.query.title,
	(nv) => {
		data.title = nv;
		listControl.value.setControlMode("grid");
	}
);

function openForm() {
	util.nextTickN(2, () => {
		listControl.value.setFormFieldAttr("_id", "rules", roleID);
	});
}

function updateValueBast(newValue) {
	data.bastRecord = newValue;
}

function isEdited(record) {

}

function cancelBast(newRecord, isEdited) {
	data.openSection = "contract-checklist";
	data.appMode = "form";
	data.initialTab = 1;

	data.checklistRecord.forEach((val, idx) => {
		if (newRecord._id == val._id) {
			data.checklistRecord.Bast = newRecord;
		}
	});

	
	data.isEdited = isEdited
	data.woSelectedItem = data.checklistRecord;

	// console.log("data.checklistRecord ===>", data.checklistRecord)
}

function updateModelValue(newValue) {
	data.checklistRecord = newValue;
}

function getJurnalType(id, record) {
	const url = "/sdp/salesorderjournaltype/get";
	axios.post(url, [id]).then(
		(r) => {
			data.jurnalType = r.data;
			if (record !== undefined) {
				record.PostingProfileID = r.data.PostingProfileID;
				record.DefaultOffset = r.data.DefaultOffset;
			}
		},
		(e) => util.showError(e)
	);
}

function preSave(record) {
	record.SalesOrderID = data.itemSalesOrder.Id
	record.SalesOrderRefNo = data.itemSalesOrder.SalesOrderRefNo;
	record.SalesOrderDate = data.itemSalesOrder.SalesOrderDate;
	record.SalesOrderName = data.itemSalesOrder.SalesOrderName;
	record.SPKNo = data.itemSalesOrder.SPKNo;

	record.CustomerID = data.customerSelectedItem.CustomerID;
	record.CustomerName = data.customerSelectedItem.CustomerName;
	record.CustomerAddress = data.customerSelectedItem.CustomerAddress;
	record.CustomerCity = data.customerSelectedItem.CustomerCity;
	record.CustomerProvince = data.customerSelectedItem.CustomerProvince;
	record.CustomerCountry = data.customerSelectedItem.CustomerCountry;
	record.CustomerPhone = data.customerSelectedItem.CustomerPhone;

	record.CustomerZipcode = parseInt(
		data.customerSelectedItem.CustomerZipcode.toString()
	);
	if (record.StartPeriod == null) {
		record.StartPeriod = moment().format("YYYY-MM-DDThh:mm:ssZ");
	}

	if (record.EndPeriod == null) {
		record.EndPeriod = moment().format("YYYY-MM-DDThh:mm:ssZ");
	}

	record.Checklist = data.checklistRecord;

	// console.log("presave =>", record)
}

function newRecord() {
	// checkListConfig.value.setGridRecords([])

	data.titleForm = "Create New Contract Checklist";
	data.initialTab = 0;
	data.checklistShow = false;
	data.isEdited = false

	data.woSelectedItem = [];
	data.checklistRecord = [];

	data.itemSalesOrder.Id = "";
	data.itemSalesOrder.SalesOrderRefNo = "";
	data.itemSalesOrder.SalesOrderDate = "";
	data.itemSalesOrder.SalesOrderName = "";
	data.itemSalesOrder.SPKNo = "";

	data.customerSelectedItem.CustomerID = "";
	data.customerSelectedItem.CustomerName = "";
	data.customerSelectedItem.CustomerAddress = "";
	data.customerSelectedItem.CustomerCity = "";
	data.customerSelectedItem.CustomerProvince = "";
	data.customerSelectedItem.CustomerCountry = "";
	data.customerSelectedItem.CustomerPhone = "";
	data.customerSelectedItem.CustomerZipcode = 0;

	openForm();
}

function editRecord(record) {
	data.titleForm = `Edit Contract Checklist | ${record._id}`;
	data.initialTab = 0;
	
	data.woSelectedItem = record.Checklist;
	data.checklistRecord = record.Checklist;
	// window.setInterval(() => {

		data.checklistShow = true;
		data.isEdited = true
	// }, 1000)

	data.itemSalesOrder.Id = record.Id
	data.itemSalesOrder.SalesOrderRefNo = record.SalesOrderRefNo;
	data.itemSalesOrder.SalesOrderDate = record.SalesOrderDate;
	data.itemSalesOrder.SalesOrderName = record.SalesOrderName;
	data.itemSalesOrder.SPKNo = record.SPKNo;

	data.customerSelectedItem.CustomerID = record.CustomerID;
	data.customerSelectedItem.CustomerName = record.CustomerName;
	data.customerSelectedItem.CustomerAddress = record.CustomerAddress;
	data.customerSelectedItem.CustomerCity = record.CustomerCity;
	data.customerSelectedItem.CustomerProvince = record.CustomerProvince;
	data.customerSelectedItem.CustomerCountry = record.CustomerCountry;
	data.customerSelectedItem.CustomerPhone = record.CustomerPhone;
	data.customerSelectedItem.CustomerZipcode = record.CustomerZipcode;

	// checkListConfig.value.setGridRecords(record.Checklist)
	openForm();
}

function itemSelectChecklist(oldValue, selectedValue, isEdited) {
	data.checklistRecord = oldValue;
	// console.log("old value =>", oldValue)
	// console.log("selected => ", selectedValue)

	data.woSelectedItem = selectedValue;
	data.assetIDSelected = selectedValue.AssetID;
	// data.woSelectedItem.Bast.collection = data.collection
	// console.log("data =>", data.woSelectedItem)
}

function payloadBuilderCustomer() {
	return {
		Take: 0,
		Sort: ["_id"],
		Select: ["_id", "Name"],
	};
}

async function getSalesOrderData(value, item) {
	data.checklistShow = false;
	data.woSelectedItem = [];
	data.checklistRecord = [];
	await axios.post(`/sdp/salesorder/find?SalesOrderNo=${value}`).then((res) => {
		const response = res.data[0];

		
		data.itemSalesOrder.Id = response._id;
		console.log("id =>", response._id)
		data.itemSalesOrder.SalesOrderRefNo = response.SalesOrderNo;
		data.itemSalesOrder.SalesOrderDate = response.SalesOrderDate;
		data.itemSalesOrder.SalesOrderName = response.Name;
		data.itemSalesOrder.SPKNo = response.SpkNo;

		item.SalesOrderDate = response.SalesOrderDate;
		item.SalesOrderName = response.Name;
		item.Dimension = response.Dimension

		item.CustomerID = response.CustomerID;
		data.checklistShow = true;
		getCustomerData(response.CustomerID, item);
		getJurnalType(response.JournalTypeID, item.ChecklistTemp);
	});
}

async function getCustomerData(idValue, item) {
	await axios.post(`/bagong/customer/get`, [idValue]).then((res) => {
		const response = res.data;

		data.customerSelectedItem.CustomerID = response._id;
		data.customerSelectedItem.CustomerName = response.Name;
		data.customerSelectedItem.CustomerAddress = response.Detail.Address;
		data.customerSelectedItem.CustomerCity = response.Detail.City;
		data.customerSelectedItem.CustomerProvince = response.Detail.Province;
		data.customerSelectedItem.CustomerCountry = response.Detail.Country;
		data.customerSelectedItem.CustomerPhone = response.Detail.Phone;
		data.customerSelectedItem.CustomerZipcode = response.Detail.Zipcode;

		item.CustomerName = response.Name;
		item.CustomerAddress = response.Detail.Address;
		item.CustomerCity = response.Detail.City;
		item.CustomerProvince = response.Detail.Province;
		item.CustomerCountry = response.Detail.Country;
		item.CustomerZipcode = response.Detail.Zipcode;
	});
}

function gridRefreshed() {
	// if (data.searchData.contract-checklistName != "" || data.searchData.DateFrom != "" || data.searchData.DateTo!= "") {
	// 	axios.post("/sdp/contract-checklist/gets-filter", data.searchData).then(async r => {
	// 		listControl.value.setGridRecords(r.data.data)
	// 	}, e => {
	// 		util.showError(e);
	// 	});
	// }
}

function onPostSave(record) {
	gridAttachment.value.Save();
}

// function onPostSave(record) {
// 	// let lines = lineConfig.value.getDataValue();
// 	// let payloadBatch = {
// 	//   contract-checklistID: record._id,
// 	//   Lines: lines.filter(function (b) {
// 	//     return b.contract-checklistID != "";
// 	//   }),
// 	// };
// 	// console.log("payloadBatch =>", lines)
// 	// axios
// 	//   .post("/sdp/contract-checklistline/save-multiple", payloadBatch)
// 	//   .then(
// 	//     (r) => {
// 	//       console.log("rrr =>", r)
// 	//     },
// 	//     (e) => {}
// 	//   );
// }

function onCancelForm(mode) {
	// checkListConfig.value.setGridRecords([])
	if (mode === "grid") {
		data.titleForm = "Contract Checklist";
	}
}

const addCancel = () => {
	data.formMode = "new";
	// record._id = "";
	// record.TrxDate = new Date();
	// record.Status = "";
	data.titleForm = "Create New Contract Checklist";
	// openForm(record);
};

const onsubmit = () => {};

function onalterFormConfig(r) {
	console.log(r);
}
function onAlterGridConfig(cfg: any) {
	cfg.setting.keywordFields = [
		"_id",
		"CustomerName",
		"SalesOrderRefNo",
		"SalesOrderName",
		"CustomerProvince",
	];
}
</script>
