<template>
  <div class="w-full">
    <s-modal :display="false" ref="deleteModal" @submit="confirmDelete">
      You will delete data ! Are you sure ?<br />
      Please be noted, this can not be undone !
    </s-modal>
    <div v-show="data.appMode != 'preview'">
      <data-list
        v-show="data.isDataList"
        :key="data.keyList"
        class="card"
        ref="listControl"
        :title="data.titleForm"
        :form-hide-submit="true"
        grid-config="/scm/inventory/journal/gridconfig"
        :form-config="fromUrlUI"
        :grid-read="
          `/scm/inventory/journal/gets-v1?TrxType=` + data.transactionType
        "
        form-read="/scm/inventory/journal/get"
        grid-mode="grid"
        grid-delete="/scm/inventory/journal/delete"
        form-keep-label
        :form-insert="data.formInsert"
        :form-update="data.formUpdate"
        :grid-fields="[
          'Enable',
          'WarehouseID',
          'WarehouseFrom',
          'WarehouseTo',
          'Status',
          'Approvers',
        ]"
        :form-fields="[
          '_id',
          'JournalTypeID',
          'PostingProfileID',
          'ReffNo',
          'Dimension',
          'DimensionFrom',
          'InventDim',
          'InventDimTo',
        ]"
        grid-sort-field="LastUpdate"
        grid-sort-direction="desc"
        :init-app-mode="data.appMode"
        :init-form-mode="data.formMode"
        :form-tabs-new="data.tabs"
        :form-tabs-edit="data.tabs"
        :form-tabs-view="data.tabs"
        @formNewData="newRecord"
        @formEditData="editRecord"
        @preSave="onPreSave"
        @postSave="onPostSave"
        @controlModeChanged="onControlModeChanged"
        @alterFormConfig="alterFormConfig"
        @alterGridConfig="alterGridConfig"
        :stay-on-form-after-save="data.stayOnForm"
        :grid-hide-new="!getProfile().canCreate"
        :grid-hide-edit="!getProfile().canUpdate"
        :grid-hide-delete="true"
        grid-hide-select
        :grid-custom-filter="customFilter"
      >
        <template #grid_header_search="{ config }">
          <s-input
            ref="refName"
            v-model="data.search.Text"
            lookup-key="_id"
            label="Text"
            class="w-full"
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
            class="w-[400px]"
            use-list
            :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
            :lookup-labels="['Label']"
            :lookup-searchs="['_id', 'Label']"
            :lookup-payload-builder="
              getDefaultList()?.length > 0
                ? (...args) =>
                    helper.payloadBuilderDimension(
                      getDefaultList(),
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
        <template #grid_item_buttons_2="{ item, config }">
          <a
            href="#"
            v-if="['DRAFT'].includes(item.Status) && getProfile().canDelete"
            @click="deleteData(item)"
            class="delete_action"
          >
            <mdicon
              name="delete"
              width="16"
              alt="delete"
              class="cursor-pointer hover:text-primary"
            />
          </a>
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
              !['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(
                item.Status
              )
            "
            :icon="`content-save`"
            class="btn_primary submit_btn"
            label="Save"
            :disabled="data.disabledFormButton"
            @click="onSave(item)"
          />
          <FormButtonsTrx
            :posting-profile-id="item.PostingProfileID"
            :disabled="data.disabledFormButton"
            :status="item.Status"
            :journalId="item._id"
            :journal-type-id="
              data.transactionType == 'Transfer' ? 'Transfer' : 'INVENTORY'
            "
            :autoReopen="false"
            moduleid="scm/new"
            @pre-reopen="preReopen"
            @pre-submit="preSubmit"
            @post-submit="postSubmit(item)"
            @error-submit="errorSubmit"
          >
          </FormButtonsTrx>
        </template>
        <template #form_tab_Line="{ item }">
          <InventoryJournalLine
            ref="lineConfig"
            v-model="item.Lines"
            :general-record="item"
            :transaction-type="data.transactionType"
            :disable-field="data.disableField"
            :is-from-ref="data.lockDimensionTf"
          ></InventoryJournalLine>
        </template>
        <template #grid_WarehouseID="{ item }">
          {{ item.InventDim.WarehouseID }}
        </template>
        <template #grid_WarehouseFrom="{ item }">
          {{ item.InventDim.WarehouseID }}
        </template>
        <template #grid_WarehouseTo="{ item }">
          {{ item.InventDimTo.WarehouseID }}
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
            :lookup-url="`/scm/inventory/journal/type/find?TransactionType=${data.transactionType}`"
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
              ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(item.Status)
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
            ref="FinancialDimension"
            :key="data.compDimensionKey"
            v-model="item.Dimension"
            :sectionTitle="`Financial Dimension ${
              data.transactionType == 'Transfer' ? 'To' : ''
            }`"
            :default-list="getProfile().Dimension"
            :readOnly="
              data.lockDimension ||
              data.lockDimensionTf ||
              ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(item.Status)
            "
            @change="
              (field, v1, v2) => {
                if (field == 'Site' && data.transactionType != 'Transfer') {
                  item.InventDim.WarehouseID = '';
                } else if (
                  field == 'Site' &&
                  data.transactionType == 'Transfer'
                ) {
                  item.InventDimTo.WarehouseID = '';
                }
              }
            "
          ></dimension-editor>
        </template>
        <template #form_input_DimensionFrom="{ item }">
          <dimension-editor
            ref="FinancialDimensionFrom"
            :key="data.dimensionFromKey"
            v-model="item.DimensionFrom"
            :sectionTitle="`Financial Dimension From`"
            :default-list="getProfile().Dimension"
            :readOnly="
              data.lockDimension ||
              ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(item.Status)
            "
            @change="
              (field, v1, v2) => {
                if (field == 'Site') {
                  item.InventDim.WarehouseID = '';
                }
              }
            "
          ></dimension-editor>
        </template>
        <template #form_input_InventDim="{ item }">
          <dimension-invent-jurnal
            v-if="
              [
                'Transfer',
                'Movement In',
                'Movement Out',
                'Stock Opname',
              ].includes(data.transactionType)
            "
            ref="InventDimControlFrom"
            :mandatory="['WarehouseID']"
            :disable-field="item.ReffNo?.length > 0 ? [] : []"
            :key="data.compInventDimensionKey"
            :defaultList="profile.Dimension"
            v-model="item.InventDim"
            :site="
              ['Movement In', 'Movement Out'].includes(data.transactionType)
                ? item.Dimension &&
                  item.Dimension.find((_dim) => _dim.Key === 'Site') &&
                  item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] !=
                    ''
                  ? item.Dimension.find((_dim) => _dim.Key === 'Site')['Value']
                  : undefined
                : item.DimensionFrom &&
                  item.DimensionFrom.find((_dim) => _dim.Key === 'Site') &&
                  item.DimensionFrom.find((_dim) => _dim.Key === 'Site')[
                    'Value'
                  ] != ''
                ? item.DimensionFrom.find((_dim) => _dim.Key === 'Site')[
                    'Value'
                  ]
                : undefined
            "
            :title-header="`Inventory Dimension ${
              data.transactionType == 'Transfer' ? 'From' : ''
            }`"
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
              ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(item.Status)
            "
            @onFieldChanged="
              (val) => {
                if (data.transactionType != 'Transfer') {
                  onFieldChanged(val, item);
                }
              }
            "
          ></dimension-invent-jurnal>
        </template>
        <template #form_input_InventDimTo="{ item }">
          <dimension-invent-jurnal
            v-if="['Transfer'].includes(data.transactionType)"
            ref="InventDimControl"
            :mandatory="['WarehouseID']"
            :disable-field="data.disableDimTo"
            :key="data.compInventDimensionKey"
            v-model="item.InventDimTo"
            :site="
              item.Dimension &&
              item.Dimension.find((_dim) => _dim.Key === 'Site') &&
              item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] != ''
                ? item.Dimension.find((_dim) => _dim.Key === 'Site')['Value']
                : undefined
            "
            title-header="Inventory Dimension To"
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
              ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(item.Status)
            "
            @onFieldChanged="
              (val) => {
                onFieldChanged(val, item);
              }
            "
          ></dimension-invent-jurnal>
        </template>
        <template #form_input_ReffNo="{ item }">
          <s-input
            v-if="data.transactionType == 'Transfer'"
            ref="refRefNo"
            v-model="item.ReffNo"
            useList
            :readOnly="
              ['SUBMITTED', 'REJECTED', 'READY', 'POSTED'].includes(item.Status)
            "
            :multiple="true"
            :allowAdd="true"
            lookup-key="_id"
            :lookup-url="`/scm/item/request/find`"
            :lookup-labels="['_id']"
            :lookup-searchs="['_id']"
            label="Ref No"
            class="w-full"
          ></s-input>
        </template>
        <template #form_footer_1="{ item }">
          <RejectionMessageList
            ref="listRejectionMessage"
            :journalType="
              data.transactionType == 'Transfer' ? 'Transfer' : 'INVENTORY'
            "
            :journalID="item._id"
          ></RejectionMessageList>
        </template>
        <template #grid_item_buttons_1="{ item }">
          <log-trx v-if="helper.isShowLog(item.Status)" :id="item._id" />
        </template>
        <template #grid_item_button_delete="{ item }">
          <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
        </template>
        <template #grid_Approvers="{ item }">
          <list-approvers :approvers="item.Approvers" />
        </template>
      </data-list>
      <s-card
        v-show="!data.isDataList"
        :title="`${data.titleForm}`"
        class="w-full bg-white suim_datalist card"
        hide-footer
        :no-gap="false"
        :hide-title="false"
      >
        <s-grid
          v-model="data.itemFulfillment"
          ref="listControlFulfill"
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
          @checkUncheckAll="onCheckUncheckAll"
          @checkUncheck="onCheckUncheck"
        >
          <template #header_search="{ config }">
            <s-input
              ref="refSite"
              label="Site"
              kind="input"
              class="w-full"
              v-model="data.searchItem.Site"
              :allow-add="false"
              useList
              lookup-key="_id"
              :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
              :lookup-labels="['Label']"
              :lookup-searchs="['_id', 'Label']"
              :caption="'Site'"
              :lookup-payload-builder="
                defaultListTF?.length > 0
                  ? (...args) =>
                      helper.payloadBuilderDimension(
                        defaultListTF,
                        data.searchItem.Site,
                        false,
                        ...args
                      )
                  : undefined
              "
              @change="
                () => {
                  data.searchItem.Keyword = '';
                  onFilterRefresh();
                }
              "
            />
            <s-input
              label="From Warehouse"
              v-model="data.searchItem.Warehouse"
              class="w-full"
              use-list
              lookup-url="/tenant/warehouse/find"
              lookup-key="_id"
              :lookup-labels="['_id', 'Name']"
              :lookup-searchs="['_id', 'Name']"
              @change="
                () => {
                  onFilterRefresh();
                }
              "
            ></s-input>
            <s-input
              ref="refItemID"
              v-model="data.searchItem.Keyword"
              label="Search keyword"
              class="w-full"
              @keyup.enter="onFilterRefresh"
            ></s-input>
          </template>
          <template #header_buttons="{ config }">
            <s-button
              icon="refresh"
              class="btn_primary refresh_btn"
              @click="onFilterRefresh"
            />
            <s-button
              label="Process"
              class="btn_primary refresh_btn"
              @click="onProcess"
            />
            <s-button
              label="Back"
              class="btn_warning back_btn"
              icon="rewind"
              @click="onBack"
            />
          </template>
          <template #paging>
            <s-pagination
              :recordCount="data.itemFulfillment.length"
              :pageCount="pageCount"
              :current-page="data.paging.currentPage"
              :page-size="data.paging.pageSize"
              @changePage="changePage"
              @changePageSize="changePageSize"
            ></s-pagination>
          </template>
          <template #item_ItemID="{ item }">
            {{ item.ItemName }}
          </template>
          <template #item_Dimension="{ item }">
            <DimensionText :dimension="item.Dimension" />
          </template>
          <template #item_WarehouseID="{ item }">
            {{ item.WarehouseID.join(",") }}
          </template>
        </s-grid>
      </s-card>
      <s-modal
        :display="data.isDialogReffNo"
        ref="fefModal"
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
    <PreviewReport
      v-if="data.appMode == 'preview'"
      class="card w-full"
      title="Preview"
      :preview="data.preview"
      @close="closePreview"
      :SourceType="route.query.type === 'Transfer' ? 'Transfer' : 'INVENTORY'"
      :SourceJournalID="data.record._id"
      :hideSignature="false"
    ></PreviewReport>
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
import {
  DataList,
  util,
  SInput,
  SButton,
  SModal,
  SCard,
  SGrid,
  SPagination,
  loadGridConfig,
} from "suimjs";
import { useRouter, useRoute } from "vue-router";
import Loader from "@/components/common/Loader.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import InventoryJournalLine from "./widget/InventoryJournalLine.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import RejectionMessageList from "./widget/RejectionMessageList.vue";
import StatusText from "@/components/common/StatusText.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import ListApprovers from "@/components/common/ListApprovers.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import SInputSkuItem from "./widget/SInputSkuItem.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import { authStore } from "@/stores/auth.js";
import moment from "moment";
import helper from "@/scripts/helper.js";
layoutStore().name = "tenant";

const featureID = "MovementIn";
const featureID2 = "MovementOut";
const featureID3 = "ItemTransfer";
const profile = authStore().getRBAC(featureID);
const profile2 = authStore().getRBAC(featureID2);
const profile3 = authStore().getRBAC(featureID3);

const defaultListIN = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);

const defaultListOUT = profile2.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);

const defaultListTF = profile3.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const pageCount = computed({
  get() {
    return Math.ceil(data.countFulfillment / data.paging.pageSize);
  },
});

const fromUrlUI = computed({
  get() {
    return route.query.type == "Transfer"
      ? "/scm/inventory/journal/formconfig"
      : "/scm/inventory/journalmimo/formconfig";
  },
});

const getProfile = () => {
  if (route.query.type === "Movement In") {
    return profile;
  } else if (route.query.type === "Movement Out") {
    return profile2;
  } else if (route.query.type === "Transfer") {
    return profile3;
  }
  return null;
};

const getDefaultList = () => {
  if (route.query.type === "Movement In") {
    return defaultListIN;
  } else if (route.query.type === "Movement Out") {
    return defaultListOUT;
  } else if (route.query.type === "Transfer") {
    return defaultListTF;
  }
  return [];
};

const refInput = ref(null);
const listControl = ref(null);
const listControlFulfill = ref(null);
const lineConfig = ref(null);
const InventDimControlFrom = ref(null);
const InventDimControl = ref(null);
const listRejectionMessage = ref(null);
const FinancialDimension = ref(null);
const FinancialDimensionFrom = ref(null);
const route = useRoute();
const router = useRouter();

const axios = inject("axios");

const roleID = [
  (v) => {
    if (v == 0) return "required";
    return "";
  },
];

const deleteModal = ref(null);
let customFilter = computed(() => {
  let filters = [
    {
      Field: "Keyword",
      Op: "$eq",
      Value: data.search.Text,
    },
  ];
  if (data.search.Text !== null && data.search.Text !== "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "_id",
          Op: "$contains",
          Value: [data.search.Text],
        },
        {
          Field: "Text",
          Op: "$contains",
          Value: [data.search.Text],
        },
      ],
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
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  formcfg: {},
  gridCfg: {},
  gridCfgReffNo: {},
  stayOnForm: true,
  lockDimension: false,
  lockDimensionTf: false,
  lockPostingProfile: false,
  lockInventDimension: false,
  isDataList: true,
  compDimensionKey: "0",
  compInventDimensionKey: "0",
  keyJournalType: util.uuid(),
  dimensionFromKey: util.uuid(),
  titleForm: route.query.title || route.query.type,
  transactionType: route.query.type,
  transactionModul: route.query.modul,
  tabs: ["General"],
  journalTypeData: {},
  formInsert: "/scm/inventory/journal/save",
  formUpdate: "/scm/inventory/journal/save",
  disableDimTo: [],
  disableField: [],
  listReffNo: [],
  itemFulfillment: [],
  listChecked: [],
  countFulfillment: 0,
  isDialogReffNo: false,
  record: {
    _id: "",
    TrxDate: new Date(),
    Status: "",
    TrxType: "",
  },
  search: {
    Text: "",
    DateFrom: null,
    DateTo: null,
    Status: "",
    Site: "",
  },
  searchItem: {
    Keyword: "",
    Site: "",
    Warehouse: "",
    Status: "POSTED",
    FulfillmentType: "Item Transfer",
    Skip: 0,
    Take: 25,
  },
  paging: {
    skip: 0,
    pageSize: 25,
    currentPage: 1,
  },
  keyList: util.uuid(),
  disabledFormButton: false,
  deleteFn: () => {},
  preview: {},
});

function newRecord(record) {
  data.stayOnForm = true;
  data.formMode = "new";
  record._id = "";
  record.TrxDate = new Date();
  record.ETA = new Date();
  record.Status = "";
  record.TrxType = data.transactionType;
  record.JournalTypeID = "";
  record.PostingProfileID = "";
  record.ReffNo = [];
  record.Lines = [];
  record.InventDim = {
    WarehouseID: "",
    AisleID: "",
    SectionID: "",
    BoxID: "",
  };
  record.InventDimTo = {
    WarehouseID: "",
    AisleID: "",
    SectionID: "",
    BoxID: "",
  };
  data.titleForm = `Create new ${data.titleForm}`;

  data.lockDimension = false;
  data.lockPostingProfile = false;
  data.lockInventDimension = false;
  data.lockDimensionTf = false;
  data.disableDimTo = [];
  if (data.transactionType == "Transfer" && data.transactionModul == "Test") {
    data.searchItem = {
      Keyword: "",
      Site: "",
      Warehouse: "",
      Status: "POSTED",
      FulfillmentType: "Item Transfer",
      Skip: 0,
      Take: 25,
    };
    data.isDataList = false;
    // listControlFulfill.value.setLoading(true);
    getPostingProfile(record, () => {
      data.countFulfillment = 0;
      data.itemFulfillment = [];
      data.listChecked = [];
      data.record = record;
      // refreshDataFulfillment();
    });
  } else {
    getPostingProfile(record, () => {
      data.record = record;
      openForm(record);
    });
  }
}

function editRecord(record) {
  if (record.ReffNo == null) {
    record.ReffNo = [];
  }
  if (
    data.transactionType == "Transfer" &&
    record.Status == "DRAFT" &&
    data.transactionModul == "Test"
  ) {
    data.isDataList = false;
    listControlFulfill.value.setLoading(true);
  }
  data.stayOnForm = true;
  data.formMode = "edit";
  data.lockDimensionTf = false;
  data.titleForm = `Edit ${route.query.title} | ${record._id}`;
  if (record.JournalTypeID) {
    getJournalType(record.JournalTypeID, "init");
  }

  let disableDimTo = [];
  if (record.ReffNo?.length > 0) {
    getPostingProfile(record, () => {
      openForm(record, () => {
        record.Lines.map((r) => {
          r._id = r.IRDetailID;
          r.ItemVarian = helper.ItemVarian(r.ItemID, r.SKU);
          r.WarehouseID = [record.InventDim.WarehouseID]; // [r.InventDim.WarehouseID];
          return r;
        });
        data.record = record;
        if (record.InventDimTo.WarehouseID) {
          disableDimTo.push("WarehouseID");
        }
        if (record.InventDimTo.AisleID) {
          disableDimTo.push("AisleID");
        }
        if (record.InventDimTo.SectionID) {
          disableDimTo.push("SectionID");
        }
        if (record.InventDimTo.BoxID) {
          disableDimTo.push("BoxID");
        }
        if (["", "DRAFT"].includes(record.Status)) {
          record.ETA = new Date();
        }
        data.disableDimTo = disableDimTo;
        if (
          data.transactionType == "Transfer" &&
          record.Status == "DRAFT" &&
          data.transactionModul == "Test"
        ) {
          data.searchItem = {
            Keyword: "",
            Site: record.Dimension.find((d) => {
              return d.Key == "Site";
            }).Value,
            Warehouse: record.InventDim.WarehouseID,
            Status: "POSTED",
            FulfillmentType: "Item Transfer",
            Skip: 0,
            Take: 25,
          };

          listControlFulfill.value.setLoading(true);
          data.listChecked = record.Lines;
          refreshDataFulfillment();
        }
      });
    });
  } else {
    data.disableDimTo = disableDimTo;
    openForm(record, () => {
      data.record = record;
    });
  }
}

function deleteData(record) {
  data.deleteFn = () => {
    axios.post("/scm/inventory/journal/delete", record).then(
      (r) => {
        listControl.value.refreshGrid();
      },
      (e) => {
        util.showError(e);
      }
    );
  };

  deleteModal.value.show();
}

function confirmDelete() {
  deleteModal.value.hide();
  data.deleteFn();
}

function openForm(record, cbOK) {
  let type = data.transactionType;
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr(
      "_id",
      "hide",
      data.formMode == "new" ? true : false
    );
    listControl.value.setFormLoading(false);
    const el = document.querySelector(
      ".form_inputs > div.flex.section_group_container > div:nth-child(1) > div > div > div:nth-child(1)"
    );
    const el3 = document.querySelector(
      "div.suim_form.pt-2 > div > div.form_inputs > div.flex.section_group_container > div:nth-child(3) > div > div > div:nth-child(1)"
    );

    if (type == "Transfer") {
      el3.style.display = "block";
    } else {
      // el3.style.display = "none";
    }
    if (record._id || (type == "Transfer" && data.transactionModul == "Test")) {
      data.tabs = ["General", "Line"];
      if (el) {
        el.style.display = "block";
      }
    } else {
      data.tabs = ["General"];
      if (el) {
        el.style.display = "none";
      }
    }

    if (data.formMode == "edit") {
      if (type == "Transfer") {
        axios
          .post(`/tenant/warehouse/get`, [record.InventDim.WarehouseID])
          .then(
            (r) => {
              const siteFrom = r.data.Dimension.find((d) => {
                return d.Key == "Site";
              }).Value;

              record.DimensionFrom.map((d) => {
                if (d.Key == "Site") {
                  d.Value = siteFrom;
                }
                return d;
              });
              data.dimensionFromKey = util.uuid();
            },
            (e) => util.showError(e)
          );
      }

      if (
        ["SUBMITTED", "READY", "REJECTED", "POSTED"].includes(record.Status)
      ) {
        if (record.Status === "REJECTED") {
          listRejectionMessage.value.fetchRecord(record._id);
        }
        setFormMode("view");
      }
    }
    if (cbOK) {
      cbOK();
    }
  });
}
function changePageSize(pageSize) {
  data.paging.pageSize = pageSize;
  data.paging.currentPage = 1;
  refreshDataFulfillment();
}
function changePage(page) {
  data.paging.currentPage = page;
  refreshDataFulfillment();
}
function onCheckUncheckAll(checked) {
  let dt = listControlFulfill.value.getRecords();
  for (let i in dt) {
    onCheckUncheck(dt[i]);
  }
}
function onCheckUncheck(val) {
  const index = data.listChecked.findIndex((item) => item._id == val._id);
  if (index !== -1) {
    data.listChecked.splice(index, 1);
  }
  if (val.isSelected == true) {
    data.listChecked.push(val);
  }
  const lines = JSON.parse(JSON.stringify(data.listChecked)).map((l, idx) => {
    let isLine = data.record.Lines.find((r) => {
      return r._id === l._id;
    });

    if (isLine) {
      return isLine;
    } else {
      return {
        IRDetailID: l._id,
        ReffNo: l.ItemRequestID,
        BatchSerials: [],
        Dimension: l.ItemRequest.Dimension,
        InventDim: l.ItemRequest.InventDimTo,
        ItemID: l.ItemID,
        LineNo: idx,
        Qty: l.DetailLines.filter((d) => {
          return d.FulfillmentType == "Item Transfer";
        }).reduce((sum, item) => sum + item.QtyFulfilled, 0),
        RemainingQty: 0,
        Remarks: l.Remarks,
        SKU: l.SKU,
        Text: "",
        UnitCost: 0,
        UnitID: l.UoM,
        ItemVarian: l.ItemVarian,
        WarehouseID: l.WarehouseID,
      };
    }
  });
  data.record.Lines = lines;
}
function setLoadingForm(loading) {
  data.disabledFormButton = loading;
  listControl.value.setFormLoading(loading);
}

function getPostingProfile(record, cbOK, cbFalse) {
  let type = data.transactionType;
  if (type == "Transfer") {
    type = "Item Transfer";
  }
  util.nextTickN(2, () => {
    axios.post(`/scm/inventory/journal/type/find?TransactionType=${type}`).then(
      (r) => {
        if (r.data.length > 0) {
          data.keyJournalType = util.uuid();
          record.JournalTypeID = r.data[0]._id;
          record.PostingProfileID = r.data[0].PostingProfileID;
        }
        if (cbOK) {
          cbOK();
        }
      },
      (e) => {
        if (cbFalse) {
          cbFalse();
        }
        util.showError(e);
      }
    );
  });
}

function onSave(record) {
  data.stayOnForm = true;
  const isInventoryJournal = ["Movement In", "Movement Out"].includes(
    data.transactionType
  );
  const dim = isInventoryJournal ? record.InventDim : record.InventDimTo;

  let validate = true;
  let validateFrom = true;
  let validateInvFrom = true;
  let validateInvTo = true;
  if (InventDimControl.value) {
    validateInvTo = InventDimControl.value.validate();
  }
  if (InventDimControlFrom.value) {
    validateInvFrom = InventDimControlFrom.value.validate();
  }

  if (!record.Text) {
    return util.showError("General Name is required");
  }

  if (!dim?.WarehouseID) {
    return util.showError("General Warehouse ID is required");
  }

  const pc = record.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = record.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = record.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;
  if (FinancialDimension.value) {
    validate = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      validate = false;
    }
  }
  if (data.transactionType == "Transfer") {
    const pc = record.DimensionFrom.find((d) => {
      return d.Key == "PC";
    }).Value;
    const cc = record.DimensionFrom.find((d) => {
      return d.Key == "CC";
    }).Value;
    const site = record.DimensionFrom.find((d) => {
      return d.Key == "Site";
    }).Value;
    if (FinancialDimensionFrom.value) {
      validateFrom = FinancialDimensionFrom.value.validate();
    } else {
      if (!pc || !cc || !site) {
        validateFrom = false;
      }
    }
  }

  for (let l = 0; l < record?.Lines?.length; l++) {
    if (!record.Lines[l].ItemID) {
      return util.showError("line field Item required");
    }
    if (typeof record.Lines[l].Qty != "number") {
      return util.showError("line field Qty required");
    }
    if (typeof record.Lines[l].UnitCost != "number") {
      return util.showError("line field UnitCost required");
    }
    // if (
    //   record.Lines[l].UnitCost == 0 &&
    //   data.transactionType == "Movement In"
    // ) {
    //   return util.showError("UnitCost item more then 0");
    // }
    record.Lines[l].Dimension = record.Dimension;
  }
  const payload = JSON.parse(JSON.stringify(record));

  if (payload?.payload?.length > 0) {
    payload.Lines.map(function (v, idx) {
      v.LineNo = idx + 1;
      if (dim.WarehouseID) {
        v.InventDim.WarehouseID = dim.WarehouseID;
      }
      if (dim.AisleID) {
        v.InventDim.AisleID = dim.AisleID;
      }
      if (dim.SectionID) {
        v.InventDim.SectionID = dim.SectionID;
      }
      if (dim.BoxID) {
        v.InventDim.BoxID = dim.BoxID;
      }
      return v;
    });
  }

  if (record.Status == "") {
    payload.Status = "DRAFT";
  }

  if (
    validateInvTo &&
    validateInvFrom &&
    validate &&
    validateFrom &&
    listControl.value.formValidate()
  ) {
    setLoadingForm(true);
    payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
    payload.ETA = helper.dateTimeNow(payload.ETA);
    listControl.value.submitForm(
      payload,
      () => {
        data.tabs = ["General", "Line"];
        setLoadingForm(false);
        setTimeout(() => {
          data.isDataList = true;
        }, 500);
      },
      () => {
        setLoadingForm(false);
      }
    );
  } else {
    return util.showError("field is required");
  }
}

function onPreSave(record) {}

function onPostSave(record) {}

function preSubmit(status, action, doSubmit) {
  setLoadingForm(true);

  util.nextTickN(2, () => {
    data.stayOnForm = data.record.Status == "DRAFT" ? true : false;
    const isInventoryJournal = ["Movement In", "Movement Out"].includes(
      data.transactionType
    );
    const dim = isInventoryJournal
      ? data.record.InventDim
      : data.record.InventDimTo;

    if (status == "DRAFT") {
      if (!InventDimControl.value) {
        if (
          !data.record.Text ||
          !data.record.PostingProfileID ||
          !data.record.JournalTypeID
        ) {
          setLoadingForm(false);
          return util.showError(
            "General field Text, JournalTypeID, PostingProfileID  is required"
          );
        }
        if (!dim.WarehouseID) {
          setLoadingForm(false);
          return util.showError("General Warehouse ID is required");
        }
      }

      if (data?.record?.Lines?.length == 0) {
        setLoadingForm(false);
        return util.showError("Line items is empty");
      }

      for (let l = 0; l < data?.record?.Lines?.length; l++) {
        if (!data?.record.Lines[l].ItemID) {
          setLoadingForm(false);
          return util.showError("line field Item required");
        }
        if (typeof data?.record.Lines[l].Qty != "number") {
          setLoadingForm(false);
          return util.showError("line field Qty required");
        }
        if (typeof data?.record.Lines[l].UnitCost != "number") {
          setLoadingForm(false);
          return util.showError("line field UnitCost required");
        }
        if (
          data?.record.Lines[l].UnitCost == 0 &&
          data.transactionType == "Movement In"
        ) {
          setLoadingForm(false);
          return util.showError("UnitCost item more then 0");
        }
      }

      if (data?.record?.Lines?.length > 0) {
        data.record.Lines.map(function (v, idx) {
          v.LineNo = idx + 1;
          v.Dimension = data.record.Dimension;
          if (dim.WarehouseID) {
            v.InventDim.WarehouseID = dim.WarehouseID;
          }
          if (dim.AisleID) {
            v.InventDim.AisleID = dim.AisleID;
          }
          if (dim.SectionID) {
            v.InventDim.SectionID = dim.SectionID;
          }
          if (dim.BoxID) {
            v.InventDim.BoxID = dim.BoxID;
          }
          return v;
        });
      }

      setRequiredAllField(true);
      util.nextTickN(2, () => {
        let validateFrom = true;
        let valid = listControl.value.formValidate();
        const pc = data.record.Dimension.find((d) => {
          return d.Key == "PC";
        }).Value;
        const cc = data.record.Dimension.find((d) => {
          return d.Key == "CC";
        }).Value;
        const site = data.record.Dimension.find((d) => {
          return d.Key == "Site";
        }).Value;
        if (isInventoryJournal && FinancialDimension.value) {
          valid = FinancialDimension.value.validate();
        } else {
          if (isInventoryJournal && (!pc || !cc || !site)) {
            valid = false;
          }
        }

        if (data.transactionType == "Transfer") {
          const pc = data.record.DimensionFrom.find((d) => {
            return d.Key == "PC";
          }).Value;
          const cc = data.record.DimensionFrom.find((d) => {
            return d.Key == "CC";
          }).Value;
          const site = data.record.DimensionFrom.find((d) => {
            return d.Key == "Site";
          }).Value;
          if (FinancialDimensionFrom.value) {
            validateFrom = FinancialDimensionFrom.value.validate();
          } else {
            if (!pc || !cc || !site) {
              validateFrom = false;
            }
          }
        }

        if (valid && validateFrom) {
          setLoadingForm(true);
          data.record.TrxDate = helper.dateTimeNow(data.record.TrxDate);
          data.record.ETA = helper.dateTimeNow(data.record.ETA);
          listControl.value.submitForm(
            data.record,
            () => {
              doSubmit();
            },
            () => {
              setLoadingForm(false);
            }
          );
        } else {
          setLoadingForm(false);
          util.showError("field is required");
        }
        setRequiredAllField(false);
      });
    }
  });
}
function postSubmit(record) {
  setLoadingForm(false);
  data.appMode = "grid";
  listControl.value.refreshList();
  listControl.value.refreshForm();
  listControl.value.setControlMode("grid");
}
function preReopen() {
  let payload = JSON.parse(JSON.stringify(data.record));
  payload.Status = "DRAFT";
  axios.post("/scm/inventory/journal/save", payload).then(
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

function setRequiredAllField(required) {
  listControl.value.getFormAllField().forEach((e) => {
    if (!["PostingProfileID"].includes(e.field)) {
      listControl.value.setFormFieldAttr(e.field, "required", required);
    }
  });
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
  // if (data.transactionType !== "Transfer") {
  //   cfg.fields.map(function (el) {
  //     if (el.field == "ETA" || el.field == "DeliveryService" || el.field == "ReffNo") {
  //       el.readType = "hide";
  //     }
  //     return el;
  //   });
  // }
  cfg.fields.map(function (el) {
    if (
      data.transactionType == "Movement In" &&
      (el.field == "ETA" ||
        el.field == "DeliveryService" ||
        el.field == "ReffNo")
    ) {
      el.readType = "hide";
    } else if (
      data.transactionType == "Movement Out" &&
      (el.field == "ETA" || el.field == "DeliveryService")
    ) {
      el.readType = "hide";
    } else if (el.field == "KoliKilo") {
      el.readType = "hide";
    }
    return el;
  });
  if (data.transactionType === "Transfer") {
    const newFields = [
      {
        field: "WarehouseFrom",
        kind: "Text",
        label: "Warehouse From",
        readType: "show",
        input: {
          field: "WarehouseFrom",
          label: "Warehouse From",
          hint: "",
          hide: false,
          placeHolder: "Warehouse From",
          kind: "text",
          disable: false,
          required: false,
          multiple: false,
        },
      },
      {
        field: "WarehouseTo",
        kind: "Text",
        label: "Warehouse To",
        readType: "show",
        input: {
          field: "WarehouseTo",
          label: "Warehouse To",
          hint: "",
          hide: false,
          placeHolder: "Warehouse To",
          kind: "text",
          disable: false,
          required: false,
          multiple: false,
        },
      },
      {
        field: "Approvers",
        kind: "Text",
        label: "Next Approver",
        readType: "show",
        input: {
          field: "Approvers",
          label: "Next Approver",
          hint: "",
          hide: false,
          placeHolder: "Next Approver",
          kind: "text",
          disable: false,
          required: false,
          multiple: false,
        },
      },
    ];
    cfg.fields = [...cfg.fields, ...newFields];
  } else {
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
  }
  let sortColm = [
    "_id",
    "Text",
    "TrxDate",
    "ETA",
    "WarehouseID",
    "WarehouseTo",
    "WarehouseFrom",
    "DeliveryService",
    "ReffNo",
    "Status",
    "Approvers",
  ];
  cfg.fields.map((f) => {
    if (f.field == "Text") {
      f.label = "Name";
    }
    f.idx = sortColm.indexOf(f.field);
    return f;
  });
  cfg.fields.sort((a, b) => (a.idx > b.idx ? 1 : -1));
}

function alterFormConfig(cfg) {
  cfg.sectionGroups = cfg.sectionGroups.map((sectionGroup) => {
    sectionGroup.sections = sectionGroup.sections.map((section) => {
      section.rows.map((row) => {
        row.inputs = row.inputs
          .filter((input) => ["Enable"].indexOf(input.field) == -1)
          .map((input) => {
            switch (data.transactionType) {
              case "Movement Out":
                if (
                  input.field === "InventDimTo" ||
                  input.field === "ETA" ||
                  input.field === "DeliveryService" ||
                  input.field === "DimensionFrom" ||
                  input.field === "KoliKilo"
                ) {
                  input.hide = true;
                }
                break;
              case "Movement In":
                if (
                  input.field === "InventDimTo" ||
                  input.field === "ETA" ||
                  input.field === "DeliveryService" ||
                  input.field === "DimensionFrom" ||
                  input.field === "KoliKilo"
                ) {
                  input.hide = true;
                }
                break;
              case "Transfer":
                if (
                  input.field === "Dimension" ||
                  input.field === "DimensionFrom"
                ) {
                  input.hide = false;
                }
                break;
              default:
                break;
            }
            if (input.field == "Text") {
              input.label = "Name";
            }
            return input;
          });
        return row;
      });
      return section;
    });
    return sectionGroup;
  });
  if (route.query.trxid !== undefined) {
    let currQuery = { ...route.query };
    listControl.value.selectData({ _id: currQuery.trxid }); //remark sementara tunggu suimjs update
    delete currQuery["trxid"];
    router.replace({ path: route.path, query: currQuery });
  } else if (route.query.id !== undefined) {
    let getUrlParam = route.query.id;
    listControl.value.selectData({ _id: getUrlParam });
    router.replace({
      query: {
        type: route.query.type,
        title: route.query.title,
      },
    });
  }
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
      switch (data.transactionType) {
        case "Movement Out":
          data.record.InventDim = JT.InventoryDimension;
          break;
        case "Movement in":
          data.record.InventDimTo = JT.InventoryDimension;
          break;
        case "Stock Opname":
          data.record.InventDimTo = JT.InventoryDimension;
          break;
        default:
          data.record.InventDim = JT.InventoryDimension;
          break;
      }
      data.record.Dimension = JT.Dimension;
      data.record.PostingProfileID = JT.PostingProfileID;
      nextTick(() => {
        data.compDimensionKey = (
          parseInt(data.compDimensionKey) + 1
        ).toString();
        data.compInventDimensionKey = (
          parseInt(data.compInventDimensionKey) + 1
        ).toString();
      });
    }
  }
}

function onFieldChanged(val, item) {
  if (lineConfig.value && item.Lines.length > 0) {
    util.nextTickN(2, () => {
      const line = lineConfig.value.getDataValue();
      if (line.length > 0) {
        for (let l = 0; l < line.length; l++) {
          line[l].InventDim = val;
        }
      }
      if (val.WarehouseID) {
        data.disableField.push("WarehouseID");
        // if (typeof val.WarehouseID == "string") {
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
    });
  }
}

function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}

function getReffNo() {
  data.listReffNo = [];
  axios
    .post("/scm/postingprofile/find-journal-ref", {
      JournalType: route.query.type,
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
    data.gridCfgReffNo = helper.generateGridCfg(colum);
  });
}
function onFilterRefresh(val) {
  data.paging = {
    skip: 0,
    pageSize: 25,
    currentPage: 1,
  };
  util.nextTickN(2, () => {
    refreshDataFulfillment();
  });
}
function refreshDataFulfillment() {
  let payload = {
    ...data.searchItem,
    Skip: (data.paging.currentPage - 1) * data.paging.pageSize,
    Take: data.paging.pageSize,
  };

  if (data.listChecked.length > 0) {
    payload.SortPurchaseRequestIDs = [
      ...new Set(
        JSON.parse(JSON.stringify(data.listChecked)).map((i) => {
          return i.ReffNo;
        })
      ),
    ];
  }

  listControlFulfill.value.setLoading(true);
  axios.post("/scm/item/request/detail/get-lines", payload).then(
    (r) => {
      r.data.data.map(function (i) {
        let isCheck = data.record.Lines.find(function (v) {
          return v.IRDetailID == i._id;
        });
        i.isSelected = isCheck ? true : false;
        i.WarehouseID = i.DetailLines.map((d) => {
          return d.InventDimFrom.WarehouseID;
        });
        i.ItemVarian = helper.ItemVarian(i.ItemID, i.SKU);
        i.ReffNo = i.ItemRequestID;
        i.Qty = i.DetailLines.filter((d) => {
          return d.FulfillmentType == "Item Transfer";
        }).reduce((sum, item) => sum + item.QtyFulfilled, 0);
      });
      setTimeout(() => {
        data.countFulfillment = r.data.count;
        data.itemFulfillment = r.data.data;
        listControlFulfill.value.setLoading(false);
      }, 500);
    },
    (e) => {
      util.showError(e);
    }
  );
}

function onProcess(val) {
  let wh = [];
  let dim = [];
  for (let i in data.record.Lines) {
    for (let idx = 0; idx < data.record.Lines[i].WarehouseID.length; idx++) {
      if (!wh.includes(data.record.Lines[i].WarehouseID[idx])) {
        wh.push(data.record.Lines[i].WarehouseID[idx]);
      }
    }
  }

  for (let i in data.record.Lines) {
    let strDim = "";
    if (
      data.record.Lines[i].Dimension == null ||
      data.record.Lines[i].Dimension == undefined ||
      data.record.Lines[i].Dimension.length == 0
    ) {
      strDim = "";
    } else {
      strDim = data.record.Lines[i].Dimension.filter(
        (el) => el.Value != "" && el.Value != null && el.Key == "Site"
      )
        .map((el) => {
          return el.Key + "=" + (el.Value == "" ? "*" : el.Value);
        })
        .join(", ");
    }
    if (!dim.includes(strDim)) {
      dim.push(strDim);
    }
  }
  if (wh.length !== 1 || dim.length !== 1) {
    if (wh.length > 1 || dim.length > 1) {
      util.showError("Please select the same warehouse and Dimension");
    } else {
      openForm(data.record, () => {
        data.isDataList = true;
        if (!data.record.Status) {
          data.record.TrxDate = new Date();
          data.tabs = ["General"];
        } else {
          data.tabs = ["General", "Line"];
        }
      });
    }
  } else {
    axios.post("/tenant/warehouse/get", wh).then((r) => {
      data.isDataList = true;
      openForm(data.record, () => {
        data.record.DimensionFrom = r.data.Dimension;
        data.record.Dimension = JSON.parse(
          JSON.stringify(data.record.Lines[0].Dimension)
        );
        data.record.InventDimTo = JSON.parse(
          JSON.stringify(data.record.Lines[0].InventDim)
        );
        data.record.InventDim = {
          WarehouseID: wh[0],
          SectionID: "",
          AisleID: "",
          BoxID: "",
        };
        data.record.Dimension.map((dimTo) => {
          if (dimTo.Key !== "Site") {
            dimTo.Value = "";
          }
          return dimTo;
        });
        data.record.ReffNo = [
          ...new Set(
            data.record.Lines.map((r) => {
              return r.ReffNo;
            })
          ),
        ];
        if (!Status) {
          data.record.TrxDate = new Date();
        }

        setTimeout(() => {
          lineConfig.value.setDataValue(data.record.Lines);
        }, 500);
      });
    });
  }
}

function onBack(val) {
  util.nextTickN(2, () => {
    data.isDataList = true;
    listControl.value.setControlMode("grid");
    listControl.value.refreshList();
  });
}

watch(
  () => route.query.type,
  (nv) => {
    data.transactionType = route.query.type;
    data.keyList = util.uuid();
    util.nextTickN(2, () => {
      data.appMode = "grid";
      if (listControl.value) {
        listControl.value.refreshList();
        listControl.value.refreshForm();
        listControl.value.setControlMode("grid");
      }
    });
  }
);

watch(
  () => route.query.title,
  (nv) => {
    data.titleForm = nv;
  }
);
onMounted(() => {
  createGridCfgRefNo();
  loadGridConfig(axios, "/scm/item/request/detail/gridconfig").then(
    (r) => {
      let cfg = r;
      let fields = [
        "ItemID",
        "QtyFulfilled",
        "QtyAvailable",
        "UoM",
        "ItemRequestID",
        "ItemType",
        "Remarks",
        "WarehouseID",
        "Dimension",
      ];
      cfg.fields = cfg.fields.map((el) => {
        if (["QtyRequested", "QtyAvailable"].includes(el.field)) {
          el.width = "130px";
        }
        if (["UoM"].includes(el.field)) {
          el.width = "70px";
        }
        if (["QtyFulfilled"].includes(el.field)) {
          el.width = "150px";
        }
        if (["Remarks"].includes(el.field)) {
          el.width = "400px";
        }
        if (["WarehouseID"].includes(el.field)) {
          el.label = "From Warehouse";
          el.width = "150px";
        }
        if (["ItemRequestID"].includes(el.field)) {
          el.label = "Item Request";
          el.width = "170px";
        }
        el.idx = fields.indexOf(el.field);
        return {
          ...el,
          input: {
            ...el.input,
            readOnly: false,
          },
        };
      });
      cfg.fields = cfg.fields
        .filter((el) => fields.includes(el.field))
        .sort((a, b) => (a.idx > b.idx ? 1 : -1));

      util.nextTickN(2, () => {
        data.gridCfg = cfg;
      });
    },
    (e) => util.showError(e)
  );
});

function onPreview() {
  data.appMode = "preview";
}

function closePreview() {
  data.appMode = "grid";
}
</script>
