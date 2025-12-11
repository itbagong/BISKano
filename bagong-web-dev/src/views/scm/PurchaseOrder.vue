<template>
  <div class="w-full">
    <s-card
      v-show="data.appMode != 'preview'"
      :title="`${titleForm}`"
      class="w-full bg-white suim_datalist card"
      hide-footer
      :no-gap="false"
      :hide-title="false"
    >
      <s-grid
        v-if="data.step == 'list-po'"
        ref="listPOControl"
        class="w-full"
        :hide-new-button="!profile.canCreate"
        :hide-delete-button="!profile.canDelete"
        :hide-detail="!profile.canUpdate"
        :hide-action="false"
        :hide-select="true"
        :config="data.gridCfgTrx"
        sort-field="LastUpdate"
        auto-commit-line
        no-confirm-delete
        read-url="/scm/purchase/order/gets-v1"
        delete-url="/scm/purchase/order/delete"
        :custom-filter="customFilter"
        form-keep-label
        @selectData="onSelectData"
        @newData="newRecord"
      >
        <template #header_search="{ config }">
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
            :items="[
              'DRAFT',
              'SUBMITTED',
              'READY',
              'POSTED',
              'REJECTED',
              'CANCELED',
            ]"
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
        <template #item_Status="{ item }">
          <status-text :txt="item.IsCanceled ? 'CANCELED' : item.Status" />
        </template>
        <template #item_Approvers="{ item }">
          <list-approvers :approvers="item.Approvers" />
        </template>
        <template #item_ReffNo="{ item }">
          {{ item.ReffNo?.join() }}
        </template>
        <template #item_WarehouseID="{ item }">
          {{ item.Location.WarehouseID }}
        </template>
        <template #item_buttons_1="{ item }">
          <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
          <div
            class="px-1"
            v-if="item.Status == 'POSTED' && item.TotalPrint > 0"
          >
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
        <template #item_button_delete="{ item }">
          <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
        </template>
      </s-grid>
      <s-grid
        v-else-if="data.step == 'list-pr'"
        ref="listPRControl"
        class="w-full grid-line-items"
        hide-sort
        :hide-new-button="true"
        :hide-delete-button="true"
        :hide-detail="true"
        :hide-action="true"
        :hide-select="false"
        :config="data.gridCfgPrline"
        auto-commit-line
        no-confirm-delete
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
              defaultList?.length > 0
                ? (...args) =>
                    helper.payloadBuilderDimension(
                      defaultList,
                      data.searchItem.Site,
                      false,
                      ...args
                    )
                : undefined
            "
            @change="
              () => {
                data.searchItem.DeliveryTo = '';
                data.searchItem.VendorID = '';
                data.searchItem.Requestor = '';
                onChangeFilter();
              }
            "
          />
          <s-input
            v-if="false"
            label="Warehouse"
            v-model="data.searchItem.DeliveryTo"
            class="w-full"
            use-list
            lookup-url="/tenant/warehouse/find"
            lookup-key="_id"
            :lookup-labels="['_id', 'Name']"
            :lookup-searchs="['_id', 'Name']"
            :lookup-payload-builder="
              (search) =>
                lookupPayloadBuilderSite(
                  search,
                  ['_id', 'Name'],
                  data.searchItem.DeliveryTo,
                  data.searchItem
                )
            "
            @change="
              () => {
                onChangeFilter();
              }
            "
          ></s-input>
          <s-input
            ref="refInput"
            label="Vendor"
            v-model="data.searchItem.VendorID"
            class="w-full"
            use-list
            :lookup-url="`/tenant/vendor/find`"
            lookup-key="_id"
            :lookup-labels="['_id', 'Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="
              () => {
                onChangeFilter();
              }
            "
          ></s-input>
          <s-input
            ref="refInput"
            label="Requestor"
            v-model="data.searchItem.Requestor"
            class="w-full"
            use-list
            lookup-key="_id"
            :lookup-url="`/tenant/employee/find`"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            :lookup-payload-builder="
              (search) =>
                lookupPayloadSearch(
                  search,
                  ['_id', 'Name'],
                  data.searchItem.Requestor,
                  data.searchItem
                )
            "
            @change="
              () => {
                onChangeFilter();
              }
            "
          ></s-input>
          <s-input
            ref="refInput"
            label="Reff No"
            :hideLabel="false"
            v-model="data.searchItem.ReffNo"
            class="w-full"
          ></s-input>
        </template>
        <template #header_buttons="{ config }">
          <s-button
            icon="refresh"
            class="btn_primary refresh_btn"
            :disabled="data.loading.processItem"
            @click="onRefreshPRLine"
          />
        </template>
        <template #header_buttons_2="{ config }">
          <s-button
            label="Process"
            class="btn_primary refresh_btn"
            :disabled="data.loading.processItem"
            @click="onProcess"
          />
          <s-button
            icon="rewind"
            label="Back"
            class="btn_warning back_btn"
            :disabled="data.loading.processItem"
            @click="onBackListPO"
          />
        </template>
        <template #item_ItemID="{ item }">
          <div class="bg-transparent">{{ item.Text }}</div>
        </template>
        <template #item_UnitID="{ item }">
          <div class="bg-transparent">{{ item.UnitID }}</div>
        </template>
        <template #item_Dimension="{ item }">
          <DimensionText :dimension="item.Dimension" />
        </template>
        <template #item_WarehouseID="{ item }">
          <div class="bg-transparent">{{ item.InventDim.WarehouseID }}</div>
        </template>
        <template #paging>
          <s-pagination
            :recordCount="data.itemsPRline.length"
            :pageCount="pageCountItem"
            :current-page="data.pagingItems.currentPage"
            :page-size="data.pagingItems.pageSize"
            @changePage="changePageItem"
            @changePageSize="changePageSizeItem"
          ></s-pagination>
        </template>
      </s-grid>
      <s-form
        v-else-if="data.step == 'form-po'"
        id="form-po-control"
        ref="formPOControl"
        v-model="data.records"
        :keep-label="true"
        :config="data.frmCfg"
        class="pt-2"
        :auto-focus="true"
        :hide-submit="true"
        :hide-cancel="false"
        :tabs="
          data.records._id
            ? ['General', 'Line', 'Attachment']
            : data.records.ReffNo?.length > 0
            ? ['General', 'Line']
            : ['General']
        "
        :mode="data.formMode"
        @fieldChange="onFieldChange"
        @cancelForm="onBackListPO"
      >
        <template #tab_Line="{ item }">
          <PurchaseLine
            ref="lineConfig"
            v-model="item.Lines"
            :general-record="item"
            purchase-type="purchase/order"
            :disable-field="data.disableField"
            :islog="true"
            :isReff="item.ReffNo?.length > 0"
            @attachment-action="
              (attachment) => {
                item.Attachment = attachment;
              }
            "
          ></PurchaseLine>
        </template>
        <template #tab_Attachment="{ item }">
          <s-grid-attachment
            :key="item._id"
            :journal-id="item._id"
            :tags="linesTag"
            :isUpdateTags="reffTags.length > 0 ? true : false"
            :add-Tags="addTags"
            :reff-tags="reffTags"
            journal-type="Purchase Order"
            ref="gridAttachment"
            @pre-Save="preSaveAttachment"
            v-model="item.Attachment"
          ></s-grid-attachment>
        </template>
        <template #input_Requestor="{ item, config }">
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
        <template #input_JournalTypeID="{ item }">
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
            :lookup-url="`/scm/purchase/order/journal/type/find`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
          ></s-input>
        </template>
        <template #input_VendorID="{ item }">
          <s-input
            ref="refVendor"
            v-model="item.VendorID"
            label="Vendor"
            field="VendorID"
            class="w-full"
            use-list
            :disabled="!['DRAFT', ''].includes(item.Status)"
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
                onFieldChange(field, v1, v2, old);
              }
            "
          ></s-input>
        </template>
        <template #input_DueDate="{ item, config }">
          <s-input
            ref="refInput"
            label="Due Date"
            v-model="item.DueDate"
            class="w-full"
            kind="date"
            keepLabel
            :disabled="true"
          ></s-input>
        </template>
        <template #input_TaxType="{ item }">
          <div class="flex gap-[10px]">
            <s-input
              ref="refInput"
              label="Tax Registration"
              v-model="item.TaxRegistration"
              class="w-full"
              keepLabel
              :hide-label="false"
              :read-only="true"
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
        <template #input_TaxCodes="{ item }">
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
        <template #input_Status="{ item }">
          <div>
            <label class="input_label"><div>Status</div></label>
            <div>
              {{
                item.Status === "POSTED" && item.IsCanceled === true
                  ? "CANCELED"
                  : item.Status
              }}
            </div>
          </div>
        </template>
        <template #input_PIC="{ item, config }">
          <s-input
            :key="data.keyPIC"
            ref="refInput"
            label="PIC"
            v-model="item.PIC"
            class="w-full"
            :required="true"
            :disabled="
              ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
            "
            use-list
            :lookup-url="`/tenant/employee/find`"
            lookup-key="_id"
            :lookup-labels="['_id', 'Name']"
            :lookup-searchs="['_id', 'Name']"
            :lookup-payload-builder="
              (search) =>
                lookupPayloadBuilderDimension(
                  search,
                  ['_id', 'Name'],
                  item.PIC,
                  item
                )
            "
          ></s-input>
        </template>
        <template #input_Dimension="{ item }">
          <dimension-editor
            ref="FinancialDimension"
            :key="data.keyDimension"
            v-model="item.Dimension"
            sectionTitle="Financial Dimension"
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
        <template #input_Location="{ item }">
          <dimension-invent-jurnal
            ref="InventDimControl"
            :mandatory="['WarehouseID']"
            :key="data.keyDimension"
            v-model="item.Location"
            :default-list="profile.Dimension"
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
            @onChange="
              (field, v1, v2, old, val) => {
                onFieldChanged(val, item);
              }
            "
          ></dimension-invent-jurnal>
        </template>
        <template #input_BillingName="{ item }">
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
                onFieldChange(field, v1, v2, old);
              }
            "
          ></s-input>
        </template>
        <template #footer_1="{ item }">
          <RejectionMessageList
            ref="listRejectionMessage"
            journalType="Purchase Order"
            :journalID="item._id"
          ></RejectionMessageList>
        </template>
        <template #buttons_1="{ item }">
          <div class="flex gap-1">
            <s-button
              class="bg-transparent hover:bg-blue-500 hover:text-black"
              label="Preview"
              icon="eye-outline"
              :disabled="data.loading.processItem"
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
              v-if="data.isClose && !item.IsCanceled"
              :disabled="data.loading.processItem"
              class="btn_primary submit_btn"
              label="Cancel"
              icon="close-box"
              tooltip="Cancel"
              @click="onCancel(item)"
            ></s-button>
            <s-button
              v-if="
                !['SUBMITTED', 'POSTED', 'READY', 'REJECTED'].includes(
                  item.Status
                )
              "
              :icon="`content-save`"
              class="btn_primary submit_btn"
              :label="'Save'"
              :disabled="data.loading.processItem"
              @click="onSave(item)"
            />
            <form-buttons-trx
              v-if="item.Lines.length > 0"
              :key="data.btnPostId"
              :status="item.Status"
              :moduleid="`scm/new`"
              :autoPost="false"
              :autoReopen="false"
              :journal-id="item._id"
              :posting-profile-id="item.PostingProfileID"
              :disabled="data.loading.processItem"
              journal-type-id="Purchase Order"
              @pre-submit="preSubmit"
              @pre-reopen="preReopen"
              @post-submit="postSubmit(item)"
              @error-submit="errorSubmit"
            />
          </div>
        </template>
      </s-form>
    </s-card>
    <PreviewReport
      v-if="data.appMode == 'preview'"
      class="card w-full"
      title="Purchase Order"
      :preview="data.preview"
      @close="closePreview"
      @update-print="onUpdatePrint"
      SourceType="Purchase Order"
      :SourceJournalID="data.records._id"
      :hideSignature="false"
    >
    </PreviewReport>
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
              class="w-full"
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
  vModelCheckbox,
} from "vue";
import { layoutStore } from "@/stores/layout.js";
import { useRouter, useRoute } from "vue-router";
import {
  loadGridConfig,
  loadFormConfig,
  SPagination,
  util,
  SInput,
  SButton,
  SForm,
  SGrid,
  SCard,
  SModal,
} from "suimjs";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import PurchaseLine from "./widget/PurchaseLine.vue";
import ListApprovers from "@/components/common/ListApprovers.vue";
import StatusText from "@/components/common/StatusText.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import moment from "moment";
import helper from "@/scripts/helper.js";
import RejectionMessageList from "./widget/RejectionMessageList.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import { authStore } from "@/stores/auth";

layoutStore().name = "tenant";
const featureID = "PurchaseOrder";
const profile = authStore().getRBAC(featureID);
const headOffice = layoutStore().headOfficeID;
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const auth = authStore();
const listPOControl = ref(null);
const listPRControl = ref(null);
const formPOControl = ref(null);
const FinancialDimension = ref(null);
const InventDimControl = ref(null);
const lineConfig = ref(null);
const gridAttachment = ref(SGridAttachment);

const axios = inject("axios");
const route = useRoute();
const router = useRouter();

let customFilter = computed(() => {
  let filters = [
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
    if (data.search.Status == "CANCELED") {
      filters.push({
        Field: "IsCanceled",
        Op: "$eq",
        Value: true,
      });
    } else {
      filters.push(
        {
          Field: "Status",
          Op: "$eq",
          Value: data.search.Status,
        },
        {
          Field: "IsCanceled",
          Op: "$ne",
          Value: true,
        }
      );
    }
  }

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});

const linesTag = computed({
  get() {
    let ReffNo = JSON.parse(
      JSON.stringify(data.records.ReffNo ? data.records.ReffNo : [])
    ).map((ref) => {
      return `${ref.slice(0, 2)}_${ref}`;
    });

    const tags =
      ReffNo && data.records._id
        ? [...[`PO_${data.records._id}`], ...ReffNo]
        : data.records._id
        ? [`PO_${data.records._id}`]
        : ReffNo;

    return tags;
  },
});

const addTags = computed({
  get() {
    return [`PO_${data.records._id}`];
  },
});

const reffTags = computed({
  get() {
    let ReffNo = JSON.parse(
      JSON.stringify(data.records.ReffNo ? data.records.ReffNo : [])
    ).map((ref) => {
      return `${ref.slice(0, 2)}_${ref}`;
    });
    return ReffNo;
  },
});

const titleForm = computed({
  get() {
    let PO = JSON.parse(JSON.stringify(data.records));
    let title = "Purchase Order";
    if (data.step == "list-po") {
      title = `Purchase Order`;
    } else if (PO._id && data.step == "list-pr") {
      title = `Edit Purchase Order | ${PO._id}`;
    } else if (!PO._id && data.step == "list-pr") {
      title = `Create Purchase Order`;
    } else if (PO._id && data.step == "form-po") {
      title = `Edit Purchase Order | ${PO._id}`;
    } else if (!PO._id && data.step == "form-po") {
      title = `Create Purchase Order`;
    }
    return title;
  },
});

const pageCountItem = computed({
  get() {
    return Math.ceil(data.countPRline / data.paging.pageSize);
  },
});

const data = reactive({
  isPreview: false,
  lockDimension: false,
  lockPostingProfile: false,
  lockInventDimension: false,
  isDialogReffNo: false,
  isClose: false,
  step: "list-po", // list-po, list-pr, form-po
  appMode: "grid",
  value: [],
  itemsPRline: [],
  itemsPOline: [],
  listChecked: [],
  disableField: [],
  listReffNo: [],
  tabs: ["General"],
  gridCfgTrx: {},
  gridCfgPrline: {},
  gridCfgLine: {},
  frmCfg: {},
  preview: {},
  gridCfgReffNo: {},
  listTTD: [],
  records: {
    Lines: [],
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
  searchItem: {
    VendorID: "",
    DeliveryTo: "",
    Site: "",
    ReffNo: "",
    Requestor: "",
  },
  loading: {
    processItem: false,
  },
  record: {
    _id: "",
    TrxDate: new Date(),
    DocumentDate: new Date(),
    PRDate: new Date(),
    PODate: new Date(),
    DueDate: new Date(),
    DeliveryDate: new Date(),
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
  keyPIC: util.uuid(),
  keyJournalType: util.uuid(),
  keyDimension: util.uuid(),
  btnPostId: util.uuid(),
  keyRequestor: util.uuid(),
  keyTaxCode: util.uuid(),
  sortDirection: "asc",
  formMode: "new",
  siteUser: "",
  currentUser: "",
  requestorName: "",
  countPRline: 0,
  paging: {
    skip: 0,
    pageSize: 20,
    currentPage: 1,
  },
  pagingItems: {
    skip: 0,
    pageSize: 20,
    currentPage: 1,
  },
});
function newRecord() {
  data.step = "list-pr";
  data.formMode = "new";
  const record = {};
  record._id = "";
  record.CompanyID = auth.companyId;
  record.Requestor = "";
  record.TrxDate = new Date();
  record.DocumentDate = new Date();
  record.PRDate = new Date();
  record.PODate = new Date();
  record.DueDate = new Date();
  record.DeliveryDate = new Date();
  record.Status = "";
  record.BillingName = "";
  record.TrxType = data.transactionType;
  record.Freight = 0;
  record.ReffNo = [];
  record.Location = {
    WarehouseID: "",
  };
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
      Value: "",
    },
    {
      Key: "Asset",
      Value: "",
    },
  ];
  record.Lines = [];
  record.Discount = {
    DiscountType: "percent",
    DiscountValue: 0,
    DiscountAmount: 0,
  };
  data.listChecked = [];
  data.titleForm = `Create New Purchase Order`;
  data.records = record;
  data.lockDimension = false;
  data.lockPostingProfile = false;
  data.lockInventDimension = false;
  data.searchItem = {
    VendorID: "",
    DeliveryTo: "",
    Site: "",
    ReffNo: "",
    Requestor: "",
  };
}
function changePageItem(page) {
  console.log("---------------------")
  data.pagingItems.currentPage = page;
  onGeneratePRLine(
    data.searchItem.Site,
    data.searchItem.DeliveryTo,
    data.searchItem.VendorID,
    data.searchItem.ReffNo,
    data.searchItem.Requestor
  );
}
function changePageSizeItem(pageSize) {
  data.paging = {
    skip: 0,
    pageSize: pageSize,
    currentPage: 1,
  }
  data.pagingItems= {
    skip: 0,
    pageSize: pageSize,
    currentPage: 1,
  }
  onGeneratePRLine(
    data.searchItem.Site,
    data.searchItem.DeliveryTo,
    data.searchItem.VendorID,
    data.searchItem.ReffNo,
    data.searchItem.Requestor
  );
}
function openForm(record) {
  util.nextTickN(2, () => {
    data.fileAttach = {};
    data.isPostFlow = false;
    formPOControl.value.setFieldAttr(
      "_id",
      "hide",
      data.formMode == "new" ? true : false
    );
    const el = document.querySelector(
      "#po-data-list .form_inputs > div.flex.section_group_container > div:nth-child(1) > div:nth-child(1) > div.flex.flex-col.gap-4 > div:nth-child(1)"
    );
    data.tabs = ["General", "Line", "Attachment"];
    if (record._id == "") {
      el ? (el.style.display = "none") : "";
      getPostingProfile(record);
      if (record.Location) {
        onFieldChanged(record.Location, record);
      }

      data.keyDimension = util.uuid();
      if (record.VendorID) {
        formPOControl.value.setFieldAttr("TaxName", "readOnly", true);
        formPOControl.value.setFieldAttr("PaymentTerms", "readOnly", true);
        formPOControl.value.setFieldAttr("TaxType", "readOnly", true);
        formPOControl.value.setFieldAttr("TaxCodes", "readOnly", true);
        formPOControl.value.setFieldAttr("TaxAddress", "readOnly", true);
      }
    } else {
      el ? (el.style.display = "block") : "";
      util.nextTickN(2, () => {
        let tabs = document.querySelector(".tab_container > div");
        tabs.addEventListener("click", function (event) {
          setMinDate(record);
        });
        if (record.VendorID) {
          formPOControl.value.setFieldAttr("TaxName", "readOnly", true);
          formPOControl.value.setFieldAttr("PaymentTerms", "readOnly", true);
          formPOControl.value.setFieldAttr("TaxType", "readOnly", true);
          formPOControl.value.setFieldAttr("TaxCodes", "readOnly", true);
          formPOControl.value.setFieldAttr("TaxAddress", "readOnly", true);
        } else {
          formPOControl.value.setFieldAttr("TaxName", "readOnly", false);
          formPOControl.value.setFieldAttr("PaymentTerms", "readOnly", false);
          formPOControl.value.setFieldAttr("TaxType", "readOnly", false);
          formPOControl.value.setFieldAttr("TaxCodes", "readOnly", false);
          formPOControl.value.setFieldAttr("TaxAddress", "readOnly", false);
        }
      });
    }
    setMinDate(record);
    data.records = record;
    formPOControl.value.setLoading(false);
  });
}
function setMinDate(record) {
  const el = document.querySelector('input[placeholder="PO date"]');
  if (el) {
    el.setAttribute("min", moment(record.PRDate).format("YYYY-MM-DD"));
  }
  const ell = document.querySelector('input[placeholder="Due date"]');
  if (ell) {
    ell.setAttribute("min", moment(record.PODate).format("YYYY-MM-DD"));
  }
  const elll = document.querySelector('input[placeholder="Delivery date"]');
  if (elll) {
    elll.setAttribute("min", moment(record.DueDate).format("YYYY-MM-DD"));
  }
}
function onSelectData(record, index) {
  axios
    .post(`/scm/purchase/order/get`, [record._id])
    .then(
      (r) => {
        data.records = r.data;
        data.listChecked = r.data.Lines.map(function (l) {
          l.VendorID = r.data.VendorID;
          l.ItemVarian = helper.ItemVarian(l.ItemID, l.SKU);
          return l;
        });
        const VendorID = r.data.VendorID;
        const WarehouseID = r.data.Location.WarehouseID;
        const site = r.data.Dimension.find((d) => {
          return d.Key == "Site";
        }).Value;

        data.searchItem.Site = site;
        data.searchItem.VendorID = VendorID;
        if (r.data.Status != "DRAFT") {
          data.formMode = "view";
          data.tabs = ["General", "Line", "Attachment"];
          data.step = "form-po";
          const total = r.data.Lines.reduce((a, b) => {
            let val = 0;
            if (b.ReceivedQty) {
              val = b.ReceivedQty;
            }
            return a + val;
          }, 0);
          if (total == 0 && r.data.Status == "POSTED") {
            data.isClose = true;
          } else {
            data.isClose = false;
          }
        } else {
          data.formMode = "edit";
          data.isClose = false;
          onGeneratePRLine(site, "", "", "", "");
        }
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      util.nextTickN(2, () => {});
    });
}
function onRefreshPRLine(){
  data.paging = {
    skip: 0,
    pageSize: 20,
    currentPage: 1,
  }
  data.pagingItems= {
    skip: 0,
    pageSize: 20,
    currentPage: 1,
  }
  onGeneratePRLine(
    data.searchItem.Site,
    data.searchItem.DeliveryTo,
    data.searchItem.VendorID,
    data.searchItem.ReffNo,
    data.searchItem.Requestor
  );
}
function onGeneratePRLine(
  Site = "",
  DeliveryTo = "",
  VendorID = "",
  ReffNo = "",
  Requestor = ""
) {
  data.step = "list-pr";
  let lines = data.records.Lines;
  let SortPriority = []
  for(let s=0; s<lines.length; s++){
    SortPriority.push(`${lines[s].PRID}|${lines[s].LineNo}`)
  }
  let payload = {
    VendorID: VendorID,
    DeliveryTo: DeliveryTo,
    Site: Site,
    ReffNo: ReffNo,
    Requestor: Requestor,
    SortPriority:SortPriority,
    Skip: (data.pagingItems.currentPage - 1) * data.pagingItems.pageSize,
    Take: data.pagingItems.pageSize,
  };
  if (
    data.searchItem.VendorID ||
    data.searchItem.DeliveryTo ||
    data.searchItem.Site ||
    data.searchItem.Requestor
  ) {
    payload = { ...payload, ...data.searchItem };
  }
  if (!payload.Site) {
    return util.showError("Please select Site");
  }

  data.loading.processItem = true;
  util.nextTickN(2, () => {
    listPRControl.value.setLoading(true);
    axios
      .post("/scm/purchase/request/get-lines-v1", payload)
      .then((r) => {
        r.data.data.map((r) => {
          // isSelected
          const Line = lines.find(function (l) {
            return (
              `${l.PRID}${l.ItemID}${l.SKU}${l.SourceLineNo}${JSON.stringify(
                l.Dimension.find((d) => {
                  return d.Key == "Site";
                })
              )}` ==
              `${r.PRID}${r.ItemID}${r.SKU}${r.LineNo}${JSON.stringify(
                r.Dimension.find((d) => {
                  return d.Key == "Site";
                })
              )}`
            );
          });

          if (Line) {
            r.isSelected = true;
            r.Qty = Line.Qty;
          }else{
            let arrItems = helper.cloneObject(data.listChecked).map((l)=>{
              return `${l.PRID}${l.ItemID}${l.SKU}${l.SourceLineNo}${JSON.stringify(
                l.Dimension.find((d) => {
                  return d.Key == "Site";
                })
              )}`
            });
            r.isSelected = arrItems.includes(`${r.PRID}${r.ItemID}${r.SKU}${r.LineNo}${JSON.stringify(
                r.Dimension.find((d) => {
                  return d.Key == "Site";
                })
              )}`);
          }
          return r;
        });
        data.itemsPRline = r.data.data;
        data.countPRline = r.data.count;
        listPRControl.value.setRecords(r.data.data);
        listPRControl.value.setLoading(false);
        data.loading.processItem = false;
      })
      .finally(() => {
        data.loading.processItem = false;
      });
  });
}
function onCheckUncheckAll(checked) {
  let items = data.listChecked;
  for (let v = 0; v < data.itemsPRline.length; v++) {
    const exists = items.find(function (i) {
      return (
        `${i.ItemID}${i.SKU}${i.PRID}${i.LineNo}` ==
        `${data.itemsPRline[v].ItemID}${data.itemsPRline[v].SKU}${data.itemsPRline[v].PRID}${data.itemsPRline[v].LineNo}`
      );
    });
    if (checked && !exists) {
      items.push(data.itemsPRline[v]);
    } else {
      const newItem = items.filter(function (i) {
        return (
          `${i.ItemID}${i.SKU}${i.PRID}${i.LineNo}` ==
          `${val.ItemID}${i.SKU}${i.PRID}${i.LineNo}`
        );
      });
      items = newItem;
    }
  }
  data.listChecked = items;
}
function onCheckUncheck(val) {
  let items = data.listChecked;
  const exists = items.find(function (i) {
    return (
      `${i.ItemID}${i.SKU}${i.PRID}${i.LineNo}` ==
      `${val.ItemID}${val.SKU}${val.PRID}${val.LineNo}`
    );
  });

  if (val.isSelected && !exists) {
    items.push(val);
  } else {
    const newItem = items.filter(function (i) {
      return (
        `${i.ItemID}${i.SKU}${i.PRID}${i.LineNo}` !=
        `${val.ItemID}${val.SKU}${val.PRID}${val.LineNo}`
      );
    });
    items = newItem;
  }
  data.listChecked = items;
}
function lookupPayloadBuilderSite(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

  //setting search
  const querySite = [
    {
      Field: "Dimension.Key",
      Op: "$eq",
      Value: "Site",
    },
    {
      Field: "Dimension.Value",
      Op: "$eq",
      Value: data.searchItem.Site,
    },
  ];
  if (data.searchItem.Site) {
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
    if (data.searchItem.Site) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}
function lookupPayloadBuilderDimension(search, select, value, item) {
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
      onFieldChange("VendorID", "", "", "");
      break;
    default:
      break;
  }
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
  const line = lineConfig.value.getDataValue();
  if (line.length > 0) {
    for (let l = 0; l < line.length; l++) {
      line[l].InventDim = val;
    }
  }
}

function getPostingProfile(record) {
  util.nextTickN(2, () => {
    formPOControl.value.setLoading(true);
    axios
      .post(`/scm/purchase/order/journal/type/find?TrxType=Purchase Order`)
      .then(
        (r) => {
          if (r.data.length > 0) {
            record.JournalTypeID = r.data[0]._id;
            record.PostingProfileID = r.data[0].PostingProfileID;
            data.records = record;
            data.keyJournalType = util.uuid();
          }
        },
        (e) => util.showError(e)
      )
      .finally(function () {
        formPOControl.value.setLoading(false);
      });
  });
}

function preSaveAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [`PO_${data.records._id}`];
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
    return util.showError("General Name is required");
  }

  if (!record.Requestor) {
    return util.showError("General Requestor is required");
  }

  if (!record.Location?.WarehouseID) {
    return util.showError("General Warehouse ID is required");
  }

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
    if (
      record.Lines[l].PRID &&
      record.Lines[l].Qty > record.Lines[l].RemainingQty
    ) {
      return util.showError("qty more than Remaining qty");
    }

    if (typeof record.Lines[l].DiscountValue != "number") {
      return util.showError("there is a line field Discount Value required");
    }
  }

  if (record?.Lines?.length > 0) {
    record.Lines.map(function (v) {
      // v.SourceLineNo = v.LineNo;
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
      data.loading.processItem = false;
      formPOControl.value.setLoading(false);
      return util.showError("field Discount Type cannot be empty");
    }
    if (typeof payload.Discount?.DiscountAmount != "number") {
      data.loading.processItem = false;
      formPOControl.value.setLoading(false);
      return util.showError("field Discount Type cannot be empty");
    }
  }

  if (validDimension && validate && formPOControl.value.validate()) {
    formPOControl.value.setLoading(true);
    let url = "/scm/purchase/order/save";
    payload.DocumentDate = helper.dateTimeNow(payload.DocumentDate);
    payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
    payload.DueDate = helper.dateTimeNow(payload.DueDate);
    axios
      .post(url, payload)
      .then(
        (r) => {
          data.records = {
            ...payload,
            _id: r.data._id,
            Status: r.data.Status,
          };
          data.btnPostId = util.uuid();
          util.nextTickN(2, () => {
            if (gridAttachment.value) {
              gridAttachment.value.Save();
            }
            postSaveAttachment();
          });
          openForm(data.records);
          return util.showInfo("Purchase Order has been successful save");
        },
        (e) => {
          util.showError(e);
        }
      )
      .finally(function () {
        data.loading.processItem = false;
        formPOControl.value.setLoading(false);
      });
  } else {
    return util.showError("Please check general required field");
  }
}
function preSubmit(status, action, doSubmit) {
  setRequiredAllField(true);
  const dim = data.records.Location;
  const pc = data.records.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = data.records.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = data.records.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;

  if (!data.records.Requestor) {
    return util.showError("General Requestor is required");
  }

  if (!data.records.BillingName) {
    return util.showError("General Billing Name is required");
  }

  if (!data.records.BillingAddress) {
    return util.showError("General Billing Address is required");
  }
  let valid = formPOControl.value.validate();
  let validate = true;
  let validDimension = true;

  if (status == "DRAFT") {
    if (InventDimControl.value) {
      validate = InventDimControl.value.validate();
    }

    if (!data.records.Name) {
      return util.showError("General Name is required");
    }

    if (!data.records.Location?.WarehouseID) {
      return util.showError("General Warehouse ID is required");
    }

    if (FinancialDimension.value) {
      validDimension = FinancialDimension.value.validate();
    } else {
      if (!pc || !cc || !site) {
        validDimension = false;
      }
    }

    for (let l = 0; l < data.records.Lines.length; l++) {
      if (!data.records.Lines[l].ItemID) {
        return util.showError("there is a line field Item required");
      }
      if (typeof data.records.Lines[l].Qty != "number") {
        return util.showError("there is a line field Qty required");
      }
      if (typeof data.records.Lines[l].UnitCost != "number") {
        return util.showError("there is a line field UnitCost required");
      }
      if (data.records.Lines[l].UnitCost == 0 && data.siteUser != headOffice) {
        return util.showError("UnitCost item which is 0");
      }
      if (
        data.records.Lines[l].PRID &&
        data.records.Lines[l].Qty > data.records.Lines[l].RemainingQty
      ) {
        return util.showError("qty more than Remaining qty");
      }

      if (typeof data.records.Lines[l].DiscountValue != "number") {
        return util.showError("there is a line field Discount Value required");
      }
    }

    if (data.records?.Lines?.length > 0) {
      data.records.Lines.map(function (v) {
        // v.SourceLineNo = v.LineNo;
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
        return v;
      });
    }

    [
      "PostingProfileID",
      "VendorID",
      "BillingName",
      "BillingAddress",
      "PIC",
      "Priority",
    ].every((e) => {
      if (!data.records[e]) {
        valid = false;
        return false;
      }
      return true;
    });

    let payload = JSON.parse(JSON.stringify(data.records));

    payload = {
      ...payload,
      ...JSON.parse(JSON.stringify(lineConfig.value.getOtherTotal())),
    };
    if (typeof payload.Discount?.DiscountValue != "number") {
      return util.showError("field Discount Type cannot be empty");
    }
    if (typeof payload.Discount?.DiscountAmount != "number") {
      return util.showError("field Discount Type cannot be empty");
    }
    if (valid && validate && validDimension) {
      formPOControl.value.setLoading(true);
      payload.DocumentDate = helper.dateTimeNow(payload.DocumentDate);
      payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
      payload.DueDate = helper.dateTimeNow(payload.DueDate);
      let url = "/scm/purchase/order/save";
      axios
        .post(url, payload)
        .then(
          (r) => {
            data.records = {
              ...payload,
              _id: r.data._id,
              Status: r.data.Status,
            };
            util.nextTickN(2, () => {
              if (gridAttachment.value) {
                gridAttachment.value.Save();
              }
              postSaveAttachment();
            });
            openForm(data.records);
            util.nextTickN(2, () => {
              doSubmit();
            });
          },
          (e) => {
            util.showError(e);
            data.loading.processItem = false;
            formPOControl.value.setLoading(false);
          }
        )
        .finally(function () {});
    } else {
      return util.showError("Please check general required field");
    }
  } else {
    doSubmit();
  }
}

function postSubmit(record, action) {
  axios.post("/scm/purchase/order/sync-journal-status", { PurchaseOrderID: record._id })
  .then((r)=>{
    data.btnPostId = util.uuid();
    data.loading.processItem = false;
    formPOControl.value.setLoading(false);
    data.step = "list-po";
    return util.showInfo("Purchase Order has been successful Submit");
  },(e)=>{
    return util.showError(e);
  })
}
function errorSubmit(e, action) {
  if (action === "Submit" || action === "Reject") {
    // calculate: qty exceeded, remaining qty: 0 PCS | total qty input (converted): 10 PCS
    // util.showError(e);
    rollbackItem(e, action);
  } else {
    data.btnPostId = util.uuid();
    data.loading.processItem = false;
    formPOControl.value.setLoading(false);
    // return util.showError(e);
  }
}

function rollbackItem(e, action) {
  let payload = JSON.parse(JSON.stringify(data.records));
  let url = "/scm/purchase/order/rollback-remaining-qty";
  axios
    .post(url, { PurchaseOrderJournalID: payload._id, Error: e })
    .then(
      (r) => {
        if (action == "Reject") {
          return util.showInfo("Purchase Order has been successful Reject");
        }
      },
      (e) => {
        return util.showError(e);
      }
    )
    .finally(function () {
      data.btnPostId = util.uuid();
      data.loading.processItem = false;
      formPOControl.value.setLoading(false);
    });
}

function preReopen() {
  let payload = JSON.parse(JSON.stringify(data.records));
  payload = {
    ...payload,
    ...JSON.parse(JSON.stringify(lineConfig.value.getOtherTotal())),
  };
  payload.Status = "DRAFT";
  axios.post("/scm/purchase/order/save", payload).then(
    (r) => {
      util.nextTickN(2, () => {
        data.loading.processItem = false;
        formPOControl.value.setLoading(false);
        data.step = "list-po";
      });
    },
    (e) => {
      return util.showError(e);
    }
  );
}
function setRequiredAllField(required) {
  formPOControl.value.getAllField().forEach((e) => {
    if (
      [
        "PostingProfileID",
        "BillingName",
        "BillingAddress",
        "PIC",
        "Priority",
      ].includes(e.field)
    ) {
      formPOControl.value.setFieldAttr(e.field, "required", required);
    }
  });
}

function onFieldChange(name, value1, value2, oldValue) {
  switch (name) {
    case "VendorID":
      formPOControl.value.setFieldAttr("TaxName", "readOnly", false);
      formPOControl.value.setFieldAttr("PaymentTerms", "readOnly", false);
      formPOControl.value.setFieldAttr("TaxType", "readOnly", false);
      formPOControl.value.setFieldAttr("TaxAddress", "readOnly", false);
      data.keyTaxCode = util.uuid();
      data.records.VendorName = "";
      data.records.PaymentTerms = "";
      data.records.TaxType = "";
      data.records.TaxName = "";
      data.records.TaxRegistration = "";
      data.records.TaxAddress = "";
      data.records.TaxCodes = [];
      data.records.DueDate = moment(new Date(data.records.TrxDate)).add(
        0,
        "days"
      );
      if (typeof value1 == "string") {
        axios.post("/bagong/vendor/get", [value1]).then(
          (r) => {
            data.records.VendorName = r.data.Name;
            data.records.PaymentTerms = r.data.PaymentTermID;
            data.records.TaxType = r.data.TaxType;
            data.records.TaxName = r.data.TaxName;
            data.records.TaxRegistration = r.data.TaxRegistrationNumber;
            data.records.TaxAddress = r.data.TaxAddress;
            let TaxCodes = [];
            if (r.data.Detail.Terms.Taxes1) {
              TaxCodes.push(r.data.Detail.Terms.Taxes1);
            }
            if (r.data.Detail.Terms.Taxes2) {
              TaxCodes.push(r.data.Detail.Terms.Taxes2);
            }
            data.records.TaxCodes = TaxCodes;
            formPOControl.value.setFieldAttr("TaxName", "readOnly", true);
            formPOControl.value.setFieldAttr("PaymentTerms", "readOnly", true);
            formPOControl.value.setFieldAttr("TaxType", "readOnly", true);
            formPOControl.value.setFieldAttr("TaxAddress", "readOnly", true);
            axios
              .post("/fico/paymentterm/get", [r.data.PaymentTermID])
              .then((p) => {
                data.records.DueDate = moment(
                  new Date(data.records.TrxDate)
                ).add(p.data.Days, "days");
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
    case "TrxDate":
      data.records.DueDate = moment(new Date(value1)).add(0, "days");
      axios
        .post("/fico/paymentterm/get", [data.records.PaymentTerms])
        .then((p) => {
          data.records.DueDate = moment(new Date(value1)).add(
            p.data.Days,
            "days"
          );
        });
      break;
    case "BillingName":
      if (value1 === "SITE020") {
        getWarehouse("WH-HO", data.records);
      } else {
        data.records.BillingAddress = data.records.DeliveryAddress;
      }

      break;
    default:
      break;
  }
  if (["PODate", "DueDate"].includes(name)) {
    setMinDate(record);
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
function onChangeFilter(v1, v2, item) {
  data.pagingItems.currentPage = 1;
  util.nextTickN(2, () => {
    onGeneratePRLine();
  });
}
function onBackListPO() {
  data.step = "list-po";
  data.searchItem = {
    VendorID: "",
    DeliveryTo: "",
    Site: "",
  };
}
function onProcess() {
  const listIR = data.listChecked.filter((ir) => {
    return ir.PRID != "";
  });
  let group = listIR.reduce((result, currentObject) => {
    const site = currentObject.Dimension.find((d) => {
      return d.Key == "Site";
    }).Value;
    const key = `${site}|${currentObject.InventDim.WarehouseID}|${currentObject.VendorID}`;
    result[key] = result[key] || [];
    result[key].push(currentObject);
    return result;
  }, {});

  if (Object.keys(group).length > 1) {
    return util.showError("only in 1 same site & warehouse & vendor");
  }

  if ([""].includes(data.records.Status) && listIR.length > 0) {
    data.records = listIR[0].Header;
    data.records._id = "";
    data.records.Name = "";
    data.records.Requestor = "";
    data.records.Status = "";
    data.records.JournalTypeID = "";
    data.records.PostingProfileID = "";
    data.records.DocumentDate = new Date();
    data.records.TrxDate = new Date();
    data.records.Created = new Date();
    data.records.LastUpdate = new Date();
    data.records.Location = {
      WarehouseID: "",
    };
    data.records.Dimension = [
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
        Value: "",
      },
      {
        Key: "Asset",
        Value: "",
      },
    ];
  }

  if ([""].includes(data.records.Status)) {
    data.records.Discount = {
      DiscountType: "percent",
      DiscountValue: 0,
      DiscountAmount: 0,
    };
    axios
      .post("/fico/paymentterm/get", [data.records.PaymentTerms])
      .then((p) => {
        data.records.DueDate = moment(new Date(data.records.TrxDate)).add(
          p.data.Days,
          "days"
        );
      });
  }
  if (listIR.length > 0) {
    data.records.Dimension.map((d) => {
      if (d.Key == "Site") {
        d.Value = Object.keys(group).at(0).split("|").at(0);
      }
      return d;
    });
    data.records.Location.WarehouseID = Object.keys(group)
      .at(0)
      .split("|")
      .at(1);
    data.records.VendorID = Object.keys(group).at(0).split("|").at(2);
  }

  data.listChecked.map(function (l) {
    delete l.Header;
    l.ItemVarian = helper.ItemVarian(l.ItemID, l.SKU);
    return l;
  });

  data.records.Lines = data.listChecked;
  data.records.ReffNo = [
    ...new Set(
      data.listChecked.map((v) => {
        return v.PRID;
      })
    ),
  ];

  data.step = "form-po";
  data.btnPostId = util.uuid();
  openForm(data.records);
}
function getRequestor(record) {
  axios.post("/bagong/employee/get", [record.Requestor]).then(
    (r) => {
      data.requestorName = r.data.Name;
    },
    (e) => util.showError(e)
  );
}
function onCancel(item) {
  let payload = {
    ID: item._id,
  };
  data.loading.processItem = true;
  axios.post("/scm/purchase/order/cancel", payload).then(
    (r) => {
      data.btnPostId = util.uuid();
      data.loading.processItem = false;
      formPOControl.value.setLoading(false);
      data.step = "list-po";
      return util.showInfo("Purchase Order has been successfully canceled");
    },
    (e) => {
      data.loading.processItem = false;
      util.showError(e);
    }
  );
}
function onPreview() {
  getRequestor(data.records);
  data.appMode = "preview";
}
function closePreview() {
  data.appMode = "form";
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
    listPOControl.value.refreshData();
  });
}
function getReffNo() {
  data.listReffNo = [];
  axios
    .post("/scm/postingprofile/find-journal-ref", {
      JournalType: "Purchase Order",
      JournalID: data.records._id,
    })
    .then(
      (r) => {
        data.listReffNo = r.data.Refferences;
        data.isDialogReffNo = true;
      },
      (e) => util.showError(e)
    );
}
function lookupPayloadSearch(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

  //setting search
  const Site = data.searchItem.Site;
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
  getByCurrentUser();
  createGridCfgRefNo();
  loadGridConfig(axios, `/scm/purchase/order/gridconfig`).then(
    (r) => {
      r.setting.idField = "LastUpdate";
      r.setting.sortable = [
        "LastUpdate",
        "Created",
        "TrxDate",
        "Status",
        "_id",
      ];
      data.gridCfgTrx = r;
    },
    (e) => util.showError(e)
  );
  loadGridConfig(axios, `/scm/purchase/line/gridconfig`).then((r) => {
    loadGridConfig(axios, `/scm/inventory/journal/line/gridconfig`).then(
      (rLine) => {
        const alterFields = [
          {
            field: "PRID",
            kind: "text",
            label: "Purchase Request No.",
            halign: "start",
            valign: "start",
            labelField: "",
            readType: "show",
            unit: "",
            input: {
              field: "PRID",
              label: "Purchase Request No.",
              hint: "",
              hide: false,
              placeHolder: "Line",
              kind: "text",
            },
          },
          {
            field: "WarehouseID",
            kind: "text",
            label: "Delivery to",
            halign: "start",
            valign: "start",
            labelField: "",
            readType: "show",
            input: {
              field: "WarehouseID",
              label: "Delivery to",
              hint: "",
              hide: false,
              placeHolder: "Delivery to",
              kind: "text",
            },
          },
          {
            field: "VendorName",
            kind: "text",
            label: "Vendor Name",
            halign: "start",
            valign: "start",
            labelField: "",
            readType: "show",
            input: {
              field: "VendorName",
              label: "Vendor Name",
              hint: "",
              hide: false,
              placeHolder: "Vendor Name",
              kind: "text",
            },
          },
          {
            field: "LineNo",
            kind: "text",
            label: "Source Line No",
            halign: "start",
            valign: "start",
            labelField: "",
            readType: "show",
            input: {
              field: "LineNo",
              label: "LineNo",
              hint: "",
              hide: false,
              placeHolder: "Source Line No",
              kind: "text",
            },
          },
        ];
        data.gridCfgPrline = rLine
        data.gridCfgPrline.fields = [
          ...rLine.fields.filter((ivLine) => {
            return !["Remarks"].includes(ivLine.field);
          }),
          ...alterFields,
          ...r.fields,
        ].filter(
          (o) =>
            ![
              "InventJournalLine",
              "SourceLineNo",
              "SKU",
              "DiscountType",
              "DiscountValue",
              "DiscountAmount",
            ].includes(o.field)
        );
        data.gridCfgLine = rLine;
      }
    );
  });
  loadFormConfig(axios, "/scm/purchase/order/formconfig").then(
    (r) => {
      data.frmCfg = r;
      if (route.query.trxid !== undefined) {
        let currQuery = { ...route.query };
        // listControl.value.selectData({ _id: currQuery.trxid }); //remark sementara tunggu suimjs update
        delete currQuery["trxid"];
        router.replace({ path: route.path, query: currQuery });
      }
    },
    (e) => util.showError(e)
  );

  if (route.query.trxid !== undefined || route.query.id !== undefined) {
    let getUrlParam = route.query.trxid || route.query.id;
    axios
      .post(`/scm/purchase/order/get`, [getUrlParam])
      .then(
        (r) => {
          data.records = r.data;
          data.listChecked = r.data.Lines.map(function (l) {
            l.VendorID = r.data.VendorID;
            l.ItemVarian = helper.ItemVarian(l.ItemID, l.SKU);
            return l;
          });
          const VendorID = r.data.VendorID;
          const site = r.data.Dimension.find((d) => {
            return d.Key == "Site";
          }).Value;

          data.searchItem.Site = site;
          data.searchItem.VendorID = VendorID;
          if (r.data.Status != "DRAFT") {
            data.formMode = "view";
            data.tabs = ["General", "Line", "Attachment"];
            data.step = "form-po";
          } else {
            data.formMode = "edit";
            onGeneratePRLine(site, "", "", "", "");
          }
        },
        (e) => util.showError(e)
      )
      .finally(() => {
        util.nextTickN(2, () => {
          router.replace({
            path: `/scm/PurchaseOrder`,
          });
        });
      });
  }
});
</script>
<style>
.adminfee,
.parkingfee,
.deliveryfee {
  padding-left: 6px;
}
</style>
