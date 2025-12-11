<template>
  <s-modal title="Confirm" display hideButtons hideClose>
    <s-card hide-title class="min-w-[700px]">
      <div class="py-3 mb-3">
        <template v-if="data.loading">
          <div class="w-full h-[30px] mb-3">
            <loader kind="skeleton" skeleton-kind="input" />
          </div>
          <div class="w-full h-[30px] mb-3">
            <loader kind="skeleton" skeleton-kind="input" />
          </div>
          <div class="w-full h-[30px] mb-3">
            <loader kind="skeleton" skeleton-kind="input" />
          </div>
        </template>
        <ul class="flex flex-col" v-else>
          <li v-for="(item, idx) in data.items" :key="idx" class="mb-6">
            <div class="w-full flex justify-between">
              <div class="mb-2 font-semibold">
                {{ item.SourceJournalID }} - {{ item.SourceType }}
              </div>
              <div class="flex gap-4 items-center justify-end">
                {{ util.formatMoney(item.CurrentSettled) }}
                <button
                  class="text-primary"
                  @click="data.showIdx = data.showIdx == idx ? -1 : idx"
                >
                  <mdicon
                    :name="idx === data.showIdx ? 'chevron-up' : 'chevron-down'"
                    size="32"
                  />
                </button>
              </div>
            </div>
            <div class="w-full" v-if="idx === data.showIdx">
              <s-grid
                ref="grid"
                :config="data.gridCfg"
                hide-action
                hide-new-button
                hide-detail
                hide-select
                hide-search
                hide-sort
                hide-refresh-button
                hide-control
                hide-save-button
                v-model="item.records"
              >
              </s-grid>
            </div>
          </li>
        </ul>
      </div>
      <template #footer>
        <div class="mt-5 flex gap-3 justify-end">
          <div class="w-full h-[30px]" v-if="data.loading">
            <loader kind="skeleton" skeleton-kind="input" />
          </div>
          <template v-else>
            <s-button
              class="btn_warning"
              label="Back"
              @click="onClose"
            ></s-button>
            <s-button
              v-if="data.items.length > 0"
              class="btn_primary"
              label="Submit"
              @click="onSubmit"
            ></s-button>
          </template>
        </div>
      </template>
    </s-card>
  </s-modal>
</template>
<script setup>
import { reactive, computed, watch, onMounted, inject } from "vue";
import {
  SGrid,
  SInput,
  SModal,
  SButton,
  SCard,
  util,
  loadGridConfig,
} from "suimjs";

import Loader from "@/components/common/Loader.vue";

const axios = inject("axios");
const props = defineProps({
  sources: { type: Array, default: () => [] },
  calcAdjustment: { type: Function },
});
const emit = defineEmits({
  "update:modelValue": null,
  close: null,
  submit: null,
});
const data = reactive({
  gridCfg: {},
  items: [],
  loading: false,
  showIdx: 0,
  items: [],
});
function onClose() {
  emit("close");
}
function onSubmit() {
  data.loading = true;
  emit(
    "submit",
    function () {
      data.loading = false;
      emit("close");
    },
    function () {
      data.loading = false;
    }
  );
}
function hasSettled(mapApply) {
  let isFound = false;
  if (mapApply == undefined) return isFound;
  Object.keys(mapApply).forEach((mp) => {
    if (isFound === true) return;
    isFound = mapApply[mp].IsSettled;
  });
  return isFound;
}

function createDataItems() {
  data.items = JSON.parse(JSON.stringify(props.sources)).reduce((acc, obj) => {
    if (hasSettled(obj.MapApply) === false) return acc;

    obj.records = Object.keys(obj.MapApply).reduce((r, mp) => {
      if (obj.MapApply[mp].IsSettled === false) return r;

      r.push({
        ...obj.MapApply[mp],
        TotalAdjustment: props.calcAdjustment(obj.MapApply[mp].Adjustments),
      });
      return r;
    }, []);

    acc.push(obj);
    return acc;
  }, []);
}
onMounted(() => {
  loadGridConfig(axios, "fico/apply/confirm/gridconfig").then(
    (r) => {
      data.gridCfg = r;
      createDataItems();
    },
    (e) => util.showError(e)
  );
});
</script>