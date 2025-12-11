<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full"
      :editor="true"
      hide-search
      hide-select
      hide-sort
      :hide-new-button="false"
      :hide-delete-button="false"
      hide-refresh-button
      :hide-detail="true"
      :hide-action="false"
      @new-data="newRecord"
      @delete-data="onDelete"
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
    >
      <template #item_FromUnit="{ item, idx }">
        <s-input
          ref="refFromUnit"
          v-model="item.FromUnit"
          :disabled="true"
        ></s-input>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, watch, inject, computed, reactive, ref } from "vue";
import {
  loadFormConfig,
  loadGridConfig,
  createFormConfig,
  util,
  SInput,
  SGrid,
  SForm,
  SButton,
} from "suimjs";
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
  gridCfg: {},
});

function newRecord() {
  const record = {};
  record.ID = props.item._id;
  record.FromUnitID = props.item._id;
  record.FromUnit = props.item.Name;
  record.ToQty = 0;
  record.ToUnit = "";
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

function setDataValue(val) {
  data.value = val;
}

onMounted(() => {
  loadGridConfig(axios, props.gridConfig).then(
    (r) => {
      let tbLine = ["FromUnit", "ToQty", "ToUnit"];
      const _fields = r.fields.filter((o) => tbLine.includes(o.field));
      data.gridCfg = { ...r, fields: _fields };
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
          if (props.item._id) {
            listControl.value.setRecords(r.data.data);
          }
        }
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      if (listControl.value) {
        listControl.value.setLoading(false);
      }
    });
});
defineExpose({
  getDataValue,
  setDataValue,
});
</script>
