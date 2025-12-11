<template>
  <Loader kind="skeleton" v-if="data.loading" />
  <div v-else>
    <status-text :txt="data.record.Status" />
    <s-form
      ref="frmCtl"
      v-model="data.record"
      :config="data.frmCfg"
      keep-label
      hide-cancel
      hide-submit
      :buttons-on-bottom="false"
      buttons-on-top
      :tabs="
        data.record.MCUTransactionID == '' ||
        data.record.MCUTransactionID == undefined
          ? ['General']
          : ['General', 'Attachment']
      "
      @submit-form="submitForm"
    >
      <template #tab_Attachment="{ item }">
        <s-grid-attachment
          :journalId="item.MCUTransactionID"
          journalType="SHE_MCU_RESULT"
          read-only
        />
      </template>
      <template #input_MCUTransactionID="{ item }">
        <div
          v-if="
            item.MCUTransactionID == '' || item.MCUTransactionID == undefined
          "
          class="h-[300px] flex justify-center items-center font-bold text-xl mt-3"
        >
          <div>SHE has not been created</div>
        </div>
        <div v-else>
          <div
            class="h-[300px] flex justify-center items-center font-bold text-xl mt-3"
          >
            <div
              class="circle_outer bg-green-100"
              v-if="data.mcuTrx?.MCUResult?.VisitResult == 'Fit'"
            >
              <div class="circle_inner bg-green-200 text-3xl text-green-900">
                Fit
              </div>
            </div>
            <div
              class="circle_outer bg-red-100"
              v-else-if="data.mcuTrx?.MCUResult?.VisitResult == 'UnFit'"
            >
              <div class="circle_inner bg-red-200 text-3xl text-red-900">
                UnFit
              </div>
            </div>
            <div class="circle_outer bg-gray-100" v-else>
              <div
                class="circle_inner bg-gray-200 text-2xl text-center text-gray-900"
              >
                Not Available
              </div>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-2">
            <s-input
              class="mt-2"
              v-if="data.mcuTrx?.MCUResult?.Notes"
              label="Visit Note"
              read-only
              v-model="data.mcuTrx.MCUResult.Notes"
            />
            <div v-else>
              <label class="input_label">Visit Note</label>
              <div class="bg-transparent italic">Data is not available</div>
            </div>
            <div v-if="data.mcuTrx?.FollowUp?.length > 0">
              <s-input
                class="mt-2"
                label="Follow Up Note"
                read-only
                v-model="
                  data.mcuTrx.FollowUp[data.mcuTrx.FollowUp.length - 1].Notes
                "
              />
            </div>
            <div v-else>
              <label class="input_label">Follow Up Note</label>
              <div class="bg-transparent italic">Data is not available</div>
            </div>
          </div>
        </div>
      </template>
    </s-form>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";

import StatusText from "@/components/common/StatusText.vue";
import Loader from "@/components/common/Loader.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import { util, SForm, SButton, SInput, loadFormConfig } from "suimjs";

const props = defineProps({
  id: { type: String, default: "" },
});

const axios = inject("axios");
const frmCtl = ref(null);

const data = reactive({
  frmCfg: {},
  record: {},
  mcuTrx: {},
  loading: false,
  loadingMcu: false,
});
watch(
  () => props.id,
  () => {
    fetchRecord();
  }
);

function fetchMcuTrx() {
  if (
    data.record.MCUTransactionID == "" ||
    data.record.MCUTransactionID == undefined
  ) {
    data.mcuTrx = {};
    return;
  }
  data.loadingMcu = true;
  axios
    .post("/she/mcutransaction/get", [data.record.MCUTransactionID])
    .then((r) => {
      data.mcuTrx = r.data;
    })
    .catch((e) => {
      data.mcuTrx = {};
      util.showError(e);
    })
    .finally(() => {
      data.loading = false;
    });
}
function fetchRecord() {
  data.loading = true;
  axios
    .post("/hcm/mcu/get", [props.id])
    .then((r) => {
      data.record = r.data;
      fetchMcuTrx();
    })
    .catch((e) => {
      data.record = {};
      util.showError(e);
    })
    .finally(() => {
      data.loading = false;
    });
}
function submitForm(record, cbOk, cbError) {
  axios
    .post("/hcm/mcu/update", record)
    .then((r) => {
      cbOk();
    })
    .catch((e) => {
      cbError();
      util.showError(e);
    });
}
onMounted(() => {
  loadFormConfig(axios, "/hcm/mcu/formconfig").then(
    (r) => {
      data.frmCfg = r;
      fetchRecord();
    },
    (e) => util.showError(e)
  );
});
</script>
<style scoped>
.circle_outer {
  @apply h-[180px] w-[180px] flex items-center justify-center  rounded-[180px];
}
.circle_inner {
  @apply h-[140px] w-[140px] flex items-center justify-center font-bold rounded-[140px];
}
</style>
