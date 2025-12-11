<template>
  <data-list
    class="card mb-4"
    ref="listControl"
    grid-config="/scm/inventory/receive/gridconfig"
    :grid-read="
      '/scm/inventory/receive/find-by-vendor?VendorID=' + modelValue.VendorID
    "
    grid-mode="grid"
    grid-sort-field="Created"
    grid-sort-direction="desc"
    @alter-grid-config="onAlterConfig"
    :init-app-mode="'grid'"
    grid-hide-new
    grid-hide-delete
    grid-hide-edit
    :grid-fields="['WarehouseID']"
    :grid-page-size="10"
    @grid-refreshed="onGridRefreshed"
    @grid-check-uncheck="onCheckUncheck"
  >
    <template #grid_header_buttons_2="{ config }">
      <s-button
        class="bg-primary text-white font-bold w-full flex justify-center"
        label="Generate"
        @click="onGenerate"
        :disabled="data.loadingGenerate"
      />
    </template>
    <template #grid_WarehouseID="{ item }">
      <s-input
        hide-label
        v-model="item.WarehouseID"
        use-list
        lookup-url="/tenant/warehouse/find"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookup-searchs="['_id', 'Name']"
        @key="item"
        read-only
      />
    </template>
  </data-list>
</template>
<script setup>
import { reactive, ref, computed, inject, watch, onMounted } from "vue";
import { DataList, util, SButton, SInput, SModal } from "suimjs";
import helper from "@/scripts/helper.js";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  jurnalType: { type: String, default: () => "" },
  jurnalId: { type: String, default: () => "" },
});
const axios = inject("axios");
const emit = defineEmits({
  "update:modelValue": null,
  close: null,
});
const data = reactive({
  otherExpense: [],
  loadingGenerate: false,
  selectedRecord: {},
});

const listControl = ref(null);

function onAlterConfig(config) {
  config.setting.keywordFields = ["_id", "ReffNo"];
  let f = config.fields.find((o) => o.field == "Dimension");
  if (f) f.readType = "hide";
}

async function onGenerate() {
  data.loadingGenerate = true;
  let mv = props.modelValue;
  let dt = selecteds.value;
  let listGrNo = dt.map((o) => o._id);
  let listPoNo = [];
  let resLines = [];
  let Lines = [];
  let MainBalanceAccount = "";
  let amountDiscByPO = [];

  for (let i in dt) {
    let o = dt[i];
    if (i == 0) {
      MainBalanceAccount = o.MainBalanceAccount;
    }
    updateTagAttachment(
      [`${props.jurnalType}_${props.jurnalType} ${props.jurnalId}`],
      [`GR_${o._id}`]
    );

    for (let j in o.ReffNo) {
      let poid = o.ReffNo[j];
      await getOtherExpense(poid);
    }

    for (let j in o.Lines) {
      let ob = o.Lines[j];
      let po = [...o.ReffNo];
      let flineid = ob.References.find((x) => x.Key == "POLineID");
      let ref = [
        {
          Key: "PO Ref No",
          Value: po.length > 0 ? po[po.length - 1] : "",
        },
        {
          Key: "PO Line ID",
          Value: flineid ? flineid.Value : "",
        },
        { Key: "GR Ref No", Value: o._id },
      ];

      if (j == 0) {
        amountDiscByPO.push(ob.DiscountGeneral.DiscountAmount);
      }
      ob.References = ref;
    }

    listPoNo = [...listPoNo, ...o.ReffNo];
    Lines = [...Lines, ...o.Lines];
  }
  listPoNo = [...new Set(listPoNo)];

  let obj = {
    LineNo: 0,
    TagObjectID1: {
      AccountType: "INVENTORY",
      AccountID: "",
      AccountIDs: [],
    },
    Account: {
      AccountType: "LEDGERACCOUNT",
      AccountID: MainBalanceAccount,
      AccountIDs: [],
    },
    Qty: 0,
    Text: "",
    UnitID: "",
    PriceEach: "",
    Amount: 0,
    DiscountType: "",
    Discount: 0,
    Taxable: true,
    TaxCodes: [],
    References: [],
  };

  for (let i in Lines) {
    let o = Lines[i];
    let ob = helper.cloneObject(obj);
    ob.LineNo = parseInt(i) + 1;
    ob.TagObjectID1.AccountID = o.ItemID;
    ob.Text = o.Text;
    ob.UnitID = o.UnitID;
    getSKUName(o.SKU, i);
    ob.Qty = o.Qty;
    ob.PriceEach = o.UnitCost;
    ob.DiscountType = o.DiscountType;
    ob.Amount = setAmount(o);
    ob.Discount = o.DiscountValue;
    ob.TaxCodes = o.TaxCodes;
    ob.References = o.References;
    resLines.push(ob);
  }

  for (let i in data.otherExpense) {
    let o = data.otherExpense[i];
    let ob = helper.cloneObject(obj);
    ob.LineNo = resLines.length + (parseInt(i) + 1);
    ob.TagObjectID1.AccountType = "EXPENSE";
    ob.TagObjectID1.AccountID = o.Expenses;
    ob.Account.AccountID = "514202";
    ob.UnitID = "PCS";
    ob.Qty = 1;
    ob.PriceEach = o.Amount;
    ob.Amount = o.Amount;
    ob.Taxable = false;
    resLines.push(ob);
  }

  // set value
  let po = mv.References.find((o) => o.Key == "PO Ref No.");
  let gr = mv.References.find((o) => o.Key == "GR Ref No.");
  if (po) po.Value = listPoNo.join();
  if (gr) gr.Value = listGrNo.join();
  mv.Lines = resLines;

  mv.TransactionType = "Good Receive";
  mv.HeaderDiscountType = "fixed";
  mv.HeaderDiscountValue = amountDiscByPO.reduce((a, b) => a + b, 0);
  data.loadingGenerate = false;
  emit("close");
}

function setAmount(dt) {
  let res = 0;
  let total = dt.Qty * dt.UnitCost;
  if (dt.DiscountType == "fixed") {
    res = total - dt.DiscountValue;
  } else {
    res = total - total * (dt.DiscountValue / 100);
  }
  return res;
}

async function getSKUName(id, idx) {
  if (!id) return;
  const url = "/tenant/itemspec/gets-info?_id=" + id;
  await axios.post(url).then(
    (r) => {
      if (r.data.length > 0) {
        let o = r.data[0];
        props.modelValue.Lines[idx].Text = o.OtherName;
      }
    },
    (e) => util.showError(e)
  );
}

function updateTagAttachment(addTags, whereTags) {
  const payload = {
    Addtags: addTags,
    Tags: whereTags,
  };
  if (payload.Tags.length > 0) {
    helper.updateTags(axios, payload);
  }
}

async function getOtherExpense(id) {
  if (!id) return;
  const url = `/scm/purchase/order/find?_id=${id}`;
  await axios.post(url).then(
    (r) => {
      for (let i in r.data) {
        let o = r.data[i];
        data.otherExpense = [...data.otherExpense, ...o.OtherExpenses];
      }
    },
    (e) => util.showError(e)
  );
}

function onGridRefreshed() {
  let dt = listControl.value.getGridRecords();
  for (let i in dt) {
    let o = dt[i];
    if (data.selectedRecord[o._id]) o.isSelected = true;
  }
}

function onCheckUncheck(o) {
  if (o.isSelected) {
    data.selectedRecord[o._id] = o;
  } else {
    delete data.selectedRecord[o._id];
  }
}

const selecteds = computed(() => {
  return Object.keys(data.selectedRecord).map(
    (key) => data.selectedRecord[key]
  );
});
</script>
