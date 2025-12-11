<template>
  <s-card hide-footer class="w-full bg-white suim_datalist">
    <data-list
      ref="listControl"
      hide-title
      no-gap
      :grid-editor="!readOnly"
      :grid-hide-delete="readOnly"
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-hide-detail
      grid-hide-select
      grid-no-confirm-delete
      init-app-mode="grid"
      grid-mode="grid"
      :grid-config="'/fico/journal/line/gridconfig'"
      :grid-fields="['Account', 'PaymentType', 'Amount']"
      new-record-type="grid"
      grid-auto-commit-line
      @grid-row-add="onGridNewRecord"
      @grid-row-delete="onGridRowDelete"
      @alter-grid-config="onAlterGridConfig"
      @grid-row-field-changed="onGridRowFieldChanged"
    >
      <template #grid_Amount="{ item }">{{
        util.formatMoney(item.Amount)
      }}</template>
      <template #grid_Account="{ item }">
        <AccountSelector
          v-model="item.Account"
          :items-type="
            journalType === 'CASH IN'
              ? ['LEDGERACCOUNT', 'CUSTOMER']
              : ['LEDGERACCOUNT', 'VENDOR', 'EXPENSE']
          "
          :read-only="readOnly && status!='DRAFT' && !(ledgerEditor && status=='READY')"
        ></AccountSelector>
      </template>
      <template
        #grid_item_buttons_1="{ prop, item }"
        v-if="journalType === 'CASH OUT'"
      >
        <s-button
          icon="note-plus-outline"
          name="apply"
          alt="apply"
          class="cursor-pointer hover:text-success mt-[-3px]"
          width="16"
          @click="onApply(item)"
          v-if="
            item.PaymentType &&
            ['PTY002', 'PTY003'].includes(item.PaymentType) &&
            item.ChequeGiroID == ''
          "
        />
      </template>
      <template #grid_PaymentType="{ item }">
        <s-input
          v-model="item.PaymentType"
          hide-label
          use-list
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTY"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="item.ChequeGiroID = ''"
        />
      </template>
    </data-list>

    <cheque
      ref="applyCheque"
      :modal-apply="data.modalApply"
      :journal="journal"
      :lines="data.records"
    />
  </s-card>
</template>
<script setup>
import { reactive, ref, watch, onMounted, inject } from "vue";
import {
  DataList,
  util,
  SModal,
  SForm,
  SButton,
  SInput,
  createFormConfig,
} from "suimjs";
import AccountSelector from "@/components/common/AccountSelector.vue";
import Cheque from "./ChequeGiro.vue";

const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false },
  journalType: { type: String, default: "CASH IN" },
  journal: { type: Object, default: {} },
});

const ledgerEditor = authStore().getRBAC("EditLedgerBeforePost").canSpecial1;

const emit = defineEmits({
  "update:modelValue": null,
  calcTotalAmount: null,
});

const data = reactive({
  appMode: "grid",
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
});
const listControl = ref(null);
const applyCheque = ref(null);

function calcTotalAmount() {
  const totalAmount = data.records.reduce((total, e) => {
    return total + e.Amount;
  }, 0);
  emit("calcTotalAmount", totalAmount);
}

function onGridNewRecord(r) {
  r.ID = "";
  r.Amount = 0;
  r.LineNo = data.records.length;
  // r.OffsetAccount = {
  //   AccountType": "CASHBANK"
  // }
  data.records.push(r);
  updateItems();
}

function onGridRowDelete(_, index) {
  const newRecords = data.records.filter((_, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  updateItems();
}

function onAlterGridConfig(config) {
  const isCashOut = props.journalType == "CASH OUT";
  const fieldReadOnly = ["ChequeGiroID"];
  let gridFields = ["Account", "PriceEach", "Qty", "Amount", "Text"];

  if (isCashOut)
    gridFields = [...gridFields, ...["ChequeGiroID", "PaymentType"]];

  const fields = config.fields.filter((e) => {
    if (fieldReadOnly.includes(e.field)) {
      e.input.readOnly = true;
    }
    return gridFields.indexOf(e.field) > -1;
  });

  config.fields = fields;
  setTimeout(() => {
    updateItems();
  }, 500);
}
function onGridRowFieldChanged(name, v1, v2, old, record) {
  if (name == "Qty") {
    record.Amount = v1 * record.PriceEach;
  }
  if (name == "PriceEach") {
    record.Amount = v1 * record.Qty;
  }
  listControl.value.setGridRecord(
    record,
    listControl.value.getGridCurrentIndex()
  );

  updateItems();

  // if (name == "Amount") {
  util.nextTickN(2, () => {
    calcTotalAmount();
  });
  // }
}
function updateItems() {
  listControl.value.setGridRecords(data.records);
}

function onApply(record) {
  applyCheque.value.onApplyCheque(record);
}

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
