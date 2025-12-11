<template>
  <div>
    <data-list
      class="mcu-detail-result"
      ref="listControl"
      form-config="/she/mcutransaction/resultdetail/formconfig"
      form-read="/she/mcutransaction/resultdetail/get"
      init-app-mode="form"
      stay-on-form-after-save
      form-hide-buttons
      hide-title
      :form-fields="['Description', 'NilaiRujukan', 'Result']"
      no-gap
      @alter-form-config="onAlterCfgForm"
    >
      <template #form_input_Description="{ item }">
        <div class="flex pl-2 mb-2 font-semibold" v-if="data.name">
          {{ data.name }}
        </div>
        <div :class="[item.Parent ? 'pl-10' : 'pl-6']">
          {{ item.Description }}
        </div>
      </template>
      <template #form_input_Result="{ item }">
        <div class="flex gap-4">
          <s-input
            v-model="item.Result"
            class="w-full"
            v-if="item.Type == 'List'"
            use-list
            :items="data.itemList"
          />
          <s-input
            v-model="item.Result"
            class="w-full"
            v-if="item.Type == 'Range'"
          />
          <mdicon
            name="circle"
            width="16"
            alt="delete"
            :class="[setIndicator(item) ? 'text-success' : 'text-primary']"
            v-if="['List', 'Range'].includes(item.Type)"
          />
        </div>
      </template>
      <template #form_input_NilaiRujukan="{ item }">
        <div class="flex">
          <div v-if="item.Type == 'List'" class="pl-2 font-semibold">
            {{ getCorrectList(item.Condition) }}
          </div>
          <div v-if="item.Type == 'Range'" class="pl-2 font-semibold">
            <div v-for="(dt, idx) in item.Range" :key="idx" class="flex gap-2">
              <div v-if="dt.Name">{{ dt.Name }}</div>
              {{ dt.Min + " - " + dt.Max }}
            </div>
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
const listControl = ref(null);
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  name: { type: String, default: () => "" },
  items: { type: Array, default: () => [] },
  gender: { type: String, default: () => "" },
});
const data = reactive({
  record: props.modelValue,
  name: "",
  itemList: props.modelValue.Condition.map((o) => {
    return { key: o.Name, text: o.Name };
  }),
});

function getName(id) {
  const url = "/she/mcuitemtemplate/get";
  axios.post(url, [id]).then(
    (r) => {
      data.name = r.data.Name;
    },
    (e) => {
      util.showError(e);
    }
  );
}

function isParent() {
  let dt = props.items;
  let currentID = props.modelValue.ID;
  let findex = dt.findIndex((o) => o.Parent == currentID);
  if (findex > -1) {
    getName(props.name);
    let fieldHide = ["Result", "Unit", "Note", "NilaiRujukan"];
    for (let i in fieldHide) {
      let o = fieldHide[i];
      listControl.value.setFormFieldAttr(o, "hide", true);
    }
  }
}

function getCorrectList(dt) {
  let res = dt.find((o) => o.Value);
  return res ? res.Name : "";
}

function setIndicator(dt) {
  let res = false;

  if (!dt.Result) return false;

  if (dt.Type == "List") {
    let correntVal = getCorrectList(dt.Condition);
    res = correntVal == dt.Result;
  }

  if (dt.Type == "Range" && dt.IsGender) {
    let findex = dt.Range.findIndex((item) => item.Name == props.gender);
    let correntVal = parseFloat(dt.Result);
    let obj = dt.Range[findex];
    res = obj.Min <= correntVal && obj.Max >= correntVal;
  }

  if (dt.Type == "Range" && !dt.IsGender) {
    let correntVal = parseFloat(dt.Result);
    let obj = dt.Range[0];
    res = obj.Min <= correntVal && obj.Max >= correntVal;
  }

  return res;
}

function onAlterCfgForm(cfg) {
  util.nextTickN(3, () => {
    listControl.value.setFormRecord(data.record);
    isParent();
  });
}
</script>

<style>
.mcu-detail-result .input_label {
  @apply hidden;
}

.mcu-detail-result .suim_form {
  @apply pt-0;
}
</style>
