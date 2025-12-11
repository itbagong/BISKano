<template>
  <data-list
    ref="listControl"
    hide-title
    no-gap
    grid-hide-search
    :grid-editor="!readOnly || status === 'SUBMITTED'"
    :grid-hide-delete="readOnly"
    :grid-hide-new="readOnly"
    grid-hide-sort
    grid-hide-refresh
    grid-hide-detail
    grid-hide-select
    grid-no-confirm-delete
    init-app-mode="grid"
    grid-mode="grid"
    :grid-config="'/fico/journal/line/gridconfig'"
    :grid-fields="['Account','Amount','Text','Critical']"
    new-record-type="grid"
    grid-auto-commit-line
    @grid-row-add="onGridNewRecord"
    @grid-row-delete="onGridRowDelete"
    @alter-grid-config="onAlterGridConfig"
    @grid-row-field-changed="onGridRowFieldChanged"
  >
    <template #grid_Account="{item}">
      <AccountSelector
        v-model="item.OffsetAccount"
        :read-only="readOnly"
        :items-type=" ['COA', 'EXP']"
      ></AccountSelector>
    </template>
    <template #grid_Amount="{item}">
      <template v-if="status === 'SUBMITTED'">{{item.Amount}}</template>
    </template>
    <template #grid_Text="{item}">
      <template v-if="status === 'SUBMITTED'">{{item.Text}}</template>
    </template>
    <template #grid_Critical="{}">
      <template v-if="status === 'SUBMITTED'">&nbsp;</template>
    </template>
  </data-list>
</template>
<script setup>
import { reactive, ref, watch } from "vue";
import { DataList, util, SInput } from "suimjs";
import AccountSelector from "@/components/common/AccountSelector.vue";

const props = defineProps({ 
  modelValue: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false},
  journalType:{ type: String, default: "Cash In" },
  status: { type:String, default:""}
});
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
const listControl = ref(null)

function calcTotalAmount() {
  const totalAmount = data.records.reduce((total, e) => {
    return total + e.Amount;
  }, 0); 
  emit("calcTotalAmount", totalAmount);
}

function onGridNewRecord(r) {
  r.ID = "";
  r.Amount = 0;
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
    const arr = ['Account','Amount','Text','Critical','PriceEach','Qty'];
    
    const fields = config.fields.filter(e=> { return arr.indexOf(e.field) > -1})
    config.fields = fields
    setTimeout(()=>{
        updateItems();
    },500)
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

    if (name == "Amount") {
        util.nextTickN(2, () => {
            calcTotalAmount();
        });
    }
}
function updateItems() { 
     listControl.value.setGridRecords(data.records);
 
   
}

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>