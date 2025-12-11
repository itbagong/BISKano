<template>
  <div class="w-full">
    <data-list
      v-show="data.appMode == 'grid'"
      id="pr-data-list"
      class="card"
      ref="listControl"
      :title="data.titleForm"
      :form-hide-submit="true"
      grid-config="/scm/purchase/request/gridconfig"
      form-config="/scm/purchase/request/formconfig"
      :grid-read="`/scm/purchase/request/gets-v1`"
      form-read="/scm/purchase/request/get"
      grid-mode="grid"
      grid-delete="/scm/purchase/request/delete"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      grid-hide-select
      form-keep-label
      :form-insert="data.formInsert"
      :form-update="data.formUpdate"
      :grid-fields="['Status', 'WarehouseID', 'Approvers']"
      :form-fields="[
        '_id',
        'JournalTypeID',
        'PostingProfileID',
        'VendorID',
        'References',
        'TaxCodes',
        'TaxType',
        'PurchaseType',
        'PIC',
        'Dimension',
        'Location',
        'Requestor',
        'BillingName',
      ]"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :form-tabs-new="data.tabs"
      :form-tabs-edit="data.tabs"
      :form-tabs-view="data.tabs"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="customFilter"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @preSave="onPreSave"
      @postSave="onPostSave"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
      @formRecordChange="onFormRecordChange"
      @alterFormConfig="alterFormConfig"
      @alterGridConfig="alterGridConfig"
      :stay-on-form-after-save="data.stayOnForm"
    >
      <template #grid_header_search="{ config }">
        <s-input
          ref="refVendor"
          v-model="data.search.VendorID"
          lookup-key="_id"
          label="Vendor"
          class="w-full"
          multiple
          use-list
          :lookup-url="`/tenant/vendor/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refName"
          v-model="data.search.Name"
          lookup-key="_id"
          label="Text"
          class="w-[400px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Trx Date From"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Trx Date To"
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
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #grid_WarehouseID="{ item }">
        {{ item.Location.WarehouseID }}
      </template>
      <template #grid_ReffNo="{ item }">
        {{ item.ReffNo?.join() }}
      </template>
      <template #grid_Approvers="{ item }">
        <list-approvers :approvers="item.Approvers" />
      </template>
      <template #form_loader>
        <loader />
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          class="bg-transparent hover:bg-blue-500 hover:text-black"
          label="Preview"
          icon="eye-outline"
          @click="onPreview"
        ></s-button>
        <s-button
          v-if="['POSTED'].includes(item.Status) || item.ReffNo?.length > 0"
          icon="open-in-new"
          label="References"
          class="btn_warning refresh_btn"
          :tooltip="`References No`"
          @click="getReffNo"
        />
        <s-button
          v-if="
            !['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          :icon="`content-save`"
          class="btn_primary submit_btn"
          label="Save"
          :disabled="data.disableButton"
          @click="onSave(item)"
        />
        <FormButtonsTrx
          :posting-profile-id="item.PostingProfileID"
          :disabled="data.disableButton"
          :status="item.Status"
          :journalId="item._id"
          :autoReopen="false"
          journal-type-id="Purchase Request"
          moduleid="scm/new"
          @pre-submit="preSubmit"
          @pre-reopen="preReopen"
          @post-submit="postSubmit(item)"
          @error-submit="errorSubmit"
        >
        </FormButtonsTrx>
      </template>
      <template #form_tab_Line="{ item }">
        <PurchaseLine
          ref="lineConfig"
          v-model="item.Lines"
          :general-record="item"
          :defaultList="defaultList"
          purchase-type="purchase/request"
          :disable-field="data.disableField"
          :isReff="item.ReffNo?.length > 0"
        ></PurchaseLine>
      </template>
      <template #form_tab_Attachment="{ item }">
        <s-grid-attachment
          :key="item._id"
          :journal-id="item._id"
          :tags="linesTag"
          :isUpdateTags="reffTags.length > 0 ? true : false"
          :add-Tags="addTags"
          :reff-tags="reffTags"
          journal-type="Purchase Request"
          ref="gridAttachment"
          @pre-Save="preSaveAttachment"
          v-model="item.Attachment"
        ></s-grid-attachment>
      </template>
      <template #form_input_Requestor="{ item, config }">
        <s-input
          :key="data.keyRequestor"
          label="Requestor"
          v-model="item.Requestor"
          class="w-full"
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :required="true"
          :keep-error-section="true"
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilderRequestor(
                search,
                ['_id', 'Name'],
                item.Requestor,
                item
              )
          "
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
          :disabled="true"
          :keepErrorSection="true"
          use-list
          :lookup-url="`/scm/purchase/request/journal/type/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_input_VendorID="{ item }">
        <s-input
          v-if="['DRAFT', ''].includes(item.Status)"
          ref="refInput"
          v-model="item.VendorID"
          label="Vendor"
          field="VendorID"
          class="w-full"
          use-list
          lookup-key="_id"
          :lookup-url="`/bagong/vendor/get-vendor-by-site`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :lookup-payload-builder="
            (search) => {
              const qp = {};
              let SiteID = item.Dimension.find((d) => {
                return d.Key === 'Site';
              }).Value;
              qp.Take = 20;
              qp.Sort = ['_id'];
              qp.Select = ['_id', 'Name'];
              qp.Where = {
                SiteID: SiteID
                  ? SiteID
                  : defaultList.length > 0
                  ? defaultList[0]
                  : 'SITE020',
              };
              return qp;
            }
          "
          @change="
            (field, v1, v2, old, ctl) => {
              onFormFieldChange(field, v1, v2, old, item);
            }
          "
        ></s-input>
      </template>
      <template #form_input_TaxType="{ item }">
        <div class="flex gap-[10px]">
          <s-input
            ref="refInput"
            label="Tax Registration"
            v-model="item.TaxRegistration"
            class="w-full"
            keepLabel
            :hide-label="false"
            :disabled="true"
          ></s-input>
          <s-input
            ref="refInput"
            label="Tax Type"
            v-model="item.TaxType"
            class="w-full"
            use-list
            :disabled="
              ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
                item.Status
              ) || item.VendorID
            "
            :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=TTY`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
          ></s-input>
        </div>
      </template>
      <template #form_input_TaxCodes="{ item }">
        <s-input
          :key="data.keyTaxCode"
          ref="refInput"
          label="Tax Codes"
          v-model="item.TaxCodes"
          class="w-full"
          use-list
          :disabled="true"
          :lookup-url="`/fico/taxcode/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :multiple="true"
        ></s-input>
      </template>
      <template #form_input_PurchaseType="{ item }">
        <s-input
          ref="refInput"
          label="Purchase Type"
          v-model="item.PurchaseType"
          class="w-full"
          :disabled="
            data.lockDimension ||
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          use-list
          :items="
            item.Dimension &&
            item.Dimension.find((_dim) => _dim.Key === 'Site') &&
            item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] ==
              headOffice
              ? ['STOCK', 'VIRTUAL', 'SERVICE', 'ASSET']
              : ['STOCK', 'VIRTUAL', 'SERVICE']
          "
        ></s-input>
      </template>
      <template #form_input_PIC="{ item, config }">
        <s-input
          :key="data.keyPIC"
          ref="refInput"
          label="PIC"
          v-model="item.PIC"
          class="w-full"
          :required="true"
          :disabled="
            data.lockDimension
              | ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
                item.Status
              )
          "
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilder(search, ['_id', 'Name'], item.PIC, item)
          "
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          ref="FinancialDimension"
          :key="data.keyDimension"
          sectionTitle="Financial Dimension"
          v-model="item.Dimension"
          :default-list="profile.Dimension"
          :readOnly="
            data.lockDimension ||
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          @change="
            (field, v1, v2) => {
              onChangeDimension(field, v1, v2, item);
            }
          "
        ></dimension-editor>
      </template>
      <template #form_input_Location="{ item }">
        <dimension-invent-jurnal
          ref="InventDimControl"
          :mandatory="['WarehouseID']"
          :default-list="profile.Dimension"
          v-model="item.Location"
          title-header="Inventory Dimension"
          :site="
            item.Dimension &&
            item.Dimension.find((_dim) => _dim.Key === 'Site') &&
            item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] != ''
              ? item.Dimension.find((_dim) => _dim.Key === 'Site')['Value']
              : undefined
          "
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
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          @defaultWH="
            (_id) => {
              getEmpWH(_id, item);
            }
          "
          @onChange="
            (field, v1, v2, old, val) => {
              onFieldChanged(val, item);
            }
          "
        ></dimension-invent-jurnal>
      </template>
      <template #form_input_BillingName="{ item }">
        <s-input
          ref="inputs"
          v-model="item.BillingName"
          label="Billing Name"
          field="BillingName"
          class="w-full"
          :disabled="!['DRAFT', ''].includes(item.Status)"
          use-list
          lookup-key="_id"
          :lookup-payload-builder="
            [...defaultList, 'SITE020'].length > 0
              ? (...args) =>
                  helper.payloadBuilderDimension(
                    [
                      ...defaultList,
                      'SITE020',
                      item.Dimension.find((d) => {
                        return d.Key === 'Site';
                      }).Value,
                    ],
                    item.BillingName,
                    false,
                    ...args
                  )
              : undefined
          "
          :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
          :lookup-labels="['Label']"
          :lookup-searchs="['_id', 'Label']"
          @change="
            (field, v1, v2, old, ctl) => {
              item.BillingAddress = '';
              onFormFieldChange(field, v1, v2, old, item);
            }
          "
        ></s-input>
      </template>
      <template #form_footer_1="{ item }">
        <RejectionMessageList
          ref="listRejectionMessage"
          journalType="Purchase Request"
          :journalID="item._id"
        ></RejectionMessageList>
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
        <div class="px-1" v-if="item.Status == 'POSTED' && item.TotalPrint > 0">
          <a href="#" class="mt-1">
            <mdicon
              name="printer"
              width="16"
              alt="printer"
              class="hover:text-primary"
            />
          </a>
        </div>
      </template>
      <template #grid_item_button_delete="{ item }">
        <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
      </template>
    </data-list>
    <div v-if="data.appMode == 'preview'" class="w-full">
      <PreviewReport
        class="card w-full"
        title="Purchase Request"
        :preview="data.preview"
        @close="closePreview"
        @update-print="onUpdatePrint"
        SourceType="Purchase Request"
        :SourceJournalID="data.record._id"
        :hideSignature="false"
      >
      </PreviewReport>
    </div>
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
                  {{ item.JournalID }}
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
import {
  reactive,
  ref,
  inject,
  computed,
  watch,
  onMounted,
  nextTick,
} from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton, SCard, SModal, SGrid } from "suimjs";
import { useRoute, useRouter } from "vue-router";
import helper from "@/scripts/helper.js";
import Loader from "@/components/common/Loader.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import PurchaseLine from "./widget/PurchaseLine.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import RejectionMessageList from "./widget/RejectionMessageList.vue";
import ListApprovers from "@/components/common/ListApprovers.vue";
import StatusText from "@/components/common/StatusText.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import { authStore } from "@/stores/auth";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import moment from "moment";
import lodash from "lodash";
import PreviewReport from "@/components/common/PreviewReport.vue";

layoutStore().name = "tenant";
const featureID = "PurchaseRequest";
const profile = authStore().getRBAC(featureID);
const auth = authStore();
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const refInput = ref(null);
const listControl = ref(null);
const lineConfig = ref(null);
const InventDimControl = ref(null);
const FinancialDimension = ref(null);
const listRejectionMessage = ref(null);
const route = useRoute();
const router = useRouter();
const headOffice = layoutStore().headOfficeID;
const axios = inject("axios");
const gridAttachment = ref(SGridAttachment);
let customFilter = computed(() => {
  const filters = [
    {
      Field: "Keyword",
      Op: "$eq",
      Value: data.search.Name,
    },
  ];
  const query = [];
  if (data.search.VendorID !== null && data.search.VendorID.length > 0) {
    filters.push({
      Field: "VendorID",
      Op: "$contains",
      Value: data.search.VendorID,
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
        {
          Field: "ReffNo",
          Op: "$contains",
          Value: [data.search.Name],
        },
        {
          Field: "POReff",
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

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

const linesTag = computed({
  get() {
    let ReffNo = JSON.parse(
      JSON.stringify(data.record.ReffNo ? data.record.ReffNo : [])
    ).map((ref) => {
      return `${ref.slice(0, 2)}_${ref}`;
    });

    const tags =
      ReffNo && data.record._id
        ? [...[`PR_${data.record._id}`], ...ReffNo]
        : data.record._id
        ? [`PR_${data.record._id}`]
        : ReffNo;

    return tags;
  },
});

const addTags = computed({
  get() {
    return [`PR_${data.record._id}`];
  },
});

const reffTags = computed({
  get() {
    let ReffNo = JSON.parse(
      JSON.stringify(data.record.ReffNo ? data.record.ReffNo : [])
    ).map((ref) => {
      return `${ref.slice(0, 2)}_${ref}`;
    });
    return ReffNo;
  },
});

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  stayOnForm: true,
  lockDimension: false,
  lockPostingProfile: false,
  lockInventDimension: false,
  isDialogReffNo: false,
  keyJournalType: util.uuid(),
  keyDimension: util.uuid(),
  keyPIC: util.uuid(),
  keyRequestor: util.uuid(),
  keyTaxCode: util.uuid(),
  titleForm: "Purchase Request",
  transactionType: "",
  journalTypeData: {},
  preview: {},
  formInsert: "/scm/purchase/request/save",
  formUpdate: "/scm/purchase/request/save",
  tabs: ["General"],
  disableField: [],
  listTTD: [],
  listReffNo: [],
  record: {
    _id: "",
    TrxDate: new Date(),
    DocumentDate: new Date(),
    PRDate: new Date(),
    ExpectedDate: new Date(),
    Status: "",
    TrxType: "",
    WarehouseID: "",
    DeliveryName: "",
    DeliveryAddress: "",
    PIC: "",
    BillingName: "",
    BillingAddress: "",
    Freight: 0,
    Lines: [],
    AttachmentID: "",
  },
  search: {
    VendorID: [],
    Site: "",
    Name: "",
    DateFrom: null,
    DateTo: null,
    Status: "",
  },
  isFilterSite: false,
  fileAttach: {},
  gridCfgReffNo: {},
  isPostFlow: false,
  siteUser: "",
  currentUser: "",
  requestorName: "",
});

function newRecord(record) {
  data.stayOnForm = true;
  data.formMode = "new";
  record._id = "";
  record.CompanyID = auth.companyId;
  record.Requestor = "";
  record.TrxDate = new Date();
  record.DocumentDate = new Date();
  record.PRDate = new Date();
  record.ExpectedDate = new Date();
  record.Status = "";
  record.TrxType = data.transactionType;
  record.Freight = 0;
  record.ReffNo = [];
  record.BillingName = "";
  record.Dimension = [
    {
      Key: "PC",
      Value: "",
    },
    {
      Key: "CC",
      Value: "",
    },
    {
      Key: "Site",
      Value: defaultList.length == 1 ? defaultList[0] : "",
    },
    {
      Key: "PC",
      Value: "",
    },
    {
      Key: "Asset",
      Value: "",
    },
  ];
  record.Location = {
    WarehouseID: "",
  };
  record.Discount = {
    DiscountType: "percent",
    DiscountValue: 0,
    DiscountAmount: 0,
  };
  data.titleForm = `Create New Purchase Request`;
  data.record = record;
  getPostingProfile(
    record,
    () => {
      data.lockDimension = false;
      data.lockPostingProfile = false;
      data.lockInventDimension = false;
      openForm(record);
    },
    () => {
      data.lockDimension = false;
      data.lockPostingProfile = false;
      data.lockInventDimension = false;
      openForm(record);
    }
  );
}
function editRecord(record) {
  data.stayOnForm = true;
  data.formMode = "edit";
  data.record = record;
  data.titleForm = `Edit Purchase Request | ${record._id}`;
  record.Lines.map(function (l) {
    l.ItemVarian = helper.ItemVarian(l.ItemID, l.SKU);
    return l;
  });

  openForm(record);
  nextTick(() => {
    if (["SUBMITTED", "READY", "REJECTED", "POSTED"].includes(record.Status)) {
      // if (record.Status === "REJECTED") {
      listRejectionMessage.value.fetchRecord(record._id);
      // }
      setFormMode("view");
    } else {
      if (record.ReffNo.length > 0) {
        getEmployee(record);
        getPostingProfile(record);
      }
    }
  });
}
function openForm(record) {
  util.nextTickN(2, () => {
    data.fileAttach = {};
    data.isPostFlow = false;
    listControl.value.setFormFieldAttr(
      "_id",
      "hide",
      data.formMode == "new" ? true : false
    );
    const el = document.querySelector(
      "#pr-data-list .form_inputs > div.flex.section_group_container > div:nth-child(1) > div:nth-child(1) > div.flex.flex-col.gap-4 > div:nth-child(1)"
    );
    if (record._id == "") {
      el ? (el.style.display = "none") : "";
      data.tabs = ["General"];
      data.keyRequestor = util.uuid();
    } else {
      el ? (el.style.display = "block") : "";
      data.tabs = ["General", "Line", "Attachment"];
      if (record.VendorID) {
        listControl.value.setFormFieldAttr("TaxName", "readOnly", true);
        listControl.value.setFormFieldAttr("PaymentTerms", "readOnly", true);
        listControl.value.setFormFieldAttr("TaxType", "readOnly", true);
        listControl.value.setFormFieldAttr("TaxCodes", "readOnly", true);
        listControl.value.setFormFieldAttr("TaxAddress", "readOnly", true);
      } else {
        listControl.value.setFormFieldAttr("TaxName", "readOnly", false);
        listControl.value.setFormFieldAttr("PaymentTerms", "readOnly", false);
        listControl.value.setFormFieldAttr("TaxType", "readOnly", false);
        listControl.value.setFormFieldAttr("TaxCodes", "readOnly", false);
        listControl.value.setFormFieldAttr("TaxAddress", "readOnly", false);
      }
    }
    listControl.value.setFormLoading(false);
  });
}
function preSaveAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [`PR_${data.record._id}`];
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

function onSave(record) {
  const dim = record.Location;
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
  let validDimension = true;
  if (InventDimControl.value) {
    validate = InventDimControl.value.validate();
  }

  if (FinancialDimension.value) {
    validDimension = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      validDimension = false;
    }
  }

  if (!record.Name) {
    return util.showError("General PR Name is required");
  }

  if (!record.Requestor) {
    return util.showError("General Requestor is required");
  }

  if (!record.Location?.WarehouseID) {
    return util.showError("General Warehouse ID is required");
  }
  let valid = true;
  for (let l = 0; l < record?.Lines?.length; l++) {
    if (!record.Lines[l].ItemID) {
      return util.showError("there is a line field Item required");
    }
    if (typeof record.Lines[l].Qty != "number") {
      return util.showError("there is a line field Qty required");
    }
    if (typeof record.Lines[l].UnitCost != "number") {
      return util.showError("there is a line field UnitCost required");
    }
    if (record.Lines[l].UnitCost == 0 && data.siteUser != headOffice) {
      return util.showError("UnitCost item which is 0");
    }
    if (typeof record.Lines[l].DiscountValue != "number") {
      return util.showError("there is a line field Discount Value required");
    }
  }

  if (record?.Lines?.length > 0) {
    record.Lines.map(function (v, idx) {
      v.RemainingQty = v.Qty;
      v.LineNo = idx + 1;
      v.SourceLineNo = idx + 1;
      // if (dim.WarehouseID) {
      //   v.InventDim.WarehouseID = dim.WarehouseID;
      // }
      // if (dim.AisleID) {
      //   v.InventDim.AisleID = dim.AisleID;
      // }
      // if (dim.SectionID) {
      //   v.InventDim.SectionID = dim.SectionID;
      // }
      // if (dim.BoxID) {
      //   v.InventDim.BoxID = dim.BoxID;
      // }
      // v.Dimension  = record.Dimension
      return v;
    });
  }

  let payload = JSON.parse(JSON.stringify(record));
  if (record.Status == "") {
    payload.Status = "DRAFT";
  }
  if (lineConfig.value) {
    payload = {
      ...payload,
      ...JSON.parse(JSON.stringify(lineConfig.value.getOtherTotal())),
    };
    if (typeof payload.Discount?.DiscountValue != "number") {
      data.disableButton = false;
      listControl.value.setFormLoading(false);
      return util.showError("field Discount Type cannot be empty");
    }
    if (typeof payload.Discount?.DiscountAmount != "number") {
      data.disableButton = false;
      listControl.value.setFormLoading(false);
      return util.showError("field Discount Type cannot be empty");
    }
  }

  if (validDimension && valid && validate && listControl.value.formValidate()) {
    listControl.value.setFormLoading(true);
    data.disableButton = true;
    payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
    payload.PRDate = helper.dateTimeNow(payload.PRDate);
    listControl.value.submitForm(
      payload,
      () => {
        data.disableButton = false;
        listControl.value.setFormLoading(false);
      },
      () => {
        data.disableButton = false;
        listControl.value.setFormLoading(false);
      }
    );
  } else {
    return util.showError("field is required");
  }
}

function onFormRecordChange(record) {}
function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "VendorID":
      // record.DeliveryName = v2;
      // record.BillingName = v2;
      listControl.value.setFormFieldAttr("TaxName", "readOnly", false);
      listControl.value.setFormFieldAttr("PaymentTerms", "readOnly", false);
      listControl.value.setFormFieldAttr("TaxType", "readOnly", false);
      listControl.value.setFormFieldAttr("TaxAddress", "readOnly", false);
      data.keyTaxCode = util.uuid();
      record.VendorName = "";
      record.PaymentTerms = "";
      record.TaxType = "";
      record.TaxName = "";
      record.TaxRegistration = "";
      record.TaxAddress = "";
      record.TaxCodes = [];
      if (typeof v1 == "string") {
        axios.post("/bagong/vendor/get", [v1]).then(
          (r) => {
            record.VendorName = r.data.Name;
            record.PaymentTerms = r.data.PaymentTermID;
            record.TaxType = r.data.TaxType;
            record.TaxName = r.data.TaxName;
            record.TaxRegistration = r.data.TaxRegistrationNumber;
            record.TaxAddress = r.data.TaxAddress;
            let TaxCodes = [];
            if (r.data.Detail.Terms.Taxes1) {
              TaxCodes.push(r.data.Detail.Terms.Taxes1);
            }
            if (r.data.Detail.Terms.Taxes2) {
              TaxCodes.push(r.data.Detail.Terms.Taxes2);
            }
            record.TaxCodes = TaxCodes;
            listControl.value.setFormFieldAttr("TaxName", "readOnly", true);
            listControl.value.setFormFieldAttr(
              "PaymentTerms",
              "readOnly",
              true
            );
            listControl.value.setFormFieldAttr("TaxType", "readOnly", true);
            listControl.value.setFormFieldAttr("TaxAddress", "readOnly", true);
            axios
              .post("/fico/paymentterm/get", [r.data.PaymentTermID])
              .then((p) => {
                record.PRDate = moment(new Date(record.TrxDate)).add(
                  p.data.Days,
                  "days"
                );
              })
              .finally(function () {
                if (lineConfig.value) {
                  util.nextTickN(2, () => {
                    lineConfig.value.gridRefreshed();
                  });
                }
              });
          },
          (e) => {
            if (lineConfig.value) {
              util.nextTickN(2, () => {
                lineConfig.value.gridRefreshed();
              });
            }
          }
        );
      }
      break;
    case "WarehouseID":
      if (typeof v1 == "string") {
        axios.post("/tenant/warehouse/get", [v1]).then((r) => {
          record.DeliveryAddress = r.data.Address;
          record.PIC = r.data.PIC;
        });
      } else {
        record.DeliveryAddress = "";
        record.PIC = "";
      }
      break;
    case "TrxDate":
      record.PRDate = moment(new Date(v1)).add(0, "days");
      axios.post("/fico/paymentterm/get", [record.PaymentTerms]).then((p) => {
        record.PRDate = moment(new Date(v1)).add(p.data.Days, "days");
      });
      break;
    case "BillingName":
      if (v1 === "SITE020") {
        getWarehouse("WH-HO", record);
      } else {
        record.BillingAddress = record.DeliveryAddress;
      }

      break;
  }
}
function onPreSave(record) {
  data.record.Freight = parseFloat(data.record.Freight);
}
function onPostSave(record) {
  data.record = record;
  if (
    record.Status === "DRAFT" &&
    gridAttachment.value &&
    data.formMode == "edit"
  ) {
    gridAttachment.value.Save();
    postSaveAttachment();
  }
}

function preSubmit(status, action, doSubmit) {
  data.stayOnForm = data.record.Status == "DRAFT" ? true : false;
  const dim = data.record.Location;
  const pc = data.record.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = data.record.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = data.record.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;

  if (!data.record.Requestor) {
    return util.showError("General Requestor is required");
  }

  if (!data.record.BillingName) {
    return util.showError("General Billing Name is required");
  }

  if (!data.record.BillingAddress) {
    return util.showError("General Billing Address is required");
  }

  data.isPostFlow = true;
  if (status == "DRAFT") {
    data.record.Freight = parseFloat(data.record.Freight);
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

    if (data.record.Lines.length == 0) {
      return util.showError("Line items is empty");
    }

    for (let l = 0; l < data.record.Lines.length; l++) {
      if (!data.record.Lines[l].ItemID) {
        return util.showError("there is a line field Item required");
      }
      if (typeof data.record.Lines[l].Qty != "number") {
        return util.showError("there is a line field Qty required");
      }
      if (typeof data.record.Lines[l].UnitCost != "number") {
        return util.showError("there is a line field UnitCost required");
      }
      if (data.record.Lines[l].UnitCost == 0 && data.siteUser != headOffice) {
        return util.showError("UnitCost item which is 0");
      }
      if (typeof data.record.Lines[l].DiscountValue != "number") {
        return util.showError("there is a line field Discount Value required");
      }
      data.record.Lines[l].RemainingQty = data.record.Lines[l].Qty;
    }
    setRequiredAllField(true);

    util.nextTickN(2, () => {
      let valid = listControl.value.formValidate();
      let validDimension = true;
      if (!valid) {
        return false;
      }

      const checkVendor = [
        "PostingProfileID",
        "VendorID",
        "BillingName",
        "BillingAddress",
        "PIC",
      ].every((e) => {
        if (!data.record[e]) {
          return false;
        }
        return true;
      });

      if (data.siteUser != headOffice) {
        valid = checkVendor;
      }
      if (FinancialDimension.value) {
        validDimension = FinancialDimension.value.validate();
      } else {
        if (!pc || !cc || !site) {
          return util.showError("Financial Dimension is required");
        }
      }
      if (validDimension && valid) {
        listControl.value.setFormLoading(true);
        data.disableButton = true;
        let payload = JSON.parse(JSON.stringify(data.record));
        payload = {
          ...payload,
          ...JSON.parse(JSON.stringify(lineConfig.value.getOtherTotal())),
        };
        if (typeof payload.Discount?.DiscountValue != "number") {
          data.disableButton = false;
          listControl.value.setFormLoading(false);
          return util.showError("field Discount Type cannot be empty");
        }
        if (typeof payload.Discount?.DiscountAmount != "number") {
          data.disableButton = false;
          listControl.value.setFormLoading(false);
          return util.showError("field Discount Type cannot be empty");
        }
        payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
        payload.PRDate = helper.dateTimeNow(payload.PRDate);
        listControl.value.submitForm(
          payload,
          () => {
            doSubmit();
          },
          () => {
            data.disableButton = false;
            listControl.value.setFormLoading(false);
          }
        );
      } else {
        return util.showError("Please check general required field");
      }
      setRequiredAllField(false);
    });
  }
}
function postSubmit(record) {
   axios.post("/scm/purchase/request/sync-journal-status", { PurchaseRequestID: record._id })
  .then((r)=>{
    data.disableButton = false;
    listControl.value.setFormLoading(false);
    data.appMode = "grid";
    listControl.value.refreshList();
    listControl.value.refreshForm();
    listControl.value.setControlMode("grid");
  },(e)=>{
    return util.showError(e);
  })
}

function preReopen() {
  let payload = JSON.parse(JSON.stringify(data.record));
  payload = {
    ...payload,
    ...JSON.parse(JSON.stringify(lineConfig.value.getOtherTotal())),
  };
  payload.Status = "DRAFT";
  axios.post("/scm/purchase/request/save", payload).then(
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

function errorSubmit(e) {
  data.disableButton = false;
  listControl.value.setFormLoading(false);
  // return util.showError(e);
}
function setRequiredAllField(required) {
  const site = data.record.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;

  listControl.value.getFormAllField().forEach((e) => {
    if (
      [
        "PostingProfileID",
        "BillingName",
        "BillingAddress",
        "PIC",
        "Requestor",
      ].includes(e.field)
    ) {
      listControl.value.setFormFieldAttr(e.field, "required", required);
    }
  });

  if (site == headOffice) {
    listControl.value.setFormFieldAttr("VendorID", "required", false);
    listControl.value.setFormFieldAttr("BillingName", "required", false);
    listControl.value.setFormFieldAttr("BillingAddress", "required", false);
  } else {
    listControl.value.setFormFieldAttr("VendorID", "required", true);
    listControl.value.setFormFieldAttr("BillingName", "required", true);
    listControl.value.setFormFieldAttr("BillingAddress", "required", true);
  }
}
function setFormMode(mode) {
  listControl.value.setFormMode(mode);
}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = route.query.title;
  }
}
function alterGridConfig(cfg) {
  cfg.sortable = ["LastUpdate", "Created", "TrxDate", "Status", "_id"];
  cfg.setting.idField = "LastUpdate";
  cfg.setting.sortable = ["LastUpdate", "Created", "TrxDate", "Status", "_id"];
}
function alterFormConfig(config) {
  if (route.query.trxid !== undefined) {
    let currQuery = { ...route.query };
    listControl.value.selectData({ _id: currQuery.trxid }); //remark sementara tunggu suimjs update
    delete currQuery["trxid"];
    router.replace({ path: route.path, query: currQuery });
  } else if (route.query.id !== undefined) {
    let getUrlParam = route.query.id;
    listControl.value.selectData({ _id: getUrlParam });
    router.replace({
      path: `/scm/PurchaseRequest`,
    });
  }
}

function setLockForm(JT) {
  data.lockDimension = JT.LockFinancialDimension;
  data.lockInventDimension = JT.LockInventoryDimension;
  data.lockPostingProfile = JT.LockPostingProfile;
}

function getWarehouse(WarehouseID, record) {
  axios.post("/tenant/warehouse/get", [WarehouseID]).then((r) => {
    if (r.data) {
      record.BillingAddress = r.data.Address;
    }
  });
}

function onFieldChanged(val, record) {
  if (val.WarehouseID) {
    data.disableField.push("WarehouseID");
    record.DeliveryAddress = "";
    record.PIC = "";
    record.DeliveryName = "";
    record.BillingAddress = "";
    axios.post("/tenant/warehouse/get", [val.WarehouseID]).then((r) => {
      if (r.data) {
        record.DeliveryAddress = r.data.Address;
        record.PIC = r.data.PIC;
        record.DeliveryName = r.data.Name;
      }

      if (record.BillingName == "SITE020") {
        getWarehouse("WH-HO", record);
      } else if (record.BillingName != "") {
        record.BillingAddress = record.DeliveryAddress;
      }
      data.keyPIC = util.uuid();
    });
  }
  if (val.AisleID) {
    data.disableField.push("AisleID");
  }
  if (val.SectionID) {
    data.disableField.push("SectionID");
  }
  if (val.BoxID) {
    data.disableField.push("BoxID");
  }
  if (!lineConfig.value) {
    return true;
  }
  // const line = lineConfig.value.getDataValue();
  // if (line.length > 0) {
  //   for (let l = 0; l < line.length; l++) {
  //     line[l].InventDim = val;
  //   }
  // }
}
function getPostingProfile(record, cbOK, cbFalse) {
  util.nextTickN(2, () => {
    listControl.value.setFormLoading(true);
    axios
      .post(`/scm/purchase/request/journal/type/find?TrxType=Purchase Request`)
      .then(
        (r) => {
          if (r.data.length > 0) {
            if (defaultList.length === 0 || defaultList.includes(headOffice)) {
              record.JournalTypeID = r.data.filter((d) => {
                return d.Name === "PR-HO";
              })[0]._id;
              record.PostingProfileID = r.data.filter((d) => {
                return d.Name === "PR-HO";
              })[0].PostingProfileID;
            } else {
              record.JournalTypeID = r.data.filter((d) => {
                return d.Name !== "PR-HO";
              })[0]._id;
              record.PostingProfileID = r.data.filter((d) => {
                return d.Name !== "PR-HO";
              })[0].PostingProfileID;
            }

            data.record = record;
            data.keyJournalType = util.uuid();
            if (cbOK) {
              cbOK();
            }
          }
        },
        (e) => {
          if (cbFalse) {
            cbFalse();
          }
          util.showError(e);
        }
      )
      .finally(function () {
        listControl.value.setFormLoading(false);
      });
  });
}
function getDetailEmployee(_id, record) {
  axios.post("/tenant/warehouse/get", [_id]).then(
    (r) => {
      record.DeliveryAddress = r.data.Address;
      record.PIC = r.data.PIC;
      record.DeliveryName = r.data.Name;
      data.keyPIC = util.uuid();
    },
    (e) => util.showError(e)
  );
}

function getEmpWH(_id, record) {
  axios.post("/tenant/warehouse/get", [_id]).then(
    (r) => {
      if (!record._id) {
        record.DeliveryAddress = r.data.Address;
        record.PIC = r.data.PIC;
        record.DeliveryName = r.data.Name;
        data.keyPIC = util.uuid();
      }
    },
    (e) => util.showError(e)
  );
}
function getEmployee(record) {
  if (
    record.ReffNo.length > 0 &&
    data.siteUser != headOffice &&
    auth.appData.Email != "satria@kanosolution.com"
  ) {
    data.lockDimension = true;
    data.lockInventDimension = true;
    setFormMode("view");
  }
}

function getRequestor(record) {
  axios.post("/bagong/employee/get", [record.Requestor]).then(
    (r) => {
      data.requestorName = r.data.Name;
    },
    (e) => util.showError(e)
  );
}

function onChangeDimension(field, v1, v2, item) {
  switch (field) {
    case "Site":
      item.VendorID = "";
      item.Requestor = "";
      item.PIC = "";
      item.DeliveryName = "";
      item.DeliveryAddress = "";
      item.Location.WarehouseID = "";
      item.BillingAddress = "";
      item.BillingName = "";
      onFormFieldChange("VendorID", "", "", "", item);
      break;
    default:
      break;
  }
}

function getByCurrentUser() {
  axios.post("/tenant/employee/get-by-current-user").then(
    (r) => {
      let Site = "";
      Site =
        r.data.Dimension &&
        r.data.Dimension.find((_dim) => _dim.Key === "Site") &&
        r.data.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
          ? r.data.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
          : undefined;

      data.siteUser = Site;
      data.currentUser = r.data._id;
      if (
        Site != headOffice &&
        auth.appData.Email != "satria@kanosolution.com"
      ) {
        data.isFilterSite = true;
        data.search.Site = Site;
      } else {
        data.isFilterSite = false;
      }
    },
    (e) => util.showError(e)
  );
}

watch(
  () => route.query.type,
  (nv) => {
    data.transactionType = route.query.type;

    util.nextTickN(2, () => {
      data.appMode = "grid";
    });
  }
);
watch(
  () => route.query.title,
  (nv) => {
    data.titleForm = nv;
  }
);

function onPreview() {
  getRequestor(data.record);
  data.appMode = "preview";
}

function closePreview() {
  data.appMode = "grid";
}
function onUpdatePrint(SourceType, SourceJournalID) {
  const payload = {
    SourceType: SourceType,
    SourceJournalID: SourceJournalID,
  };
  helper.updatePrint(axios, "scm", "purchase", payload);
}

function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
function lookupPayloadBuilder(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

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
  if (Site.length > 0) {
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
    if (Site.length > 0) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}
function lookupPayloadBuilderRequestor(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

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

  if (Site.length > 0) {
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

    if (Site.length > 0) {
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

function lookupPayloadBuilderLogin(search, select, value, item) {
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
          { Field: "Label", Op: "$contains", Value: [search] },
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

function getReffNo() {
  data.listReffNo = [];
  axios
    .post("/scm/postingprofile/find-journal-ref", {
      JournalType: "Purchase Request",
      JournalID: data.record._id,
    })
    .then(
      (r) => {
        data.listReffNo = r.data.Refferences;
        data.isDialogReffNo = true;
      },
      (e) => util.showError(e)
    );
}

function redirectReff(item) {
  let page = "";
  let query = {};
  if (item.JournalType == "Item Request") {
    page = "scm-ItemRequest";
    query = { id: item.JournalID };
  } else if (item.JournalType == "Movement In") {
    page = "scm-InventoryJournal";
    query = {
      id: item.JournalID,
      type: "Movement In",
      title: "Movement In",
    };
  } else if (item.JournalType == "Movement Out") {
    page = "scm-InventoryJournal";
    query = {
      id: item.JournalID,
      type: "Movement Out",
      title: "Movement Out",
    };
  } else if (item.JournalType == "Transfer") {
    page = "scm-InventoryJournal";
    query = {
      id: item.JournalID,
      type: "Transfer",
      title: "Item Transfer",
    };
  } else if (item.JournalType == "Purchase Request") {
    page = "scm-PurchaseRequest";
    query = { id: item.JournalID };
  } else if (item.JournalType == "Purchase Order") {
    page = "scm-PurchaseOrder";
    query = { id: item.JournalID };
  } else if (item.JournalType == "Inventory Receive") {
    page = "scm-InventTrx";
    query = {
      id: item.JournalID,
      type: "Inventory Receive",
      title: "Inventory Receive",
    };
  } else if (item.JournalType == "Inventory Issuance") {
    page = "scm-InventTrx";
    query = {
      id: item.JournalID,
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
  createGridCfgRefNo();
  getByCurrentUser();
});
</script>
<style>
.adminfee,
.parkingfee,
.deliveryfee {
  padding-left: 6px;
}
</style>
