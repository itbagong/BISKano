<template>
  <s-grid
    class="basic-cause-line"
    ref="BasicCauseLine"
    :config="data.cfgGridBasicCauseLine"
    hide-search
    hide-sort
    hide-refresh-button
    hide-edit
    hide-select
    hide-paging
    editor
    auto-commit-line
    no-confirm-delete
    @new-data="newBasicCause"
  >
    <template #item_button_delete="{ item, idx }">
      <a @click="deleteBasicCause(item)" class="delete_action">
        <mdicon
          name="delete"
          width="16"
          alt="delete"
          class="cursor-pointer hover:text-primary"
        />
      </a>
    </template>
    <template #item_BCDetail="{ item, idx }">
      <s-input
        label=""
        v-model="item.BCDetail"
        use-list
        :lookup-url="
          item.BCType
            ? '/tenant/masterdata/find?MasterDataTypeID=BasicCauseDetail&ParentID=' +
              item.BCType
            : '/tenant/masterdata/find?MasterDataTypeID=BasicCauseDetail'
        "
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookup-searchs="['_id', 'Name']"
        :key="item.BCType"
      />
    </template>
  </s-grid>
</template>
<script setup>
import { onMounted, inject, reactive, ref } from "vue";
import { loadGridConfig, util, SGrid, SInput } from "suimjs";
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const BasicCauseLine = ref(null);

const data = reactive({
  record: props.modelValue,
  cfgGridBasicCauseLine: {},
});

function loadGridBasicCauseline() {
  let url = `/she/investigasi/basiccause/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridBasicCauseLine = r;
      updateGridLine(data.record.BasicCause, "BasicCause");
    },
    (e) => {}
  );
}

function newBasicCause() {
  let r = {};
  const noLine = data.record.BasicCause.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.DCType = "";
  r.DCDetail = "";
  r.SubDCDetail = "";
  r.Description = "";
  data.record.BasicCause.push(r);
  updateGridLine(data.record.BasicCause, "BasicCause");
}
function deleteBasicCause(r) {
  data.record.BasicCause = data.record.BasicCause.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.BasicCause, "BasicCause");
}
function updateGridLine(record, type) {
  record.map((obj, idx) => {
    obj.LineNo = parseInt(idx) + 1;
    return obj;
  });
  if (type == "BasicCause") {
    BasicCauseLine.value.setRecords(record);
  }
}

onMounted(() => {
  loadGridBasicCauseline();
});
defineExpose({});
</script>
