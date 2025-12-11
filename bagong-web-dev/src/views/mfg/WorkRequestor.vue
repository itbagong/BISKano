<template>
  <div class="w-full">
    <data-list
      v-show="data.appMode == 'grid'"
      id="wr-data-list"
      class="card grid-line-items"
      ref="listControl"
      :title="data.titleForm"
      form-hide-submit
      grid-config="/mfg/work/request/gridconfig"
      form-config="/mfg/work/request/formconfig"
      grid-read="/mfg/work/request/gets"
      form-read="/mfg/work/request/get"
      grid-mode="grid"
      grid-delete="/mfg/work/request/delete"
      form-keep-label
      grid-hide-delete
      form-insert="/mfg/work/request/save"
      form-update="/mfg/work/request/save"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :grid-fields="['Department', 'Dimension', 'Status']"
      :form-fields="[
        '_id',
        'Name',
        'Department',
        'Dimension',
        'WorkRequestType',
        'JournalTypeID',
        'EquipmentType',
        'EquipmentNo',
        'StartDownTime',
        'TargetFinishTime',
        'Kilometer',
        'SourceType',
        'SourceID',
      ]"
      :form-tabs-new="['General']"
      :form-tabs-edit="['General', 'Attachment']"
      :form-tabs-view="['General', 'Attachment']"
      :stay-on-form-after-save="data.stayOnFormAfterSave"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="customFilter"
      @controlModeChanged="onControlModeChanged"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @alterGridConfig="onAlterGridConfig"
      @form-field-change="onFormFieldChange"
    >
      <template #grid_header_search="{ config }">
        <s-input
          ref="refrequestor"
          v-model="data.search.requestor"
          lookup-key="_id"
          label="Requestor"
          class="w-full"
          use-list
          :lookup-url="`/tenant/employee/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
          :lookup-payload-builder="
            (search) =>
              lookupPayloadSearch(
                search,
                ['_id', 'Name'],
                data.search.requestor,
                data.search
              )
          "
        ></s-input>
        <s-input
          ref="refName"
          v-model="data.search.Text"
          lookup-key="_id"
          label="Text"
          class="w-[400px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="WR Date From"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="WR Date To"
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
          class="w-[400px]"
          use-list
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
      <template #form_tab_Attachment="{ item }">
        <s-grid-attachment
          v-model="item.Attachment"
          ref="gridAttachment"
          journalType="WorkRequestor"
          :journalId="item._id"
          :tags="linesTag"
          @pre-Save="preSaveAttachment"
        />
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #grid_Department="{ item }">
        {{
          item.Dimension &&
          item.Dimension.find((_dim) => _dim.Key === "CC") &&
          item.Dimension.find((_dim) => _dim.Key === "CC")["Value"] != ""
            ? item.Dimension.find((_dim) => _dim.Key === "CC")["Value"]
            : ""
        }}
      </template>
      <template #grid_Dimension="{ item }">
        {{ item.Site }}
      </template>
      <template #form_input_Name="{ item, config }">
        <s-input
          ref="refRequestor"
          label="Requestor"
          v-model="item.Name"
          class="w-full"
          :disabled="data.statusDisabled.includes(item.Status)"
          :required="true"
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeRequestor(v1, item);
            }
          "
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilderRequestor(search, config, item.Name, item)
          "
        ></s-input>
      </template>
      <template #form_input_Department="{ item, config }">
        <s-input
          :key="data.keyDept"
          ref="refDepartment"
          label="Department"
          v-model="item.Department"
          class="w-full"
          :required="false"
          :disabled="true"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=DME`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_input_JournalTypeID="{ item }">
        <s-input
          :key="data.keyJournalType"
          ref="refInput"
          label="Journal Type"
          v-model="item.JournalTypeID"
          class="w-full"
          :required="false"
          :disabled="true"
          use-list
          :lookup-url="`/mfg/workrequestor/journal/type/find`"
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
      <template #form_input_SourceType="{ item }">
        <s-input
          ref="refInput"
          label="Source Type"
          v-model="item.SourceType"
          class="w-full"
          :disabled="data.statusDisabled.includes(item.Status)"
          use-list
          :items="data.sourceTypeOption"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onFormFieldChange('SourceType', v1, v2, old, item);
            }
          "
        ></s-input>
      </template>
      <template #form_input_SourceID="{ item }">
        <s-input
          v-if="item.SourceType == 'Sales Order'"
          ref="refInput"
          label="Ref No."
          v-model="item.SourceID"
          class="w-full"
          :required="true"
          :disabled="data.statusDisabled.includes(item.Status)"
          :use-list="true"
          :lookup-url="`/sdp/salesorder/find`"
          lookup-key="_id"
          :lookup-labels="['SalesOrderNo']"
          :lookup-searchs="['_id', 'SalesOrderNo']"
          keepLabel
          :hide-label="false"
        ></s-input>
        <s-input
          v-else
          ref="refInput"
          label="Ref No."
          v-model="item.SourceID"
          class="w-full"
          :required="true"
          :read-only="true"
          keepLabel
          :hide-label="false"
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          :key="data.keyFinancialDimension"
          ref="FinancialDimension"
          sectionTitle="Financial Dimension"
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :readOnly="data.statusDisabled.includes(item.Status)"
          @change="
            (field, v1, v2) => {
              onChangeDimension(field, v1, v2, item);
            }
          "
        ></dimension-editor>
      </template>
      <template #form_input_WorkRequestType="{ item }">
        <s-input
          ref="refWorkRequestType"
          v-model="item.WorkRequestType"
          :required="true"
          :disabled="
            data.statusDisabled.includes(item.Status) ||
            item.SourceType == 'Sales Order'
          "
          label="Work Request Type"
          use-list
          :items="['Production', 'Service']"
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              setAttr(v1, v2, item);
            }
          "
        ></s-input>
      </template>
      <template #form_input_EquipmentType="{ item, config }">
        <s-input
          ref="refEquipmentType"
          :required="true"
          v-model="item.EquipmentType"
          :disabled="data.statusDisabled.includes(item.Status)"
          label="Asset Type"
          use-list
          :use-list="true"
          :lookup-payload-builder="
            (search) =>
              lookupPayloadAssetType(search, config, item.EquipmentType, item)
          "
          :lookup-url="`/tenant/assetgroup/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          keepLabel
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onFormFieldChange('EquipmentType', v1, v2, old, item);
            }
          "
        ></s-input>
      </template>
      <template #form_input_EquipmentNo="{ item, config }">
        <s-input
          :key="data.keyEquipmentNo"
          ref="refEquipmentNo"
          v-model="item.EquipmentNo"
          :disabled="data.statusDisabled.includes(item.Status)"
          label="Asset No"
          use-list
          :use-list="true"
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilder(search, config, item.EquipmentNo, item)
          "
          :lookup-url="
            item.EquipmentType
              ? `/tenant/asset/find?GroupID=${item.EquipmentType}`
              : `/tenant/asset/find`
          "
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          keepLabel
          class="w-full"
        ></s-input>
      </template>
      <template #form_input_StartDownTime="{ item }">
        <s-input
          ref="refStartDowntime"
          v-model="item.StartDownTime"
          :disabled="
            data.disableField ||
            data.statusDisabled.includes(item.Status) ||
            item.SourceType == 'Sales Order'
          "
          label="Start Down Time"
          class="w-full"
          kind="datetime-local"
        ></s-input>
      </template>
      <template #form_input_TargetFinishTime="{ item }">
        <s-input
          ref="refTargetFinishTime"
          v-model="item.TargetFinishTime"
          :disabled="
            data.disableField ||
            data.statusDisabled.includes(item.Status) ||
            item.SourceType == 'Sales Order'
          "
          label="Target Finish Time"
          class="w-full"
          kind="datetime-local"
        ></s-input>
      </template>
      <template #form_input_Kilometers="{ item }">
        <s-input
          ref="refKilometer"
          v-model="item.Kilometers"
          :disabled="
            data.disableField || data.statusDisabled.includes(item.Status)
          "
          label="Kilometers"
          class="w-full"
          kind="Number"
        ></s-input>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          class="bg-transparent hover:bg-blue-500 hover:text-black"
          label="Preview"
          icon="eye-outline"
          @click="onPreview"
        ></s-button>
        <s-button
          v-if="['', 'DRAFT'].includes(item.Status)"
          icon="content-save"
          class="btn_primary"
          label="Save"
          :disabled="data.loading.isProcess"
          @click="postSave(item)"
        />
        <form-buttons-trx
          :key="data.btnTrx"
          :disabled="data.loading.isProcess"
          :status="item.Status"
          :moduleid="`mfg`"
          :autoPost="false"
          :autoReopen="false"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          journal-type-id="Work Request"
          @preSubmit="trxPreSubmit"
          @pre-reopen="preReopen"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
        />
      </template>
      <template #form_footer_1="{ item }">
        <RejectionMessageList
          v-if="item.Status === 'REJECTED'"
          ref="listRejectionMessage"
          journalType="Work Request"
          :journalID="item._id"
        ></RejectionMessageList>
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #grid_item_button_delete="{ item }">
        <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
      </template>
    </data-list>

    <PreviewReport
      v-if="data.appMode == 'preview'"
      class="card w-full"
      title="Preview"
      :preview="data.preview"
      @close="closePreview"
      SourceType="Work Request"
      :SourceJournalID="data.record._id"
      :hideSignature="false"
    >
    </PreviewReport>
  </div>
</template>

<script setup>
import {
  reactive,
  ref,
  nextTick,
  inject,
  onMounted,
  computed,
  watch,
} from "vue";
import { useRoute, useRouter } from "vue-router";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth";
import { DataList, util, SForm, SInput, SButton, loadFormConfig } from "suimjs";
import helper from "@/scripts/helper.js";
import moment from "moment";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import RejectionMessageList from "../scm/widget/RejectionMessageList.vue";

layoutStore().name = "tenant";
const featureID = "WorkRequest";
const profile = authStore().getRBAC(featureID);
const headOffice = layoutStore().headOfficeID;
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);

const listControl = ref(null);
const AttachmentConfig = ref(null);
const gridAttachment = ref(null);
const FinancialDimension = ref(null);
const refEquipmentType = ref(null);
const axios = inject("axios");
const route = useRoute();
const router = useRouter();

const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];
let customFilter = computed(() => {
  let filters = [];
  let filtersWR = [];
  if (data.search.Text !== null && data.search.Text !== "") {
    filtersWR.push(
      {
        Field: "_id",
        Op: "$contains",
        Value: [data.search.Text],
      },
      {
        Field: "Name",
        Op: "$contains",
        Value: [data.search.Text],
      },
      {
        Field: "EquipmentNo",
        Op: "$contains",
        Value: [data.search.Text],
      },
      {
        Field: "SourceID",
        Op: "$contains",
        Value: [data.search.Text],
      }
    );
  }

  if (data.search.requestor !== null && data.search.requestor !== "") {
    filters.push({
      Field: "Name",
      Op: "$eq",
      Value: data.search.requestor,
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
  let items = [
    {
      Op: "$or",
      items: filtersWR,
    },
  ];
  if (data.search.Text !== null && data.search.Text !== "") {
    filters = [...filters, ...items];
  }

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

const linesTag = computed({
  get() {
    let ReffNo = data.record.ReffNo
      ? [`${data.record.SourceID.slice(0, 2)}_${data.record.SourceID}`]
      : [];
    const tags =
      ReffNo && data.record._id
        ? [...[`WR_${data.record._id}`], ...ReffNo]
        : data.record._id
        ? [`WR_${data.record._id}`]
        : ReffNo;

    return tags;
  },
});

const addTags = computed({
  get() {
    return [`WR_${data.record._id}`];
  },
});

const reffTags = computed({
  get() {
    let ReffNo = data.record.ReffNo
      ? [`${data.record.SourceID.slice(0, 2)}_${data.record.SourceID}`]
      : [];
    return ReffNo;
  },
});

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  titleForm: "Work Request",
  statusDisabled: [
    "READY",
    "POSTED",
    "SUBMITTED",
    "IN PROGRESS",
    "COMPLETED",
    "REJECTED",
  ],
  btnTrx: util.uuid(),
  keyDept: util.uuid(),
  keyJournalType: util.uuid(),
  keyEquipmentNo: util.uuid(),
  keyFinancialDimension: util.uuid(),
  disableField: false,
  stayOnFormAfterSave: true,
  isEdit: true,
  record: {
    _id: "",
    Status: "",
    Dimension: [],
  },
  loading: {
    isProcess: false,
  },
  sourceTypeOption: [
    "Assembly",
    "Routine Check",
    "Sales Order",
    "Non-Routine",
    "Breakdown",
  ],
  search: {
    Text: "",
    requestor: "",
    DateFrom: null,
    DateTo: null,
    Status: "",
    Site: "",
  },
  listTTD: [],
  requestorName: "",
});

function getRequestor(record) {
  axios.post("/bagong/employee/get", [record.Name]).then(
    (r) => {
      data.requestorName = r.data.Name;
    },
    (e) => util.showError(e)
  );
}

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Work Request`;

  util.nextTickN(2, async () => {
    record._id = "";
    record.Status = "";
    record.Kilometers = 0;
    record.StatusApproval = "";
    record.TrxDate = new Date();
    record.StartDownTime = moment().format("YYYY-MM-DDTHH:mm");
    record.TargetFinishTime = moment().format("YYYY-MM-DDTHH:mm");
    data.record = record;
    getPostingProfile(record);
    listControl.value.setFormFieldAttr("Equipment", "readOnly", false);
  });
  openForm(record);
}

function editRecord(record) {
  data.titleForm = `Edit Work Request | ${record._id}`;
  record.StartDownTime = moment(
    moment(record.StartDownTime ? record.StartDownTime : new Date()).format(
      "YYYY-MM-DDTHH:mm:00Z"
    )
  ).format("YYYY-MM-DDTHH:mm");
  record.TargetFinishTime = moment(
    moment(
      record.TargetFinishTime ? record.TargetFinishTime : new Date()
    ).format("YYYY-MM-DDTHH:mm:00Z")
  ).format("YYYY-MM-DDTHH:mm");
  data.record = record;
  openForm(record);

  nextTick(() => {
    if (record.EquipmentType == "PRT" || record.EquipmentType == "ELC") {
      listControl.value.setFormFieldAttr("Kilometers", "required", false);
    } else {
      listControl.value.setFormFieldAttr(
        "Kilometers",
        "required",
        record.WorkRequestType == "Service"
      );
    }
    if (record.SourceID) {
      getPostingProfile(record);
      getAsset(record);
    } else {
      listControl.value.setFormFieldAttr("Equipment", "readOnly", false);
      listControl.value.setFormFieldAttr("EquipmentType", "readOnly", false);
    }
    getApproval(record);
  });
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", roleID);
    listControl.value.setFormFieldAttr(
      "_id",
      "hide",
      record._id ? false : true
    );
    const el = document.querySelector(
      "#wr-data-list .form_inputs > div.flex.section_group_container > div:nth-child(1) > div > div > div:nth-child(1)"
    );
    if (record._id == "") {
      el.style.display = "none";
    } else {
      el.style.display = "block";
    }
    if (data.statusDisabled.includes(record.Status)) {
      listControl.value.setFormMode("view");
    }
    data.btnTrx = util.uuid();
    data.isEdit = false;
  });
}

function preSaveAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [`WR_${data.record._id}`];
    return asset;
  });
}

function postSaveAttachment() {
  const payload = {
    Addtags: addTags.value,
    Tags: reffTags.value,
  };
  if (payload.Tags.length > 0) {
    helper.updateTags(axios, payload);
  }
}

function onAlterGridConfig(cfg) {
  util.nextTickN(2, () => {
    cfg.setting.idField = "LastUpdate";
    cfg.setting.sortable = [
      "LastUpdate",
      "Created",
      "TrxDate",
      "Status",
      "_id",
    ];
  });
}

function getJournalType(_id, item) {
  item.PostingProfileID = "";
  axios
    .post("/mfg/workrequestor/journal/type/find?_id=" + _id, { sort: ["-_id"] })
    .then(
      (r) => {
        if (r.data.length > 0) {
          item.PostingProfileID = r.data[0].PostingProfileID;
        }
      },
      (e) => util.showError(e)
    );
}

function onChangeRequestor(_id, record) {
  record.Department = "";
  if (_id) {
    axios.post("/bagong/employee/get", [_id]).then(
      (r) => {
        record.Department = r.data.Detail.Department;
        data.keyDept = util.uuid();
      },
      (e) => util.showError(e)
    );
  } else {
    data.keyDept = util.uuid();
  }
}

function postSave(record) {
  let payload = JSON.parse(JSON.stringify(record));
  payload.Status = "DRAFT";
  payload.StartDownTime = moment(record.StartDownTime)
    .utc()
    .format("YYYY-MM-DDTHH:mm:00Z");
  payload.TargetFinishTime = moment(record.TargetFinishTime)
    .utc()
    .format("YYYY-MM-DDTHH:mm:00Z");

  payload.TrxType = "Work Request";
  payload.Kilometers = payload.Kilometers ? payload.Kilometers : 0;

  const pc = record.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = record.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = record.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;

  let validate = true;
  if (FinancialDimension.value) {
    validate = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      validate = false;
    }
  }

  if (!record.Name) {
    return util.showError("field Requestor is required");
  }
  if (!record.EquipmentType) {
    return util.showError("field Asset Type  is required");
  }
  if (!record.WorkRequestType) {
    return util.showError("field Work Request Type is required");
  }
  if (!["PRT", "ELC"].includes(record.EquipmentType)) {
    if (record.WorkRequestType == "Service" && record.Kilometers <= 0) {
      return util.showError("field Kilometers is required");
    }
  }

  if (validate) {
    if (listControl.value) {
      listControl.value.setFormLoading(true);
    }
    data.loading.isProcess = true;
    payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
    axios
      .post("/mfg/work/request/save", payload)
      .then(
        (r) => {
          let newRecord = { ...payload, ...r.data };
          newRecord.StartDownTime = moment(
            moment(r.data.StartDownTime).format("YYYY-MM-DDTHH:mm:00Z")
          ).format("YYYY-MM-DD HH:mm");
          newRecord.TargetFinishTime = moment(
            moment(r.data.TargetFinishTime).format("YYYY-MM-DDTHH:mm:00Z")
          ).format("YYYY-MM-DD HH:mm");
          listControl.value.setFormMode("edit");
          listControl.value.setFormRecord(newRecord);
          data.record = newRecord;
          data.stayOnFormAfterSave = true;
          util.nextTickN(2, () => {
            if (gridAttachment.value) {
              gridAttachment.value.Save();
              postSaveAttachment();
            }
          });

          return util.showInfo("Work Request has been successful save");
        },
        (e) => {
          return util.showError(e);
        }
      )
      .finally(function () {
        data.loading.isProcess = false;
        if (listControl.value) {
          listControl.value.setFormLoading(false);
        }
      });
  } else {
    return util.showError("field is required");
  }
}

function trxPostSubmit() {
  if (listControl.value) {
    listControl.value.setFormLoading(false);
  }
  listControl.value.refreshForm();
  listControl.value.setControlMode("grid");
  listControl.value.refreshList();
  data.titleForm = `Work Request`;
  data.loading.isProcess = false;
  return util.showInfo("Work Request has been successful save");
}

function trxErrorSubmit() {
  if (listControl.value) {
    listControl.value.setFormLoading(false);
  }
  data.loading.isProcess = false;
}
function trxPreSubmit(status, action, doSubmit) {
  let payload = JSON.parse(JSON.stringify(data.record));
  let validate = true;
  payload.Status = "DRAFT";
  payload.StartDownTime = moment(data.record.StartDownTime)
    .utc()
    .format("YYYY-MM-DDTHH:mm:00Z");
  payload.TargetFinishTime = moment(data.record.TargetFinishTime)
    .utc()
    .format("YYYY-MM-DDTHH:mm:00Z");

  payload.TrxType = "Work Request";

  const pc = payload.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = payload.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = payload.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;

  if (FinancialDimension.value) {
    validate = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      validate = false;
    }
  }

  if (!payload.Name) {
    return util.showError("field Requestor is required");
  }
  if (!payload.EquipmentType) {
    return util.showError("field Asset Type  is required");
  }
  if (!payload.WorkRequestType) {
    return util.showError("field Work Request Type is required");
  }

  if (!["PRT", "ELC"].includes(payload.EquipmentType)) {
    if (payload.WorkRequestType == "Service" && payload.Kilometers <= 0) {
      return util.showError("field Kilometers is required");
    }
  }

  if (validate) {
    if (listControl.value) {
      listControl.value.setFormLoading(true);
    }
    data.loading.isProcess = true;
    if (data.record.Status == "DRAFT") {
      payload.TrxType = "Work Request";
      payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
      axios.post("/mfg/work/request/save", payload).then(
        (r) => {
          util.nextTickN(2, () => {
            if (gridAttachment.value) {
              gridAttachment.value.Save();
            }
            postSaveAttachment();
            doSubmit();
          });
        },
        (e) => {
          util.showError(e);
        }
      );
    } else {
      util.nextTickN(2, () => {
        doSubmit();
      });
    }
  } else {
    return util.showError("field is required");
  }
}

function preReopen() {
  let payload = JSON.parse(JSON.stringify(data.record));
  payload.Status = "DRAFT";
  payload.StartDownTime = moment(data.record.StartDownTime)
    .utc()
    .format("YYYY-MM-DDTHH:mm:00Z");
  payload.TargetFinishTime = moment(data.record.TargetFinishTime)
    .utc()
    .format("YYYY-MM-DDTHH:mm:00Z");
  data.loading.isProcess = true;
  axios.post("/mfg/work/request/save", payload).then(
    (r) => {
      util.nextTickN(2, () => {
        listControl.value.refreshForm();
        listControl.value.setControlMode("grid");
        listControl.value.refreshList();
        data.titleForm = `Work Request`;
        data.loading.isProcess = false;
      });
    },
    (e) => {
      data.loading.isProcess = false;
      return util.showError(e);
    }
  );
}

function setAttr(v1, v2, item) {
  let disable = false;
  item.Kilometer = 0;
  if (v1 == "Production") {
    disable = true;
  }

  util.nextTickN(2, () => {
    data.disableField = disable;
    listControl.value.setFormFieldAttr(
      "Kilometers",
      "required",
      !disable && item.WorkRequestType == "Service"
    );
  });
}

function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Work Request";
    data.isEdit = true;
  }
}

function lookupPayloadBuilderRequestor(search, config, value, item) {
  const qp = {};
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
  const profileSite = profile.Dimension.filter((_dim) => _dim.Key === "Site")
    .map((d) => {
      return d.Value;
    })
    .filter((x) => x != null);

  let Site =
    item.Dimension &&
    item.Dimension.find((_dim) => _dim.Key === "Site") &&
    item.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? item.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : null;

  let siteOffice =
    item.Dimension &&
    item.Dimension.find((_dim) => _dim.Key === "Site") &&
    item.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? item.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : null;

  Site = [...profileSite, ...[Site]].filter((x) => x != null);
  const querySite = [
    {
      Field: "Dimension.Key",
      Op: "$eq",
      Value: "Site",
    },
    {
      Field: "Dimension.Value",
      Op: "$in",
      Value: Site,
    },
  ];

  if (profileSite.length != 0 && Site.length > 0 && siteOffice != headOffice) {
    qp.Where = {
      Op: "$and",
      items: querySite,
    };
  }
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "_id", Op: "$contains", Value: [search] },
          { Field: "Name", Op: "$contains", Value: [search] },
        ],
      },
    ];
    if (
      profileSite.length != 0 &&
      Site.length > 0 &&
      siteOffice != headOffice
    ) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }

  return qp;
}

function lookupPayloadBuilder(search, config, value, item) {
  const qp = {};
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
  const Site =
    item.Dimension &&
    item.Dimension.find((_dim) => _dim.Key === "Site") &&
    item.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? item.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : undefined;
  const querySite = [
    {
      Field: "Dimension.Key",
      Op: "$eq",
      Value: "Site",
    },
    {
      Field: "Dimension.Value",
      Op: "$eq",
      Value: Site,
    },
  ];
  if (Site && Site != headOffice) {
    qp.Where = {
      Op: "$and",
      items: querySite,
    };
  }
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "_id", Op: "$contains", Value: [search] },
          { Field: "Name", Op: "$contains", Value: [search] },
        ],
      },
    ];
    if (Site && Site != headOffice) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }

  return qp;
}
function lookupPayloadAssetType(search, config, value, item) {
  const qp = {};
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
  const CC =
    item.Dimension &&
    item.Dimension.find((_dim) => _dim.Key === "CC") &&
    item.Dimension.find((_dim) => _dim.Key === "CC")["Value"] != ""
      ? item.Dimension.find((_dim) => _dim.Key === "CC")["Value"]
      : undefined;

  const query = [
    {
      Field: "_id",
      Op: "$ne",
      Value: "UNT",
    },
  ];

  if (CC == "PLN") {
    qp.Where = {
      Op: "$and",
      items: query,
    };
  }
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "_id", Op: "$contains", Value: [search] },
          { Field: "Name", Op: "$contains", Value: [search] },
        ],
      },
    ];
    if (CC && CC == "PLN") {
      items = [...items, ...query];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }

  return qp;
}
function lookupPayloadSearch(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;
  //setting search
  const Site =
    profile.Dimension &&
    profile.Dimension.find((_dim) => _dim.Key === "Site") &&
    profile.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? profile.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : undefined;
  const querySite = [
    {
      Field: "Dimension.Key",
      Op: "$eq",
      Value: "Site",
    },
    {
      Field: "Dimension.Value",
      Op: "$eq",
      Value: Site,
    },
  ];
  if (Site) {
    qp.Where = {
      Op: "$and",
      items: querySite,
    };
  }
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "_id", Op: "$contains", Value: [search] },
          { Field: select[1], Op: "$contains", Value: [search] },
        ],
      },
    ];
    if (Site) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}
function getPostingProfile(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormLoading(true);
    axios
      .post(`/mfg/workrequestor/journal/type/find`)
      .then(
        (r) => {
          if (r.data.length > 0) {
            record.JournalTypeID = r.data[0]._id;
            record.PostingProfileID = r.data[0].PostingProfileID;
            data.keyJournalType = util.uuid();
          }

          record.Dimension.map((d) => {
            if (d.Key == "CC") {
              d.Value = "OPS";
            }
            return d;
          });
          data.keyFinancialDimension = util.uuid();
        },
        (e) => util.showError(e)
      )
      .finally(function () {
        listControl.value.setFormLoading(false);
      });
  });
}
function getAsset(record) {
  util.nextTickN(2, () => {
    axios
      .post(`/tenant/asset/get`, [record.EquipmentNo])
      .then(
        (r) => {
          record.EquipmentType = r.data.GroupID;
          if (record.SourceID) {
            listControl.value.setFormFieldAttr("Equipment", "readOnly", true);
            listControl.value.setFormFieldAttr(
              "EquipmentType",
              "readOnly",
              true
            );
          }
        },
        (e) => {
          util.showError(e);
        }
      )
      .finally(function () {});
  });
}

function onPreview() {
  getRequestor(data.record);
  data.appMode = "preview";
}
function closePreview() {
  data.appMode = "grid";
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
function onFormFieldChange(field, v1, v2, old, record) {
  switch (field) {
    case "SourceType":
      record.SourceID = "";
      if (v1 == "Sales Order") {
        listControl.value.setFormFieldAttr("WorkRequestType", "readOnly", true);
        listControl.value.setFormFieldAttr("StartDownTime", "readOnly", true);
        listControl.value.setFormFieldAttr(
          "TargetFinishTime",
          "readOnly",
          true
        );
        record.WorkRequestType = "Production";
      } else {
        listControl.value.setFormFieldAttr(
          "WorkRequestType",
          "readOnly",
          false
        );
        listControl.value.setFormFieldAttr("StartDownTime", "readOnly", false);
        listControl.value.setFormFieldAttr(
          "TargetFinishTime",
          "readOnly",
          false
        );
      }
      break;
    case "EquipmentType":
      record.EquipmentNo = "";
      if (v1 == "PRT" || v1 == "ELC") {
        listControl.value.setFormFieldAttr("Kilometers", "required", false);
      } else {
        listControl.value.setFormFieldAttr(
          "Kilometers",
          "required",
          true && record.WorkRequestType == "Service"
        );
      }
      break;
    default:
      break;
  }
  if (route.query.id !== undefined) {
    let currQuery = { ...route.query };
    listControl.value.selectData({ _id: currQuery.id });
    delete currQuery["id"];
    router.replace({ path: route.path, query: currQuery });
  }
}
function onChangeDimension(field, v1, v2, item) {
  switch (field) {
    case "Site":
      item.Name = "";
      item.EquipmentNo = "";
      break;
    case "CC":
      item.EquipmentType = "";
      item.EquipmentNo = "";
      break;
    default:
      break;
  }
}
function getApproval(record) {
  if (record.Status === "DRAFT") {
    return true;
  }
  axios
    .post("/fico/approvallog/get", {
      ID: record._id,
    })
    .then(
      (r) => {
        let ttd = [];
        ttd = r.data.map((d) => {
          return {
            name: d.Date
              ? d.Text.split(d.Status == "APPROVED" ? " By " : " from ")[1]
              : d.Status == "APPROVED"
              ? d.Text.split(d.Status == "APPROVED" ? " By " : " from ")[1]
              : "",
            date: d.Date ? moment(d.Date).format("DD-MMM-yyyy hh:mm:ss") : "",
          };
        });
        data.listTTD = ttd;
      },
      (e) => util.showError(e)
    );
}
onMounted(() => {});
</script>
