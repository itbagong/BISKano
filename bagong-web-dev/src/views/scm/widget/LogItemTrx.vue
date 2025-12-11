<template>
  <div class="px-1">
    <a
      v-if="!props.hideButton"
      href="#"
      @click="data.openContent = true"
      class="mt-1"
    >
      <mdicon
        name="information-outline"
        width="16"
        alt="edit"
        class="cursor-pointer hover:text-primary"
      />
    </a>
    <s-modal
      v-if="data.openContent"
      title="Log Item"
      class="model-reject"
      hide-buttons
      display
      ref="reject"
      @beforeHide="data.openContent = false"
    >
      <div class="min-w-[1000px] max-h-[600px] overflow-auto">
        <span v-if="data.logs.length > 0">Item Variant: {{ data.logs.length > 0 ? data.logs[0].ItemVarian : '' }}</span>
        <loader kind="skeleton" skeleton-kind="list" v-if="data.loading" />
        <s-grid
          v-else
          v-model="data.logs"
          ref="listControlTrx"
          class="w-full grid-line-items"
          hide-sort
          auto-commit-line
          no-confirm-delete
          form-keep-label
          hide-search
          :hide-new-button="true"
          :hide-refresh-button="true"
          :hide-delete-button="true"
          :hide-detail="true"
          :hide-action="true"
          :hide-select="true"
          :config="data.gridCfg"
        >
        </s-grid>
      </div>
    </s-modal>
  </div>
</template>
<script setup>
import { reactive, ref, watch, onMounted, inject } from "vue";
import { loadGridConfig, SButton, SModal, SGrid, SCard, util } from "suimjs";
import Loader from "@/components/common/Loader.vue";
import moment from "moment";
import { provide } from "vue";

const axios = inject("axios");
const props = defineProps({
  id: { type: String, default: "" },
  itemId: {type: String, default: ""},
  sku: {type: String, default: ""},
  gridConfig: { type: String, default: "" },
  showContent: { type: Boolean, default: false },
  hideButton: { type: Boolean, default: false },
});
const data = reactive({
  openContent: props.showContent === undefined ? false : props.showContent,
  loading: false,
  logs: [],
  gridCfg: {},
});
function fetchLog() {
  data.loading = true;
  data.logs = [];
  
  let payload = {
    ItemID: props.itemId,
    SKU: props.sku
  };
  console.log(payload);

  axios
    .post(`/scm/inventory/trx/get-display-prev`, payload)
    .then(
      (r) => {
        data.logs = r.data ?? [];
      },
      (err) => {
        util.showError(err);
      }
    )
    .finally(() => {
      data.loading = false;
    });
}

watch(
  () => data.openContent,
  (val) => {
    if (val) {
      fetchLog();
    }
  }
);
onMounted(() => {
  loadGridConfig(axios, props.gridConfig).then(
    (r) => {
      data.gridCfg = r;
      if (data.openContent === true) fetchLog();
    },
    (e) => util.showError(e)
  );
});
</script>
