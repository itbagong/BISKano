<template>
  <data-list
    class="card MCUitemTemplate"
    ref="listControl"
    title="Master Item Template"
    grid-config="/she/mcuitemtemplate/gridconfig"
    form-config="/she/mcuitemtemplate/formconfig"
    grid-read="/she/mcuitemtemplate/gets"
    form-read="/she/mcuitemtemplate/get"
    grid-mode="grid"
    grid-delete="/she/mcuitemtemplate/delete"
    form-keep-label
    form-insert="/she/mcuitemtemplate/save"
    form-update="/she/mcuitemtemplate/save"
    :init-app-mode="data.appMode"
    grid-hide-select
    @form-edit-data="openForm"
    @form-new-data="newData"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    stay-on-form-after-save
    :form-tabs-edit="['General', 'Lines', 'Instruction']"
    :form-fields="['Dimension', 'IsActive']"
  >
    <template #form_input_Dimension="{ item, mode }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :read-only="mode == 'view'"
      ></dimension-editor-vertical>
    </template>
    <template #form_input_IsActive="{ item }">
      <label class="input_label">
        <div>Status</div>
      </label>
      <s-toggle
        v-model="item.IsActive"
        class="w-[120px] mt-0.5"
        yes-label="active"
        no-label="inactive"
      />
    </template>
    <template #form_tab_Lines="{ item, mode }">
      <lines
        v-for="(dt, idx) in item.Lines"
        v-model="item.Lines[idx]"
        :key="dt.ID"
        :data-items="item.Lines"
        @updateItems="
          (val) => {
            item.Lines = val;
          }
        "
      />
      <div class="flex mt-4">
        <s-button
          class="bg-success text-white font-bold"
          label="Add Item"
          @click="
            item.Lines.push({
              ID: util.uuid(),
              Parent: '',
              Condition: [],
              Range: [{ Name: '', Min: 0, Max: 0 }],
              Level: 0,
            })
          "
        ></s-button>
      </div>
    </template>
    <template #form_tab_Instruction="{ item, mode }">
      <instructions v-model="item.Instruction" :jurnal-id="item._id" />
    </template>
  </data-list>
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
  loadGridConfig,
} from "suimjs";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import { layoutStore } from "@/stores/layout.js";
import SToggle from "@/components/common/SButtonToggle.vue";
import Lines from "./widget/DetailMCUItemTemplate.vue";
import Instructions from "./widget/McuItemTemplateInstruction.vue";

layoutStore().name = "tenant";
const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  record: {},
});

function newData(r) {
  r.TrxDate = new Date();
}
</script>

<style>
.MCUitemTemplate .items-start.gap-2.grid.gridCol4,
.MCUitemTemplate .items-start.gap-2.grid.gridCol3 {
  @apply gap-6;
}
</style>
