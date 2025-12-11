<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/scm/asset-acquisition/gridconfig"
      form-config="/scm/asset-acquisition/formconfig"
      grid-read="/scm/asset-acquisition/gets"
      form-read="/scm/asset-acquisition/get"
      grid-mode="grid"
      grid-delete="/scm/asset-acquisition/delete"
      form-keep-label
      form-insert="/scm/asset-acquisition/save"
      form-update="/scm/asset-acquisition/save"
      :grid-fields="['Enable', 'Status']"
      :form-fields="[
        'Dimension',
        'TransferFrom',
        'JournalTypeID',
        'PostingProfileID',
      ]"
      :form-tabs-new="['General', 'Transfer Item', 'Register Asset']"
      :form-tabs-edit="['General', 'Transfer Item', 'Register Asset']"
      :form-tabs-view="['General', 'Transfer Item', 'Register Asset']"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      form-hide-submit
      :stay-on-form-after-save="true"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @alterGridConfig="onAlterGridConfig"
      :grid-custom-filter="data.customFilter"
      @gridResetCustomFilter="resetGridHeaderFilter"
    >
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilter"
          v-model="data.customFilter"
          hide-filter-text
          @initNewItem="initNewItemFilter"
          @preChange="changeFilter"
          @change="refreshGrid"
        >
          <template #filter_1="{ item }">
            <s-input
              class="w-[200px]"
              keep-label
              label="Transfer Name"
              v-model="item.TransferName"
            />
          </template>
        </grid-header-filter>
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_tab_Transfer_Item="{ item }">
        <s-grid
          v-model="item.ItemTranfers"
          ref="LineControlTranfer"
          class="w-full grid-line-items"
          hide-search
          hide-sort
          hide-refresh-button
          hide-select
          hide-detail
          auto-commit-line
          no-confirm-delete
          :editor="true"
          :hide-new-button="data.statusDisabled.includes(data.record.Status)"
          :hide-delete-button="data.statusDisabled.includes(data.record.Status)"
          :hide-detail="data.statusDisabled.includes(data.record.Status)"
          :hide-action="data.statusDisabled.includes(data.record.Status)"
          :config="data.gridCfgTransfer"
          @new-data="newRecordTransfer"
          @delete-data="onDeleteTransfer"
        >
          <template #item_ItemID="{ item, idx }">
            <s-input-sku-item
              v-model="item.ItemVarian"
              :record="item"
              :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
                item.ItemID,
                item.SKU
              )}`"
              :read-only="data.statusDisabled.includes(data.record.Status)"
              @afterOnChange="onAfterOnChange"
            ></s-input-sku-item>
          </template>
          <template #item_UnitID="{ item }">
            <s-input
              ref="refUnitID"
              v-model="item.UnitID"
              :read-only="data.statusDisabled.includes(data.record.Status)"
              class="w-full"
            ></s-input>
          </template>
          <template #item_UnitCost="{ item }">
            <s-input
              ref="refUnitCost"
              v-model="item.UnitCost"
              :read-only="data.statusDisabled.includes(data.record.Status)"
              kind="number"
              class="w-full text-right"
            ></s-input>
          </template>
          <template #item_Qty="{ item }">
            <s-input
              ref="refQty"
              v-model="item.Qty"
              :disabled="data.statusDisabled.includes(data.record.Status)"
              :read-only="data.statusDisabled.includes(data.record.Status)"
              kind="number"
              class="w-full text-right"
            ></s-input>
          </template>
          <template #item_SourceType="{ item }">
            <s-input
              ref="refSourceType"
              v-model="item.SourceType"
              :disabled="data.statusDisabled.includes(data.record.Status)"
              :read-only="data.statusDisabled.includes(data.record.Status)"
              use-list
              :items="['Purchase Request', 'Work Order (Assembly)']"
              class="w-full"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  onChangeSourceType(v1, v2, item);
                }
              "
            ></s-input>
          </template>
          <template #item_SourceReffNo="{ item }">
            <s-input
              ref="refSourceReffNo"
              v-model="item.SourceReffNo"
              :disabled="data.statusDisabled.includes(data.record.Status)"
              :read-only="data.statusDisabled.includes(data.record.Status)"
              hide-label
              class="w-full"
            ></s-input>
          </template>
        </s-grid>
      </template>
      <template #form_tab_Register_Asset="{ item }">
        <s-grid
          v-model="item.AssetRegisters"
          ref="LineControlRegister"
          class="w-full grid-line-items"
          hide-search
          hide-sort
          hide-refresh-button
          :hide-new-button="true"
          :hide-delete-button="true"
          :hide-detail="true"
          :hide-action="true"
          hide-select
          auto-commit-line
          no-confirm-delete
          :config="data.gridCfgRegister"
          form-keep-label
        >
          <template #item_ItemID="{ item, idx }">
            <s-input
              ref="refItemID"
              v-model="item.ItemID"
              :disabled="true"
              :read-only="true"
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
              :read-only="true"
              use-list
              :lookup-url="`/tenant/itemspec/find`"
              lookup-key="_id"
              :lookup-labels="['SKU']"
              :lookup-searchs="['_id', 'SKU']"
              class="w-full"
            ></s-input>
          </template>
          <template #item_DoesFixedAssetNumberIsExist="{ item }">
            <s-input
              v-if="!data.statusDisabled.includes(data.record.Status)"
              ref="refDoesFixedAssetNumberIsExist"
              v-model="item.DoesFixedAssetNumberIsExist"
              :disabled="data.statusDisabled.includes(data.record.Status)"
              :read-only="data.statusDisabled.includes(data.record.Status)"
              kind="bool"
              class="w-full"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  onChangeAssetID(v1, v2, item);
                }
              "
            ></s-input>
          </template>
          <template #item_AssetGroup="{ item }">
            <s-input
              ref="refAssetGroup"
              v-model="item.AssetGroup"
              :disabled="data.statusDisabled.includes(data.record.Status)"
              :read-only="data.statusDisabled.includes(data.record.Status)"
              use-list
              :lookup-url="`/tenant/assetgroup/find`"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              class="w-full"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  onChangeAssetID(v1, v2, item);
                }
              "
            ></s-input>
          </template>
          <template #item_AssetID="{ item }">
            <s-input
              ref="refAssetID"
              v-model="item.AssetID"
              :disabled="data.statusDisabled.includes(data.record.Status)"
              use-list
              :lookup-url="
                item.DoesFixedAssetNumberIsExist
                  ? `/bagong/asset/gets-filter?GroupID=${
                      item.AssetGroup
                    }&HasAcquisitionDate=${false}`
                  : `/fico/fixedassetnumberlist/gets-filter?FixedAssetGrup=${
                      item.AssetGroup
                    }&IsUsed=${false}`
              "
              lookup-key="_id"
              :lookup-labels="['_id']"
              :lookup-searchs="['_id', 'Name']"
              class="w-full"
            ></s-input>
          </template>
        </s-grid>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          :readOnly="data.statusDisabled.includes(data.record.Status)"
        ></dimension-editor>
      </template>
      <template #form_input_TransferFrom="{ item }">
        <dimension-invent-jurnal
          id="dimension-invent-Transfer"
          v-model="item.TransferFrom"
          title-header="Transfer From"
          :hide-field="[
            'VariantID',
            'Size',
            'Grade',
            'BatchID',
            'SerialNumber',
            'SpecID',
            'InventDimID',
          ]"
          :disabled="data.statusDisabled.includes(data.record.Status)"
        ></dimension-invent-jurnal>
      </template>
      <template #form_input_JournalTypeID="{ item }">
        <s-input
          ref="refInput"
          label="Journal Type"
          v-model="item.JournalTypeID"
          class="w-full"
          :required="true"
          :disabled="data.statusDisabled.includes(data.record.Status)"
          use-list
          :lookup-url="`/scm/asset-acquisition/journal/type/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="
            (field, v1, v2, old, ctlRef) => {
              getJournalType(v1, item);
            }
          "
        ></s-input>
      </template>
      <template #form_input_PostingProfileID="{ item }">
        <s-input
          ref="refInput"
          label="Posting Profile"
          v-model="item.PostingProfileID"
          class="w-full"
          :required="true"
          :disabled="data.statusDisabled.includes(data.record.Status)"
          use-list
          :lookup-url="`/fico/postingprofile/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          v-if="['', 'DRAFT'].includes(item.Status)"
          icon="content-save"
          class="btn_primary"
          label="Save"
          :disabled="data.isProcess"
          @click="onSave(item)"
        />
        <form-buttons-trx
          :key="data.btnTrxId"
          :status="item.Status"
          :moduleid="`scm/new`"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="'Asset Acquisition'"
          :auto-post="false"
          :disabled="data.isProcess"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  DataList,
  util,
  loadGridConfig,
  loadFormConfig,
  createFormConfig,
  SGrid,
  SInput,
  SButton,
} from "suimjs";
import helper from "@/scripts/helper.js";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import SInputSkuItem from "./widget/SInputSkuItem.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";

layoutStore().name = "tenant";
const separatorID = "~~";

const listControl = ref(null);
const refInput = ref([]);
const LineControlTranfer = ref(null);
const LineControlRegister = ref(null);
const gridHeaderFilter = ref(null);

const axios = inject("axios");
const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];
const data = reactive({
  btnTrxId: util.uuid(),
  appMode: "grid",
  formMode: "edit",
  titleForm: "Asset Acquisition",
  isProcess: false,
  gridCfgTransfer: {},
  gridCfgRegister: {},
  valueTranfers: [],
  valueRegister: [],
  statusDisabled: [
    "WAITING",
    "READY",
    "POSTED",
    "SUBMITTED",
    "IN PROGRESS",
    "COMPLETED",
  ],
  record: {
    _id: "",
  },
  customFilter: null,
});

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Asset Acquisition`;
  record._id = "";
  record.Status = "";
  record.TrxDate = new Date();
  record.TransferDate = new Date();
  record.valueTranfers = [];
  record.valueRegister = [];
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Asset Acquisition | ${record._id}`;
  updateRecordsTransfer();
  openForm(record);
}

function openForm(record) {
  util.nextTickN(2, () => {
    data.btnTrxId = util.uuid();
    data.record = record;
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
    listControl.value.setFormFieldAttr(
      "_id",
      "hide",
      record._id ? false : true
    );
    if (data.statusDisabled.includes(record.Status)) {
      listControl.value.setFormMode("view");
    } else {
      listControl.value.setFormMode("edit");
    }
  });
}

function onAlterGridConfig(cfg) {
  util.nextTickN(2, () => {
    cfg.sortable = ["Created", "TrxDate", "_id"];
    cfg.setting.idField = "Created";
    cfg.setting.sortable = ["Created", "TrxDate", "_id"];
  });
}

function newRecordTransfer() {
  const record = {};
  record.LineNo = 0;
  record.ItemVarian = "";
  record.ItemID = "";
  record.SKU = "";
  record.Description = "";
  record.Qty = 0;
  record.UnitID = "";
  record.UnitCost = 0;
  record.SourceType = "";
  LineControlTranfer.value.setRecords([
    ...LineControlTranfer.value.getRecords(),
    record,
  ]);
  updateRecordsTransfer();
}

function onDeleteTransfer(record, index) {
  const newRecords = LineControlTranfer.value.getRecords().filter((dt, idx) => {
    return idx != index;
  });
  LineControlTranfer.value.setRecords(newRecords);
  updateRecordsTransfer();
}

function updateRecordsTransfer(record, index) {
  util.nextTickN(2, () => {
    const records = LineControlTranfer.value.getRecords();
    records.map((r) => {
      if (r.ItemID != "" && r.SKU != "") {
        r.ItemVarian = helper.ItemVarian(r.ItemID, r.SKU);
      } else {
        r.ItemVarian = "";
      }
      return r;
    });
    data.valueTranfers = records;
  });
}

function onAfterOnChange(item) {}
function onChangeItemTranfer(v1, v2, item) {
  if (typeof v1 != "string") {
    item.UoM = "";
    item.Text = "";
    item.SKU = "";
    item.UnitCost = 0;
  } else {
    axios.post("/tenant/item/get", [v1]).then(
      (r) => {
        item.UnitID = r.data.DefaultUnitID;
        item.UnitCost = r.data.CostUnit;
        item.Item = r.data;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function onChangeSKUTranfer(v1, v2, item) {
  if (typeof v1 != "string") {
    item.Text = "";
  } else {
    axios.post("/tenant/itemspec/gets-detail", [v1]).then(
      (r) => {
        item.Text = r.data.length == 0 ? "" : r.data[0].Description;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}

function onChangeSourceType(v1, v2, item) {
  item.SourceReffNo = "";
}

function onChangeAssetID(v1, v2, item) {
  item.AssetID = "";
}

function onSave(record, status = "DRAFT", doSubmit) {
  let payload = record;
  if (LineControlTranfer.value && LineControlTranfer.value.getRecords()) {
    let tf = LineControlTranfer.value.getRecords();
    payload.ItemTranfers = tf;
  }
  if (LineControlRegister.value && LineControlRegister.value.getRecords()) {
    let reg = LineControlRegister.value.getRecords();
    payload.AssetRegisters = reg;
  }

  const param = JSON.parse(JSON.stringify(payload));
  param.Status = status;
  data.isProcess = true;
  listControl.value.setFormLoading(true);
  axios
    .post("/scm/asset-acquisition/save", param)
    .then(
      (r) => {
        if (r.data.Status == "DRAFT") {
          if (!r.data.Dimension) {
            r.data.Dimension = [];
          }
          listControl.value.setFormMode("edit");
          listControl.value.setFormRecord(r.data);
          listControl.value.setFormLoading(false);
          updateRecordsTransfer();
        } else {
          util.nextTickN(2, () => {
            doSubmit();
          });
        }
      },
      (e) => {
        util.showError(e);
      }
    )
    .finally(function () {
      data.isProcess = false;
      listControl.value.setFormLoading(false);
    });
}

function trxPostSubmit(record) {
  listControl.value.setControlMode("grid");
  listControl.value.refreshList();
  listControl.value.refreshForm();
}

function trxPreSubmit(status, action, doSubmit) {
  let validateJournalType = true;
  if (refInput.value) {
    validateJournalType = refInput.value.validate();
  }

  if (data.record.Status == "DRAFT" && validateJournalType) {
    if (!data.record.JournalTypeID || !data.record.PostingProfileID) {
      return util.showError(
        "field JournalTypeID or PostingProfileID is required"
      );
    }
    if (data.record.ItemTranfers.length == 0) {
      return util.showError("Item Tranfers is required");
    }
    onSave(data.record, "SUBMITTED", doSubmit);
  } else {
    util.nextTickN(2, () => {
      doSubmit();
    });
  }
}

function getJournalType(_id, item) {
  item.PostingProfileID = "";
  axios
    .post("/scm/asset-acquisition/journal/type/find?_id=" + _id, {
      sort: ["-_id"],
    })
    .then(
      (r) => {
        if (r.data.length > 0) {
          item.PostingProfileID = r.data[0].PostingProfileID;
        }
      },
      (e) => util.showError(e)
    );
}

watch(
  () => data.valueTranfers,
  (nv) => {
    util.nextTickN(2, () => {
      const group = nv.reduce((result, currentObject) => {
        const key = `${currentObject.ItemID}_${currentObject.SKU}`;
        if (
          currentObject.ItemID &&
          currentObject.SKU &&
          currentObject.Qty > 0
        ) {
          // Create an array for the key if it doesn't exist
          result[key] = result[key] || [];
          // Push the current object to the array
          result[key].push(currentObject);
        }
        return result;
      }, {});

      const register = [];
      for (let key in group) {
        const ast = LineControlRegister.value
          .getRecords()
          .filter(function (val) {
            return `${val.ItemID}_${val.SKU}` == key;
          });
        let idxAsset = 0;
        let keyAsset = "";
        for (let idx = 0; idx < group[key].length; idx++) {
          if (keyAsset != key) {
            idxAsset = 0;
            keyAsset = key;
          }

          for (let asset = 0; asset < group[key][idx].Qty; asset++) {
            const reg = {
              ...group[key][idx],
              ...{
                _id: util.uuid(),
                DoesFixedAssetNumberIsExist: ast[idxAsset]
                  ? ast[idxAsset].DoesFixedAssetNumberIsExist
                  : false,
                AssetGroup: ast[idxAsset] ? ast[idxAsset].AssetGroup : "",
                AssetID: ast[idxAsset] ? ast[idxAsset].AssetID : "",
              },
            };
            register.push(reg);
            idxAsset++;
          }
        }
      }
      data.valueRegister = register;
      LineControlRegister.value.setRecords(register);
    });
  },
  { deep: true }
);

function resetGridHeaderFilter() {
  gridHeaderFilter.value.reset();
}
function initNewItemFilter(item) {
  item.TransferName = "";
}
function changeFilter(item, filters) {
  if (item.TransferName.length > 0) {
    filters.push({
      Op: "$contains",
      Field: "TransferName",
      Value: [item.TransferName],
    });
  }
}
function refreshGrid() {
  listControl.value.refreshGrid();
}

onMounted(() => {
  loadGridConfig(axios, `/scm/asset-acquisition/item/transfer/gridconfig`).then(
    (r) => {
      let addColms = [];
      const hideColms = ["InventJournalLine", "Item", "LineNo"];
      let sortColm = [
        "LineNo",
        "ItemID",
        "SKU",
        "Text",
        "SourceType",
        "SourceReffNo",
        "UnitID",
        "UnitCost",
        "Qty",
      ];
      const colms = [
        {
          field: "ItemID",
          label: "Item",
          kind: "text",
        },
        {
          field: "UnitID",
          label: "UnitID",
          kind: "text",
        },
        {
          field: "UnitCost",
          label: "Unit Cost",
          kind: "number",
        },
        {
          field: "Qty",
          label: "Qty",
          kind: "number",
        },
      ];
      for (let index = 0; index < colms.length; index++) {
        addColms.push({
          field: colms[index].field,
          kind: colms[index].kind,
          label: colms[index].label,
          readType: "show",
          labelField: "",
          readOnly: data.statusDisabled.includes(data.record.Status),
          input: {
            field: colms[index].field,
            label: colms[index].label,
            hint: "",
            hide: false,
            placeHolder: colms[index].label,
            kind: colms[index].kind,
            readOnly: data.statusDisabled.includes(data.record.Status),
          },
        });
      }
      const _fields = [...r.fields, ...addColms].filter((o) => {
        if (["LineNo", "Qty", "UnitCost"].includes(o.field)) {
          o.width = "150px";
        } else if (["Description"].includes(o.field)) {
          o.width = "400px";
        } else {
          o.width = "300px";
        }
        o.idx = sortColm.indexOf(o.field);
        return !hideColms.includes(o.field);
      });
      data.gridCfgTransfer = {
        ...r,
        fields: _fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
    },
    (e) => util.showError(e)
  );
  loadGridConfig(
    axios,
    `/scm/asset-acquisition/asset/register/gridconfig`
  ).then(
    (r) => {
      const colms = [];
      let sortColms = [
        "ItemID",
        "SKU",
        "AssetGroup",
        "DoesFixedAssetNumberIsExist",
        "AssetID",
      ];
      let addColms = [];
      for (let index = 0; index < colms.length; index++) {
        addColms.push({
          field: colms[index].field,
          kind: "text",
          label: colms[index].label,
          readType: "show",
          labelField: "",
          input: {
            field: colms[index].field,
            label: colms[index].label,
            hint: "",
            hide: false,
            placeHolder: colms[index].label,
            kind: colms[index].kind,
          },
        });
      }
      const _fields = [...r.fields, ...addColms].filter((o) => {
        o.idx = sortColms.indexOf(o.field);
        return o;
      });
      data.gridCfgRegister = {
        ...r,
        fields: _fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
    },
    (e) => util.showError(e)
  );
});
</script>
