<template>
  <routine-grid
    v-if="data.appMode == ''"
    :title="data.title"
    @selectdata="selectDataRoutine"
    @onPostAddNew="selectDataRoutine"
  />
  <routine-details-grid
    v-else-if="data.appMode == 'grid-details'"
    :title="'Routine Details'"
    :dataRoutine="data.selectedRoutine"
    @back="data.appMode = ''"
    @selectdata_detail="onSelectRoutineDetail"
  />
  <routine-asset-detail
    v-else-if="data.appMode == 'asset-details'"
    :title="'Asset Details'"
    :dataRoutine="data.selectedRoutine"
    :dataRoutineAsset="data.selectedRoutineAsset"
    @back="data.appMode = 'grid-details'"
    @post-save="updateRoutineAsset"
  />
</template>
<script setup>
import { reactive, ref, computed, inject, nextTick } from "vue";
import { layoutStore } from "@/stores/layout.js";
import RoutineGrid from "./widget/RoutineList.vue";
import RoutineDetailsGrid from "./widget/RoutineDetails.vue";
import RoutineAssetDetail from "./widget/RoutineAssetDetail.vue";

const axios = inject("axios");

layoutStore().name = "tenant";

const data = reactive({
  appMode: "",
  formMode: "edit",
  title: "Routine",
  selectedRoutine: {},
  selectedRoutineAsset: {},
});
function onSelectRoutineDetail(dt) {
  data.selectedRoutineAsset = dt;
  nextTick(() => {
    data.appMode = "asset-details";
  });
}
function selectDataRoutine(record) {
  axios
    .post(`/mfg/routine/gets?_id=${record._id}`, {
      Skip: 0,
      Take: 25,
      Sort: ["_id"],
    })
    .then(
      (r) => {
        data.selectedRoutine = r.data.data[0];
        data.appMode = "grid-details";
      },
      (e) => {
        util.showError(e);
      }
    );
}
function updateRoutineAsset(status, onRequest) {
  const payload = { ...data.selectedRoutineAsset, StatusCondition: status };
  axios.post("/mfg/routine/detail/save", payload).then(
    (r) => {
      if (onRequest) {
        onRequest();
      } else {
        data.appMode = "grid-details";
      }
    },
    (e) => util.showError(e)
  );
}
</script>
