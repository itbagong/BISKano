<template>
  <div>
    <data-list
      class="card datalist-trayekritase"
      :class="data.selectedRevenueType.toLowerCase().replace(' ', '_')"
      ref="listControl"
      title="Ritase"
      grid-config="/bagong/siteentry_trayekritase/gridconfig"
      form-config="/bagong/siteentry_trayekritase/formconfig"
      :grid-read="
        '/bagong/siteentry_trayekritase/gets?SiteEntryAssetID=' +
        siteEntryAssetID
      "
      form-read="/bagong/siteentry_trayekritase/get"
      grid-mode="grid"
      grid-delete="/bagong/siteentry_trayekritase/delete"
      form-keep-label
      :form-insert="data.uriInsert"
      :form-update="data.uriUpdate"
      :form-fields="[
        'ConfigDeposit',
        'IsFlat',
        'OtherIncome',
        'PassengerIncome',
        'RitasePassenger',
        'RitaseIncome',
        'TerminalExpense',
        'FixExpense',
        'OtherExpense',
        'RevenueType',
        'IsFullDeposit',
        'TargetNonFlat',
        'RitaseDeposit',
        'ShiftID',
        'TicketNumbers',
        'StartKM',
        'EndKM',
      ]"
      :grid-fields="[
        'ConfigDeposit',
        'ConfigPremi',
        'RitaseSummary',
        'RitasePremi',
        'Bonus',
      ]"
      form-focus
      stayOnFormAfterSave
      @preSave="preSave"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @form-field-change="onFormFieldChanged"
      :formHideSubmit="!data.isGoDetail || readOnly"
      @alterGridConfig="alterGridConfig"
      :grid-hide-edit="hideEdit"
    >
      <!-- -->
      <template #form_buttons_1="{ item }">
        <s-button
          v-if="data.haveDraft"
          :disabled="data.disabledFormButton"
          @click="onSubmit(item._id)"
          class="btn_primary"
          label="Submit"
        ></s-button>
      </template>

      <template #form_input_RevenueType="{ item, mode }">
        <div class="flex justify-between">
          <div class="flex gap-4">
            <s-input
              class="min-w-[300px] max-w-[300px]"
              label="Type"
              v-model="item.RevenueType"
              use-list
              :disabled="data.isGoDetail"
              :items="['Premi', 'Setoran']"
              @change="
                (name, v1) => {
                  ritasePremi.selected = '';
                  item.RitasePremi = [];
                }
              "
            />
            <s-input
              class="min-w-[300px] max-w-[300px]"
              v-if="item.RevenueType == 'Premi'"
              label="Ritase"
              v-model="ritasePremi.selected"
              use-list
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=MRP"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :disabled="data.isGoDetail"
              placeholder="Ritase"
              keep-label
              @change="(...args) => onChangeRitasePremi(item, ...args)"
            />
          </div>
          <div class="flex justify-end" v-if="mode !== 'view'">
            <s-button
              icon="close"
              v-if="data.isGoDetail"
              label="Reset"
              @click="resetDetail(item)"
              class="btn_error h-[35px]"
            />
            <s-button
              icon="arrow-right"
              :disabled="!validBtnGoDetail(item)"
              v-else
              label="Go"
              @click="geDetail(item)"
              class="btn_warning h-[35px]"
            />
          </div>
        </div>

        <header-calc
          class="mb-2"
          v-if="data.isGoDetail"
          :revenue-type="item.RevenueType"
          :revenue="revenue"
          :income="income"
          :expense="expense"
          :method="item.ConfigPremi.Method"
          :target1="item.ConfigPremi.Target1"
          :target2="item.ConfigPremi.Target2"
          :percent1="item.ConfigPremi.Percent1"
          :percent2="item.ConfigPremi.Percent2"
          :bonus1="bonus1"
          :bonus2="bonus2"
          :total-bonus="totalBonus"
          :total-method="totalMethod"
        />
        <ritase-journal
          v-if="data.haveSubmitted && item.RevenueType == 'Setoran'"
          @reOpen="reOpen"
          v-model="item.DepositIncome"
          journal-type="CUSTOMER"
        />
      </template>

      <template #grid_ConfigDeposit="{ item }">
        <div v-if="item.RevenueType == 'Setoran'" class="flex flex-col gap-1">
          <div>
            <b>Target flat :</b>
            {{ item.ConfigDeposit.TargetFlat }}
          </div>
          <div>
            <b>Target non flat :</b>
            {{ item.ConfigDeposit.TargetNonFlat }}
          </div>
          <div>
            <b>Toll :</b>
            {{ item.ConfigDeposit.IsToll }}
          </div>
        </div>
        <template v-else> &nbsp; </template>
      </template>

      <template #grid_ConfigPremi="{ item }">
        <div v-if="item.RevenueType == 'Premi'" class="flex flex-col gap-1">
          <div>
            <b>Persen-1 :</b>
            {{ item.ConfigPremi.Percent1 * 100 }}%
          </div>
          <div>
            <b>Target-1 :</b>
            {{ util.formatMoney(item.ConfigPremi.Target1) }}
          </div>
          <template v-if="item.ConfigPremi.Method == 2">
            <div>
              <b>Persen-2 :</b>
              {{ item.ConfigPremi.Percent2 * 100 }}%
            </div>
            <div>
              <b>Target-2 :</b>
              {{ util.formatMoney(item.ConfigPremi.Target2) }}
            </div>
          </template>
        </div>
        <template v-else>&nbsp;</template>
      </template>

      <template #grid_RitaseSummary="{ item }">
        <div v-if="item.RevenueType == 'Premi'" class="flex flex-col gap-1">
          <div>
            <b>Total ritase :</b>
            {{ util.formatMoney(item.RitaseSummary.TotalRitaseIncome) }}
          </div>
          <div>
            <b>Total other income :</b>
            {{ util.formatMoney(item.RitaseSummary.TotalOtherIncome) }}
          </div>
          <!-- <div>
            <b>Total terminal expense :</b>
            {{ util.formatMoney(item.RitaseSummary.TotalTerminalExpense) }}
          </div> -->
          <div>
            <b>Total fix expense :</b>
            {{ util.formatMoney(item.RitaseSummary.TotalFixExpense) }}
          </div>
          <div>
            <b>Total other expense :</b>
            {{ util.formatMoney(item.RitaseSummary.TotalOtherExpense) }}
          </div>
          <div>
            <b>Total bonus :</b>
            {{ util.formatMoney(item.RitaseSummary.TotalBonus) }}
          </div>
          <div>
            <b>Total profit :</b>
            {{ util.formatMoney(item.RitaseSummary.TotalMethod) }}
          </div>
        </div>
        <template v-else>&nbsp;</template>
      </template>
      <template #grid_Bonus="{ item }">
        <div class="text-right">
          {{ util.formatMoney(item.RitaseSummary.TotalBonus) }}
        </div>
      </template>

      <template #form_input_ConfigDeposit="{ item }">
        <s-input
          label="Target Deposit Flat"
          v-if="item.CategoryDeposit == 'Flat'"
          v-model="item.ConfigDeposit.TargetFlat"
          read-only
          kind="number"
        />
        <s-input
          label="Target Deposit Non Flat"
          v-else-if="item.CategoryDeposit == 'Non Flat'"
          v-model="item.ConfigDeposit.TargetNonFlat"
          read-only
          kind="number"
        />
        <s-input
          label="Target Rent"
          v-else
          v-model="item.ConfigDeposit.TargetRent"
          :read-only="item.CategoryDeposit != 'Rent'"
          kind="number"
        />
      </template>

      <template #form_input_RitasePassenger="{ item, mode }">
        <template v-if="data.isGoDetail && item.RevenueType == 'Premi'">
          <s-input
            v-if="item.TrayekName === SPECIAL_TRAYEKNAME"
            v-model="item.Kurs"
            kind="number"
            label="Kurs"
            class="w-[200px] mb-5"
            :read-only="mode == 'view'"
          />

          <tarif
            :special-trayek-name="SPECIAL_TRAYEKNAME"
            :special-terminal-from="SPECIAL_TERMINALFROM"
            :trayek-name="item.TrayekName"
            :kurs="item.Kurs"
            :read-only="
              item?.PassengerIncome?.JournalID != '' &&
              item?.PassengerIncome?.JournalID != undefined
            "
            v-model="data.matrixTarif"
            @calc-total-penumpang-a="calcTotalPenumpangA"
            @calc-total-penumpang-b="calcTotalPenumpangB"
            @calc-total-amount-a="calcTotalAmountRitaseA"
            @calc-total-amount-b="calcTotalAmountRitaseB"
          />
          <div class="grid grid-cols-2 gap-4 my-4">
            <div class="font-semibold">
              <div>
                Total Penumpang Naik: {{ util.formatMoney(total.penumpangA) }}
              </div>
              <div>Total: {{ util.formatMoney(total.ritaseA) }}</div>
            </div>
            <div class="font-semibold">
              <div>
                Total Penumpang Naik: {{ util.formatMoney(total.penumpangB) }}
              </div>
              <div>Total: {{ util.formatMoney(total.ritaseB) }}</div>
            </div>
          </div>
        </template>
        <div class="flex">
          <ritase-journal
            v-if="data.haveSubmitted"
            @reOpen="reOpen"
            v-model="item.PassengerIncome"
            journal-type="CUSTOMER"
          />
        </div>
      </template>

      <template #form_input_OtherIncome="{ item, mode }">
        <template v-if="data.isGoDetail">
          <income
            v-model="item.OtherIncome"
            :grid-config-url="
              readOnly
                ? '/bagong/siteincome-read/grid/gridconfig'
                : '/bagong/siteincome/grid/gridconfig'
            "
            @calc="calcTotalAmountOtherIncome"
            :attch-kind="data.attchKind"
            :attch-ref-id="data.record._id"
            :attch-tag-prefix="data.attchKind + '_OTHER'"
            :tag-upload="`SE_TRAYEK_${siteEntryAssetID}`"
            @preOpenAttch="preOpenAttch"
            @reOpen="reOpen"
          />
          <div class="my-3 font-bold">
            Total: {{ util.formatMoney(total.otherIncome) }}
          </div>
        </template>
      </template>

      <template #form_input_PassengerIncome="{ item, config, mode }">
        <template v-if="item.PassengerIncome">
          <s-input
            kind="number"
            keep-label
            :label="config.label"
            v-model="item.PassengerIncome.Amount"
          ></s-input>
        </template>
      </template>

      <template #form_input_OtherExpense="{ item, mode }">
        <template v-if="data.isGoDetail && item.RevenueType == 'Premi'">
          <ritase-expense
            v-model="item.OtherExpense"
            @calc="calcTotalAmountOtherExpense"
            :attch-kind="data.attchKind"
            :attch-ref-id="item._id"
            :attch-tag-prefix="data.attchKind + '_OTHER'"
            :tag-upload="`SE_TRAYEK_${siteEntryAssetID}`"
            @preOpenAttch="preOpenAttch"
            @reOpen="reOpen"
          />
          <div class="my-3 font-bold">
            Total: {{ util.formatMoney(total.otherExpense) }}
          </div>
        </template>
      </template>

      <template #form_input_FixExpense="{ item, mode }">
        <template v-if="data.isGoDetail && item.RevenueType == 'Premi'">
          <ritase-expense
            v-model="item.FixExpense"
            @calc="calcTotalAmountFixExpense"
            :attch-kind="data.attchKind"
            :attch-ref-id="item._id"
            :attch-tag-prefix="data.attchKind + '_FIX'"
            :tag-upload="`SE_TRAYEK_${siteEntryAssetID}`"
            @preOpenAttch="preOpenAttch"
            @reOpen="reOpen"
          />

          <div class="my-3 font-bold">
            Total: {{ util.formatMoney(total.fixExpense) }}
          </div>
        </template>
        <div class="flex">
          <ritase-journal
            v-if="data.haveSubmitted"
            @reOpen="reOpen"
            v-model="data.journalExpense"
            journal-type="VENDOR"
          />
        </div>
      </template>

      <template #form_input_RitaseDeposit="{ item }">
        <template
          v-if="
            item.CategoryDeposit == 'Flat' || item.CategoryDeposit == 'Rent'
          "
          >&nbsp;</template
        >
      </template>

      <template #form_input_IsFullDeposit="{ item }">
        <template
          v-if="
            item.CategoryDeposit == 'Flat' || item.CategoryDeposit == 'Rent'
          "
          >&nbsp;</template
        >
      </template>

      <template #form_input_ShiftID="{ item, mode }">
        <template v-if="data.isGoDetail">
          <div class="input_label">Shift</div>
          <s-input
            :read-only="readOnly || mode == 'view'"
            class="min-w-[100px]"
            hide-label
            use-list
            v-model="item.ShiftID"
            :items="data.shiftItems"
          />
        </template>
        <div v-else>&nbsp;</div>
      </template>

      <template #form_input_TicketNumbers="{ item, mode }">
        <s-list-editor
          no-gap
          :config="{}"
          v-model="item.TicketNumbers"
          :allow-add="!(mode == 'view' || readOnly)"
          hide-header
          hide-select
          @validate-item="addTicketNumber"
          ref="ticketNumber"
        >
          <template #item="{ item }">
            <div class="grow grid grid-cols-2 gap-4">
              <s-input kind="text" v-model="item.Start" label="Start" />
              <s-input kind="text" v-model="item.End" label="End" />
            </div>
          </template>
        </s-list-editor>
      </template>
      <template #form_input_StartKM="{ item, config }">
        <div class="flex gap-4">
          <s-input
            keep-label
            class="w-full"
            :label="config.label"
            v-model="item.StartKM"
            kind="number"
          ></s-input>
          <uploader
            ref="gridAttachmentKMStart"
            :journalId="item._id"
            :journalType="data.attchKind"
            :config="config"
            :tags="[
              `${data.attchKind}_KM_START_${item._id}`,
              `SE_TRAYEK_${siteEntryAssetID}`,
            ]"
            :tags-for-get="[`${data.attchKind}_KM_START_${item._id}`]"
            :key="1"
            bytag
            hide-label
            single-save
            @close="emit('refreshAttach')"
            @preOpen="preOpenAttch"
          />
        </div>
      </template>
      <template #form_input_EndKM="{ item, config }">
        <div class="flex gap-4">
          <s-input
            keep-label
            class="w-full"
            :label="config.label"
            v-model="item.EndKM"
            kind="number"
          ></s-input>
          <uploader
            ref="gridAttachmentKMEnd"
            :journalId="item._id"
            :journalType="data.attchKind"
            :config="config"
            :tags="[
              `${data.attchKind}_KM_END_${item._id}`,
              `SE_TRAYEK_${siteEntryAssetID}`,
            ]"
            :tags-for-get="[`${data.attchKind}_KM_END_${item._id}`]"
            :key="1"
            bytag
            hide-label
            single-save
            @close="emit('refreshAttach')"
            @preOpen="preOpenAttch"
          />
        </div>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, onMounted, inject, ref, computed, watch } from "vue";
import {
  SForm,
  createFormConfig,
  util,
  SInput,
  SButton,
  DataList,
  SListEditor,
} from "suimjs";
import HeaderCalc from "./RitaseHeaderCalc.vue";

import helper from "@/scripts/helper.js";
import Income from "./Income.vue";
import Tarif from "./RitaseTarif.vue";
// import TerminalExpense from "./RitaseTerminalExpense.vue";
import RitaseExpense from "./RitaseExpense.vue";

import RitaseJournal from "./RitaseJournal.vue";

import Uploader from "@/components/common/Uploader.vue";

const fixExpense = ref(null);
const listControl = ref(null);
const ticketNumber = ref(null);

const gridAttachmentKMStart = ref(null);
const gridAttachmentKMEnd = ref(null);

const axios = inject("axios");

const SPECIAL_TRAYEKNAME = "KUPANG-DILI";
const SPECIAL_TERMINALFROM = "DIL";

const props = defineProps({
  siteEntryAssetID: { type: String, default: "" },
  site: { type: String, default: "" },
  assetID: { type: String, default: "" },
  trayekName: { type: String, default: "" },
  hideEdit: { type: Boolean, default: false },
});

const data = reactive({
  status: "",
  appMode: "grid",
  formMode: !props.hideEdit ? "edit" : "view",
  isGoDetail: false,
  trayek: undefined,
  ritase: undefined,
  matrixTarif: [],
  matrixTerminalExpense: [],
  listRitasePremi: generateListRitase(),
  selectedRevenueType: "",
  selectedRitasePremi: "",
  record: {},
  structConfigDeposit: {
    TargetFlat: 0,
    TargetNonFlat: 0,
    TargetRent: 0,
    IsToll: false,
  },
  structConfigPremi: {
    Method: 0,
    Target1: 0,
    Target2: 0,
    Percent1: 0,
    Percent2: 0,
  },
  formCfgTicketNum: {},
  shiftItems: [],
  disabledFormButton: false,
  attchKind: "SE_TRAYEK_RITASE",
  uriInsert: "/bagong/siteentry_trayekritase/insert",
  uriUpdate: "/bagong/siteentry_trayekritase/update",
  selectedAttchTag: "",
  haveDraft: true,
  haveSubmitted: false,
  journalExpense: {},
});
const emit = defineEmits({
  refreshAttach: null,
});

function preOpenAttch(readOnly) {
  if (readOnly) return;
  preSave(data.record);
  const mode = listControl.value.getFormMode();
  axios
    .post(
      mode == "new"
        ? "/bagong/siteentry_trayekritase/insert"
        : "/bagong/siteentry_trayekritase/update",
      data.record
    )
    .then((r) => {
      listControl.value.setFormMode(data.formMode);
      data.record._id = r.data._id;
    });
}
function reOpen() {
  save();
}
function save() {
  listControl.value.setFormLoading(true);
  const cb = () => {
    listControl.value.setFormLoading(false);
  };
  listControl.value.submitForm(data.record, cb, cb);
}

const ritasePremi = reactive({
  list: generateListRitase(),
  selected: "",
});

// calculate
const total = reactive({
  passengers: 0,
  allAmountRitase: 0,
  allRitase: 0,
  penumpangA: 0,
  penumpangB: 0,
  ritaseA: 0,
  ritaseB: 0,
  otherIncome: 0,
  terminalExpense: 0,
  fixExpense: 0,
  otherExpense: 0,
});

const income = computed({
  get() {
    if (data.record.RevenueType == "Premi")
      return total.otherIncome + total.ritaseA + total.ritaseB;
    else
      return parseInt(total.otherIncome) + parseInt(data.record.AmountDeposit);
  },
});

const expense = computed({
  get() {
    if (data.record.RevenueType == "Premi")
      return total.terminalExpense + total.otherExpense + total.fixExpense;
    else return 0;
  },
});

const revenue = computed({
  get() {
    const r = parseInt(income.value) - parseInt(expense.value);
    return data.record.RevenueType == "Premi" ? r - totalBonus.value : r;
  },
});

const revenueKotor = computed({
  get() {
    return income.value - total.otherExpense;
  },
});

const bonus1 = computed({
  get() {
    if (data.record.ConfigPremi.Method == 1) return getCalcBonus1Method1();
    else if (data.record.ConfigPremi.Method == 2) return getCalcBonus1Method2();
    else if (data.record.ConfigPremi.Method == 3) return getCalcBonus1Method3();
    return 0;
  },
});

const bonus2 = computed({
  get() {
    if (data.record.ConfigPremi.Method == 2) return getCalcBonus2Method2();
    return 0;
  },
});

const totalBonus = computed({
  get() {
    return bonus1.value + bonus2.value;
  },
});

const totalMethod = computed({
  get() {
    if (data.record.ConfigPremi.Method == 1) return getCalcTotalMethod1();
    else if (data.record.ConfigPremi.Method == 2) return getCalcTotalMethod2();
    else if (data.record.ConfigPremi.Method == 3) return getCalcTotalMethod3();
    else return 0;
  },
});
function getCalcTotalMethod1() {
  return revenue.value - bonus1.value;
}

function getCalcTotalMethod2() {
  return (
    total.otherIncome +
    total.ritaseA +
    total.ritaseB -
    (total.terminalExpense + total.otherExpense + totalBonus.value) -
    total.fixExpense
  );
}

function getCalcTotalMethod3() {
  return (
    total.otherIncome +
    total.ritaseA +
    total.ritaseB -
    (total.terminalExpense + total.otherExpense + bonus1.value) -
    total.fixExpense
  );
}

function getCalcBonus1Method1() {
  if (revenue.value > data.record.ConfigPremi.Target1) {
    return roundThousandths(
      (revenue.value - data.record.ConfigPremi.Target1) *
        data.record.ConfigPremi.Percent1
    );
  }
  return 0;
}

function getCalcBonus1Method2() {
  if (revenueKotor.value > data.record.ConfigPremi.Target2) {
    return roundThousandths(
      (data.record.ConfigPremi.Target2 - data.record.ConfigPremi.Target1) *
        data.record.ConfigPremi.Percent1
    );
  }
  if (revenueKotor.value > data.record.ConfigPremi.Target1) {
    return roundThousandths(
      (revenueKotor.value - data.record.ConfigPremi.Target1) *
        data.record.ConfigPremi.Percent1
    );
  }
  return 0;
}

function getCalcBonus2Method2() {
  if (revenueKotor.value > data.record.ConfigPremi.Target2) {
    return roundThousandths(
      (revenueKotor.value - data.record.ConfigPremi.Target2) *
        data.record.ConfigPremi.Percent2
    );
  }
  return 0;
}

function getCalcBonus1Method3() {
  if (revenueKotor.value > data.record.ConfigPremi.Target1) {
    return roundThousandths(
      (revenueKotor.value - data.record.ConfigPremi.Target1) *
        data.record.ConfigPremi.Percent1
    );
  }
  return 0;
}
function roundThousandths(val) {
  const n = 1000;
  const r = Math.floor(val / n) * n;
  return r;
}

function calcTotalPenumpangA(val) {
  total.penumpangA = val;
  calcTotalPenumpangAllRitase();
}

function calcTotalPenumpangB(val) {
  total.penumpangB = val;
  calcTotalPenumpangAllRitase();
}

function calcTotalAmountRitaseA(val) {
  total.ritaseA = val;
  calcTotalAmountAllRitase();
}

function calcTotalAmountRitaseB(val) {
  total.ritaseB = val;
  calcTotalAmountAllRitase();
}

function calcTotalPenumpangAllRitase() {
  total.allRitase = total.penumpangA + total.penumpangB;
}

function calcTotalAmountAllRitase() {
  total.allAmountRitase = total.ritaseA + total.ritaseB;
}

function calcTotalAmountOtherIncome(val) {
  total.otherIncome = val;
}

function calcTotalAmountTerminalExpense(val) {
  total.terminalExpense = val;
}

function calcTotalAmountFixExpense(val) {
  total.fixExpense = val;
}

function calcTotalAmountOtherExpense(val) {
  total.otherExpense = val;
}
// calc
function addTicketNumber(record) {
  ticketNumber.value.setValidateItem(true);
}
function validBtnGoDetail(record) {
  if (record.RevenueType == "") return false;
  else if (record.RevenueType == "Setoran") return true;
  else if (record.RevenueType == "Premi") return record.RitasePremi.length > 0;
}

function generateListRitase() {
  let list = [];
  for (let i = 1; i <= 3; i++) {
    let j = i * 2 - 1;
    let k = j + 1;
    list.push(j + " & " + k);
  }
  list.push("3 & 1");
  list.push("4 & 1");
  return list;
}

function onChangeRitasePremi(record, name, value, v2) {
  const res = v2 !== "" ? v2.split("&") : [];
  record.RitasePremi = res.map((e, i) => {
    return parseInt(e);
  });
}
function setFormLoading(loading) {
  data.disabledFormButton = loading;
  listControl.value.setFormLoading(loading);
}
function onSubmit() {
  const valid = listControl.value.formValidate();
  if (valid) {
    setFormLoading(true);
    listControl.value.submitForm(data.record, doSubmit);
  }
}

function doSubmit() {
  const url = "/bagong/postingprofile/post";
  const param = {
    JournalType: "SITEENTRY_TRAYEK",
    JournalID: data.record._id,
    Op: "Submit",
    Text: "",
  };
  axios.post(url, param).then(
    (r) => {
      setFormLoading(false);
      listControl.value.setControlMode("grid");
      listControl.value.refreshList();
    },
    (e) => {
      util.showError(e);
      setFormLoading(false);
    }
  );
}

function getTerminalPassenger(tarifs, ritase) {
  return Object.values(
    tarifs.reduce((group, obj) => {
      if (obj.From == obj.To) return group;

      group[obj.From] = group[obj.From] ?? {
        TerminalID: obj.From,
        TerminalName: obj.FromName,
        Index: Object.keys(group).length,
        TotalPassenger: 0,
        Amount: 0,
        Ritase: ritase,
        Passengers: [],
      };
      if (group[obj.To] != undefined) return group;
      group[obj.From].Passengers.push({
        Ritase: ritase,
        Tariff: obj.Rate,
        From: obj.From,
        FromName: obj.FromName,
        To: obj.To,
        ToName: obj.ToName,
        Total: 0,
        Income: 0,
      });
      return group;
    }, {})
  );
}

function generateNewMatrixTarif(ritases) {
  data.matrixTarif = [
    getTerminalPassenger(data.ritase.Tarif, 1),
    getTerminalPassenger(data.ritase.Tarif.reverse(), 2),
  ];
}

function generateEditMatrixTarif(incomes, Passengers) {
  let idx = -1;
  let currentRitase = 1;
  const arr = incomes.reduce((ar, e, i) => {
    const obj = { ...e, ...{ Index: i, Passengers: [] } };
    obj.Passengers = Passengers.filter(
      (e2) => e2.From == e.TerminalID && e2.Ritase == e.Ritase
    );

    if (currentRitase != e.Ritase) {
      currentRitase = e.Ritase;
      idx = 0;
    } else {
      idx++;
    }

    obj.CombinedAmount =
      idx == 0 ? obj.Amount : ar[i - 1].CombinedAmount + obj.Amount;

    ar.push(obj);

    return ar;
  }, []);

  data.matrixTarif = Object.values(
    arr.reduce((group, obj) => {
      const { Ritase } = obj;
      group[Ritase] = group[Ritase] ?? [];
      group[Ritase].push(obj);
      return group;
    }, {})
  );
}

function generateMatrixTerminalExpense(terminal, terminalExpense) {
  data.matrixTerminalExpense = terminalExpense.reduce((group, obj) => {
    let idx = group.findIndex(
      (e) => e.ExpenseID == obj.ExpenseID && e.ExpenseName == obj.ExpenseName
    );
    if (idx == -1) {
      group.push({
        ExpenseID: obj.ExpenseID,
        ExpenseName: obj.ExpenseName,
        List: JSON.parse(JSON.stringify(terminal)),
      });
      idx = group.length - 1;
    }
    const idxTerminal = group[idx].List.findIndex(
      (e) => e.TerminalID == obj.TerminalID
    );

    group[idx].List[idxTerminal] = {
      ...group[idx].List[idxTerminal],
      ...{
        Enable: true,
        ExpenseCategory: obj.ExpenseCategory,
        ExpenseValue: obj.ExpenseValue ?? 0,
        Value: obj.ExpenseCategory == "Per Person" ? obj.Value ?? 0 : 1,
        Amount:
          obj.ExpenseCategory == "Per Person"
            ? obj.Amount ?? 0
            : obj.ExpenseValue,
        ExpenseID: obj.ExpenseID,
        ExpenseName: obj.ExpenseName,
      },
    };

    return group;
  }, []);
}

function openForm(record) {
  ticketNumber.value.setRecord({ Start: "", End: "" });
  data.record = record;
}
function checkJournal(items) {
  if (Array.isArray(items) && items.length > 0)
    return items[0].JournalID !== "";
  return false;
}

const readOnly = computed({
  get() {
    return data.haveSubmitted;
  },
});
watch(
  () => data.record.OtherIncome,
  (nv) => {
    if (data.record.RevenueType == "Premi") {
      const passengerIncome = Array.isArray(data.record?.PassengerIncome)
        ? data.record.PassengerIncome
        : [];
      const otherIncome = Array.isArray(nv) ? nv : [];
      const otherExpense = Array.isArray(data.record?.OtherExpense)
        ? data.record.OtherExpense
        : [];
      const fixExpense = Array.isArray(data.record?.FixExpense)
        ? data.record.FixExpense
        : [];

      const totalAmount = data.record.OtherExpense.reduce(
        (accumulator, transaction) => {
          return accumulator + transaction.Amount;
        },
        0
      );
      const res = {
        LineNo: 0,
        ID: data.record.OtherExpense[0].ID,
        Name: "Expense",
        Amount: totalAmount,
        Notes: data.record.OtherExpense[0].Notes,
        CashBankID: data.record.OtherExpense[0].CashBankID,
        ApprovalStatus: data.record.OtherExpense[0].ApprovalStatus,
        JournalID: data.record.OtherExpense[0].JournalID,
        VoucherID: data.record.OtherExpense[0].VoucherID,
      };
      data.journalExpense = res;
      checkLineStatus([
        ...passengerIncome,
        ...otherIncome,
        ...otherExpense,
        ...fixExpense,
      ]);
    }
  },
  { deep: true }
);

function checkLineStatus(lines) {
  if (lines.length == 0) {
    data.haveDraft = true;
    data.haveSubmitted = false;
  } else {
    const drafts = lines?.filter((e) => e?.JournalID == "");
    data.haveDraft = drafts.length > 0;

    const submitteds = lines?.filter((e) => e?.JournalID != "");
    data.haveSubmitted = submitteds.length > 0;
  }
}

function editRecord(record) {
  data.isGoDetail = false;

  ritasePremi.selected = record.RitasePremi.join(" & ");
  data.selectedRevenueType = record.RevenueType;
  util.nextTickN(2, () => {
    if (record.RevenueType == "Premi") {
      generateEditMatrixTarif(record.RitaseIncome, record.RitasePassenger);
      checkLineStatus([
        record.PassengerIncome,
        ...record.OtherIncome,
        ...record.OtherExpense,
        ...record.FixExpense,
      ]);
      record.TicketNumbers = record.TicketNumbers ?? [];
    } else {
      checkLineStatus([record.DepositIncome]);
    }
    if (data.haveSubmitted) listControl.value.setFormMode("view");
    data.isGoDetail = true;
    openForm(record);
  });
}

function newRecord(record) {
  data.status = "DRAFT";
  record._id = "";
  record.TrayekName = props.trayekName;
  record.AmountDeposit = 0;
  record.RevenueType = "";
  record.RitasePremi = [];
  record.SiteEntryAssetID = props.siteEntryAssetID;
  record.ConfigDeposit = { ...data.structConfigDeposit };
  record.ConfigPremi = { ...data.structConfigPremi };
  record.TicketNumbers = [];
  data.isGoDetail = false;
  ritasePremi.selected = "";
  data.matrixTarif = [];
  data.selectedRevenueType = "";
  data.haveSubmitted = false;
  data.haveDraft = true;

  util.nextTickN(2, () => {
    resetDetail(record);
    openForm(record);
  });
}

function toogleELFormSection(nIndex, display) {
  // console.log("toogleELFormSection", nIndex, display);
  document
    .querySelectorAll(
      ".datalist-trayekritase .section:nth-child(" + nIndex + ")"
    )
    .forEach((el) => {
      el.style.display = display;
    });
}

function geDetail(record) {
  data.selectedRevenueType = record.RevenueType;
  data.haveSubmitted = false;
  data.haveDraft = true;
  if (record.RevenueType == "Premi") {
    generateNewMatrixTarif(record.RitasePremi);

    record.ConfigPremi = {
      ...record.ConfigPremi,
      ...data.ritase?.Detail?.ConfigPremi,
    };
    record.OtherIncome = [];
    // record.FixExpense =
    let dtFixExpense = [];
    let dtOtherExpense = [];
    data.ritase.FixExpense.map((e) => {
      const res = {
        ID: e._id,
        Name: e.Name,
        ExpenseTypeID: "",
        Amount: e.Value,
        Notes: "",
        Vendor: "",
        CashBankID: "",
        Urgent: false,
        ApprovalStatus: "",
        JournalID: "",
        VoucherID: "",
        ExpenseTypeID: e.ExpenseType,
        ExpenseCategory: e.ExpenseCategory,
        Value: 1,
        TotalAmount: e.Value,
      };
      if (e.ExpenseType == "Fix Expense") dtFixExpense.push(res);
      else dtOtherExpense.push(res);
    });
    record.FixExpense = [...dtFixExpense];
    record.OtherExpense = [...dtOtherExpense];
  } else {
    record.ConfigDeposit = {
      ...record.ConfigDeposit,
      ...data.ritase?.Detail?.ConfigDeposit,
    };
  }
  data.isGoDetail = true;
}

function resetDetail(record) {
  data.selectedRevenueType = "";
  data.isGoDetail = false;
  record.CategoryDeposit = "Flat";
  record.RitaseDeposit = 1;
  record.TargetFlat = 0;
  record.TargetNonFlat = 0;
  record.RevenueType = "";
  record.Ritase = "";
  record.RitaseIncome = [];
  record.RitasePassenger = [];
  record.OtherIncome = [];
  record.TerminalExpense = [];
  record.FixExpense = [];
  record.Kurs = 0;
  record.RitaseSummary = {
    TotalRitasePassenger: 0,
    TotalRitaseIncome: 0,
    TotalOtherIncome: 0,
    TotalTerminalExpense: 0,
    TotalOtherExpense: 0,
    TotalFixExpense: 0,
    TotalBonus: 0,
    TotalMethod: 0,
  };
  record.ConfigPremi = {
    ...record.ConfigPremi,
  };

  record.ConfigDeposit = {
    ...record.ConfigDeposit,
  };
  record.IsFlat = false;
  // toogleELFormSection("n+2", "none");
}

function fetchTrayek(trayekID) {
  axios.post("/bagong/trayek/get", [trayekID]).then(
    (r) => {
      data.trayek = r.data;
    },
    (e) => {
      data.trayek = undefined;
      util.showError(e);
    }
  );
}

function fetchitase(trayekID) {
  axios.post("/bagong/trayek/get-ritase", { _id: trayekID }).then(
    (r) => {
      r.data.Terminals.sort((a, b) => {
        return a.TerminalSort - b.TerminalSort;
      });
      data.ritase = r.data;
    },
    (e) => {
      data.ritase = undefined;
      util.showError(e);
    }
  );
}

function fechAsset() {
  axios.post("/bagong/asset/get", [props.assetID]).then(
    (r) => {
      fetchTrayek(r.data.DetailUnit.TrayekID);
      fetchitase(r.data.DetailUnit.TrayekID);
    },
    (e) => {
      data.ritase = undefined;
      data.trayek = undefined;
      util.showError(e);
    }
  );
}

function preSave(record) {
  if (record.RevenueType == "Premi") {
    let lineNo = 1;
    let ritaseIncome = [];
    let ritasePasanger = [];
    data.matrixTarif.forEach((el) => {
      el.forEach((el2) => {
        ritaseIncome.push(el2);
        ritasePasanger = [...ritasePasanger, ...el2.Passengers];
      });
    });

    record.OtherIncome = record.OtherIncome.map((e) => {
      e.LineNo = lineNo;
      lineNo++;
      return e;
    });
    record.OtherExpense = record.OtherExpense.map((e) => {
      e.LineNo = lineNo;
      lineNo++;
      return e;
    });
    record.FixExpense = record.FixExpense.map((e) => {
      e.LineNo = lineNo;
      lineNo++;
      return e;
    });
    record.RitaseIncome = ritaseIncome;
    record.RitasePassenger = ritasePasanger;
    record.RitaseSummary = {
      TotalRitasePassenger: total.allRitase,
      TotalRitaseIncome: total.allAmountRitase,
      TotalOtherIncome: total.otherIncome,
      TotalTerminalExpense: 0,
      TotalOtherExpense: total.otherExpense,
      TotalFixExpense: total.expense,
      TotalBonus: totalBonus.value,
      TotalMethod: totalMethod.value,
    };
    record.Revenue = revenue.value;
    record.Income = income.value;
    record.Expense = expense.value;
  } else {
    record.Revenue = revenue.value;
    record.Income = income.value;
    record.Expense = expense.value;
  }
}

function onFormFieldChanged(name, v1, v2, old, record) {
  if (name == "IsFullDeposit") {
    // listControl.value.setFormFieldAttr("RitaseDeposit", "readOnly", v1);
    // record.RitaseDeposit = 1;
    record.ConfigDeposit.TargetNonFlat =
      data.ritase.Detail.ConfigDeposit.TargetNonFlat;
    record.ConfigDeposit.TargetFlat =
      data.ritase.Detail.ConfigDeposit.TargetFlat;
    if (v1) {
      record.ConfigDeposit.TargetNonFlat =
        data.ritase.Detail.ConfigDeposit.TargetFlat;
    }
  } else if (name == "CategoryDeposit") {
    record.IsFullDeposit = false;
    record.RitaseDeposit = 0;
    record.ConfigDeposit.TargetNonFlat =
      data.ritase.Detail.ConfigDeposit.TargetNonFlat;
  } else if (name == "RitaseDeposit") {
    if (record.CategoryDeposit == "Non Flat") {
      record.ConfigDeposit.TargetNonFlat =
        data.ritase.Detail.ConfigDeposit.TargetNonFlat * v1;
    }
  }
}

// function genFormCfgTicketNum() {
//   const cfgTicketNum = createFormConfig("", true);
//   cfgTicketNum
//     .addSection("", true)
//     .addRow(
//       {
//         field: "Start1",
//         kind: "string",
//       },
//       {
//         field: "End1",
//         kind: "string",
//       }
//     )
//     .addRow(
//       {
//         field: "Start2",
//         kind: "string",
//       },
//       {
//         field: "End2",
//         kind: "string",
//       }
//     )
//     .addRow(
//       {
//         field: "Start3",
//         kind: "string",
//       },
//       {
//         field: "End3",
//         kind: "string",
//       }
//     )
//     .addRow(
//       {
//         field: "Start4",
//         kind: "string",
//       },
//       {
//         field: "End4",
//         kind: "string",
//       }
//     );
//   data.formCfgTicketNum = cfgTicketNum.generateConfig();
// }

function fetchSiteShift() {
  const url = "/bagong/sitesetup/get";
  axios.post(url, [props.site]).then(
    (r) => {
      const dataShift = r.data.Shift.map((v) => v.ShiftID);
      data.shiftItems = dataShift;
    },
    (e) => {
      data.shiftItems = [];
      util.showError(e);
    }
  );
}
function alterGridConfig(cfg) {
  cfg.fields.splice(
    3,
    0,
    helper.gridColumnConfig({ field: "Bonus", label: "Bonus", kind: "number" })
  );
}
function getRecords() {
  return listControl.value.getGridRecords();
}
defineExpose({
  getRecords,
});
onMounted(() => {
  // genFormCfgTicketNum();
  fechAsset();
  fetchSiteShift();
});
</script>

<style>
.datalist-trayekritase .suim_form .section_group .section:nth-child(n + 2) {
  display: none;
}
.datalist-trayekritase.setoran .suim_form .section_group .section:nth-child(2) {
  display: block;
}
.datalist-trayekritase.setoran
  .suim_form
  .section_group
  .section:nth-child(10) {
  display: block;
}
.datalist-trayekritase.premi
  .suim_form
  .section_group
  .section:nth-child(n + 3) {
  display: block;
}

.datalist-trayekritase .suim_form .modal_fullbg + div {
  position: fixed;
}

.datalist-trayekritase .section_group .section {
  @apply border-[1px] p-4 mb-14;
}

.datalist-trayekritase .suim_form .checkboxOffset {
  margin-top: 17px;
}

.datalist-trayekritase .section_group .section:nth-child(2) .grid:nth-child(2) {
  @apply grid-cols-3;
}

.datalist-trayekritase .section_group .section:nth-child(2) .grid {
  @apply gap-4;
}

.datalist-trayekritase .input-flat-nonflat .suim_input > div {
  @apply flex-row gap-4;
}

.datalist-trayekritase #form_inputs {
  min-height: 230px;
}
</style>
