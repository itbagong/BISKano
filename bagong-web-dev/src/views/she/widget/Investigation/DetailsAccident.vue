<template>
  <s-form
    ref="detailAccodentFormCtl"
    v-model="data.record"
    :config="data.formCfgDtlAccident"
    keep-label
    only-icon-top
    :buttons-on-top="false"
    hide-cancel
    auto-focus
  >
    <template #input_AccidentAttachment="{ item, idx }">
      <s-grid-attachment
        :key="data.record._id"
        :journal-id="data.record._id"
        :tags="linesTagAccident"
        journal-type="Accident"
        ref="gridAccidentAttachment"
        @pre-Save="preSaveAccidentAttachment"
      ></s-grid-attachment>
    </template>
    <template #input_AccidentType="{ item, idx }">
      <s-grid
        class="accidentType_lines"
        ref="AccidentType"
        :config="data.cfgGridAccidentType"
        hide-search
        hide-sort
        hide-refresh-button
        hide-edit
        hide-select
        hide-paging
        editor
        auto-commit-line
        no-confirm-delete
        @new-data="newAccidentType"
      >
        <template #item_button_delete="{ item, idx }">
          <a @click="deleteAccidentType(item)" class="delete_action">
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
  </s-form>
</template>
<script setup>
import { onMounted, inject, reactive, ref, computed } from "vue";
import { loadGridConfig, loadFormConfig, util, SForm, SGrid } from "suimjs";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import helper from "@/scripts/helper.js";
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);
const AccidentType = ref(null);
const gridAccidentAttachment = ref(null);
const linesTagAccident = computed({
  get() {
    const tags = [data.record._id];
    return tags;
  },
});

const data = reactive({
  record: props.modelValue,
  formCfgDtlAccident: {},
  cfgGridAccidentType: {},
});

function preSaveAccidentAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [`inv-dtl-accident-${data.record._id}`];
    return asset;
  });
}

function loadFromDetailAccident() {
  let url = `/she/investigasi/detailsaccident/formconfig`;
  loadFormConfig(axios, url).then(
    (r) => {
      data.formCfgDtlAccident = r;
      loadGridAccidentType();
    },
    (e) => {}
  );
}
function loadGridAccidentType() {
  let url = `/she/investigasi/accidenttypeline/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridAccidentType = r;
      updateGridLine(data.record.AccidentType, "AccidentType");
    },
    (e) => {}
  );
}

function newAccidentType() {
  let r = {};
  const noLine = data.record.AccidentType.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.Type = "";
  r.Explaination = "";
  data.record.AccidentType.push(r);
  updateGridLine(data.record.AccidentType, "AccidentType");
}
function deleteAccidentType(r) {
  data.record.AccidentType = data.record.AccidentType.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.AccidentType, "AccidentType");
}
function updateGridLine(record, type) {
  record.map((obj, idx) => {
    obj.NoLine = parseInt(idx) + 1;
    return obj;
  });
  if (type == "AccidentType") {
    AccidentType.value.setRecords(record);
  }
}

function onSaveAttachment() {
  gridAccidentAttachment.value.Save();
}

onMounted(() => {
  loadFromDetailAccident();
});
defineExpose({
  onSaveAttachment,
});
</script>
