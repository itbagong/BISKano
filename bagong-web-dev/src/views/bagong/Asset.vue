<template>
  <div class="w-full">
    <div
      v-if="data.loadingDetail || data.loadingRef"
      class="w-full fixed h-full bg-[#94a3b86b] z-[999] flex justify-center items-center"
    >
      <loader kind="circle" />
    </div>
    <data-list
      class="card"
      ref="listControl"
      title="Asset"
      grid-config="/tenant/asset/gridconfig"
      form-config="/tenant/asset/formconfig"
      grid-read="/bagong/asset/get-assets"
      form-read="/tenant/asset/get"
      grid-mode="grid"
      grid-delete="/tenant/asset/delete"
      form-keep-label
      form-insert="/tenant/asset/insert"
      form-update="/tenant/asset/update"
      form-hide-submit
      :grid-fields="[
        'AssetType',
        'Purpose',
        'Site',
        'AcquisitionDate',
        'LifeTime',
        'AcquisitionAmount',
        'DepreciationAmount',
        'NBV',
        'LatestCustomer',
      ]"
      :form-fields="['AssetType', 'AdjustmentAccount', 'Dimension']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      :form-tabs-edit="data.tabsList"
      :form-tabs-view="data.tabsList"
      @alterGridConfig="alterGridConfig"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @formFieldChange="handleFieldChange"
      @postSave="onPostSave"
      @gridRowDeleted="ongridRowDeleted"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="data.customFilter"
    >
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilterCtl"
          v-model="data.customFilter"
          @initNewItem="initNewItemFilter"
          @preChange="changeFilter"
          :fieldsText="['Name']"
          hide-filter-date
          hide-filter-status
          @change="refreshGrid"
        >
          <template #filter_2="{ item }">
            <s-input
              class="w-[200px] filter-text"
              label="Police No | Hull No"
              v-model="item.PoliceHullNum"
            />
            <s-input
              useList
              label="Site"
              lookup-url="/tenant/dimension/find?DimensionType=Site"
              lookup-key="_id"
              :lookup-labels="['Label']"
              :lookup-searchs="['_id', 'Label']"
              v-model="item.SiteIDs"
              class="min-w-[180px]"
              multiple
            />
            <s-input
              useList
              label="Customer"
              lookup-url="/tenant/customer/find"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              v-model="item.CustomerIDs"
              class="min-w-[180px]"
              multiple
            />
            <s-input
              useList
              label="Asset Type"
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=AUT"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              v-model="item.AssetTypes"
              class="min-w-[180px]"
              multiple
            />
          </template>
        </grid-header-filter>
      </template>
      <!-- Slot Grid -->
      <template #grid_AssetType="{ item }">
        <s-input
          v-if="item.AssetType"
          v-model="item.AssetType"
          hide-label
          use-list
          :lookup-url="`/tenant/masterdata/find?_id=${item.AssetType}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          read-only
        />
      </template>
      <template #grid_Purpose="{ item }">
        <div v-if="item.DetailUnit">
          {{ item.DetailUnit.Purpose }}
        </div>
        <div v-else>&nbsp;</div>
      </template>
      <template #grid_AcquisitionDate="{ item }">
        <div v-if="item.AcquisitionDate">
          {{ moment(item.AcquisitionDate).format("DD-MM-YYYY") }}
        </div>
      </template>
      <template #grid_LifeTime="{ item }">
        <div v-if="item.LifeTime">
          {{ moment(item.LifeTime).format("DD-MM-YYYY") }}
        </div>
      </template>
      <template #grid_AcquisitionAmount="{ item }">
        <div>{{ util.formatMoney(item.AcquisitionAmount, {}) }}</div>
      </template>
      <template #grid_DepreciationAmount="{ item }">
        <div>{{ util.formatMoney(item.DepreciationAmount, {}) }}</div>
      </template>
      <template #grid_NBV="{ item }">
        <div>{{ util.formatMoney(item.NBV, {}) }}</div>
      </template>
      <template #grid_Site="{ item }">
        <dimension-editor
          hide-label
          :dimNames="['Site']"
          v-model="item.Dimension"
          read-only
        />
      </template>
      <template #grid_LatestCustomer="{ item, header }">
        <s-input
          v-if="item.LatestCustomer"
          v-model="item.LatestCustomer"
          hide-label
          use-list
          :lookup-url="`${header.input.lookupUrl}?${header.input.lookupKey}=${item.LatestCustomer}`"
          :lookup-key="header.input.lookupKey"
          :lookup-labels="header.input.lookupLabels"
          read-only
        />
      </template>

      <!-- Slot Form Tab -->
      <template #form_tab_Detail="{ item }">
        <bagong-detail
          :groupID="
            ['UNT', 'PRT', 'ELC'].includes(item.GroupID) ? item.GroupID : 'UNT'
          "
          :dimension="item.Dimension"
          v-model="data.detail"
          :recordId="item._id"
          @closeAttch="closeAttch"
          v-if="!data.loading && Object.keys(data.detail).length > 0"
        />
      </template>
      <template #form_tab_References="{}">
        <ref-template
          :ReferenceTemplate="data.refTemplate.ReferenceTemplateID"
          :readOnly="false"
          v-model="data.references"
          v-if="!data.loading"
        />
      </template>
      <template #form_tab_Depreciation="{}">
        <bagong-depreciation
          v-model="data.depreciation"
          v-if="!data.loading && Object.keys(data.depreciation).length > 0"
        />
        <bagong-depreciation-activity
          v-model="data.depreciationActivity"
          :references="data.references"
          :depreciation="data.depreciation"
          v-if="
            !data.loadingDetail && Object.keys(data.depreciation).length > 0
          "
        />
      </template>
      <template #form_tab_Other_Info="{}">
        <bagong-user-info
          title="User info"
          v-model="data.userinfo"
          v-if="!data.loadingDetail"
        />
      </template>
      <template #form_tab_Checklist="{ item, mode }">
        <Checklist
          v-if="!data.loading"
          v-model="data.checklist"
          :checklist-id="data.refTemplate.ChecklistTemplateID"
        />
      </template>

      <template #form_tab_Attachment="{ item, mode }">
        <attachment
          :journalId="item._id"
          :journalType="'ASSET_ASSET'"
          :tags="attachmentByTag(item)"
          ref="gridAttachment"
          single-save
        />
      </template>
      -->

      <!-- Slot Form Input -->
      <template #form_input_AssetType="{ item, mode }">
        <s-input
          class="min-w-[100px]"
          v-model="item.AssetType"
          keep-label
          label="Asset type"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=${getAssetTypeID(
            data.groupID
          )}`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          :read-only="mode == 'view'"
        />
      </template>
      <template #form_input_AdjustmentAccount="{ config, item, mode }">
        <s-input
          class="min-w-[100px]"
          keep-label
          :label="config.label"
          use-list
          v-model="item.AdjustmentAccount"
          lookup-url="/tenant/ledgeraccount/find"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          :read-only="mode == 'view'"
        />
        <s-form
          class="pt-2"
          v-model="data.bgAsset"
          :config="data.frmCfg"
          keep-label
          only-icon-top
          hide-submit
          hide-cancel
          ref="frmBagongAsset"
        />
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :read-only="mode == 'view'"
        ></dimension-editor-vertical>
      </template>

      <!-- Slot Form Buttons -->
      <template #form_buttons_1="{ item, mode }">
        <template v-if="profile.canUpdate && mode !== 'new'">
          <s-button
            icon="content-save"
            class="btn_primary"
            label="Save"
            @click="onSave(item)"
          />
        </template>
      </template>

      <!-- Slot Form Footer -->
      <template #form_footer_1></template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, inject, onMounted, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SForm, SInput, SButton, loadFormConfig } from "suimjs";
import { authStore } from "@/stores/auth.js";
import helper from "@/scripts/helper.js";
import refTemplate from "./widget/RefTemplate.vue";
import BagongDetail from "./widget/BagongAssetDetail.vue";
import BagongDepreciation from "./widget/BagongAssetDepreciation.vue";
import BagongDepreciationActivity from "./widget/BagongAssetDepreciationActivity.vue";
import BagongUserInfo from "./widget/BagongAssetUserInfo.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";
import Loader from "@/components/common/Loader.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import Checklist from "@/components/common/Checklist.vue";
import moment from "moment";

layoutStore().name = "tenant";

const FEATUREID = "Asset";
const profile = authStore().getRBAC(FEATUREID);
const dimensionSite = profile.Dimension?.filter((e) => e.Key === "Site").map(
  (e) => e.Value
);

const gridHeaderFilterCtl = ref(null);
const listControl = ref(null);
const frmBagongAsset = ref(null);
const gridAttachment = ref(null);

const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  tabsList: profile.canSpecial1
    ? [
        "General",
        "Detail",
        "References",
        "Depreciation",
        "Other Info",
        "Checklist",
        "Attachment",
      ]
    : ["General", "Detail", "Other Info", "Checklist", "Attachment"],
  record: {},
  detail: {},
  depreciation: {},
  depreciationActivity: [],
  checklist: [],
  references: [],
  userinfo: [],
  loading: false,
  refTemplate: {},
  frmCfg: {},
  bgAsset: {},
  groupID: "",
  loadFormCfg: true,
  customFilter: null,
  loadingDetail: false,
  loadingRef: false,
});
function resetGridHeaderFilter() {
  gridHeaderFilterCtl.value.reset();
}
function refreshGrid() {
  listControl.value.refreshGrid();
}
function initNewItemFilter(item) {
  item.SiteIDs = [];
  item.CustomerIDs = [];
  item.AssetTypes = [];
  item.PoliceHullNum = "";
}
function changeFilter(item, filters) {
  if (item.SiteIDs.length > 0) {
    filters.push({
      Op: "$and",
      Items: [
        {
          Op: "$eq",
          Field: "Dimension.Key",
          Value: "Site",
        },
        {
          Op: "$in",
          Field: "Dimension.Value",
          Value: [...item.SiteIDs],
        },
      ],
    });
  }
  if (item.CustomerIDs.length > 0) {
    filters.push({
      Op: "$in",
      Field: "LatestCustomer",
      Value: item.CustomerIDs,
    });
  }
  if (item.AssetTypes.length > 0) {
    filters.push({
      Op: "$in",
      Field: "AssetType",
      Value: item.AssetTypes,
    });
  }
  if (item.PoliceHullNum != "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Op: "$contains",
          Field: "PoliceNum",
          Value: [item.PoliceHullNum],
        },
        {
          Op: "$contains",
          Field: "HullNum",
          Value: [item.PoliceHullNum],
        },
      ],
    });
  }
}

function newRecord(record) {
  record._id = "";
  record.Name = "";
  data.formMode = "new";
  data.references = [];
  data.checklist = [];
  data.userinfo = [];
  data.detail = {};
  data.depreciation = {};
  data.depreciationActivity = [];
  data.bgAsset = {};
  data.bgAsset.IsActive = true;
  data.groupID = "";
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.references = [];
  data.checklist = [];
  data.userinfo = [];
  data.bgAsset = {};
  data.detail = {};
  data.depreciation = {};
  data.depreciationActivity = [];
  data.groupID = record.GroupID;
  getDetailData(record._id, record.GroupID);
  if (record.GroupID) getRefTemplateId(record.GroupID);
  openForm(record);
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", [
      (v) => {
        let vLen = 0;
        let consistsInvalidChar = false;

        v.split("").forEach((ch) => {
          vLen++;
          const validCar =
            "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
              ch
            ) >= 0;
          if (!validCar) consistsInvalidChar = true;
          //console.log(ch,vLen,validCar)
        });

        if (vLen < 3 || consistsInvalidChar)
          return "fixed asset number must be generated first";
        return "";
      },
    ]);
    setAttrFrmBgAsset(record.GroupID, "TrayekID");
    showHideDropdownDriverType(record.GroupID);
  });
}

function getAssetTypeID(groupID) {
  switch (groupID) {
    case "UNT":
      return "AUT";
      break;
    case "ELC":
      return "AOT";
      break;
    case "PRT":
      return "APT";
      break;
    default:
      return "AUT";
      break;
  }
}

function showHideDropdownDriverType(value) {
  switch (value) {
    case undefined:
    case null:
    case "ELC":
    case "PRT":
      listControl.value.setFormFieldAttr("DriveType", "hide", true);
      break;

    default:
      listControl.value.setFormFieldAttr("DriveType", "hide", false);
      break;
  }
}

function handleFieldChange(name, v1, v2, old, record) {
  if (name == "GroupID") {
    getRefTemplateId(v1);
    setAttrFrmBgAsset(v1, "TrayekID");
    data.groupID = v1;
    showHideDropdownDriverType(v1);
  }
}

async function getReadyFixedAssetNumber(id) {
  if (id == "") return;

  let payload = {
    AssetGroup: id,
  };
  let result = "";
  try {
    const r = await axios.post(
      "/fico/fixedassetnumber/get-ready-fixed-asset-number",
      payload
    );
    result = r.data.FixedAssetNumber;
  } catch (e) {
    util.showError(e);
  }
  return result;
}

function getRefTemplateId(id) {
  data.loadingRef = true;
  const url = "/tenant/assetgroup/get";
  axios.post(url, [id]).then(
    (r) => {
      data.refTemplate = r.data;
      data.loadingRef = false;
      data.depreciation = {
        AssetDuration: r.data.AssetDuration,
        DepreciationPeriod: r.data.DepreciationPeriod,
      };
    },
    (e) => {
      data.loadingRef = false;
    }
  );
}

function getDetailData(id, assetId) {
  data.loadingDetail = true;
  const url = "/bagong/asset/get";
  const emptyObj = {
    DetailElectronic: {
      OtherInfo: {},
    },
    DetailProperty: {
      OtherInfo: {},
    },
    DetailUnit: {
      OtherInfo: {},
    },
  };
  axios.post(url, [id]).then(
    (r) => {
      data.loadingDetail = false;
      // if (r.data.length > 0) {
      let obj = r.data;
      // console.log(obj);
      data.references = obj.References;
      data.checklist = obj.ChecklistTemp;
      data.userinfo = [...obj.UserInfo];
      // console.log(data.userinfo, obj.UserInfo);
      data.detail = obj;
      data.depreciation = obj.Depreciation;
      data.depreciationActivity = obj.DepreciationActivity;
      data.bgAsset.IsActive = obj.IsActive;
      data.bgAsset.TrayekID = obj.TrayekID;
      // } else {

      // }
    },
    (e) => {
      util.showError(e);
      data.detail = emptyObj;
      data.loadingDetail = false;
    }
  );
}
async function onSave(record) {
  // set dimension site equel with last site in user info detail
  let vSite = "";
  if (data.userinfo.length > 0) {
    // get site in user info last data
    var dimArr = ["PC", "CC", "Site", "Asset"];
    vSite = data.userinfo[data.userinfo.length - 1].SiteID;
    const dimKey = Object.groupBy(record.Dimension, (obj) => obj.Key);

    for (var i = 0; i < dimArr.length; i++) {
      if (dimKey[dimArr[i]] === undefined) {
        record.Dimension.push({
          Key: dimArr[i],
          Value: "",
        });
      }
      if (dimArr[i] == "Site") {
        const findIdx = record.Dimension.findIndex(
          (data) => data.Key == "Site"
        );
        record.Dimension[findIdx].Value = vSite;
      }
    }
  }

  if (data.formMode == "new") {
    listControl.value.setFormLoading(true);
    record._id = await getReadyFixedAssetNumber(record.GroupID);
    listControl.value.setFormLoading(false);
  }

  if (record._id != "") {
    listControl.value.submitForm(
      record,
      () => {},
      () => {}
    );
  } else {
    util.showError("fixed asset number must be generated first");
  }
}

function onPostSave(record) {
  // save to bagong after save tenant
  let objMode = {
    new: "insert",
    edit: "update",
  };
  const url = "/bagong/asset/" + objMode[data.formMode];
  let param = data.detail;
  param.TrayekID = data.bgAsset.TrayekID;
  param.IsActive = data.bgAsset.IsActive;
  param.References = data.references;
  param.UserInfo = data.userinfo;
  param.Depreciation = data.depreciation;
  param.DepreciationActivity = data.depreciationActivity;
  param.Name = record.Name;
  param._id = record._id;
  param.ChecklistTemp = data.checklist;
  param.Dimension = record.Dimension;
  param.LatestCustomer =
    data.userinfo.length > 0 ? data.userinfo.at(-1).CustomerID : "";

  for (let i in param.UserInfo) {
    let o = param.UserInfo[i];
    o.AssetDateFromString = moment(o.AssetDateFrom).format("YYYYMMDD");
    o.AssetDateToString = moment(o.AssetDateTo).format("YYYYMMDD");
  }

  axios.post(url, param).then(
    (r) => {},
    (e) => {
      util.showError(e.error);
    }
  );

  // update data to fixed asset number
  if (data.formMode == "new") {
    updateFixedAssetNumberList(record._id, true);
  }
}
function closeAttch() {
  gridAttachment.value.refreshGrid();
}
function setAttrFrmBgAsset(param, field) {
  let isUnit = param == "UNT";
  frmBagongAsset.value.setFieldAttr(field, "hide", !isUnit);
}

function ongridRowDeleted(params) {
  const url = "/bagong/asset/find?_id=" + params._id;
  axios.post(url).then(
    (r) => {
      if (r.data.length > 0) onDeleteDetail(r.data[0]);
    },
    (e) => {}
  );

  // update data to fixed asset number
  updateFixedAssetNumberList(params._id, false);
}

function updateFixedAssetNumberList(id, isUsed) {
  let payload = {
    FixedAssetNumber: id,
    IsUsed: isUsed,
  };

  axios.post("/fico/fixedassetnumber/update-is-used", payload).then(
    (r) => {},
    (e) => {
      util.showError(e);
    }
  );
}

function onDeleteDetail(id) {
  const url = "/bagong/asset/delete";
  axios.post(url, id).then(
    (r) => {},
    (e) => {}
  );
}

function alterGridConfig(cfg) {
  const filteredFields = cfg.fields.filter(
    (el) =>
      [
        "AcquisitionAccount",
        "DepreciationAccount",
        "DisposalAccount",
        "AdjustmentAccount",
      ].indexOf(el.field) == -1
  );
  let newFields = [
    {
      field: "Purpose",
      label: "Purpose",
      readType: "show",
      input: {
        field: "Purpose",
        label: "Purpose",
      },
    },
    {
      field: "Site",
      label: "Site",
      readType: "show",
      input: {
        field: "Site",
        label: "Site",
      },
    },
    {
      field: "LatestCustomer",
      label: "Customer",
      labelField: "LatestCustomer",
      readType: "show",
      input: {
        lookupUrl: "/tenant/customer/find",
        lookupKey: "_id",
        lookupLabels: ["Name"],
        lookupSearchs: ["_id", "Name"],
        useList: true,
        field: "LatestCustomer",
        label: "Customer",
      },
    },
    {
      field: "AcquisitionDate",
      label: "Acquisition date",
      readType: "show",
      input: {
        field: "AcquisitionDate",
        label: "Acquisition date",
      },
    },
    {
      field: "LifeTime",
      label: "Life time",
      readType: "show",
      input: {
        field: "LifeTime",
        label: "Life time",
      },
    },
    {
      field: "AcquisitionAmount",
      label: "Acquisition amount",
      readType: "show",
      input: {
        field: "AcquisitionAmount",
        label: "Acquisition amount",
      },
    },
    {
      field: "DepreciationAmount",
      label: "Depreciation amount",
      readType: "show",
      input: {
        field: "DepreciationAmount",
        label: "Depreciation amount",
      },
    },
    {
      field: "NBV",
      label: "NBV",
      readType: "show",
      input: {
        field: "NBV",
        label: "NBV",
      },
    },
  ];
  if (!profile.canSpecial1) {
    newFields = newFields.filter((o) =>
      ["Purpose", "Site", "LatestCustomer"].includes(o.field)
    );
  }
  cfg.fields = [...filteredFields, ...newFields];
}
function attachmentByTag(record) {
  return [`ASSET_ASSET_${record._id}`];
}

onMounted(() => {
  loadFormConfig(axios, "/bagong/asset/formconfig").then(
    (r) => {
      data.frmCfg = r;
    },
    (e) => util.showError(e)
  );
});

// watch(
//   () => data.userinfo,
//   (nv) => {
//     if (nv && nv.length > 0) {
//       data.userinfo = nv;
//     }
//   },
//   { deep: true }
// );

// watch(
//   () => data.groupID,
//   (nv) => {
//     data.groupID = nv;
//   },
//   { deep: true }
// );
</script>
