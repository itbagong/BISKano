<template>
  <div class="w-full">
    <div class="grid grid-cols-1 gap-4">
      <div class="grid grid-cols-4 gap-4">
        <template v-for="(dt, idx) in data.refList.Items">
          <slot
            :name="'ref_' + dt.Label.replace(/ /g, '_')"
            :item="dt"
            :value="data.refValues[idx]"
          >
            <s-input
              :kind="dt.ReferenceType"
              v-model="data.refValues[idx].Value"
              :label="dt.Label"
              class="w-full"
              v-if="!['lookup', 'items'].includes(dt.ReferenceType)"
              :multi-row="dt.ReferenceType === 'textarea' ? 3 : 0"
            />
            <s-input
              :kind="dt.ReferenceType"
              v-model="data.refValues[idx].Value"
              :label="dt.Label"
              class="w-full"
              lookup-key="_id"
              :lookup-labels="formatLookup(dt.ConfigValue, 'labels')"
              :lookup-searchs="formatLookup(dt.ConfigValue, 'search')"
              :lookup-url="formatLookup(dt.ConfigValue, 'url')"
              use-list
              v-if="dt.ReferenceType == 'lookup'"
            />
            <s-input
              :kind="dt.ReferenceType"
              v-model="data.refValues[idx].Value"
              :label="dt.Label"
              class="w-full"
              v-if="dt.ReferenceType == 'items'"
              use-list
              :items="formatItemList(dt.ConfigValue)"
            />
          </slot>
        </template>
      </div>
      <div class="card_title grow text-xs font-semibold">Manual input</div>
      <template v-for="(dt, idx) in data.refValues" :key="'input_' + idx">
        <div
          class="grid grid-cols-2 gap-4"
          v-if="idx > data.refList.Items.length - 1"
        >
          <s-input kind="text" v-model="dt.Key" class="w-full" label="Label" />
          <div class="flex gap-2">
            <s-input
              kind="text"
              v-model="dt.Value"
              class="w-full"
              label="Config value"
            />
            <mdicon
              name="delete"
              size="16"
              class="cursor-pointer mt-5"
              @click="data.refValues.splice(idx, 1)"
            />
          </div>
        </div>
      </template>
      <div class="flex justify-end my-4">
        <s-button
          class="btn_primary"
          label="Add"
          @click="data.refValues.push({ Key: '', Value: '' })"
        />
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, inject, watch, onMounted } from "vue";
import moment from "moment";
import {
  SInput,
  util,
  SForm,
  formInput,
  createFormConfig,
  SButton,
} from "suimjs";
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false },
  ReferenceTemplate: { type: String, default: "" },
});

const axios = inject("axios");
const emit = defineEmits({
  "update:modelValue": null,
  getItems: null,
});

const data = reactive({
  refList: {},
  refValues: [],
});

function getReferenceTemplate(id) {
   if (id == "") {
      data.refList = {
         Items: []
      }
      data.refValues = props.modelValue;
      return;
  }
  const url = "/tenant/referencetemplate/get";
  axios.post(url, [id]).then(
    (r) => {
      data.refList = r.data;
      data.refValues =
        props.modelValue.length == 0
          ? mappingRefValues(r.data.Items)
          : props.modelValue;
      emit("getItems", r.data.Items);
    },
    (e) => {}
  );
}

function mappingRefValues(sources) {
  const defaultValue = {
    date: null,
    number: 0,
  };
  const mv = props.modelValue;

  const res = sources.map((el) => {
    const ky = el.Label;
    const f = mv.filter((o) => o.Key == ky);
    const v = f.length == 0 ? defaultValue[el.ReferenceType] ?? "" : f[0].Value;
    return {
      Key: ky,
      Value: ky.toLowerCase().includes("date") ? null : v,
    };
  });
  return res;
}

function formatItemList(sources) {
  let dt = sources.split("|");
  const res = dt.map((el) => {
    return {
      text: el,
      value: el,
    };
  });
  return res;
}

function formatLookup(sources, flag) {
  let splited = sources.split("|");
  if (splited.length == 0) return "";

  if (flag == "url") {
    return splited[0];
  }
  if (flag == "labels") {
    return [splited[2].split(",")[1]];
  }
  if (flag == "search") {
    return splited[2].split(",");
  }
}

watch(
  () => data.refValues,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

watch(
  () => props.ReferenceTemplate,
  (nv) => {
    getReferenceTemplate(nv);
  },
  { deep: true }
);

onMounted(() => {
  getReferenceTemplate(props.ReferenceTemplate);
});
</script>
