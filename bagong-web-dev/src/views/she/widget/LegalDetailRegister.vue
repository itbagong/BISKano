<template>
  <s-grid
    ref="gridCtl"
    v-model="data.record"
    no-gap
    auto-commit-line
    editor
    hide-action
    hide-search
    hide-sort
    hide-refresh-button
    hide-select
    hide-detail
    no-confirm-delete
    hide-footer
    :config="data.gridCfg"
    @newData="onNewRecord"
  >
    <template #item_Subject="{ item }">
      <div>
        <s-input
          kind="text"
          v-model="item.Subject"
          multi-row="3"
          class="w-full"
          @change="onRowFieldChanged"
        />
        <div class="grid grid-cols-2 gap-2 pb-4">
          <s-button
            label="add detail"
            tooltip="add detail"
            class="btn_success"
            @click="addDetail(item)"
          />
          <s-button
            label="delete"
            tooltip="delete"
            class="btn_primary"
            @click="deleteDetail(item)"
          />
        </div>
      </div>
    </template>
    <template #item_ActivityPoints="{ item }">
      <div class="w-full" v-for="(ap, idx) in item.ActivityPoints" :key="idx">
        <div class="flex gap-2">
          <s-input
            kind="text"
            v-model="ap.Value"
            multi-row="3"
            class="w-full"
            @change="onRowFieldChanged"
          />
          <s-toggle
            v-if="false"
            v-model="ap.IsComply"
            class="w-[120px] mt-0.5"
            yes-label="comply"
            no-label="not comply"
          />
        </div>
      </div>
    </template>
  </s-grid>
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
import { SInput, util, SButton, SGrid } from "suimjs";
import SToggle from "@/components/common/SButtonToggle.vue";
import helper from "@/scripts/helper.js";

const gridCtl = ref(null);

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
  countPlant: null,
});

const data = reactive({
  record: [],
  gridCfg: {},
});

function gridCfgDetail() {
  data.gridCfg = {
    setting: {
      idField: "",
      keywordFields: ["_id", "Subject"],
      sortable: ["_id"],
    },
    fields: [
      helper.gridColumnConfig({
        field: "Subject",
        label: "Subject",
        width: "500px",
      }),
      helper.gridColumnConfig({
        field: "ActivityPoints",
        label: "Activity Point",
      }),
    ],
  };
  data.record = props.modelValue.map((dt) => {
    dt.ID = util.uuid();
    return dt;
  });
  updateItems();
}

function onNewRecord(r) {
  r = {
    ID: util.uuid(),
    Subject: "",
    ActivityPoints: [
      {
        _id: util.uuid(),
        Value: "",
        IsComply: false,
      },
    ],
  };

  data.record.push(r);
  updateItems();
}

function updateItems() {
  nextTick(() => {
    gridCtl.value.setRecords(data.record);
  });
}

function addNewActivity(parent, item = { Value: "", IsComply: false }) {
  parent.push(item);
  calculatePlan();
}
function onRowFieldChanged() {
  nextTick(() => {
    emit("update:modelValue", data.record);
  });
}

function deleteData(item, idx) {
  item.splice(idx, 1);
  calculatePlan();
}

function onDeleteSubject(id) {
  let filtered = helper.cloneObject(data.record.filter((o) => o.ID !== id));
  data.record = filtered;
  onRowFieldChanged();
  updateItems();
  calculatePlan();
}

function calculatePlan() {
  let res = data.record.reduce((acc, item) => {
    return (acc += item.ActivityPoints.length);
  }, 0);

  emit("countPlant", res);
}

function addDetail(r) {
  r.ActivityPoints.push({
    _id: util.uuid(),
    Value: "",
    IsComply: false,
  });
}

function deleteDetail(r) {
  data.record = data.record.filter((v) => {
    return v.ID != r.ID;
  });
  updateItems();
}

onMounted(() => {
  gridCfgDetail();
});
</script>
