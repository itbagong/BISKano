<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-manpower grid-line-items"
      editor
      hide-search
      hide-sort
      :hide-new-button="false"
      hide-refresh-button
      :hide-detail="true"
      :hide-select="true"
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @delete-data="onDelete"
    >
      <template #item_ExpenseType="{ item, idx }">
        <s-input
          ref="refExpenseType"
          v-model="item.ExpenseType"
          use-list
          :lookup-url="`/tenant/expensetype/find`"
          lookup-key="_id"
          :lookup-labels="['LedgerAccountID', 'Name']"
          :lookup-searchs="['_id', 'LedgerAccountID', 'Name']"
          class="w-full"
        ></s-input>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid } from "suimjs";
const axios = inject("axios");
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
  itemID: { type: String, default: () => "" },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  readOnly: { type: Boolean, defaule: false },
  hideDetail: { type: Boolean, defaule: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  typeMoveIn: "line",
  gridCfg: {},
});

function newRecord() {
  const record = {};
  record.BoMID = props.itemID;
  record.ExpenseType = "";
  record.ActivityName = "";
  record.EmployeeQuantity = 0;
  record.StandartHour = 0;
  record.RatePerHour = 0;
  record.Description = "";
  listControl.value.setRecords([...listControl.value.getRecords(), record]);
}

function onDelete(record, index) {
  const newRecords = listControl.value.getRecords().filter((dt, idx) => {
    return idx != index;
  });
  listControl.value.setRecords(newRecords);
}

function getDataValue() {
  return listControl.value.getRecords();
}

function onSelectDataLine() {
  return listControl.value.getRecords().filter((el) => el.isSelected == true);
}

onMounted(() => {
  loadGridConfig(axios, props.gridConfig).then(
    (r) => {
      const _fields = r.fields.filter((o) => {
        if (
          ["EmployeeQuantity", "StandartHour", "RatePerHour"].includes(o.field)
        ) {
          o.width = "150px";
        } else {
          o.width = "300px";
        }
        return o;
      });
      data.gridCfg = {
        ...r,
        fields: _fields,
      };
      data.gridCfg = r;
    },
    (e) => util.showError(e)
  );
  axios
    .post(props.gridRead, {
      Skip: 0,
      Take: 0,
      Sort: ["_id"],
    })
    .then(
      (r) => {
        if (listControl.value) {
          listControl.value.setRecords(r.data.data);
        }
      },
      (e) => util.showError(e)
    );
});
defineExpose({
  getDataValue,
  onSelectDataLine,
});
</script>
