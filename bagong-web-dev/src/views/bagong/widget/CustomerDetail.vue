<template>
  <div>
    <s-form
      ref="form"
      :config="data.config"
      v-model="data.value"
      keep-label
      @field-change="handleFieldChange"
      :mode="readOnly ? 'view' : 'edit'"
    >
      <template #buttons>&nbsp;</template>
    </s-form>
  </div>
</template>

<script setup>
import { reactive, onMounted, inject, ref } from "vue";
import { SForm, loadFormConfig, util } from "suimjs";
import helper from "@/scripts/helper.js";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  readOnly: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  config: {},
  value: props.modelValue,
});

const form = ref(null);
const axios = inject("axios");

function handleFieldChange(name, v1, v2, old) {
  const value = data.value;
  if (name == "SameAsBillAddress") {
    if (v1) {
      value.NPWPAddress = value.Address;
      value.NPWPCity = value.City;
      value.NPWPProvince = value.Province;
      value.NPWPCountry = value.Country;
      value.NPWPPhone = value.Phone;
      value.NPWPZipcode = value.Zipcode;
      form.value.setFieldAttr("SameAsDeliverAddress", "hide", true);
    } else {
      value.NPWPAddress = "";
      value.NPWPCity = "";
      value.NPWPProvince = "";
      value.NPWPCountry = "";
      value.NPWPPhone = "";
      value.NPWPZipcode = "";
      form.value.setFieldAttr("SameAsDeliverAddress", "hide", false);
    }
  }
  if (name == "SameAsDeliverAddress") {
    if (v1) {
      value.NPWPAddress = value.DeliveryAddress;
      value.NPWPCity = value.DeliveryCity;
      value.NPWPProvince = value.DeliveryProvince;
      value.NPWPCountry = value.DeliveryCountry;
      value.NPWPPhone = value.DeliveryPhone;
      value.NPWPZipcode = value.DeliveryZipcode;
      form.value.setFieldAttr("SameAsBillAddress", "hide", true);
    } else {
      value.NPWPAddress = "";
      value.NPWPCity = "";
      value.NPWPProvince = "";
      value.NPWPCountry = "";
      value.NPWPPhone = "";
      value.NPWPZipcode = "";
      form.value.setFieldAttr("SameAsBillAddress", "hide", false);
    }
  }
  if (name == "SameAsBillAddr") {
    if (v1) {
      value.DeliveryAddress = value.Address;
      value.DeliveryCity = value.City;
      value.DeliveryProvince = value.Province;
      value.DeliveryCountry = value.Country;
      value.DeliveryPhone = value.Phone;
      value.DeliveryZipcode = value.Zipcode;
    } else {
      value.DeliveryAddress = "";
      value.DeliveryAddress = "";
      value.DeliveryCity = "";
      value.DeliveryProvince = "";
      value.DeliveryCountry = "";
      value.DeliveryPhone = "";
      value.DeliveryZipcode = "";
    }
  }
}

onMounted(() => {
  loadFormConfig(axios, "/bagong/customerdetail/formconfig").then(
    (r) => {
      data.config = r;

      util.nextTickN(2, () => {
        const cfgTax1 = form.value.getField("Tax1");

        form.value.setFieldAttr("Tax1", "lookupPayloadBuilder", (search) => {
          return helper.payloadBuilderTaxCodes(
            search,
            cfgTax1,
            data.value.Tax1,
            "Sales"
          );
        });

        const cfgTax2 = form.value.getField("Tax2");

        form.value.setFieldAttr("Tax2", "lookupPayloadBuilder", (search) => {
          return helper.payloadBuilderTaxCodes(
            search,
            cfgTax2,
            data.value.Tax2,
            "Sales"
          );
        });
      });
    },
    (e) => util.showError(e)
  );
});
</script>
