<template>
  <s-grid
    class="PicaLine"
    ref="PicaLine"
    :config="data.cfgGridPicaLine"
    hide-search
    hide-sort
    hide-refresh-button
    hide-edit
    hide-select
    hide-paging
    hide-new-button
    editor
    auto-commit-line
    no-confirm-delete
  >
  </s-grid>
</template>
<script setup>
import { onMounted, inject, reactive, ref } from "vue";
import { loadGridConfig, util, SGrid } from "suimjs";
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const PicaLine = ref(null);

const data = reactive({
  value: props.modelValue,
  cfgGridPicaLine: {},
});

function updateGridLines() {
  PicaLine.value.setRecords(data.value.PICA);
}

function updateGridLine(record, type) {
  updateGridLines();
}

function loadGridPicaline() {
  let url = `/she/investigasi/pica/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridPicaLine = r;
      updateGridLine(data.record.PICA, "Pica");
    },
    (e) => {}
  );
}

onMounted(() => {
  loadGridPicaline();
});
defineExpose({
  updateGridLines,
});
</script>
