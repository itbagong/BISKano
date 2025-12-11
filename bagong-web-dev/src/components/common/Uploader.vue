<template>
  <div>
    <label class="input_label flex mb-1" v-if="!hideLabel">
      {{ config.label }}
    </label>
    <div
      class="box-uploader-style"
      @click="onOpenContent"
      :class="[
        readOnly
          ? 'cursor-not-allowed pointer-events-none opacity-50 select-none'
          : '',
      ]"
    >
      <mdicon name="upload" size="16" />
    </div>
    <div
      v-show="data.modal"
      class="w-full h-full absoluted top-0 left-0 z-50 flex items-center justify-center fixed"
    >
      <div
        class="min-w-[500px] w-[1000px] max-h-[600px] overflow-auto p-0 border shadow-md"
      >
        <div class="flex justify-between bg-primary py-2 px-4">
          <div class="text-white flex gap-2 text-base font-semibold">
            {{ config.label }}
          </div>
          <s-button class="btn_warning" label="Close" @click="close" />
        </div>
        <attachment
          v-if="journalId !== ''"
          class="p-4 bg-white"
          ref="gridAttachment"
          :journalId="journalId"
          :journalType="journalType"
          :is-single-upload="isSingleUpload"
          :read-only="readOnly"
          :singleSave="singleSave"
          :tags="tags"
          :by-tag="bytag"
          @preGets="preGets"
          @preSave="preSave"
        />
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, computed, watch } from "vue";
import { util, SButton, SModal } from "suimjs";
import Attachment from "@/components/common/SGridAttachment.vue";

const gridAttachment = ref(null);
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  journalId: { type: String, default: "" },
  journalType: { type: String, default: "" },
  config: { type: Object, default: () => {} },
  tags: { type: Array, default: [] },
  tagsForGet: { type: Array, default: [] },
  bytag: { type: Boolean, default: false },
  isSingleUpload: { type: Boolean, default: false },
  readOnly: { type: Boolean, default: false },
  singleSave: { type: Boolean, default: false },
  hideLabel: { type: Boolean, default: false },
});

const emit = defineEmits({
  preOpen: null,
  "update:modelValue": null,
  close: null,
});

const data = reactive({
  modal: false,
});
function onOpenContent() {
  if (props.readOnly) return;
  emit("preOpen");
  util.nextTickN(2, () => {
    data.modal = true;
  });
}
function close() {
  emit("close");
  data.modal = false;
}
function Save(journalId, journalType) {
  gridAttachment.value.Save(journalId, journalType);
}

function refreshGrid() {
  if (props.journalId) gridAttachment.value.refreshGrid();
}

function preSave(param) {
  param.Asset = {
    ...param.Asset,
    ...{ Tags: props.tags },
  };
}
function preGets(param) {
  if (props.bytag) {
    param.JournalID = undefined;
    param.JournalType = undefined;
    param.Tags = props.tagsForGet.length > 0 ? props.tagsForGet : props.tags;
  }
}
watch(
  () => data.modal,
  (nv) => {
    if (nv) refreshGrid();
    emit("modal", nv);
  }
);

defineExpose({
  Save,
});
</script>
<style>
.box-uploader-style {
  @apply border-2 border-dashed border-zinc-200 w-12 h-12 mr-2 rounded-md grid place-content-center relative cursor-pointer;
}
.box-uploader-style:hover {
  @apply border-zinc-500 text-zinc-500;
}
</style>
