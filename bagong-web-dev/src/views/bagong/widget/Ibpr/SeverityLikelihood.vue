<template>
  <data-list
    ref="listControl"
    hide-title
    no-gap
    grid-editor
    grid-hide-search
    grid-hide-sort
    grid-hide-refresh
    grid-hide-detail
    grid-hide-select
    grid-no-confirm-delete
    grid-sort-field="Level"
    grid-sort-direction="asc"
    init-app-mode="grid"
    grid-mode="grid"
    :grid-config="`/bagong/${kind}/gridconfig`"
    :grid-read="`/bagong/${kind}/gets?Type=${type}`"
    new-record-type="grid"
    grid-auto-commit-line
    :grid-fields="['Level']"
    @grid-row-add="newRecord"
    @grid-row-delete="onGridRowDelete"
    @grid-row-save="onGridRowSave"
    @grid-refreshed="onGridRefreshed"
    :grid-hide-new="!profile.canCreate"
    :grid-hide-edit="!profile.canUpdate"
    :grid-hide-delete="!profile.canDelete"
  >
    <template #grid_header_buttons_1>
      <s-button
        icon="content-save"
        class="btn_primary"
        @click="onSave"
        v-if="profile.canCreate"
      />
    </template>
  </data-list>
</template>
<script setup>
import { authStore } from "@/stores/auth.js";
import { reactive, onMounted, inject, ref, nextTick } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { SCard, util, DataList, SButton, SInput } from "suimjs";
layoutStore().name = "tenant";
const axios = inject("axios");
const auth = authStore();

const props = defineProps({
  kind: { type: String, default: "" },
  profile: { type: Object, default: () => {} },
  type: { type: String, default: "" },
});

const data = reactive({
  records: [],
});

const listControl = ref(null);

function newRecord(r) {
  r.ID = "";
  r.Value = 0;
  data.records.push(r);
  updateItems();
}

function onGridRowDelete(_, index) {
  const deletedID = data.records.find((dt, idx) => {
    return idx == index;
  });
  onDelete(deletedID._id);
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  updateItems();
}

function onGridRowSave(record, index) {
  record.suimRecordChange = false;
  data.records[index] = record;
  updateItems();
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
}

function setRecords() {
  data.records = listControl.value.getGridRecords();
}

function onSave(params) {
  const url = "/bagong/" + props.kind + "/save";
  for (let i in data.records) {
    let o = data.records[i];
    o.CompanyID = auth.companyId;
    o.Type = props.type;
    axios.post(url, o).then(
      (r) => {
        const idx = parseInt(i) + 1;
        if (idx == data.records.length) util.showInfo("data has been saved");
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function onDelete(id) {
  const url = "/bagong/" + props.kind + "/delete";
  axios.post(url, { _id: id }).then(
    (r) => {
      util.showInfo("data has been deleted");
    },
    (e) => {
      util.showError(e);
    }
  );
}

function onGridRefreshed(r) {
  setRecords();
}
</script>
