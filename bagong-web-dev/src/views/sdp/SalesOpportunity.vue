<template>
	<div class="w-full">
		<data-list
			class="card"
			ref="listControl"
			:title="data.titleForm"
			:form-hide-submit="false"
			grid-config="/sdp/salesopportunity/grid/gridconfig"
			form-config="/sdp/salesopportunity/formconfig"
			grid-read="/sdp/salesopportunity/gets"
			form-read="/sdp/salesopportunity/get"
			grid-mode="grid"
			grid-delete="/sdp/salesopportunity/delete"
			form-keep-label
			form-insert="/sdp/salesopportunity/insert"
			form-update="/sdp/salesopportunity/update"
			grid-sort-field="Created"
			grid-sort-direction="desc"
			:grid-fields="[
				'SalesType',
				'SalesStage',
				'OppStatus',
				'OppStatusReason',
				'Customer',
			]"
			:form-fields="['Dimension', 'Customer']"
			:init-app-mode="data.appMode"
			:init-form-mode="data.formMode"
			:form-tabs-new="['General', 'Lines']"
			:form-tabs-edit="data.formEdit"
			:formInitialTab="data.formInitialTab"
			:grid-custom-filter="customFilter"
			form-default-mode="edit"
			@formNewData="newRecord"
			@formEditData="editRecord"
			@formFieldChange="onFormFieldChanged"
			@preSave="onPreSave"
			@post-save="onPostSave"
			@controlModeChanged="onControlModeChanged"
			@alterGridConfig="onAlterGridConfig"
			:grid-hide-new="!profile.canCreate"
			:grid-hide-edit="!profile.canUpdate"
			:grid-hide-delete="!profile.canDelete"
			stay-on-form-after-save
		>
			<template #grid_header_search="{config}">
				<s-input
					ref="refCustomer"
					v-model="data.search.Customer"
					lookup-key="_id"
					label="Customer"
					class="w-full"
					multiple
					use-list
					:lookup-url="`/tenant/customer/find`"
					:lookup-labels="['Name']"
					:lookup-searchs="['_id', 'Name']"
					@change="refreshData"
				></s-input>
				<s-input
					ref="refName"
					v-model="data.search.Name"
					lookup-key="_id"
					label="Text"
					class="w-[300px]"
					@keyup.enter="refreshData"
				></s-input>
				<s-input
					kind="date"
					label="Date From"
					v-model="data.search.DateFrom"
					@change="refreshData"
				></s-input>
				<s-input
					kind="date"
					label="Date To"
					v-model="data.search.DateTo"
					@change="refreshData"
				></s-input>
				<s-input
					ref="refStatus"
					v-model="data.search.Status"
					lookup-key="_id"
					label="Status"
					class="w-[250px]"
					use-list
					:lookup-url="`/tenant/masterdata/find?MasterDataTypeID=SOS`"
					:lookup-labels="['Name']"
					:lookup-searchs="['_id', 'Name']"
					@change="refreshData"
				></s-input>
			</template>
			<!-- slot grid -->
			<template #grid_SalesType="{item}">
				{{ setGridName(item, "SalesType") }}
			</template>
			<template #grid_SalesStage="{item}">
				{{ setGridName(item, "SalesStage") }}
			</template>
			<template #grid_OppStatus="{item}">
				{{ setGridName(item, "OppStatus") }}
			</template>
			<template #grid_OppStatusReason="{item}">
				{{ setGridName(item, "OppStatusReason") }}
			</template>
			<template #grid_Customer="{item}">
				{{ setGridName(item, "Customer") }}
			</template>

			<!-- slot form -->
			<template #form_input_Dimension="{item}">
				<dimension-editor
					v-model="item.Dimension"
					:default-list="profile.Dimension"
				></dimension-editor>
			</template>
			<template #form_input_Customer="{item}">
				<SelesOpportunityCustomer
					v-model="item.Customer"
				></SelesOpportunityCustomer>
			</template>

			<!-- slot tab -->
			<template #form_tab_Lines="{item}">
				<SalesOpportunityLine
					v-model="item.Lines"
					:item="item"
					grid-config="/sdp/opportunity/line/gridconfig"
					form-config="/sdp/opportunity/line/formconfig"
					:trx-type="data.record.TransactionType"
				></SalesOpportunityLine>
			</template>
			<template #form_tab_Competitor="{item}">
				<CompetitorOpportunity
					v-model="item.Competitors"
					:item="item"
					grid-config="/sdp/opportunity/competitor/gridconfig"
					form-config="/sdp/opportunity/competitor/formconfig"
				></CompetitorOpportunity>
			</template>
			<template #form_tab_Event="{item}">
				<EventOpportunity
					v-model="item.Events"
					:item="item"
					grid-config="/sdp/opportunity/event/gridconfig"
					form-config="/sdp/opportunity/event/formconfig"
				></EventOpportunity>
			</template>
			<template #form_tab_Bond="{item}">
				<BondOpportunity
					v-model="item.Bonds"
					:item="item"
					grid-config="/sdp/opportunity/bond/gridconfig"
					form-config="/sdp/opportunity/bond/formconfig"
				></BondOpportunity>
			</template>
			<template #form_tab_Attachment="{item}">
				<s-grid-attachment
					:journal-id="item._id"
					journal-type="Sales Opportunity"
					ref="gridAttachment"
					v-model="item.Attachment"
				></s-grid-attachment>
			</template>
			<!-- 
      <template #form_tab_Bond="{ item }">
        <BondOpportunity
          ref="BondConfig"
          v-model="data.record"
          :item="item"
          :itemID="item._id"
          :hide-detail="true"
          grid-config="/sdp/opportunity/bond/gridconfig"
          :grid-read="'/sdp/opportunity/bond/gets?OpportunityId=' + item._id"
        ></BondOpportunity>
      </template>
      <template #form_tab_Attachment="{ item }">
        <SelesOpportunityAttachment
          ref="AttachmentConfig"
          v-model="data.record"
          :item="item"
          :itemID="item._id"
          :hide-detail="true"
        ></SelesOpportunityAttachment>
      </template> -->
			<!-- <template #form_buttons_2="{ item }">
        <div class="flex gap-[2px] ml-2">
          <s-button
            class="btn_primary"
            label="Save & Submit"
            @click="postSubmit(item)"
          />
          <s-button
            class="btn_primary"
            label="Process"
            @click="postProcess(item)"
          />
          <s-button
            class="btn_primary"
            label="Approve"
            @click="postApprove(item)"
          />
        </div>
      </template> -->
		</data-list>
	</div>
</template>

<script setup>
import {reactive, ref, inject, computed, watch, onMounted} from "vue";
import {layoutStore} from "@/stores/layout.js";
import {
	createFormConfig,
	DataList,
	util,
	SInput,
	SButton,
	SCard,
	SForm,
} from "suimjs";
import {authStore} from "@/stores/auth.js";

import SalesOpportunityLine from "./widget/SalesOpportunityLine.vue";
import CompetitorOpportunity from "./widget/CompetitorOpportunity.vue";
import EventOpportunity from "./widget/EventOpportunity.vue";
import BondOpportunity from "./widget/BondOpportunity.vue";
import SelesOpportunityAttachment from "./widget/SelesOpportunityAttachment.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import SelesOpportunityCustomer from "./widget/SelesOpportunityCustomer.vue";
import GridAttachment from "@/components/common/GridAttachment.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import moment from "moment";

layoutStore().name = "tenant";

const featureID = "SalesOpportunity";
// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})
const profile = authStore().getRBAC(featureID);

const listControl = ref(null);
const gridAttachment = ref(SGridAttachment);
const lineConfig = ref(null);
const axios = inject("axios");
let customFilter = computed(() => {
	const filters = [];
	if (data.search.Customer !== null && data.search.Customer.length > 0) {
		filters.push({
			Field: "Customer",
			Op: "$contains",
			Value: data.search.Customer,
		});
	}
	if (data.search.Name !== null && data.search.Name !== "") {
		filters.push({
			Field: "Name",
			Op: "$contains",
			Value: [data.search.Name],
		});
	}
	if (
		data.search.DateFrom !== null &&
		data.search.DateFrom !== "" &&
		data.search.DateFrom !== "Invalid date"
	) {
		filters.push({
			Field: "OppStartDate",
			Op: "$gte",
			Value: moment(data.search.DateFrom).utc().format("YYYY-MM-DDT00:mm:00Z"),
		});
	}
	if (
		data.search.DateTo !== null &&
		data.search.DateTo !== "" &&
		data.search.DateTo !== "Invalid date"
	) {
		filters.push({
			Field: "OppEndDate",
			Op: "$lte",
			Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
		});
	}
	if (data.search.Status !== null && data.search.Status !== "") {
		filters.push({
			Field: "OppStatus",
			Op: "$eq",
			Value: data.search.Status,
		});
	}
	if (filters.length == 1) return filters[0];
	else if (filters.length > 1) return {Op: "$and", Items: filters};
	else return null;
});
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
	appMode: "grid",
	formMode: "edit",
	titleForm: "Sales Opportunity",
	formInitialTab: 0,
	record: {
		_id: "",
		OppStartDate: new Date(),
		OppEndDate: new Date(),
		CompanyID: "",
		TransactionType: "",
	},
	search: {
		Customer: [],
		Name: "",
		DateFrom: null,
		DateTo: null,
		Status: "",
	},
	stayOnForm: true,
	salesTypeList: [],
	salesStageList: [],
	oppStatusList: [],
	oppStatusReasonList: [],
	customerList: [],
	formEdit: ["General", "Lines", "Competitor", "Event", "Bond", "Attachment"],
});

function newRecord(record) {
	record._id = "";
	record.OppStartDate = new Date();
	record.OppEndDate = new Date();
	record.CompanyID = "DEMO00";
	record.TransactionType = "Asset";
	data.formMode = "new";
	data.titleForm = "Create New Sales Opportunity";
	openForm(record);
}

function editRecord(record) {
	data.formMode = "edit";
	data.record = record;

	const nontender = data.salesTypeList.find((e) => {
		return e.Name == "Non Tender";
	});
	// console.log(nontender)
	if (record.SalesType == nontender._id) {
		data.formEdit = ["General", "Lines", "Competitor", "Attachment"];
	} else {
		data.formEdit = [
			"General",
			"Lines",
			"Competitor",
			"Event",
			"Bond",
			"Attachment",
		];
	}
	// console.log(record)
	data.titleForm = `Edit Sales Opportunity - ${record.OpportunityNo} | ${record.Name}`;
	// data.titleForm = `Edit Sales Opportunity | ${record.Name}`;
	openForm(record);
}

function openForm() {
	util.nextTickN(2, () => {
		listControl.value.setFormFieldAttr("_id", "rules", roleID);
	});
}

async function getSalesType() {
	try {
		const dataresponse = await axios.post(
			`/tenant/masterdata/find?MasterDataTypeID=SOOT`
		);
		data.salesTypeList = dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
}

async function getSalesStage() {
	try {
		const dataresponse = await axios.post(
			`/tenant/masterdata/find?MasterDataTypeID=SOSS`
		);
		data.salesStageList = dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
}

async function getOppStatus() {
	try {
		const dataresponse = await axios.post(
			`/tenant/masterdata/find?MasterDataTypeID=SOS`
		);
		data.oppStatusList = dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
}

async function getOppStatusReason() {
	try {
		const dataresponse = await axios.post(
			`/tenant/masterdata/find?MasterDataTypeID=SOSR`
		);
		data.oppStatusReasonList = dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
}

async function getCustomer() {
	try {
		const dataresponse = await axios.post(`/tenant/customer/find`);
		data.customerList = dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
}

function setGridName(item, field) {
	if (field == "SalesType") {
		const res = data.salesTypeList.find((e) => {
			return e._id == item.SalesType;
		});
		if (res === undefined) {
			return "";
		}
		return res.Name;
	} else if (field == "SalesStage") {
		const res = data.salesStageList.find((e) => {
			return e._id == item.SalesStage;
		});
		if (res === undefined) {
			return "";
		}
		return res.Name;
	} else if (field == "OppStatus") {
		const res = data.oppStatusList.find((e) => {
			return e._id == item.OppStatus;
		});
		if (res === undefined) {
			return "";
		}
		return res.Name;
	} else if (field == "OppStatusReason") {
		const res = data.oppStatusReasonList.find((e) => {
			return e._id == item.OppStatusReason;
		});
		if (res === undefined) {
			return "";
		}
		return res.Name;
	} else if (field == "Customer") {
		const res = data.customerList.find((e) => {
			return e._id == item.Customer;
		});
		if (res === undefined) {
			return "";
		}
		return res.Name;
	}
}
function onAlterGridConfig(cfg) {
	cfg.setting.idField = "Created";
	cfg.setting.sortable = ["_id", "Created", "OppStartDate"];
}

function refreshData() {
	util.nextTickN(2, () => {
		listControl.value.refreshGrid();
	});
}

const onFormFieldChanged = (name, v1, v2, old, record) => {
	if (name === "TransactionType") {
		record.TransactionType = v1;
	}
}

// function postSave(record) {
//   let payload = {
//     StockOpname: record,
//     Lines: [],
//   };
//   if (lineConfig.value && lineConfig.value.getDataValue()) {
//     let dv = lineConfig.value.getDataValue();
//     let WarehouseID = "";
//     if (record.InventoryDimension) {
//       WarehouseID = record.InventoryDimension.WarehouseID;
//     }
//     dv.map(function (d) {
//       d.InventoryDimension = {
//         WarehouseID: WarehouseID,
//         AisleID: d.AisleID,
//         SectionID: d.SectionID,
//         BoxID: d.BoxID,
//       };
//       return d;
//     });
//     payload.Lines = dv;
//   }
//   axios.post("/sdp/salesopportunity/save-opname", payload).then(
//     (r) => {
//       listControl.value.refreshForm();
//       listControl.value.setControlMode("grid");
//       listControl.value.refreshList();
//       return util.showInfo("stock opname has been successful save");
//     },
//     (e) => {
//       return util.showError(e);
//     }
//   );
// }

// function postSubmit(record) {
//   let payload = {
//     StockOpname: record,
//     Lines: [],
//   };
//   if (lineConfig.value && lineConfig.value.getDataValue()) {
//     let dv = lineConfig.value.getDataValue();
//     let WarehouseID = "";
//     if (record.InventoryDimension) {
//       WarehouseID = record.InventoryDimension.WarehouseID;
//     }
//     dv.map(function (d) {
//       d.InventoryDimension = {
//         WarehouseID: WarehouseID,
//         AisleID: d.AisleID,
//         SectionID: d.SectionID,
//         BoxID: d.BoxID,
//       };
//       return d;
//     });
//     payload.Lines = dv;
//   }
//   axios.post("/sdp/salesopportunity/submit", payload).then(
//     (r) => {
//       listControl.value.refreshForm();
//       listControl.value.setControlMode("grid");
//       listControl.value.refreshList();
//       return util.showInfo("stock opname has been successful submit");
//     },
//     (e) => {
//       return util.showError(e);
//     }
//   );
// }

// function postProcess(record) {
//   let payload = {
//     StockOpnameID: record._id,
//     StockOpnameDate: record.StockOpnameDate,
//   };
//   axios.post("/sdp/salesopportunity/process", payload).then(
//     (r) => {
//       data.stayOnForm = false;
//       listControl.value.refreshForm();
//       listControl.value.setControlMode("grid");
//       listControl.value.refreshList();
//     },
//     (e) => {
//       return util.showError(e);
//     }
//   );
// }

// function postAskReview(record) {
//   if (checkItemExists()) {
//     return util.showError("there are the same items and sku");
//   }
//   let payload = {
//     StockOpname: record,
//     Lines: [],
//   };
//   if (lineConfig.value && lineConfig.value.getDataValue()) {
//     let dv = lineConfig.value.getDataValue();
//     let WarehouseID = "";
//     if (record.InventoryDimension) {
//       WarehouseID = record.InventoryDimension.WarehouseID;
//     }
//     dv.map(function (d) {
//       d.InventoryDimension = {
//         WarehouseID: WarehouseID,
//         AisleID: d.AisleID,
//         SectionID: d.SectionID,
//         BoxID: d.BoxID,
//       };
//       return d;
//     });
//     payload.Lines = dv;
//   }
//   axios.post("/sdp/salesopportunity/ask-review", payload).then(
//     (r) => {
//       listControl.value.refreshForm();
//       listControl.value.setControlMode("grid");
//       listControl.value.refreshList();
//       return util.showInfo("stock opname has been successful submit");
//     },
//     (e) => {
//       return util.showError(e);
//     }
//   );
// }

// function postApprove(record) {
//   axios
//     .post("/sdp/salesopportunity/approve", { StockOpnameID: record._id })
//     .then(
//       (r) => {
//         listControl.value.refreshForm();
//         listControl.value.setControlMode("grid");
//         listControl.value.refreshList();
//       },
//       (e) => {
//         return util.showError(e);
//       }
//     );
// }
function onPreSave(record) {}
function onControlModeChanged(mode) {
	if (mode === "grid") {
		data.titleForm = "Sales Opportunity";
	}
}

function onPostSave(record) {
	gridAttachment.value.Save();
}

onMounted(() => {
	getSalesType();
	getSalesStage();
	getOppStatus();
	getOppStatusReason();
	getCustomer();
});
</script>
