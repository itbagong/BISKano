<template>
  <div>
    <loader v-if="data.loading" kind="skeleton" />
    <s-grid
      v-show="!data.loading"
      ref="gridref"
      class="w-full"
      editor
      :config="data.gridCfg"
      hide-select
      hide-search
      hide-detail
      hide-sort
      hide-refresh-button
      :hide-new-button="readOnly || controlAdd"
      :hide-delete-button="readOnly"
      :no-confirm-delete="!singleSave"
      @new-data="onGridNewRecord"
      @delete-data="onGridRowDelete"
      @rowFieldChanged="onGridRowFieldChanged"
      autoCommitLine
      hidePaging
    >
      <template #item_STNK="{ item }">
        <uploader
          ref="gridAttachmentSTNK"
          :journalId="item.ID"
          :journalType="attchKind"
          :config="{label: `STNK : ${item.LineNo}`}"
          :tags="[`${data.attchKind}_${recordId}`, `${data.attchKind}_REGINFO_STNK_${recordId}`,`${data.attchKind}_REGINFO_STNK_${recordId}_${item.ID}`]"
          :tags-for-get="[`${data.attchKind}_REGINFO_STNK_${recordId}_${item.ID}`]"
          :key="1"
          bytag
          hide-label
          single-save
          @close="emit('close')"
        />
      </template>
      <template #item_BPKB="{ item }">
        <uploader
          ref="gridAttachmentBPKB"
          :journalId="item.ID"
          :journalType="attchKind"
          :config="{label: `BPKB : ${item.LineNo}`}"
          :tags="[`${data.attchKind}_${recordId}`, `${data.attchKind} _REGINFO_BPKB_${recordId}`,`${data.attchKind}_REGINFO_BPKB_${recordId}_${item.ID}`]"
          :tags-for-get="[`${data.attchKind}_REGINFO_BPKB_${recordId}_${item.ID}`]"
          :key="1"
          bytag
          hide-label
          single-save
          @close="emit('close')"
        />
      </template>
      <template #item_Kir="{ item }">
        <uploader
          ref="gridAttachmentKir"
          :journalId="item.ID"
          :journalType="attchKind"
          :config="{label: `Kir : ${item.LineNo}`}"
          :tags="[`${data.attchKind}_${recordId}`, `${data.attchKind} _REGINFO_KIR_${recordId}`,`${data.attchKind}_REGINFO_KIR_${recordId}_${item.ID}`]"
          :tags-for-get="[`${data.attchKind}_REGINFO_KIR_${recordId}_${item.ID}`]"
          :key="1"
          bytag
          hide-label
          single-save
          @close="emit('close')"
        />
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, watch, computed } from "vue";
import { SGrid, util, loadGridConfig } from "suimjs";
import Uploader from "@/components/common/Uploader.vue";

const axios = inject("axios");

import helper from "@/scripts/helper.js";
import Loader from "@/components/common/Loader.vue";
const gridref = ref(null);
const gridAttachmentSTNK = ref(null);
const gridAttachmentBPKB = ref(null);
const gridAttachmentKir = ref(null);

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  recordId:  { type: String, default: '' },
  readOnly: { type: Boolean, default: false },
  singleSave: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  close: null,
});

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  _idDelete: [],
  gridCfg: {},
  loading: false,
  attchKind: 'ASSET_ASSET'
});

function onGridRowDelete(_, index) {
  const newRecords = data.records.filter((_, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  gridref.value.setRecords(data.records);
  updateItems();
}
function getMaxLineNo() {
  const obj = data.records.reduce(
    (acc, c) => {
      return acc.LineNo > c.LineNo ? acc : c;
    },
    { LineNo: 0 }
  );
  return obj.LineNo ?? 0;
}

function onGridNewRecord() {
  const r = {};
  r.LineNo = getMaxLineNo() + 1;
  r.ID = util.uuid();
  r.PrevPoliceNum = "";
  r.AnnualExpDate = null;
  r.YearsExpDate = null;
  r.KirTestNum = "";
  r.KirExpDate = null;
  r.STNK = false;
  r.Tax = false;
  r.Kir = false;
  data.records.push(r);
  updateItems();
}
function updateItems() {
  if (!gridref.value) return;
  gridref.value.setRecords(data.records);
  emit("update:modelValue", data.records);
}
onMounted(() => {
  data.loading =true
  loadGridConfig(axios, "/bagong/asset/detail/registerinfo/gridconfig").then(
    (r) => {
      r.setting.sortable = ["TrxDate"];
      data.gridCfg = r;
      data.loading = false
      updateItems()
    },
    (e) => {
      data.loading = false
      util.showError(e)
    }
  );
});
</script>
