<template>
	<div class="w-full">
		<s-modal
			title="Sales Opportunity"
			class="p-4"
			:display="false"
			ref="generateModal"
			hideButtons
			@submit="closeModalOpportunity"
		>
			<s-input
				ref="siteID"
				useList
				class="mb-4 min-w-[240px]"
				label="Opportunity"
				:disabled="data.formMode === 'edit'"
				lookup-url="/sdp/salesopportunity/find"
				lookup-key="_id"
				:lookup-labels="['Name']"
				:lookup-searchs="['_id', 'Name']"
				v-model="data.cacheopportunity"
				placeholder="Opportunity"
			></s-input>
			<s-button
				v-if="data.formMode === 'new'"
				class="w-full btn_success text-center"
				label="Create"
				@click="CreatewithOpportunity"
			></s-button>
		</s-modal>
		<data-list
			v-show="data.appMode == 'grid'"
			class="card"
			ref="listControl"
			:title="data.titleForm"
			grid-config="/sdp/salesquotation/gridconfig"
			grid-read="/sdp/salesquotation/gets"
			form-read="/sdp/salesquotation/get"
			grid-mode="grid"
			form-config="/sdp/salesquotation/formconfig"
			grid-delete="/sdp/salesquotation/delete"
			form-keep-label
			form-insert="/sdp/salesquotation/insert"
			form-update="/sdp/salesquotation/update"
			:form-tabs-new="['General', 'Line', 'Editor']"
			:form-tabs-edit="['General', 'Line', 'Editor']"
			:form-tabs-view="['General', 'Line', 'Editor']"
			:form-fields="[
				'Dimension',
				'Customer',
				'Address',
				'AddressDelivery',
				'PostingProfileID',
				'JournalType',
				'TaxCodes',
				'HeaderDiscountValue',
			]"
			grid-sort-field="Created"
			grid-sort-direction="desc"
			:grid-fields="['Customer', 'Status', 'OpportunityNo']"
			:grid-custom-filter="customFilter"
			:formInitialTab="data.formInitialTab"
			:init-app-mode="data.appMode"
			:init-form-mode="data.formMode"
			@formNewData="newRecord"
			@formEditData="editRecord"
			@pre-save="onPreSave"
			@controlModeChanged="onCancelForm"
			@form-field-change="onFormFieldChanged"
			@alterGridConfig="onAlterGridConfig"
      @alterFormConfig="onAlterFormConfig"
			:formHideSubmit="['', 'DRAFT'].indexOf(data.record.Status) === -1"
			:grid-hide-new="!profile.canCreate"
			:grid-hide-edit="!profile.canUpdate"
			:grid-hide-delete="!profile.canDelete"
			stay-on-form-after-save
		>
			<!--  @grid-refreshed="onGridRefreshed" -->
			<template #grid_Customer="{item}">
				<s-input
					ref="refCustomer"
					v-model="item.Customer"
					class="w-50"
					read-only
					use-list
					:lookup-url="`/tenant/customer/find`"
					lookup-key="_id"
					:lookup-labels="['Name']"
					:lookup-searchs="['_id', 'Name']"
				></s-input>
				<!-- <div
          v-for="hdr in data.CustomersName.filter(
            (dt) => dt._id === item.Customer
          )"
        >
          {{ hdr.Name }}
        </div> -->
			</template>

			<template #grid_OpportunityNo="{item}">
				<s-input
					v-model="item.OpportunityNo"
					class="w-50"
					read-only
					use-list
					:lookup-url="`/sdp/salesopportunity/find`"
					lookup-key="_id"
					:lookup-labels="['OpportunityNo']"
					:lookup-searchs="['_id', 'Name']"
				></s-input>
			</template>

			<template #grid_Status="{item}">
				<status-text :txt="item.Status" />
			</template>

			<template #form_input_Customer="{item, config}">
				<s-input
					v-if="data.opportunity"
					:field="config.field"
					:kind="config.kind"
					:label="config.label"
					:disabled="data.opportunity"
					:caption="config.caption"
					:hint="config.hint"
					:multi-row="config.multiRow"
					:use-list="config.useList"
					:items="config.items"
					:rules="config.rules"
					:required="config.required"
					:read-only="false"
					:lookup-url="config.lookupUrl"
					:lookup-key="config.lookupKey"
					:allow-add="config.allowAdd"
					:lookup-format1="config.lookupFormat1"
					:lookup-format2="config.lookupFormat2"
					:lookup-payload-builder="
						(search) =>
							lookupCustomerPayloadBuilder(search, config, item[config.field])
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
				/>
			</template>

			<template #form_input_Dimension="{item}">
				<DimensionEditorVertical
					v-model="item.Dimension"
					:readOnly="['SUBMITTED', 'READY', 'POSTED'].includes(item.Status)"
					:default-list="profile.Dimension"
				/>
			</template>

			<template #form_input_Address="{item, config}">
				<div class="bg-transparent max-w-xs">
					{{
						item["Address"] && item["Address"] != ""
							? item["Address"]
							: "&nbsp;"
					}}
				</div>
			</template>

			<template #form_input_AddressDelivery="{item, config}">
				<div class="bg-transparent max-w-xs">
					{{
						item["AddressDelivery"] && item["AddressDelivery"] != ""
							? item["AddressDelivery"]
							: "&nbsp;"
					}}
				</div>
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

			<template #form_input_JournalType="{item}">
				<s-input
					ref="refInput"
					label="Journal Type"
					v-model="item.JournalType"
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

			<template #form_tab_Line="{item}">
				<SalesQuotationLine
					ref="lineConfig"
					v-model="item.Lines"
					grid-config="/sdp/salesquotation/line/gridconfig"
					form-config="/sdp/salesquotation/line/formconfig"
					:opportunity="data.opportunity"
					@recalc="ReCalc"
					:sales-price-book="item.SalesPriceBook"
					:tax-codes="item.TaxCodes"
					:trx-type="data.record.TransactionType"
				></SalesQuotationLine>
			</template>

			<template #form_tab_Editor="{item}">
				<SalesQuotationEditor
					:model-value="item"
					@update:model-value="updateEditor"
					form-config="/sdp/salesquotation/editor/formconfig"
					:form-default-mode="data.formMode"
				>
				</SalesQuotationEditor>
			</template>

			<template #form_buttons_1="{item, config}">
				<s-button
					class="bg-transparent hover:bg-blue-500 hover:text-black"
					label="Preview"
					icon="eye-outline"
					@click="saveForm"
				></s-button>
				<!-- <s-button
					class="bg-transparent hover:bg-blue-500 hover:text-black"
					label="Submit"
					icon="check"
				></s-button> -->
				<s-button
					class="bg-transparent hover:bg-blue-500 hover:text-black"
					label="Action"
					icon="eye-outline"
					@click="openModalOpportunity"
				></s-button>

				<form-buttons-trx
					:status="item.Status"
					:journal-id="item._id"
					:posting-profile-id="item.PostingProfileID"
					journal-type-id="Sales Quotation"
					moduleid="sdp/new"
					@preSubmit="trxPreSubmit"
					@postSubmit="trxPostSubmit"
				/>
			</template>

			<template #grid_item_buttons_1="{item, config}">
				<log-trx :id="item._id" />
				<!-- <a href="#" @click="sentMail(item)" class="mark_action">
          <mdicon
            name="email-fast-outline"
            width="24"
            alt="edit"
            class="cursor-pointer hover:text-primary"
          />
        </a> -->
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

			<template #grid_header_search="{config}">
				<s-input
					ref="refCustomer"
					v-model="data.searchData.Customer"
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
					kind="text"
					label="Name"
					class="w-full"
					v-model="data.searchData.Text"
					@keyup.enter="refreshData"
				></s-input>
				<s-input
					kind="date"
					label="Date From"
					v-model="data.searchData.DateFrom"
					@change="refreshData"
				></s-input>
				<s-input
					kind="date"
					label="Date To"
					v-model="data.searchData.DateTo"
					@change="refreshData"
				></s-input>
				<s-input
					kind="text"
					label="Status"
					class="w-full"
					v-model="data.searchData.Status"
					use-list
					:items="['DRAFT', 'SUBMITTED', 'SENT', 'READY', 'POSTED']"
					@change="refreshData"
				></s-input>
			</template>
		</data-list>

		<PreviewReport
			v-if="data.appMode == 'preview'"
			v-model="data.record"
			grid-read="/sdp/salesquotation/gets"
			grid-config="/sdp/salesquotation/line/preview/gridconfig"
			@cancel-click="CancelPreview"
		>
			<template #buttons_1="props">
				<div class="flex gap-[1px] mr-2">
          <form-buttons-trx
						:status="props.item.Status"
						:journal-id="props.item._id"
						:posting-profile-id="props.item.PostingProfileID"
						journal-type-id="Sales Quotation"
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
import {
	reactive,
	ref,
	watch,
	computed,
	onMounted,
	inject,
	onBeforeMount,
} from "vue";
import {layoutStore} from "../../stores/layout.js";
// import { DataList, util } from "suimjs";
import {DataList, util, SInput, SModal, SButton} from "suimjs";
import {useRoute, useRouter} from 'vue-router'
import {authStore} from "@/stores/auth.js";
import moment from "moment";
import SalesQuotationLine from "./widget/SalesQuotationLine.vue";
import SalesQuotationEditor from "./widget/SalesQuotationEditor.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import PreviewReport from "./widget/PreviewSalesQuotation.vue";
import SInputBuilder from "@/components/common/SInputBuilder.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import {Axios} from "axios";

layoutStore().name = "tenant";

const featureID = "SalesQuotation";
// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})
const profile = authStore().getRBAC(featureID);

const lineConfig = ref(null);
const axios = inject("axios");
let customFilter = computed(() => {
	const filters = [];
	if (
		data.searchData.Customer !== null &&
		data.searchData.Customer.length > 0
	) {
		filters.push({
			Field: "Customer",
			Op: "$contains",
			Value: data.searchData.Customer,
		});
	}
	if (data.searchData.Text !== null && data.searchData.Text !== "") {
		filters.push({
			Field: "QuotationName",
			Op: "$contains",
			Value: [data.searchData.Text],
		});
	}
	if (
		data.searchData.DateFrom !== null &&
		data.searchData.DateFrom !== "" &&
		data.searchData.DateFrom !== "Invalid date"
	) {
		filters.push({
			Field: "QuotationDate",
			Op: "$gte",
			Value: moment(data.searchData.DateFrom)
				.utc()
				.format("YYYY-MM-DDT00:mm:00Z"),
		});
	}
	if (
		data.searchData.DateTo !== null &&
		data.searchData.DateTo !== "" &&
		data.searchData.DateTo !== "Invalid date"
	) {
		filters.push({
			Field: "QuotationDate",
			Op: "$lte",
			Value: moment(data.searchData.DateTo)
				.utc()
				.format("YYYY-MM-DDT23:59:00Z"),
		});
	}
	if (data.searchData.Status !== null && data.searchData.Status !== "") {
		filters.push({
			Field: "Status",
			Op: "$eq",
			Value: data.searchData.Status,
		});
	}
	if (filters.length == 1) return filters[0];
	else if (filters.length > 1) return {Op: "$and", Items: filters};
	else return null;
});
const listControl = ref(DataList);
const generateModal = ref(SModal);

const route = useRoute();
const router = useRouter();

const data = reactive({
	title: null,
	appMode: "grid",
	formMode: "edit",
	titleForm: "Sales Quotation",
	allowDelete: route.query.allowdelete === "true",
	formAssets: {},
	preview: {},
	record: {
		Dimension: [],
		Lines: [],
		Status: "",
	},

	// Search Data
	CustomFiltering: undefined,
	searchData: {
		Customer: [],
		Text: "",
		QuotationDate: null,
		DateFrom: null,
		DateTo: null,
		Status: "",
	},

	// Get data Opportunity
	cacheopportunity: undefined,
	opportunity: undefined,

	// Get Data grid Customer
	CustomersName: [],

	isSelected: false,
	formInitialTab: 0,
});

const debounce = createDebounce();

watch(
	() => route.query.objname,
	() => {
		util.nextTickN(2, () => {
			listControl.value.refreshList();
			listControl.value.refreshForm();
		});
	}
);

// watch(
//   () => JSON.stringify(data.searchData),
//   async (v) => {
//     const Filter = JSON.parse(v);
//     const customFilter = [];

//     if (Filter.Status) {
//       customFilter.push({
//         Op: "$contains",
//         Field: "Status",
//         Value: [Filter.Status],
//       });
//     }

//     if (Filter.Customer) {
//       const customer = await getDataCustomerContains(Filter.Customer);
//       if (customer) {
//         customFilter.push({
//           Op: "$contains",
//           Field: "Customer",
//           Value: [customer._id],
//         });
//       }
//     }

//     if (Filter.QuotationDate) {
//       const QuotationDate = new Date(Filter.QuotationDate);
//       const now = new Date();
//       if (QuotationDate.getTime() <= now.getTime()) {
//         customFilter.push({
//           Op: "$eq",
//           Field: "QuotationDate",
//           Value: QuotationDate,
//         });
//       }
//     }

//     if (customFilter.length > 0) {
//       data.CustomFiltering = {
//         Op: "$and",
//         Items: customFilter,
//       };
//     } else {
//       data.CustomFiltering = undefined;
//     }

//     debounce(() => {
//       listControl.value.refreshList();
//     }, 500);
//   }
// );

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

watch(
	() => route.query.title,
	(nv) => {
		data.title = nv;
		listControl.value.setControlMode("grid");
	}
);

function CancelPreview() {
	data.appMode = "grid";
	listControl.value.setControlMode("form");
}

function closeModalOpportunity() {
	generateModal.value.hide();
}

function openModalOpportunity() {
	generateModal.value.show();
}

async function CreatewithOpportunity() {
	try {
		data.opportunity = data.cacheopportunity;
		await axios.post(`/sdp/salesquotation/action-sq`, {
			_id: data.opportunity,
		});

		closeModalOpportunity();
		listControl.value.setControlMode("grid");
		listControl.value.refreshList();
		listControl.value.refreshForm();
	} catch (error) {
		util.showError(error);
	}
}

async function sentMail(item) {
	try {
		await axios.post(`/sdp/salesquotation/send-email`, {
			_id: item._id,
		});
		util.showInfo("Success Send....");
		debounce(() => {
			listControl.value.refreshList();
		}, 1);
	} catch (error) {
		util.showError(error);
	}
}

function openForm() {
	util.nextTickN(2, () => {
		// listControl.value.setFormFieldAttr("_id", "rules", roleID);
	});
}

async function newRecord() {
	data.formMode = "new";
	data.titleForm = "Create New Sales Quotation";
	data.opportunity = undefined;

	data.cacheopportunity = {};
	const records = {
		PostingProfileID: "",
		Status: "",
		CompanyID: "DEMO00",
		TransactionType: "Asset",
	};

	records.QuotationDate = new Date();

	setTimeout(() => {
		data.record = {...records};
		listControl.value.setFormRecord({...records});
	}, 1);

	openForm();
}

async function editRecord(record) {
	data.formMode = "edit";
	data.record = record;
	if (record.Customer) {
		const promisebatch = [];
		promisebatch.push(getDataCustomer(record.Customer));

		const res = await Promise.all(promisebatch);
		const customer = res[0];
		data.record.PersonalContact = customer.Detail.PersonalContact;
		data.record.Address = customer.Detail.Address;
		data.record.AddressDelivery = customer.Detail.DeliveryAddress;
		data.record.City = customer.Detail.City;
		data.record.CityDelivery = customer.Detail.DeliveryCity;
		data.record.Province = customer.Detail.Province;
		data.record.ProvinceDelivery = customer.Detail.DeliveryProvince;
		data.record.Country = customer.Detail.Country;
		data.record.CountryDelivery = customer.Detail.DeliveryCountry;
		data.record.Zipcode = customer.Detail.Zipcode;
		data.record.ZipcodeDelivery = customer.Detail.DeliveryZipcode;
		data.record.Name = customer.Detail.PersonalContact;
	}

	if (record.OpportunityNo) {
		data.cacheopportunity = record.OpportunityNo;
		data.opportunity = record.OpportunityNo;
	}

	data.titleForm = `Edit Sales Quotation - ${record.QuotationNo} | ${record.QuotationName}`;
	// data.titleForm = `Edit Sales Quotation | ${record._id}`;

	if (record.FooterAsset == "") {
		record.FooterAsset = undefined;
	}
	// else {
	// 	record.FooterAsset = getAsset(record.FooterAsset);
	// }

	if (record.LetterHeadAsset == "") {
		record.LetterHeadAsset = undefined;
	}

	if (record.JournalType) {
		getJournalType(record.JournalType, "init", {});
	}
	// else {
	// 	record.LetterHeadAsset = getAsset(record.LetterHeadAsset);
	// }

	openForm();
}

function onPreSave(record) {
	console.log(record);
}

function updateEditor(newValue) {
	listControl.value.setFormRecord(newValue);
}

function onCancelForm(mode) {
	if (mode === "grid") {
		data.titleForm = "Sales Quotation";
	}
}

const onFormFieldChanged = (name, v1, v2, old, record) => {
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

	if (name == "Customer") {
		getDataCustomer(v1).then((customer) => {
			record.Address = customer.Detail.Address;
			record.AddressDelivery = customer.Detail.DeliveryAddress;
			record.City = customer.Detail.City;
			record.CityDelivery = customer.Detail.DeliveryCity;
			record.Province = customer.Detail.Province;
			record.ProvinceDelivery = customer.Detail.DeliveryProvince;
			record.Country = customer.Detail.Country;
			record.CountryDelivery = customer.Detail.DeliveryCountry;
			record.Zipcode = customer.Detail.Zipcode;
			record.ZipcodeDelivery = customer.Detail.DeliveryZipcode;
			record.Name = customer.Name;
			const TaxCodes = [];
			if (customer.Detail.Tax1 !== "") {
				TaxCodes.push(customer.Detail.Tax1);
			}
			if (customer.Detail.Tax2 !== "") {
				TaxCodes.push(customer.Detail.Tax2);
			}

			changeTaxcode(TaxCodes, []);
			record.TaxCodes = TaxCodes;
		});
	}

	if (name === "TaxCodes") {
		changeTaxcode(v1, old);
	}

	if (name === "TransactionType") {
		record.TransactionType = v1;
	}

	listControl.value.setGridRecord(
		record,
		listControl.value.getGridCurrentIndex()
	);
};

const getDataCustomer = async (idCustomer) => {
	try {
		const dataresponse = await axios.post(`/bagong/customer/get`, [idCustomer]);
		return dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
};

const getDataCustomerContains = async (search) => {
	try {
		const dataresponse = await axios.post(`/tenant/customer/find`, {
			where: {
				op: "$contains",
				field: "Name",
				value: [search],
			},
		});
		return dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
};

async function saveForm() {
	try {
		// const record = listControl.value.getFormRecord();
		const record = data.record;
		// data.record = record;

		const URLdata =
			data.formMode == "edit"
				? "/sdp/salesquotation/update"
				: "/sdp/salesquotation/insert";

		await axios.post(URLdata, record);
		data.appMode = "preview";
		data.preview = record;
	} catch (error) {
		util.showError(error);
	}
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

		console.log(calc, calc.DiscountAmount, headerdiscount);
		listControl.value.setFormRecord({...records, ...calc});
	});
}

async function getLastSalesQuotation() {
	try {
		const dataresponse = await axios.post(`/sdp/salesquotation/find`, {
			Sort: ["-Created"],
			Take: 1,
		});
		return dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
}

function createDebounce() {
	let timeout = null;
	return function (fnc, delayMs) {
		clearTimeout(timeout);
		timeout = setTimeout(() => {
			fnc();
		}, delayMs || 500);
	};
}

async function onGridRefreshed() {
	try {
		const records = listControl.value.getGridRecords();
		const customers = records.map((rec) => rec.Customer);
		data.CustomersName = (
			await axios.post(`/tenant/customer/find`, {
				where: {
					op: "$in",
					field: "_id",
					value: customers,
				},
			})
		).data.map((dat) => ({Name: dat.Name, _id: dat._id}));

		// return dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
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
				() => {}
			);
		}
	});
}
function closePreview() {
  data.appMode = "grid";
}
function trxPostSubmit(record) {
	closePreview()
  setModeGrid();
}
function setModeGrid() {
	listControl.value.setControlMode("grid");
	listControl.value.refreshList();
}
function getJournalType(_id, action, item) {
	axios
		.post("/sdp/salesorderjournaltype/find?_id=" + _id, {sort: ["-_id"]})
		.then(
			(r) => {
				const dt = r.data[0];
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

function lookupCustomerPayloadBuilder(search, config, value) {
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

	if (qp.Where != undefined) {
		const items = [{Op: "$eq", Field: `IsActive`, Value: true}];
		items.push(qp.Where);
		qp.Where = {
			Op: "$and",
			items: items,
		};
	} else {
		qp.Where = {Op: "$eq", Field: `IsActive`, Value: true};
	}

	return qp;
}

function refreshData() {
	util.nextTickN(2, () => {
		listControl.value.refreshGrid();
	});
}
function onAlterGridConfig(cfg) {
	cfg.setting.idField = "Created";
	cfg.setting.sortable = ["_id", "Created", "QuotationDate"];
}
function onAlterFormConfig(cfg) {
  if (route.query.id !== undefined) {
    let currQuery = {...route.query};
    listControl.value.selectData({ _id: currQuery.id });
    delete currQuery['id'];
    router.replace({ path: route.path, query: currQuery });
  }
}
async function GetTaxes(_id) {
	return (await axios.post(`/fico/taxcode/get`, [_id])).data;
}

function changeTaxcode(v1, old) {
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
</script>
