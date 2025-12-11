<template>
  <s-grid
    class="ExternalReportLine"
    ref="ExternalReportLine"
    :config="data.cfgGridExternalReportLine"
    hide-search
    hide-sort
    hide-refresh-button
    hide-edit
    hide-select
    hide-paging
    editor
    auto-commit-line
    no-confirm-delete
    @new-data="newExternalReport"
  >
    <template #item_NeedReport="{ item, idx }">
      <s-toggle
        v-model="item.NeedReport"
        class="w-[100px] mt-0.5"
        yes-label="Yes"
        no-label="No"
      />
    </template>
    <template #item_button_delete="{ item, idx }">
      <a @click="deleteExternalReport(item)" class="delete_action">
        <mdicon
          name="delete"
          width="16"
          alt="delete"
          class="cursor-pointer hover:text-primary"
        />
      </a>
    </template>
  </s-grid>
</template>
<script setup>
import { onMounted, inject, reactive, ref } from "vue";
import { loadGridConfig, util, SGrid } from "suimjs";
import SToggle from "@/components/common/SButtonToggle.vue";
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const ExternalReportLine = ref(null);

const data = reactive({
  record: props.modelValue,
  cfgGridExternalReportLine: {},
});

function loadGridExternalReportLine() {
  let url = `/she/investigasi/externalreport/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridExternalReportLine = r;
      updateGridLine(data.record.ExternalReport, "ExternalReport");
    },
    (e) => {}
  );
}

function newExternalReport() {
  let r = {};
  const noLine = data.record.ExternalReport.length + 1;
  r._id = util.uuid();
  r.LineNo = noLine;
  r.Date = new Date();
  r.Reporter = "";
  r.ThirdParty = "";
  r.NeedReport = true;
  r.PidThirdParty = "";
  r.Remark = "";
  data.record.ExternalReport.push(r);
  updateGridLine(data.record.ExternalReport, "ExternalReport");
}
function deleteExternalReport(r) {
  data.record.ExternalReport = data.record.ExternalReport.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.ExternalReport, "ExternalReport");
}
function updateGridLine(record, type) {
  record.map((obj, idx) => {
    obj.LineNo = parseInt(idx) + 1;
    return obj;
  });
  if (type == "ExternalReport") {
    ExternalReportLine.value.setRecords(record);
  }
}

onMounted(() => {
  loadGridExternalReportLine();
});
defineExpose({});
</script>
