<template>
  <div class="w-full flex">
    <div class="basis-2/5 pr-6 mt-2">
      <div class="border">
        <div class="border-b bg-slate-100">
          <div class="font-semibold text-center py-2">
            Last Information Tire
          </div>
        </div>
        <div class="border-b p-3">
          <div class="mb-3">
            <p class="text-slate-500 font-semibold">
              Last Date of Tire Replacement
            </p>
            <p>12 Agustus 2023</p>
          </div>
          <div class="mb-3">
            <p class="text-slate-500 font-semibold">Last Position Changed</p>
            <p>Surabaya</p>
          </div>
          <div>
            <p class="text-slate-500 font-semibold">Tire Type</p>
            <p>ABS</p>
          </div>
        </div>
        <div class="border-b p-3">
          <div>
            <p class="text-slate-500 font-semibold mb-1">
              Last Position of Changed Tire
            </p>
            <div class="flex flex-row items-center">
              <div
                v-for="index in 3"
                :key="index"
                :class="`flex flex-row items-center ${
                  index !== 1 ? 'ml-4' : ''
                }`"
              >
                <mdicon
                  size="16"
                  name="checkbox-marked-outline"
                  class="text-green-600"
                />
                <p class="ml-1">Tire {{ index }}</p>
              </div>
            </div>
          </div>
        </div>
        <div class="p-3">
          <div>
            <p class="text-slate-500 font-semibold mb-1">
              Tire Serial Type No.
            </p>
            <div class="flex flex-col">
              <div
                v-for="index in 3"
                :key="index"
                class="flex flex-row justify-between mb-1"
              >
                <p>Tire {{ index }}</p>
                <p class="font-semibold">205/45ZR19 98W</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="basis-3/5">
      <s-form
        ref="inputTireUsage"
        v-model="data.record"
        :config="data.formCfg"
        keep-label
        :buttons-on-top="false"
      >
        <template #input_TirePosition>
          <s-grid
            ref="tireList"
            class="w-full"
            editor
            hide-control
            hide-action
            :config="data.gridCfg"
            form-keep-label
          >
            <template #item_Position="{ item }">
              <s-input v-model="item.Position" read-only />
            </template>
          </s-grid>
        </template>
      </s-form>
    </div>
  </div>
</template>
<script setup>
import { reactive, onMounted, inject, ref } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  SForm,
  loadGridConfig,
  loadFormConfig,
  util,
  SGrid,
  SInput,
} from "suimjs";
import tireListData from "@/data/json/tirelist-data.json";

layoutStore().name = "tenant";

const axios = inject("axios");

const props = defineProps({
  siteEntryAssetID: { type: String, default: "" },
  modelValue: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  formCfg: {},
  gridCfg: {},
  record:
    props.modelValue == null || props.modelValue == undefined
      ? {}
      : props.modelValue,
});

const inputTireUsage = ref(null);
const tireList = ref(null);

function getFormRecord(id) {
  const url = "/bagong/siteentry_miningusage/find?SiteEntryAssetID=" + id;
  axios.post(url).then(
    (r) => {
      if (r.data.length > 0) {
        data.record = { ...r.data[0] };
        r.data[0].TirePosition.forEach((v) => {
          const existing = tireListData.find((x) => x.Position == v.Position);
          if (existing) {
            existing.TireType = v.TireType;
            existing.SerialNum = v.SerialNum;
          }
        });
        tireList.value.setRecords(tireListData);
      } else {
        data.record = {
          IsTireChange: false,
          TirePosition: [],
          TireChangePlan: new Date().toISOString(),
          TireType: "",
        };
        const initTireList = tireListData.map((v) => {
          return {
            ...v,
            TireType: "",
            SerialNum: "",
          };
        });
        tireList.value.setRecords(initTireList);
      }
      emit("update:modelValue", data.record);
    },
    (e) => util.showError(e)
  );
}

function getSelectedTire() {
  const selected = tireList.value.getSelected();
  return selected.value;
}

defineExpose({
  getSelectedTire,
});

onMounted(() => {
  loadFormConfig(axios, "/bagong/siteentry_miningusage/formconfig").then(
    (r) => {
      data.formCfg = r;
      util.nextTickN(2, () => {
        const fieldRemoved = ["IsOilChange", "OilUsage", "OilNotes"];
        fieldRemoved.forEach((field) => {
          inputTireUsage.value.removeField(field);
        });
      });
    },
    (e) => util.showError(e)
  );
  loadGridConfig(axios, "/bagong/tire_position/gridconfig").then(
    (r) => {
      data.gridCfg = r;
    },
    (e) => util.showError(e)
  );
  setTimeout(() => {
    getFormRecord(props.siteEntryAssetID);
  }, 500);
});
</script>
