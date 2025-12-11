<template>
  <div class="w-full">
    <s-card
      :title="cardTitle"
      class="w-full bg-white suim_datalist card"
      hide-footer
      :no-gap="false"
      :hide-title="false"
    >
      <s-grid
        v-if="data.appMode == 'grid'"
        v-model="data.value"
        ref="listControl"
        class="w-full grid-line-items"
        hide-search
        hide-sort
        :hide-new-button="true"
        :hide-delete-button="true"
        hide-refresh-button
        :hide-detail="true"
        :hide-action="true"
        auto-commit-line
        no-confirm-delete
        :config="data.gridCfg"
        form-keep-label
      >
        <template #header_search="{ config }">
          <s-input
            v-model="data.search.SourceJournalID"
            kind="text"
            label="Search Source Journal"
            class="w-full"
            hideLabel
          ></s-input>
          <s-input
            v-model="data.search.CompanyID"
            useList
            label="Company"
            class="w-[500px]"
            lookup-url="/tenant/company/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-search="['_id', 'Name']"
            hideLabel
          ></s-input>
          <s-input
            v-model="data.search.WarehouseID"
            useList
            label="Warehouse"
            class="w-[500px]"
            lookup-url="/tenant/warehouse/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-search="['_id', 'Name']"
            hideLabel
          ></s-input>
          <s-input
            v-model="data.search.SourceType"
            useList
            label="Source Type"
            class="min-w-[200px]"
            :items="['INVENTORY', 'PURCHASE', 'WORKORDER']"
            hideLabel
          ></s-input>
          <s-input
            v-model="data.search.Status"
            useList
            label="Status"
            class="min-w-[200px]"
            :items="['Planned', 'Reserved', 'Confirmed']"
            hideLabel
          ></s-input>
        </template>
        <template #header_buttons="{ config }">
          <s-button
            icon="refresh"
            class="btn_primary refresh_btn"
            @click="GridRefreshed"
          />
        </template>
        <template #header_buttons_2="{ config }">
          <s-button
            label="Process"
            class="btn_primary refresh_btn"
            @click="onSelectData"
          />
        </template>
        <template #item_Item="{ item }">
          {{ item.Item.Name }}
        </template>
        <template #item_SKU="{ item }">
          <s-input
            ref="refSKU"
            v-model="item.SKU"
            :disabled="true"
            use-list
            :lookup-url="`/tenant/itemspec/find`"
            lookup-key="_id"
            :lookup-labels="['SKU']"
            :lookup-searchs="['_id', 'SKU']"
            class="w-full"
          ></s-input>
        </template>
        <template #item_WarehouseID="{ item }">
          <s-input
            ref="refWarehouseID"
            v-model="item.WarehouseID"
            :disabled="true"
            use-list
            :lookup-url="`/tenant/warehouse/find`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
          ></s-input>
        </template>
      </s-grid>
      <s-form
        v-else
        ref="formCtlInventTrx"
        v-model="data.records"
        :keep-label="true"
        :config="data.frmCfg"
        class="pt-2"
        :auto-focus="true"
        :hide-submit="true"
        :hide-cancel="true"
        :tabs="['General', 'Line']"
        :mode="data.formModeInventTrx"
        @cancelForm="onCancelForm"
      >
        <template #tab_Line="{ item }">
          <div>
            <s-grid
              v-show="data.appLine == ''"
              v-model="data.valueLine"
              ref="LineControl"
              class="w-full grid-line-items"
              hide-search
              hide-sort
              :hide-new-button="true"
              :hide-delete-button="true"
              hide-refresh-button
              :hide-detail="true"
              :hide-action="false"
              hide-select
              auto-commit-line
              no-confirm-delete
              :config="data.gridLineCfg"
              form-keep-label
            >
              <template #item_Item="{ item }"> {{ item.Item.Name }} </template>
              <template #item_SKU="{ item }">
                <s-input
                  ref="refSKU"
                  v-model="item.SKU"
                  :disabled="true"
                  use-list
                  :lookup-url="`/tenant/itemspec/find`"
                  lookup-key="_id"
                  :lookup-labels="['SKU']"
                  :lookup-searchs="['_id', 'SKU']"
                  class="w-full"
                ></s-input>
              </template>
              <template #item_Qty="{ item }">
                <s-input
                  ref="refQty"
                  kind="number"
                  v-model="item.Qty"
                  label="Qty"
                  hide-label
                  :disabled="false"
                  class="w-full"
                ></s-input>
              </template>
              <template #item_buttons_1="{ item }">
                <a
                  href="#"
                  @click="onSelectLineDIM(item, idx)"
                  class="edit_action"
                >
                  <mdicon
                    name="eye"
                    width="16"
                    alt="edit"
                    class="cursor-pointer hover:text-primary"
                  />
                </a>
                <a
                  href="#"
                  @click="onSelectLineBatchSN(item, idx)"
                  class="edit_action"
                >
                  <mdicon
                    name="format-list-numbered"
                    width="16"
                    alt="edit"
                    class="cursor-pointer hover:text-primary"
                  />
                </a>
              </template>
            </s-grid>
            <s-form
              v-show="data.appLine == 'InventDim'"
              ref="formCtlInventDim"
              v-model="data.valueInventDim"
              :keep-label="true"
              :config="data.fromCfgDim"
              class="pt-2"
              :auto-focus="true"
              :hide-submit="true"
              :hide-cancel="true"
              mode="view"
            >
              <template #input_Dimension="{ item }">
                <dimension-editor
                  v-model="item.Dimension"
                  :readOnly="true"
                ></dimension-editor>
              </template>
              <template #input_InventDim="{ item }">
                <div>
                  <dimension-invent-jurnal
                    v-model="item.InventDim"
                    title-header="Inventory Dimension"
                    :readOnly="true"
                    :hideField="['BatchID', 'SerialNumber', 'InventDimID']"
                  ></dimension-invent-jurnal>
                </div>
              </template>
            </s-form>
            <s-grid
              v-show="data.appLine == 'BatchSN'"
              v-model="data.valueBatchSN"
              ref="BatchSNControl"
              class="w-full grid-line-items"
              hide-search
              hide-sort
              :hide-new-button="true"
              :hide-delete-button="true"
              hide-refresh-button
              :hide-detail="true"
              :hide-action="true"
              hide-select
              auto-commit-line
              no-confirm-delete
              :config="data.gridCfgBatchSN"
              form-keep-label
            >
            </s-grid>
          </div>
        </template>
        <template #input__id="{ item }">
          <s-input
            v-model="item._id"
            useList
            label="ID"
            class="w-full"
            lookup-url="/scm/inventory/receive/find"
            lookup-key="_id"
            :lookup-labels="['_id', 'Name']"
            :lookup-search="['_id', 'Name']"
            @change="
              (field, v1, v2, old, ctlRef) => {
                onChangeInventTrx(v1, v2, item);
              }
            "
          ></s-input>
        </template>
        <template #input_TrxType="{ item }">
          <s-input
            v-model="item.TrxType"
            useList
            label="Trx Type"
            class="w-full"
            :items="['Inventory Receive', 'Inventory Issuance']"
            :disabled="data.formModeInventTrx == 'edit' ? false : true"
          ></s-input>
        </template>
        <template #input_CompanyID="{ item }">
          <s-input
            v-model="item.CompanyID"
            useList
            label="Company"
            class="w-full"
            lookup-url="/tenant/company/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-search="['_id', 'Name']"
            :disabled="data.formModeInventTrx == 'edit' ? false : true"
          ></s-input>
        </template>
        <template #input_Dimension="{ item }">
          <div>
            <dimension-editor
              v-model="item.Dimension"
              :readOnly="data.formModeInventTrx == 'edit' ? false : true"
              :hideField="[
                'VariantID',
                'Size',
                'Grade',
                'SpecID',
                'BatchID',
                'SerialNumber',
                'InventDimID',
              ]"
            ></dimension-editor>
          </div>
        </template>
        <template #buttons="{ item }">
          <s-button
            :icon="`rewind`"
            class="btn_warning back_btn"
            :label="'Back'"
            @click="onBackForm"
          />
        </template>
        <template #buttons_1="{ item }">
          <s-button
            :icon="`content-save`"
            class="btn_primary submit_btn"
            :label="'Save'"
            @click="onSave"
          />
          <!-- <form-buttons-trx
            :status="`DRAFT`"
            :moduleid="`scm`"
            :journal-id="data.journalID"
            :posting-profile-id="data.records.PostingProfileID"
            journal-type-id="Inventory Receive"
            @preSubmit="trxPreSubmit"
            @postSubmit="trxPostSubmit"
          /> -->
        </template>
      </s-form>
    </s-card>
  </div>
</template>

<script setup>
import { reactive, ref, inject, computed, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { useRoute } from "vue-router";
import {
  loadGridConfig,
  loadFormConfig,
  createFormConfig,
  util,
  SInput,
  SButton,
  SForm,
  SGrid,
  SCard,
} from "suimjs";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";

layoutStore().name = "tenant";

const listControl = ref(null);
const formCtlInventTrx = ref(null);
const LineControl = ref(null);
const axios = inject("axios");
const route = useRoute();
const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];
const data = reactive({
  value: [],
  valueLine: [],
  valueInventDim: {},
  valueBatchSN: [],
  appMode: "grid",
  appLine: "",
  formMode: "edit",
  formInsert: "/scm/inventory/trx/receipt/save",
  formUpdate: "/scm/inventory/trx/receipt/save",
  journalID: "",
  formModeInventTrx: "edit",
  gridCfg: {},
  frmCfg: {},
  gridLineCfg: {},
  fromCfgDim: {},
  gridCfgBatchSN: {},
  records: {},
  record: null,
  search: {
    CompanyID: "DEMO00",
    WarehouseID: "W-MLG001",
    SourceType: "",
    SourceJournalID: "",
    Status: "Planned",
  },
});

const cardTitle = computed(() => {
  if (data.appMode == "grid") return "Inventory Transaction";
  const formMode = data.appMode;
  return formMode == "grid"
    ? `Inventory Transaction`
    : "Inventory Transaction - ";
});

function onCancelForm(record) {
  data.appMode = "grid";
  util.nextTickN(2, () => {
    GridRefreshed();
  });
}
function onBackForm(record) {
  const tab = document.querySelector(
    ".tab_container > div.tab_selected"
  ).textContent;
  if (data.appLine == "" || tab == "General") {
    data.appMode = "grid";
    util.nextTickN(2, () => {
      GridRefreshed();
    });
  } else if (data.appLine == "InventDim") {
    util.nextTickN(2, () => {
      data.appLine = "";
    });
  } else if (data.appLine == "BatchSN") {
    util.nextTickN(2, () => {
      data.appLine = "";
    });
  }
}

function setFormRequired(isRequired) {
  formCtlInventTrx.value.setFieldAttr("CompanyID", "required", isRequired);
  formCtlInventTrx.value.setFieldAttr("Name", "required", isRequired);
  formCtlInventTrx.value.setFieldAttr("WarehouseID", "required", isRequired);
  formCtlInventTrx.value.setFieldAttr("SectionID", "required", isRequired);
  formCtlInventTrx.value.setFieldAttr("TrxDate", "required", isRequired);
  formCtlInventTrx.value.setFieldAttr(
    "PostingProfileID",
    "required",
    isRequired
  );
}

function onSave() {
  setFormRequired(true);
  const valid = formCtlInventTrx.value.validate();
  const line = data.valueLine;
  data.records.Lines = line;
  data.records.status = "DRAFT";
  let url = "/scm/inventory/receive/save";
  if (data.records._id) {
    url = "/scm/inventory/receive/update-lines";
  }
  if (valid) {
    axios.post(url, data.records).then(
      (r) => {
        onCancelForm();
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function onChangeInventTrx(v1, v2, item) {
  if (typeof v1 != "string") {
    data.records = {
      Dimension: {},
    };
    data.formModeInventTrx = "edit";
  } else {
    axios.post("/scm/inventory/receive/get", [v1]).then(
      (r) => {
        util.nextTickN(2, () => {
          if (r.data.Dimension == null) {
            r.data.Dimension = {};
          }
          data.records = r.data;
          data.formModeInventTrx = "view";
        });
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function trxPostSubmit(record) {}

function trxPreSubmit(status, action, doSubmit) {
  setFormRequired(true);
  const valid = formCtlInventTrx.value.validate();
  const line = data.valueLine;
  data.records.Lines = line;
  if (valid) {
    axios.post("/scm/inventory/receive/save", data.records).then(
      (r) => {
        data.records = r.data;
        data.journalID = r.data._id;
        if (status == "DRAFT") {
          util.nextTickN(2, () => {
            doSubmit();
          });
        }
      },
      (e) => {}
    );
  }
}

function onSelectData(record, index) {
  const Invent = listControl.value.getRecords();
  const Selected = Invent.filter((el) => el.isSelected === true);
  Selected.map(function (val) {
    val.LineNo = val.SourceLineNo;
    val.SourceLine = val.SourceLineNo;
    val.UnitID = val.TrxUnitID;
    val.Qty = 0;
    val.BatchSerials = [];
    val.Dimension = [];
    return val;
  });
  if (Selected.length > 0) {
    data.appMode = "from";
    data.appLine = "";
    data.records = {
      _id: "",
      Name: "",
      TrxDate: new Date(),
    };
    data.valueLine = Selected;
    data.formModeInventTrx = "edit";
  } else {
    return util.showError("Please choose inventory transaction");
  }
}

function onSelectLineDIM(record, index) {
  data.appLine = "InventDim";
  util.nextTickN(2, () => {
    data.valueInventDim = {
      InventDim: record.InventDim,
      Dimension: [],
    };
  });
}

function onSelectLineBatchSN(record, index) {
  data.appLine = "BatchSN";
}

function GridRefreshed() {
  listControl.value.setLoading(true);
  let payload = { ...data.search, ...{ Statuses: [data.search.Status] } };
  axios
    .post("/scm/inventory/trx/gets-filter", payload)
    .then(
      (r) => {
        r.data.data.map(function (i) {
          i.WarehouseID = i.InventDim.WarehouseID;
          i.AisleID = i.InventDim.AisleID;
          i.SectionID = i.InventDim.SectionID;
          i.BoxID = i.InventDim.BoxID;
          i.VariantID = i.InventDim.VariantID;
          i.Size = i.InventDim.Size;
          i.Grade = i.InventDim.Grade;
        });
        listControl.value.setRecords(r.data.data);
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      util.nextTickN(2, () => {
        listControl.value.setLoading(false);
      });
    });
}
function genCfgInventDim() {
  const cfg = createFormConfig("", true);
  cfg.addSection("General2", false).addRow(
    {
      field: "InventDim",
      kind: "text",
      label: "InventDim",
    },
    {
      field: "Dimension",
      kind: "text",
      label: "Dimension",
    }
  );
  data.fromCfgDim = cfg.generateConfig();
}

function genCfgBatchSN() {
  const BatchSN = ["BatchID", "SerialNumber", "Qty"];
  let CfgBatchSN = [];
  for (let index = 0; index < BatchSN.length; index++) {
    CfgBatchSN.push({
      field: BatchSN[index],
      kind: "Text",
      label: BatchSN[index],
      readType: "show",
      input: {
        field: BatchSN[index],
        label: BatchSN[index],
        hint: "",
        hide: false,
        placeHolder: BatchSN[index],
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    });
  }
  data.gridCfgBatchSN = {
    setting: { idField: "", keywordFields: ["_id", "Name"], sortable: ["_id"] },
    fields: CfgBatchSN,
  };
}

onMounted(() => {
  loadGridConfig(axios, `/scm/inventory/trx/receipt/gridconfig`).then(
    (r) => {
      let hideColm = ["InventDim"];
      const Dim = ["WarehouseID"];
      let InventoryDimension = [];
      for (let index = 0; index < Dim.length; index++) {
        InventoryDimension.push({
          field: Dim[index],
          kind: "Text",
          label: Dim[index],
          readType: "show",
          input: {
            field: Dim[index],
            label: Dim[index],
            hint: "",
            hide: false,
            placeHolder: Dim[index],
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
          },
        });
      }
      const _fields = [...r.fields, ...InventoryDimension].filter((o) => {
        return !hideColm.includes(o.field);
      });
      data.gridCfg = { ...r, fields: _fields };
      GridRefreshed();
    },
    (e) => util.showError(e)
  );
  loadFormConfig(axios, "/scm/inventory/receive/formconfig").then(
    (r) => {
      data.frmCfg = r;
    },
    (e) => util.showError(e)
  );
  loadGridConfig(axios, `/scm/inventory/receive/line/gridconfig`).then(
    (r) => {
      let hideColm = ["InventJournalLine", "LineNo"];
      const Line = [
        "LineNo",
        "SKU",
        "UnitID",
        "Text",
        "OriginalQty",
        "SettledQty",
        "TrxQty",
        "Qty",
      ];
      let colmLine = [
        "SourceType",
        "SourceJournalID",
        "SourceTrxType",
        "SourceLine",
        "LineNo",
        "Item",
        "SKU",
        "UnitID",
        "Text",
        "OriginalQty",
        "SettledQty",
        "TrxQty",
        "Qty",
      ];
      let InventJournalLine = [];
      for (let index = 0; index < Line.length; index++) {
        InventJournalLine.push({
          field: Line[index],
          kind: ["OriginalQty", "SettledQty", "TrxQty", "Qty"].includes(
            Line[index]
          )
            ? "number"
            : "text",
          label: Line[index],
          readType: "show",
          labelField: "",
          input: {
            field: Line[index],
            label: Line[index],
            hint: "",
            hide: false,
            placeHolder: Line[index],
            kind: ["OriginalQty", "SettledQty", "TrxQty", "Qty"].includes(
              Line[index]
            )
              ? "number"
              : "text",
          },
        });
      }
      const _fields = [...r.fields, ...InventJournalLine].filter((o) => {
        if (["OriginalQty", "SettledQty", "TrxQty", "Qty"].includes(o.field)) {
          o.width = "300px";
        } else {
          o.width = "400px";
        }
        o.idx = colmLine.indexOf(o.field);
        return !hideColm.includes(o.field);
      });
      data.gridLineCfg = {
        ...r,
        fields: _fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
    },
    (e) => util.showError(e)
  );
  genCfgInventDim();
  genCfgBatchSN();
});
</script>
<style>
.row_action {
  justify-content: center;
  align-items: center;
  gap: 2px;
}
</style>
