<template>
  <cash-journal
    :type="cfg.type"
    :featureID="cfg.featureID"
    :is-submission="cfg.isSubmission"
    :journal-type="cfg.journalType"
    url-line-grid-config="fico/cashout/line/gridconfig"
    :key="cfg.type"
  />
</template>

<script setup>
import { authStore } from "@/stores/auth";
import { layoutStore } from "@/stores/layout.js";
import CashJournal from "./widget/Cash/Journal.vue";
import { useRoute } from "vue-router";
import { computed } from "vue";
import { component } from "v-viewer";

layoutStore().name = "tenant";
const route = useRoute();

const cfg = computed({
  get() {
    switch (route.query.id) {
      case "SubmissionCashOut":
        return {
          type: "SUBMISSION CASH OUT",
          featureID: "SubmissionCashOut",
          journalType: "SUBMISSION CASH OUT",
          isSubmission: true,
        };
      default:
        return {
          type: "CASH OUT",
          journalType: "CASH OUT",
          featureID: "CashOut",
          isSubmission: false,
        };
    }
  },
});
</script>
