<template>
  <div class="w-full">
    <s-card
      :title="cardTitle"
      class="w-full bg-white suim_datalist card"
      :no-gap="false"
      :hide-title="false"
    >
      <s-grid
        v-if="data.appMode == 'grid'"
        v-model="data.value"
        ref="listControl"
        class="w-full"
        hide-search
        hide-sort
        delete-url="/tenant/item/delete"
        sortField="Created"
        sortDirection="desc"
        :hide-new-button="true"
        :hide-delete-button="!profile.canDelete"
        hide-refresh-button
        :hide-detail="!profile.canUpdate"
        :hide-action="false"
        auto-commit-line
        :config="data.gridCfg"
        form-keep-label
        @selectData="onSelectData"
        @rowDeleted="onRowDeleted"
      >
        <template #header_search="{ config }">
          <s-input
            ref="refItemID"
            v-model="data.search.ItemIDs"
            label="Item"
            class="w-full"
            multiple
            use-list
            :lookup-url="`/tenant/item/find`"
            lookup-key="_id"
            :lookup-labels="['_id', 'Name']"
            :lookup-searchs="['_id', 'Name', 'OtherName']"
            @change="onFilterRefresh"
          ></s-input>
          <s-input
            kind="date"
            label="Date From"
            class="min-w-[120px]"
            v-model="data.search.DateFrom"
            @change="onFilterRefresh"
          ></s-input>
          <s-input
            kind="date"
            label="Date To"
            class="min-w-[120px]"
            v-model="data.search.DateTo"
            @change="onFilterRefresh"
          ></s-input>
        </template>
        <template #header_buttons="{ config }">
          <s-button
            icon="refresh"
            class="btn_primary refresh_btn"
            @click="refreshData"
          />

          <s-button
            v-if="profile.canCreate"
            icon="plus"
            class="btn_primary new_btn"
            @click="newRecord"
          />
        </template>
        <template #item_ItemGroupID="{ item }">
          {{ item.ItemGroupID }}
        </template>
        <template #item_LedgerAccountIDStock="{ item }">
          {{ item.LedgerAccountIDStock }}
        </template>
        <template #item_DefaultUnitID="{ item }">
          {{ item.DefaultUnitID }}
        </template>
        <template #paging>
          <div
            v-if="pageCount > 1"
            class="flex gap-2 justify-center pagination"
          >
            <mdicon
              name="arrow-left"
              class="cursor-pointer"
              :class="{
                'opacity-25': data.paging.currentPage == 1,
              }"
              @click="changePage(data.paging.currentPage - 1)"
            />
            <div class="pagination_info">
              Page {{ data.paging.currentPage }} of {{ pageCount }}
            </div>
            <mdicon
              name="arrow-right"
              class="cursor-pointer"
              :class="{ 'opacity-25': data.paging.currentPage == pageCount }"
              @click="changePage(data.paging.currentPage + 1)"
            />
          </div>
        </template>
      </s-grid>
      <s-form
        v-else
        ref="formCtlItem"
        v-model="data.records"
        :keep-label="true"
        :config="data.frmCfg"
        class="pt-2"
        :auto-focus="true"
        :hide-submit="true"
        :hide-cancel="false"
        :tabs="tabsList"
        :mode="data.formMode"
        @cancelForm="onCancelForm"
      >
        <template #tab_Specification="{ item }">
          <ItemSpecification
            ref="specificationConfig"
            typeSpec="item"
            :item="item"
          ></ItemSpecification>
        </template>
        <template #tab_Item_Balance="{ item }">
          <transaction-balance
            :key="data.keyItemBalance"
            v-model="data.filterBalance"
            url-gets="/scm/item/balance/gets-by-warehouse-and-section"
            url-config="/scm/item/balance/dimension/gridconfig"
            transaction-type="Inventory"
            :jurnal-id="item._id"
            :dim-list="[
              'WarehouseID',
              'SectionID',
              'VariantID',
              'SKU',
              'Size',
              'Grade',
              'BatchID',
              'SerialNumber',
            ]"
          />
        </template>
        <template #tab_Item_Transaction="{ item }">
          <transaction-history
            ref="transactionHistory"
            transaction-type="Inventory"
            config-url="/scm/inventory/trx/dimension/gridconfig"
            url="/scm/inventory/trx/gets-by-balance"
            hide-dim-finance
            :jurnal-id="item._id"
            :param="item"
          />
        </template>
        <template #input__id="{ item }">
          <s-input
            ref="refId"
            label="ID"
            v-model="item._id"
            class="w-full"
            :keepErrorSection="true"
            :required="true"
            :disabled="data.formMode == 'new' ? false : true"
          ></s-input>
        </template>
        <template #input_ItemGroupID="{ item }">
          <s-input
            ref="refItemGroupID"
            label="Item Group"
            v-model="item.ItemGroupID"
            class="w-full"
            :keepErrorSection="true"
            use-list
            :lookup-url="`/tenant/itemgroup/find`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="
              (field, v1, v2, old, ctlRef) => {
                onChangeItemGroup(v1, v2, item);
              }
            "
          ></s-input>
        </template>
        <template #input_PhysicalDimension="{ item }">
          <div class="flex gap-2">
            <s-input
              v-for="(value, key, index) in item.PhysicalDimension"
              v-model="item.PhysicalDimension[key]"
              :label="data.dim.find((v) => v.key == key).label"
              kind="checkbox"
              class="w-full"
            ></s-input>
          </div>
        </template>
        <template #input_FinanceDimension="{ item }">
          <div class="flex gap-2">
            <s-input
              v-for="(value, key, index) in item.FinanceDimension"
              v-model="item.FinanceDimension[key]"
              :label="data.dim.find((v) => v.key == key).label"
              kind="checkbox"
              class="w-full"
            ></s-input>
          </div>
        </template>
        <template #buttons_1="{ item }">
          <s-button
            class="btn_primary submit_btn"
            :icon="`content-save`"
            :label="'Save'"
            :disabled="data.isProcess"
            @click="onSubmit"
          />
        </template>
      </s-form>
    </s-card>
  </div>
</template>

<script setup>
import { reactive, ref, inject, computed, onMounted, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
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
import ItemSpecification from "./widget/ItemSpecification.vue";
import TransactionBalance from "@/components/common/TransactionBalance.vue";
import TransactionHistory from "@/components/common/TransactionHistory.vue";

layoutStore().name = "tenant";
const featureID = "ItemMaster";
const profile = authStore().getRBAC(featureID);

const listControl = ref(null);
const formCtlItem = ref(null);
const specificationConfig = ref(null);
const refId = ref(null);
const transactionHistory = ref(null);

const axios = inject("axios");
const route = useRoute();
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "",
  tabspecKey: "spec",
  value: [],
  recordsCount: 0,
  records: {},
  gridCfg: {},
  frmCfg: {},
  stayOnForm: true,
  isProcess: false,
  search: {
    ItemIDs: [],
    DateFrom: new Date(),
    DateTo: new Date(),
    Check: {
      IsEnabledSpecVariant: false,
      IsEnabledSpecSize: false,
      IsEnabledSpecGrade: false,
      IsEnabledItemBatch: false,
      IsEnabledItemSerial: false,
      IsEnabledLocationWarehouse: false,
      IsEnabledLocationAisle: false,
      IsEnabledLocationSection: false,
      IsEnabledLocationBox: false,
    },
  },
  dim: [
    {
      key: "IsEnabledSpecVariant",
      label: "VariantID",
    },
    {
      key: "IsEnabledSpecSize",
      label: "Size",
    },
    {
      key: "IsEnabledSpecGrade",
      label: "Grade",
    },
    {
      key: "IsEnabledItemBatch",
      label: "Batch",
    },
    {
      key: "IsEnabledItemSerial",
      label: "Serial",
    },
    {
      key: "IsEnabledLocationWarehouse",
      label: "WarehouseID",
    },
    {
      key: "IsEnabledLocationAisle",
      label: "AisleID",
    },
    {
      key: "IsEnabledLocationSection",
      label: "SectionID",
    },
    {
      key: "IsEnabledLocationBox",
      label: "BoxID",
    },
  ],
  paging: {
    skip: 0,
    pageSize: 25,
    currentPage: 1,
  },
  listSpec: [],
  gridCfgSpec: {},
  keyItemBalance: util.uuid(),
});

const cardTitle = computed(() => {
  if (data.appMode == "grid") return "Item";
  const formMode = data.appMode;
  return formMode == "grid"
    ? `Item`
    : data.formMode == `edit`
    ? `Item - ${data.records.Name}`
    : `Item`;
});

const tabsList = computed({
  get() {
    if (data.formMode == `edit` && data.records._id) {
      return ["General", "Specification", "Item Balance", "Item Transaction"];
    }
    return ["General"];
  },
});

function newRecord() {
  data.records = {
    _id: "",
    Name: "",
    Enable: true,
    IsActive: true,
    PhysicalDimension: {
      IsEnabledSpecVariant: false,
      IsEnabledSpecSize: false,
      IsEnabledSpecGrade: false,
      IsEnabledItemBatch: false,
      IsEnabledItemSerial: false,
      IsEnabledLocationWarehouse: false,
      IsEnabledLocationAisle: false,
      IsEnabledLocationSection: false,
      IsEnabledLocationBox: false,
    },
    FinanceDimension: {
      IsEnabledSpecVariant: false,
      IsEnabledSpecSize: false,
      IsEnabledSpecGrade: false,
      IsEnabledItemBatch: false,
      IsEnabledItemSerial: false,
      IsEnabledLocationWarehouse: false,
      IsEnabledLocationAisle: false,
      IsEnabledLocationSection: false,
      IsEnabledLocationBox: false,
    },
  };
  data.titleForm = `Create New Item`;
  data.appMode = `form`;
  data.formMode = `new`;
  util.nextTickN(2, () => {
    formCtlItem.value.setFieldAttr("_id", "hide", true);
  });
}

function onCancelForm(record) {
  data.appMode = "grid";
  util.nextTickN(2, () => {
    refreshData();
  });
}

function onSubmit(record) {
  // let validate = true;
  // if (refId.value) {
  //   validate = refId.value.validate();
  // }
  // if (validate) {

  // } else {
  //   formCtlItem.value.setLoading(false);
  // }

  let validateGeneral = formCtlItem.value.validate();
  let validateSpec = true;
  if (data.records._id) {
    const isSpecGradeID = data.records.PhysicalDimension.IsEnabledSpecGrade;
    const isSpecSizeID = data.records.PhysicalDimension.IsEnabledSpecSize;
    const isSpecVariantID = data.records.PhysicalDimension.IsEnabledSpecVariant;
    const lineSpec = specificationConfig.value.getDataValue().filter((s) => {
      return (
        s.SKU == "" ||
        ([null, ""].includes(s.SpecGradeID) && isSpecGradeID) ||
        ([null, ""].includes(s.SpecVariantID) && isSpecVariantID) ||
        ([null, ""].includes(s.SpecSizeID) && isSpecSizeID)
      );
    });
    if (lineSpec.length > 0) {
      validateSpec = false;
      util.showError("Specifications The line is empty");
    }
  }

  if (validateGeneral && validateSpec) {
    data.isProcess = true;
    formCtlItem.value.setLoading(true);
    axios.post("/tenant/item/save", data.records).then(
      (r) => {
        data.records = r.data;
        data.formMode = `edit`;
        onPostSave(r.data);
      },
      (e) => {
        formCtlItem.value.setLoading(false);
        data.isProcess = false;
        util.showError(e);
      }
    );
  }
}

function onPostSave(record) {
  if (specificationConfig.value && specificationConfig.value.getDataValue()) {
    let dv = specificationConfig.value.getDataValue();
    dv.map(function (sp) {
      if (!record.PhysicalDimension.IsEnabledSpecVariant) {
        sp.SpecVariantID = "";
      }
      if (!record.PhysicalDimension.IsEnabledSpecSize) {
        sp.SpecSizeID = "";
      }
      if (!record.PhysicalDimension.IsEnabledSpecGrade) {
        sp.SpecGradeID = "";
      }
      return sp;
    });
    if (dv.length > 0) {
      // return onCancelForm();
      axios
        .post("/tenant/itemspec/save-multiple", dv)
        .then(
          (r) => {
            // onCancelForm();
          },
          (e) => {
            util.showError(e);
          }
        )
        .finally(function () {
          util.nextTickN(2, () => {
            data.isProcess = false;
            formCtlItem.value.setLoading(false);
            transactionHistory.value.refreshData();
          });
        });
    } else {
      util.nextTickN(2, () => {
        data.isProcess = false;
        formCtlItem.value.setLoading(false);
      });
    }
  } else {
    // onCancelForm();
    data.isProcess = false;
    formCtlItem.value.setLoading(false);
  }
}
function changePage(page) {
  data.paging.currentPage = page;
  refreshData();
}
function refreshData() {
  const payload = {
    ...data.search,
    Skip: (data.paging.currentPage - 1) * data.paging.pageSize,
    Take: data.paging.pageSize,
    Sort: ["-Created"],
  };
  listControl.value.setLoading(true);
  const _fields = data.gridCfg.fields.filter((o) => {
    const isDim = [
      "IsEnabledItemBatch",
      "IsEnabledItemSerial",
      "IsEnabledLocationAisle",
      "IsEnabledLocationBox",
      "IsEnabledLocationSection",
      "IsEnabledLocationWarehouse",
      "IsEnabledSpecGrade",
      "IsEnabledSpecSize",
      "IsEnabledSpecVariant",
    ].includes(o.field);

    if (isDim) {
      o.readType = data.search.Check[o.field] ? "show" : "hide";
    }
    return o;
  });
  data.gridCfg = { ...data.gridCfg, fields: _fields };
  axios.post("/scm/item/gets", payload).then(
    (r) => {
      util.nextTickN(2, () => {
        data.recordsCount = r.data.total;
        data.value = r.data.data.map(function (val) {
          val = { ...val, ...val.InventDim };
          return val;
        });
        listControl.value.setLoading(false);
      });
    },
    (e) => {
      util.showError(e);
    }
  );
}

function onSelectData(val, index) {
  data.appMode = `form`;
  axios.post("/tenant/item/get", [val._id]).then(
    (r) => {
      data.records = r.data;
      data.formMode = `edit`;
      util.nextTickN(2, () => {
        formCtlItem.value.setFieldAttr("_id", "hide", false);
        data.keyItemBalance = util.uuid();
        transactionHistory.value.refreshData();
      });
    },
    (e) => {
      util.showError(e);
    }
  );
}

function onRowDeleted(val) {
  refreshData();
}

function onFilterRefresh(val) {
  util.nextTickN(2, () => {
    refreshData();
  });
}
function onChangeItemGroup(v1, v2, item) {
  if (typeof v1 == "string") {
    axios.post("/tenant/itemgroup/get", [v1]).then(
      (r) => {
        item.ItemType = r.data.ItemType;
        item.LedgerAccountIDStock = r.data.LedgerAccountIDStock;
        item.DefaultUnitID = r.data.DefaultUnitID;
        item.CostUnitCalcMethod = r.data.CostUnitCalcMethod;
        item.CostUnit = r.data.CostUnit;
        item.PhysicalDimension = r.data.PhysicalDimension;
        item.FinanceDimension = r.data.FinanceDimension;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}
const pageCount = computed({
  get() {
    const count = Math.ceil(data.recordsCount / data.paging.pageSize);
    return count;
  },
});

onMounted(() => {
  listControl.value.setLoading(true);
  loadGridConfig(axios, `/tenant/item/gridconfig`).then(
    (r) => {
      const hideColm = ["OtherName"];
      const Qty = [];
      let customColl = [];
      for (let index = 0; index < Qty.length; index++) {
        customColl.push({
          field: Qty[index],
          kind: "number",
          label: Qty[index],
          readType: "show",
          labelField: "",
          input: {
            field: Qty[index],
            label: Qty[index],
            hint: "",
            hide: false,
            placeHolder: Qty[index],
            kind: "number",
            disable: false,
            required: false,
            multiple: false,
          },
        });
      }
      const dim = data.dim;
      let dimColl = [];
      for (let index = 0; index < dim.length; index++) {
        dimColl.push({
          field: dim[index].key,
          kind: "text",
          label: dim[index].label,
          readType: "show",
          labelField: "",
          input: {
            field: dim[index].key,
            label: dim[index].label,
            hint: "",
            hide: false,
            placeHolder: dim[index].label,
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
          },
        });
      }
      const _fields = [...r.fields, ...customColl].filter((o) => {
        return !hideColm.includes(o.field);
      });
      data.gridCfg = { ...r, fields: [..._fields, ...dimColl] };
      refreshData();
    },
    (e) => util.showError(e)
  );
  loadFormConfig(axios, "/tenant/item/formconfig").then(
    (r) => {
      data.frmCfg = r;
    },
    (e) => util.showError(e)
  );
});
</script>
<style>
.row_action {
  justify-content: center;
  align-items: center;
  gap: 2px;
}
</style>
