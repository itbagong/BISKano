<template>
  <div class="w-full">
    <data-list
      title="Cheque & Giro"
      ref="listControl"
      grid-hide-sort
      grid-hide-detail
      grid-hide-select
      grid-no-confirm-delete
      :init-app-mode="data.appMode"
      :grid-mode="data.appMode"
      new-record-type="grid"
      :grid-config="'/fico/cheque/gridconfig'"
      grid-read="/fico/cheque/gets"
      grid-auto-commit-line
      grid-hide-new
      grid-hide-delete
      :grid-fields="['CashBookID', 'Kind']"
    >
      <template #grid_item_buttons="{ item }">
        <div class="flex justify-center w-24" v-if="item.Status == 'Reserved'">
          <s-button
            class="btn_primary new_btn"
            label="Release Fund"
            @click="releaseFund(item)"
          />
        </div>
      </template>
      <template #grid_CashBookID="{ item }">
        <s-input
          v-model="item.CashBookID"
          hide-label
          use-list
          lookup-url="/tenant/cashbank/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :read-only="true"
        />
      </template>
      <template #grid_Kind="{ item }">
        <s-input
          v-model="item.Kind"
          hide-label
          use-list
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=CGT"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :read-only="true"
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject } from "vue";
import { DataList, util, SInput, SButton, SModal } from "suimjs";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";

layoutStore().name = "tenant";

const FEATUREID = 'Administrator'
const profile = authStore().getRBAC(FEATUREID)

const listControl = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
});

function releaseFund(record) {
  const url = "/fico/cg/release-fund";
  axios.post(url, record).then(
    (r) => {
      listControl.value.refreshGrid();
    },
    (e) => util.showError(e)
  );
}
</script>
