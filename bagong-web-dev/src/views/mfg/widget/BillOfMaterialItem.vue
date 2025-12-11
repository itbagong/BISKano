<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-material grid-line-items"
      editor
      hide-search
      hide-sort
      :hide-new-button="false"
      hide-refresh-button
      :hide-detail="true"
      :hide-select="true"
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @delete-data="onDelete"
    >
      <template #item_ItemID="{ item, idx }">
        <!-- <s-input
          ref="refItemID"
          v-model="item.ItemID"
          :disabled="false"
          use-list
          :lookup-url="`/tenant/item/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeItem(v1, v2, item);
            }
          "
        ></s-input> -->
        <s-input-sku-item
          v-model="item.ItemVarian"
          :record="item"
          :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
            item.ItemID,
            item.SKU
          )}`"
        ></s-input-sku-item>
      </template>
      <!-- <template #item_SKU="{ item }">
        <s-input
          ref="refSKU"
          v-model="item.SKU"
          :disabled="false"
          use-list
          :lookup-url="`/tenant/itemspec/gets-info?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Description']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeSKU(v1, v2, item);
            }
          "
        ></s-input>
      </template> -->
      <template #item_Description="{ item }">
        <s-input
          ref="refDescription"
          v-model="item.Description"
          :disabled="true"
          multiRow="2"
          :keepErrorSection="true"
        ></s-input>
      </template>
      <template #item_UoM="{ item }">
        <s-input
          ref="refUom"
          v-model="item.UoM"
          :disabled="false"
          use-list
          :lookup-url="`/tenant/unit/gets-filter?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
      <template #item_Qty="{ item }">
        <s-input
          ref="refQty"
          v-model="item.Qty"
          kind="number"
          @change="
            (name, v1) => {
              item.Total = v1 * item.UnitPrice;
            }
          "
        ></s-input>
      </template>
      <template #item_UnitPrice="{ item }">
        <s-input
          ref="refUnitPrice"
          v-model="item.UnitPrice"
          kind="number"
          @change="
            (name, v1) => {
              item.Total = v1 * item.Qty;
            }
          "
        ></s-input>
      </template>
      <template #item_Total="{ item }">
        <div class="text-right">
          {{ util.formatMoney(item.Total, { decimal: 0 }) }}
        </div>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid } from "suimjs";
import helper from "@/scripts/helper.js";
import SInputSkuItem from "../../scm/widget/SInputSkuItem.vue";
const axios = inject("axios");
const refItemID = ref(null);
const refSKU = ref(null);
const refUom = ref(null);
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
  itemID: { type: String, default: () => "" },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  readOnly: { type: Boolean, defaule: false },
  hideDetail: { type: Boolean, defaule: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  typeMoveIn: "line",
  gridCfg: {},
});

function newRecord() {
  const record = {};
  record.PurchaseRequestID = props.itemID;
  record.ItemVarian = "";
  record.ItemID = "";
  record.SKU = "";
  record.Description = "";
  record.UoM = "";
  record.Qty = 0;
  record.UnitPrice = 0;
  record.Total = 0;
  listControl.value.setRecords([...listControl.value.getRecords(), record]);
}

function onDelete(record, index) {
  const newRecords = listControl.value.getRecords().filter((dt, idx) => {
    return idx != index;
  });
  listControl.value.setRecords(newRecords);
}

function onChangeItem(v1, v2, item) {
  item.UoM = "";
  item.SKU = "";
  item.Description = "";
  if (typeof v1 == "string") {
    axios.post("/tenant/item/get", [v1]).then(
      (r) => {
        item.UoM = r.data.DefaultUnitID;
        item.UnitPrice = r.data.CostUnit;
        item.Total = r.data.CostUnit * item.Qty;
        item.Item = r.data;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function onChangeSKU(v1, v2, item) {
  item.Description = "";
  if (typeof v1 == "string") {
    axios.post("/tenant/itemspec/gets-detail", [v1]).then(
      (r) => {
        item.Description = r.data.length == 0 ? "" : r.data[0].Description;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function getDataValue() {
  return listControl.value.getRecords();
}

function onSelectDataLine() {
  return listControl.value.getRecords().filter((el) => el.isSelected == true);
}

onMounted(() => {
  loadGridConfig(axios, props.gridConfig).then(
    (r) => {
      let tbLine = ["ItemID", "UoM", "Qty", "UnitPrice", "Total"];
      const _fields = r.fields.filter((o) => {
        if (["UoM", "Qty", "UnitPrice", "Total"].includes(o.field)) {
          o.width = "150px";
        } else {
          o.width = "400px";
        }
        if (["Total"].includes(o.field)) {
          o.input.disable = true;
        }
        o.idx = tbLine.indexOf(o.field);
        return tbLine.includes(o.field);
      });
      data.gridCfg = {
        ...r,
        fields: _fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
    },
    (e) => util.showError(e)
  );
  axios
    .post(props.gridRead, {
      Skip: 0,
      Take: 0,
      Sort: ["_id"],
    })
    .then(
      (r) => {
        if (listControl.value) {
          console.log(r.data.data);
          r.data.data.map((dt) => {
            dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
            return dt;
          });

          listControl.value.setRecords(r.data.data);
        }
      },
      (e) => util.showError(e)
    );
});
defineExpose({
  getDataValue,
  onSelectDataLine,
});
</script>
