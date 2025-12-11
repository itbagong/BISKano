<template>
	<div class="flex flex-col gap-2">
		<data-list
			ref="listControl"
			title="Sales Quotation"
			hide-title
			no-gap
			grid-editor
			grid-hide-select
			grid-hide-search
			grid-hide-sort
			grid-hide-refresh
			:grid-hide-delete="props.opportunity"
			grid-no-confirm-delete
			grid-auto-commit-line
			:grid-hide-new="props.opportunity"
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
				'UnitPrice',
				'TaxCodes',
				'Discount',
			]"
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
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'Asset'
					)"
					:key="`gridassetinput ${hdrindex}`"
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
						:disabled="
							item.Item && item.Item.length > 0
								? true
								: false || hdr.input.readOnly || props.opportunity
						"
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

			<template #grid_Item="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'Item'
					)"
					:key="`griditeminput ${hdrindex}`"
				>
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
								lookupPayloadBuilder(
									search,
									['ID', 'Text'],
									item.ItemVarian,
									item
								)
						"
						:disabled="
							item.Asset && item.Asset.length > 0
								? true
								: false || hdr.input.readOnly || props.opportunity
						"
						@afterOnChange="
							(val) => {
								onGridRowFieldChanged(
									hdr.input.field,
									val.Item._id,
									val.Item._id,
									item[hdr.input.field],
									item
								);
								item.Spesifications = [val.ItemSpec._id];
								item[hdr.input.field] = val.Item._id;
							}
						"
					></s-input-sku-item>
					<div v-else></div>

					<!-- <s-input
						hide-label
						:ctl-ref="{rowIndex: item.index}"
						:field="hdr.input.field"
						:kind="hdr.input.kind"
						:label="
							hdr.input.kind == 'checkbox' || hdr.input.kind == 'bool'
								? ''
								: hdr.input.label
						"
						:disabled="
							item.Asset && item.Asset.length > 0
								? true
								: false || hdr.input.readOnly || props.opportunity
						"
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
					/> -->
				</div>
			</template>

			<template #grid_Spesifications="{item}">
				<div
					v-if="data.defaultgridconfig"
					v-for="(hdr, hdrindex) in data.defaultgridconfig.filter(
						(dt) => dt.field === 'Spesifications'
					)"
					:key="`gridSpesificationsinput ${hdrindex}`"
				>
					<s-input
						v-if="item.Asset != '' && item.Asset != undefined"
						hide-label
						:ctl-ref="{rowIndex: item.index}"
						:field="hdr.input.field"
						:kind="hdr.input.kind"
						:label="
							hdr.input.kind == 'checkbox' || hdr.input.kind == 'bool'
								? ''
								: hdr.input.label
						"
						:disabled="hdr.input.readOnly"
						:caption="hdr.input.caption"
						:hint="hdr.input.hint"
						:multi-row="hdr.input.multiRow"
						:use-list="hdr.input.useList"
						:items="hdr.input.items"
						:rules="hdr.input.rules"
						:required="hdr.input.required"
						:read-only="hdr.input.readOnly"
						:lookup-url="
							item.Asset !== '' && item.Asset !== undefined
								? `/tenant/masterdata/find?MasterDataTypeID=SPC`
								: `/tenant/specvariant/find`
						"
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
						:lookup-payload-builder="
							(search) =>
								lookupSpecsPayloadBuilder(
									search,
									hdr.input,
									item[hdr.input.field]
								)
						"
						@focus="rowFieldFocus"
						@change="rowFieldChanged"
						v-model="item[hdr.input.field]"
						ref="inputs"
					/>
					<div v-else></div>
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
						hide-label
						:ctl-ref="{rowIndex: item.index}"
						:field="hdr.input.field"
						:kind="hdr.input.kind"
						:label="
							hdr.input.kind == 'checkbox' || hdr.input.kind == 'bool'
								? ''
								: hdr.input.label
						"
						:disabled="hdr.input.readOnly || props.opportunity"
						:caption="hdr.input.caption"
						:hint="hdr.input.hint"
						:multi-row="hdr.input.multiRow"
						:use-list="hdr.input.useList"
						:items="hdr.input.items"
						:rules="hdr.input.rules"
						:required="hdr.input.required"
						:read-only="hdr.input.readOnly || props.opportunity"
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
						class="input_field text-right"
						:value="valueUnitPrice.get(item['UnitPrice'])"
						@change="(val) => valueUnitPrice.set(val, item)"
						ref="control"
						:disabled="hdr.inputdisabled"
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
						:disabled="hdr.input.readOnly || props.opportunity"
						:caption="hdr.input.caption"
						:hint="hdr.input.hint"
						:multi-row="hdr.input.multiRow"
						:use-list="hdr.input.useList"
						:items="hdr.input.items"
						:rules="hdr.input.rules"
						:required="hdr.input.required"
						:read-only="hdr.input.readOnly || props.opportunity"
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
						:disabled="hdr.input.readOnly || props.opportunity"
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
						v-model="item['UoM']"
						ref="inputs"
					/>
				</div>
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
		</data-list>
	</div>
</template>

<script setup>
import {DataList, util, SInput} from "suimjs";
import {inject, ref, reactive, watch, onMounted} from "vue";
import DimensionItem from "./DimensionItem.vue";
import helper from "@/scripts/helper.js";
import moment from "moment";
import SInputSkuItem from "../../scm/widget/SInputSkuItem.vue";

const axios = inject("axios");

const props = defineProps({
	gridConfig: {type: String, default: () => ""},
	formConfig: {type: String, default: () => ""},
	modelValue: {type: Array, default: () => []},
	opportunity: {type: Boolean, default: () => false},
	taxCodes: {type: Array, default: () => []},
	salesPriceBook: {type: String, default: () => ""},
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
		if (!dt.SKUItem && dt.Item) {
			dt.SKUItem = helper.ItemVarian(dt.Item, dt.Spesifications);
		}
		dt.ItemVariant = {};
		return dt;
	}),

	changed: false,
	defaultgridconfig: {},
	currentIndex: -1,
	recordChanged: false,
	SalesPriceBook: {},
	cfg: {},
});

watch(() => props.trxType, (nv) => {
	util.nextTickN(2, () => {
		AlterGridConfig(data.cfg);
	});
})

watch(
	() => props.modelValue,
	(nt) => {
		if (data.changed == false) {
			const records = (nt ?? []).map((dt, index) => {
				dt.suimRecordChange = false;
				dt.index = index;
				dt.ItemVariant = {};

				if (!dt.SKUItem && dt.Item) {
					dt.SKUItem = helper.ItemVarian(dt.Item, dt.Spesifications);
				}
				return dt;
			});

			data.records = records;
			listControl.value.setGridRecords(records);
		}
		data.changed = false;
	}
);

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

const listControl = ref(null);

watch(
	() => props.salesPriceBook,
	async (_id) => {
		if (_id != "") {
			data.SalesPriceBook = await GetSalesPriceBook(_id);
		}
	}
);

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
		if (record.ContractPeriod > 0) {
			record.Amount = v1 * record.Qty * record.ContractPeriod;
		} else {
			record.Amount = v1 * record.Qty;
		}
		listControl.value.setGridRecord(
			record,
			listControl.value.getGridCurrentIndex()
		);
		updateItems();
	},
};

const newRecord = () => {
	const records = listControl.value.getGridRecords();
	records.push({
		Item: "",
		SKUItem: "",
		ItemVariant: {},
		Description: "",
		Qty: 0,
		ContractPeriod: 1,
		UnitPrice: 0,
		Amount: 0,
		Discount: 0,
		Taxable: props.taxCodes.length > 0,
		TaxCodes: [...props.taxCodes],
		GetTax: [],
		index: records.length,
	});

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

function onGridRowFieldChanged(name, v1, v2, old, record) {
	if (name == "Item") {
		if (v1 != "") {
			const itemtenant = getItemTenant(v1);

			Promise.all([itemtenant]).then(([respitemtenant]) => {
				record.PhysicalDimension = respitemtenant.PhysicalDimension;
				record.FinanceDimension = respitemtenant.FinanceDimension;
				record.UoM = respitemtenant.DefaultUnitID;
			});
		} else {
			if (record) {
				record.PhysicalDimension = {};
				record.FinanceDimension = {};
				record.UoM = undefined;
				record.Spesifications = undefined;
			} else {
				record = {
					PhysicalDimension: {},
					FinanceDimension: {},
				};
			}
		}

		if (record) {
			record.Asset = undefined;
		}
	}

	if (name == "Asset") {
		if (v1.length <= 0 && record) {
			record.Spesifications = undefined;
		}

		if (record) {
			record.Item = undefined;
		}
	}

	if (name == "Qty") {
		record.Amount = v1 * record.UnitPrice * record.ContractPeriod;
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
	}
	if (name == "ContractPeriod") {
		record.Amount = v1 * record.UnitPrice * record.Qty;

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

	listControl.value.setGridRecord(
		record,
		listControl.value.getGridCurrentIndex()
	);

	setTimeout(() => {
		updateItems();
	}, 1);
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
	if (records.length > 0) {
		records.map((record) => {
			const gridcfg = assetfield;

			return gridcfg;
		});
	}
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

			// if (record.Taxable && record.TaxAmount && record.TaxAmount != 0) {
			// 	calc.TaxAmount += record.TaxAmount;
			// 	taxamount = record.TaxAmount;
			// }

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

onMounted(async () => {
	if (props.salesPriceBook != "") {
		data.SalesPriceBook = await GetSalesPriceBook(props.salesPriceBook);
	}
});

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

function lookupSpecsPayloadBuilder(search, config, value) {
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

	return qp;
}

function lookupPayloadBuilder(search, select, value, item) {
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
</script>
