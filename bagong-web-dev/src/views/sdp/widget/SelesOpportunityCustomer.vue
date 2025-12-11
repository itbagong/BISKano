<template>
    <div class="flex gap-2 filter-line-card">
        <div class="filter-line-box">
            <s-input
                v-model="customerDetail"
                label="Customer"
                use-list
                :lookup-url="`/tenant/customer/find`"
                lookup-key="_id"
                :lookup-labels="['_id', 'Name']"
                :lookup-searchs="['_id', 'Name']"
                class="w-100"
                :key="props.modelValue"
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
}
.w-100 {
    width: 100%;
}
</style>

<script setup>
import {SInput} from "suimjs";
import {inject, onMounted, ref, reactive, computed, watch} from "vue";
import moment from 'moment'
const axios = inject("axios");

const props = defineProps({
	modelValue: {type: Array, default: () => []},
});

const emit = defineEmits({
	"update:modelValue": null,
	recalc: null,
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
});

const customerDetail = computed({
    get() {
        const id = props.modelValue;
        getDataCustomerDetail(id);
        return id;
    },
    set(v) {
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
    } catch (error) {
        // util.showError(error);
    }
    // const dt = res.data.Detail;
    // return dt;
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
// bagong/customer/get
// ["M-136"]
</script>