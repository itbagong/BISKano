<template>
  <div
    class="grid md:grid-cols-12 grid-cols-1 gap-2 divide-x [&>*]:p-2 min-h-[calc(100vh-280px)]"
  >
    <div class="col-span-2">
      <StageTracking
        ref="stepCtl"
        v-model="data.selectedTrack"
        @select="selectTrack"
        @refresh="refresh"
      />
    </div>
    <div class="col-span-10">
      <div v-if="data.selectedTrack == 'Mapping'">
        <Mapping :employeeID="recordTalent.EmployeeID"></Mapping>
      </div>
      <div v-if="data.selectedTrack == 'Assesment'">
        <Assesment :talentDevelopmentID="recordTalent._id"></Assesment>
      </div>
      <div v-if="data.selectedTrack == 'SKActing'">
        <SK form-config="/hcm/talentdevelopmentsk/acting/formconfig" :talentDevelopmentID="recordTalent._id" :employeeID="recordTalent.EmployeeID" type="ACTING" journal-type-lookup-url="/hcm/journaltype/find?TransactionType=Talent%20Development%20-%20Promotion%20-%20Tracking%20SK%20Acting"></SK>
      </div>
      <div v-if="data.selectedTrack == 'SKTetap'">
        <SK form-config="/hcm/talentdevelopmentsk/permanent/formconfig" :talentDevelopmentID="recordTalent._id" :employeeID="recordTalent.EmployeeID" type="PERMANENT" journal-type-lookup-url="/hcm/journaltype/find?TransactionType=Talent%20Development%20-%20Promotion%20-%20Tracking%20SK%20Tetap"></SK>
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, watch, inject } from "vue";
import { DataList, SButton, util } from "suimjs";
import StageTracking from "./StageTracking.vue";
import Mapping from "./Tracking/Mapping.vue";
import Assesment from "./Tracking/Assesment.vue"
import SK from "./Tracking/SK.vue"

const stepCtl = ref(null);

const props = defineProps({
  recordTalent: { type: Object, default: {} },
});
const data = reactive({
  selectedTrack: "",
  // pshycoTab: ["General", "Material"],
  // selectedPshycoTab: "General",
  records: [],
});
function selectTrack(id) {
  data.selectedTrack = id;
  // console.log(id)
}

function refresh() {
  stepCtl.value?.refresh();
}
</script>
<style></style>
