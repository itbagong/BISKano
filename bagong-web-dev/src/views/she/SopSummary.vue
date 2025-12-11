<template>
  <data-list
    ref="listControl"
    title="SOP Summary"
    grid-config="/she/mastersopsummary/gridconfig"
    grid-read="/she/mastersopsummary/gets"
    grid-mode="grid"
    form-keep-label
    init-app-mode="grid"
    @form-edit-data="openForm"
    :form-fields="['Dimension']"
    :grid-fields="['Dimension', 'Attachment']"
    v-if="data.appMode == 'grid'"
    grid-hide-new
    grid-hide-delete
  >
    <template #form_input_Dimension="{ item, mode }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :read-only="mode == 'view'"
      ></dimension-editor-vertical>
    </template>
    <template #grid_Attachment="{ item }">
      <uploader
        :config="{}"
        read-only
        :tags="[`SHE_SOP_SUMMARY_${item._id}_${effDate(item.EffectiveDate)}`]"
      />
    </template>
  </data-list>

  <div v-else-if="data.appMode == 'detail'" class="card w-full">
    <div class="card_title grow mb-2">Document History</div>
    <s-grid
      :config="data.cfgDetail"
      hide-new-button
      hide-action
      hide-select
      :read-url="`/she/mastersop/gets?Status=APPROVED&DocumentRefno=${data.record._id}`"
    >
      <template #header_buttons_2="{}">
        <s-button
          label="Back"
          tooltip="Back"
          icon="rewind"
          class="btn_warning back_btn"
          @click="onCancel"
        />
      </template>
      <template #item_Attachment="{ item }">
        <uploader
          ref="gridAttachment"
          :journalId="item._id"
          :config="{}"
          journalType="SHE_SOP"
          single-save
          read-only
        />
      </template>
    </s-grid>
  </div>
</template>

<script setup>
import {
  reactive,
  ref,
  inject,
  watch,
  computed,
  onMounted,
  nextTick,
} from "vue";

import {
  SInput,
  util,
  SButton,
  SGrid,
  SCard,
  DataList,
  SModal,
  loadGridConfig,
} from "suimjs";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import Uploader from "@/components/common/Uploader.vue";
import { layoutStore } from "@/stores/layout.js";
import moment from "moment";

layoutStore().name = "tenant";

const axios = inject("axios");
const listControl = ref(null);

const data = reactive({
  appMode: "grid", //grid|detail,
  record: {},
  cfgDetail: {},
});

function openForm(r) {
  data.record = r;
  data.appMode = "detail";
}

function onCancel() {
  data.appMode = "grid";
}

function effDate(params) {
  return moment(params).utc().format("YYYY-MM-DDTHH:mm:ss");
}

onMounted(() => {
  loadGridConfig(axios, "/she/mastersop/gridconfig").then(
    (r) => {
      data.cfgDetail = r;
    },
    (e) => util.showError(e)
  );
});
</script>
