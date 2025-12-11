<template>
  <div class="w-full flex mb-4 gap-2">
    <div class="flex justify-end grow gap-2">
      <s-button
        class="btn_secondary hover:bg-white hover:text-primary"
        label=""
        icon="keyboard-backspace"
        @click="cancelForm"
      ></s-button>
      <h1 class="grow flex">
        Generate Payroll Site:
        <s-input
          useList
          hide-label
          lookup-url="/tenant/dimension/find?DimensionType=Site"
          lookup-key="_id"
          :lookup-labels="['Label']"
          :lookup-searchs="['_id', 'Label']"
          v-model="dataParameter.SiteID"
          read-only
          :key="dataParameter.SiteID"
          class="mx-1"
        ></s-input>
        ,period:
        {{ dataParameter.FullPeriod }}
      </h1>
      <s-button
        class="btn_primary"
        label="Save"
        icon="content-save"
        @click="saveForm"
        :disabled="data.loading"
      ></s-button>
    </div>
  </div>
  <div class="w-full">
    <s-grid
      class="se-grid mt-4"
      ref="gridDetail"
      :config="data.config"
      :modelValue="data.records"
      hideControl
      hideSelect
      hideAction
      :editor="data.isEditor"
      @rowFieldChanged="rowFieldChanged"
    >
      <template
        v-for="(col, index) in data.config.fields"
        v-slot:[col.slot]="item"
      >
        <!-- {{ col.field }} {{ item.header.field }} -->
        <div v-if="col.input.readOnly" :class="col.class">
          {{ convertMoney(item.item[item.header.field]) }}
        </div>
        <div v-else :class="col.class">
          <s-input
            class="focus-within:text-black active:text-black"
            :field="item.header.field"
            kind="number"
            v-model="item.item[item.header.field]"
            @change="onHandleChange(item, item.header.field)"
          >
          </s-input>
        </div>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { inject, onMounted, reactive, ref, nextTick } from "vue";
import { SGrid, SInput, SButton, util, SModal } from "suimjs";
import moment from "moment";

const axios = inject("axios");
const gridDetail = ref(null);

const props = defineProps({
  dataParameter: {
    SiteID: { type: String, default: "" },
    DateStart: {
      type: Date,
      default: function () {
        return new Date();
      },
    },
    DateEnd: {
      type: Date,
      default: function () {
        return new Date();
      },
    },
    Period: { type: String },
  },
});

const emit = defineEmits({
  cancelForm: null,
});

const data = reactive({
  config: {
    fields: [],
    setting: {
      idField: "_id",
    },
  },
  record: {},
  records: [],
  payroll: [],
  isEditor: true,
  disabledByData: {},
  fieldsTotal: [],
  loading: false,
});

function cancelForm() {
  // data.isSelect = false;
  emit("cancelForm");
}

onMounted(() => {
  fetchData();
});

function fetchData() {
  nextTick(() => {
    nextTick(() => {
      gridDetail.value.setLoading(true);
      let payload = props.dataParameter;
      axios.post("/bagong/payroll/get", payload).then(
        async (r) => {
          gridDetail.value.setLoading(false);

          generateColumn(r.data.Details);
          data.payroll = r.data.Details;
          genDisableByData(r.data.Details);
        },
        (e) => {
          gridDetail.value.setLoading(false);
          util.showError(e);
        }
      );
    });
  });
}

function generateColumn(d) {
  let columns = [
    {
      label: "",
      field: "_id",
      readOnly: true,
      class: "",
      groupColumn: "",
      isShow: false,
    },
    {
      label: "Employee ID",
      field: "EmployeeID",
      readOnly: true,
      class: "",
      groupColumn: "",
      isShow: true,
    },
    {
      label: "Name",
      field: "Name",
      readOnly: true,
      class: "",
      groupColumn: "",
      isShow: true,
    },
    {
      label: "Dimension",
      field: "Dimension",
      readOnly: false,
      class: "",
      groupColumn: "",
      isShow: false,
    },
    {
      label: "Period",
      field: "Period",
      readOnly: false,
      class: "",
      groupColumn: "",
      isShow: false,
    },
    {
      label: "Base Salary",
      field: "BaseSalary",
      readOnly: true,
      class: "text-right",
      groupColumn: "",
      isShow: true,
    },
  ];

  d.forEach((x) => {
    if (x.Benefits?.length > 0) {
      x.Benefits.forEach((f) => {
        if (!data.fieldsTotal.includes("B@" + f.ID))
          data.fieldsTotal.push("B@" + f.ID);
        columns.push({
          field: f.ID,
          label: f.Name,
          readOnly: !f.IsManual,
          kind: "number",
          class: "bgBenefit",
          groupColumn: "Benefit",
          isShow: true,
        });
      });
    }
    if (x.Deductions?.length > 0) {
      x.Deductions.forEach((f) => {
        // if (!columns.some(c => c.field === f.ID))
        if (!data.fieldsTotal.includes("D@" + f.ID))
          data.fieldsTotal.push("D@" + f.ID);
        columns.push({
          field: f.ID,
          label: f.Name,
          readOnly: !f.IsManual,
          kind: "number",
          class: "bgDeduction",
          groupColumn: "Deduction",
          isShow: true,
        });
      });
    }
  });

  // columns = columns.filter((value, index, array) => { return array.indexOf(value) === index })
  columns = [...new Map(columns.map((item) => [item["field"], item])).values()];

  columns.push({
    field: "TPH",
    label: "Take Home Pay",
    readOnly: true,
    class: "text-right",
    groupColumn: "",
    isShow: true,
  });

  columns.forEach((x) => {
    data.config.fields.push({
      field: x.field,
      label: x.label,
      halign: "start",
      valign: "start",
      labelField: "",
      readType: x.isShow ? "show" : "hide",
      input: {
        field: x.field,
        label: x.label,
        lookupUrl: "",
        readOnly: x.readOnly,
        kind: x.kind,
        decimal: 0,
      },
      slot: "item_" + x.field,
      class: x.class,
      groupColumn: x.groupColumn,
    });
  });
  generateData(d);
}

async function generateData(d) {
  d.forEach((x) => {
    // for await (let x of d) {
    let record = {};
    let sumBenefit = 0;
    if (x.Benefits?.length > 0) {
      x.Benefits.forEach((f) => {
        // console.log(sumBenefit, f.Amount)
        sumBenefit = sumBenefit + f.Amount;
        record[f.ID] = Math.round(f.Amount);
        // console.log("total benefits:", sumBenefit)
      });
    }

    let sumDeduction = 0;
    if (x.Deductions?.length > 0) {
      x.Deductions.forEach((f) => {
        // console.log(sumDeduction, f.Amount)
        sumDeduction = sumDeduction + f.Amount;
        record[f.ID] = Math.round(f.Amount);
        // console.log("total deductons:", sumDeduction)
      });
    }
    record._id = x._id;
    record.EmployeeID = x.EmployeeID;
    record.Name = x.Name;
    record.BaseSalary = x.BaseSalary;
    record.Dimension = x.Dimension;
    record.Benefits = x.Benefits;
    record.Deductions = x.Deductions;
    record.TPH = x.BaseSalary + sumBenefit - sumDeduction;

    // await data.records.push(record)
    data.records.push(record);
    // }
  });
  gridDetail.value.setRecords(data.records);
  // console.log("records", data.records)
}

// function rowFieldChanged(n, v1, v2) {
//   console.log(n, v1, v2);
// }

// function handleChange(name, v1, v2, record) {
//   console.log(name, v1, v2, record);
// }
function onHandleChange(r, idField) {
  util.nextTickN(2, () => {
    let benefit = 0;
    let deductions = 0;
    for (let i in data.fieldsTotal) {
      let o = data.fieldsTotal[i];
      let key = o.split("@");
      if (key[0] == "B") {
        benefit += r.item[key[1]];
        if (o == "B@" + idField) {
          let f = r.item.Benefits.find((o) => o.ID == key[1]);
          if (f) f.Amount = r.item[key[1]];
        }
      }
      if (key[0] == "D") {
        deductions += r.item[key[1]];
        if (o == "D@" + idField) {
          let f = r.item.Deductions.find((o) => o.ID == key[1]);
          if (f) f.Amount = r.item[key[1]];
        }
      }
    }
    let thp = r.item.BaseSalary + benefit - deductions;
    r.item.TPH = thp;
  });
}

function saveForm() {
  nextTick(() => {
    nextTick(() => {
      data.loading = true;
      gridDetail.value.setLoading(true);
      let records = gridDetail.value.getRecords();
      records.forEach((x, i) => {
        Object.keys(x).forEach((o) => {
          let findBenefit = data.payroll[i].Benefits.find((z) => z.ID == o);
          if (findBenefit) findBenefit.Amount = x[o];

          let findDeduction = data.payroll[i].Deductions.find((z) => z.ID == o);
          if (findDeduction) findDeduction.Amount = x[o];

          data.payroll[i][o] = x[o] !== undefined ? x[o] : data.payroll[i][o];
        });
      });
      var payload = { Details: JSON.parse(JSON.stringify(data.payroll)) };
      axios.post("/bagong/payroll/save", payload).then(
        async (r) => {
          data.loading = false;
          gridDetail.value.setLoading(false);
          util.showInfo("Data has been saved");
        },
        (e) => {
          gridDetail.value.setLoading(false);
          data.loading = false;
          util.showError(e);
        }
      );
    });
  });
}

function genDisableByData(src) {
  let res = {};
  for (let i in src) {
    let o = src[i];
    for (let b in o.Benefits) {
      let ben = o.Benefits[b];
      res[o._id + "#" + ben.ID] = !ben.IsManual;
    }

    for (let b in o.Deductions) {
      let de = o.Deductions[b];
      res[o._id + "#" + de.ID] = !de.IsManual;
    }
  }
  data.disabledByData = res;
}

function isReadOnly(id, field) {
  return data.disabledByData[id + "#" + field];
}

function convertMoney(src) {
  const isNumber = typeof src == "number";
  return isNumber ? util.formatMoney(src) : src;
}
</script>
