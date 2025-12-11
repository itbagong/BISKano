<template>
  <s-card class="w-full bg-white suim_datalist card" hide-footer>
    <div class="flex mb-4">
      <s-input
        label="Type"
        use-list
        :items="['IBPR', 'RSCA']"
        v-model="data.type"
        class="w-[300px]"
      />
    </div>
    <s-tab
      :tabs="mapTabByType[data.type]"
      @activeTab="getActiveTab"
      ref="customTab"
    >
      <template #tab_Severity_body>
        <addable
          kind="severity"
          :profile="profile"
          :type="data.type"
          v-if="data.type"
        />
      </template>
      <template #tab_Impact_body>
        <addable
          kind="severity"
          :profile="profile"
          :type="data.type"
          v-if="data.type"
        />
      </template>
      <template #tab_Likelihood_body>
        <addable
          kind="likelihood"
          :profile="profile"
          :type="data.type"
          v-if="data.type"
        />
      </template>
      <template #tab_Risk_Matrix_body
        ><matrix
          ref="matrixRisk"
          :profile="profile"
          :type="data.type"
          v-if="data.type"
      /></template>
    </s-tab>
  </s-card>
</template>
<script setup>
import { reactive, onMounted, inject, ref, nextTick, watch } from "vue";
import { SCard, SInput } from "suimjs";
import STab from "@/components/common/STab.vue";
import Addable from "./widget/Ibpr/SeverityLikelihood.vue";
import Matrix from "./widget/Ibpr/Matrix.vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";

layoutStore().name = "tenant";
const FEATUREID = "MIBPR";
const profile = authStore().getRBAC(FEATUREID);

const matrixRisk = ref(null);
const customTab = ref(null);

function getActiveTab(name) {
  data.tabName = name;
  if (name == "Risk Matrix" && data.type) {
    matrixRisk.value.getMatrixList();
    matrixRisk.value.getSources();
  }
}

const mapTabByType = {
  IBPR: ["Severity", "Likelihood", "Risk Matrix"],
  RSCA: ["Impact", "Likelihood", "Risk Matrix"],
};

const data = reactive({
  type: "",
  tabName: "",
});

watch(
  () => data.type,
  (nv) => {
    customTab.value.setCurrentTab(0);
  }
);
</script>
