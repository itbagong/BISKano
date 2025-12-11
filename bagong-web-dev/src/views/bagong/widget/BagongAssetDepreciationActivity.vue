<template>
  <div class="title section_title">Depreciation Activity</div>
  <data-list
    ref="listControl"
    hide-title
    grid-hide-search
    grid-hide-sort
    grid-hide-refresh
    grid-hide-detail
    grid-hide-select
    grid-hide-delete
    grid-auto-commit-line
    new-record-type="grid"
    grid-editor-no-form
    grid-config="/bagong/asset/detail/depreciationactivity/gridconfig"
    :init-app-mode="data.appMode"
    :grid-mode="data.appMode"
    @alterGridConfig="onAlterGridConfig"
  >
    <template #grid_header_buttons>
      <s-button
        icon="import"
        class="btn_primary"
        label="Generate"
        :disabled="data.statusGenerate"
        @click="genDepreciationActivity"
      />
    </template>
    <template #grid_item_buttons_1="{ props, item }">
      <s-button
        icon="pencil"
        class="btn_secondary"
        label="Adjust"
        :disabled="item.JournalID !== ''"
        @click="
          () => {
            data.selectedRecord = item;
            util.nextTickN(2, () => {
              data.openAdjustment = true;
            });
          }
        "
      />
    </template>
  </data-list>
  <s-modal
    v-if="data.openAdjustment"
    title="Asset Adjustment"
    class="w-1/2"
    display
    ref="reject"
    @beforeHide="close"
    hideButtons
  >
    <div class="w-[350px]">
      <div class="py-3 mb-3">
        <s-input
          v-model="data.adjustmentAmount"
          label="Adjustment amount"
          class="w-full"
          :required="true"
        ></s-input>
      </div>
      <div class="mt-5">
        <s-button
          class="bg-primary text-white font-bold w-full flex justify-center"
          label="Submit"
          @click="onApplyAdjustment()"
        ></s-button>
      </div>
    </div>
  </s-modal>
</template>

<script setup>
import { reactive, onMounted, inject, watch, ref } from "vue";
import { DataList, SButton, SInput, util, SModal } from "suimjs";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  references: { type: Object, default: () => [] },
  depreciation: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  appMode: "grid",
  records: props.modelValue,
  selectedRecord: {},
  adjustmentAmount: 0,
  gridCfg: {},
  loadingGrid: false,
  statusGenerate: false,
  openAdjustment: false,
});

const listControl = ref(null);
const axios = inject("axios");

function genDepreciationActivity() {
  data.records = [];

  const acquisitionCost = props.references[3].Value;
  let url = "/bagong/asset/generate-depreciation";
  let payload = {
    AcquisitionCost: acquisitionCost,
    AssetDuration: props.depreciation.AssetDuration,
    DepreciationDate: props.depreciation.DepreciationDate,
    DepreciationPeriod: props.depreciation.DepreciationPeriod,
    ResidualAmount: props.depreciation.ResidualAmount,
  };
  axios.post(url, payload).then((r) => {
    try {
      data.records = r.data;
      updateItems();
    } catch (err) {
      util.showError(err);
    }
  });
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
  emit("update:modelValue", data.records);
}

function checkStatusGenerate(dataRef, dataDep) {
  data.statusGenerate =
    dataRef[3] == undefined ||
    dataRef[3].Value <= 0 ||
    dataDep.AssetDuration <= 0 ||
    dataDep.DepreciationPeriod.length <= 0 ||
    !dataDep.DepreciationDate;
}
function onAlterGridConfig(config) {
  setTimeout(() => {
    updateItems();
    checkStatusGenerate(props.references, props.depreciation);
  }, 500);
}
onMounted(() => {
  setTimeout(() => {
    checkStatusGenerate(props.references, props.depreciation);
  }, 2000);
});
function close() {
  data.openAdjustment = false;
  data.selectedRecord = {};
  data.adjustmentAmount = 0;
}
function onApplyAdjustment() {
  const adjustmentAmount = parseFloat(data.adjustmentAmount);

  if (!adjustmentAmount || isNaN(adjustmentAmount)) {
    util.showError("Please enter a valid adjustment amount.");
    return;
  }
  const selectedRecordIndex = data.records.findIndex(
    (record) => record._id === data.selectedRecord._id
  );
  if (selectedRecordIndex !== -1) {
    const startIndex = selectedRecordIndex + 1;
    data.records[selectedRecordIndex].AdjustmentAmount = adjustmentAmount;

    const remainingRecords = data.records.slice(startIndex);
    const duration = remainingRecords.length;

    const newDepreciationAmount =
      (adjustmentAmount - props.depreciation.ResidualAmount) / duration;

    let currentBookValue = adjustmentAmount;

    for (let i = startIndex; i < data.records.length; i++) {
      data.records[i].DepreciationAmount = newDepreciationAmount;
      data.records[i].NetBookValue = currentBookValue - newDepreciationAmount;
      currentBookValue = data.records[i].NetBookValue;
    }
    updateItems();

    close();
  } else {
    util.showError("Selected record not found.");
  }
}
watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

watch(
  () => [props.references, props.depreciation],
  (nv) => {
    const dataRef = nv[0];
    const dataDep = nv[1];
    checkStatusGenerate(dataRef, dataDep);
  },
  { deep: true }
);
</script>
