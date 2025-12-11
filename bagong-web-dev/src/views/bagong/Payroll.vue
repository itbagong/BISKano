<template>
  <div class="w-full card p-4 payroll">
    <s-modal
      title="Payroll Site"
      class="p-4"
      :display="false"
      ref="generateModal"
      hideButtons
      @submit="confirmDelete"
    >
      <s-input
        kind="month"
        label="Period"
        class="mb-4 w-[240px]"
        v-model="data.detail.Period"
      ></s-input>
      <!-- <s-input
        ref="siteID"
        useList
        class="mb-4 min-w-[240px]"
        label="Site"
        lookup-url="/tenant/dimension/find?DimensionType=Site"
        lookup-key="_id"
        :lookup-labels="['Label']"
        :lookup-searchs="['_id', 'Label']"
        v-model="data.detail.SiteID"
        placeholder="Site"
      ></s-input> -->
      <div class="mb-3">
       <dimension-editor
          v-model="data.detail.Site"
          :default-list="profile.Dimension"
          :dim-names="['Site']"
      ></dimension-editor>
      </div>
      <s-button
        class="w-full btn_primary flex justify-center"
        label="Generate"
        @click="generateData"
      ></s-button>
    </s-modal>
    <div v-if="!data.isSelected" class="m-4">
      <s-grid
        ref="gridPayroll"
        class="se-grid"
        :config="data.config"
        :modelValue="data.records"
        @select-data="selectData"
        hide-controls
        hideDeleteButton
        hideSort
        hideRefreshButton
        hideNewButton
        :hide-edit="!profile.canUpdate"
      >
        <template #header_search>
          <div class="grow">
            <h1>Payroll</h1>
          </div>
          <s-input
            kind="text"
            label="Search"
            class="grow"
            v-model="data.search.text"
            hideLabel
          ></s-input>
          <s-input
            kind="month"
            label="Period"
            class="w-[150px]"
            v-model="data.search.period"
            hideLabel
            @change="handleChange"
          ></s-input>
          <div class="min-w-[200px]">
            <dimension-editor
                hide-label
                :default-list="profile.Dimension"
                v-model="data.search.Dimension"
                :required-fields="[]"
                :dim-names="['Site']" 
                @change="handleChange"
            ></dimension-editor>
          </div>
          
          <s-button class="btn_primary" v-if="profile.canCreate" icon="plus" @click="add"></s-button>
        </template>
        <template #item_Period="{ item }">
          {{ moment(item.Period).format("MMM YYYY") }}
        </template>
        <template #item_Benefit="{ item }">
          <div class="bgBenefit text-right">
            {{ util.formatMoney(item.Benefit) }}
          </div>
        </template>
        <template #item_Deduction="{ item }">
          <div class="bgDeduction text-right">
            {{ util.formatMoney(item.Deduction) }}
          </div>
        </template>
        <template #item_BaseSalary="{ item }">
          <div class="text-right">
            {{ util.formatMoney(item.BaseSalary) }}
          </div>
        </template>
        <template #item_TakeHomePay="{ item }">
          <div class="text-right">
            {{ util.formatMoney(item.TakeHomePay) }}
          </div>
        </template>
      </s-grid>
    </div>
    <div v-else class="m-4">
      <payroll-detail
        :dataParameter="data.detail"
        @cancelForm="cancelSelect"
      ></payroll-detail>
    </div>
  </div>
</template>
<script setup>
import { authStore } from "@/stores/auth";
import { reactive, ref, watch, onMounted, inject, nextTick } from "vue";
import helper from "@/scripts/helper.js";
import { layoutStore } from "@/stores/layout.js";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import { DataList, SGrid, SInput, SButton, SModal, util } from "suimjs";
import { useRoute } from "vue-router";
import PayrollDetail from "./widget/PayrollDetail.vue";
import moment from "moment";

layoutStore().name = "tenant";

const featureID = "Payroll";
const profile = authStore().getRBAC(featureID);
const dimensionSite = profile.Dimension?.filter(e => e.Key === 'Site').map(e=> e.Value);

const route = useRoute();
const axios = inject("axios");

const gridPayroll = ref(null);
const generateModal = ref(null);
// const theme = ref('blue-theme');

const data = reactive({
  config: {
    fields: [],
    setting: {},
  },
  record: {},
  records: [],
  search: {
    text: "",
    period: "",
    siteID: "",
    Dimension:[]
  },
  isSelected: false,
  detail: {
    SiteID: "",
    DateStart: "",
    DateEnd: "",
    Period: "",
    // DateStart: "2023-09-01T00:00:00+07:00",
    // DateEnd: "2023-09-10T00:00:00+07:00",
  },
  bgDeduction: {
    backgroundColor: "green",
    color: "white",
  },
});

onMounted(() => {
  // console.log(moment().format("YYYY-MM"))
  data.search.period = moment().format("YYYY-MM");
  generateConfig();
  getData();
});

function cancelSelect(v) {
  data.isSelected = v;
}

function add() {
  generateModal.value.show();
}

function selectData(record) {
  var date =
    record.Period.substr(0, 4) +
    "-" +
    record.Period.substr(4, 2) +
    "-01T00:00:00+07:00";
  data.detail = {
    SiteID: record.SiteID,
    SiteName: record.Name,
    FullPeriod: moment(record.Period).format("MMMM YYYY"),
    // DateStart: moment(date).startOf('month'),
    DateStart: new Date(date),
    DateEnd: moment(record.Period).endOf("month"),
    Period: record.Period,
  };

  data.isSelected = true;
}

function generateConfig() {
  let columns = [
    {
      field: "SiteID",
      label: "SiteID",
      show: false,
    },
    {
      field: "Name",
      label: "Site Name",
      show: "show",
    },
    {
      field: "Period",
      label: "Period",
      show: true,
    },
    {
      field: "BaseSalary",
      label: "Base Salary",
      show: true,
    },
    {
      field: "Benefit",
      label: "Benefit",
      show: true,
    },
    {
      field: "Deduction",
      label: "Deduction",
      show: true,
    },
    {
      field: "TakeHomePay",
      label: "Take Home Pay",
      show: true,
    },
  ];

  let fields = [];
  columns.map((x) =>
    fields.push({
      field: x.field,
      label: x.label,
      halign: "start",
      valign: "start",
      labelField: "",
      readType: x.show ? "show" : "hide",
      input: {
        lookupUrl: "",
        readOnly: x.readOnly,
      },
    })
  );

  data.config.fields = fields;
}

function getData() {
  nextTick(() => {
    nextTick(() => {
      gridPayroll.value.setLoading(true);

      var month = data.search.period.substr(5, 2);
      // var startOfMonth = moment(data.search.period).startOf('month')
      var startOfMonth =
        data.search.period.substr(0, 4) +
        "-" +
        data.search.period.substr(5, 2) +
        "-01T00:00:00+07:00";
      var endOfMonth = moment(data.search.period).endOf("month");

      let payload = {
        SiteID: data.search.Dimension.find(e=> e.Key == 'Site').Value,
        DateStart: startOfMonth,
        DateEnd: endOfMonth,
        // DateStart: "2023-09-01T00:00:00+07:00",
        // DateEnd: "2023-09-10T00:00:00+07:00",
        Skip: 0,
      };

      let arrData = [];
      axios.post("/bagong/payroll/get-site", payload).then(
        async (r) => {
          // console.log(r)
          gridPayroll.value.setLoading(false);
          arrData = JSON.parse(JSON.stringify(r.data.Details));

          data.records = r.data.Details;
          arrData.map((x, i) => {
            // data.records.push({ SiteID: x.SiteID, Name: x.Name })
            let sumBenefit = data.records[i].Benefit.reduce(
              (a, b) => +a + +b.Amount,
              0
            );
            data.records[i].Benefit = Math.round(sumBenefit);

            let sumDeduction = data.records[i].Deduction.reduce(
              (a, b) => +a + +b.Amount,
              0
            );
            data.records[i].Deduction = Math.round(sumDeduction);

            let THP = data.records[i].BaseSalary + sumBenefit - sumDeduction;
            data.records[i].TakeHomePay = Math.round(THP);
          });
        },
        (e) => {
          // gridDetail.value.setLoading(false)
          util.showError(e);
        }
      );
    });
  });
}

function generateData() {
  data.detail.SiteID =  data.detail.Site?.length == 0 ? '' :  data.detail.Site[0].Value,
  data.detail.Period = data.detail.Period;
  data.detail.DateStart = moment(data.detail.Period).startOf("month");
  data.detail.DateEnd = moment(data.detail.Period).endOf("month");
  data.detail.FullPeriod = moment(data.detail.Period).format("MMM YYYY");
  data.isSelected = !data.isSelected;
  generateModal.value.hide();
}

function handleChange(v) {
  nextTick(() => {
    nextTick(() => {
      getData();
    });
  });
}

// function onFocus(name, v1, v1, r) {
//   console.log(name, v1, v1, r);
// }
</script>
<style>
.payroll .se-grid thead tr th:has(.bgBenefit) {
  @apply bg-emerald-200;
}

.payroll .se-grid tbody tr td:has(.bgBenefit) {
  @apply bg-emerald-200;
}

.payroll .se-grid tbody tr td:has(.bgDeduction) {
  @apply bg-rose-200;
}
</style>
