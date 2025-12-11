<template>
	<div class="w-full">
		<s-modal
			title="Sales Quotation"
			class="p-4"
			:display="false"
			ref="generateModal"
			hideButtons
			@submit="closeModalQuotation"
		>
			<s-input
				ref="siteID"
				useList
				class="mb-4 min-w-[240px]"
				label="Quotation"
				:disabled="data.formMode === 'edit'"
				lookup-url="/sdp/salesquotation/find?Status=POSTED"
				lookup-key="_id"
				:lookup-labels="['QuotationName']"
				:lookup-searchs="['_id', 'QuotationName']"
				v-model="data.cachequotation"
				placeholder="Quotation"
			></s-input>
			<s-button
				v-if="data.formMode === 'new'"
				class="w-full btn_success text-center"
				label="Create"
				@click="CreateWithQuotation"
			></s-button>
		</s-modal>

		<data-list
			v-show="data.appMode == 'grid'"
			stay-on-form-after-save
			class="card"
			ref="listControl"
			:title="data.titleForm"
			grid-config="/sdp/salesorder/grid/gridconfig"
			form-config="/sdp/salesorder/formconfig"
			grid-read="/sdp/salesorder/gets"
			form-read="/sdp/salesorder/get"
			grid-mode="grid"
			grid-delete="/sdp/salesorder/delete"
			form-keep-label
			form-insert="/sdp/salesorder/insert"
			form-update="/sdp/salesorder/update"
			:form-tabs-new="['General']"
			:form-tabs-edit="[
				'General',
				'Line',
				'Editor',
				'Manpower Fullfillment',
				'Breakdown Cost',
				'Attachment',
				'Checklist',
				'References',
			]"
			:form-tabs-view="[
				'General',
				'Line',
				'Editor',
				'Manpower Fullfillment',
				'Breakdown Cost',
				'Attachment',
				'Checklist',
				'References',
			]"
			:form-fields="[
				'SalesOpportunityRefNo',
				'Dimension',
				'CustomerID',
				'PostingProfileID',
				'JournalTypeID',
				'TaxCodes',
				'HeaderDiscountValue',
			]"
			:grid-fields="['SalesOpportunityRefNo', 'CustomerID', 'Status']"
			grid-sort-field="Created"
			grid-sort-direction="desc"
			:formInitialTab="data.formInitialTab"
			:init-app-mode="data.appMode"
			:init-form-mode="data.formMode"
			:grid-custom-filter="customFilter"
			@formNewData="newRecord"
			@formEditData="editRecord"
			@form-field-change="onFormFieldChanged"
			@controlModeChanged="onCancelForm"
			@pre-save="onPreSave"
			@post-save="onPostSave"
			@alterGridConfig="onAlterGridConfig"
      @alterFormConfig="onAlterFormConfig"
			:formHideSubmit="['', 'DRAFT'].indexOf(data.record.Status) === -1"
			:grid-hide-new="!profile.canCreate"
			:grid-hide-edit="!profile.canUpdate"
			:grid-hide-delete="!profile.canDelete"
		>
			<!-- <template #form_loader>
        <loader />
      </template> -->

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
					:items="['DRAFT', 'SUBMITTED', 'SENT', 'READY', 'POSTED']"
					@change="refreshData"
				></s-input>
			</template>

			<template #form_tab_References="{item, mode}">
				<References
					:ReferenceTemplate="data.jurnalType.ReferenceTemplateID"
					:readOnly="readOnly || item.mode == 'view'"
					v-model="item.References"
				></References>
			</template>

			<!-- slot grid -->
			<template #grid_SalesOpportunityRefNo="{item}">
				<s-input
					ref="refCustomerID"
					v-model="item.SalesOpportunityRefNo"
					class="w-50"
					read-only
					use-list
					:lookup-url="`/sdp/salesopportunity/find`"
					lookup-key="_id"
					:lookup-labels="['OpportunityNo']"
					:lookup-searchs="['_id', 'OpportunityNo']"
				></s-input>
			</template>
			<template #grid_CustomerID="{item}">
				<s-input
					ref="refCustomerID"
					v-model="item.CustomerID"
					class="w-50"
					read-only
					use-list
					:lookup-url="`/tenant/customer/find`"
					lookup-key="_id"
					:lookup-labels="['Name']"
					:lookup-searchs="['_id', 'Name']"
				></s-input>
			</template>
			<template #grid_Status="{item}">
				<status-text :txt="item.Status" />
			</template>

			<!-- slot form input -->
			<template #form_input_SalesOpportunityRefNo="{item}">
				<s-input
					label="Opportunity Ref No"
					ref="refCustomerID"
					v-model="item.SalesOpportunityRefNo"
					class="w-50"
					read-only
					use-list
					:lookup-url="`/sdp/salesopportunity/find`"
					lookup-key="_id"
					:lookup-labels="['OpportunityNo']"
					:lookup-searchs="['_id', 'OpportunityNo']"
				></s-input>
			</template>
			<template #form_input_Dimension="{item}">
				<DimensionEditorVertical
					v-model="item.Dimension"
					:readOnly="['SUBMITTED', 'READY', 'POSTED'].includes(item.Status)"
					:default-list="profile.Dimension"
					@change="onDimensionChange"
				/>
			</template>
			<template #form_input_CustomerID="{item}">
				<SelesOrderCustomer
					v-model="item.CustomerID"
					:customer="item.Customer"
					:quotation="data.quotation"
					@change="onChangeTaxCodesFromCustomer"
					:form-mode="data.formMode"
				/>
			</template>
			<template #form_input_PostingProfileID="{item, config}">
				<label class="input_label">{{ config.label }}</label>
				<div>
					{{ item.PostingProfileID === "" ? "-" : item.PostingProfileID }}
				</div>
				<!-- <s-input
          ref="refInput"
          label="Posting Profile"
          v-model="item.PostingProfileID"
          class="w-full"
          :required="true"
          :disabled="['SUBMITTED', 'READY', 'POSTED'].includes(item.Status)"
          :keepErrorSection="true"
          use-list
          :lookup-url="`/fico/postingprofile/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          read-only
        ></s-input> -->
			</template>
			<template #form_input_JournalTypeID="{item}">
				<s-input
					ref="refInput"
					label="Journal Type"
					v-model="item.JournalTypeID"
					class="w-full"
					:required="true"
					:disabled="['SUBMITTED', 'READY', 'POSTED'].includes(item.Status)"
					:keepErrorSection="true"
					use-list
					:lookup-url="`/sdp/salesorderjournaltype/find`"
					lookup-key="_id"
					:lookup-labels="['Name']"
					:lookup-searchs="['_id', 'Name']"
					@change="
						(field, v1, v2, old, ctlRef) => {
							getJournalType(v1, 'change', item);
						}
					"
				></s-input>
			</template>

			<!-- slot form tab -->
			<template #form_tab_Line="{item}">
				<SalesOrderLine
					ref="lineConfig"
					v-model="item.Lines"
					grid-config="/sdp/salesorder/line/gridconfig"
					form-config="/sdp/salesorder/line/formconfig"
					:opportunity="data.quotation"
					:tax="item.TaxCodes"
					:salesPriceBook="item.SalesPriceBook"
					@recalc="ReCalc"
					:journalTypeID="item.JournalTypeID"
					:trxType="data.record.TransactionType"
				></SalesOrderLine>
			</template>

			<template #form_tab_Checklist="{item, mode}">
				<Checklist
					v-model="item.Checklists"
					:checklist-id="data.jurnalType.ChecklistTemplateID"
					:readOnly="readOnly || mode == 'view'"
				/>
			</template>

			<template #form_tab_Editor="{item}">
				<SalesOrderEditor
					:model-value="item"
					@update:model-value="updateEditor"
					form-config="/sdp/salesorder/editor/formconfig"
					:form-default-mode="data.formMode"
				>
				</SalesOrderEditor>
			</template>

			<template #form_tab_Manpower_Fullfillment="{item}">
				<SalesOrderManPower
					ref="manpowerConfig"
					v-model="item.ManPower"
					grid-config="/sdp/salesorder/manpower/gridconfig"
					form-config="/sdp/salesorder/manpower/formconfig"
				>
				</SalesOrderManPower>
			</template>

			<template #form_tab_Breakdown_Cost="{item}">
				<SalesOrderBreakdownCost
					ref="breakdowncostConfig"
					v-model="item.BreakdownCost"
					grid-config="/sdp/salesorder/breakdowncost/gridconfig"
					form-config="/sdp/salesorder/breakdowncost/formconfig"
				>
				</SalesOrderBreakdownCost>
			</template>

			<template #form_input_TaxCodes="{item, config}">
				<s-input
					:field="config.field"
					:kind="config.kind"
					:label="config.label"
					@change="
						(name, v1, v2, old) => onFormFieldChanged(name, v1, v2, old, item)
					"
					:disabled="config.readOnly || mode == 'view'"
					:caption="config.caption"
					:hint="config.hint"
					:multi-row="config.multiRow"
					:use-list="config.useList"
					:items="config.items"
					:rules="config.rules"
					:required="config.required"
					:read-only="config.readOnly"
					:lookup-url="config.lookupUrl"
					:lookup-key="config.lookupKey"
					:allow-add="config.allowAdd"
					:lookup-format1="config.lookupFormat1"
					:lookup-format2="config.lookupFormat2"
					:lookup-payload-builder="
						(search) =>
							lookupTaxCodesPayloadBuilder(search, config, item[config.field])
					"
					:decimal="config.decimal"
					:date-format="config.dateFormat"
					:multiple="config.multiple"
					:keep-label="keepLabel"
					:lookup-labels="config.lookupLabels"
					:lookup-searchs="
						config.lookupSearchs && config.lookupSearchs.length == 0
							? config.lookupLabels
							: config.lookupSearchs
					"
					v-model="item[config.field]"
					:class="{checkboxOffset: config.kind == 'checkbox'}"
					ref="inputs"
				>
				</s-input>
			</template>

			<template #form_input_HeaderDiscountValue="{item, config}">
				<s-input
					:field="config.field"
					kind="text"
					:label="
						config.kind == 'checkbox' || config.kind == 'bool'
							? ''
							: config.label
					"
					:disabled="config.readOnly"
					:caption="config.caption"
					:hint="config.hint"
					:multi-row="config.multiRow"
					:use-list="config.useList"
					:items="config.items"
					:rules="config.rules"
					:required="config.required"
					:read-only="config.readOnly"
					:lookup-url="config.lookupUrl"
					:lookup-key="config.lookupKey"
					:allow-add="config.allowAdd"
					:lookup-format1="config.lookupFormat1"
					:lookup-format2="config.lookupFormat2"
					:decimal="config.decimal"
					:date-format="config.dateFormat"
					:multiple="config.multiple"
					:lookup-labels="config.lookupLabels"
					:lookup-searchs="
						config.lookupSearchs && config.lookupSearchs.length == 0
							? config.lookupLabels
							: config.lookupSearchs
					"
					@focus="rowFieldFocus"
					@change="rowFieldChanged"
					:model-value="valueDiscountPrice.get(item[config.field])"
					@update:model-value="(val) => valueDiscountPrice.set(val, item)"
					ref="inputs"
				/>

				<!-- <input
					type="number"
					:placeholder="config.caption || config.label"
					class="input_field text-right w-[100px]"
					:value="valueDiscountPrice.get(item['HeaderDiscountValue'])"
					@change="(val) => valueDiscountPrice.set(val, item)"
					ref="control"
					:disabled="configdisabled"
				/> -->
			</template>

			<template #form_tab_Attachment="{item}">
				<s-grid-attachment
					:journal-id="item._id"
					journal-type="Sales Order"
					ref="gridAttachment"
					v-model="item.Attachment"
				></s-grid-attachment>
			</template>

			<!-- slot form btn -->
			<template #form_buttons_1="{item, config}">
				<s-button
					class="bg-transparent hover:bg-blue-500 hover:text-black"
					label="Preview"
					icon="eye-outline"
					@click="saveForm"
				></s-button>
				<s-button
					class="bg-transparent hover:bg-blue-500 hover:text-black"
					label="Action"
					icon="eye-outline"
					@click="openModalQuotation"
				></s-button>
				<!-- <FormButtonsTrx
					:posting-profile-id="item.PostingProfileID"
					:status="item.Status"
					:journalId="item._id"
					journal-type-id="Sales Order"
					moduleid="sdp/new"
          			@pre-submit="preSubmit"
					@post-submit="postSubmit(item)"
				>
				</FormButtonsTrx> -->
				<form-buttons-trx
					:status="item.Status"
					:journal-id="item._id"
					:posting-profile-id="item.PostingProfileID"
					journal-type-id="Sales Order"
					moduleid="sdp/new"
					@preSubmit="trxPreSubmit"
					@postSubmit="trxPostSubmit"
					@errorSubmit="trxErrorSubmit"
				/>
			</template>

			<!-- slot grid btn -->
			<template #grid_item_buttons_1="{item, config}">
				<log-trx :id="item._id" />
				<a href="#" @click="duplicateSO(item)" class="mark_action">
					<mdicon
						name="content-copy"
						width="15"
						alt="edit"
						class="cursor-pointer hover:text-primary"
					/>
				</a>
			</template>
		</data-list>

		<PreviewReport
			v-if="data.appMode == 'preview'"
			v-model="data.record"
			grid-read="/sdp/salesorder/gets"
			grid-config="/sdp/salesorder/line/preview/gridconfig"
			@cancel-click="CancelPreview"
		>
			<template #buttons_1="props">
				<div class="flex gap-[1px] mr-2">
          <form-buttons-trx
						:status="props.item.Status"
						:journal-id="props.item._id"
						:posting-profile-id="props.item.PostingProfileID"
						journal-type-id="Sales Order"
						moduleid="sdp/new"
						@preSubmit="trxPreSubmit"
						@postSubmit="trxPostSubmit"
						@errorSubmit="trxErrorSubmit"
          />
				</div>
			</template>	
		</PreviewReport>
	</div>
</template>

<script setup>
import {reactive, ref, watch, computed, onMounted, inject, nextTick} from "vue";
import {layoutStore} from "../../stores/layout.js";
import {DataList, util, SInput, SModal, SButton} from "suimjs";
import {useRoute, useRouter} from "vue-router";
import {authStore} from "@/stores/auth.js";
import Checklist from "@/components/common/Checklist.vue";

import SalesOrderLine from "./widget/SalesOrderLine.vue";
import SalesOrderEditor from "./widget/SalesOrderEditor.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import PreviewReport from "./widget/PreviewSalesOrder.vue";
import SalesOrderBreakdownCost from "./widget/SalesOrderBreakdownCost.vue";
import SalesOrderManPower from "./widget/SalesOrderManPower.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import SelesOrderCustomer from "./widget/SelesOrderCustomer.vue";
import Loader from "@/components/common/Loader.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import References from "@/components/common/References.vue";
import moment from "moment";
layoutStore().name = "tenant";

const featureID = "SalesOrder";
// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})
const profile = authStore().getRBAC(featureID);

const lineConfig = ref(null);
const axios = inject("axios");

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
let customFilter = computed(() => {
	const filters = [];
	if (data.search.Customer !== null && data.search.Customer.length > 0) {
		filters.push({
			Field: "CustomerID",
			Op: "$contains",
			Value: data.search.Customer,
		});
	}
	if (data.search.Name !== null && data.search.Name !== "") {
		filters.push({
			Field: "SpkNo",
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
			Field: "SalesOrderDate",
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
			Field: "SalesOrderDate",
			Op: "$lte",
			Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
		});
	}
	if (data.search.Status !== null && data.search.Status !== "") {
		filters.push({
			Field: "Status",
			Op: "$eq",
			Value: data.search.Status,
		});
	}
	if (filters.length == 1) return filters[0];
	else if (filters.length > 1) return {Op: "$and", Items: filters};
	else return null;
});
const listControl = ref(DataList);
const gridAttachment = ref(SGridAttachment);
const generateModal = ref(SModal);

const route = useRoute();
const router = useRouter();

const data = reactive({
	title: null,
	appMode: "grid",
	formMode: "edit",
	titleForm: "Sales Order",
	allowDelete: route.query.allowdelete === "true",
	formAssets: {},
	preview: {},
	record: {
		Dimension: [],
		Lines: [],
		BreakdownCost: [],
		SalesOrderDate: new Date(),
		Status: "",
		PostingProfileID: "",
		Customer: {
			PersonalContact: "",
			Address: "",
			AddressDelivery: "",
			City: "",
			CityDelivery: "",
			Province: "",
			ProvinceDelivery: "",
			Country: "",
			CountryDelivery: "",
			Zipcode: "",
			ZipcodeDelivery: "",
			CustomerName: "",
		},
		References: [],
		CustomerID: "",
	},
	jurnalType: {},
	search: {
		Customer: [],
		Name: "",
		DateFrom: null,
		DateTo: null,
		Status: "",
	},
	cachequotation: undefined,
	quotation: undefined,
	isSelected: false,
	formInitialTab: 0,
	stayOnForm: false,
	warehouseId: "",
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

const valueDiscountPrice = {
	get(v) {
		return util.formatMoney(v, {decimal: 0});
	},
	set(v, record) {
		const v1 = Number(String(v).replaceAll(",", ""));
		record["HeaderDiscountValue"] = v1;
		onFormFieldChanged("HeaderDiscountValue", v1, 0, 0, record);
	},
};

function CancelPreview() {
	data.appMode = "grid";
	listControl.value.setControlMode("grid");
}

function closeModalQuotation() {
	generateModal.value.hide();
}
function openModalQuotation() {
	generateModal.value.show();
}

async function CreateWithQuotation() {
	try {
		data.quotation = data.cachequotation;
		await axios.post(`/sdp/salesorder/action-so`, {_id: data.quotation});
		listControl.value.setControlMode("grid");
		listControl.value.refreshList();
		listControl.value.refreshForm();
		closeModalQuotation();
	} catch (error) {
		util.showError(error);
	}
}

function openForm(record) {
	util.nextTickN(2, () => {
		data.record = record;
		listControl.value.setFormFieldAttr("_id", "rules", roleID);
		listControl.value.setFormRecord(record);
	});
}

function newRecord() {
	data.formMode = "new";
	data.titleForm = "Create New Sales Order";
	data.quotation = undefined;
	const record = {
		PostingProfileID: "",
		Status: "",
		HeaderDiscountValue: 0.0,
		TransactionType: "Asset",
	};

	openForm(record);
}

async function editRecord(record) {
	data.formMode = "edit";
	data.quotation = undefined;
	if (record.Lines.length > 0) {
		record.Lines.map((e) => {
			e.Uom = e.UoM;
			return e;
		});
	}
	if (record.CustomerID) {
		const promisebatch = [];
		const flagpromise = {};
		promisebatch.push(getDataCustomer(record.CustomerID));

		const res = await Promise.all(promisebatch);
		const customer = res[0];
		data.record.Customer = {
			PersonalContact: customer.Detail.PersonalContact,
			Address: customer.Detail.Address,
			AddressDelivery: customer.Detail.DeliveryAddress,
			City: customer.Detail.City,
			CityDelivery: customer.Detail.DeliveryCity,
			Province: customer.Detail.Province,
			ProvinceDelivery: customer.Detail.DeliveryProvince,
			Country: customer.Detail.Country,
			CountryDelivery: customer.Detail.DeliveryCountry,
			Zipcode: customer.Detail.Zipcode,
			ZipcodeDelivery: customer.Detail.DeliveryZipcode,
			CustomerName: customer.Name,
		};
	}
	if (record.SalesQuotationRefNo) {
		data.cachequotation = record.SalesQuotationRefNo;
		data.quotation = record.SalesQuotationRefNo;
	}
	if (record.JournalTypeID) {
		getJournalType(record.JournalTypeID, "init", {});
	}

	data.titleForm = `Edit Sales Order - ${record.SalesOrderNo} | ${record.Name}`;

	if (record.FooterAsset == "") {
		record.FooterAsset = undefined;
	}

	if (record.LetterHeadAsset == "") {
		record.LetterHeadAsset = undefined;
	}

	openForm(record);
	nextTick(() => {
		if (["SUBMITTED", "READY", "POSTED"].includes(record.Status)) {
			// setFormMode("view");
			data.formMode = "view";
		}
	});
}

function onPreSave(record) {
	record.WarehouseID = data.warehouseId;
}

function onPostSave(record) {
	gridAttachment.value.Save();
}

function onCancelForm(mode) {
	if (mode === "grid") {
		data.titleForm = "Sales Order";
	}
}

const onFormFieldChanged = async (name, v1, v2, old, record) => {
	if (name == "Customer") {
		const promisebatch = [];
		const flagpromise = {};
		promisebatch.push(getDataCustomer(v1));

		const res = await Promise.all(promisebatch);
		const customer = res[0];
		// getDataCustomer(v1).then((customer) => {
		record.Customer = {
			PersonalContact: customer.Detail.PersonalContact,
			Address: customer.Detail.Address,
			AddressDelivery: customer.Detail.DeliveryAddress,
			City: customer.Detail.City,
			CityDelivery: customer.Detail.DeliveryCity,
			Province: customer.Detail.Province,
			ProvinceDelivery: customer.Detail.DeliveryProvince,
			Country: customer.Detail.Country,
			CountryDelivery: customer.Detail.DeliveryCountry,
			Zipcode: customer.Detail.Zipcode,
			ZipcodeDelivery: customer.Detail.DeliveryZipcode,
			CustomerName: customer.Name,
		};
		// });
	}

	if (name == "HeaderDiscountType") {
		if (v1 == "fixed") {
			record.HeaderDiscountAmount = record.HeaderDiscountValue;
		} else {
			record.HeaderDiscountAmount =
				record.SubTotalAmount * (record.HeaderDiscountValue / 100);
		}

		record.TotalAmount =
			record.SubTotalAmount +
			record.TaxAmount -
			(record.DiscountAmount + record.HeaderDiscountAmount);
	}

	if (name == "HeaderDiscountValue") {
		if (record.HeaderDiscountType == "fixed") {
			record.HeaderDiscountAmount = v1;
		} else {
			record.HeaderDiscountAmount = record.SubTotalAmount * (v1 / 100);
		}

		record.TotalAmount =
			record.SubTotalAmount +
			record.TaxAmount -
			(record.DiscountAmount + record.HeaderDiscountAmount);
	}

	if (name == "TaxCodes") {
		const remove = (old || []).filter((_old) => {
			const same = v1.find((_v1) => _v1 === _old);
			if (!same) {
				return true;
			}

			return false;
		});

		const add = v1.filter((_old) => {
			const same = (old || []).find((_v1) => _v1 === _old);
			if (!same) {
				return true;
			}

			return false;
		});

		if (remove.length > 0) {
			remove.map(async (nv) => {
				const Taxes = await GetTaxes(nv);
				const record = listControl.value.getFormRecord();

				if ((record.Lines || []).length > 0) {
					record.Lines.map((line) => {
						if (!line.TaxCodes) {
							line.TaxCodes = [];
						}

						line.TaxCodes = line.TaxCodes.filter((txcode) => txcode !== nv);

						if (line.TaxCodes && line.TaxCodes.length > 0) {
							line.Taxable = true;
						} else {
							line.Taxable = false;
						}

						if (Taxes.InvoiceOperation == "Decrease") {
							record.TaxAmount += line.Amount * Taxes.Rate;
						} else if (Taxes.InvoiceOperation == "Increase") {
							record.TaxAmount -= line.Amount * Taxes.Rate;
						}

						return line;
					});
				}

				const headerdiscount = isNaN(record.HeaderDiscountAmount)
					? 0
					: record.HeaderDiscountAmount;

				record.TotalAmount =
					record.SubTotalAmount -
					record.TaxAmount -
					(record.DiscountAmount + headerdiscount);
				listControl.value.setFormRecord(record);
			});
		}

		if (add.length > 0) {
			add.map(async (nv) => {
				const Taxes = await GetTaxes(nv);
				const record = listControl.value.getFormRecord();
				if ((record.Lines || []).length > 0) {
					record.Lines.map((line) => {
						if (!line.TaxCodes) {
							line.TaxCodes = [];
						}

						const available = line.TaxCodes.find((txcode) => txcode === nv);
						if (!available) {
							line.TaxCodes.push(nv);
						}

						if (line.TaxCodes && line.TaxCodes.length > 0) {
							line.Taxable = true;
						} else {
							line.Taxable = false;
						}

						if (Taxes.InvoiceOperation == "Decrease") {
							record.TaxAmount -= line.Amount * Taxes.Rate;
						} else if (Taxes.InvoiceOperation == "Increase") {
							record.TaxAmount += line.Amount * Taxes.Rate;
						}

						return line;
					});
				}

				const headerdiscount = isNaN(record.HeaderDiscountAmount)
					? 0
					: record.HeaderDiscountAmount;

				record.TotalAmount =
					record.SubTotalAmount +
					record.TaxAmount -
					(record.DiscountAmount + headerdiscount);
				listControl.value.setFormRecord(record);
			});
		}
	}

	if (name == "TransactionType") {
		record.TransactionType = v1;
	}

	listControl.value.setGridRecord(
		record,
		listControl.value.getGridCurrentIndex()
	);
};

const onDimensionChange = (field, v1, v2, old, ctl) => {
	if (field == "Site") {
		getWarehouseBySiteId(v1);
	}
}

function getWarehouseBySiteId(siteId) {
	axios.post(
			`/tenant/warehouse/find`,
			{
				Sort: ["Name"],
				Select: ["_id", "Name"],
				Take: 20,
				Where: {
					Op: "$and",
					Items: [
						{
							Op: "$eq",
							Field: "Dimension.Key",
							Value: "Site",
						},
						{
							Op: "$eq",
							Field: "Dimension.Value",
							Value: siteId,
						}
					],
				}
			},
		)
		.then((r) => {
			if (r.data) {
				data.warehouseId = r.data[0]._id;
			}
		});
}

const getDataCustomer = async (idCustomer) => {
	try {
		const dataresponse = await axios.post(`/bagong/customer/get`, [idCustomer]);
		return dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
};

function saveForm() {
	const record = data.record;
	const URLdata =
		data.formMode == "edit" ||
		["SUBMITTED", "READY", "POSTED"].includes(data.record.Status)
			? "/sdp/salesorder/update"
			: "/sdp/salesorder/insert";

	axios.post(URLdata, record).then((r) => {
		// console.log(r);
		gridAttachment.value.Save(r._id, "Sales Order");
		data.appMode = "preview";
		data.preview = record;
	});
}

function trxPreSubmit(status, action, doSubmit) {
	if (status == "DRAFT") {
		trxSubmit(doSubmit);
	}
}

function trxSubmit(doSubmit) {
	util.nextTickN(2, () => {
		const valid = listControl.value.formValidate();
		if (valid) {
			listControl.value.submitForm(
				data.record,
				() => {
					doSubmit();
				},
				() => {
					setLoadingForm(false);
				}
			);
		}
	});
}
function closePreview() {
  data.appMode = "grid";
}
function trxPostSubmit(record) {
	setLoadingForm(false);
	closePreview()
  setModeGrid();
}
function trxErrorSubmit(e) {
	setLoadingForm(false);
}

function setModeGrid() {
	listControl.value.setControlMode("grid");
	listControl.value.refreshList();
}

async function ReCalc(calcs) {
	const calc = await calcs;
	const records = listControl.value.getFormRecord();
	util.nextTickN(2, () => {
		const headerdiscount = isNaN(records.HeaderDiscountAmount)
			? 0
			: records.HeaderDiscountAmount;

		calc.TotalAmount =
			calc.SubTotalAmount +
			calc.TaxAmount -
			(calc.DiscountAmount + headerdiscount);
		listControl.value.setFormRecord({...records, ...calc});
	});
}

function getJournalType(_id, action, item) {
	axios
		.post("/sdp/salesorderjournaltype/find?_id=" + _id, {sort: ["-_id"]})
		.then(
			(r) => {
				const dt = r.data[0];
				data.jurnalType = dt;

				// console.log(dt);
				if (["", "DRAFT"].includes(data.record.Status)) {
					if (action === "change" && data.formMode === "new") {
						item.PostingProfileID = dt.PostingProfileID;
						data.record.PostingProfileID = dt.PostingProfileID;
					} else if (action === "change" && data.formMode === "edit") {
						data.record.PostingProfileID = dt.PostingProfileID;
					}
				}
			},
			(e) => util.showError(e)
		);
}

async function duplicateSO(item) {
	try {
		const dataresponse = await axios.post(`/sdp/salesorder/duplicate-so`, {
			_id: item._id,
		});
		setModeGrid();
		// return dataresponse.data[0];
	} catch (error) {
		util.showError(error);
	}
}

function refreshData() {
	util.nextTickN(2, () => {
		listControl.value.refreshGrid();
	});
}

function onAlterGridConfig(cfg) {
	cfg.setting.idField = "Created";
	cfg.setting.sortable = ["_id", "Created", "SalesOrderDate"];
}
function onAlterFormConfig(cfg) {
  if (route.query.id !== undefined) {
    let currQuery = {...route.query};
    listControl.value.selectData({ _id: currQuery.id });
    delete currQuery['id'];
    router.replace({ path: route.path, query: currQuery });
  }
}
async function onChangeTaxCodesFromCustomer(customerID) {
	const record = listControl.value.getFormRecord();
	// console.log(record.CustomerID, record.Dimension, customerID)
	if (customerID !== null) {
		const promisebatch = [];
		promisebatch.push(getDataCustomer(customerID));
		const res = await Promise.all(promisebatch);
		const customer = res[0];

		const TaxCodes = [];
		if (customer) {
			if (customer.Detail.Tax1 !== "") {
				TaxCodes.push(customer.Detail.Tax1);
			}
			if (customer.Detail.Tax2 !== "") {
				TaxCodes.push(customer.Detail.Tax2);
			}
		}

		record.TaxCodes = TaxCodes;
		// console.log(record)

		listControl.value.setGridRecord(
			record,
			listControl.value.getGridCurrentIndex()
		);
	}
}

function updateEditor(newValue) {
	listControl.value.setFormRecord(newValue);
}

function setLoadingForm(loading) {
	listControl.value.setFormLoading(loading);
}

function lookupTaxCodesPayloadBuilder(search, config, value) {
	const qp = {};
	if (search != "") data.filterTxt = search;
	qp.Take = 20;
	qp.Sort = [config.lookupLabels[0]];
	qp.Select = config.lookupLabels;
	let idInSelect = false;
	const selectedFields = config.lookupLabels.map((x) => {
		if (x == config.lookupKey) {
			idInSelect = true;
		}
		return x;
	});
	if (!idInSelect) {
		selectedFields.push(config.lookupKey);
	}
	qp.Select = selectedFields;

	//setting search
	if (search.length > 0 && config.lookupSearchs.length > 0) {
		if (config.lookupSearchs.length == 1)
			qp.Where = {
				Field: config.lookupSearchs[0],
				Op: "$contains",
				Value: [search],
			};
		else
			qp.Where = {
				Op: "$or",
				items: config.lookupSearchs.map((el) => {
					return {Field: el, Op: "$contains", Value: [search]};
				}),
			};
	}

	if (config.multiple && value && value.length > 0 && qp.Where != undefined) {
		const whereExisting =
			value.length == 1
				? {Op: "$eq", Field: config.lookupKey, Value: value[0]}
				: {
						Op: "$or",
						items: value.map((el) => {
							return {Field: config.lookupKey, Op: "$eq", Value: el};
						}),
				  };

		qp.Where = {Op: "$or", items: [qp.Where, whereExisting]};
	}

	// if (qp.Where != undefined) {
	// 	const items = [{Op: "$eq", Field: `IsActive`, Value: true}];
	// 	items.push(qp.Where);
	// 	qp.Where = {
	// 		Op: "$and",
	// 		items: items,
	// 	};
	// } else {
	// 	qp.Where = {Op: "$eq", Field: `IsActive`, Value: true};
	// }

	if (qp.Where != undefined) {
		const items = [
			{Op: "$eq", Field: `IsActive`, Value: true},
			{Op: "$eq", Field: `Modules`, Value: "Sales"},
		];
		items.push(qp.Where);
		qp.Where = {
			Op: "$and",
			items: items,
		};
	} else {
		qp.Where = {
			Op: "$and",
			items: [
				{Op: "$eq", Field: `IsActive`, Value: true},
				{Op: "$eq", Field: `Modules`, Value: "Sales"},
			],
		};
	}

	return qp;
}

async function GetTaxes(_id) {
	return (await axios.post(`/fico/taxcode/get`, [_id])).data;
}
</script>
