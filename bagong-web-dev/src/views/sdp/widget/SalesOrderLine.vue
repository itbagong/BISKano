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
			:grid-hide-delete="props.quotation"
			grid-no-confirm-delete
			grid-auto-commit-line
			:grid-hide-new="props.quotation"
			init-app-mode="grid"
			grid-mode="grid"
			form-keep-label
			new-record-type="grid"
			:grid-config="props.gridConfig"
			:grid-fields="[
				'Asset',
				'Item',
				'Spesifications',
				'Qty',
				'Description',
				'Uom',
				'Account',
				'UnitPrice',
				'Discount',
				// 'DiscountType',
			]"
			:form-tabs-edit="['General', 'Checklist', 'References']"
			:form-tabs-view="['General', 'Checklist', 'References']"
			:form-config="props.formConfig"
			:form-fields="['Dimension', 'TaxCodes']"
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
					v-if="item.Item == '' || item.Item == undefined"
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
				<div v-else></div>
			</template>

			<template #grid_Item="{item}">
				<!-- <s-input
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
				></s-input> -->
				<s-input-sku-item
					v-if="item.Asset == '' || item.Asset == undefined"
					ref="refItemVarian"
					@update:model-value="
						(newValue) => {
							if (newValue) {
								const valuesplit = newValue.split('~~');
								data.dataSpesification = [valuesplit[1]];
								item.Spesifications = [valuesplit[1]];
								item['Item'] = valuesplit[0];

								onGridRowFieldChanged(
									'Item',
									valuesplit[0],
									valuesplit[0],
									item['Item'],
									item
								);
							} else {
								data.dataSpesification = [];
								item.Spesifications = [];
								item['Item'] = undefined;

								onGridRowFieldChanged(
									'Item',
									undefined,
									undefined,
									item['Item'],
									item
								);
							}

							item.SKUItem = newValue;
						}
					"
					:model-value="item.SKUItem"
					:record="item.ItemVariant"
					:lookup-url="`/tenant/item/gets-detail`"
					:lookup-payload-builder="
						(search) =>
							lookupPayloadBuilderSinput(
								search,
								['ID', 'Text'],
								item.ItemVarian,
								item
							)
					"
					:disabled="item.Asset && item.Asset.length > 0 ? true : false"
				></s-input-sku-item>

				<div v-else></div>
			</template>

			<template #grid_Spesifications="{item}">
				<s-input
					v-if="item.Asset != '' && item.Asset != undefined"
					ref="refSpesifications"
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
						(search) => lookupPayloadBuilder(search, item.Spesifications, item)
					"
				></s-input>
				<div v-else></div>
			</template>

			<template #grid_Qty="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'Qty'
					)"
					:key="`gridSpesificationsinput ${hdrindex}`"
				>
					<s-input
						hide-label
						:ctl-ref="{rowIndex: item.index}"
						:field="hdr.input.field"
						:kind="hdr.input.kind"
						:label="
							hdr.input.kind == 'checkbox' || hdr.input.kind == 'bool'
								? ''
								: hdr.input.label
						"
						:disabled="hdr.input.readOnly || props.quotation"
						:caption="hdr.input.caption"
						:hint="hdr.input.hint"
						:multi-row="hdr.input.multiRow"
						:use-list="hdr.input.useList"
						:items="hdr.input.items"
						:rules="hdr.input.rules"
						:required="hdr.input.required"
						:read-only="hdr.input.readOnly || props.quotation"
						:lookup-url="hdr.input.lookupUrl"
						:lookup-key="hdr.input.lookupKey"
						:allow-add="hdr.input.allowAdd"
						:lookup-format1="hdr.input.lookupFormat1"
						:lookup-format2="hdr.input.lookupFormat2"
						:decimal="hdr.input.decimal"
						:date-format="hdr.input.dateFormat"
						:multiple="hdr.input.multiple"
						:lookup-labels="hdr.input.lookupLabels"
						:lookup-searchs="
							hdr.input.lookupSearchs && hdr.input.lookupSearchs.length == 0
								? hdr.input.lookupLabels
								: hdr.input.lookupSearchs
						"
						@focus="rowFieldFocus"
						@change="rowFieldChanged"
						v-model="item[hdr.input.field]"
						ref="inputs"
					/>
				</div>
			</template>

			<template #grid_Description="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'Description'
					)"
					:key="`gridSpesificationsinput ${hdrindex}`"
				>
					<s-input
						hide-label
						:ctl-ref="{rowIndex: item.index}"
						:field="hdr.input.field"
						:kind="hdr.input.kind"
						:label="
							hdr.input.kind == 'checkbox' || hdr.input.kind == 'bool'
								? ''
								: hdr.input.label
						"
						:disabled="hdr.input.readOnly || props.quotation"
						:caption="hdr.input.caption"
						:hint="hdr.input.hint"
						:multi-row="hdr.input.multiRow"
						:use-list="hdr.input.useList"
						:items="hdr.input.items"
						:rules="hdr.input.rules"
						:required="hdr.input.required"
						:read-only="hdr.input.readOnly || props.quotation"
						:lookup-url="hdr.input.lookupUrl"
						:lookup-key="hdr.input.lookupKey"
						:allow-add="hdr.input.allowAdd"
						:lookup-format1="hdr.input.lookupFormat1"
						:lookup-format2="hdr.input.lookupFormat2"
						:decimal="hdr.input.decimal"
						:date-format="hdr.input.dateFormat"
						:multiple="hdr.input.multiple"
						:lookup-labels="hdr.input.lookupLabels"
						:lookup-searchs="
							hdr.input.lookupSearchs && hdr.input.lookupSearchs.length == 0
								? hdr.input.lookupLabels
								: hdr.input.lookupSearchs
						"
						@focus="rowFieldFocus"
						@change="rowFieldChanged"
						v-model="item[hdr.input.field]"
						ref="inputs"
					/>
				</div>
			</template>

			<template #grid_UnitPrice="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'UnitPrice'
					)"
					:key="`gridUnitPriceinput ${hdrindex}`"
				>
					<input
						type="text"
						:placeholder="hdr.input.caption || hdr.input.label"
						class="input_field text-right w-[100px]"
						:value="valueUnitPrice.get(item['UnitPrice'])"
						@change="(val) => valueUnitPrice.set(val, item)"
						ref="control"
						:disabled="hdr.inputdisabled"
					/>
				</div>
			</template>

			<template #grid_Discount="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'Discount'
					)"
					:key="`gridDiscountinput ${hdrindex}`"
				>
					<s-input
						hide-label
						:ctl-ref="{rowIndex: item.index}"
						:field="hdr.input.field"
						kind="text"
						:label="
							hdr.input.kind == 'checkbox' || hdr.input.kind == 'bool'
								? ''
								: hdr.input.label
						"
						:disabled="hdr.input.readOnly || props.quotation"
						:caption="hdr.input.caption"
						:hint="hdr.input.hint"
						:multi-row="hdr.input.multiRow"
						:use-list="hdr.input.useList"
						:items="hdr.input.items"
						:rules="hdr.input.rules"
						:required="hdr.input.required"
						:read-only="hdr.input.readOnly"
						:lookup-url="hdr.input.lookupUrl"
						:lookup-key="hdr.input.lookupKey"
						:allow-add="hdr.input.allowAdd"
						:lookup-format1="hdr.input.lookupFormat1"
						:lookup-format2="hdr.input.lookupFormat2"
						:decimal="hdr.input.decimal"
						:date-format="hdr.input.dateFormat"
						:multiple="hdr.input.multiple"
						:lookup-labels="hdr.input.lookupLabels"
						:lookup-searchs="
							hdr.input.lookupSearchs && hdr.input.lookupSearchs.length == 0
								? hdr.input.lookupLabels
								: hdr.input.lookupSearchs
						"
						@focus="rowFieldFocus"
						@change="rowFieldChanged"
						:model-value="valueDiscountPrice.get(item[hdr.input.field])"
						@update:model-value="(val) => valueDiscountPrice.set(val, item)"
						ref="inputs"
					/>
					<!-- <input
						type="text"
						:placeholder="hdr.input.caption || hdr.input.label"
						class="input_field text-right w-[100px]"
						:value="valueDiscountPrice.get(item['Discount'])"
						@change="(val) => valueDiscountPrice.set(val, item)"
						ref="control"
						:disabled="hdr.inputdisabled"
					/> -->
				</div>
			</template>

			<template #grid_Uom="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'Uom'
					)"
					:key="`gridSpesificationsinput ${hdrindex}`"
				>
					<s-input
						hide-label
						:ctl-ref="{rowIndex: item.index}"
						:field="hdr.input.field"
						:kind="hdr.input.kind"
						:label="
							hdr.input.kind == 'checkbox' || hdr.input.kind == 'bool'
								? ''
								: hdr.input.label
						"
						:disabled="hdr.input.readOnly || props.quotation"
						:caption="hdr.input.caption"
						:hint="hdr.input.hint"
						:multi-row="hdr.input.multiRow"
						:use-list="hdr.input.useList"
						:items="hdr.input.items"
						:rules="hdr.input.rules"
						:required="hdr.input.required"
						:read-only="hdr.input.readOnly"
						:lookup-url="hdr.input.lookupUrl"
						:lookup-key="hdr.input.lookupKey"
						:allow-add="hdr.input.allowAdd"
						:lookup-format1="hdr.input.lookupFormat1"
						:lookup-format2="hdr.input.lookupFormat2"
						:decimal="hdr.input.decimal"
						:date-format="hdr.input.dateFormat"
						:multiple="hdr.input.multiple"
						:lookup-labels="hdr.input.lookupLabels"
						:lookup-searchs="
							hdr.input.lookupSearchs && hdr.input.lookupSearchs.length == 0
								? hdr.input.lookupLabels
								: hdr.input.lookupSearchs
						"
						@focus="rowFieldFocus"
						@change="rowFieldChanged"
						v-model="item[hdr.input.field]"
						ref="inputs"
					/>
				</div>
			</template>

			<template #grid_Account="{item}">
				<AccountSelector
					v-model="item.Account"
					:items-type="['LEDGERACCOUNT']"
					hide-account-type
					hide-label
				></AccountSelector>
			</template>

			<template #form_tab_References="{item}">
				<References
					:ReferenceTemplate="data.jurnalType.ReferenceTemplateID"
					:readOnly="readOnly || item.mode == 'view'"
					v-model="item.References"
				/>
			</template>

			<template #form_tab_Checklist="{item, mode}">
				<Checklist
					v-model="item.Checklists"
					:checklist-id="data.jurnalType.ChecklistTemplateID"
					:readOnly="readOnly || mode == 'view'"
				/>
			</template>

			<!-- <template #grid_DiscountType="{item}">
				<s-input
					ref="refDiscountType"
					hide-label
					v-model="item.DiscountType"
					use-list
					:items="['Value', 'percent']"
					class="w-100"
					@change="
						(field, v1, v2, old, ctlRef) => {
							onGridRowFieldChanged('DiscountType', v1, v2, old, item);
						}
					"
				></s-input>
			</template> -->

			<template #form_input_TaxCodes="{item, config}">
				<s-input
					:field="config.field"
					:kind="config.kind"
					:label="config.label"
					@change="
						(name, v1, v2, old) => onFormFieldChanged(name, v1, v2, old, item)
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
import {DataList, util, SInput} from "suimjs";
import {inject, ref, reactive, watch, onMounted} from "vue";
import DimensionItem from "./DimensionItem.vue";
import moment from "moment";
import AccountSelector from "@/components/common/AccountSelector.vue";
import helper from "@/scripts/helper.js";
import SInputSkuItem from "../../scm/widget/SInputSkuItem.vue";
import References from "@/components/common/References.vue";
import Checklist from "@/components/common/Checklist.vue";

const axios = inject("axios");

const props = defineProps({
	gridConfig: {type: String, default: () => ""},
	formConfig: {type: String, default: () => ""},
	modelValue: {type: Array, default: () => []},
	quotation: {type: Boolean, default: () => false},
	tax: {type: Array, default: () => []},
	salesPriceBook: {type: String, default: () => ""},
	journalTypeID: {type: String, default: () => ""},
	trxType: {type: String, default: () => ""},
});

const emit = defineEmits({
	"update:modelValue": null,
	recalc: null,
});

const data = reactive({
	records: (props.modelValue ?? []).map((dt, index) => {
		dt.suimRecordChange = false;
		dt.index = index;
		// dt.SKUItem = helper.ItemVarian(dt.Item, dt.Spesifications[0]);
		dt.ItemVariant = {};

		if (!dt.SKUItem && dt.Item) {
			dt.SKUItem = helper.ItemVarian(dt.Item, dt.Spesifications);
		}

		return dt;
	}),

	jurnalType: {},
	changed: false,
	defaultgridconfig: {},
	currentIndex: -1,
	recordChanged: false,
	SalesPriceBook: {},
	TaxRates: [],
	dataSpesification: [],
	cfg: {},
});

const valueDiscountPrice = {
	get(v) {
		return util.formatMoney(v, {decimal: 0});
	},
	set(v, record) {
		const v1 = Number(String(v).replaceAll(",", ""));
		record["Discount"] = v1;
		onGridRowFieldChanged("Discount", v1, 0, 0, record);
	},
};

const valueUnitPrice = {
	get(v) {
		return util.formatMoney(v, {decimal: 0});
	},
	set(v, record) {
		const v1 = Number(String(v.target.value).replaceAll(",", ""));
		const spb = (data.SalesPriceBook.Lines || []).find((line) => {
			if (record.Item === line.ItemID) {
				return true;
			}

			if (record.Asset === line.AssetID) {
				return true;
			}

			return false;
		});
		if (spb && v1 < spb.MinPrice) {
			util.showWarning("Unit Price must not be lower than Sales Price Book");
		}

		record.UnitPrice = v1;
		record.Amount = v1 * record.Qty * record.ContractPeriod;
		listControl.value.setGridRecord(
			record,
			listControl.value.getGridCurrentIndex()
		);
		updateItems();
	},
};

watch(
	() => props.modelValue,
	(nt) => {
		if (data.changed == false) {
			const records = (nt ?? []).map((dt, index) => {
				dt.suimRecordChange = false;
				dt.index = index;
				// dt.SKUItem = helper.ItemVarian(dt.Item, dt.Spesifications[0]);
				dt.ItemVariant = {};

				if (!dt.SKUItem && dt.Item) {
					dt.SKUItem = helper.ItemVarian(dt.Item, dt.Spesifications);
				}

				return dt;
			});

			listControl.value.setGridRecords(records);
		}

		data.changed = false;
	}
);

watch(() => props.trxType, (nv) => {
	util.nextTickN(2, () => {
		AlterGridConfig(data.cfg);
	});
});

const listControl = ref(null);

watch(
	() => props.salesPriceBook,
	async (_id) => {
		if (_id != "") {
			data.SalesPriceBook = await GetSalesPriceBook(_id);
		}
	}
);

watch(
	() => props.journalTypeID,
	async (_id) => {
		if (_id) {
			console.log(_id);
			getJournalType(_id, "init", {});
		}
	}
);

const newRecord = () => {
	const records = listControl.value.getGridRecords();
	records.push({
		Asset: "",
		Item: "",
		SKUItem: "",
		ItemVariant: {},
		Description: "",
		Qty: 0,
		ContractPeriod: 1,
		UnitPrice: 0,
		Amount: 0,
		Discount: 0,
		DiscountType: "",
		Taxable: props.tax && props.tax.length > 0 ? true : false,
		TaxCodes: props.tax,
		GetTax: [],
		index: records.length,
		StartDate: new Date(),
		EndDate: new Date(),
	});
	// console.log(records, props.tax)

	listControl.value.setGridRecords(records);
	updateItems();
};

async function onFormFieldChanged(name, v1, v2, old, record) {
	if (name == "TaxCodes") {
		if (v1 && v1.length > 0) {
			record.Taxable = true;
		} else {
			record.Taxable = false;
		}
	}

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

function getJournalType(_id, action, item) {
	axios
		.post("/sdp/salesorderjournaltype/find?_id=" + _id, {sort: ["-_id"]})
		.then(
			(r) => {
				const dt = r.data[0];
				data.jurnalType = dt;
			},
			(e) => util.showError(e)
		);
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
	const typeVal = typeof v1 === "string";
	if (name == "Item") {
		if (v1 && typeVal) {
			const itemtenant = getItemTenant(v1);

			Promise.all([itemtenant]).then(([respitemtenant]) => {
				record.PhysicalDimension = respitemtenant.PhysicalDimension;
				record.FinanceDimension = respitemtenant.FinanceDimension;
				record.Uom = respitemtenant.DefaultUnitID;
			});
		} else {
			if (record) {
				record.PhysicalDimension = {};
				record.FinanceDimension = {};
				record.Uom = undefined;
				record.Spesifications = [];
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
			data.dataSpesification = [];
		}
		if (!typeVal) {
			record.Asset = "";
			record.Item = "";
			record.Spesifications = [];
			data.dataSpesification = [];
		}
	}

	if (name == "Qty") {
		record.Amount = v1 * record.UnitPrice * record.ContractPeriod;
		// data.TaxeRates.map((tax) => {
		// 	record.TaxAmount += record.Amount * (tax.Rate / 100);
		// });
	}
	if (name == "UnitPrice") {
		const spb = (data.SalesPriceBook.Lines || []).find((line) => {
			if (record.Item === line.ItemID) {
				return true;
			}

			if (record.Asset === line.AssetID) {
				return true;
			}

			return false;
		});
		if (spb && v1 < spb.MinPrice) {
			util.showWarning("Unit Price must not be lower than Sales Price Book");
		}

		record.Amount = v1 * record.Qty * record.ContractPeriod;
		// data.TaxeRates.map((tax) => {
		// 	record.TaxAmount += record.Amount * (tax.Rate / 100);
		// });
	}
	if (name == "ContractPeriod") {
		if (v1 == undefined || v1 == "" || v1 < 1) v1 = 1;
		record.Amount = v1 * record.UnitPrice * record.Qty;
		// data.TaxeRates.map((tax) => {
		// 	record.TaxAmount += record.Amount * (tax.Rate / 100);
		// });

		const datestart = new Date(record.StartDate);
		const dateend = new Date(record.StartDate);

		switch (String(record.Uom).toLowerCase()) {
			case "days":
				dateend.setDate(datestart.getDate() + v1);
				break;
			case "day":
				dateend.setDate(datestart.getDate() + v1);
				break;
			case "month":
				dateend.setMonth(datestart.getMonth() + v1);
				break;
			case "months":
				dateend.setMonth(datestart.getMonth() + v1);
				break;

			default:
				break;
		}

		record.EndDate = moment(dateend).local().format();
	}

	if (name == "Discount") {
		if (record.DiscountType == "fixed") {
			record.DiscountAmount = v1;
		} else if (record.DiscountType == "percent") {
			record.DiscountAmount = record.Amount * (v1 / 100);
		}
	}

	if (name == "DiscountType") {
		if (v1 == "fixed") {
			record.DiscountAmount = record.Discount;
		} else if (v1 == "percent") {
			record.DiscountAmount = record.Amount * (record.Discount / 100);
		}
	}

	if (name == "StartDate") {
		const datestart = new Date(v1);
		const dateend = new Date(v1);

		switch (String(record.Uom).toLowerCase()) {
			case "days":
				dateend.setDate(datestart.getDate() + record.ContractPeriod);
				break;
			case "day":
				dateend.setDate(datestart.getDate() + record.ContractPeriod);
				break;
			case "month":
				dateend.setMonth(datestart.getMonth() + record.ContractPeriod);
				break;
			case "months":
				dateend.setMonth(datestart.getMonth() + record.ContractPeriod);
				break;

			default:
				break;
		}

		record.EndDate = moment(dateend).local().format();

		// const days =
		// 	record.UoM == "MONTH"
		// 		? record.ContractPeriod * 30
		// 		: record.ContractPeriod;
		// record.EndDate = moment(v1)
		// 	.add(days || 0, "days")
		// 	.format();
	}

	if (name == "Uom") {
		const datestart = new Date(record.StartDate);
		const dateend = new Date(record.StartDate);

		switch (String(v1).toLowerCase()) {
			case "days":
				dateend.setDate(datestart.getDate() + record.ContractPeriod);
				break;
			case "day":
				dateend.setDate(datestart.getDate() + record.ContractPeriod);
				break;
			case "month":
				dateend.setMonth(datestart.getMonth() + record.ContractPeriod);
				break;
			case "months":
				dateend.setMonth(datestart.getMonth() + record.ContractPeriod);
				break;

			default:
				break;
		}

		record.EndDate = moment(dateend).local().format();
	}

	listControl.value.setGridRecord(
		record,
		listControl.value.getGridCurrentIndex()
	);

	updateItems();
}

async function gridRefreshed() {
	listControl.value.setGridRecords(data.records);
}

const AlterGridConfig = (gridconfigs) => {
	data.cfg = gridconfigs;

	const records = data.records;

	gridconfigs.fields?.map((fields) => {
		if (props.trxType == 'Asset') {
			// hide in asset
			["Item", "Dimension"]?.map((flds) => {
				if (fields.field == flds) {
					fields.readType = "hide";
				}
			});

			// show in asset
			[
				"Asset",
				"Spesifications",
				"Shift",
				"ContractPeriod",
				"Dimension",
			]?.map((flds) => {
				if (fields.field == flds) {
					fields.readType = "show";
				}
			})
		} else {
			// hide in asset
			["Item", "Dimension"]?.map((flds) => {
				if (fields.field == flds) {
					fields.readType = "show";
				}
			});

			// show in asset
			[
				"Asset",
				"Spesifications",
				"Shift",
				"ContractPeriod",
				"Dimension",
			]?.map((flds) => {
				if (fields.field == flds) {
					fields.readType = "hide";
				}
			})
		}
		
		return fields;
	});

	const assetfield = gridconfigs.fields;

	data.defaultgridconfig = assetfield;
	// if (records.length > 0) {
	// 	records.map((record) => {
	// 		const gridcfg = assetfield;
	// 		if (record.Item !== "") {
	// 			const itemtenant = getItemTenant(record.Item);
	// 			const itemspec = getItemSpecs(record.Item);

	// 			Promise.all([itemtenant, itemspec]).then(
	// 				([respitemtenant, resitemspec]) => {
	// 					record.PhysicalDimension = respitemtenant.PhysicalDimension;
	// 					record.FinanceDimension = respitemtenant.FinanceDimension;
	// 					record.Uom = respitemtenant.DefaultUnitID;
	// 					record.Spesifications = resitemspec.map(
	// 						(item) => item.SpecVariantID
	// 					);
	// 				}
	// 			);

	// 			gridcfg.map((config) => {
	// 				if (config.field === "Asset") {
	// 					if (record.Item !== "") {
	// 						config.input.disable = true;
	// 						config.input.readOnly = true;
	// 						config.input.readOnlyOnEdit = true;
	// 						config.input.readOnlyOnNew = true;
	// 					} else {
	// 						config.input.disable = false;
	// 						config.input.readOnly = false;
	// 						config.input.readOnlyOnEdit = false;
	// 						config.input.readOnlyOnNew = false;
	// 					}
	// 				}
	// 				return config;
	// 			});

	// 			record.Asset = "";
	// 		}

	// 		if (record.Asset !== "") {
	// 			gridcfg.map((config) => {
	// 				if (config.field === "Item") {
	// 					if (record.Asset !== "") {
	// 						config.input.disable = true;
	// 						config.input.readOnly = true;
	// 						config.input.readOnlyOnEdit = true;
	// 						config.input.readOnlyOnNew = true;
	// 					} else {
	// 						config.input.disable = false;
	// 						config.input.readOnly = false;
	// 						config.input.readOnlyOnEdit = false;
	// 						config.input.readOnlyOnNew = false;
	// 					}
	// 				}
	// 				return config;
	// 			});

	// 			record.Item = "";
	// 		}

	// 		console.log("AlterGridConfig - Map : ", gridcfg, records, assetfield);
	// 		return gridcfg;
	// 	});
	// }
};

async function getItemsTax(TaxCodes) {
	try {
		const dataresponse = await axios.post(`/fico/taxsetup/find`, {
			Select: ["_id", "Rate", "InvoiceOperation"],
			Where: {
				Op: "$in",
				Field: "_id",
				Value: TaxCodes,
			},
		});

		return dataresponse.data;
	} catch (error) {
		util.showError(error);
	}
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

function updateItems() {
	const records = listControl.value.getGridRecords();
	data.records = [...records];
	const committedRecords = records.filter(
		(dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
	);

	util.nextTickN(2, () => {
		emit("recalc", ReCalc(committedRecords));
		emit("update:modelValue", committedRecords);
		data.changed = true;
	});
}

const ReCalc = async (Lines) => {
	if (Lines.length > 0) {
		const calc = {
			SubTotalAmount: 0,
			TaxAmount: 0,
			DiscountAmount: 0,
			TotalAmount: 0,
		};
		for (const record of Lines) {
			calc.SubTotalAmount += record.Amount;

			let taxamount = 0;
			if (record.TaxCodes && record.TaxCodes.length > 0) {
				const GetTaxes = (await getItemsTax(record.TaxCodes)).map((item) => ({
					Rate: item.Rate,
					_id: item._id,
					InvoiceOperation: item.InvoiceOperation,
				}));

				for (const tax of GetTaxes) {
					if (tax.InvoiceOperation == "Decrease") {
						taxamount -= record.Amount * tax.Rate;
					} else if (tax.InvoiceOperation == "Increase") {
						taxamount += record.Amount * tax.Rate;
					}
				}

				calc.TaxAmount = taxamount;
			}

			let discountAmount = 0;
			if (record.Discount && record.Discount > 0) {
				// if (record.DiscountType == "") {
				// 	record.DiscountType = "fixed";
				// }

				if (record.DiscountType == "fixed") {
					discountAmount = record.Discount;
				} else if (record.DiscountType == "percent") {
					discountAmount = record.Amount * (record.Discount / 100);
				} else if (record.DiscountType == "percent") {
					discountAmount = record.Amount * (record.Discount / 100);
				}
			}

			calc.DiscountAmount += discountAmount;
			calc.TotalAmount += record.Amount + taxamount - discountAmount;
		}

		return calc;
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

function rowFieldChanged(name, v1, v2) {
	const currentIndex = data.currentIndex;
	const current = data.records[currentIndex];

	onGridRowFieldChanged(name, v1, v2, current, current);
	// emit("rowFieldChanged", name, v1, v2, current, current);
	// emit("update:modelValue", data.items);
}

async function GetSalesPriceBook(_id) {
	try {
		const dataresponse = await axios.post(`/sdp/salespricebook/find`, {
			Where: {
				Op: "$eq",
				Field: "_id",
				Value: _id,
			},
		});

		return dataresponse.data[0];
	} catch (error) {
		util.showError(error);
	}
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

	if (
		item.Item !== "" &&
		item.Item !== null &&
		data.dataSpesification.length > 0
	) {
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
	} else if (
		item.Item !== "" &&
		item.Item !== null &&
		item.Spesifications.length > 0 &&
		data.dataSpesification.length == 0
	) {
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

onMounted(async () => {
	if (props.salesPriceBook != "") {
		data.SalesPriceBook = await GetSalesPriceBook(props.salesPriceBook);
	}
	if (props.journalTypeID) {
		getJournalType(props.journalTypeID, "init", {});
	}
});

function lookupPayloadBuilderSinput(search, select, value, item) {
	const qp = {};
	qp.Take = 20;
	qp.Sort = [select[0]];
	qp.Select = select;

	//setting search
	const query = [
		// {
		// 	Field: "ExcludeItemGroupID",
		// 	Op: "$nin",
		// 	Value: ["GRP0026", "GRP0023"],
		// },
	];
	qp.Where = {
		Op: "$and",
		items: query,
	};
	if (search !== "" && search !== null) {
		let items = [
			{
				Op: "$or",
				items: [
					{Field: "Text", Op: "$contains", Value: [search]},
					{Field: "ID", Op: "$contains", Value: [search]},
				],
			},
		];
		items = [...items, ...query];
		qp.Where = {
			Op: "$and",
			items: items,
		};
	}
	return qp;
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
</script>
