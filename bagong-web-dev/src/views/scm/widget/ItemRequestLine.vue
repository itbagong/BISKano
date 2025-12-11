<template>
  <div class="flex flex-col gap-2">
    <data-list
      v-show="!data.loading"
      :key="data.generalRecord._id"
      class="IRLine grid-line-items"
      ref="listControl"
      no-gap
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-no-confirm-delete
      gridAutoCommitLine
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      :form-hide-cancel="
        !['SUBMITTED', 'POSTED', 'READY', 'REJECTED'].includes(
          data.generalRecord.Status
        )
      "
      :form-hide-submit="true"
      :grid-hide-control="
        ['SUBMITTED', 'POSTED', 'READY', 'REJECTED'].includes(
          data.generalRecord.Status
        )
      "
      :grid-hide-detail="
        !['SUBMITTED', 'POSTED', 'READY', 'REJECTED'].includes(
          data.generalRecord.Status
        )
      "
      :grid-hide-delete="
        ['SUBMITTED', 'POSTED', 'READY', 'REJECTED'].includes(
          data.generalRecord.Status
        )
      "
      :grid-editor="['', 'DRAFT', 'READY'].includes(data.generalRecord.Status)"
      form-config="/scm/item/request/detail/formconfig"
      :form-fields="[
        'DetailLines',
        'Dimension',
        'QtyRemaining',
        'Remarks',
        'Complete',
        'SKU',
      ]"
      :grid-fields="
        data.generalRecord.Status === 'READY'
          ? [
              'ItemID',
              'QtyRequested',
              'QtyAvailable',
              'UoM',
              'Remarks',
              'Complete',
            ]
          : ['ItemID', 'QtyRequested', 'QtyAvailable', 'UoM', 'Remarks']
      "
      @grid-row-add="newRecord"
      @form-field-change="onFormFieldChanged"
      @form-edit-data="openForm"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="onGridRowFieldChanged"
      @grid-refreshed="gridRefreshed"
      @grid-row-save="onGridRowSave"
      @post-save="onFormPostSave"
      @alter-form-config="onAlterFormConfig"
      @form-loaded="formLoadedItem"
      form-focus
    >
      <template #grid_header_buttons_1="{ item }">
        <s-button
          v-if="
            !['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              generalRecord.Status
            )
          "
          class="btn_primary"
          label="Get Available Stock"
          @click="emit('getAvailStock', item)"
        />
      </template>
      <template #grid_ItemID="{ item, idx }">
        <s-input-sku-item
          v-model="item.ItemVarian"
          :record="item"
          :lookup-url="`/tenant/item/gets-detail?_id=${item.ItemVarian}`"
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            )
          "
        ></s-input-sku-item>
      </template>
      <template #grid_UoM="{ item }">
        <s-input
          ref="refUom"
          v-model="item.UoM"
          :key="item.UoM"
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            )
          "
          use-list
          :lookup-url="`/tenant/unit/gets-filter?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
        ></s-input>
      </template>
      <template #grid_QtyRequested="{ item }">
        <s-input
          v-if="
            !data.generalRecord.WOReff &&
            ['', 'DRAFT'].includes(data.generalRecord.Status)
          "
          v-model="item.QtyRequested"
          kind="number"
          class="w-full"
        ></s-input>
        <div style="text-align: right" v-else>
          {{ helper.formatNumberWithDot(item.QtyRequested) }}
        </div>
      </template>
      <template #grid_QtyAvailable="{ item }">
        <div
          :class="
            item.QtyRequested <= item.QtyAvailable
              ? 'text-green-700 font-semibold text-right'
              : 'text-red-700 font-semibold text-right'
          "
        >
          {{ helper.formatNumberWithDot(item.QtyAvailable) }}
        </div>
      </template>
      <template #grid_Complete="{ item }">
        <s-input
          v-model="item.Complete"
          :read-only="
            !['', 'DRAFT'].includes(item.Status) || !data.status.Postingers
          "
          kind="checkbox"
          class="w-full"
        ></s-input>
      </template>
      <template #grid_Remarks="{ item }">
        <s-input
          v-model="item.Remarks"
          :read-only="
            !['', 'DRAFT', 'READY'].includes(data.generalRecord.Status)
          "
          kind="text"
          class="w-full"
        ></s-input>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          v-if="
            ['READY'].includes(data.generalRecord.Status) &&
            data.status.Postingers
          "
          class="btn_primary"
          label="Save"
          icon="content-save"
          @click="onSave(item)"
        />
        <!-- sementara hide dulu -->
        <!-- <s-button
          v-if="
            ['SUBMITTED'].includes(data.generalRecord.Status)
          "
          :icon="cancelIcon" 
          class="btn_warning back_btn"
          label="Back"
          @click="onCancelForm"
        /> -->
      </template>
      <template #form_input_QtyRemaining="{ item }">
        <s-input
          label="Qty Remaining"
          v-model="item.QtyRemaining"
          disabled
          kind="number"
          :class="
            item.QtyRemaining > 0
              ? 'text-red-700 font-semibold'
              : 'text-black font-semibold'
          "
        ></s-input>
      </template>
      <template #form_input_Complete="{ item }">
        <s-input
          label="Complete"
          v-model="item.Complete"
          :read-only="
            !['READY'].includes(generalRecord.Status) ||
            item.QtyFulfilled === item.QtyRequested
          "
          kind="bool"
          class="w-full"
        ></s-input>
      </template>
      <template #form_input_Remarks="{ item }">
        <s-input
          label="Remarks"
          v-model="item.Remarks"
          :read-only="
            !['', 'DRAFT'].includes(item.Status) || !data.status.Postingers
          "
          kind="text"
          class="w-full"
        ></s-input>
      </template>
      <template #form_input_DetailLines="r">
        <s-grid
          v-model="r.item.DetailLines"
          ref="controlDetailLines"
          class="w-full grid-fulfillment grid-line-items"
          :editor="['READY'].includes(data.generalRecord.Status)"
          hide-search
          hide-sort
          hide-header
          :hide-new-button="
            !['READY'].includes(data.generalRecord.Status) ||
            !data.status.Postingers
          "
          :hide-delete-button="
            !['READY'].includes(data.generalRecord.Status) ||
            !data.status.Postingers
          "
          :hide-detail="true"
          :hide-action="
            !['READY'].includes(data.generalRecord.Status) ||
            !data.status.Postingers
          "
          :config="data.gridCfgDetailLines"
          hide-refresh-button
          hide-select
          auto-commit-line
          no-confirm-delete
          form-keep-label
          @row-field-changed="
            (name, v1, v2, record, old) =>
              onFullfilmentRowFieldChanged(item, name, v1, record)
          "
          @new-data="newRecordLineFulfillment"
          @delete-data="deleteRecordLineFulfillment"
          @controlModeChanged="onControlModeChanged"
        >
          <template #item_FulfillmentType="{ item }">
            <s-input
              v-model="item.FulfillmentType"
              :key="item.FulfillmentType"
              :disabled="
                !['READY'].includes(data.generalRecord.Status) ||
                !data.status.Postingers
              "
              use-list
              class="w-full"
              :items="data.fulfillmentType"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  item.WarehouseID = '';
                  item.QtyAvailable = 0;
                  item.InventDimFrom = {};
                }
              "
            ></s-input>
          </template>
          <template #item_QtyFulfilled="{ item }">
            <s-input
              v-if="
                ['READY'].includes(data.generalRecord.Status) &&
                data.status.Postingers
              "
              ref="refUom"
              v-model="item.QtyFulfilled"
              kind="number"
              class="w-full"
            ></s-input>
            <div style="text-align: right" v-else>
              {{ helper.formatNumberWithDot(item.QtyFulfilled) }}
            </div>
          </template>
          <template #item_UoM="{ item }">
            <s-input
              ref="refUom"
              v-model="item.UoM"
              :key="item.UoM"
              disabled
              use-list
              :lookup-url="`/tenant/unit/gets-filter?ItemID=${data.lineRecord.ItemID}`"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              class="w-full"
            ></s-input>
          </template>
          <template #item_WarehouseID="{ item }">
            <s-input
              hide-label
              label="From warehouse"
              v-model="item.WarehouseID"
              :key="item.WarehouseID"
              class="w-full"
              :disabled="
                !['READY'].includes(data.generalRecord.Status) ||
                !data.status.Postingers ||
                ['', null, 'Purchase Request', 'Assembly'].includes(
                  item.FulfillmentType
                )
              "
              :use-list="
                !['', null, 'Purchase Request', 'Assembly'].includes(
                  item.FulfillmentType
                )
              "
              :lookup-url="`/scm/item/balance/get-available-warehouse?ItemID=${data.lineRecord.ItemID}&SKU=${data.lineRecord.SKU}&ItemRequestID=${data.generalRecord._id}&FulfillmentType=${item.FulfillmentType}`"
              lookup-key="_id"
              :lookup-labels="['Text']"
              :lookup-searchs="['_id', 'Text']"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  onGetsAvailableWarehouse(v1, item);
                }
              "
            ></s-input>
          </template>
          <template #item_QtyAvailable="{ item }">
            <div
              :class="
                item.QtyFulfilled <= item.QtyAvailable
                  ? 'text-green-700 font-semibold text-right'
                  : 'text-red-700 font-semibold text-right'
              "
            >
              {{ helper.formatNumberWithDot(item.QtyAvailable) }}
            </div>
          </template>
        </s-grid>
      </template>
      <template #form_input_SKU="{ item }">
        <!-- /tenant/itemspec/find|SKU|SKU -->
        <s-input
          label="SKU"
          ref="refUom"
          v-model="item.SKU"
          :key="item.SKU"
          :disabled="true"
          use-list
          :lookup-url="`/tenant/itemspec/find?_id=${item.SKU}`"
          lookup-key="_id"
          :lookup-labels="['SKU', 'SpecVariantID']"
          :lookup-searchs="['_id', 'SKU', 'SpecVariantID']"
          class="w-full"
        ></s-input>
      </template>
    </data-list>
    <div v-show="data.loading" class="loading">
      loading data from server ...
    </div>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, DataList, SGrid, SButton } from "suimjs";
import helper from "@/scripts/helper.js";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import SInputSkuItem from "./SInputSkuItem.vue";

const axios = inject("axios");
const listControl = ref(null);
const lineConfig = ref(null);
const controlDetailLines = ref(null);
const refItemID = ref(null);
const separatorID = "~~";
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  generalRecord: { type: Object, default: () => {} },
  approval: {
    type: Object,
    default: {
      Approvers: false,
      Postingers: false,
      Submitters: false,
    },
  },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
  getAvailStock: null,
});

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Item Request Fulfillment",
  generalRecord: props.generalRecord,
  records: props.modelValue.map((dt) => {
    dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
    dt.suimRecordChange = false;
    return dt;
  }),
  fulfillmentType: [
    {
      text: "Item Transfer",
      value: "Item Transfer",
    },
    {
      text: "Purchase Request",
      value: "Purchase Request",
    },
    {
      text: "Movement Out",
      value: "Movement Out",
    },
    {
      text: "Assembly",
      value: "Assembly",
    },
  ],
  fulfillmentTypeNoMO: [
    {
      text: "Item Transfer",
      value: "Item Transfer",
    },
    {
      text: "Purchase Request",
      value: "Purchase Request",
    },
    {
      text: "Assembly",
      value: "Assembly",
    },
  ],
  listAvailableWarehouse: [],
  gridCfgDetailLines: {},
  lineRecord: {},
  loading: false,
  status: {
    Approvers: false,
    Postingers: false,
    Submitters: false,
  },
});
function newRecord() {
  const record = {};
  record.ItemRequestID = data.generalRecord._id;
  record.ItemVarian = "";
  record.ItemID = "";
  record.SKU = "";
  record.QtyAvailable = 0;
  record.QtyRequested = 0;
  if (data.generalRecord.InventDimTo?.WarehouseID) {
    record.WarehouseID = data.generalRecord.InventDimTo.WarehouseID;
  }
  data.records.push(record);
  listControl.value.setGridRecords(data.records);
  updateItems();
}
function openForm(record) {
  data.lineRecord = record;
  util.nextTickN(2, () => {
    listControl.value.refreshForm();
  });
  // onGetsAvailableWarehouse();
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
  if (data.records) {
    listControl.value.setGridRecords(data.records);
    updateItems();
  }
}
function newRecordLineFulfillment() {
  if (
    data.lineRecord.QtyRequested ===
    controlDetailLines.value.getRecords().length
  ) {
    return util.showError("Record can not bigger than Quantity Requested");
  }
  const record = {};
  record.FulfillmentType = "";
  record.QtyFulfilled = 0;
  record.WarehouseID = "";
  record.QtyAvailable = 0;
  record.InventDimFrom = {};
  record.Remarks = "";
  record.UoM = data.lineRecord.UoM;
  controlDetailLines.value.setRecords([
    ...controlDetailLines.value.getRecords(),
    record,
  ]);
  const newRecords = controlDetailLines.value.getRecords();
  updateRecordsDetailLines(newRecords);
}
function deleteRecordLineFulfillment(record, index) {
  const newRecords = record.items.filter((dt, idx) => {
    return idx != index;
  });
  updateRecordsDetailLines(newRecords);
}
function onFormFieldChanged(name, v1, v2, record, old) {
  // console.log(name, v1, v2, record, old)
}
function onFullfilmentRowFieldChanged(item, name, v1, record) {
  // console.log(item, name, v1, record)
}
function updateRecordsDetailLines(records) {
  util.nextTickN(2, () => {
    data.records.map(function (r) {
      if (r._id == data.lineRecord._id) {
        r.DetailLines = records;
      }
      return r;
    });
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
  listControl.value.setGridRecords(data.records);
}
async function formLoadedItem(record) {
  if (!["POSTED"].includes(data.generalRecord.Status)) {
    for (const d of record.DetailLines) {
      await onGetsAvailableWarehouse(d.WarehouseID, d);
    }
  }
}
// function setDataGrids() {
//   data.records = props.modelValue.map((dt) => {
//     dt.suimRecordChange = false;
//     return dt;
//   });
//   // if (data.records !== null) {
//   //   util.nextTickN(10, () => {
//   //     listControl.value.setGridRecords(data.records);
//   //   })
//   // }
// }

function onChangeItem(v1, v2, item) {
  if (typeof v1 != "string") {
    item.UoM = "";
  } else {
    axios.post("/tenant/item/get", [v1]).then(
      (r) => {
        item.UoM = r.data.DefaultUnitID;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function onAlterFormConfig(cfg) {
  cfg.setting.title = data.titleForm;
  cfg.setting.showTitle = true;
}
function alterGridConfig(cfg) {
  console.log(cfg);
}
async function onGetsAvailableWarehouse(_id, item) {
  await axios
    .post(
      `/scm/item/balance/get-available-warehouse?ItemID=${data.lineRecord.ItemID}&SKU=${data.lineRecord.SKU}&ItemRequestID=${data.generalRecord._id}&FulfillmentType=${item.FulfillmentType}`
    )
    .then(
      (r) => {
        data.listAvailableWarehouse = r.data;
        const wh = r.data.find(function (v) {
          return v._id == _id;
        });
        if (wh) {
          item.InventDimFrom = wh.InventDim;
          item.QtyAvailable = wh.Qty;
        }
      },
      (e) => {
        return util.showError(e);
      }
    );
}

// function onCancelForm() {
//   listControl.value.cancelForm();
//   util.nextTickN(2, () => {
//     getLines(props.generalRecord);
//   });
// }
function onSave(record) {
  if (record.QtyFulfilled > record.QtyRequested) {
    return util.showError(
      "The Qty Fulfilled is greater than requested, Please check your Qty Fulfilled!"
    );
  }
  // if (
  //   record.DetailLines.find(
  //     (o) => o.FulfillmentType == "Item Transfer" && o.WarehouseID === ""
  //   )
  // ) {
  //   return util.showError("From warehouse is required");
  // }
  axios.post("/scm/item/request/detail/save", record).then(
    (r) => {
      listControl.value.cancelForm();
      util.nextTickN(2, () => {
        getLines(props.generalRecord);
      });
    },
    (e) => util.showError(e)
  );
}
function getLines(record) {
  data.loading = true;
  axios
    .post("/scm/item/request/detail/gets?ItemRequestID=" + record._id, {
      Skip: 0,
      Take: 0,
      Sort: ["_id"],
    })
    .then(
      (r) => {
        data.records = r.data.data.map((dt) => {
          dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
          return dt;
        });
        if (listControl.value) {
          listControl.value.setGridRecords(data.records);
        }

        util.nextTickN(2, () => {
          updateItems();
        });
      },
      (e) => util.showError(e)
    )
    .finally(() => (data.loading = false));
}
function fetchApproveSource() {
  axios
    .post("/fico/postingprofile/get-approval-by-source-user", {
      JournalType: "Item Request",
      JournalID: props.generalRecord._id,
    })
    .then(
      (r) => {
        data.status = r.data;
      },
      (e) => util.showError(e)
    )
    .finally(() => {});
}
onMounted(() => {
  loadGridConfig(axios, "/scm/item/request/detail/gridconfig").then(
    (r) => {
      let cfg = r;
      let fields = [
        "ItemID",
        "Description",
        "ItemType",
        "UoM",
        "QtyRequested",
        "QtyAvailable",
        "Remarks",
      ];
      if (!["", "DRAFT"].includes(data.generalRecord.Status)) {
        fields = [...fields, "QtyFulfilled", "Complete"];
      }
      cfg.fields = cfg.fields.map((el) => {
        if (["QtyRequested", "QtyAvailable", "UoM"].includes(el.field)) {
          el.width = "130px";
        }
        if (["Remarks"].includes(el.field)) {
          el.width = "400px";
        }

        return {
          ...el,
          input: {
            ...el.input,
            readOnly: false,
          },
        };
      });
      if (!["", "DRAFT"].includes(data.generalRecord.Status)) {
        cfg.fields = cfg.fields.map((el) => {
          return {
            ...el,
            input: {
              ...el.input,
              readOnly: !["WarehouseID", "QtyAvailable"].includes(el.field),
            },
          };
        });
      }
      cfg.fields = cfg.fields.filter((el) => fields.includes(el.field));
      util.nextTickN(2, () => {
        if (listControl.value) {
          listControl.value.setGridConfig(cfg);
        }

        getLines(props.generalRecord);
      });
    },
    (e) => util.showError(e)
  );
  loadGridConfig(axios, `/scm/item/request/detail/line/gridconfig`).then(
    (r) => {
      const newFields = [
        {
          field: "QtyAvailable",
          kind: "number",
          label: "Qty Available",
          readType: "show",
          disable: true,
          input: {
            field: "QtyAvailable",
            label: "Qty Available",
            hint: "",
            hide: false,
            placeHolder: "QtyAvailable",
            kind: "number",
            disable: true,
            required: false,
            multiple: false,
          },
        },
      ];
      const fields = [...r.fields, ...newFields];
      r.fields = fields;
      data.gridCfgDetailLines = r;
    },
    (e) => util.showError(e)
  );
  fetchApproveSource();
});
defineExpose({
  getLines,
});
watch(
  () => data.lineRecord.DetailLines,
  (nv) => {
    const sum = nv.reduce((a, b) => {
      return a + b.QtyFulfilled;
    }, 0);
    data.lineRecord.QtyFulfilled = sum;
    data.lineRecord.QtyRemaining =
      data.lineRecord.QtyRequested - data.lineRecord.QtyFulfilled;
    if (!data.lineRecord.Complete) {
      data.lineRecord.Complete =
        data.lineRecord.QtyFulfilled == data.lineRecord.QtyRequested;
    }
  },
  { deep: true }
);
</script>
