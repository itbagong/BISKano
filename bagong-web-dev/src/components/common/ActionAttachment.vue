<template>
  <div class="px-1">
    <s-tooltip :tooltip="tooltip">
      <template #content>
        <a
          v-if="!props.hideButton"
          href="#"
          @click="onOpenContent"
          :class="buttonClass"
        >
          <mdicon
            :name="icon"
            :width="iconWidth"
            alt="edit"
            class="cursor-pointer hover:text-primary"
          />
        </a>
        <s-button
          v-if="props.actionOnHeader"
          @click="onOpenContent"
          icon="attachment"
          class="btn_primary"
          label="Attachment"
        ></s-button>
      </template>
    </s-tooltip>
    <s-modal
      v-if="data.openContent"
      title="Attachment"
      class="model-reject"
      display
      ref="reject"
      @beforeHide="close"
      hideButtons
    >
      <div class="min-w-[500px] w-[1200px] max-h-[600px] overflow-auto">
        <s-grid-attachment
          :journalId="refId"
          :journalType="kind"
          :tags="tags"
          single-save
          by-tag
          ref="gridAttachmentCtl"
          :read-only="readOnly"
          @preSave="preSave"
          @preGets="preGets"
        />
      </div>
    </s-modal>
  </div>
</template>
<script setup>
import { reactive, ref, watch, onMounted, inject } from "vue";
import { SButton, STooltip, SModal, SCard, util } from "suimjs";
import Loader from "./Loader.vue";
import moment from "moment";
import SGridAttachment from "@/components/common/SGridAttachment.vue";

const axios = inject("axios");
const props = defineProps({
  kind: { type: String, default: "" },
  refId: { type: String, default: "" },
  tags: { type: Array, default: [] },
  showContent: { type: Boolean, default: false },
  hideButton: { type: Boolean, default: false },
  readOnly: { type: Boolean, default: false },
  actionOnHeader: { type: Boolean, default: false },
  tagsForGet: { type: Array, default: [] },
  tooltip: { type: String, default: "Upload attachment" },
  buttonClass: { type: String, default: "" },
  icon: { type: String, default: "upload" },
  iconWidth: { type: Number, default: 16 },
});
const emit = defineEmits({
  preOpenContent: null,
  close: null,
});

const data = reactive({
  openContent: props.showContent === undefined ? false : props.showContent,

  loading: false,
});
const gridAttachmentCtl = ref(null);

function onOpenContent() {
  emit("preOpen");
  data.openContent = true;
}
function close() {
  emit("close");
  data.openContent = false;
}
defineExpose({
  deleteAttch,
});
function deleteAttch(cb) {
  axios
    .post("/asset/delete-by-journal", {
      JournalType: props.kind,
      JournalID: props.RefID,
      Tags: props.tags,
    })
    .then(
      (r) => {},
      (err) => {}
    )
    .finally(() => {
      cb();
    });
}

function preSave(param) {
  param.Asset = {
    ...param.Asset,
    ...{ Tags: props.tags },
  };
}
function preGets(param) {
  param.JournalID = undefined;
  param.JournalType = undefined;
  param.Tags = props.tagsForGet.length > 0 ? props.tagsForGet : props.tags;
}
</script>
No newline at end of file
