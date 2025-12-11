<template>
  <s-grid
    class="DirectCauseLine"
    ref="DirectCauseLine"
    :config="data.cfgGridDirectCauseLine"
    hide-search
    hide-sort
    hide-refresh-button
    hide-edit
    hide-select
    hide-paging
    editor
    auto-commit-line
    no-confirm-delete
    @new-data="newDirectCause"
  >
    <template #item_button_delete="{ item, idx }">
      <a @click="deleteDirectCause(item)" class="delete_action">
        <mdicon
          name="delete"
          width="16"
          alt="delete"
          class="cursor-pointer hover:text-primary"
        />
      </a>
    </template>
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

const DirectCauseLine = ref(null);

const data = reactive({
  record: props.modelValue,
  cfgGridDirectCauseLine: {},
});

function loadGridDirectCauseline() {
  let url = `/she/investigasi/directcause/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridDirectCauseLine = r;
      updateGridLine(data.record.DirectCause, "DirectCause");
    },
    (e) => {}
  );
}

function newDirectCause() {
  let r = {};
  const noLine = data.record.DirectCause.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.DCType = "";
  r.DCDetail = "";
  r.SubDCDetail = "";
  r.Description = "";
  data.record.DirectCause.push(r);
  updateGridLine(data.record.DirectCause, "DirectCause");
}
function deleteDirectCause(r) {
  data.record.DirectCause = data.record.DirectCause.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.DirectCause, "DirectCause");
}
function updateGridLine(record, type) {
  record.map((obj, idx) => {
    obj.LineNo = parseInt(idx) + 1;
    return obj;
  });
  if (type == "DirectCause") {
    DirectCauseLine.value.setRecords(record);
  }
}

onMounted(() => {
  loadGridDirectCauseline();
});
defineExpose({});
</script>
