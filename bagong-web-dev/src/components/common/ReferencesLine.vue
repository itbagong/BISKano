<template>
  <div class="w-full">
    <div class="grid grid-cols-2 gap-4">
      <template v-for="(dt, idx) in data.refList.Items" :key="'input_' + idx">
        <s-input
          :read-only="readOnly"
          :kind="dt.ReferenceType"
          v-model="data.refValues[idx].Value"
          :label="dt.Label"
          class="w-full"
          v-if="!['lookup', 'items'].includes(dt.ReferenceType)"
          :multi-row="dt.ReferenceType === 'textarea' ? 3 : 0"
        />
        <s-input
          :read-only="readOnly"
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
          :read-only="readOnly"
          :kind="dt.ReferenceType"
          v-model="data.refValues[idx].Value"
          :label="dt.Label"
          class="w-full"
          v-if="dt.ReferenceType == 'items'"
          use-list
          :items="formatItemList(dt.ConfigValue)"
        />
      </template>
    </div>

    <div v-if="!hideManualInput">
      <div class="card_title grow text-xs font-semibold mt-2">Manual input</div>
      <div class="grid grid-cols-1 gap-4">
        <template v-for="(dt, idx) in data.refValues" :key="'input_' + idx">
          <div
            class="grid grid-cols-2 gap-4"
            v-if="idx > data.refList.Items.length - 1"
          >
            <s-input
              kind="text"
              v-model="dt.Key"
              class="w-full"
              label="Label"
            />
            <div class="flex gap-2">
              <s-input
                :read-only="readOnly"
                kind="text"
                v-model="dt.Value"
                class="w-full"
                label="Config value"
              />
              <mdicon
                v-if="!readOnly"
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
            v-if="!readOnly"
            class="btn_primary"
            label="Add"
            @click="data.refValues.push({ Key: '', Value: '' })"
          />
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, inject, watch, onMounted } from "vue";
import { SInput, SButton } from "suimjs";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false },
  refList: {
    type: Object,
    default: () => {
      Items: [];
    },
  },
  hideManualInput: { type: Boolean, default: false },
});

const axios = inject("axios");
const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  refList: props.refList,
  refValues: props.modelValue,
});

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
  () => props.refList,
  (nv) => {
    data.refList = nv;
  },
  { deep: true }
);
watch(
  () => props.refValues,
  (nv) => {
    data.refValues = nv;
  },
  { deep: true }
);
</script>
