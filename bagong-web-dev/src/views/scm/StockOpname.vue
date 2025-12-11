<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="data.titleForm"
      :form-hide-submit="true"
      grid-config="/scm/stock-opname/gridconfig"
      form-config="/scm/stock-opname/formconfig"
      grid-read="/scm/stock-opname/gets"
      form-read="/scm/stock-opname/get"
      grid-mode="grid"
      grid-delete="/scm/stock-opname/delete"
      form-keep-label
      form-insert="/scm/stock-opname/save"
      form-update="/scm/stock-opname/save"
      :grid-fields="['Enable', 'WarehouseID', 'Status']"
      :form-fields="[
        '_id',
        'JournalTypeID',
        'PostingProfileID',
        'Dimension',
        'InventDim',
      ]"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :form-tabs-new="['General', 'Line']"
      :form-tabs-edit="['General', 'Line']"
      :form-tabs-view="['General', 'Line']"
      :formInitialTab="data.formInitialTab"
      form-default-mode="edit"
      :grid-custom-filter="customFilter"
      @alter-grid-config="alterGridConfig"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      :stay-on-form-after-save="data.stayOnForm"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <!-- @preSave="onPreSave" -->
      <template #grid_header_search="{ config }">
        <s-input
          ref="refName"
          v-model="data.search.Name"
          lookup-key="_id"
          label="Text"
          class="w-full"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Adj Date From"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Adj Date To"
          v-model="data.search.DateTo"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refStatus"
          v-model="data.search.Status"
          lookup-key="_id"
          label="Status"
          class="w-[300px]"
          use-list
          :items="['DRAFT', 'SUBMITTED', 'READY', 'POSTED', 'REJECTED']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refSite"
          v-model="data.search.Site"
          lookup-key="_id"
          label="Site"
          class="w-[450px]"
          use-list
          :disabled="defaultList?.length === 1"
          :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
          :lookup-labels="['Label']"
          :lookup-searchs="['_id', 'Label']"
          :lookup-payload-builder="
            defaultList?.length > 0
              ? (...args) =>
                  helper.payloadBuilderDimension(
                    defaultList,
                    data.search.Site,
                    false,
                    ...args
                  )
              : undefined
          "
          @change="refreshData"
        ></s-input>
      </template>
      <template #form_loader>
        <loader />
      </template>
      <template #form_tab_Line="{ item }">
        <div v-if="data.loadingLine" class="loading">
          loading data from server ...
        </div>
        <StockOpnameLine
          v-else
          :key="util.uuid()"
          ref="lineConfig"
          v-model="item.Lines"
          :general-record="item"
          :isLoading="data.loadingLine"
        ></StockOpnameLine>
      </template>
      <template #grid_WarehouseID="{ item }">
        <s-input
          hide-label
          label="Warehouse ID"
          v-model="item.InventDim.WarehouseID"
          class="w-full"
          read-only
          use-list
          :lookup-url="`/tenant/warehouse/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_input_JournalTypeID="{ item }">
        <s-input
          :key="data.keyJournalType"
          ref="refInput"
          label="Journal Type"
          v-model="item.JournalTypeID"
          class="w-full"
          :required="true"
          :disabled="true"
          :keepErrorSection="true"
          use-list
          :lookup-url="`/scm/inventory/journal/type/find?TransactionType=Stock%20Opname`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="
            (field, v1, v2, old, ctlRef) => {
              getJournalType(v1, 'change');
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
          :disabled="
            data.lockPostingProfile ||
            ['SUBMITTED', 'READY', 'POSTED'].includes(item.Status)
          "
          :keepErrorSection="true"
          use-list
          :lookup-url="`/fico/postingprofile/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          :key="data.compDimensionKey"
          v-model="item.Dimension"
          sectionTitle="Financial Dimension"
          :default-list="profile.Dimension"
          :readOnly="
            data.lockDimension ||
            ['SUBMITTED', 'READY', 'POSTED'].includes(item.Status)
          "
        ></dimension-editor>
      </template>
      <template #form_input_InventDim="{ item }">
        <dimension-invent-jurnal
          ref="RefDimensionInventory"
          :key="data.compInventDimensionKey"
          v-model="item.InventDim"
          :defaultList="profile.Dimension"
          :site="
            item.Dimension &&
            item.Dimension.find((_dim) => _dim.Key === 'Site') &&
            item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] != ''
              ? item.Dimension.find((_dim) => _dim.Key === 'Site')['Value']
              : undefined
          "
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
            data.lockInventDimension ||
            ['SUBMITTED', 'READY', 'POSTED'].includes(item.Status)
          "
          :mandatory="['WarehouseID']"
          @onFieldChanged="(v1, v2, _item) => {}"
        ></dimension-invent-jurnal>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          v-if="
            !['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(item.Status)
          "
          :icon="`content-save`"
          class="btn_primary submit_btn"
          label="Save"
          @click="onPreSave(item)"
        />

        <FormButtonsTrx
          :key="data.btnTrx"
          :posting-profile-id="item.PostingProfileID"
          :status="item.Status"
          :journalId="item._id"
          :autoPost="false"
          journal-type-id="Stock Opname"
          moduleid="scm"
          @pre-submit="preSubmit"
          @post-submit="postSubmit(item)"
        >
        </FormButtonsTrx>
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import {
  reactive,
  ref,
  inject,
  computed,
  watch,
  nextTick,
  onMounted,
} from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  createFormConfig,
  DataList,
  util,
  SInput,
  SButton,
  SCard,
  SForm,
} from "suimjs";
import { useRouter, useRoute } from "vue-router";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import StockOpnameLine from "./widget/StockOpnameLine.vue";
import Loader from "@/components/common/Loader.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import { authStore } from "@/stores/auth.js";
import helper from "@/scripts/helper.js";
import moment from "moment";
layoutStore().name = "tenant";

const featureID = "StockOpname";

// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'StockOpname'})

const profile = authStore().getRBAC(featureID);
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);

const listControl = ref(null);
const lineConfig = ref(null);
const RefDimensionInventory = ref(null);
const axios = inject("axios");
const route = useRoute();
const router = useRouter();

let currentTab = computed(() => {
  if (listControl.value == null) {
    return 0;
  } else {
    return listControl.value.getFormCurrentTab();
  }
});

let customFilter = computed(() => {
  const filters = [];
  if (data.search.Name !== null && data.search.Name !== "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "_id",
          Op: "$contains",
          Value: [data.search.Name],
        },
        {
          Field: "Name",
          Op: "$contains",
          Value: [data.search.Name],
        },
      ],
    });
  }
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "TrxDate",
      Op: "$gte",
      Value: moment(data.search.DateFrom).utc().format("YYYY-MM-DDT00:mm:00Z"),
    });
  }
  if (
    data.search.DateTo !== null &&
    data.search.DateTo !== "" &&
    data.search.DateTo !== "Invalid date"
  ) {
    filters.push({
      Field: "TrxDate",
      Op: "$lte",
      Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
    });
  }
  if (data.search.Status !== null && data.search.Status !== "") {
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: data.search.Status,
    });
  }

  if (
    data.search.Site !== undefined &&
    data.search.Site !== null &&
    data.search.Site !== ""
  ) {
    filters.push(
      {
        Field: "Dimension.Key",
        Op: "$eq",
        Value: "Site",
      },
      {
        Field: "Dimension.Value",
        Op: "$eq",
        Value: data.search.Site,
      }
    );
  }

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Stock Opname",
  formInitialTab: 0,
  record: {
    _id: "",
    TrxDate: new Date(),
    Status: "",
  },
  search: {
    Name: "",
    DateFrom: null,
    DateTo: null,
    Status: "",
    Site: "",
  },
  journalTypeData: {},
  stayOnForm: true,
  lockDimension: false,
  lockPostingProfile: false,
  lockInventDimension: false,
  keyJournalType: "",
  compDimensionKey: "0",
  compInventDimensionKey: "0",
  disabledFormButton: false,
  btnTrx: "",
  dimInventory: "",
  loadingLine: false,
});

function newRecord(record) {
  record._id = "";
  record.StockOpnameDate = new Date();
  record.InputDate = new Date();
  record.TrxDate = new Date();
  record.Status = "";
  data.formMode = "new";
  data.titleForm = "Create New Stock Opname";
  openForm(record, () => {
    getPostingProfile(record);
  });
}

function editRecord(record) {
  data.formMode = "edit";
  data.record = record;
  data.titleForm = `Edit Stock Opname | ${record._id}`;
  data.dimInventory = Object.values(record.InventDim).join("|");
  openForm(record, () => {
    if (["SUBMITTED", "READY", "POSTED"].includes(record.Status)) {
      setFormMode("view");
    }
  });
}

function openForm(record, cbOK) {
  // if (record._id) {
  //   lineConfig.value.getLoadLine();
  // }
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr(
      "_id",
      "hide",
      data.formMode == "new" ? true : false
    );
    const el = document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(1) > div > div > div:nth-child(1)"
    );
    if (record._id) {
      if (el) {
        el.style.display = "block";
      }
    } else {
      if (el) {
        el.style.display = "none";
      }
    }
    if (cbOK) {
      cbOK();
    }
  });
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
function alterGridConfig(cfg) {
  cfg.sortable = ["LastUpdate", "Created", "TrxDate", "_id"];
  cfg.setting.idField = "LastUpdate";
  cfg.setting.sortable = ["LastUpdate", "Created", "TrxDate", "_id"];
  cfg.fields.push({
    field: "WarehouseID",
    kind: "Text",
    label: "Warehouse",
    readType: "show",
    input: {
      field: "WarehouseID",
      label: "Warehouse",
      hint: "",
      hide: false,
      placeHolder: "Warehouse",
      kind: "text",
      disable: false,
      required: false,
      multiple: false,
    },
  });

  let sortColm = ["_id", "Name", "TrxDate", "WarehouseID", "Created", "Status"];
  cfg.fields.map((f) => {
    if (["Created", "CompanyID"].includes(f.field)) {
      f.readType = "hide";
    }
    f.idx = sortColm.indexOf(f.field);
    return f;
  });
  cfg.fields.sort((a, b) => (a.idx > b.idx ? 1 : -1));
}
function setLoadingForm(loading) {
  data.disabledFormButton = loading;
  listControl.value.setFormLoading(loading);
}
function getJournalType(_id, action) {
  axios
    .post("/scm/inventory/journal/type/find?_id=" + _id, { sort: ["-_id"] })
    .then(
      (r) => {
        data.journalTypeData = r.data[0];
        setDefaultRecordFromJournalType(r.data[0], action);
        setLockForm(r.data[0]);
      },
      (e) => util.showError(e)
    );
}
function setLockForm(JT) {
  data.lockDimension = JT.LockFinancialDimension;
  data.lockInventDimension = JT.LockInventoryDimension;
  data.lockPostingProfile = JT.LockPostingProfile;
}
function setDefaultRecordFromJournalType(JT, action) {
  if (["", "DRAFT"].includes(data.record.Status)) {
    if (action === "change") {
      data.record.InventDim = JT.InventoryDimension;
      data.record.Dimension = JT.Dimension;
      data.record.PostingProfileID = JT.PostingProfileID;
      nextTick(() => {
        data.compDimensionKey = util.uuid();
        data.compInventDimensionKey = util.uuid();
      });
    }
  }
}
function setRequiredAllField(required) {
  listControl.value.getFormAllField().forEach((e) => {
    if (!["PostingProfileID"].includes(e.field)) {
      listControl.value.setFormFieldAttr(e.field, "required", required);
    }
  });
}
function checkItemExists() {
  if (lineConfig.value && lineConfig.value.getDataValue()) {
    let dv = lineConfig.value.getDataValue();
    var Items = dv.map(function (item) {
      return `${item["ItemID"]}|${item["SKU"]}`;
    });

    return new Set(Items).size !== dv.length;
  }
  return false;
}
function preSubmit(status, action, doSubmit) {
  data.stayOnForm = data.record.Status == "DRAFT" ? true : false;
  if (status == "DRAFT") {
    // @change="
    //         (field, v1, v2, old, ctlRef) => {
    //           item.Gap = v1 - item.QtyInSystem;
    //           if (item.Gap > 0) {
    //             item.Remarks = 'OVER'
    //           } else if (item.Gap < 0) {
    //             item.Remarks = 'MINUS'
    //           } else if (item.Gap === 0) {
    //             item.Remarks = 'OK'
    //           }
    //         }
    //       "
    setRequiredAllField(true);
    util.nextTickN(2, () => {
      const valid = listControl.value.formValidate();
      if (valid) {
        setLoadingForm(true);
        data.record?.Lines?.map((o) => {
          o.Gap = parseFloat(o.QtyActual) - o.QtyInSystem;
          if (o.Gap > 0) {
            o.Remarks = "OVER";
          } else if (o.Gap < 0) {
            o.Remarks = "MINUS";
          } else if (o.Gap === 0) {
            o.Remarks = "OK";
          }
          o.QtyActual = o.QtyActual.toString();
          return o;
        });

        listControl.value.submitForm(
          data.record,
          () => {
            setLoadingForm(false);
            doSubmit();
          },
          () => {
            setLoadingForm(false);
          }
        );
      }
      setRequiredAllField(false);
    });
  } else {
    doSubmit();
  }
}
function postSubmit(record) {
  setLoadingForm(false);
  data.stayOnForm = false;
  data.appMode = "grid";
  listControl.value.refreshList();
  listControl.value.setControlMode("grid");
  // util.nextTickN(2, () => {
  //   axios.post("/scm/stock-opname/get", [data.record._id]).then((res) => {
  //     listControl.value.setFormRecord(res.data);
  //     data.formMode = ["SUBMITTED", "READY", "REJECTED", "POSTED"].includes(
  //       res.data.Status
  //     )
  //       ? "view"
  //       : "edit";
  //     setFormMode(data.formMode);
  //     data.record = res.data;
  //     listControl.value.refreshForm();
  //     data.btnTrx = util.uuid();
  //   });
  // });
}

function onFieldChanged(record) {
  let isInv = data.dimInventory != Object.values(record.InventDim).join("|");
  if (record.InventDim.WarehouseID && isInv) {
    data.loadingLine = true;
    axios.post("/scm/stock-opname/get-lines", record.InventDim).then(
      (r) => {
        record.Lines = r.data.map((d) => {
          d.QtyActual = "";
          d.Gap = "";
          d.Remarks = "";
          return d;
        });
        // listControl.value.setFormRecord(record);
        if (data.formMode != "new") {
          util.nextTickN(2, () => {
            lineConfig.value.gridRefreshed();
          });
        }
        data.dimInventory = Object.values(record.InventDim).join("|");
        data.loadingLine = false;
      },
      (e) => {
        data.loadingLine = false;
        return util.showError(e);
      }
    );
  }
}

function onPreSave(record) {
  // for (let i = 0; i < record?.Lines?.length; i++) {
  //   if (record.Lines[i].QtyActual == "") {
  //     return util.showError("line field Qty Actual required");
  //   }
  // }

  record?.Lines?.map((o) => {
    o.Gap = parseFloat(o.QtyActual) - o.QtyInSystem;
    if (o.Gap > 0) {
      o.Remarks = "OVER";
    } else if (o.Gap < 0) {
      o.Remarks = "MINUS";
    } else if (o.Gap === 0) {
      o.Remarks = "OK";
    }
    o.QtyActual = o.QtyActual.toString();
    return o;
  });

  const validInvDin = RefDimensionInventory.value.validate();
  const valid = listControl.value.formValidate();

  if (valid && validInvDin) {
    let payload = JSON.parse(JSON.stringify(record));
    if (payload.Status == "") {
      payload.Status = "DRAFT";
    }
    data.dimInventory = Object.values(payload.InventDim).join("|");
    listControl.value.setFormLoading(true);
    listControl.value.submitForm(
      payload,
      () => {
        listControl.value.setFormLoading(false);
      },
      () => {
        listControl.value.setFormLoading(false);
      }
    );
  }
}

function setFormMode(mode) {
  listControl.value.setFormMode(mode);
}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Stock Opname";
  }
}
function getPostingProfile(record) {
  util.nextTickN(2, () => {
    setLoadingForm(true);
    axios
      .post(`/scm/inventory/journal/type/find?TransactionType=Stock Opname`)
      .then(
        (r) => {
          if (r.data.length > 0) {
            data.keyJournalType = util.uuid();
            record.JournalTypeID = r.data[0]._id;
            record.PostingProfileID = r.data[0].PostingProfileID;
          }
        },
        (e) => util.showError(e)
      )
      .finally(function () {
        // getDetailEmployee("", record);
        setLoadingForm(false);
      });
  });
}

watch(
  () => currentTab.value,
  (nv) => {
    if (nv == 1) {
      let record = listControl.value.getFormRecord();
      onFieldChanged(record);
    }
  }
);

onMounted(() => {
  if (defaultList.length === 1) {
    data.search.Site = defaultList[0];
  }
  if (route.query.trxid !== undefined) {
    let getUrlParam = route.query.trxid;
    axios
      .post(`/scm/stock-opname/get`, [getUrlParam])
      .then(
        (r) => {
          listControl.value.setControlMode("form");
          listControl.value.setFormRecord(r.data);

          data.record = r.data;
        },
        (e) => util.showError(e)
      )
      .finally(() => {
        util.nextTickN(2, () => {
          router.replace({
            path: `/scm/StockOpname`,
          });
        });
      });
  }
});
</script>
