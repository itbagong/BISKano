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
      :buttons-on-bottom="false"
      buttons-on-top
      :tabs="['General', 'Attachment']"
      @submit-form="submitForm"
    >
      <template #tab_Attachment="{ item }">
        <s-grid-attachment
          ref="attchCtl"
          :journalId="item._id"
          journalType="HCMOLPlottings"
          is-single-upload
        />
      </template>
    </s-form>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";

import SGridAttachment from "@/components/common/SGridAttachment.vue";
import StatusText from "@/components/common/StatusText.vue";
import Loader from "@/components/common/Loader.vue";

import { util, SForm, SButton, loadFormConfig } from "suimjs";

const props = defineProps({
  id: { type: String, default: "" },
});

const axios = inject("axios");
const frmCtl = ref(null);
const attchCtl = ref(null);

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
    .post("/hcm/olplotting/get", [props.id])
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
    .post("/hcm/olplotting/update", record)
    .then((r) => {
      attchCtl.value?.Save();
      cbOk();
    })
    .catch((e) => {
      cbError();
      util.showError(e);
    });
}
function onAlterFormConfig(cfg) {
  cfg.sectionGroups = cfg.sectionGroups.map((sectionGroup) => {
    sectionGroup.sections = sectionGroup.sections.map((section) => {
      section.rows.map((row) => {
        row.inputs = row.inputs.filter(
          (input) => !["OfferingLetter"].includes(input.field)
        );
        return row;
      });
      return section;
    });
    return sectionGroup;
  });
}
onMounted(() => {
  loadFormConfig(axios, "/hcm/olplotting/formconfig").then(
    (r) => {
      r.sectionGroups = r.sectionGroups.map((sectionGroup) => {
        sectionGroup.sections = sectionGroup.sections.filter((section) => section.name == 'General');
        return sectionGroup;
      });
      data.frmCfg = r;
      fetchRecord();
    },
    (e) => util.showError(e)
  );
});
</script>
