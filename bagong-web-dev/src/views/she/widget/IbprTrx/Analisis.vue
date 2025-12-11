<template>
  <div class="overflow-x-scroll relative h-[650px]">
    <s-grid
      :class="`trx_IBPR_Analisis_${tabId}`"
      ref="listControl"
      :config="data.cfg"
      hide-search
      hide-sort
      hide-refresh-button
      hide-select
      hide-footer
      hide-new-button
      hide-action
      editor
      auto-commit-line
      no-confirm-delete
      @new-data="newLine"
      @delete-data="onDeleteDetail"
      @row-Field-Changed="onChangeRow"
    >
      <template #item_SituasiAktivitas="{ item }">
        <s-input
          v-model="item.SituasiAktivitas"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=IBPRActivity`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_RutinNonRutin="{ item }">
        <s-input
          v-model="item.RutinNonRutin"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=IBPRRoutine`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_NormalAbnormalEmergency="{ item }">
        <s-input
          v-model="item.NormalAbnormalEmergency"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=IBPRStatus`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_PeraturanTerkait="{ item }">
        <s-input
          v-model="item.PeraturanTerkait"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/she/legalregister/find`"
          :lookup-labels="['LegalNo']"
          :lookup-searchs="['LegalNo']"
        />
      </template>
      <template #item_Lingkup="{ item }">
        <s-input
          v-model="item.Lingkup"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=IBPRType`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_BahayaK3LAspekLingkungan="{ item }">
        <s-input
          v-model="item.BahayaK3LAspekLingkungan"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=IBPRAspek`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_ResikoK3LDampakLingkungan="{ item }">
        <s-input
          v-model="item.ResikoK3LDampakLingkungan"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=IBPRDampak`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_CreatedBy="{ item }">
        <s-input
          v-model="item.CreatedBy"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/employee/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_UpdatedBy="{ item }">
        <s-input
          v-model="item.UpdatedBy"
          use-list
          read-only
          lookup-key="_id"
          :lookup-url="`/tenant/employee/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['Name']"
        />
      </template>
      <template #item_CurrentActions="{ item }">
        <current-action
          v-model="item.CurrentActions"
          :item="item"
          :cfg="currentActionCfg"
          :scope="item.Lingkup"
          @field-change="onChangeRow"
        />
      </template>
      <template #item_RiskRating="{ item }">
        <div :class="[`tab_${tabId}_matrix_div_${item.ID}`]">
          {{ calMatrix(item, `tab_${tabId}_matrix_div_${item.ID}`) }}
          <div
            class="flex gap-4 font-semibold h-full p-4 h-full"
            v-if="item.Severity && item.Probability"
          >
            <div>{{ riskMatrix[item.RiskRating].RiskID }}</div>
            <div>({{ riskMatrix[item.RiskRating].Value }})</div>
          </div>
        </div>
      </template>
      <template #item_Validasi="{ item }">
        <uploader
          ref="gridAttachment"
          :journalId="`${jurnalId}#${tabId}#${item.ID}`"
          :config="{}"
          journalType="IBPR_TRX"
          single-save
          @modal="onModal"
        />
      </template>
      <template #item_PeluangTeridentifikasi="{ item }">
        <s-toggle
          v-model="item.PeluangTeridentifikasi"
          class="w-[80px] mt-0.5"
        />
      </template>
    </s-grid>
  </div>
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

import {
  SInput,
  util,
  SButton,
  SGrid,
  SCard,
  DataList,
  SModal,
  loadGridConfig,
} from "suimjs";
import currentAction from "./CurrentAction.vue";
import helper from "@/scripts/helper.js";
import Uploader from "@/components/common/Uploader.vue";
import SToggle from "@/components/common/SButtonToggle.vue";

const axios = inject("axios");
const listControl = ref(null);

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  masterIbpr: { type: Array, default: () => [] },
  formatRowSpan: { type: Function },
  currentActionCfg: { type: Object, default: {} },
  riskMatrix: { type: Object, default: {} },
  calMatrix: { type: Function },
  setBgMatrix: { type: Function },
  tabId: { type: Number, default: 0 },
  jurnalId: { type: String, default: 0 },
});
const data = reactive({
  appMode: "grid",
  record: props.modelValue ?? [],
  cfg: {},
});

const mapTab = {
  Initial: "/she/ibprtrx/initialrisk/gridconfig",
  Residual: "/she/ibprtrx/residualrisk/gridconfig",
  Opportunity: "/she/ibprtrx/opportunity/gridconfig",
};

const emit = defineEmits({
  "update:modelValue": null,
  rowChange: null,
  modal: null,
});

function refreshGrid() {
  util.nextTickN(3, () => {
    listControl.value.setRecords(data.record);
    util.nextTickN(3, () => {
      if (data.record.length > 0)
        props.formatRowSpan(`trx_IBPR_Analisis_${props.tabId}`);
    });
  });
}

function buildInitialRisk() {
  let dt = helper.cloneObject(props.masterIbpr);
  if (data.record.length == 0) {
    let objInitialRisk = {
      CurrentActions: {},
      Severity: "",
      Probability: "",
      RiskRating: "",
    };
    for (let i in dt) {
      dt[i] = { ...dt[i], ...helper.cloneObject(objInitialRisk) };
    }
    data.record = dt;
  } else {
    for (let i in dt) {
      let cur = data.record[i];
      dt[i] = { ...dt[i], ...cur };
    }
    data.record = dt;
  }
  refreshGrid();
}

function onChangeRow(name, v1, v2, current) {
  emit("rowChange", name, v1, v2, current);
}

function onModal(nv) {
  emit("modal", nv);
}

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  }
);

watch(
  () => props.masterIbpr,
  (nv) => {
    buildInitialRisk();
  }
);

onMounted(() => {
  loadGridConfig(axios, mapTab[props.tabId]).then(
    (r) => {
      data.cfg = r;
      buildInitialRisk();
    },
    (e) => {}
  );
});
</script>
<style>
.trx_IBPR_Analisis_Residual tr td:not(:first-child),
.trx_IBPR_Analisis_Opportunity tr td:not(:first-child) {
  min-width: 150px;
}

.trx_IBPR_Analisis_Opportunity tr td:nth-child(12),
.trx_IBPR_Analisis_Opportunity tr td:nth-child(16) {
  min-width: 250px;
}

.trx_IBPR_Analisis_Initial tr th,
.trx_IBPR_Analisis_Residual tr th,
.trx_IBPR_Analisis_Opportunity tr th {
  @apply sticky top-0 bg-gray-100;
}
</style>
