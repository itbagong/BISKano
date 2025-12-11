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
      title="Log"
      class="model-reject"
      hide-buttons
      display
      ref="reject"
      @beforeHide="data.openContent = false"
    >
      <div class="min-w-[500px] max-h-[600px] overflow-auto">
        <loader kind="skeleton" skeleton-kind="list" v-if="data.loading" />
        <ul v-else class="flex flex-col">
          <li v-for="(log, idx) in data.logs" :key="idx" class="pb-3">
            <status-text :txt="log.Status" />
            <div class="flex justify-between text-[0.8rem] pl-4">
              <div>{{ log.Text }}</div>
              <div>
                {{
                  log.Date == null
                    ? ""
                    : moment(log.Date).format("DD-MMM-yyyy HH:mm:ss")
                }}
              </div>
            </div>
            <div class="pl-4">{{ log.Reason }}</div>
          </li>
        </ul>
      </div>
    </s-modal>
  </div>
</template>
<script setup>
import { reactive, ref, watch, onMounted, inject } from "vue";
import { SButton, SModal, SCard, util } from "suimjs";
import StatusText from "./StatusText.vue";
import Loader from "./Loader.vue";
import moment from "moment";

const axios = inject("axios");
const props = defineProps({
  id: { type: String, default: "" },
  showContent: { type: Boolean, default: false },
  hideButton: { type: Boolean, default: false },
});
const data = reactive({
  openContent: props.showContent === undefined ? false : props.showContent,
  loading: false,
  logs: [],
});
function fetchLog() {
  data.loading = true;
  data.logs = [];
  axios
    .post(`/fico/approvallog/get`, { ID: props.id })
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
  (nv) => {
    if (nv) fetchLog();
  }
);
onMounted(() => {
  if (data.openContent === true) fetchLog();
});
</script>
