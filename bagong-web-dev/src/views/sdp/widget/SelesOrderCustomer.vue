<template>
    <div class="flex gap-2 filter-line-card">
        <div class="filter-line-box">
            <s-input
                v-if="props.quotation !== ''"
                :key="data.customerName"
                label="Customer"
                v-model="data.customerName"
                read-only
                class="w-100"
            ></s-input>
            <s-input
                v-else
                :key="data.customerID"
                label="Customer"
                v-model="customerDetail"
                use-list
                :lookup-url="`/tenant/customer/find`"
                lookup-key="_id"
                :lookup-labels="['Name']"
                :lookup-searchs="['_id', 'Name']"
                class="w-100"
                :disabled="props.quotation !== ''"
                :lookup-payload-builder="
                    (search) =>
                        lookupPayloadBuilder(search, customerDetail)
                "
            ></s-input>
        </div>
        <div class="filter-line-box">
            <s-input
                ref="refAddress"
                label="Address"
                v-model="data.customer.Address"
                disabled
                :multi-row="2"
                class="w-100"
            ></s-input>
        </div>
        <div class="filter-line-box">
            <s-input
                ref="refCity"
                label="City"
                v-model="data.customer.City"
                disabled
                class="w-100 mr-2"
            ></s-input>
            <s-input
                ref="refProvince"
                label="Province"
                v-model="data.customer.Province"
                disabled
                class="w-100"
            ></s-input>
        </div>
        <div class="filter-line-box">
            <s-input
                ref="refCountry"
                label="Country"
                v-model="data.customer.Country"
                disabled
                class="w-100 mr-2"
            ></s-input>
            <s-input
                ref="refZipcode"
                label="Zipcode"
                v-model="data.customer.Zipcode"
                disabled
                class="w-100"
            ></s-input>
        </div>

        <div class="title section_title w-100">Primary Contact Information</div>
        <div class="filter-line-box">
            <s-input
                ref="refPersonalContact"
                label="Name"
                v-model="data.customer.PersonalContact"
                disabled
                class="w-100 mr-2"
            ></s-input>
            <s-input
                ref="refEmail"
                label="Email"
                v-model="data.customer.Email"
                disabled
                class="w-100"
            ></s-input>
        </div>
        <div class="filter-line-box">
            <s-input
                ref="refBusinessPhoneNo"
                label="Business Phone No."
                v-model="data.customer.BusinessPhoneNo"
                disabled
                class="w-100 mr-2"
            ></s-input>
            <s-input
                ref="refMobilePhoneNo"
                label="Mobile Phone No."
                v-model="data.customer.MobilePhoneNo"
                disabled
                class="w-100"
            ></s-input>
        </div>

        <div class="title section_title w-100">Delivery Address</div>
        <div class="filter-line-box">
            <s-input
                ref="refDeliveryAddress"
                label="Address"
                v-model="data.customer.DeliveryAddress"
                disabled
                :multi-row="2"
                class="w-100"
            ></s-input>
        </div>
        <div class="filter-line-box">
            <s-input
                ref="refDeliveryCity"
                label="City"
                v-model="data.customer.DeliveryCity"
                disabled
                class="w-100 mr-2"
            ></s-input>
            <s-input
                ref="refDeliveryCountry"
                label="Country"
                v-model="data.customer.DeliveryCountry"
                disabled
                class="w-100"
            ></s-input>
        </div>
        <div class="filter-line-box">
            <s-input
                ref="refDeliveryProvince"
                label="Province"
                v-model="data.customer.DeliveryProvince"
                disabled
                class="w-100 mr-2"
            ></s-input>
            <s-input
                ref="refDeliveryZipcode"
                label="Zipcode"
                v-model="data.customer.DeliveryZipcode"
                disabled
                class="w-100"
            ></s-input>
        </div>
    </div>
</template>

<style scoped>
.filter-line-card {
    width: 100%;
    display: block;
}
.filter-line-box {
    margin-bottom: 1rem;
    display: flex;
    width: 100% !important;
}
.w-100 {
    width: 100%;
}
</style>

<script setup>
import {DataList, util, loadGridConfig, SGrid, SInput} from "suimjs";
import {inject, onMounted, ref, reactive, computed, watch} from "vue";
import moment from 'moment'
const axios = inject("axios");

const props = defineProps({
	modelValue: {type: Array, default: () => []},
    quotation: {type: String, default: ""},
    formMode: {type: String, default: ""},
});

const emit = defineEmits({
	"update:modelValue": null,
	change: null,
});

const data = reactive({
	records: props.modelValue,
    customer: {
        Address: "",
        City: "",
        Province: "",
        Country: "",
        Zipcode: "",
        //Primary Contact Information
        PersonalContact: "",
        Email: "",
        BusinessPhoneNo: "",
        MobilePhoneNo: "",
        //Delivery Address
        DeliveryAddress: "",
        DeliveryCity: "",
        DeliveryCountry: "",
        DeliveryProvince:"",
        DeliveryZipcode: "",
    },
    customerID: "",
    customerName: "",
});

const customerDetail = computed({
    get() {
        const id = props.modelValue;
        getDataCustomerDetail(id);
        return id;
    },
    set(v) {
        emit("change", v)
        emit("update:modelValue", v);
    },
});

watch(
  () => customerDetail,
  async (nv) => {
    getDataCustomerDetail(nv);
    // console.log(res)
    // data.customer = res;
  },
  { deep: true }
);

async function getDataCustomerDetail(params) {
    try {
        const res = await axios.post(`/bagong/customer/get`, [params]);
        data.customer = res.data.Detail;
        data.customerID = res.data._id;
        data.customerName = res.data.Name
        // listControl.value.refreshForm();
    } catch (error) {
        // util.showError(error);
    }
}

function lookupPayloadBuilder(search, value) {
	const qp = {};
	if (search != "") data.filterTxt = search;
	qp.Take = 20;
	qp.Sort = ["_id"];
	qp.Select = ["_id", "Name"];
	//setting search
    if (search !== '' && search !== null) {
        qp.Where = { 
            Op: "$or",
            items: [
                { Field: "_id", Op: "$contains", Value: [search] },
                { Field: "Name", Op:"$contains", Value: [search] }
            ]
        }
    }
    if ((value !== '' && value !== null && value.length !== 0 ) && (search == '' || search == null)) {
        qp.Where = { 
            Op: "$or",
            items: [
                { Field: "_id", Op: "$contains", Value: [value] },
                { Field: "Name", Op:"$contains", Value: [value] }
            ]
        }
    }

	return qp;
}
</script>