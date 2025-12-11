<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      form-hide-submit
      :grid-hide-new="true"
      grid-config="/scm/inventory/receive/gridconfig"
      form-config="/scm/inventory/receive/formconfig"
      grid-read="/scm/inventory/receive/gets"
      form-read="/scm/inventory/receive/get"
      grid-mode="grid"
      grid-delete="/scm/inventory/receive/delete"
      form-keep-label
      grid-hide-delete
      grid-hide-select
      form-insert="/scm/inventory/receive/save"
      form-update="/scm/inventory/receive/save"
      form-default-mode="view"
      :form-fields="['Dimension', 'CompanyID', 'WarehouseID', 'SectionID']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :form-tabs-new="['General', 'Line']"
      :form-tabs-edit="['General', 'Line']"
      :form-tabs-view="['General', 'Line']"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
    >
      <template #form_tab_Line="{ item }">
        <s-grid
          v-model="item.Lines"
          ref="LineControl"
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
          :config="data.gridLineCfg"
          form-keep-label
        >
          <template #item_Item="{ item }">
            {{ item.Item.Name }}
          </template>
        </s-grid>
      </template>
      <template #form_buttons_2="{ item }">
        <div class="flex gap-[2px] ml-2">
          <form-buttons-trx
            :status="item.Status"
            :moduleid="`scm`"
            :journal-id="item._id"
            :posting-profile-id="item.PostingProfileID"
            journal-type-id="Inventory Receive"
            @preSubmit="trxPreSubmit"
            @postSubmit="trxPostSubmit"
          />
        </div>
      </template>
      <template #form_input_CompanyID="{ item }">
        <s-input
          v-model="item.CompanyID"
          useList
          label="Company"
          class="w-full"
          lookup-url="/tenant/company/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-search="['_id', 'Name']"
          :disabled="true"
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <div>
          <dimension-invent-jurnal
            v-model="item.Dimension"
            title-header="Inventory Dimension"
            :readOnly="true"
            :hideField="['BatchID', 'SerialNumber', 'InventDimID']"
          ></dimension-invent-jurnal>
        </div>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import { layoutStore } from "@/stores/layout.js";
import {
  DataList,
  util,
  SForm,
  SInput,
  SButton,
  SGrid,
  loadGridConfig,
  loadFormConfig,
  createFormConfig,
} from "suimjs";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import ItemRequestLine from "./widget/ItemRequestLine.vue";
import SMenuButton from "./widget/SMenuButton.vue";

layoutStore().name = "tenant";
const listControl = ref(null);
const lineConfig = ref(null);
const axios = inject("axios");

const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Invent Receive Journals",
  gridLineCfg: {},
  fromCfgDim: {},
  record: {
    _id: "",
    Status: "",
  },
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Invent Receive Journals`;
  record._id = "";
  record.Status = "";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Invent Receive Journals | ${record._id}`;
  data.record = record;
  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
  });
}

function trxPostSubmit(record) {
  listControl.value.refreshForm();
  listControl.value.setControlMode("grid");
  listControl.value.refreshList();
}

function trxPreSubmit(status, action, doSubmit) {
  doSubmit();
}

function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Invent Receive Journals";
  }
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
onMounted(() => {
  loadGridConfig(axios, `/scm/inventory/receive/line/gridconfig`).then(
    (r) => {
      let hideColm = ["InventJournalLine", "LineNo"];
      const Line = ["LineNo", "SKU", "UnitID", "Text", "Qty"];
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
        "Qty",
      ];
      let InventJournalLine = [];
      for (let index = 0; index < Line.length; index++) {
        InventJournalLine.push({
          field: Line[index],
          kind: ["Qty"].includes(Line[index]) ? "number" : "text",
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
});
</script>
