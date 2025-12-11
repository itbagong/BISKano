<template>
  <div class="mt-5">
    <data-list
      ref="listControl"
      hide-title
      no-gap
      grid-hide-control
      grid-hide-select
      grid-hide-detail
      grid-hide-delete
      :init-app-mode="data.appMode"
      :grid-mode="data.appMode"
      :grid-fields="['CashBank']"
      grid-config="/fico/cashtransaction/gridconfig"
      form-config="/fico/cashtransaction/formconfig"
    >
      <template #grid_CashBank="{ item }">
        <div>{{ item.CashBank?.Name }}</div>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, inject } from "vue";
import { DataList, loadGridConfig, util } from "suimjs";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  items: { type: Array, default: () => [] },
});

const listControl = ref(null);

const axios = inject("axios");

const data = reactive({
  appMode: "grid",
});

function updateItems() {
  listControl.value.setGridRecords(props.items);
}

onMounted(() => {
  loadGridConfig(axios, "/fico/cashtransaction/gridconfig").then(
    (r) => {
      util.nextTickN(2, () => {
        setTimeout(() => {
          const hideFields = [
            "_id",
            "SourceJournalID",
            "SourceLineNo",
            "CompanyID",
            "Dimension",
            "CashReconID",
          ];
          hideFields.map((field) => {
            listControl.value.removeGridField(field);
          });
          updateItems();
        }, 500);
      });
    },
    (e) => util.showError(e)
  );
});
</script>
