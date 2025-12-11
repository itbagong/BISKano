<template>
  <Loader kind="skeleton" v-if="data.loading" />
  <div v-else>
    <status-text :txt="data.record.Status" />
    <s-form
      mode="view"
      ref="frmCtl"
      v-model="data.record"
      :config="data.frmCfg"
      keep-label
      hide-cancel
      hide-submit
      :buttons-on-bottom="false"
      buttons-on-top
      @submit-form="submitForm"
    >
      <template #input_TrainingStatus="{ item }">
        <div>
          <label class="input_label">Training status</label>
          <div class="bg-transparent">
            {{ item.TrainingStatus ? "Open" : "Close" }}
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

import { util, SForm, SButton, loadFormConfig } from "suimjs";

const props = defineProps({
  id: { type: String, default: "" },
});

const axios = inject("axios");
const frmCtl = ref(null);

const data = reactive({
  frmCfg: {},
  record: {},
  loading: "",
});
watch(
  () => props.id,
  () => {
    fetchRecord();
  }
);
function fetchRecord() {
  data.loading = true;
  axios
    .post("/hcm/training/get", [props.id])
    .then((r) => {
      data.record = r.data;
    })
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loading = false;
    });
}
function submitForm(record, cbOk, cbError) {
  axios
    .post("/hcm/training/update", record)
    .then((r) => {
      cbOk();
    })
    .catch((e) => {
      cbError();
      util.showError(e);
    });
}
function filterFormCfg(cfg) {
  return {
    ...cfg,
    sectionGroups: cfg.sectionGroups.map((group) => ({
      ...group,
      sections: group.sections.map((section) => ({
        ...section,
        rows: section.rows.filter((row) =>
          row.inputs.some((field) => field.field === "TrainingStatus")
        ),
      })),
    })),
  };
}
onMounted(() => {
  loadFormConfig(axios, "/hcm/training/formconfig").then(
    (r) => {
      data.frmCfg = filterFormCfg(r);
      fetchRecord();
    },
    (e) => util.showError(e)
  );
});
</script>
