<template>
  <div class="flex flex-col gap-2">
    <data-list
      :key="data.keyDatalist"
      ref="listControl"
      class="grid-line-items"
      title="Inventory Journal Line"
      hide-title
      no-gap
      grid-hide-select
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      :grid-hide-new="props.isFromRef || data.generalRecord.ReffNo.length > 0"
      grid-no-confirm-delete
      gridAutoCommitLine
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      :init-form-mode="data.formMode"
      :form-hide-submit="
        ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
          data.generalRecord.Status
        )
      "
      :grid-hide-control="
        ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
          data.generalRecord.Status
        )
      "
      :grid-hide-delete="
        ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
          data.generalRecord.Status
        )
      "
      :grid-editor="['', 'DRAFT'].includes(data.generalRecord.Status)"
      grid-config="/scm/inventory/journal/line/gridconfig"
      form-config="/scm/inventory/journal/line/formconfig"
      :form-fields="['ItemID', 'SKU', 'Dimension', 'InventDim', 'BatchSerials']"
      :grid-fields="['ItemID', 'Dimension', 'UnitID', 'Qty', 'UnitCost']"
      @grid-row-add="newRecord"
      @form-field-change="onFormFieldChanged"
      @form-edit-data="openForm"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="onGridRowFieldChanged"
      @grid-refreshed="gridRefreshed"
      @grid-row-save="onGridRowSave"
      @post-save="onFormPostSave"
      @alterGridConfig="alterGridConfig"
      @alterFormConfig="alterFormConfig"
      form-focus
    >
      <template #grid_Dimension="{ item }">
        <DimensionText :dimension="item.Dimension" />
      </template>
      <template #grid_ItemID="{ item, idx }">
        <s-input-sku-item
          v-model="item.ItemVarian"
          :record="item"
          :lookup-url="`/tenant/item/gets-detail?_id=${item.ItemVarian}`"
          :disabled="
            ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
              data.generalRecord.Status
            ) || props.isFromRef
          "
          @afterOnChange="onAfterOnChange"
        ></s-input-sku-item>
      </template>
      <template #grid_UnitID="{ item, idx }">
        <s-input
          v-model="item.UnitID"
          :key="item.UnitID"
          :disabled="
            ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
              data.generalRecord.Status
            ) || props.isFromRef
          "
          use-list
          :lookup-url="`/tenant/unit/gets-filter?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          @change="updateItems"
        ></s-input>
      </template>
      <template #grid_Qty="{ item }">
        <s-input
          v-if="
            ['', 'DRAFT'].includes(data.generalRecord.Status) &&
            !props.isFromRef
          "
          v-model="item.Qty"
          @change="updateItems"
          kind="number"
          class="w-full"
        ></s-input>
        <div v-else class="text-right">
          {{ helper.formatNumberWithDot(item.Qty) }}
        </div>
      </template>
      <template #grid_UnitCost="{ item }">
        <s-input
          v-if="['', 'DRAFT'].includes(data.generalRecord.Status)"
          v-model="item.UnitCost"
          kind="number"
          class="w-full"
          @change="updateItems"
        ></s-input>
        <div v-else class="text-right">
          {{ helper.formatNumberWithDot(item.UnitCost) }}
        </div>
      </template>
      <template #form_input_ItemID="{ item, idx }">
        <s-input
          ref="refItemID"
          label="ITEM"
          v-model="item.ItemID"
          :disabled="
            ['', 'DRAFT', 'SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
              data.generalRecord.Status
            )
          "
          use-list
          :lookup-url="`/tenant/item/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeItem(v1, v2, item);
            }
          "
        ></s-input>
      </template>
      <template #form_input_SKU="{ item }">
        <s-input
          v-model="item.SKU"
          label="SKU"
          :disabled="
            ['', 'DRAFT', 'SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
              data.generalRecord.Status
            )
          "
          use-list
          :lookup-url="`/tenant/itemspec/gets-info?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Description']"
          :lookup-searchs="['_id', 'SKU', 'Description']"
          class="w-full"
          @change="(...args) => handleChangeSKU(...args, item)"
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          sectionTitle="Financial Dimension"
          :readOnly="
            data.lockDimension ||
            ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
              data.generalRecord.Status
            )
          "
        ></dimension-editor>
      </template>
      <template #form_input_InventDim="{ item }">
        <DimensionInventJurnal
          v-model="item.InventDim"
          title-header="Inventory Dimension"
          :hide-field="[
            'BatchID',
            'SerialNumber',
            'SpecID',
            'InventDimID',
            'VariantID',
            'Size',
            'Grade',
          ]"
          :readOnly="
            ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
              data.generalRecord.Status
            )
          "
          :disable-field="data.disableField"
        ></DimensionInventJurnal>
      </template>
      <template #form_input_BatchSerials="{ item }">
        <BatchSerialLine
          ref="lineBatchSNConfig"
          v-model="item.BatchSerials"
          :general-record="data.generalRecord"
          :line-record="item"
        ></BatchSerialLine>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, inject, nextTick } from "vue";
import { DataList, SGrid, SInput, util } from "suimjs";
import helper from "@/scripts/helper.js";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import BatchSerialLine from "./InventoryJournalLineBatchSN.vue";
import SInputSkuItem from "./SInputSkuItem.vue";
const separatorID = "~~";
const axios = inject("axios");

const listControl = ref(null);
const refItemID = ref(null);

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  disableField: { type: Array, default: () => [] },
  transactionType: { type: String, default: () => "" },
  generalRecord: { type: Object, default: () => {} },
  isFromRef: { type: Boolean, default: false },
});
const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  generalRecord: props.generalRecord,
  records: props.modelValue.map((dt) => {
    dt.suimRecordChange = false;
    return dt;
  }),
  disableField: props.disableField,
  listcfg: {},
  keyDatalist: util.uuid(),
});

function onQueryUnitID(search, config, value) {
  let qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [config.lookupLabels[0]];
  qp.Select = config.lookupLabels;
  let idInSelect = false;
  const selectedFields = config.lookupLabels.map((x) => {
    if (x == config.lookupKey) {
      idInSelect = true;
    }
    return x;
  });
  if (!idInSelect) {
    selectedFields.push(config.lookupKey);
  }
  qp.Select = selectedFields;

  //setting search
  if (search.length > 0 && config.lookupSearchs.length > 0) {
    if (config.lookupSearchs.length == 1)
      qp.Where = {
        Field: config.lookupSearchs[0],
        Op: "$contains",
        Value: [search],
      };
    else
      qp.Where = {
        Op: "$or",
        items: config.lookupSearchs.map((el) => {
          return { Field: el, Op: "$contains", Value: [search] };
        }),
      };
  }

  if (config.multiple && value && value.length > 0 && qp.Where != undefined) {
    const whereExisting =
      value.length == 1
        ? { Op: "$eq", Field: config.lookupKey, Value: value[0] }
        : {
            Op: "$or",
            items: value.map((el) => {
              return { Field: config.lookupKey, Op: "$eq", Value: el };
            }),
          };

    qp.Where = { Op: "$or", items: [qp.Where, whereExisting] };
  }

  if (qp.Where != undefined) {
    const items = [{ Op: "$contains", Field: `_id`, Value: [value] }];
    items.push(qp.Where);
    qp.Where = {
      Op: "$and",
      items: items,
    };
  } else {
    qp.Where = { Op: "$contains", Field: `_id`, Value: [value] };
  }

  return qp;
}

function newRecord() {
  const record = {};
  record.LineNo = data.records.length + 1; //listControl.value.getRecords();
  record.ItemID = "";
  record.SKU = "";
  record.Qty = 0;
  record.UnitCost = 0;
  record.UnitID = "";
  record.Text = "";
  if (props.generalRecord?.Dimension) {
    record.Dimension = props.generalRecord?.Dimension;
  }
  data.records.push(record);
  listControl.value.setGridRecords(data.records);
  updateItems();
}
function openForm(record) {
  nextTick(() => {
    if (
      ["SUBMITTED", "READY", "REJECTED", "POSTED"].includes(
        data.generalRecord.Status
      )
    ) {
      data.formMode = "view";
      setFormMode("view");
    }
  });
}
function setFormMode(mode) {
  listControl.value.setFormMode(mode);
}
function onFormPostSave(record, index) {
  record.suimRecordChange = false;
  if (listControl.value.getFormMode() == "new") {
    data.records.push(record);
  } else {
    data.records[index] = record;
  }
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onGridRowSave(record, index) {
  record.suimRecordChange = false;
  data.records[index] = record;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onGridRowDelete(record, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  updateItems();
}
function alterGridConfig(cfg) {
  data.listcfg = cfg;
  cfg.fields.map(function (fields) {
    if (fields.field == "ItemID") {
      fields.width = "350px";
    } else if (["RemainingQty", "SKU"].includes(fields.field)) {
      fields.readType = "hide";
      fields.input.hide = true;
    } else if (
      fields.field == "UnitCost" &&
      props.transactionType == "Movement Out"
    ) {
      fields.width = "200px";
      fields.input.readOnly = true;
    } else if (["Qty", "UnitID", "UnitCost"].includes(fields.field)) {
      fields.width = "200px";
    }
    return fields;
  });
}
function alterFormConfig(cfg) {
  cfg.sectionGroups = cfg.sectionGroups.map((sectionGroup) => {
    sectionGroup.sections = sectionGroup.sections.map((section) => {
      section.rows.map((row) => {
        row.inputs.map((inputs) => {
          if (
            [
              "ItemID",
              "SKU",
              "RemainingQty",
              "Qty",
              "UnitID",
              "UnitCost",
              "Remarks",
            ].includes(inputs.field)
          ) {
            inputs.readOnly = true;
          }
          return inputs;
        });
        return row;
      });
      return section;
    });
    return sectionGroup;
  });
}

function onFormFieldChanged(name, v1, v2, old, record) {}

function onGridRowFieldChanged(name, v1, v2, old, record) {
  data.records = listControl.value.getGridRecords();
  util.nextTickN(2, () => {
    updateItems();
  });
}

function updateItems() {
  const committedRecords = data.records.filter(
    (dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
  );
  emit("update:modelValue", committedRecords);
  emit("recalc");
}
function gridRefreshed() {
  listControl.value.setGridLoading(true);
  data.records.map((r) => {
    r.ItemVarian = helper.ItemVarian(r.ItemID, r.SKU);
    return r;
  });

  listControl.value.setGridRecords(data.records);
  listControl.value.setGridLoading(false);
}

function onAfterOnChange(item) {
  let dim = [
    "Movement In",
    "Movement Out",
    "Stock Opname",
    "Transfer",
  ].includes(props.transactionType)
    ? data.generalRecord.InventDim
    : data.generalRecord.InventDimTo;
  if (!dim) {
    dim = {
      WarehouseID: "",
      AisleID: "",
      SectionID: "",
      BoxID: "",
    };
  }
  if (props.transactionType == "Movement Out") {
    item.UnitCost = 0;
  }
  item.InventDim = {
    ...item.ItemSpec,
    WarehouseID: dim.WarehouseID ? dim.WarehouseID : "",
    AisleID: dim.AisleID ? dim.AisleID : "",
    SectionID: dim.SectionID ? dim.SectionID : "",
    BoxID: dim.BoxID ? dim.BoxID : "",
  };
  if (dim.WarehouseID) {
    data.disableField.push("WarehouseID");
  }
  if (dim.AisleID) {
    data.disableField.push("AisleID");
  }
  if (dim.SectionID) {
    data.disableField.push("SectionID");
  }
  if (dim.BoxID) {
    data.disableField.push("BoxID");
  }
}
function onChangeItem(v1, v2, item) {
  item.SKU = "";
  if (typeof v1 != "string") {
    item.UnitID = "";
  } else {
    axios.post("/tenant/item/get", [v1]).then(
      (r) => {
        item.UnitID = r.data.DefaultUnitID;
        item.Item = r.data;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}
function handleChangeSKU(name, v1, v2, old, ctlRef, item, idx) {
  axios.post("/tenant/itemspec/gets-detail", [v1]).then((r) => {
    const res = r.data[0];
    let dim = [
      "Movement In",
      "Movement Out",
      "Stock Opname",
      "Transfer",
    ].includes(props.transactionType)
      ? data.generalRecord.InventDim
      : data.generalRecord.InventDimTo;
    if (!dim) {
      dim = {
        WarehouseID: "",
        AisleID: "",
        SectionID: "",
        BoxID: "",
      };
    }
    item.Text = res.Description;
    item.InventDim = {
      ...item.InventDim,
      VariantID: res.SpecVariantID,
      Size: res.SpecSizeID,
      Grade: res.SpecGradeID,
      WarehouseID: dim.WarehouseID ? dim.WarehouseID : "",
      AisleID: dim.AisleID ? dim.AisleID : "",
      SectionID: dim.SectionID ? dim.SectionID : "",
      BoxID: dim.BoxID ? dim.BoxID : "",
    };
    if (dim.WarehouseID) {
      data.disableField.push("WarehouseID");
    }
    if (dim.AisleID) {
      data.disableField.push("AisleID");
    }
    if (dim.SectionID) {
      data.disableField.push("SectionID");
    }
    if (dim.BoxID) {
      data.disableField.push("BoxID");
    }
  });
}

function getDataValue() {
  return listControl.value.getGridRecords();
}

function setDataValue(records) {
  data.records = records;
  listControl.value.setGridLoading(true);
  setTimeout(() => {
    gridRefreshed();
  }, 1000);
}

onMounted(() => {
  setTimeout(() => {}, 500);
});

defineExpose({
  getDataValue,
  setDataValue,
});
</script>
