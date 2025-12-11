<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-lines"
      :editor="['', 'DRAFT'].includes(props.item.Status)"
      hide-search
      hide-select
      hide-sort
      hide-delete-button
      hide-refresh-button
      :hide-new-button="
        ['SUBMITTED', 'READY', 'POSTED', 'REJECTED'].includes(props.item.Status)
      "
      :hide-detail="true"
      hide-paging
      hide-action
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
      @new-data="newLine"
      @delete-data="deleteDetail"
    >
      <template #item_LineNo="{ item }">
        <div class="grid grid-cols-1 gap-2" v-if="item.Parent">
          {{ item.LineNo }}
        </div>
        <div v-else></div>
      </template>
      <template #item_StepsTask="{ item }">
        <div class="grid grid-cols-1 gap-2" v-if="item.Parent">
          <s-input
            v-model="item.StepsTask"
            :readOnly="
              ['SUBMITTED', 'READY', 'POSTED', 'REJECTED'].includes(
                props.item.Status
              )
            "
          />
          <div
            class="grid grid-cols-2 gap-2 pb-4"
            v-if="
              !['SUBMITTED', 'READY', 'POSTED', 'REJECTED'].includes(
                props.item.Status
              )
            "
          >
            <s-button
              label="add detail"
              tooltip="add detail"
              class="btn_success"
              @click="addDetail(item)"
            />
            <s-button
              label="delete"
              tooltip="delete"
              class="btn_primary"
              @click="deleteDetail(item)"
            />
          </div>
        </div>
        <div v-else></div>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid, SButton } from "suimjs";
const axios = inject("axios");
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  item: { type: Object, default: () => {} },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  formMode: { type: String, default: () => "new" },
  activeFields: { type: Array, default: () => [] },
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

function newLine(r) {
  r = {};
  const noLine = data.value.filter((obj) => obj.Parent == true).length + 1;
  r.ID = util.uuid();
  r.Parent = true;
  r.LineNo = noLine;
  r.StepsTask = "";
  r.HazardAndRiskSteps = "";
  r.HazardCode = "";
  r.Recommendation = "";
  data.value.push(r);
  updateGridLines();
}

function addDetail(r) {
  let obj = {};
  obj.ID = util.uuid();
  obj.Parent = false;
  obj.LineNo = r.LineNo;
  obj.StepsTask = r.StepsTask;
  obj.HazardAndRiskSteps = "";
  obj.HazardCode = "";
  obj.Recommendation = "";
  data.value.push(obj);
  updateGridLines();
}

function deleteDetail(r) {
  data.value = data.value.filter((obj) => obj.LineNo !== r.LineNo);
  let no = 1;
  for (let i in data.value) {
    let obj = data.value[i];
    if (obj.Parent) {
      let LineNo = no++;
      obj.LineNo = LineNo;
    } else {
      obj.LineNo = data.value[parseInt(i) - 1].LineNo;
    }
  }
  updateGridLines();
}

function updateGridLines() {
  data.value.sort((a, b) => {
    if (a.LineNo !== b.LineNo) {
      return a.LineNo - b.LineNo;
    }
    return b.Parent - a.Parent;
  });
  listControl.value.setRecords(data.value);
  util.nextTickN(3, () => {
    formatRowSpan();
  });
}

function formatRowSpan() {
  const myTable = document.querySelectorAll(
    ".tb-lines .suim_table [name='grid_body']"
  );
  if (myTable.length > 0) {
    for (let i = 0, row; (row = myTable[0].rows[i]); i++) {
      const firstCell = row.cells[0];
      const secondCell = row.cells[1];
      firstCell.classList.remove("hidden");
      secondCell.classList.remove("hidden");
      if (data.value[i].Parent) {
        let lengthRow = data.value.filter(
          (o) => o.LineNo == data.value[i].LineNo
        );
        firstCell.rowSpan = lengthRow.length;
        secondCell.rowSpan = lengthRow.length;
      } else {
        firstCell.classList.add("hidden");
        secondCell.classList.add("hidden");
      }
    }
  }
}

function getDataValue() {
  return listControl.value.getRecords();
}
onMounted(() => {
  loadGridConfig(axios, props.gridConfig).then(
    (r) => {
      data.gridCfg = r;
      console.log("===========", r);
      updateGridLines();
    },
    (e) => util.showError(e)
  );
});
defineExpose({
  getDataValue,
});
</script>
<style>
.tb-line > div:nth-child(2) > div {
  overflow-x: auto;
  padding-bottom: 100px;
}
.tb-line > div:nth-child(2) > div > table {
  width: calc(100% + 40%) !important;
}
</style>
