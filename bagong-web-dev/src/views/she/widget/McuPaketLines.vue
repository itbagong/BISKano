<template>
  <div class="grid grid-cols-3 gap-12">
    <div class="flex gap-4">
      <s-input
        class="w-full"
        v-model="data.record.Name"
        useList
        label="Jenis Pemeriksaan"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['_id', 'Name']"
        lookupUrl="/she/mcuitemtemplate/find"
        :lookup-payload-builder="
          (search) => lookupPayloadBuilder(search, data.record.Name)
        "
        @change="getDetails"
      />
      <mdicon
        name="delete"
        width="16"
        alt="delete"
        class="cursor-pointer hover:text-primary pt-4"
        @click="dataItems.splice(index, 1)"
      />
    </div>
    <div class="cols-pan-2">
      <label class="input_label">Rincian</label>
      <div v-for="(dt, idx) in data.record.Lines" :key="idx">
        <div class="flex gap-2" v-if="!data.loading">
          <s-input
            kind="checkbox"
            v-model="dt.IsSelected"
            @change="onChangeCheck(idx)"
          />
          <div class="flex">
            <div class="pr-3" v-if="dt.Parent"></div>
            {{ dt.Description }}
          </div>
        </div>
      </div>
      <loader kind="linier" skeleton-kind="input" v-if="data.loading" />
    </div>
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
  loadGridConfig,
} from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import Loader from "@/components/common/Loader.vue";
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  dataItems: { type: Array, default: () => [] },
  index: { type: Number, default: () => 0 },
  selectedId: { type: Array, default: () => [] },
});

const data = reactive({
  record: props.modelValue,
  loading: false,
});

function getDetails() {
  util.nextTickN(2, () => {
    data.loading = true;
    const url = "/she/mcuitemtemplate/get";
    axios.post(url, [data.record.Name]).then(
      (r) => {
        data.record.Lines = r.data.Lines;
        data.loading = false;
      },
      (e) => {
        data.loading = false;
      }
    );
  });
}

function lookupPayloadBuilder(search, value) {
  const qp = {};
  qp.Where = {
    Op: "$or",
    items: [
      { Field: "_id", Op: "$eq", Value: value },
      { Field: "_id", Op: "$nin", Value: props.selectedId },
    ],
  };
  return qp;
}

function onChangeCheck(idx) {
  util.nextTickN(2, () => {
    let val = data.record.Lines[idx].IsSelected;
    let parentID = data.record.Lines[idx].Parent;
    let currentID = data.record.Lines[idx].ID;
    for (let i in data.record.Lines) {
      let o = data.record.Lines[i];
      if (o.ID == parentID) {
        o.IsSelected = true;
      }

      if (currentID == o.Parent) {
        o.IsSelected = val;
      }
    }
  });
}
</script>
