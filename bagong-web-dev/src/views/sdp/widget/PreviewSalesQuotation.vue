<template>
	<s-card
		class="w-full bg-white card"
		hide-footer
		:no-gap="false"
		:hide-title="true"
	>
		<div class="pt-2">
			<!-- tab header -->
			<div class="mb-2 flex header">
				<div class="grow form_button_top">
					<div class="flex items-center justify-end w-full">
						<div class="grow">&nbsp;</div>
						<slot name="buttons_1" :item="modelValue" />
						<slot name="buttons" :item="modelValue">
							<s-button
								icon="file-pdf-box"
								class=""
								label="PDF"
								@click="onDownloadPDF"
							/>

							<s-button
								icon="rewind"
								class="btn_warning back_btn"
								label="Back"
								@click="onCancelForm"
							/>
						</slot>
						<slot name="buttons_2" :item="modelValue" />
					</div>
				</div>
			</div>

			<div
				class="flex justify-center"
				style="font-family: 'Times New Roman', Times, serif; font-size: 11px"
			>
				<div class="max-w-[596px] border shadow-md p-5">
					<div v-if="props.modelValue.LetterHeadFirst" class="w-full">
						<img
							:src="'/v1/asset/view?id=' + props.modelValue.LetterHeadAsset"
							class="w-full"
						/>
					</div>
					<div>Nomer : {{ props.modelValue.QuotationNo }}</div>
					<div>Perihal : {{ props.modelValue.QuotationName }}</div>
					<div class="grid grid-cols-3 gap-3">
						<div />
						<div />
						<div>
							Malang,
							{{ moment(props.modelValue.QuotationDate).format("DD-MM-YYYY") }}
						</div>
					</div>
					<div class="grid grid-cols-3 gap-3">
						<div />
						<div />
						<div>Kepada Yth:</div>
					</div>
					<div class="grid grid-cols-3 gap-3">
						<div />
						<div />
						<div v-if="data.customerName && data.customerName.length > 0">
							{{ data.customerName }}
						</div>
						<div v-else>-</div>
					</div>
					<div class="grid grid-cols-3 gap-3">
						<div />
						<div />
						<div>{{ props.modelValue.PersonalContact }}</div>
					</div>
					<div class="grid grid-cols-3 gap-3">
						<div />
						<div />
						<div>
							{{
								`${props.modelValue.Address}, ${props.modelValue.City}, ${props.modelValue.Province}, ${props.modelValue.Country}, ${props.modelValue.Zipcode}`
							}}
						</div>
					</div>
					<div>Dengan Hormat</div>
					<div>
						Bersama dengan ini kami sampaikan
						<u>{{ props.modelValue.QuotationName }}</u> dengan rincian sebagai
						berikut:
					</div>
					<div>
						<s-grid
							ref="gridCtl"
							class="w-full"
							hide-select
							:config="data.listCfg"
							hide-search
							hide-control
							hide-detail
							hide-sort
							hide-delete-button
							hide-refresh-button
							hide-new
							hide-action
							hide-footer
							:grid-fields="['Item', 'Uom', 'Description']"
						>
							<template #item_Description="{item}: {item: any}">
								{{ item["Description"] }}
							</template>
							<template #item_Item="{item}: {item: any}">
								<s-input
									v-if="item.Asset !== '' && item.Asset !== undefined"
									ref="refItemAsset"
									hide-label
									v-model="item.Asset"
									use-list
									:lookup-url="`/tenant/asset/find`"
									lookup-key="_id"
									:lookup-labels="['Name']"
									:lookup-searchs="['_id', 'Name']"
									class="w-100"
									read-only
								></s-input>
								<s-input
									v-else
									ref="refItemItem"
									hide-label
									v-model="item.Item"
									use-list
									:lookup-url="`/tenant/item/find`"
									lookup-key="_id"
									:lookup-labels="['Name']"
									:lookup-searchs="['_id', 'Name']"
									class="w-100"
									read-only
								></s-input>
							</template>
							<template #item_Uom="{item}: {item: any}">
								<div>{{ item.UoM }}</div>
							</template>
						</s-grid>
					</div>
					<div v-if="props.modelValue.JournalType !== 'SalesQuotMining'">
						<div class="grid grid-cols-3 gap-3">
							<div />
							<div />
							<div>
								<div class="grid grid-cols-2 gap-2">
									<div>Subtotal Amount</div>
									<div>
										{{
											util.formatMoney(props.modelValue.SubTotalAmount, {
												decimal: 0,
											})
										}}
									</div>
								</div>
							</div>
						</div>
						<div class="grid grid-cols-3 gap-3">
							<div />
							<div />
							<div>
								<div class="grid grid-cols-2 gap-2">
									<div>Tax</div>
									<div>
										{{
											util.formatMoney(props.modelValue.TaxAmount, {decimal: 0})
										}}
									</div>
								</div>
							</div>
						</div>
						<div class="grid grid-cols-3 gap-3">
							<div />
							<div />
							<div>
								<div class="grid grid-cols-2 gap-2">
									<div>Discount</div>
									<div>
										{{
											util.formatMoney(props.modelValue.DiscountAmount, {
												decimal: 0,
											})
										}}
									</div>
								</div>
							</div>
						</div>
						<div class="grid grid-cols-3 gap-3">
							<div />
							<div />
							<div>
								<div class="grid grid-cols-2 gap-2">
									<div>Total Amount</div>
									<div>
										{{
											util.formatMoney(props.modelValue.TotalAmount, {
												decimal: 0,
											})
										}}
									</div>
								</div>
							</div>
						</div>
					</div>
					<div class="ql-editor" v-html="props.modelValue.Editor"></div>
					<div v-if="props.modelValue.FooterLastPage" class="w-full">
						<img
							:src="'/v1/asset/view?id=' + props.modelValue.FooterAsset"
							class="w-full"
						/>
					</div>
				</div>
			</div>
		</div>
	</s-card>
</template>

<script setup lang="ts">
import {inject, reactive, ref, onMounted, defineEmits} from "vue";
import moment from "moment";
import {SCard, SGrid, loadGridConfig, util, SButton, SInput} from "suimjs";

const props = defineProps({
	modelValue: {type: Object, default: () => {}},
	gridConfig: {type: [String, Object], default: () => {}},
});

const data = reactive({
	listCfg: {},
	customerName: '',
});

const emit = defineEmits({
	cancelClick: null,
});

const axios = inject("axios");
const gridCtl = ref(SGrid);

function onCancelForm() {
	emit("cancelClick", props.modelValue);
}

async function onDownloadPDF() {
	try {
		const data = await axios.post(`/sdp/salesquotation/print-pdf`, {
			_id: props.modelValue._id,
			public_url: import.meta.env.VITE_API_URL,
		});

		goto(`/asset/view?id=${data.data}`);
	} catch (error) {
		util.showError(error);
	}
}

function loadgridconfig() {
	loadGridConfig(axios, props.gridConfig).then(
		(r) => {
			if (props.modelValue.JournalType === 'SalesQuotMining') {
				r.fields = r.fields.filter((v) => !['Item', 'Qty', 'ContractPeriod', 'Amount'].includes(v.field));
			}
			data.listCfg = r;
		},
		(e) => util.showError(e)
	);
}

function goto(destination) {
	window.location.href = import.meta.env.VITE_API_URL + destination;
}

function getCustomer(id) {
	if (!id) return;

	const url = "/bagong/customer/get?_id=" + id;
	axios.post(url, [id]).then(
		(r) => {
			data.customerName = r.data.Name;
		},
		(e) => util.showError(e)
	);
}

onMounted(() => {
	util.nextTickN(2, () => {
		loadgridconfig();
		getCustomer(props.modelValue.Customer);
		gridCtl.value.setRecords(props.modelValue.Lines);
	})
});
</script>

<style>
.row_action {
	justify-content: center;
	align-items: center;
	gap: 2px;
}

table.suim_table > thead > tr > th {
	text-align: center !important;
}
</style>
