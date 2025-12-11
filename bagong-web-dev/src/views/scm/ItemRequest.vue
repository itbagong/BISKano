<template>
  <div class="w-full">
    <data-list
      v-show="data.isPreview == false"
      class="card"
      ref="listControl"
      :title="data.titleForm"
      form-hide-submit
      grid-config="/scm/item/request/gridconfig"
      form-config="/scm/item/request/formconfig"
      grid-read="/scm/item/request/gets-v1"
      form-read="/scm/item/request/get"
      grid-mode="grid"
      grid-delete="/scm/item/request/delete"
      form-keep-label
      form-insert="/scm/item/request/save"
      form-update="/scm/item/request/save"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      grid-hide-select
      :form-fields="[
        'TrxDate',
        'Dimension',
        'InventDimTo',
        'JournalTypeID',
        'PostingProfileID',
        'Requestor',
        'Department',
      ]"
      :grid-fields="['WarehouseID', 'Status', 'Priority', 'Approvers']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :form-tabs-new="['General']"
      :form-tabs-edit="['General', 'Line', 'Attachment']"
      :form-tabs-view="['General', 'Line', 'Attachment']"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
      @alter-grid-config="alterGridConfig"
      @alter-form-config="alterFormConfig"
      :stay-on-form-after-save="data.stayOnForm"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="customFilter"
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
          label="Request  Date From"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Request Date To"
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
          ref="refsite"
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

      <template #form_tab_Line="{ item }">
        <ItemRequestLine
          ref="lineConfig"
          :key="data.lineKey"
          v-model="data.lineRecords"
          :general-record="item"
          :approval="data.approval"
          @get-avail-stock="() => onGetStock(item)"
        ></ItemRequestLine>
      </template>
      <template #form_tab_Attachment="{ item }">
         <!-- :tags="[`IR_${item._id}`, item.WOReff ? `WO_${item.WOReff}`:'']" -->
        <s-grid-attachment
          :journal-id="item._id"
          :tags="tagsAttacment"
          :isUpdateTags="reffTags.length > 0 ? true : false"
          :add-Tags="[`IR_${item._id}`]"
          :reff-tags="reffTags"
          journal-type="Item Request"
          ref="gridAttachment"
          v-model="item.Attachment"
          @pre-Save="preSaveAttachment"
        ></s-grid-attachment>
      </template>
      <template #grid_WarehouseID="{ item }">
        {{ item.InventDimTo.WarehouseID }}
      </template>
      <template #grid_Priority="{ item }">
        {{ item.Priority }}
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #grid_Approvers="{ item }">
        <list-approvers :approvers="item.Approvers" />
      </template>
      <template #form_input_TrxDate="{ item }">
        <s-input
          label="Request Date"
          kind="date"
          v-model="item.TrxDate"
          class="w-full"
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              item.Status
            ) || item.WOReff != ''
          "
        ></s-input>
      </template>
      <template #form_input_Requestor="{ item, config }">
        <s-input
          :key="data.keyRequestor"
          label="Requestor"
          v-model="item.Requestor"
          class="w-full"
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              item.Status
            ) || item.WOReff != ''
          "
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :required="true"
          :keep-error-section="true"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onRequestorChange(v1, item);
            }
          "
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilder(search, config, item.Requestor, item)
          "
        ></s-input>
      </template>
      <template #form_input_Department="{ item }">
        <s-input
          label="Department"
          v-model="item.Department"
          class="w-full"
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              item.Status
            ) || data.disableDepartment
          "
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=DME`"
          lookup-key="_id"
          :lookup-labels="['Name']"
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
          :required="true"
          read-only
          :keepErrorSection="true"
          use-list
          :lookup-url="`/scm/item/request/journal/type/find`"
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
          read-only
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
          ref="FinancialDimension"
          :key="data.keyDimension"
          v-model="item.Dimension"
          sectionTitle="Financial Dimension"
          :default-list="profile.Dimension"
          :readOnly="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          @change="
            (field, v1, v2) => {
              onChangeDimension(field, v1, v2, item);
            }
          "
        ></dimension-editor>
      </template>
      <template #form_input_InventDimTo="{ item }">
        <dimension-invent-jurnal
          :key="data.keyDimension"
          ref="InventDimControl"
          v-model="item.InventDimTo"
          :defaultList="profile.Dimension"
          :hide-field="[
            'BatchID',
            'SerialNumber',
            'SpecID',
            'InventDimID',
            'VariantID',
            'Size',
            'Grade',
          ]"
          :site="
            item.Dimension &&
            item.Dimension.find((_dim) => _dim.Key === 'Site') &&
            item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] != ''
              ? item.Dimension.find((_dim) => _dim.Key === 'Site')['Value']
              : undefined
          "
          title-header="Inventory Dimension"
          :mandatory="['WarehouseID']"
          :readOnly="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
        ></dimension-invent-jurnal>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          class="bg-transparent hover:bg-blue-500 hover:text-black"
          label="Preview"
          icon="eye-outline"
          @click="onPreview"
        ></s-button>
        <s-button
          v-if="['POSTED'].includes(item.Status)"
          icon="open-in-new"
          label="References"
          class="btn_warning refresh_btn"
          :tooltip="`References No`"
          @click="getReffNo"
        />
        <s-button
          v-if="
            ['', 'DRAFT'].includes(item.Status) ||
            (['READY'].includes(item.Status) &&
              data.approval.Postingers == true)
          "
          icon="content-save"
          :disabled="data.disabledFormButton"
          class="btn_primary s-button"
          label="Save"
          @click="onSave(item)"
        />

        <s-button
          v-if="
            ['READY'].includes(item.Status) && data.approval.Postingers == true
          "
          icon="swap-horizontal"
          class="btn_warning s-button"
          label="Fulfillment"
          @click="selectFulfillment"
        />
        <FormButtonsTrx
          :key="data.btnTrx"
          :disabled="data.disabledFormButton"
          :posting-profile-id="item.PostingProfileID"
          :status="item.Status"
          :journalId="item._id"
          journal-type-id="Item Request"
          :autoPost="false"
          :autoReopen="false"
          moduleid="scm/new"
          @pre-submit="preSubmit"
          @post-submit="postSubmit"
          @pre-reopen="preReopen"
          @error-submit="errorSubmit"
          @approvalSource="approvalSource"
        >
        </FormButtonsTrx>
      </template>
      <template #form_footer_1="{ item }">
        <RejectionMessageList
          ref="listRejectionMessage"
          journalType="Item Request"
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

    <div v-if="data.isPreview == true" class="w-full">
      <PreviewReport
        class="card w-full"
        title="Preview"
        :preview="data.preview"
        @close="closePreview"
        SourceType="Item Request"
        :SourceJournalID="data.record._id"
        :hideSignature="false"
      >
      </PreviewReport>
    </div>
    <s-modal
      :display="data.isDialogFulfillment"
      ref="deleteModal"
      title="Fulfillment"
      class="w-1/2"
      @beforeHide="data.isDialogFulfillment = false"
    >
      <div
        :class="`min-w-[700px]  ${
          ['Item Transfer', 'Movement Out'].includes(data.fulfillmentType)
            ? 'min-h-[300px]'
            : 'min-h-[200px]'
        }`"
      >
        <s-card hide-title>
          <div>
            <s-input
              v-model="data.fulfillmentType"
              label="Fulfillment type"
              class="w-full"
              use-list
              :required="true"
              :keepErrorSection="true"
              :items="data.listFulfillmentType"
              @change="onChangeType"
            ></s-input>
          </div>
          <div
            v-if="
              ['Item Transfer', 'Movement Out'].includes(data.fulfillmentType)
            "
            class="h-[500px] overflow-auto"
          >
            <s-grid
              v-model="data.listfulFillment"
              ref="gridDailyRptControl"
              class="w-full grid-line-items"
              hide-search
              hide-sort
              hide-new-button
              hide-delete-button
              hide-refresh-button
              hide-detail
              hide-action
              hide-select
              hide-control
              hide-footer
              auto-commit-line
              no-confirm-delete
              form-keep-label
              hide-paging
              :config="data.gridCfgFulfillment"
            >
              <template #item_ItemID="{ item, idx }">
                <s-input-sku-item
                  v-model="item.ItemVarian"
                  :key="data.keyRefItem"
                  :record="item"
                  :lookup-url="`/tenant/item/gets-detail?_id=${item.ItemVarian}`"
                  :disabled="true"
                ></s-input-sku-item>
              </template>
              <template #item_WarehouseID="{ item, idx }">
                <s-input
                  v-model="item.WarehouseID"
                  hide-label
                  label="From warehouse"
                  class="w-full"
                  :disabled="false"
                  :use-list="true"
                  :lookup-url="`/scm/item/balance/get-available-warehouse?ItemID=${item.ItemID}&SKU=${item.SKU}&ItemRequestID=${data.record._id}&FulfillmentType=${data.fulfillmentType}`"
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
            </s-grid>
          </div>
        </s-card>
      </div>
      <template #buttons="{ item }">
        <s-button
          class="w-[50px] btn_primary text-white font-bold flex justify-center"
          label="Apply"
          @click="confirmFulfillment"
        ></s-button>
        <s-button
          class="w-[50px] btn_warning text-white font-bold flex justify-center"
          label="CANCEL"
          @click="data.isDialogFulfillment = false"
        ></s-button>
      </template>
    </s-modal>
    <s-modal
      :display="data.isDialogReffNo"
      ref="refModal"
      title="List Ref No."
      @before-hide="data.isDialogReffNo = false"
    >
      <div :class="`min-w-[500px]`">
        <s-card hide-title>
          <div class="max-h-[500px] overflow-auto">
            <s-grid
              v-model="data.listReffNo"
              class="w-full grid-line-items"
              hide-search
              hide-sort
              hide-new-button
              hide-delete-button
              hide-refresh-button
              hide-detail
              hide-action
              hide-select
              hide-control
              auto-commit-line
              no-confirm-delete
              form-keep-label
              hide-paging
              :config="data.gridCfgReffNo"
            >
              <template #item_JournalID="{ item, idx }">
                <a
                  href="#"
                  class="text-blue-600 border-gray-400"
                  @click="redirectReff(item)"
                >
                  {{ item.ReffNo }}
                </a>
              </template>
            </s-grid>
          </div>
        </s-card>
      </div>
      <template #buttons="{ item }">
        <s-button
          class="w-[50px] btn_warning text-white font-bold flex justify-center"
          label="CANCEL"
          @click="data.isDialogReffNo = false"
        ></s-button>
      </template>
    </s-modal>
  </div>
</template>

<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { layoutStore } from "@/stores/layout.js";
import {
  DataList,
  util,
  SForm,
  SInput,
  SButton,
  loadFormConfig,
  SModal,
  SGrid,
} from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import ItemRequestLine from "./widget/ItemRequestLine.vue";
import RejectionMessageList from "./widget/RejectionMessageList.vue";
import StatusText from "@/components/common/StatusText.vue";
import SingleAttachment from "./widget/SingleAttachment.vue";
import SInputSkuItem from "./widget/SInputSkuItem.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import lodash from "lodash";
import PreviewReport from "@/components/common/PreviewReport.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import ListApprovers from "@/components/common/ListApprovers.vue";
import moment from "moment";
import { authStore } from "@/stores/auth.js";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";

const featureID = "ItemRequest";

// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'ItemRequest'})

const profile = authStore().getRBAC(featureID);
const headOffice = layoutStore().headOfficeID;
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const route = useRoute();
const router = useRouter();

const listControl = ref(null);
const lineConfig = ref(null);
const InventDimControl = ref(null);
const listRejectionMessage = ref(null);
const gridAttachment = ref(null);
const FinancialDimension = ref(null);
let customFilter = computed(() => {
  let filters = [
    {
      Field: "Keyword",
      Op: "$eq",
      Value: data.search.Text,
    },
  ];
  let filtersIR = [];
  if (data.search.Text !== null && data.search.Text !== "") {
    filtersIR.push(
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
        Field: "WOReff",
        Op: "$contains",
        Value: [data.search.Text],
      },
      {
        Field: "Priority",
        Op: "$contains",
        Value: [data.search.Text],
      }
    );
  }

  if (data.search.requestor !== null && data.search.requestor !== "") {
    filters.push({
      Field: "Requestor",
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
      items: filtersIR,
    },
  ];
  if (data.search.Text !== null && data.search.Text !== "") {
    filters = [...filters, ...items];
  }

  if (filters.length == 1) return { Op: "$and", Items: filters }; //filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});
const reffTags = computed({
  get() {
    return data.record.WOReff ? ["WO_" + data.record.WOReff] : [];
  },
});
const tagsAttacment = computed({
  get() {
    let tags = [`IR_${data.record._id}`]
    if (data.record.WOReff){
      tags = [...tags, ...["WO_" + data.record.WOReff] ]
    }
    return data.record._id ? tags : [];
  },
});
const axios = inject("axios");
const data = reactive({
  isPreview: false,
  isDialogFulfillment: false,
  isDialogReffNo: false,
  appMode: "grid",
  formMode: "edit",
  titleForm: "Item Request",
  btnTrx: util.uuid(),
  lineKey: util.uuid(),
  keyRequestor: util.uuid(),
  keyDimension: util.uuid(),
  keyRefItem: "",
  listReffNo: [],
  menus: [
    [
      { label: "Item Request", emit: "ItemRequest" },
      { label: "Purchase Request", emit: "PurchaseRequest" },
      { label: "Item Transfer", emit: "ItemTransfer" },
    ],
  ],
  menusLine: [
    [
      { label: "Item Request", emit: "ItemRequest" },
      { label: "Purchase Request", emit: "LinePurchaseRequest" },
      { label: "Item Transfer", emit: "LineItemTransfer" },
    ],
  ],
  record: {
    _id: "",
    DocumentDate: new Date(),
    RequestDate: new Date(),
    TrxDate: new Date(),
    Dimension: [],
    Status: "",
  },
  preview: [],
  lineRecords: [],
  listfulFillment: [],
  approval: {
    Approvers: false,
    Postingers: false,
    Submitters: false,
  },
  journalTypeData: {},
  gridCfgFulfillment: {},
  gridCfgReffNo: {},
  stayOnForm: true,
  disableDepartment: false,
  disabledFormButton: false,
  keyJournalType: util.uuid(),
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
  fulfillmentType: "",
  listFulfillmentType: [
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
});

function getRequestor(record) {
  axios.post("/bagong/employee/get", [record.Requestor]).then(
    (r) => {
      data.requestorName = r.data.Name;
    },
    (e) => util.showError(e)
  );
}

function newRecord(record) {
  data.formMode = "new";
  data.titleForm = `Create New Item Request`;
  data.lineRecords = [];
  record._id = "";
  record.DocumentDate = new Date();
  record.RequestDate = new Date();
  record.TrxDate = new Date();
  record.InventDimTo = {};
  record.Dimension = [];
  record.Status = "";
  record.WOReff = "";
  data.record = record;
  getPostingProfile(record);
  getDetailEmployee("", data.record);
  openForm(record);
}

function editRecord(record) {
  data.formMode = "edit";
  data.titleForm = `Edit Item Request | ${record._id}`;
  data.record = record;
  // getLines(record);
  openForm(record);
  util.nextTickN(2, () => {
    if (["SUBMITTED", "READY", "REJECTED", "POSTED"].includes(record.Status)) {
      if (record.Status === "REJECTED") {
        listRejectionMessage.value.fetchRecord(record._id);
      }
      setFormMode("view");
    }
    if (record.WOReff) {
      getPostingProfile(record);
    }
    data.lineKey = util.uuid();
  });
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr(
      "_id",
      "hide",
      data.formMode == "new" ? true : false
    );
    const el = document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(1) > div > div > div:nth-child(1)"
    );
    if (record._id == "") {
      el.style.display = "none";
    } else {
      el.style.display = "block";
    }
    data.btnTrx = util.uuid();
  });
}
function onFormFieldChange(name, v1, v2, old, record) {
  if (route.query.id !== undefined) {
    let currQuery = { ...route.query };
    listControl.value.selectData({ _id: currQuery.id }); //remark sementara tunggu suimjs update
    delete currQuery["id"];
    router.replace({ path: route.path, query: currQuery });
  }
}

function getDetailEmployee(_id, record) {
  let payload = [];
  if (_id) {
    payload = [_id];
  }
  record.Requestor = _id;
  if (!_id) {
    axios.post("/tenant/employee/get-emp-warehouse", payload).then(
      (r) => {
        record.Requestor = r.data._id;
        data.keyRequestor = util.uuid();
      },
      (e) => util.showError(e)
    );
  }
}
function onFieldChanged(val, record) {
  if (val.WarehouseID) {
    record.Requestor = "";
    axios.post("/tenant/warehouse/get", [val.WarehouseID]).then((r) => {
      if (r.data) {
        record.Requestor = r.data.PIC;
      }
    });
  }
}
function alterGridConfig(cfg) {
  cfg.sortable = ["LastUpdate", "Created", "TrxDate", "_id"];
  cfg.setting.idField = "Created";
  cfg.setting.sortable = ["LastUpdate", "Created", "TrxDate", "_id"];

  const warehousField = {
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
  };
  cfg.fields.splice(6, 0, warehousField);
}
function alterFormConfig(config) {
  if (route.query.id !== undefined) {
    let getUrlParam = route.query.id;
    listControl.value.selectData({ _id: getUrlParam }); //remark sementara tunggu suimjs update
    router.replace({
      path: `/scm/ItemRequest`,
    });
  }
}
function onRequestorChange(v1, record) {
  if (v1) {
    // axios.post("/bagong/employee/get", [v1]).then((r) => {
    //   record.Department = r.data.Detail.Department;
    //   if (record.Department != "") {
    //     data.disableDepartment = true;
    //   } else {
    //     data.disableDepartment = false;
    //   }
    // });
    getDetailEmployee(v1, record);
  }
}
function setFormMode(mode) {
  listControl.value.setFormMode(mode);
}
function setLoadingForm(loading) {
  data.disabledFormButton = loading;
  listControl.value.setFormLoading(loading);
}
function getJournalType(_id, action) {
  axios
    .post("/scm/item/request/journal/type/find?_id=" + _id, { sort: ["-_id"] })
    .then(
      (r) => {
        data.journalTypeData = r.data[0];
        setDefaultRecordFromJournalType(r.data[0], action);
      },
      (e) => util.showError(e)
    );
}
function getPostingProfile(record) {
  util.nextTickN(2, () => {
    axios.post(`/scm/item/request/journal/type/find`).then(
      (r) => {
        if (r.data.length > 0) {
          data.keyJournalType = util.uuid();
          record.JournalTypeID = r.data[0]._id;
          record.PostingProfileID = r.data[0].PostingProfileID;
        }
      },
      (e) => util.showError(e)
    );
  });
}
function setDefaultRecordFromJournalType(JT, action) {
  if (["", "DRAFT"].includes(data.record.Status)) {
    if (action === "change") {
      data.record.PostingProfileID = JT.PostingProfileID;
    }
  }
}
function saveLine(record, doSubmit = undefined) {
  let line = data.lineRecords;
  let payload = {
    ItemRequestID: record._id,
    ItemRequestDetails: line.map((o) => {
      const newLine = o;
      newLine.ItemRequestID = record._id;
      newLine.WarehouseID = record.InventDimTo.WarehouseID;
      newLine.Dimension = record.Dimension;
      return newLine;
    }),
  };
  setLoadingForm(true);
  axios
    .post("/scm/item/request/detail/save-multiple", payload)
    .then(
      (r) => {
        if (doSubmit) {
          doSubmit();
        } else {
          util.nextTickN(2, () => {
            if (["DRAFT", "SUBMITTED", ""].includes(record.Status)) {
              // listControl.value.setControlMode("grid");
              listControl.value.refreshForm();
              setFormMode("edit");
            } else {
              listControl.value.refreshList();
            }

            setLoadingForm(false);
            return util.showInfo("item request has been successful save");
          });
        }
      },
      (e) => {
        setLoadingForm(false);
        return util.showError(e);
      }
    )
    .finally(function () {
      setLoadingForm(false);
    });
}

function onSave(record) {
  let validate = true;
  const pc = record.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = record.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = record.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;

  if (InventDimControl.value) {
    validate = InventDimControl.value.validate();
  }
  // if (!record.Requestor) {
  //   return util.showError("General Requestor is required");
  // }

  if (!record.InventDimTo?.WarehouseID) {
    return util.showError("General Warehouse ID is required");
  }

  if (FinancialDimension.value) {
    validate = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      validate = false;
    }
  }

  const payload = JSON.parse(JSON.stringify(record));
  if (record.Status == "") {
    payload.Status = "DRAFT";
  }
  if (validate && listControl.value.formValidate()) {
    setLoadingForm(true);
    payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
    axios.post("/scm/item/request/save", payload).then(
      (r) => {
        data.record = r.data;
        listControl.value.setFormRecord(r.data);
        data.titleForm = `Edit Item Request | ${r.data._id}`;
        util.nextTickN(2, () => {
          if (gridAttachment.value) {
            gridAttachment.value.Save();
          }
          Promise.all([saveLine(r.data)]).then(() => {});
        });
      },
      (e) => {
        setLoadingForm(false);
        return util.showError(e);
      }
    );
  } else {
    return util.showError("field is required");
  }
}

function preSubmit(status, action, doSubmit) {
  const dim = data.record.InventDimTo;
  if (status == "DRAFT") {
    setLoadingForm(true);
    if (!InventDimControl.value) {
      if (
        !data.record.Name ||
        !data.record.PostingProfileID ||
        !data.record.JournalTypeID
      ) {
        return util.showError(
          "General field Name, JournalTypeID, PostingProfileID  is required"
        );
      }
      if (!dim.WarehouseID) {
        return util.showError("General Warehouse ID is required");
      }
    }

    if (data.lineRecords.length == 0) {
      setLoadingForm(false);
      return util.showError("Line items is empty");
    }
    if (data.lineRecords.find((o) => o.QtyRequested === 0)) {
      setLoadingForm(false);
      return util.showError("Qty requested cannot be 0");
    }
    if (data.lineRecords.find((o) => o.ItemID === "")) {
      setLoadingForm(false);
      return util.showError("Line Item requested");
    }

    util.nextTickN(2, () => {
      const valid = listControl.value.formValidate();
      if (valid) {
        onGetStock(
          data.record,
          () => {
            let payload = JSON.parse(JSON.stringify(data.record));
            if (action === "Reject") {
              payload.Status = "DRAFT";
            }
            setLoadingForm(true);
            payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
            axios.post("/scm/item/request/save", payload).then(
              (r) => {
                util.nextTickN(2, () => {
                  if (gridAttachment.value) {
                    gridAttachment.value.Save();
                  }
                  saveLine(r.data, doSubmit);
                });
              },
              (e) => {
                setLoadingForm(false);
                return util.showError(e);
              }
            );
          },
          () => {
            setLoadingForm(false);
          }
        );
      } else {
        setLoadingForm(false);
        return util.showError("Please check required field");
      }
      setFormRequired(false);
    });
  } else if (status == "READY" && action !== "Reject") {
    if (data.lineRecords.find((o) => o.Complete === false)) {
      return util.showError("Fulfillment is not completed yet");
    }
    const checkRemarks = data.lineRecords.find(
      (o) => o.Remarks === "" && o.QtyFulfilled < o.QtyRequested
    );
    if (checkRemarks) {
      return util.showError(
        "Remarks is required if Qty Fulfilled less than Requested"
      );
    }
    saveLine(data.record, doSubmit);
  } else {
    doSubmit();
  }
}
function postSubmit(r, action) {
  axios.post("/scm/item/request/detail/post-submit", { ItemRequestID: data.record._id })
    .then((r)=>{
      listControl.value.refreshForm();
      setLoadingForm(false);
      util.nextTickN(2, () => {
        data.stayOnForm = false;
        listControl.value.setControlMode("grid");
        listControl.value.refreshList();
      });
    },(e)=>{
      return util.showError(e);
    })

  // if (action == "Submit"){
    
  // }else{
  //     listControl.value.refreshForm();
  //     setLoadingForm(false);
  //     util.nextTickN(2, () => {
  //       data.stayOnForm = false;
  //       listControl.value.setControlMode("grid");
  //       listControl.value.refreshList();
  //     });
  // }
}

function preReopen(status, doSubmit) {
  let payload = JSON.parse(JSON.stringify(data.record));
  payload.Status = "DRAFT";
  axios.post("/scm/item/request/save", payload).then(
    (r) => {
      util.nextTickN(2, () => {
        data.stayOnForm = false;
        listControl.value.setControlMode("grid");
        listControl.value.refreshList();
      });
    },
    (e) => {
      return util.showError(e);
    }
  );
}

function errorSubmit(e, action) {
  setLoadingForm(false);
}
function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    if (
      ![
        "RequestType",
        "WOReff",
        "PostingProfileID",
        "ReffNo",
        "Note",
        "Remarks",
      ].includes(e.field)
    ) {
      listControl.value.setFormFieldAttr(e.field, "required", required);
    }
  });
}
const multiGroupBy = (seq, keys) => {
  if (!keys.length) return seq;
  var first = keys[0];
  var rest = keys.slice(1);
  return lodash.mapValues(lodash.groupBy(seq, first), function (value) {
    return multiGroupBy(value, rest);
  });
};
function onProcess(record) {
  util.nextTickN(2, () => {
    let newLines = [];
    data.lineRecords.map((line) => {
      line.DetailLines.map((dline) => {
        let obj = {
          ...dline,
          ...line,
        };
        newLines.push(obj);
      });
    });
    const groupedItems = multiGroupBy(newLines, [
      "FulfillmentType",
      "WarehouseID",
    ]);
    const linerecs = [];
    for (const item in groupedItems) {
      for (const wr in groupedItems[item]) {
        linerecs.push({
          FulfillmentType: item,
          WarehouseID: wr,
          Lines: groupedItems[item][wr],
        });
      }
    }
    linerecs.map((item) => {
      process2Trx(item);
    });
    util.nextTickN(2, () => {
      data.appMode = "grid";
      listControl.value.refreshList();
      listControl.value.refreshForm();
      listControl.value.setControlMode("grid");
    });
  });
}
function process2Trx(lineRec) {
  let url = "";
  let payload = {
    Name: data.record.Name,
    Dimension: data.record.Dimension,
    InventDim: {
      WarehouseID: lineRec.WarehouseID,
    },
    Lines: lineRec.Lines.map((o, i) => {
      return {
        LineNo: i + 1,
        ItemID: o.ItemID,
        SKU: o.SKU,
        Text: o.Description,
        Qty: o.QtyFulfilled,
        UnitID: o.UoM,
        Dimension: data.record.Dimension,
        InventDim: {
          WarehouseID: lineRec.WarehouseID,
        },
      };
    }),
  };
  switch (lineRec.FulfillmentType) {
    case "Movement In":
      url = `/scm/inventory/journal/save`;
      payload = {
        ...payload,
        TrxType: "Movement In",
        InventDimTo: {
          WarehouseID: lineRec.WarehouseID,
        },
      };
      break;
    case "Movement Out":
      url = `/scm/inventory/journal/save`;
      payload = {
        ...payload,
        TrxType: "Movement Out",
        InventDim: {
          WarehouseID: lineRec.WarehouseID,
        },
      };
      break;
    case "Item Transfer":
      url = `/scm/inventory/journal/save`;
      payload = {
        ...payload,
        TrxType: "Transfer",
        InventDim: {
          WarehouseID: lineRec.WarehouseID,
        },
      };
      break;
    case "Purchase Request":
      url = `/scm/purchaserequest/save`;
      break;
    case "Assembly":
      url = `/mfg/work/order/save`;
      break;

    default:
      break;
  }
  if (url) {
    util.nextTickN(2, () => {
      axios.post(url, payload).then((r) => {
        switch (lineRec.FulfillmentType) {
          case "Purchase Request":
            const payloadLinePR = {
              PurchaseRequestID: r.data._id,
              PurchaseRequestDetails: payload.Lines,
            };

            axios
              .post("/scm/purchaserequest/detail/save-multiple", payloadLinePR)
              .then(
                (r) => {},
                (e) => {
                  return util.showError(e);
                }
              );
            break;
          case "Assembly":
            const payloadLineWO = {
              WorkOrderID: r.data._id,
              WorkOrderDetails: payload.Lines,
            };

            axios
              .post("/mfg/work/order/detail/save-multiple", payloadLineWO)
              .then(
                (r) => {},
                (e) => {
                  return util.showError(e);
                }
              );
            break;

          default:
            break;
        }
      });
    });
  }
  return;
}
function onGetStock(record, cbOK, cbFalse) {
  if (data.lineRecords.length > 0) {
    const payload = data.lineRecords.map((item) => {
      item.InventDim = data.record.InventDimTo;
      return item;
    });
    axios.post("/scm/item/balance/get-qty", payload).then(
      (r) => {
        data.lineRecords.map((o) => {
          o.QtyAvailable =
            r.data.find(
              (item) => item.ItemID === o.ItemID && item.SKU === o.SKU
            )?.Qty ?? 0;
          return o;
        });
        cbOK();
      },
      (e) => {
        cbFalse();
        return util.showError(e);
      }
    );
  }
}
function onControlModeChanged(mode) {
  data.appMode = mode;
  if (mode === "grid") {
    data.titleForm = "Item Request";
  }
}
function preSaveAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [`IR_${data.record._id}`];
    return asset;
  });
}

function getReffNo() {
  data.listReffNo = [];
  listControl.value.setFormLoading(true);
  axios
    .post("/scm/item/request/detail/reference", {
      ItemRequestID: data.record._id,
    })
    .then(
      (r) => {
        listControl.value.setFormLoading(false);
        data.isDialogReffNo = true;
        data.listReffNo = r.data.Data;
      },
      (e) => {
        listControl.value.setFormLoading(false);
        util.showError(e);
      }
    );
}

function onChangeDimension(field, v1, v2, item) {
  switch (field) {
    case "Site":
      item.Requestor = "";
      break;
    default:
      break;
  }
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
  if (Site.length > 0 && !item.WOReff) {
    qp.Where = {
      Op: "$and",
      items: querySite,
    };
  }
  if (profileSite.length == 0 || profileSite.includes(headOffice)) {
    delete qp.Where;
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
    if (Site.length > 0 && !item.WOReff) {
      items = [...items, ...querySite];
    }
    if (profileSite.length == 0 || profileSite.includes(headOffice)) {
      items = [
        {
          Op: "$or",
          items: [
            { Field: "_id", Op: "$contains", Value: [search] },
            { Field: "Name", Op: "$contains", Value: [search] },
          ],
        },
      ];
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
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
function onPreview() {
  getRequestor(data.record);
  data.isPreview = true;
}

function closePreview() {
  data.isPreview = false;
}
function approvalSource(approval) {
  console.log(approval);
  data.approval = approval;
}

function onChangeType(field, v1) {
  const fulFillment = JSON.parse(JSON.stringify(data.lineRecords))
    .filter((f) => {
      return f.isSelected == true;
    })
    .map((dt) => {
      // dt.FulfillmentType = data.fulfillmentType;
      // dt.ItemVarian = helper.ItemVarian(dt.ItemID, dt.SKU);
      // dt.WarehouseID = "";
      // dt.QtyFulfilled = dt.QtyRequested;
      // dt.InventDimFrom = {};
      return {
        _id: dt._id,
        ItemID: dt.ItemID,
        ItemID: dt.ItemID,
        SKU: dt.SKU,
        FulfillmentType: v1,
        ItemVarian: helper.ItemVarian(dt.ItemID, dt.SKU),
        WarehouseID: "",
        QtyFulfilled: dt.QtyRequested,
        InventDimFrom: {},
        QtyAvailable: 0,
      };
    });
  if (fulFillment.length > 0) {
    data.listfulFillment = fulFillment;
    data.lineRecords
      .filter((f) => {
        return f.isSelected == true;
      })
      .map((dt) => {
        dt.DetailLines = fulFillment.filter((r) => {
          return r._id === dt._id;
        });
        return dt;
      });
    data.keyRefItem = util.uuid();
  }
}
function selectFulfillment() {
  data.fulfillmentType = "";
  data.listfulFillment = [];
  const fulFillment = JSON.parse(JSON.stringify(data.lineRecords))
    .filter((f) => {
      return f.isSelected == true;
    })
    .map((dt) => {
      return {
        _id: dt._id,
        ItemID: dt.ItemID,
        SKU: dt.SKU,
        FulfillmentType: data.fulfillmentType,
        ItemVarian: helper.ItemVarian(dt.ItemID, dt.SKU),
        WarehouseID: "",
        QtyFulfilled: dt.QtyRequested,
        InventDimFrom: {},
        QtyAvailable: 0,
      };
    });
  if (fulFillment.length > 0) {
    data.listfulFillment = fulFillment;
    data.lineRecords
      .filter((f) => {
        return f.isSelected == true;
      })
      .map((dt) => {
        dt.DetailLines = fulFillment.filter((r) => {
          return r._id === dt._id;
        });
        return dt;
      });

    data.isDialogFulfillment = true;
    data.keyRefItem = util.uuid();
  } else {
    util.showError("select fulfillment");
  }
}
async function confirmFulfillment() {
  data.lineRecords.map((v) => {
    const fulFill = data.listfulFillment.find((f) => {
      return f._id == v._id;
    });
    if (fulFill) {
      v.Complete = true;
      v.QtyFulfilled = v.QtyRequested;
      v.DetailLines.map((l) => {
        l.UoM = v.UoM;
      });
    }
    return v;
  });

  for (let i = 0; i < data.lineRecords.length; i++) {
    await axios.post(`/scm/item/request/detail/save`, data.lineRecords[i]).then(
      (r) => {
        return r.data.data;
      },
      (e) => {
        return [];
      }
    );
  }
  lineConfig.value.getLines(data.record);
  data.isDialogFulfillment = false;
}
function onGetsAvailableWarehouse(_id, item) {
  axios
    .post(
      `/scm/item/balance/get-available-warehouse?ItemID=${item.ItemID}&SKU=${item.SKU}&ItemRequestID=${data.record._id}&FulfillmentType=${data.fulfillmentType}`
    )
    .then(
      (r) => {
        const wh = r.data.find(function (v) {
          return v._id == _id;
        });
        if (wh) {
          item.FulfillmentType = data.fulfillmentType;
          item.InventDimFrom = wh.InventDim;
          item.QtyAvailable = wh.Qty;
        }
      },
      (e) => {
        return util.showError(e);
      }
    );
}
function generateGridCfg(colum) {
  let addColm = [];
  for (let index = 0; index < colum.length; index++) {
    addColm.push({
      field: colum[index].field,
      kind: colum[index].kind,
      label: colum[index].label,
      readType: "show",
      labelField: "",
      width: colum[index].width,
      readOnly: colum[index].readOnly,
      input: {
        field: colum[index].field,
        label: colum[index].label,
        hint: "",
        hide: false,
        placeHolder: colum[index].label,
        kind: colum[index].kind,
        width: colum[index].width,
        readOnly: colum[index].readOnly,
      },
    });
  }
  return {
    fields: addColm,
    setting: {
      idField: "_id",
      keywordFields: ["_id", "Name"],
      sortable: ["_id"],
    },
  };
}
function createGridCfgFulfillment(load = false) {
  const colum = [
    {
      field: "ItemID",
      kind: "text",
      label: "Item",
      width: "200px",
      readOnly: true,
    },
    {
      field: "WarehouseID",
      kind: "text",
      label: "From warehouse",
      width: "150px",
      readOnly: true,
    },
  ];
  util.nextTickN(2, () => {
    data.gridCfgFulfillment = generateGridCfg(colum);
  });
}
function redirectReff(item) {
  let page = "";
  let query = {};
  if (item.JournalType == "Item Request") {
    page = "scm-ItemRequest";
    query = { id: item.ReffNo };
  } else if (item.JournalType == "Movement In") {
    page = "scm-InventoryJournal";
    query = {
      id: item.ReffNo,
      type: "Movement In",
      title: "Movement In",
    };
  } else if (item.JournalType == "Movement Out") {
    page = "scm-InventoryJournal";
    query = {
      id: item.ReffNo,
      type: "Movement Out",
      title: "Movement Out",
    };
  } else if (item.JournalType == "Transfer") {
    page = "scm-InventoryJournal";
    query = {
      id: item.ReffNo,
      type: "Transfer",
      title: "Item Transfer",
    };
  } else if (item.JournalType == "Purchase Request") {
    page = "scm-PurchaseRequest";
    query = { id: item.ReffNo };
  } else if (item.JournalType == "Purchase Order") {
    page = "scm-PurchaseOrder";
    query = { id: item.ReffNo };
  } else if (item.JournalType == "Inventory Receive") {
    page = "scm-InventTrx";
    query = {
      id: item.ReffNo,
      type: "Inventory Receive",
      title: "Inventory Receive",
    };
  } else if (item.JournalType == "Inventory Issuance") {
    page = "scm-InventTrx";
    query = {
      id: item.ReffNo,
      type: "Inventory Issuance",
      title: "Inventory Issuance",
    };
  }
  const url = router.resolve({
    name: page,
    query: query,
  });
  window.open(url.href, "_blank");
}
function createGridCfgRefNo(load = false) {
  const colum = [
    {
      field: "JournalType",
      kind: "text",
      label: "Journal Type",
      width: "200px",
      readOnly: true,
    },
    {
      field: "JournalID",
      kind: "text",
      label: "Journal Ref",
      width: "150px",
      readOnly: true,
    },
  ];
  util.nextTickN(2, () => {
    data.gridCfgReffNo = generateGridCfg(colum);
  });
}
onMounted(() => {
  createGridCfgFulfillment();
  createGridCfgRefNo();
  if (route.query.trxid !== undefined || route.query.id !== undefined) {
    let getUrlParam = route.query.trxid || route.query.id;
    axios
      .post(`/scm/item/request/get`, [getUrlParam])
      .then(
        (r) => {
          editRecord(r.data);
        },
        (e) => util.showError(e)
      )
      .finally(() => {
        util.nextTickN(2, () => {
          router.replace({
            path: `/scm/ItemRequest`,
          });
        });
      });
  }
});
</script>
