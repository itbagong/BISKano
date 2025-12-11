<template>
  <div class="flex flex-col gap-2">
    <s-grid
      v-show="data.typeMoveIn == 'line'"
      ref="listControl"
      class="w-full tb-line grid-line-items"
      :editor="data.value.Status == '' || data.value.Status == 'Draft'"
      hide-search
      hide-select
      hide-sort
      hide-new-button
      hide-delete-button
      hide-refresh-button
      :hide-detail="true"
      :hide-action="props.item.Status == 'Closed'"
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
    >
      <template #item_ItemID="{ item, idx }">
        <s-input
          ref="refItemID"
          v-model="item.ItemID"
          :disabled="true"
          use-list
          :lookup-url="`/tenant/item/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
      <template #item_SKU="{ item }">
        <s-input
          ref="refSKU"
          v-model="item.SKU"
          :disabled="true"
          use-list
          :lookup-url="`/tenant/itemspec/find?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['SKU']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
      <template #item_UoM="{ item }">
        <s-input
          ref="refUom"
          v-model="item.UoM"
          :disabled="true"
          use-list
          :lookup-url="`/tenant/unit/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
      <template #item_Qty="{ item }">
        <s-input
          ref="refQTY"
          v-model="item.Qty"
          :disabled="!['Open', 'PartialIssued'].includes(props.item.Status)"
          kind="number"
          class="w-full"
        ></s-input>
      </template>
      <template #item_AisleID="{ item }">
        <s-input
          ref="refAisleID"
          v-model="item.AisleID"
          :disabled="
            !['Open', 'PartialReceived'].includes(props.item.Status) ||
            isAisleID
          "
          use-list
          :lookup-url="`/tenant/aisle/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
      <template #item_SectionID="{ item }">
        <s-input
          ref="refSectionID"
          v-model="item.SectionID"
          :disabled="
            !['Open', 'PartialReceived'].includes(props.item.Status) ||
            isSectionID
          "
          use-list
          :lookup-url="`/tenant/section/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
      <template #item_BoxID="{ item }">
        <s-input
          ref="refBoxID"
          v-model="item.BoxID"
          :disabled="
            !['Open', 'PartialReceived'].includes(props.item.Status) || isBoxID
          "
          use-list
          :lookup-url="`/tenant/box/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
      <template #item_buttons_1="{ item, idx }">
        <a
          v-if="!['Closed'].includes(props.item.Status)"
          href="#"
          @click="onSelectData(item, idx)"
          class="edit_action"
        >
          <mdicon
            name="pencil"
            width="16"
            alt="edit"
            class="cursor-pointer hover:text-primary"
          />
        </a>
      </template>
    </s-grid>
    <div v-show="data.typeMoveIn == 'Batch'">
      <div class="mb-2">
        <s-form
          ref="formCtlBatch"
          v-model="data.recordBatch"
          :keep-label="true"
          :config="data.IsEnabledItemBatch ? data.fromCfgBatch : data.fromCfgSN"
          class="pt-2"
          :auto-focus="true"
          :hide-submit="true"
          :hide-cancel="true"
          mode="view"
        >
        </s-form>
      </div>
      <div class="mt-5">
        <s-grid
          ref="listControlBatch"
          class="w-full grid-line-items"
          :editor="true"
          hide-search
          hide-select
          hide-sort
          :hide-detail="false"
          :hide-new-button="['Closed'].includes(props.item.Status)"
          :hide-delete-button="['Closed'].includes(props.item.Status)"
          @new-data="newRecordBatch"
          @delete-data="onDeleteBatch"
          @selectData="onSelectDataBatch"
          auto-commit-line
          no-confirm-delete
          :config="data.gridCfgBatch"
          hide-refresh-button
          form-keep-label
        >
          <template #item_BatchID="{ item }">
            <s-input
              ref="refBatchID"
              v-model="item.BatchID"
              class="w-full"
              :disabled="props.item.Status == 'Closed'"
              :keepErrorSection="true"
              use-list
              :lookup-url="`/tenant/itembatch/find?ItemID=${data.recordBatch.ItemID}`"
              lookup-key="_id"
              :lookup-labels="['_id']"
              :lookup-searchs="['_id', '_id']"
            ></s-input>
          </template>
          <template #item_Qty="{ item }">
            <s-input
              ref="refBatchQty"
              v-model="item.Qty"
              class="w-full"
              :disabled="props.item.Status == 'Closed'"
              kind="number"
            ></s-input>
          </template>
          <template #header_buttons_1="{ item }">
            <s-button
              icon="rewind"
              class="btn_warning back_btn"
              label="Back"
              @click="data.typeMoveIn = 'line'"
            />
          </template>
        </s-grid>
      </div>
    </div>
    <div v-show="data.typeMoveIn == 'SerialNumber'">
      <div class="mb-2">
        <s-form
          ref="formCtlBatch"
          v-model="data.recordBatch"
          :keep-label="true"
          :config="data.IsEnabledItemBatch ? data.fromCfgBatch : data.fromCfgSN"
          class="pt-2"
          :auto-focus="true"
          :hide-submit="true"
          :hide-cancel="true"
          mode="view"
        >
        </s-form>
      </div>
      <div class="mt-5">
        <s-grid
          ref="listControlSN"
          class="w-full grid-line-items"
          :editor="true"
          hide-search
          hide-select
          hide-sort
          hide-action
          :hide-detail="true"
          hide-new-button
          hide-delete-button
          auto-commit-line
          no-confirm-delete
          :config="data.gridCfgSN"
          hide-refresh-button
          form-keep-label
        >
          <template #header_buttons_1="{ item }">
            <s-button
              icon="rewind"
              class="btn_warning back_btn"
              label="Back"
              @click="
                data.IsEnabledItemBatch
                  ? (data.typeMoveIn = 'Batch')
                  : (data.typeMoveIn = 'line')
              "
            />
          </template>
          <template #item_BatchID="{ item, idx }">
            <s-input
              v-if="data.IsEnabledItemBatch"
              ref="refBatchID"
              v-model="item.BatchID"
              :disabled="true"
              class="w-full"
            ></s-input>
            <s-input
              v-else
              ref="refGridBatchID"
              v-model="item.BatchID"
              class="w-full"
              :disabled="props.item.Status == 'Closed'"
              :keepErrorSection="true"
              use-list
              :lookup-url="`/tenant/itembatch/find?ItemID=${data.recordBatch.ItemID}`"
              lookup-key="_id"
              :lookup-labels="['_id']"
              :lookup-searchs="['_id', '_id']"
            ></s-input>
          </template>
          <template #item_SerialNumberID="{ item }">
            <s-input
              ref="refSerialNumberID"
              v-model="item.SerialNumberID"
              :disabled="props.item.Status == 'Closed'"
              class="w-full"
              :keepErrorSection="true"
              use-list
              :lookup-url="`/tenant/itemserial/find?ItemID=${data.recordBatch.ItemID}`"
              lookup-key="_id"
              :lookup-labels="['_id']"
              :lookup-searchs="['_id', '_id']"
            ></s-input>
          </template>
        </s-grid>
      </div>
    </div>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import {
  loadGridConfig,
  createFormConfig,
  util,
  SInput,
  SButton,
  SGrid,
  SForm,
} from "suimjs";
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
  formMode: { type: String, default: () => "new" },
  activeFields: { type: Array, default: () => [] },
  readOnly: { type: Boolean, defaule: false },
  hideDetail: { type: Boolean, defaule: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);
const listControlBatch = ref(null);
const listControlSN = ref(null);
const isAisleID = computed({
  get() {
    if (
      props.item.InventoryDimension &&
      props.item.InventoryDimension.AisleID
    ) {
      return true;
    } else {
      return false;
    }
  },
  set(v) {
    emit("update:item", v);
  },
});

const isSectionID = computed({
  get() {
    if (
      props.item.InventoryDimension &&
      props.item.InventoryDimension.SectionID
    ) {
      return true;
    } else {
      return false;
    }
  },
  set(v) {
    emit("update:item", v);
  },
});

const isBoxID = computed({
  get() {
    if (props.item.InventoryDimension && props.item.InventoryDimension.BoxID) {
      return true;
    } else {
      return false;
    }
  },
  set(v) {
    emit("update:item", v);
  },
});

const data = reactive({
  value: props.modelValue,
  typeMoveIn: "line",
  fromCfgBatch: {},
  fromCfgSN: {},
  gridCfg: {},
  gridCfgBatch: {},
  gridCfgSN: {},
  disabledBatch: true,
  IsEnabledItemBatch: false,
  recordBatch: {
    ItemID: "",
    SKU: "",
    Description: "",
    Qty: 0,
    Item: {
      PhysicalDimension: {
        IsEnabledItemBatch: false,
        IsEnabledItemSerial: false,
      },
    },
    Unit: {},
  },
  recordSerialNumber: {
    MovementInID: "",
    ItemID: "",
    BatchID: "",
    Qty: 0,
    QtySerialNumber: 0,
    Unit: {},
  },
  tempListBatch: [],
  tempListSerialNumber: [],
});
function getBatchItem() {
  const batch = data.tempListBatch.filter(function (v) {
    return (
      v.GoodIssueID == props.itemID &&
      v.ItemID == data.recordBatch.ItemID &&
      v.SKU == data.recordBatch.SKU
    );
  });
  return batch;
}
function onSelectData(record, index) {
  if (record.Qty == 0 || (record.Qty == 0 && props.item.Status != "Closed")) {
    return util.showError("There are fields Qty that are 0");
  }

  // const isBatch = record.Item.PhysicalDimension.IsEnabledItemBatch;
  // const isSerial = record.Item.PhysicalDimension.IsEnabledItemSerial;
  const isBatch = true;
  const isSerial = true;
  const fnBatch = () => {
    data.typeMoveIn = "Batch";
    axios.post("/tenant/unit/get", [record.UoM]).then(
      (r) => {
        const Unit = r.data;
        record.Unit = Unit;
        data.recordBatch = record;
        data.IsEnabledItemBatch = true;
        const batch = data.tempListBatch.filter(function (v) {
          return (
            v.GoodIssue == props.itemID &&
            v.ItemID == record.ItemID &&
            v.SKU == record.SKU
          );
        });

        const removeBatch = data.tempListBatch.filter(function (v) {
          return v.ItemID != record.ItemID && v.SKU != record.SKU;
        });

        const removeSN = data.tempListSerialNumber.filter(function (v) {
          return v.ItemID != record.ItemID && v.SKU != record.SKU;
        });
        if (batch.length != record.Qty) {
          data.tempListBatch = removeBatch;
          data.tempListSerialNumber = removeSN;
        }
        listControlBatch.value.setRecords(batch);
      },
      (e) => {
        util.showError(e);
      }
    );
  };
  const fnSerial = () => {
    data.typeMoveIn = "SerialNumber";
    axios.post("/tenant/unit/get", [record.UoM]).then(
      (r) => {
        const Unit = r.data;
        data.IsEnabledItemBatch = false;
        record.Unit = Unit;
        data.recordSerialNumber = record;
        let listSerial = [];
        let Ratio = record.Qty * Unit.Ratio;
        for (let i = 0; i < Ratio; i++) {
          listSerial.push({
            GoodIssue: props.itemID,
            ItemID: record.ItemID,
            SKU: record.SKU,
            BatchID: "",
            SerialNumberID: "",
            status: "new",
          });
        }
        let tempSN = getTempSerialItem(record);
        if (tempSN.length == 0) {
          data.tempListSerialNumber = [
            ...data.tempListSerialNumber,
            ...listSerial,
          ];
          listControlSN.value.setRecords(data.tempListSerialNumber);
        } else {
          if (tempSN.length != Ratio) {
            let removeSN = data.tempListSerialNumber.filter(function (v) {
              return v.ItemID != record.ItemID && v.SKU != record.SKU;
            });
            data.tempListSerialNumber = removeSN;
            listControlSN.value.setRecords(listSerial);
          } else {
            listControlSN.value.setRecords(tempSN);
          }
        }
      },
      (e) => {
        util.showError(e);
      }
    );
  };
  if (record.ItemID != "" && record.SKU != "" && record.Qty != 0) {
    if (isBatch) {
      fnBatch();
    } else if (isSerial) {
      fnSerial();
    } else if (!isBatch && !isSerial) {
      fnBatch();
    }
  }
}
function newRecordBatch() {
  const record = {};
  record._id = util.uuid();
  record.GoodIssue = props.itemID;
  record.ItemID = data.recordBatch.ItemID;
  record.SKU = data.recordBatch.SKU;
  record.BatchID = "";
  record.Qty = 0;

  const batchItem = getBatchItem();
  const sum = batchItem.reduce((accumulator, object) => {
    return accumulator + object.Qty;
  }, 0);
  if (sum > data.recordBatch.Qty) {
    util.showError("Quantity does not match");
  } else {
    data.tempListBatch.push(record);
    if (batchItem.length == 0) {
      listControlBatch.value.setRecords(data.tempListBatch);
    } else {
      const batch = getBatchItem();
      listControlBatch.value.setRecords(batch);
    }
  }
}
function onDeleteBatch(record, index) {
  const delRecords = listControlBatch.value.getRecords().find((dt, idx) => {
    return idx == index;
  });
  const indexOfObject = data.tempListBatch.findIndex((obj, idx) => {
    return obj._id == delRecords._id;
  });
  data.tempListBatch.splice(indexOfObject, 1);

  const SN = data.tempListSerialNumber.filter(function (v) {
    return v.BatchID != delRecords.BatchID;
  });
  data.tempListSerialNumber = SN;
  const batch = getBatchItem();
  listControlBatch.value.setRecords(batch);
}
function onSelectDataBatch(record, index) {
  const sum = data.tempListBatch.reduce((accumulator, object) => {
    return accumulator + object.Qty;
  }, 0);
  if (sum > data.recordBatch.Qty && data.value.Status != "Closed") {
    return util.showError("Total batch is greater than qty");
  }

  let TbBatch = listControlBatch.value.getRecords();
  const isExist = TbBatch.filter(function (v) {
    return v.BatchID == record.BatchID;
  });
  if (isExist.length > 1) {
    return util.showError("There is the same batch");
  }

  if (record.BatchID != "" && record.Qty > 0) {
    data.typeMoveIn = "SerialNumber";
    listControlSN.value.setRecords([]);
    data.IsEnabledItemBatch = true;
    record.Ratio = data.recordBatch.Unit.Ratio;
    console.log(data.recordBatch);
    data.recordSerialNumber = record;
    let listBatch = [];
    for (let i = 0; i < record.Qty * record.Ratio; i++) {
      listBatch.push({
        GoodIssue: record.GoodIssue,
        ItemID: record.ItemID,
        BatchID: record.BatchID,
        SKU: record.SKU,
        SerialNumberID: "",
      });
    }
    const tempSN = data.tempListSerialNumber.filter(function (v) {
      return (
        v.GoodIssue == props.itemID &&
        v.ItemID == record.ItemID &&
        v.BatchID == record.BatchID &&
        v.SKU == record.SKU
      );
    });
    if (tempSN.length == 0) {
      data.tempListSerialNumber = [...data.tempListSerialNumber, ...listBatch];
      listControlSN.value.setRecords(listBatch);
    } else {
      if (tempSN.length != record.Qty * record.Ratio) {
        const removeSN = data.tempListSerialNumber.filter(function (v) {
          return (
            v.ItemID != record.ItemID &&
            v.BatchID != record.BatchID &&
            v.SKU != record.SKU
          );
        });
        data.tempListSerialNumber = removeSN;
        listControlSN.value.setRecords([...sn, ...listBatch]);
      } else {
        listControlSN.value.setRecords(sn);
      }
    }
  } else {
    util.showError("There are fields that are empty");
  }
}

function getDataValue() {
  return listControl.value.getRecords();
}

function getDataValueBatch() {
  return data.tempListBatch;
}
function getDataValueSerialNumber() {
  return data.tempListSerialNumber;
}

function genCfgFrmBatch() {
  const cfg = createFormConfig("", true);
  cfg.addSection("", true).addRow(
    {
      field: "ItemID",
      label: "Item",
      kind: "text",
      useList: true,
      allowAdd: false,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["_id", "Name"],
      lookupUrl: "/tenant/item/find",
    },
    {
      field: "SKU",
      kind: "text",
      label: "SKU",
      useList: true,
      allowAdd: false,
      lookupKey: "_id",
      lookupLabels: ["SKU"],
      lookupSearchs: ["_id", "SKU"],
      lookupUrl: "/tenant/itemspec/find",
    },
    {
      field: "Qty",
      kind: "number",
      label: "Quantity",
    },
    {
      field: "Description",
      kind: "text",
      label: "Description",
      multiRow: "3",
    }
  );
  data.fromCfgBatch = cfg.generateConfig();
}

onMounted(() => {
  genCfgFrmBatch();
  loadGridConfig(axios, props.gridConfig).then(
    (r) => {
      let tbLine = [
        "ItemID",
        "SKU",
        "Description",
        "Qty",
        "UoM",
        "Remarks",
        "AisleID",
        "SectionID",
        "BoxID",
      ];
      let InventoryDimension = [
        {
          field: "AisleID",
          kind: "Text",
          label: "AisleID",
          readType: "show",
          input: {
            field: "AisleID",
            label: "AisleID",
            hint: "",
            hide: false,
            placeHolder: "AisleID",
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
          },
        },
        {
          field: "SectionID",
          kind: "Text",
          label: "SectionID",
          readType: "show",
          input: {
            field: "SectionID",
            label: "SectionID",
            hint: "",
            hide: false,
            placeHolder: "SectionID",
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
          },
        },
        {
          field: "BoxID",
          kind: "Text",
          label: "BoxID",
          readType: "show",
          input: {
            field: "BoxID",
            label: "BoxID",
            hint: "",
            hide: false,
            placeHolder: "BoxID",
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
          },
        },
      ];
      const _fields = [...r.fields, ...InventoryDimension].filter((o) => {
        if (["UoM", "QtyInSystem", "QtyActual", "Gap"].includes(o.field)) {
          o.width = "300px";
        } else if (
          [
            "Description",
            "Remarks",
            "NoteStaff",
            "Note",
            "NoteAdjustment",
          ].includes(o.field)
        ) {
          o.width = "800px";
        } else {
          o.width = "400px";
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
  loadGridConfig(axios, "/scm/good-issue/batch/gridconfig").then(
    (r) => {
      data.gridCfgBatch = r;
    },
    (e) => util.showError(e)
  );
  loadGridConfig(axios, "/scm/good-issue/serial/gridconfig").then(
    (r) => {
      const _fields = r.fields.filter((o) =>
        ["BatchID", "SerialNumberID"].includes(o.field)
      );
      data.gridCfgSN = { ...r, fields: _fields };
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
          r.data.data.map(function (v) {
            if (v.InventoryDimension) {
              v.AisleID = v.InventoryDimension.AisleID;
              v.BoxID = v.InventoryDimension.BoxID;
              v.SectionID = v.InventoryDimension.SectionID;
            }
            return v;
          });
          listControl.value.setRecords(r.data.data);
        }
      },
      (e) => util.showError(e)
    );
});
defineExpose({
  getDataValue,
  getDataValueBatch,
  getDataValueSerialNumber,
});
</script>
