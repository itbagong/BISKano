<template>
    <s-card
		hide-footer
		class="w-full bg-white card"
		:no-gap="true"
		:hide-title="true"
	>
	<div
		v-if="data.showPreview == false"
	>
		<s-form
            v-if="data.frmCfg1 && data.frmCfg1.setting"
			ref="formCtl"
			:config="data.frmCfg1"
            v-model="props.modelValue"
            mode="new"
			class="pt-2"
			hide-submit
            keep-label
			@cancelForm="cancelForm"
		>
			<template #buttons_1="{item, config}">
				<s-button icon='eye-outline' label='Preview' class='btn_default' @click="() => (data.showPreview = true)"/>
				<s-button icon='key-change' label='Print' class='btn_default'/>
			</template>
		</s-form>

		<s-form
            v-if="data.frmCfg2 && data.frmCfg2.setting"
			ref="formCtl2"
			:config="data.frmCfg2"
			v-model="props.modelValue"
			class="pt-2"
			hide-buttons
			hide-submit
			hide-cancel
            keep-label
		>
			<template #input_CompanySender="{item}">
				<s-input
					v-model="props.itemSalesOrder.CompanyID"
					kind="text"
					disabled
					keep-label
					label="Company"
					class="w-full"
				></s-input>
			</template>
			<template #input_CompanyRecipient="{item}">
				<s-input
					v-model="props.itemSalesOrder.CompanyID"
					kind="text"
					disabled
					keep-label
					label="Company"
					class="w-full"
				></s-input>
			</template>
			<template #input_AddressRecipient="{item}">
				<s-input
					kind="text"
					v-model="props.itemCustomer.CustomerAddress"
					keep-label
					label="Address"
					class="w-full"
				></s-input>
			</template>
			<template #input_PhoneRecipient="{item}">
				<s-input
					kind="text"
					v-model="props.itemCustomer.CustomerPhone"
					keep-label
					label="Phone"
					class="w-full"
				></s-input>
			</template>
		</s-form>

		<s-form
            v-if="data.frmCfg3 && data.frmCfg3.setting"
			ref="formCtl3"
			:config="data.frmCfg3"
			v-model="props.modelValue"
			class="pt-2"
			hide-buttons
			hide-submit
			hide-cancel
            keep-label
		>
		</s-form>
		
		<div class="title section_title">Kelengkapan</div>
		<div class="flex flex-row suim_form">
			<div class="flex basis-1/4 flex-col my-4">
				<div class="flex basis-1/4" v-for="(item, idx) in props.modelValue.collection.section1" :key="idx">
					<s-input
						kind="checkbox"
						class="w-full"
						v-model="props.modelValue.collection.section1[idx].value"
						:label="item.key"
					/>
				</div>
			</div>
			<div class="flex basis-1/4 flex-col my-4">
				<div class="flex basis-1/4" v-for="(item, idx) in props.modelValue.collection.section2" :key="idx">
					<s-input
						kind="checkbox"
						class="w-full"
						v-model="props.modelValue.collection.section2[idx].value"
						:label="item.key"
					/>
				</div>
			</div>
			<div class="flex basis-1/4 flex-col my-4">
				<div class="flex basis-1/4" v-for="(item, idx) in props.modelValue.collection.section3" :key="idx">
					<s-input
						kind="checkbox"
						class="w-full"
						v-model="props.modelValue.collection.section3[idx].value"
						:label="item.key"
					/>
				</div>
			</div>
			<div class="flex basis-1/4 flex-col my-4">
				<div class="flex basis-1/4" v-for="(item, idx) in props.modelValue.collection.section4" :key="idx">
					<s-input
						kind="checkbox"
						class="w-full"
						v-model="props.modelValue.collection.section4[idx].value"
						:label="item.key"
					/>
				</div>
			</div>
		</div>

		<s-form
            v-if="data.frmCfg5 && data.frmCfg5.setting"
			ref="formCtl5"
			:config="data.frmCfg5"
			v-model="props.modelValue"
			class="pt-2"
			hide-buttons
			hide-submit
			hide-cancel
            keep-label
		>
		</s-form>
	</div>
		<div class="pt-2"
			v-else
		>
		<s-button icon='rewind' label='Back' class='btn_warning back_btn origin-top-right' @click="() => (data.showPreview = false)"/>
		<br />
		<div class="container border-solid flex mb-1">
			<div class="flex basis-2/3" name="logo-bagong">
				<img src="@/assets/img/bagong.png" width="250"/>
			</div>
			<div class="flex basis-1/3">
				<div name="address">
				<p>
					Jl. Panglima Sudirman 8 Kepanjen - Malang <br />
					Jawa Timur 65163:Indonesia <br />
					Ph: +62 341 395524, +62 341 393382 <br />
					Fax: +62 341 395724
				</p>
				</div>
			</div>
		</div>
		<div class="border-solid border-8 border-black-60">
			<!-- <div class="flex basis-3/3"> -->
				<h2 class="text-center">SURAT JALAN / BERITA ACARA SERAH TERIMA BARANG</h2>
				<p>
					SJ / BAST NO : {{props.modelValue.BastNO}}<br>
					Tanggal BAST : {{props.modelValue.BastDate}}<br>
					Usage Date:  {{props.modelValue.UsageDate}}<br>
					Project Site : <br>
				</p>
			<!-- </div> -->
		</div>
		<div class="flex flex-row border-solid border-8 border-black-60">
			<div class="basis-1/4 border-solid border-2">O1</div>
			<div class="basis-1/4 border-solid border-2">Data Pengiriman</div>
			<div class="basis-1/4 border-solid border-2">02</div>
			<div class="basis-1/4 border-solid border-2">Data Penerima</div>
		</div>
		<div class="flex flex-row border-solid border-8 border-black-60">
			<div class="basis-1/2 border-solid border-2">
				<p>
					Lokasi      	: {{props.modelValue.LocationSender}}<br>
					Nama        	: {{props.modelValue.NameSender}}<br>
					Perusahaan  	: {{props.modelValue.CompanySender}}<br>
					Alamat			: {{props.modelValue.AddressSender}}
				</p>
			</div>
			<div class="basis-1/2 border-solid border-2">
				<p>
					Lokasi      	: {{props.modelValue.LocationRecipient}}<br>
					Nama        	: {{props.modelValue.NameRecipient}}<br>
					Perusahaan  	: {{props.modelValue.CompanyRecipient}}<br>
					Alamat/NoTelp	: {{props.modelValue.AddressRecipient}}
				</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/4 border-solid border-2">
				<p>O3</p>
			</div>
			<div class="basis-1/2 text-center">
				<p>BARANG YANG DIKIRIM HARUS DALAM KONDISI LENGKAP DAN SIAP / READY FOR USE</p>
			</div>
		</div>
		<div class="flex flex-row border-solid border-8 border-black-60">
			<span>&nbsp;</span>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/4 border-solid border-2">
				<p>O4</p>
			</div>
			<div class="basis-1/2 text-center">
				<p>SHIPMENT INFORMATION / DATA UNIT / BARANG</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="flex basis-1/2">
				<p>
					Model / Jenis Barang: {{props.modelValue.ItemModel}}<br>
					Nomor Polisi: {{props.modelValue.PoliceNumber}}<br>
					Nomor Rangka: {{props.modelValue.FrameNumber}}<br>
					Nomor Mesin: {{props.modelValue.MachineNumber}}<br>
					Tahun: {{props.modelValue.ProductionYear}}
				</p>
			</div>
			<div class="flex basis-1/2">
				<p>
					SMU / Km:   {{props.modelValue.SMU}}&nbsp;&nbsp; City: {{props.modelValue.City}} <br>
					Fuel Status: {{props.modelValue.FuelStatus}}% <br>
					Moda Pengiriman: {{props.modelValue.ModaPengiriman}}
				</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/4 border-solid border-2">
				<p>O5</p>
			</div>
			<div class="basis-1/2 text-center">
				<p>KELENGKAPAN</p>
			</div>
		</div>
		<div class="border-solid border-8 border-black-60">
			<div class="flex flex-row suim_form">
				<div class="flex basis-1/4 flex-col my-4">
					<div class="flex basis-1/4" v-for="(item, idx) in props.modelValue.collection.section1" :key="idx">
						<s-input
							kind="checkbox"
							class="w-full"
							:disabled="true"
							v-model="props.modelValue.collection.section1[idx].value"
							:label="item.key"
						/>
					</div>
				</div>
				<div class="flex basis-1/4 flex-col my-4">
					<div class="flex basis-1/4" v-for="(item, idx) in props.modelValue.collection.section2" :key="idx">
						<s-input
							kind="checkbox"
							class="w-full"
							:disabled="true"
							v-model="props.modelValue.collection.section2[idx].value"
							:label="item.key"
						/>
					</div>
				</div>
				<div class="flex basis-1/4 flex-col my-4">
					<div class="flex basis-1/4" v-for="(item, idx) in props.modelValue.collection.section3" :key="idx">
						<s-input
							kind="checkbox"
							class="w-full"
							:disabled="true"
							v-model="props.modelValue.collection.section3[idx].value"
							:label="item.key"
						/>
					</div>
				</div>
				<div class="flex basis-1/4 flex-col my-4">
					<div class="flex basis-1/4" v-for="(item, idx) in props.modelValue.collection.section4" :key="idx">
						<s-input
							kind="checkbox"
							class="w-full"
							:disabled="true"
							v-model="props.modelValue.collection.section4[idx].value"
							:label="item.key"
						/>
					</div>
				</div>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/4 border-solid border-2">
				<p>O6</p>
			</div>
			<div class="basis-1/2 text-center">
				<p>KOLOM UNTUK PENGIRIMAN BARANG</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/1 border-solid border-2">
				<p>{{props.modelValue.DeliveryDetail}}</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/4 border-solid border-2">
				<p>O7</p>
			</div>
			<div class="basis-1/2 text-center">
				<p>KOLOM PENERIMA</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/1 border-solid border-2">
				<p>{{props.modelValue.ReceiptDetail}}</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/4 border-solid border-2">
				<p>O8</p>
			</div>
			<div class="basis-1/2 text-center">
				<p>KOLOM KETERANGAN</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/1 border-solid border-2">
				<p>{{props.modelValue.Notes}}</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/3 border-solid border-2 text-center">
				<p>Pengiriman</p>
			</div>
			<div class="basis-1/3 border-solid border-2 text-center">
				<p>Pembawa</p>
			</div>
			<div class="basis-1/3 border-solid border-2 text-center">
				<p>Penerima</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60 h-36">
			<div class="basis-1/3 border-solid border-2 text-center">
				<p>&nbsp;</p>
			</div>
			<div class="basis-1/3 border-solid border-2 text-center">
				<p>&nbsp;</p>
			</div>
			<div class="basis-1/3 border-solid border-2 text-center">
				<p>&nbsp;</p>
			</div>
		</div>
		<div class="flex border-solid border-8 border-black-60">
			<div class="basis-1/1 text-center">
				<p>{{props.modelValue.Notes2}}</p>
			</div>
		</div>
	</div>
	</s-card>
</template>

<script setup>
import { onMounted, reactive, ref, inject } from "vue";
import { SCard,
	SForm,
	util,
	SInput,
	SModal,
	SButton, 
loadFormConfig } from "suimjs";
import {layoutStore} from "../../../stores/layout.js";
import {useRoute} from "vue-router";

layoutStore().name = "tenant";

const axios = inject("axios");

const formCtl = ref(null);

const props = defineProps({
	isEdited: {type: Boolean, default: () => false},
	modelValue: {type: Object, default: () => {}},
	itemSalesOrder: {type: Object, default: () => {}},
	itemCustomer: {type: Object, default: () => {}},
	assetId: {type: String, default: () => ""},
    formConfig1: {type: [String, Object], default: () => {}},
	formConfig2: {type: [String, Object], default: () => {}},
	formConfig3: {type: [String, Object], default: () => {}},
	formConfig5: {type: [String, Object], default: () => {}},
	formDefaultMode: {type: String, default: "edit"},
});

const emit = defineEmits({
  "update:modelValue": null,
  "update:cancelBast": null,
  recalc: null,
});

const data = reactive({
    frmCfg1: undefined,
	frmCfg2: undefined,
	frmCfg3: undefined,
	frmCfg5: undefined,
	showPreview: false,
    record: props.modelValue,
})

function cancelForm() {
	emit("update:cancelBast", props.modelValue, props.isEdited)
	// console.log("cancel form")
}

async function getAsset() {
	await axios.post("/bagong/asset/find?_id="+props.assetId).then((res) => {
		const response = res.data[0]

		props.modelValue.ItemModel = response.DetailUnit.UnitType
		props.modelValue.PoliceNumber = response.DetailUnit.PoliceNum
		props.modelValue.FrameNumber = response.DetailUnit.ChassisNum
		props.modelValue.MachineNumber = response.DetailUnit.MachineNum
		props.modelValue.ProductionYear = response.DetailUnit.ProductionYear
	})
}

function getFormConfig() {
    if (props.formConfig1 == undefined || props.formConfig1 == "") return;
    loadFormConfig(axios, props.formConfig1).then(
        (r) => {
            // console.log("res =>", r)
            data.frmCfg1 = r
        },
		(e) => util.showError(e))

	if (props.formConfig2 == undefined || props.formConfig2 == "") return;
	loadFormConfig(axios, props.formConfig2).then(
        (r) => {
            // console.log("res =>", r)
            data.frmCfg2 = r
        },
		(e) => util.showError(e))

	if (props.formConfig3 == undefined || props.formConfig3 == "") return;
	loadFormConfig(axios, props.formConfig3).then(
        (r) => {
            // console.log("res =>", r)
            data.frmCfg3 = r
        },
		(e) => util.showError(e))

	if (props.formConfig5 == undefined || props.formConfig5 == "") return;
	loadFormConfig(axios, props.formConfig5).then(
        (r) => {
            // console.log("res =>", r)
            data.frmCfg5 = r
        },
		(e) => util.showError(e))
	
}

function init() {
	getFormConfig()
	getAsset()
	// console.log("record =>", data.record)
}

onMounted(() => {
    init()
})
</script>