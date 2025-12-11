<template>
  <div>
    <s-form ref="form" :config="data.config" v-model="data.value" keep-label
      		:mode="readOnly?'view':'edit'">
      <template #buttons>&nbsp;</template>
    </s-form>
  </div>
</template>

<script setup>
import { reactive, onMounted, inject,ref } from "vue";
import { SForm, loadFormConfig, util } from "suimjs";
import helper from "@/scripts/helper.js";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  readOnly: { type: Boolean, default: false },
});


const form = ref(null);
const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  config: {},
  value: props.modelValue,
});

const axios = inject("axios");

onMounted(() => {
  loadFormConfig(axios, "/bagong/vendor/term/formconfig").then(
    (r) => {
      data.config = r; 
      util.nextTickN(2, () => {
        const cfgTax1 = form.value.getField("Taxes1")
 
        
        form.value.setFieldAttr("Taxes1","lookupPayloadBuilder", (search)=>{
          return helper.payloadBuilderTaxCodes(search, cfgTax1, data.value.Tax1, 'Purchase')
        });

        const cfgTax2 = form.value.getField("Taxes2")
        
        form.value.setFieldAttr("Taxes2","lookupPayloadBuilder", (search)=>{
          return helper.payloadBuilderTaxCodes(search, cfgTax2, data.value.Tax2, 'Purchase')
        });
      })
    },
    (e) => util.showError(e)
  );
});
</script>
