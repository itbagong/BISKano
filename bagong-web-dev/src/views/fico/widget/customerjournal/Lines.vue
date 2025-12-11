<template>
  <div class="flex flex-col gap-2">
    <data-list
      ref="listControl"
      title="Posting Profile"
      hide-title
      no-gap
      grid-editor
      grid-hide-select
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-no-confirm-delete
      gridAutoCommitLine
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      grid-config="/fico/customerjournal/line/gridconfig"
      form-config="/fico/customerjournal/line/formconfig"
      :form-fields="[
        'TagObjectID1',
        'TagObjectID2',
        'OffsetAccount',
        'Dimension',
      ]"
      :grid-fields="[
        'TagObjectID1',
        'TagObjectID2',
        'OffsetAccount',
        'Dimension',
        'Account'
      ]"
      @grid-row-add="newRecord"
      @form-field-change="onFormFieldChanged"
      @form-edit-data="openForm"
      @alter-grid-config="onAlterGridConfig"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="onGridRowFieldChanged"
      @grid-refreshed="gridRefreshed"
      @grid-row-save="onGridRowSave"
      @post-save="onFormPostSave"
      form-focus
    >
      <template #grid_Dimension="{ item }">
        <DimensionText :dimension="item.Dimension" />
      </template>
      <template #grid_TagObjectID1="{ item }">
        <AccountSelector v-model="item.TagObjectID1" hide-account-type hide-label />
      </template>
      <template #grid_TagObjectID2="{ item }">
        <AccountSelector v-model="item.TagObjectID2" hide-account-type hide-label />
      </template>
      <template #grid_OffsetAccount="{ item }">
        <AccountSelector v-model="item.OffsetAccount" hide-account-type hide-label />
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor v-model="item.Dimension"></dimension-editor>
      </template>
      <template #form_input_TagObjectID1="{ item }">
        <AccountSelector v-model="item.TagObjectID1" hide-account-type />
      </template>
      <template #form_input_TagObjectID2="{ item }">
        <AccountSelector v-model="item.TagObjectID2" hide-account-type />
      </template>
      <template #form_input_OffsetAccount="{ item }">
        <AccountSelector v-model="item.OffsetAccount" hide-account-type />
      </template>
      <template #grid_Account="{ item }">
        <AccountSelector v-model="item.Account" :items-type="[ 'LEDGERACCOUNT']"></AccountSelector>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { onMounted } from "vue";
import { reactive, ref, watch } from "vue";
import { DataList, SButton, util } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  records: props.modelValue.map((dt) => {
    dt.suimRecordChange = false;
    return dt;
  }),
});

function newRecord() {
  const record = {};
  record.suimRecordChange = false;
  record.Qty = 1;
  record.PriceEach = 1;
  record.Amount = 1;
  record.Taxable = true;
  record.TagObjectID1 = {
    AccountType: "ASSET",
    AccountID: "",
  };
  record.UnitID = "Each";
  //openForm(record);
  //return record;
  data.records.push(record);
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onAlterGridConfig(config) {
  const fields = config.fields.filter(e=> e.field != "OffsetAccount")
  fields.push({
     "field": "Account",
      "label":  "Account",
      "halign": "start",
      "valign": "start",
      "labelField": "", 
      "input": {
          "lookupUrl": ""
      }, 
      "readType": "show"
  }) 
  config.fields = fields;
}
function openForm(record) {
  updateJournalType(record.JournalTypeID);
}

function onGridRowDelete(record, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onFormPostSave(record, index) {
  record.suimRecordChange = false;
  if (listControl.value.getFormMode() == "new") {
    data.records.push(record);
  } else {
    data.records[index] = record;
  }
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onGridRowSave(record, index) {
  record.suimRecordChange = false;
  data.records[index] = record;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onFormFieldChanged(name, v1, v2, old, record) {
  if (name == "Qty") {
    record.Amount = v1 * record.PriceEach;
  }
  if (name == "PriceEach") {
    record.Amount = v1 * record.Qty;
  }
  if (name == "JournalTypeID") updateJournalType(v1);
  updateItems();
}

function updateJournalType(id) {
  // baca journaltypeid dan assisgn object tag

  // hide jia tag1==NONE

  // hide jika tag2==NONE
  util.nextTickN(1, () => {
    listControl.value.removeFormField("TagObjectID2");
  });
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
}

function updateItems() {
  const committedRecords = data.records.filter(
    (dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
  );
  emit("update:modelValue", committedRecords);
  emit("recalc");
}

function gridRefreshed() {
  listControl.value.removeGridField("TagObjectID2");
  listControl.value.setGridAttr("TagObjectID1", "label", "Asset");
  listControl.value.setGridRecords(data.records);
}

onMounted(() => {
  setTimeout(() => {}, 500);
});
</script>
