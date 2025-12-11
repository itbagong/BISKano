<template>
  <div
    class="mb-4 w-full"
    :style="{ 'padding-left': stylePaddingChild(data.record.Level) + 'px' }"
  >
    <data-list
      ref="listControl"
      form-config="/she/mcuitemtemplate/lines/formconfig"
      form-read="/she/mcuitemtemplate/lines/get"
      form-keep-label
      init-app-mode="form"
      @alter-form-config="onAlterCfgForm"
      form-hide-buttons
      :form-fields="[
        'IsGender',
        'Range',
        'Condition',
        'Unit',
        'Description',
        'AnswerValue',
        'Type',
      ]"
      class="DetailMCUItemList"
      hide-title
    >
      <template #form_input_IsGender="{ item }">
        <label class="input_label">
          <div>Range</div>
        </label>
        <s-toggle
          v-model="item.IsGender"
          class="w-[150px] mt-0.5"
          yes-label="Gender"
          no-label="Non-Gender"
        />
      </template>
      <template #form_input_Range="{ item }">
        <div v-for="(dt, idx) in item.Range" :key="idx">
          <label class="input_label" v-if="dt.Name">
            <div>{{ dt.Name }}</div>
          </label>
          <div class="grid grid-cols-2 gap-2">
            <s-input
              v-model="dt.Min"
              kind="number"
              class="w-full mb-2"
              keep-label
              label="Min"
            />
            <s-input
              v-model="dt.Max"
              kind="number"
              class="w-full mb-2"
              keep-label
              label="Max"
            />
          </div>
        </div>
      </template>

      <template #form_input_Description_footer="{ item }">
        <div class="flex mt-2">
          <s-button
            label="Delete"
            class="btn_primary mr-2"
            @click="onDelete(item.ID)"
          />
          <s-button label="add child" class="btn_success" @click="onAddChild" />
        </div>
      </template>

      <template #form_input_Condition="{ item }">
        <div class="">
          <label class="input_label">
            <div>Condition</div>
          </label>
          <div
            v-for="(dt, idx) in item.Condition"
            :key="idx"
            class="w-full flex gap-4 mb-3"
          >
            <s-input v-model="dt.Name" kind="text" class="w-full" />
            <div class="flex gap-4">
              <s-input v-model="dt.Value" kind="checkbox" />
              <s-input
                v-model="dt.Vnumber"
                kind="number"
                class="w-20"
                v-if="item.AssessmentTypeIsNumber"
              />
              <s-input
                v-model="dt.Letter"
                use-list
                class="w-32"
                v-if="!item.AssessmentTypeIsNumber"
                :lookup-url="'/tenant/masterdata/find?MasterDataTypeID=DISC'"
                lookup-key="_id"
                :lookup-labels="['Name']"
              />
              <uploader
                :journalId="dt.ID"
                :config="{}"
                journalType="MCU_ITEM_TEMPLATE_LINES_LIST"
                is-single-upload
                single-save
                hide-label
                class="mt-[-5px]"
              />
              <mdicon
                name="delete"
                width="16"
                alt="delete"
                class="cursor-pointer hover:text-primary"
                @click="item.Condition.splice(idx, 1)"
              />
            </div>
          </div>
          <div class="flex">
            <s-button
              icon="plus"
              class="btn_primary"
              tooltip="add new"
              @click="
                item.Condition.push({
                  ID: util.uuid(),
                  Name: '',
                  Value: false,
                })
              "
            />
          </div>
        </div>
      </template>
      <template #form_input_AnswerValue="{ item }">
        <div class="flex gap-4">
          <s-input
            v-model="item.AnswerValue"
            kind="number"
            class="w-full basis-1/2"
            label="Answer Value"
            keep-label
            v-if="['List', 'String'].includes(item.Type)"
          />
          <uploader
            :journalId="item.ID"
            :config="{ label: 'Attachment' }"
            journalType="MCU_ITEM_TEMPLATE_LINES"
            is-single-upload
            single-save
            v-if="['List', 'String'].includes(item.Type)"
          />
        </div>
      </template>
      <template #form_input_Type="{ item }">
        <div class="grid grid-cols-1 gap-4">
          <s-input
            class="w-full"
            v-model="item.Type"
            label="Type"
            use-list
            :items="['List', 'Range', 'String']"
            @change="onTypeChange"
            v-if="!checkParent(data.record.ID)"
          />

          <div class="" v-if="item.Type == 'List'">
            <label class="input_label flex mb-1">Assessment Type </label>
            <s-toggle
              v-model="item.AssessmentTypeIsNumber"
              class="w-[150px] mt-0.5"
              yes-label="Number"
              no-label="Letter"
              @change="item.QuestionTypeIsMost = false"
            />
          </div>

          <div v-if="item.Type == 'List' && !item.AssessmentTypeIsNumber">
            <label class="input_label flex mb-1">Question Type </label>
            <s-toggle
              v-model="item.QuestionTypeIsMost"
              class="w-[150px] mt-0.5"
              yes-label="Most"
              no-label="Least"
            />
          </div>
        </div>
      </template>
    </data-list>
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

import { SInput, util, SButton, SGrid, SCard, DataList } from "suimjs";
import helper from "@/scripts/helper.js";
import SToggle from "@/components/common/SButtonToggle.vue";
import Uploader from "@/components/common/Uploader.vue";

const axios = inject("axios");
const listControl = ref(null);

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  dataItems: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
  updateItems: null,
});

const data = reactive({
  record: props.modelValue,
  formCfg: {},
});

function onDelete(id) {
  let dt = props.dataItems;
  let res = dt.filter((o) => o.ID !== id);
  dt = res.filter((o) => {
    return !o.Parent.includes(id);
  });
  emit("updateItems", dt);
}

function onAddChild() {
  let dt = data.record;
  let item = props.dataItems;
  let idx = item.findIndex((o) => o.ID == dt.ID);
  let lenChild = item.filter((o) => o.Parent.includes(dt.ID));
  let idxSlice = idx + lenChild.length;
  let parent = (dt.Parent ? dt.Parent + "#" : "") + dt.ID;
  item.splice(idxSlice + 1, 0, {
    ID: util.uuid(),
    Parent: parent,
    Condition: [],
    IsGender: false,
    Range: [
      {
        Name: "",
        Min: 0,
        Max: 0,
      },
    ],
    Level: dt.Level + 1,
  });
  dt.Type = "";
  onTypeChange();
  emit("update:modelValue", dt);
}

function checkParent(id) {
  let dt = props.dataItems;
  let res = dt.findIndex((o) => o.Parent.includes(id));
  return res > -1;
}

function onReset() {
  util.nextTickN(3, () => {
    data.record.Unit = "";
    data.record.Condition = [];
    data.record.Min = 0;
    data.record.Max = 0;
    data.record.IsGender = false;
    data.record.Range = [{ Name: "", Min: 0, Max: 0 }];
    data.record.AssessmentTypeIsNumber = false;
    data.record.QuestionTypeIsMost = false;
    data.record.AnswerValue = 0;
  });
}

function formatMCUGender(val) {
  let res = [];
  let arr = val ? ["Laki - Laki", "Perempuan"] : [""];

  for (let i in arr) {
    let o = {
      Name: arr[i],
      Min: 0,
      Max: 0,
    };
    res.push(o);
  }

  data.record.Range = res;
}

function onAlterCfgForm(cfg) {
  listControl.value.setFormRecord(data.record);
  const hideField = ["Result", "Note"];
  util.nextTickN(3, () => {
    onTypeChange(true);
    if (checkParent(data.record.ID)) {
      listControl.value.setFormFieldAttr("Type", "hide", true);
    }
  });
}

function onTypeChange(isFirst = false) {
  util.nextTickN(3, () => {
    let type = data.record.Type;
    let obj = {
      String: [],
      Range: ["Unit", "Min", "Max", "IsGender", "Range"],
      List: ["Condition"],
    };

    const allField = ["IsGender", "Unit", "Condition", "Range", "Min", "Max"];

    for (let i in allField) {
      let o = allField[i];
      if (!type) {
        listControl.value.setFormFieldAttr(o, "hide", true);
      } else {
        let val = obj[type].includes(o);
        listControl.value.setFormFieldAttr(o, "hide", !val);
      }
    }
    if (!isFirst) onReset();
  });
}

function stylePaddingChild(params) {
  if (params == 0) return "";
  else return params * 30;
}

watch(
  () => data.record.IsGender,
  (nv) => {
    formatMCUGender(nv);
  }
);
</script>

<style>
.gridCol3 {
  @apply grid-cols-3;
}

.DetailMCUItemList .flex.flex-col.gap-4 > .grid.gridCol4 {
  @apply gridCol3;
}
</style>
