<template> 
  <div class="w-full bg-white p-3">
    <div class="grid grid-cols-5 gap-4 mb-5">
      <div class="col-span-4">
        <s-input
          label="Cash Bank"
          v-model="data.cashbankId"
          use-list
          lookup-url="/tenant/cashbank/find"
          lookup-key="_id"
          :lookup-labels="['Name', '_id']"
          :lookup-searchs="['_id', 'Name']"
          @change="onChangeCashBank"
        />
      </div>
      <div>
        <div
          class="mt-3"
          v-if="
            data.loadingFetchRecon || data.loadingStart || data.loadingRecon
          "
        >
          <loader kind="skeleton" skeleton-kind="input" />
        </div>
        <s-button
          v-else
          class="btn_primary mt-3"
          :disabled="disableRecon"
          label="Start Reconcile"
          @click="onStart"
        />
      </div>
      <s-input
        label="Actual Balance"
        v-model="data.recon.ReconBalance"
        kind="number"
        read-only
      />

      <s-input
        label="Reconciliation Balance"
        v-model="data.inputBalance"
        kind="number"
        :read-only="disableRecon || data.loadingFetchRecon === true"
      />
      <s-input label="Reconcile Date" v-model="data.date" read-only />
      <s-input label="Last Reconcile Date" v-model="data.date" read-only />
      <s-input
        label="Last Balance"
        v-model="data.recon.PreviousBalance"
        kind="number"
        read-only
      />
    </div>
    <div class="border-t pt-5">
      <div class="flex justify-between items-center mb-4">
        <div class="flex gap-4 items-center">
          <div class="text-right">
            <label class="input_label">
              <div>Difference</div>
            </label>
            <div class="bg-transparent">
              {{ util.formatMoney(diference) }}
            </div>
          </div>

          <!-- <s-input
            v-model="data.diference"
            label="Differnce"
            kind="number"
            disabled
            class="min-w-[220px]"
          /> -->
          <s-button
            :disabled="!data.isStart"
            class="btn_primary mt-3"
            label="Adjustment"
          />
        </div>

        <div class="mt-3 w-[100px]" v-if="data.loadingRecon">
          <loader kind="skeleton" skeleton-kind="input" />
        </div>
        <s-button
          v-else
          :disabled="!data.isStart"
          @click="onReconcile"
          class="btn_primary mt-3"
          label="Reconcile"
        />
      </div>
      <div class="w-full mb-3">
        <data-list
          grid-hide-control
          grid-hide-detail
          grid-hide-select
          no-gap
          class="list-bank-recon"
          init-app-mode="grid"
          grid-mode="grid"
          ref="listTrx"
          grid-config="/fico/cashrecongrid/gridconfig"
          :grid-fields="['_id', 'SourceType']"
        >
          <template #grid_SourceType="{ item }">
            <div class="flex gap-3 items-center">
              <input type="checkbox" @change="onCheckTrx(item)" />
              {{ item.SourceType }}
            </div>
          </template>
        </data-list>
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed } from "vue";
import { DataList, SForm, SInput, util, SButton } from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";


import moment from "moment";
import Loader from "@/components/common/Loader.vue";
 
layoutStore().name = "tenant";

const FEATUREID = 'BankReconciliation'
const profile = authStore().getRBAC(FEATUREID)



const auth = authStore();

const axios = inject("axios");
const listTrx = ref(null);
const data = reactive({
  cashbankId: "",
  recon: {
    CashBankID: "",
    CompanyID: "",
    Created: "",
    Diff: 0,
    LastUpdate: "",
    Lines: [],
    Name: "",
    PreviousBalance: 0,
    PreviousReconDate: "",
    ReconBalance: 0,
    ReconDate: "",
    ReconJournalTypeID: "",
    Status: "",
    _id: "",
  },
  inputBalance: 0,
  balance: 0,
  date: moment().format("DD MMM YYYY"),
  diference: 0,
  loadingFetchRecon: false,
  loadingStart: false,
  loadingRecon: false,
  listTrx: [],
  isStart: false,
});
const diference = computed({
  get() {
    return data.recon.ReconBalance - data.inputBalance;
  },
});

const disableRecon = computed({
  get() {
    return data.cashbankId == "" || data.cashbankId == null;
  },
});

const selectedTrx = computed({
  get() {
    return data.listTrx.filter((e) => e.isSelected === true);
  },
});
function onCheckTrx(item) {
  if (item.isSelected == undefined || item.isSelected == false)
    item.isSelected = true;
  else item.isSelected = false;

  const idx = data.listTrx.findIndex((e) => e._id == item._id);
  data.listTrx[idx] = { ...item };
}
async function fetchReconCurrent(param) {
  let result = {};
  try {
    const r = await axios.post("/fico/cashrecon/get-current-recon", param);
    result = r.data;
  } catch (e) {
    util.showError(e);
  }
  return result;
}
async function fetchReconPrevious(param) {
  let result = {};

  try {
    const r = await axios.post("/fico/cashrecon/get-previous-recon", param);
    result = r.data;
  } catch (e) {
    util.showError(e);
  }
  return result;
}
async function fetchCurentBalance(param) {
  let balance = 0;

  try {
    const r = await axios.post("/fico/cashbalance/get-current", param);
    balance = r.data.Balance;
  } catch (e) {
    util.showError(e);
  }
  return balance;
}
async function getRecon() {
  data.loadingFetchRecon = true;
  const param = {
    CompanyID: auth.companyId,
    CashBankID: data.cashbankId,
  };
  // data.balance = await fetchCurentBalance(param);

  const current = await fetchReconCurrent(param);

  const previous = await fetchReconPrevious(param);

  if (current._id == "") data.recon = { ...data.recon, ...previous };
  else data.recon = { ...data.recon, ...current };
  data.loadingFetchRecon = false;
}
async function fetchSaveRecon(param) {
  let result = {};

  try {
    const r = await axios.post("/fico/cashrecon/save", param);
    result = r.data;
  } catch (e) {
    util.showError(e);
  }
  return result;
}
async function fetchStartRecon(param) {
  let result = {};
  try {
    const r = await axios.post("/fico/cashrecon/start-recon", param);
    result = r.data;
  } catch (e) {
    util.showError(e);
  }
  return result;
}
async function getTrasaction() {
  const param = {
    CompanyID: auth.companyId,
    CashBankID: data.cashbankId,
    ReconDate: new Date(),
  };
  try {
    const r = await axios.post("/fico/cashrecon/get-transactions", param);
    data.listTrx = r.data;
    listTrx.value.setGridRecords(data.listTrx);
  } catch (e) {
    listTrx.value.setGridRecords([]);
    util.showError(e);
  }
}
async function onStart() {
  data.isStart = false;
  data.loadingStart = true;
  let param = {
    ...data.recon,
    ...{
      ReconDate: new Date(),
      ReconBalance: data.inputBalance,
      CompanyID: auth.companyId,
      CashBankID: data.cashbankId,
    },
  };
  if (moment(data.recon.ReconDate).isSame(moment(), "day") === false) {
    param = await fetchSaveRecon(param);
  }
  data.recon = await fetchStartRecon(param);
  await getTrasaction();

  data.isStart = true;
  data.loadingStart = false;
}
async function onReconcile() {
  data.loadingRecon = true;
  const param = {
    CashReconID: data.recon._id,
    CashTransactionIDs: selectedTrx.value.map((e) => e._id),
  };
  try {
    const r = await axios.post("/fico/cashrecon/reconcile", param);
    util.showInfo("Data has been saved");
    reset();
  } catch (e) {
    util.showError(e);
  }
  data.loadingRecon = false;
}
function reset() {
  data.recon = {
    CashBankID: "",
    CompanyID: "",
    Created: "",
    Diff: 0,
    LastUpdate: "",
    Lines: [],
    Name: "",
    PreviousBalance: 0,
    PreviousReconDate: "",
    ReconBalance: 0,
    ReconDate: "",
    ReconJournalTypeID: "",
    Status: "",
    _id: "",
  };
  data.inputBalance = 0;
  data.cashbankId = "";
  data.isStart = false;
  data.diference = 0;
  data.listTrx = [];
  listTrx.value.setGridRecords([]);
}
function onChangeCashBank(name, v1) {
  if (v1 == "" || v1 == null) return;
  util.nextTickN(2, () => {
    getRecon();
  });
}

onMounted(() => {});
</script>
<style>
.list-bank-recon .suim_grid  > div > div {
  max-height: 600px;
  overflow: auto;
}

.list-bank-recon table.suim_table  thead > tr{
  position: sticky; top: 0
}

.list-bank-recon table.suim_table  tbody{
  position: relative;
}

</style>