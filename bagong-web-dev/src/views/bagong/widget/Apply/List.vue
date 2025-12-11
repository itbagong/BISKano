<template>
  <div class="flex flex-col gap-y-4">
    <div class="bg-white p-2 flex gap-3 items-center">
      <AccountSelector
        v-if="kind === 'apply'"
        class="mb-3 grow"
        v-model="data.OffsetAccount"
      ></AccountSelector>
      <div class="grow" v-else>
        <button
          class="bg-white h-[2rem] flex justify-center items-center"
          @click="emit('back')"
        >
          <mdicon name="chevron-left" size="23" />Back
        </button>
      </div>
      <s-button icon="refresh" class="btn_primary" @click="refreshData" />

      <div class="flex gap-3 items-center">
        <s-button
          class="btn_primary"
          label="Confirm"
          @click="data.showConfirm = true"
        />
      </div>
    </div>

    <div class="grid grid-cols-2 gap-4 apply-list">
      <div class="bg-white p-2">
        <div
          v-if="data.loadingSource"
          class="flex justify-center gap-3 items-center"
          style="min-height: calc(100vh - 300px)"
        >
          <loader kind="circle" />
          ...Loading
        </div>
        <s-list
          v-else
          :key="data.keyComponentSource"
          ref="listSourceCtl"
          class="w-full"
          :config="{ setting: {} }"
          :col="1"
          hide-delete-button
          hide-new-button
          hide-sort
          v-model="data.dataSource"
          hide-control
          hide-footer
        >
          <template #item="{ item }">
            <item-source
              :kind="kind"
              @onSelect="onSelectSource"
              @onReset="onResetSource"
              :item="item"
              :selected-source="data.selectedSource"
              :selected-invoice="data.selectedInvoice"
              :calcOustanding="calcOustandingSource"
            />
          </template>
        </s-list>
      </div>
      <div class="bg-white p-2">
        <div
          v-if="data.loadingInvoice"
          class="flex justify-center gap-3 items-center"
          style="min-height: calc(100vh - 300px)"
        >
          <loader kind="circle" />
          ...Loading
        </div>
        <s-list
          v-else
          :key="data.keyComponentApply"
          ref="listApplyCtl"
          class="w-full"
          :config="{ setting: {} }"
          :col="1"
          hide-delete-button
          hide-new-button
          hide-sort
          v-model="data.dataInvoice"
          hide-control
          hide-footer
        >
          <template #item="{ item }">
            <item-invoice
              :kind="kind"
              :item="item"
              :selected-source="data.selectedSource"
              @onHover="onHoverInvoice"
              @onLeave="onLeaveInvoice"
              @onReset="onResetInvoice"
              @onSettled="onSettledInvoice"
              @onUncheck="onUncheckInvoice"
              :calcAdjustment="calcAdjustment"
              :calcOustandingSource="calcOustandingSource"
              :calcOustanding="calcOustandingInvoice"
            >
              <template #adjustments="{ item }">
                <adjusment
                  :read-only="
                    item.SourceRecID !== '' && item.SourceRecID !== null
                  "
                  :selected-source-id="data.selectedSource._id"
                  :adjusments="item.Adjustments"
                  :source-rec-id="item.CurrentSourceRecID"
                  :key="data.selectedSource._id + item.CurrentSourceRecID"
                  @submit="(...args) => onSubmitAdjustment(item, ...args)"
                />
              </template>
            </item-invoice>
          </template>
        </s-list>
      </div>
    </div>
  </div>
  <modal-confirm
    :sources="data.dataSource"
    v-if="data.showConfirm"
    @close="data.showConfirm = false"
    @submit="onSubmit"
    :calcAdjustment="calcAdjustment"
  ></modal-confirm>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed } from "vue";
import {
  SList,
  SForm,
  SInput,
  loadFormConfig,
  util,
  SButton,
  SModal,
} from "suimjs";
import moment from "moment";
import AccountSelector from "@/components/common/AccountSelector.vue";
import Adjusment from "./Adjusment.vue";
import ItemSource from "./ItemSource.vue";
import ItemInvoice from "./ItemInvoice.vue";
import ModalConfirm from "./ModalConfirm.vue";
import { authStore } from "@/stores/auth";
import Loader from "@/components/common/Loader.vue";
const auth = authStore();

const listSourceCtl = ref(null);
const listApplyCtl = ref(null);
const axios = inject("axios");

const props = defineProps({
  journalId: { type: String, default: "" },
  kind: { type: String, default: "apply" }, //in|out|apply
});
const emit = defineEmits({
  back: null,
  postSubmit: null,
});

const data = reactive({
  OffsetAccount: {
    AccountType: "CUSTOMER",
  },
  dataSource: [],
  dataInvoice: [],
  keyComponentSource: 0,
  keyComponentApply: 0,
  showConfirm: false,
  loadingSource: false,
  loadingInvoice: false,
  selectedSource: {
    _id: "",
    MapApply: {},
    Amount: 0,
    Settled: 0,
    CurrentSettled: 0,
  },
  selectedInvoice: { _id: "" },
});
const cfg = computed({
  get() {
    let obj = {
      urlSource: "/fico/apply/get-cash-bank",
      urlInvoice: "/fico/apply/get-invoice",
      urlSubmit: "/fico/apply/apply",
      paramSource: {
        Account: data.OffsetAccount.AccountType,
        ID: data.OffsetAccount.AccountID,
      },
      paramInvoice: {
        Account: data.OffsetAccount.AccountType,
        ID: data.OffsetAccount.AccountID,
      },
      selectSource: () => {},
      getParamSubmit: () => {
        return getParamSubmitApply();
      },
      parseDataSource: (dt) => {
        return getParamDataSourceApply(dt);
      },
    };

    if (["in", "out"].includes(props.kind))
      obj = {
        ...obj,
        ...{
          urlSource: "/fico/cashjournal/get",
          urlInvoice: "/fico/apply/get-invoice-cash-journal",
          urlSubmit: "/fico/apply/apply-cash-journal",
          paramSource: [props.journalId],
          paramInvoice: {
            CashScheduleID: data.selectedSource?.Account?.AccountID,
            Type: props.kind,
          },
          selectSource: function () {
            fetchDataInvoice(this.paramInvoice);
          },
          getParamSubmit: () => {
            return getParamSubmitInOut();
          },
          parseDataSource: (dt) => {
            return getParamDataSourceInOut(dt);
          },
        },
      };

    return obj;
  },
});

function onSelectSource(record) {
  const sourceId = record._id;
  const idx = getIndexSource(record._id);
  if (idx === -1) return;

  const source = data.dataSource[idx];
  data.selectedSource =
    data.selectedSource._id === source._id ? { _id: "", MapApply: {} } : source;

  cfg.value.selectSource();
}
function onResetSource(record) {
  data.dataInvoice.forEach((el) => {
    if (el.CurrentSourceRecID !== record._id || el.CurrentSourceRecID == "")
      return;
    onResetInvoice(el);
  });
}

function onHoverInvoice(record) {
  data.selectedInvoice = { ...record };
}
function onLeaveInvoice(record) {
  data.selectedInvoice = {};
}
function getIndexInvoice(invoiceId) {
  return data.dataInvoice.findIndex((e) => e._id === invoiceId);
}
function getIndexSource(sourceId) {
  return data.dataSource.findIndex((e) => e._id === sourceId);
}
function onResetInvoice(record) {
  const invoiceId = record._id;
  const idx = getIndexInvoice(record._id);
  if (idx === -1) return;

  const oldValueSettled = record.Settled;

  const CurrentSourceRecID = record.CurrentSourceRecID;
  const Adjustments = [];

  data.dataInvoice[idx] = {
    ...data.dataInvoice[idx],
    ...{
      Settled: 0,
      CurrentSourceRecID: "",
      Adjustments,
    },
  };

  setSourceSettled(CurrentSourceRecID, oldValueSettled * -1);
  setMapApply(CurrentSourceRecID, invoiceId, oldValueSettled * -1, false, []);
}
function onSettledInvoice(record, oustanding) {
  const invoiceId = record._id;
  const idx = getIndexInvoice(record._id);
  if (idx === -1) return;

  const Settled = oustanding;
  const CurrentSourceRecID = data.selectedSource._id;

  data.dataInvoice[idx] = {
    ...data.dataInvoice[idx],
    ...{
      Settled,
      CurrentSourceRecID,
    },
  };

  setSourceSettled(CurrentSourceRecID, Settled);
  setMapApply(
    CurrentSourceRecID,
    invoiceId,
    Settled,
    true,
    data.dataInvoice[idx].Adjustments
  );
}

function onSubmitAdjustment(record, newAdjustments, cbReset) {
  const invoiceId = record._id;
  const idx = getIndexInvoice(invoiceId);
  if (idx === -1) return;

  const newTotalAdjustment = calcAdjustment(newAdjustments);

  const oustandingInvoice = calcOustandingInvoice(record, newTotalAdjustment);
  const oustandingSource = calcOustandingSource(data.selectedSource);

  if (props.kind == "apply" && oustandingInvoice > oustandingSource)
    return cbReset();

  const currentTotalAdjustment = calcAdjustment(record.Adjustments);

  const diffTotalAdjustment = newTotalAdjustment - currentTotalAdjustment;

  let sourceCurrentSettled = 0;
  let CurrentSourceRecID = record.CurrentSourceRecID;
  let Settled = 0;
  let Adjustments = JSON.parse(JSON.stringify(newAdjustments));

  if (CurrentSourceRecID !== "" && CurrentSourceRecID !== null) {
    Settled = record.Settled - diffTotalAdjustment;
    sourceCurrentSettled = diffTotalAdjustment * -1;
  } else {
    CurrentSourceRecID = data.selectedSource._id;
    Settled = record.Oustanding - newTotalAdjustment;
    sourceCurrentSettled = Settled;
  }

  data.dataInvoice[idx] = {
    ...data.dataInvoice[idx],
    ...{
      Settled,
      CurrentSourceRecID,
      Adjustments,
    },
  };
  setSourceSettled(CurrentSourceRecID, sourceCurrentSettled);
  setMapApply(CurrentSourceRecID, invoiceId, Settled, true, Adjustments);
}
function setSourceSettled(sourceId, settled) {
  const idx = getIndexSource(sourceId);
  if (idx === -1) return;
  let CurrentSettled = data.dataSource[idx].CurrentSettled + settled;
  data.dataSource[idx] = {
    ...data.dataSource[idx],
    ...{
      CurrentSettled,
    },
  };
}
async function onUncheckInvoice(record) {
  const param = [
    {
      CompanyID: record.CompanyID,
      TrxDate: new Date(),
      SourceRecID: record.SourceRecID,
      Applies: [
        {
          ApplyToRecID: record._id,
          IsUnchecked: true,
          Adjustments: record.Adjustments,
          Amount: record.Settled,
        },
      ],
    },
  ];

  await submitApply("/fico/apply/apply", param);
  fetchDataInvoice(cfg.value.paramInvoice);
  setSourceSettled(record.SourceRecID, record.Settled * -1);
}

function calcOustandingSource(
  source = { Amount: 0, Settled: 0, CurrentSettled: 0 }
) {
  const amount = source.Amount ?? 0;
  const settled = source.Settled ?? 0;
  const currentSettled = source.CurrentSettled ?? 0;

  return parseInt(amount - (settled + currentSettled));
}

function calcOustandingInvoice(
  invoice = { Amount: 0, Settled: 0 },
  totalAdjusment = 0
) {
  const amount = invoice.Amount ?? 0;
  const settled = invoice.Settled ?? 0;

  return parseInt(amount - (settled + totalAdjusment));
}

function calcAdjustment(adjustments = []) {
  if (!Array.isArray(adjustments)) return 0;
  return adjustments.reduce((total, e) => {
    total += e.Amount;
    return total;
  }, 0);
}

function setMapApply(sourceID, invoiceID, Settled, IsSettled, Adjustments) {
  data.dataSource.forEach((e) => {
    if (e._id !== sourceID) return;
    e.MapApply[invoiceID] = {
      ...e.MapApply[invoiceID],
      ...{ Settled, IsSettled, Adjustments },
    };
  });
}

function getParamDataSourceApply(dt) {
  return dt.map((el) => {
    el.CurrentSettled = 0;
    el.CurrentOustanding = 0;
    el.MapApply = {};
    return el;
  });
}
function getParamDataSourceInOut(dt) {
  return dt.Lines.map((el, idx) => {
    el.Created = dt.TrxDate;
    el._id = String(el.LineNo);
    el.SourceType = dt.CashJournalType;
    el.Oustanding = el.TotalAmount;
    el.CurrentSettled = 0;
    el.CurrentOustanding = 0;
    el.MapApply = {};
    el.CompanyID = dt.CompanyID;
    return el;
  });
}

function getParamSubmitApply() {
  const param = [];
  data.dataSource.forEach((el) => {
    const Applies = [];
    Object.keys(el.MapApply).forEach((mp) => {
      let obj = el.MapApply[mp];
      if (obj.IsSettled == true) {
        let Amount = obj.Settled;
        let IsUnchecked = false;

        Applies.push({
          ApplyToRecID: mp,
          IsUnchecked,
          Amount,
          Adjustment: obj.Adjustments.map((adj) => {
            return {
              Dimension: [],
              From: el.Account,
              To: adj.Account,
              Text: adj.Text,
              Amount: adj.Amount,
            };
          }),
        });
      }
    });

    if (Applies.length > 0) {
      param.push({
        CompanyID: el.CompanyID,
        TrxDate: new Date(),
        SourceRecID: el._id,
        Applies,
      });
    }
  });
  return param;
}
function getParamSubmitInOut() {
  const param = {
    CashJournalID: props.journalId,
    Applies: [],
  };
  data.dataSource.forEach((el) => {
    Object.keys(el.MapApply).forEach((mp) => {
      let obj = el.MapApply[mp];
      if (obj.IsSettled === true) {
        const idx = data.dataInvoice.findIndex((el) => el._id == mp);
        const applyTo = data.dataInvoice[idx];
        param.Applies.push({
          CompanyID: el.CompanyID,
          Source: {
            Module: "",
            JournalID: "",
            LineNo: el.LineNo,
            JournalType: "",
          },
          ApplyTo: {
            Module: applyTo.SourceType,
            JournalID: applyTo.SourceJournalID,
            LineNo: applyTo.SourceLineNo,
            JournalType: "",
            RecordID: applyTo._id,
          },
          Adjustment: [],
          Amount: obj.Settled,
        });
      }
    });
  });

  return [param];
}

async function refreshData() {
  data.selectedSource = { _id: "", MapApply: {} };
  data.selectedInvoice = { _id: "" };

  const param = cfg.value.paramSource;
  await fetchDataSource(param);
  if (props.kind === "apply") {
    fetchDataInvoice(param);
  }
}
async function fetchDataSource(param) {
  data.loadingSource = true;
  data.dataSource = [];
  const url = cfg.value.urlSource;
  try {
    const r = await axios.post(url, param);
    data.dataSource = cfg.value.parseDataSource(r.data);
  } catch (e) {
    util.showError(e);
  } finally {
    data.loadingSource = false;
    data.keyComponentSource++;
  }
}
async function fetchDataInvoice(param) {
  data.loadingInvoice = true;
  data.dataInvoice = [];
  const url = cfg.value.urlInvoice;
  axios
    .post(url, param)
    .then(
      (r) => {
        const dt = [...r.data]
          .map((el) => {
            setApplyItem(el);
            return el;
          })
          .sort((a, b) =>
            a.CurrentSourceRecID > b.CurrentSourceRecID ? -1 : 1
          );

        data.dataInvoice = dt;
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      data.loadingInvoice = false;
      data.keyComponentApply++;
    });
}
function setApplyItem(item) {
  let r = JSON.parse(JSON.stringify(item));
  let CurrentSourceRecID = "";

  data.dataSource.map((el) => {
    if (
      el.MapApply[item._id] !== undefined &&
      el.MapApply[item._id].IsSettled === true
    ) {
      r = el.MapApply[item._id];
      CurrentSourceRecID = el._id;
    } else {
      let Name = `${r.SourceJournalID} - ${r.VoucherNo}`;
      let IsSettled = false;
      let Settled = 0;
      let Adjustments = [];

      // if(el._id === item.SourceRecID){
      //   IsSettled = r.Settled > 0
      //   Settled = r.Settled
      //   EditedSettled = r.Settled
      //   Adjustments = Array.isArray(r.Adjustments) ? r.Adjustments :[];
      //   Change =  false;
      // }
      el.MapApply[item._id] = { Name, IsSettled, Settled, Adjustments };
    }
    return el;
  });
  item.SourceRecID = r.SourceRecID ?? "";
  item.Settled = r.Settled;
  item.CurrentSourceRecID = CurrentSourceRecID;
  item.Adjustments = Array.isArray(r.Adjustments) ? r.Adjustments : [];
}

async function onSubmit(cbSucces, cbError) {
  const param = cfg.value.getParamSubmit();
  const url = cfg.value.urlSubmit;
  // console.log(param, url);
  await submitApply(url, param, cbSucces, cbError);
  refreshData();
  emit("postSubmit");
}
async function submitApply(
  url,
  param,
  cbSucces = function () {},
  cbError = function () {}
) {
  try {
    await axios.post(url, param);
    cbSucces();
  } catch (e) {
    util.showError(e);
    cbError();
  }
}

onMounted(() => {
  refreshData();
});
</script>
<style scoped>
.acitve-item{
  @apply bg-slate-300; 
  opacity: 0.75 !important;
}
.active-apply{
  @apply bg-gray-100 border-slate-300 border-[1px]; 
  /* opacity: 0.75 !important; */
}
.disable-apply{
  @apply pointer-events-none cursor-not-allowed opacity-50
}
.card-apply.disable-apply:hover{
  background: white !important
}
 
.edited-item{
  @apply bg-slate-100  ring-slate-300  ring-[1px];  
}
</style>
<style>
.apply-list .label-number{
  text-align: right;
}
.apply-list  .suim_list ul{
  grid-template-columns: repeat(1, minmax(0, 1fr)) !important; 
  gap:0 !important;
  max-height:calc(100vh - 265px);
  overflow: auto;
}
.apply-list  .suim_list ul li:hover  > div >div {
  @apply bg-slate-100;
  color: #222 !important
  
}
.apply-list  .suim_list ul li:hover  label.input_label{
  color: #222 !important
}
.apply-list  .suim_list ul li .card-apply .input-number label.input_label>div{
  text-align: right;
}

.apply-list .suim_list ul li > div{
  padding: 0 !important;
}
/* .apply-list > .suim_list ul li */
</style>