<template>
  <cash-journal
    :type="cfg.type"
    :featureID="cfg.featureID"
    :is-submission="cfg.isSubmission"
    :journal-type="cfg.journalType"
    url-line-grid-config="fico/cashin/line/gridconfig"
    :key="cfg.type"
  />
</template>

<script setup>
import { authStore } from "@/stores/auth";
import { layoutStore } from "@/stores/layout.js";
import CashJournal from "./widget/Cash/Journal.vue";
import { useRoute } from "vue-router";
import { computed } from "vue";

layoutStore().name = "tenant";
const route = useRoute();

const cfg = computed({
  get() {
    switch (route.query.id) {
      case "SubmissionCashIn":
        return {
          type: "SUBMISSION CASH IN",
          featureID: "SubmissionCashIn",
          journalType: "SUBMISSION CASH IN",
          isSubmission: true,
        };
      default:
        return {
          type: "CASH IN",
          featureID: "CashIn",
          journalType: "CASH IN",
          isSubmission: false,
        };
    }
  },
});
</script>
