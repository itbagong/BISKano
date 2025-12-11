<template>
  <div>
    <data-list
      ref="listControl"
      title="Posting Profile"
      hide-title
      no-gap
      grid-editor
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-hide-detail
      grid-hide-select
      grid-no-confirm-delete
      grid-hide-new 
      grid-hide-delete
      init-app-mode="grid"
      grid-mode="grid"
      grid-config="/bagong/site_expense/gridconfig"
      new-record-type="grid"
      grid-auto-commit-line
      @grid-row-save="onGridRowSave" 
      @grid-row-field-changed="onGridRowFieldChanged"
      @alterGridConfig="alterGridConfig"
      :grid-fields="['Amount','Notes']"
    >
      <template #grid_Amount="{ item }">
        <div v-if="item.ExpenseCategory == 'Value'">
          <s-input
            :read-only="readOnly"
            kind="number"
            field="Amount"
            v-model="item.Amount"
            @change="(...args) => onChangeCalcAmountValue(item.ID, ...args)"
          ></s-input>
        </div>
        <div v-else class="flex gap-2 justify-end items-center w-full">
          <div class="max-w-[100px]">
            <s-input
              :read-only="readOnly"
              kind="number"
              field="Value"
              v-model="item.Value"
              @change="(...args) => onChangeCalcAmount(item.ID, ...args)"
            ></s-input>
          </div>x
          <div class="font-semibold text-[0.75rem]">{{ util.formatMoney(item.Amount) }}</div>=
          <div class="font-semibold text-[0.75rem]">{{ util.formatMoney(item.TotalAmount) }}</div>
        </div>
      </template> 
      <template #grid_Notes="{ item }">
        <template v-if="readOnly">{{item.Notes}}</template>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, inject } from "vue";
import { DataList, SInput, util, loadGridConfig } from "suimjs";

const axios = inject("axios");
const props = defineProps({
  siteEntryAssetID: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
  readOnly: {type: Boolean, default: false}
});

const emit = defineEmits({
  "update:modelValue": null,
  calcTotalAmount: null,
});

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
});

const listControl = ref(null);

function calcTotalAmount() {
  const totalAmount = data.records.reduce((total, e) => {
    return total + e.TotalAmount;
  }, 0);
 
  emit("calcTotalAmount", totalAmount);
}
function onGridRowSave(record, index) {
  record.suimRecordChange = false;
  data.records[index] = record;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
  console.log(name)
  if (name == "Value") {
    record.TotalAmount = v1 * record.Amount;
  }
  if (name == "Amount") {
    record.TotalAmount = v1 * record.Value;
  }

  listControl.value.setGridRecord(
    record,
    listControl.value.getGridCurrentIndex()
  );

  updateItems();
    if (
    ["Value", "Amount"].includes(name)
  ) {
    util.nextTickN(2, () => {
      calcTotalAmount();
    });
  }
}

onMounted(() => {
  
});

function alterGridConfig(config) { 
  console.log("alterGridConfig", config)
  config.fields = config.fields.reduce((acc, obj) => {
      if(["ExpenseTypeID","UnitID","Value"].includes(obj.field) || (props.readOnly === false && obj.field === "LineNo")) return acc
     
      if(props.readOnly === true) obj.input.readOnly =  true
    
      acc.push(obj)
      return acc
   },[]) 

  setTimeout(() => {
    updateItems();
  }, 1200)
}

defineExpose({
  updateItems,
});

function updateItems() {
  listControl.value.setGridRecords(data.records);
  util.nextTickN(2, () => {
    calcTotalAmount();
  });
}

function onChangeCalcAmountValue(id, name, v1) {
  let el = data.records.find((x) => {
    return x.ID == id;
  });
  el.TotalAmount = parseInt(v1); 
  util.nextTickN(2, () => {
    calcTotalAmount();
  });
}
function onChangeCalcAmount(id, name, v1) {
  let el = data.records.find((x) => {
    return x.ID == id;
  });
  el.TotalAmount = parseInt(el.Amount) * parseInt(v1);

  util.nextTickN(2, () => {
    calcTotalAmount();
  });
}
</script>
