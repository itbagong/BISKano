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
					<div>Nomer : {{ props.modelValue.SalesOrderNo }}</div>
					<div>Perihal : {{ props.modelValue.Name }}</div>
					<div class="grid grid-cols-3 gap-3">
						<div />
						<div />
						<div>
							Malang,
							{{ moment(props.modelValue.SalesOrderDate).format("DD-MM-YYYY") }}
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
						<div>{{ data.customer ? data.customer.PersonalContact : "" }}</div>
					</div>
					<div class="grid grid-cols-3 gap-3">
						<div />
						<div />
						<div>{{ data.customer ? data.customer.Address : "" }}</div>
					</div>
					<div>Dengan Hormat</div>
					<div>
						Bersama dengan ini kami sampaikan
						<u>{{ props.modelValue.Name }}</u> dengan rincian sebagai berikut:
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
									:lookup-labels="['_id', 'Name']"
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
									:lookup-labels="['_id', 'Name']"
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
					<div>
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
											util.formatMoney(props.modelValue.TaxAmount, {
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
import {inject, reactive, ref, onMounted, watch, defineEmits} from "vue";
import moment from "moment";
import {SCard, SGrid, loadGridConfig, util, SButton, SInput} from "suimjs";

const props = defineProps({
	modelValue: {type: Object, default: () => {}},
	gridConfig: {type: [String, Object], default: () => {}},
});

const data = reactive({
	listCfg: {},
	customer: {
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
});

const emit = defineEmits({
	cancelClick: null,
});

const axios = inject("axios");
const gridCtl = ref(SGrid);

async function onDownloadPDF() {
	try {
		const data = await axios.post(`/sdp/salesorder/print-pdf`, {
			_id: props.modelValue._id,
			public_url: import.meta.env.VITE_API_URL,
		});

		goto(`/asset/view?id=${data.data}`);
	} catch (error) {
		util.showError(error);
	}
}

function onCancelForm() {
	emit("cancelClick", props.modelValue);
}

watch(
	() => props.modelValue.Lines,
	(nv) => {
		gridCtl.value.setRecords(props.modelValue.Lines);
	}
);
watch(
	() => props.modelValue.Customer,
	(nv) => {
		data.customer = nv;
	}
);

// function romanize(num) {
// 	if (isNaN(num)) return NaN;
// 	var digits = String(+num).split(""),
// 		key = [
// 			"",
// 			"C",
// 			"CC",
// 			"CCC",
// 			"CD",
// 			"D",
// 			"DC",
// 			"DCC",
// 			"DCCC",
// 			"CM",
// 			"",
// 			"X",
// 			"XX",
// 			"XXX",
// 			"XL",
// 			"L",
// 			"LX",
// 			"LXX",
// 			"LXXX",
// 			"XC",
// 			"",
// 			"I",
// 			"II",
// 			"III",
// 			"IV",
// 			"V",
// 			"VI",
// 			"VII",
// 			"VIII",
// 			"IX",
// 		],
// 		roman = "",
// 		i = 3;
// 	while (i--)
// 		roman = (key[+Number(digits.pop() || "0") + i * 10] || "") + roman;
// 	return Array(+digits.join("") + 1).join("M") + roman;
// }

function loadgridconfig() {
	loadGridConfig(axios, props.gridConfig).then(
		(r) => {
			data.listCfg = r;
		},
		(e) => util.showError(e)
	);
}

onMounted(() => {
	loadgridconfig();
});

function goto(destination) {
	window.location.href = import.meta.env.VITE_API_URL + destination;
}
</script>
