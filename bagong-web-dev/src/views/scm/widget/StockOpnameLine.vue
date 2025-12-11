<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      v-model="data.records"
      class="w-full grid-line-items"
      :editor="true"
      hide-search
      hide-select
      hide-sort
      hide-new-button
      hide-delete-button
      hide-refresh-button
      :hide-detail="true"
      :hide-action="true"
      hide-paging
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
    >
      <template #item_ItemID="{ item, idx }">
        <!-- <s-input-sku-item
          v-model="item.ItemVarian"
          :record="item"
          :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
            item.ItemID,
            item.SKU
          )}`"
          :read-only="true"
        ></s-input-sku-item> -->
        {{ item.ItemName }}
      </template>
      <template #item_UnitID="{ item }">
        <!-- <s-input
          ref="refUom"
          v-model="item.UnitID"
          disabled
          readOnly
          use-list
          :lookup-url="`/tenant/unit/gets-filter?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input> -->
        {{ item.UnitName }}
      </template>
      <template #item_QtyActual="{ item }">
        <s-input
          ref="refQtyActual"
          v-model="item.QtyActual"
          :disabled="!['', 'DRAFT'].includes(props.generalRecord.Status)"
          kind="number"
          class="w-full text-right"
          @change="
            (name, v1) => {
              if (typeof v1 == 'number') {
                item.Gap = parseFloat(v1) - item.QtyInSystem;
                if (item.Gap > 0) {
                  item.Remarks = 'OVER';
                } else if (item.Gap < 0) {
                  item.Remarks = 'MINUS';
                } else if (item.Gap === 0) {
                  item.Remarks = 'OK';
                }
              } else {
                item.Gap = '';
                item.Remarks = '';
              }
            }
          "
        ></s-input>
      </template>
      <template #item_Gap="{ item }">
        <div class="w-full text-right">
          {{ item.QtyActual == "" ? "" : helper.formatNumberWithDot(item.Gap) }}
        </div>
      </template>
      <template #item_QtyInSystem="{ item }">
        <div class="w-full text-right">
          {{ helper.formatNumberWithDot(item.QtyInSystem) }}
        </div>
      </template>
      <template #item_AisleID="{ item }">
        <!-- <s-input
          ref="refAisleID"
          v-model="item.InventDim.AisleID"
          readOnly
          disabled
          use-list
          :lookup-url="`/tenant/aisle/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeAisleID(v1, v2, item);
            }
          "
        ></s-input> -->
        {{ item.AisleName }}
      </template>
      <template #item_SectionID="{ item }">
        <!-- <s-input
          ref="refSectionID"
          v-model="item.InventDim.SectionID"
          readOnly
          disabled
          use-list
          :lookup-url="`/tenant/section/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeSectionID(v1, v2, item);
            }
          "
        ></s-input> -->
        {{ item.SectionName }}
      </template>
      <template #item_BoxID="{ item }">
        <!-- <s-input
          ref="refBoxID"
          v-model="item.InventDim.BoxID"
          readOnly
          disabled
          use-list
          :lookup-url="`/tenant/box/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeBoxID(v1, v2, item);
            }
          "
        ></s-input> -->
        {{ item.BoxName }}
      </template>
      <template #item_Remarks="{ item }">
        <!-- <s-input
          ref="refRemarks"
          v-model="item.Remarks"
          readOnly
          disabled
          multiRow="2"
        ></s-input> -->
        <div
          class="text-center"
          :class="
            item.Remarks == 'MINUS'
              ? 'text-minus'
              : item.Remarks == 'OVER'
              ? 'text-over'
              : 'text-ok'
          "
        >
          {{ item.Remarks }}
        </div>
      </template>
      <template #item_Note="{ item }">
        <s-input
          ref="refNote"
          v-model="item.Note"
          :readOnly="!['', 'DRAFT'].includes(props.generalRecord.Status)"
          multiRow="2"
        ></s-input>
      </template>
      <template #item_NoteStaff="{ item }">
        <s-input
          ref="refNoteStaff"
          v-model="item.NoteStaff"
          readOnly
          disabled
          multiRow="2"
        ></s-input>
      </template>
    </s-grid>
    <!-- </data-list> -->
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { util, SInput, DataList, SGrid, loadGridConfig } from "suimjs";
import helper from "@/scripts/helper.js";
import SInputSkuItem from "./SInputSkuItem.vue";
const axios = inject("axios");
const refItemID = ref(null);
const refSKU = ref(null);
const refUom = ref(null);
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  generalRecord: { type: Object, default: () => {} },
  isLoading: { type: Boolean, default: true },
});
const separatorID = "~~";
const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  generalRecord: props.generalRecord,
  records: props.modelValue.map((dt) => {
    dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
    return dt;
  }),
  gridCfg: {
    fields: [],
    setting: {},
  },
});
const listControl = ref(null);
const listControlBatch = ref(null);
const listControlSN = ref(null);
const isAisleID = computed({
  get() {
    if (
      props.generalRecord.InventDim &&
      props.generalRecord.InventDim.AisleID
    ) {
      return true;
    } else {
      return true;
    }
  },
  set(v) {
    emit("update:item", v);
  },
});

const isSectionID = computed({
  get() {
    if (
      props.generalRecord.InventDim &&
      props.generalRecord.InventDim.SectionID
    ) {
      return true;
    } else {
      return true;
    }
  },
  set(v) {
    emit("update:item", v);
  },
});

const isBoxID = computed({
  get() {
    if (props.generalRecord.InventDim && props.generalRecord.InventDim.BoxID) {
      return true;
    } else {
      return true;
    }
  },
  set(v) {
    emit("update:item", v);
  },
});

function getDataValue() {
  return listControl.value.getRecords();
}

function setDataValue(val) {
  listControl.value.setLoading(true);
  listControl.value.setRecords(val);
  listControl.value.setLoading(false);
  updateItems();
}
function gridRefreshed() {
  data.records = props.modelValue.map((dt) => {
    if (dt.ItemID != "" && dt.SKU != "") {
      dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
    } else {
      dt.ItemVarian = "";
    }
    return dt;
  });
  util.nextTickN(2, () => {
    listControl.value.setRecords(data.records);
  });
}
function updateItems() {
  const committedRecords = data.records.filter(
    (dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
  );
  emit("update:modelValue", committedRecords);
  emit("recalc");
}

function onChangeAisleID(v1, v2, item) {
  if (typeof v1 == "string") {
    let WarehouseID = "";
    if (props.generalRecord.InventDim) {
      WarehouseID = props.generalRecord.InventDim.WarehouseID;
    }
    axios
      .post(
        `/scm/item/balance/find?ItemID=${item.ItemID}&SKU=${item.SKU}&InventoryDimension.WarehouseID=${WarehouseID}&InventoryDimension.AisleID=${v1}&InventoryDimension.SectionID=${item.SectionID}&InventoryDimension.BoxID=${item.BoxID}`
      )
      .then(
        (r) => {
          const result = r.data.at(0);
          if (result) {
            item.QtyInSystem = result.Qty;
          } else {
            item.QtyInSystem = 0;
          }
        },
        (e) => {
          util.showError(e);
        }
      );
  }
}

function onChangeSectionID(v1, v2, item) {
  if (typeof v1 == "string") {
    let WarehouseID = "";
    if (props.generalRecord.InventDim) {
      WarehouseID = props.generalRecord.InventDim.WarehouseID;
    }
    axios
      .post(
        `/scm/item/balance/find?ItemID=${item.ItemID}&SKU=${item.SKU}&InventoryDimension.WarehouseID=${WarehouseID}&InventoryDimension.AisleID=${item.AisleID}&InventoryDimension.SectionID=${v1}&InventoryDimension.BoxID=${item.BoxID}`
      )
      .then(
        (r) => {
          const result = r.data.at(0);
          if (result) {
            item.QtyInSystem = result.Qty;
          } else {
            item.QtyInSystem = 0;
          }
        },
        (e) => {
          util.showError(e);
        }
      );
  }
}

function onChangeBoxID(v1, v2, item) {
  if (typeof v1 == "string") {
    let WarehouseID = "";
    if (props.generalRecord.InventDim) {
      WarehouseID = props.generalRecord.InventDim.WarehouseID;
    }
    axios
      .post(
        `/scm/item/balance/find?ItemID=${item.ItemID}&SKU=${item.SKU}&InventoryDimension.WarehouseID=${WarehouseID}&InventoryDimension.AisleID=${item.AisleID}&InventoryDimension.SectionID=${item.SectionID}&InventoryDimension.BoxID=${v1}`
      )
      .then(
        (r) => {
          const result = r.data.at(0);
          if (result) {
            item.QtyInSystem = result.Qty;
          } else {
            item.QtyInSystem = 0;
          }
        },
        (e) => {
          util.showError(e);
        }
      );
  }
}

async function getLineQtySystem() {
  const dv = JSON.parse(JSON.stringify(listControl.value.getRecords()));
  for (let i = 0; i < dv.length; i++) {
    let WarehouseID = "";
    if (props.generalRecord.InventDim) {
      WarehouseID = props.generalRecord.InventDim.WarehouseID;
    }
    await axios
      .post(
        `/scm/item/balance/find?ItemID=${
          dv[i].ItemID ? dv[i].ItemID : ""
        }&SKU=${
          dv[i].SKU ? dv[i].SKU : ""
        }&InventoryDimension.WarehouseID=${WarehouseID}&InventoryDimension.AisleID=${
          dv[i].AisleID ? dv[i].AisleID : ""
        }&InventoryDimension.SectionID=${
          dv[i].SectionID ? dv[i].SectionID : ""
        }&InventoryDimension.BoxID=${dv[i].BoxID ? dv[i].BoxID : ""}`
      )
      .then(
        (r) => {
          const result = r.data.at(0);
          if (result) {
            dv[i].QtyInSystem = result.Qty;
          } else {
            dv[i].QtyInSystem = 0;
          }
        },
        (e) => {
          util.showError(e);
        }
      );
  }
  listControl.value.setRecords(dv);
  listControl.value.setLoading(false);
}

function getLoadLine() {
  axios
    .post(props.gridRead, {
      Skip: 0,
      Take: 0,
      Sort: ["_id"],
    })
    .then(
      (r) => {
        if (listControl.value) {
          r.data.data.map(function (v) {
            if (v.InventDim) {
              v.AisleID = v.InventDim.AisleID;
              v.BoxID = v.InventDim.BoxID;
              v.SectionID = v.InventDim.SectionID;
            }
            return v;
          });
          data.records.push(r.data.data);
          listControl.value.setRecords(r.data.data);
          updateItems();
        }
      },
      (e) => util.showError(e)
    );
}
// function alterGridConfig(cfg) {
//   const showFields = [
//     "ItemID",
//     "SKU",
//     "Text",
//     "AisleID",
//     "SectionID",
//     "BoxID",
//     "UnitID",
//     "QtyInSystem",
//     "QtyActual",
//     "Gap",
//     "Remarks",
//     "Note",
//   ];
//   loadGridConfig(axios, "/scm/inventory/journal/line/gridconfig").then((r) => {
//     loadGridConfig(axios, "/scm/inventory/dimension/gridconfig").then((dim) => {
//       const newFields = [...r.fields.filter((o) => o.field !== 'Remarks'), ...dim.fields, ...cfg.fields];
//       cfg.fields = newFields.filter((o) => showFields.includes(o.field));
//       cfg.fields = cfg.fields.map((o) => {
//         if (o.field !== 'QtyActual') {
//           o.input.disable = true;
//         }
//         return o;
//       })
//       util.nextTickN(2, () => {
//         listControl.value.gridRefreshed()
//       })
//     });
//   });
// }
watch(
  () => isAisleID,
  (nv) => {
    if (nv.value) {
      const dv = JSON.parse(JSON.stringify(listControl.value.getRecords()));
      dv.map(function (d) {
        d.AisleID = props.generalRecord.InventDim.AisleID;
        d.SectionID = props.generalRecord.InventDim.SectionID
          ? props.generalRecord.InventDim.SectionID
          : d.SectionID;
        d.BoxID = props.generalRecord.InventDim.BoxID
          ? props.generalRecord.InventDim.BoxID
          : d.BoxID;
        return d;
      });
      listControl.value.setLoading(true);
      // util.nextTickN(2, () => {
      //   listControl.value.setRecords(dv);
      //   getLineQtySystem();
      // });
    }
  },
  { deep: true }
);

watch(
  () => isBoxID,
  (nv) => {
    if (nv.value) {
      const dv = JSON.parse(JSON.stringify(listControl.value.getRecords()));
      listControl.value.setLoading(true);
      dv.map(function (d) {
        d.AisleID = props.generalRecord.InventDim.AisleID
          ? props.generalRecord.InventDim.AisleID
          : d.AisleID;
        d.SectionID = props.generalRecord.InventDim.SectionID
          ? props.generalRecord.InventDim.SectionID
          : d.SectionID;
        d.BoxID = props.generalRecord.InventDim.BoxID;
        return d;
      });
      // util.nextTickN(2, () => {
      //   listControl.value.setRecords(dv);
      //   getLineQtySystem();
      // });
    }
  },
  { deep: true }
);

watch(
  () => isSectionID,
  (nv) => {
    if (nv.value) {
      const dv = JSON.parse(JSON.stringify(listControl.value.getRecords()));
      listControl.value.setLoading(true);
      dv?.map(function (d) {
        d.AisleID = props.generalRecord.InventDim.AisleID
          ? props.generalRecord.InventDim.AisleID
          : d.AisleID;
        d.SectionID = props.generalRecord.InventDim.SectionID;
        d.BoxID = props.generalRecord.InventDim.BoxID
          ? props.generalRecord.InventDim.BoxID
          : d.BoxID;
        return d;
      });
      // util.nextTickN(2, () => {
      //   listControl.value.setRecords(dv);
      //   getLineQtySystem();
      // });
    }
  },
  { deep: true }
);

watch(
  () => props.isLoading,
  (nv) => {
    listControl.value.setLoading(nv);
  },
  { deep: true }
);

onMounted(() => {
  if (listControl.value) {
    listControl.value.setLoading(true);
  }
  loadGridConfig(axios, "/scm/stock-opname/line/gridconfig").then((res) => {
    let showFields = [
      "ItemID",
      "SKU",
      "Text",
      "AisleID",
      "SectionID",
      "BoxID",
      "UnitID",
      "QtyActual",
      "Note",
      "QtyInSystem",
      "Gap",
      "Remarks",
    ];
    // if (![""].includes(props.generalRecord.Status)) {
    //   showFields = [...showFields, "QtyInSystem", "Gap", "Remarks"];
    // }
    const cfg = res;
    loadGridConfig(axios, "/scm/inventory/journal/line/gridconfig").then(
      (r) => {
        loadGridConfig(axios, "/scm/inventory/dimension/gridconfig").then(
          (dim) => {
            const newFields = [
              ...r.fields.filter((o) => o.field !== "Remarks"),
              ...dim.fields,
              ...cfg.fields,
            ];
            cfg.fields = newFields.filter((o) => showFields.includes(o.field));
            cfg.fields = cfg.fields.map((o) => {
              if (o.field !== "QtyActual") {
                o.input.disable = true;
                o.input.readOnly = true;
                o.width = "200px";
              }
              if (o.field == "QtyActual") {
                o.width = "250px";
              }
              if (o.field == "SKU") {
                o.input.hide = true;
                o.readType = "hide";
              }
              if (o.field == "ItemID") {
                o.width = "500px";
                o.label = "Item Varian";
              }
              return o;
            });
            data.gridCfg = cfg;
            if (listControl.value) {
              listControl.value.setLoading(false);
            }
          }
        );
      }
    );
  });
});

defineExpose({
  getLoadLine,
  setDataValue,
  gridRefreshed,
});
</script>
<style>
.text-minus {
  color: red;
  font-weight: 600;
}

.text-ok {
  color: black;
  font-weight: 600;
}

.text-over {
  color: green;
  font-weight: 600;
}
/* .tb-line > div:nth-child(2) > div {
  overflow-x: auto;
  padding-bottom: 100px;
}
.tb-line > div:nth-child(2) > div > table {
  width: calc(100% + 40%) !important;
} */
</style>
