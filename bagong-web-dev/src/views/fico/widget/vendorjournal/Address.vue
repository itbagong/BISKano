<template>
  <div> 
    <s-form ref="form" :config="data.config" v-model="data.value" keep-label :mode="readOnly?'view':'edit'">
      <template #buttons>&nbsp;</template>
      <template #input_BankName="{ item, config }">
        <s-input
          v-model="item.BankName"
          keep-label
          :label="config.label"
          use-list
          :items="props.itemsBank"
          @change="onChangeBankName"
          v-if="props.itemsBank.length > 0"
        ></s-input>
      </template>
    </s-form>
  </div>
</template>

<script setup>
import { reactive, onMounted, inject, watch, ref } from "vue";
import { SForm, SInput, loadFormConfig, util } from "suimjs";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  bankDetail: { type: Array, default: () => [] },
  itemsBank: { type: Array, default: () => [] },
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

function onChangeBankName(name, v1) {
  let f = props.bankDetail.find((o) => o.BankName == v1);
  if (f) {
    data.value.BankAccountNo = f.BankAccountNo;
    data.value.BankAccountName = f.BankAccountName;
  }
}

onMounted(() => {
  loadFormConfig(axios, "/fico/vendorjournal/address/formconfig").then(
    (r) => {
      data.config = r;
    },
    (e) => util.showError(e)
  );
});

watch(
  () => props.itemsBank,
  (nv) => {
    if (nv.length > 0) {
      data.value.BankName = nv[0];
      let f = props.bankDetail.find((o) => o.BankName == nv[0]);
      if (f) {
        data.value.BankAccountNo = f.BankAccountNo;
        data.value.BankAccountName = f.BankAccountName;
      }
    } else {
      data.value.BankName = "";
      data.value.BankAccountNo = "";
      data.value.BankAccountName = "";
    }
  }
);
</script>
