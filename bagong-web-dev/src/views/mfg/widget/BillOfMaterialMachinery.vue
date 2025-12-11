<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-machinery grid-line-items"
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
  record.MachineCode = "";
  record.MachineName = "";
  record.StandartHour = 0;
  record.RatePerHour = 0;
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
      const colms = [
        {
          field: "MachineCode",
          label: "Machine Code",
          kind: "text",
        },
        {
          field: "MachineName",
          label: "Machine Name",
          kind: "text",
        },
        {
          field: "StandartHour",
          label: "Standart Hour",
          kind: "number",
        },
        {
          field: "RatePerHour",
          label: "Rate PerHour",
          kind: "number",
        },
      ];
      let tbLine = [
        "MachineCode",
        "MachineName",
        "StandartHour",
        "RatePerHour",
      ];
      let addColms = [];
      for (let index = 0; index < colms.length; index++) {
        addColms.push({
          field: colms[index].field,
          kind: colms[index].kind,
          label: colms[index].label,
          readType: "show",
          labelField: "",
          input: {
            field: colms[index].field,
            label: colms[index].label,
            hint: "",
            hide: false,
            placeHolder: colms[index].label,
            kind: colms[index].kind,
          },
        });
      }
      const _fields = [...r.fields, ...addColms].filter((o) => {
        if (["StandartHour", "StandartHour"].includes(o.field)) {
          o.width = "150px";
        } else {
          o.width = "300px";
        }
        o.idx = tbLine.indexOf(o.field);
        return tbLine.includes(o.field);
      });
      data.gridCfg = {
        ...r,
        fields: _fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
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
