<template>
  <div class="w-full">
    <s-card
      v-show="data.isPreview == false"
      :title="`${cardTitle} ${
        !data.appMode ? '' : '| Items ' + data.listChecked.length
      }`"
      class="w-full bg-white suim_datalist card"
      hide-footer
      :no-gap="false"
      :hide-title="false"
    >
      <s-grid
        v-if="data.appMode == ''"
        v-model="data.value"
        ref="listControlTrx"
        class="w-full grid-line-items"
        hide-sort
        :hide-new-button="true"
        :hide-refresh-button="true"
        :hide-delete-button="true"
        :hide-detail="true"
        :hide-action="false"
        :hide-select="true"
        auto-commit-line
        no-confirm-delete
        :config="data.gridCfgTrx"
        form-keep-label
      >
        <template #header_search="{ config }">
          <s-input
            ref="refItemID"
            v-model="data.InventTrx.search"
            lookup-key="_id"
            label="Search"
            class="w-full"
            @keyup.enter="refreshRecords(true)"
          ></s-input>
          <s-input
            label="Trx Date"
            kind="date"
            class="w-[200px]"
            v-model="data.InventTrx.TrxDate"
          ></s-input>
          <s-input
            ref="refStatus"
            label="Status"
            v-model="data.InventTrx.Status"
            class="w-[250px]"
            use-list
            :items="[
              'All',
              'DRAFT',
              'SUBMITTED',
              'READY',
              'POSTED',
              'REJECTED',
            ]"
            @change="refreshRecords(true)"
          ></s-input>
          <s-input
            ref="refSite"
            v-model="data.InventTrx.Site"
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
                      data.InventTrx.Site,
                      false,
                      ...args
                    )
                : undefined
            "
            @change="refreshRecords(true)"
          ></s-input>
          <button
            @click="changeSortDirection"
            v-if="!hideSort"
            class="sort_btn"
          >
            <mdicon :name="sortIcon" size="18" />
          </button>
          <select
            v-model="data.sortField"
            class="sort_select border-b"
            @change="refreshRecords(true)"
          >
            <option value="">No Sort</option>
            <option v-for="f in data.sortable" :value="f">
              {{ f }}
            </option>
          </select>
        </template>
        <template #header_buttons="{ config }">
          <s-button
            :disabled="data.isProcess"
            icon="refresh"
            class="btn_primary refresh_btn"
            @click="refreshRecords(true)"
          />
          <s-button
            v-if="profile.canCreate"
            :disabled="data.isProcess"
            icon="plus"
            class="btn_primary new_btn"
            @click="newRecords"
          />
        </template>
        <template #item_Created="{ item }">
          <div class="text-right">
            {{ moment(item.Created).local().format("DD-MMM-YYYY HH:mm") }}
          </div>
        </template>
        <template #item_ReffNo="{ item }">
          {{ item.ReffNo.join(", ") }}
        </template>
        <!-- <template #item_Approvers="{ item }">
          <list-approvers :approvers="item.Approvers" />
        </template> -->
        <template #item_Status="{ item }">
          <status-text :txt="item.Status" />
        </template>
        <template #item_buttons_1="{ item }">
          <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
          <a
            v-if="profile.canUpdate"
            href="#"
            @click="onSelectTex(item, idx)"
            class="edit_action"
          >
            <mdicon
              name="pencil"
              width="16"
              alt="edit"
              class="cursor-pointer hover:text-primary"
            />
          </a>
        </template>
        <template #item_button_delete="{ item }">
          <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
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
      <s-grid
        v-if="data.appMode == 'grid'"
        v-model="data.valueItem"
        ref="listControl"
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
            v-model="data.search.SourceJournalID"
            kind="text"
            label="Search Source Journal"
            class="w-full"
            :disabled="data.loading.processItem"
          ></s-input>
          <s-input
            :key="WHSearch"
            label="Warehouse"
            v-model="data.search.WarehouseID"
            class="w-[500px]"
            use-list
            lookup-url="/tenant/warehouse/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="
              (field, v1, v2, old, ctlRef) => {
                onChangeWarehouse(v1, v2, item);
              }
            "
            :lookup-payload-builder="
              (search) =>
                lookupPayloadBuilder(
                  search,
                  ['_id', 'Name'],
                  data.search.WarehouseID,
                  item
                )
            "
          ></s-input>
          <s-input
            ref="refInput"
            label="Vendor"
            v-model="data.search.VendorID"
            class="w-[600px]"
            use-list
            :lookup-url="`/tenant/vendor/find`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="
              (field, v1, v2, old, ctlRef) => {
                onChangeSourceType(v1, v2, item);
              }
            "
          ></s-input>
          <s-input
            v-model="data.search.SourceType"
            useList
            label="Source Type"
            class="min-w-[200px]"
            :items="
              data.trxType == 'Inventory Issuance'
                ? ['INVENTORY', 'SALES ORDER', 'WORKORDER']
                : ['INVENTORY', 'PURCHASE', 'WORKORDER']
            "
            :disabled="data.loading.processItem"
            @change="
              (field, v1, v2, old, ctlRef) => {
                onChangeSourceType(v1, v2, item);
              }
            "
          ></s-input>
        </template>
        <template #header_buttons="{ config }">
          <s-button
            icon="refresh"
            class="btn_primary refresh_btn"
            :disabled="data.loading.processItem"
            @click="onFilterRefreshItems"
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
            @click="onBackGridTrx"
          />
        </template>
        <template #item_Item="{ item }">
          {{ item.ItemName }}
        </template>
        <template #item_SKU="{ item }">
          {{ item.SKUName }}
        </template>
        <template #item_OriginDestinationName="{ item }">
          {{
            item.SourceType === "PURCHASE"
              ? item.VendorName
              : item.OriginDestinationName
          }}
        </template>
        <template #item_Qty="{ item }">
          <div class="text-right">
            {{ helper.formatNumberWithDot(item.Qty) }}
          </div>
        </template>
        <template #item_TrxQty="{ item }">
          <div class="text-right">
            {{ helper.formatNumberWithDot(item.TrxQty) }}
          </div>
        </template>
        <template #item_SettledQty="{ item }">
          <div class="text-right">
            {{ helper.formatNumberWithDot(item.SettledQty) }}
          </div>
        </template>
        <template #item_OriginalQty="{ item }">
          <div class="text-right">
            {{ helper.formatNumberWithDot(item.OriginalQty) }}
          </div>
        </template>
        <template #paging>
          <s-pagination
            :recordCount="data.itemFulfillment.length"
            :pageCount="pageCountItem"
            :current-page="data.pagingItems.currentPage"
            :page-size="data.pagingItems.pageSize"
            @changePage="changePageItem"
            @changePageSize="changePageSizeItem"
          ></s-pagination>
        </template>
      </s-grid>
      <s-form
        id="form-inventory-trx"
        v-if="data.appMode == 'form'"
        ref="formCtlInventTrx"
        v-model="data.records"
        :keep-label="true"
        :config="data.frmCfg"
        class="pt-2"
        :auto-focus="true"
        :hide-submit="true"
        :hide-cancel="true"
        :tabs="
          data.records._id
            ? ['General', 'Line', 'Attachment']
            : ['General', 'Line']
        "
        :mode="data.formModeInventTrx"
        @cancelForm="onCancelForm"
        @fieldChange="onFieldChange"
      >
        <template #tab_Line="{ item }">
          <div>
            <s-grid
              v-show="data.appLine == ''"
              v-model="data.valueLine"
              ref="LineControl"
              class="w-full grid-line-items"
              hide-search
              hide-sort
              :hide-new-button="true"
              :hide-delete-button="true"
              hide-refresh-button
              :hide-detail="true"
              :hide-action="false"
              hide-select
              auto-commit-line
              no-confirm-delete
              :config="data.gridLineCfg"
              form-keep-label
            >
              <template #item_ItemName="{ item }">
                {{ item.ItemName }}
              </template>
              <template #item_SKU="{ item }">
                <s-input
                  v-if="item.SKU"
                  ref="refSKU"
                  v-model="item.SKU"
                  :disabled="true"
                  use-list
                  :lookup-url="`/tenant/itemspec/gets-info`"
                  lookup-key="_id"
                  :lookup-labels="['Description']"
                  :lookup-searchs="['_id', 'SKU', 'Description']"
                  class="w-full"
                ></s-input>
              </template>
              <template #item_TrxQty="{ item }">
                <div class="text-right">
                  {{ helper.formatNumberWithDot(item.TrxQty) }}
                </div>
              </template>
              <template #item_SettledQty="{ item }">
                <div class="text-right">
                  {{ helper.formatNumberWithDot(item.SettledQty) }}
                </div>
              </template>
              <template #item_OriginalQty="{ item }">
                <div class="text-right">
                  {{ helper.formatNumberWithDot(item.OriginalQty) }}
                </div>
              </template>
              <template #item_VendorID="{ item }">
                {{
                  item.SourceType === "PURCHASE"
                    ? item.VendorName
                    : item.OriginDestinationName
                }}
              </template>
              <template #item_RemainingQty="{ item }">
                <div class="text-right">
                  {{
                    helper.formatNumberWithDot(
                      item.OriginalQty - item.SettledQty
                    )
                  }}
                </div>
              </template>
              <template #item_Qty="{ item }">
                <s-input
                  v-if="data.formModeInventTrx != 'view'"
                  ref="refQty"
                  kind="number"
                  v-model="item.Qty"
                  label="Qty"
                  hide-label
                  class="w-full"
                ></s-input>
                <div v-else class="text-right">
                  {{ helper.formatNumberWithDot(item.Qty) }}
                </div>
              </template>
              <template #item_buttons_1="{ item }">
                <a
                  href="#"
                  @click="onSelectLineDIM(item, idx)"
                  class="edit_action"
                >
                  <mdicon
                    name="eye"
                    width="16"
                    alt="edit"
                    class="cursor-pointer hover:text-primary"
                  />
                </a>
                <a
                  href="#"
                  @click="onSelectLineBatchSN(item, idx)"
                  class="edit_action"
                >
                  <mdicon
                    name="format-list-numbered"
                    width="16"
                    alt="edit"
                    class="cursor-pointer hover:text-primary"
                  />
                </a>
              </template>
            </s-grid>
            <s-form
              v-show="data.appLine == 'InventDim'"
              ref="formCtlInventDim"
              v-model="data.valueInventDim"
              :keep-label="true"
              :config="data.fromCfgDim"
              class="pt-2"
              :auto-focus="true"
              :hide-submit="true"
              :hide-cancel="true"
              mode="view"
            >
              <template #input_Dimension="{ item }">
                <dimension-editor
                  :key="data.keyDimItem"
                  v-model="item.Dimension"
                  :readOnly="true"
                  :default-list="profile.Dimension"
                ></dimension-editor>
              </template>
              <template #input_InventDim="{ item }">
                <div>
                  <dimension-invent-jurnal
                    :key="data.keyDimItem"
                    v-model="item.InventDim"
                    title-header="Inventory Dimension"
                    :readOnly="
                      ['SUBMITTED', 'POSTED', 'READY'].includes(
                        data.records.Status
                      )
                    "
                    :hideField="[
                      'BatchID',
                      'SerialNumber',
                      'InventDimID',
                      'SpecID',
                      'VariantID',
                      'Size',
                      'Grade',
                    ]"
                    :disableField="data.disableField"
                    @alterFormConfig="onAlterFormConfig"
                  ></dimension-invent-jurnal>
                </div>
              </template>
              <template #input_General1="{ item }">
                <div class="section grow">
                  <div class="flex flex-col gap-4">
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"
                            ><div>Vendor ID</div></label
                          >
                          <div class="bg-transparent">{{ item.VendorID }}</div>
                        </div>
                      </div>
                    </div>
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"><div>Variant</div></label>
                          <div class="bg-transparent">{{ item.ItemName }}</div>
                        </div>
                      </div>
                    </div>
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"
                            ><div>Source type</div></label
                          >
                          <div class="bg-transparent">
                            {{ item.SourceType }}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"
                            ><div>Source JournalID</div></label
                          >
                          <div class="bg-transparent">
                            {{ item.SourceJournalID }}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"
                            ><div>Source Type</div></label
                          >
                          <div class="bg-transparent">
                            {{ item.SourceTrxType }}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
              <template #input_General2="{ item }">
                <div class="section grow">
                  <div class="flex flex-col gap-4">
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"><div>Unit</div></label>
                          <div class="bg-transparent">
                            {{ item.UnitID }}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"
                            ><div>Unit Cost</div></label
                          >
                          <div class="bg-transparent">
                            {{ item.UnitCost }}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"><div>TrxQty</div></label>
                          <div class="bg-transparent">
                            {{ item.TrxQty }}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"
                            ><div>SettledQty</div></label
                          >
                          <div class="bg-transparent">
                            {{ item.SettledQty }}
                          </div>
                        </div>
                      </div>
                    </div>
                    <div class="w-full items-start gap-2 grid gridCol1">
                      <div class="col-auto">
                        <div>
                          <label class="input_label"
                            ><div>OriginalQty</div></label
                          >
                          <div class="bg-transparent">
                            {{ item.OriginalQty }}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </s-form>
            <s-grid
              v-show="data.appLine == 'BatchSN'"
              v-model="data.valueBatchSN"
              ref="BatchSNControl"
              class="w-full grid-line-items"
              hide-search
              hide-sort
              :hide-new-button="
                ['SUBMITTED', 'POSTED', 'READY'].includes(data.records.Status)
                  ? true
                  : false
              "
              :hide-delete-button="false"
              hide-refresh-button
              :hide-detail="true"
              :hide-action="
                ['SUBMITTED', 'POSTED', 'READY'].includes(data.records.Status)
                  ? true
                  : false
              "
              hide-select
              auto-commit-line
              no-confirm-delete
              :config="data.gridCfgBatchSN"
              form-keep-label
              @newData="newRecordBatchSN"
              @deleteData="deleteRecordBatchSN"
            >
              <template #item_BatchID="{ item }">
                <s-input
                  ref="refBatchID"
                  v-model="item.BatchID"
                  class="w-full"
                  :disabled="data.formModeInventTrx == 'view' ? true : false"
                  use-list
                  :lookup-url="`/tenant/itembatch/find`"
                  lookup-key="_id"
                  :lookup-labels="['_id']"
                  :lookup-searchs="['_id', '_id']"
                ></s-input>
              </template>
              <template #item_SerialNumber="{ item }">
                <s-input
                  ref="refSerialNumberID"
                  v-model="item.SerialNumber"
                  :disabled="data.formModeInventTrx == 'view' ? true : false"
                  class="w-full"
                  use-list
                  :lookup-url="`/tenant/itemserial/find`"
                  lookup-key="_id"
                  :lookup-labels="['_id']"
                  :lookup-searchs="['_id', '_id']"
                ></s-input>
              </template>
              <template #item_Qty="{ item }">
                <s-input
                  ref="refQty"
                  kind="number"
                  v-model="item.Qty"
                  label="Qty"
                  hide-label
                  :disabled="data.formModeInventTrx == 'view' ? true : false"
                  class="w-full"
                ></s-input>
              </template>
            </s-grid>
          </div>
        </template>
        <template #tab_Attachment="{ item }">
          <s-grid-attachment
            v-if="item._id"
            ref="gridAttachment"
            :journal-id="item._id"
            :journal-type="
              data.trxType == 'Inventory Issuance'
                ? 'Inventory Issuance'
                : 'Inventory Receive'
            "
            :tags="linesTag"
            v-model="item.Attachment"
            @pre-Save="preSaveAttachment"
          />
        </template>
        <template #input__id="{ item }">
          <s-input
            v-model="item._id"
            label="ID"
            class="w-full"
            :disabled="item.Status != ''"
          ></s-input>
        </template>
        <template #input_TrxType="{ item }">
          <s-input
            v-model="item.TrxType"
            label="Trx Type"
            class="w-full"
            :disabled="true"
          ></s-input>
        </template>
        <template #input_ReffNo="{ item }">
          <s-input
            ref="refReffNo"
            label="Reff No"
            v-model="item.ReffNo"
            class="w-full"
            :required="true"
            :disabled="true"
            use-list
            :lookup-url="`/scm/purchase/order/find`"
            lookup-key="_id"
            :lookup-labels="['_id']"
            :lookup-searchs="['_id']"
            :multiple="true"
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
            :lookup-url="`/scm/inventory/journal/type/find?TransactionType=${data.trxType}`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
          ></s-input>
        </template>
        <template #input_PostingProfileID="{ item }">
          <s-input
            :key="util.uuid()"
            ref="refInput"
            label="Posting Profile"
            v-model="item.PostingProfileID"
            class="w-full"
            :required="true"
            :disabled="true"
            :keepErrorSection="true"
            use-list
            :lookup-url="`/fico/postingprofile/find`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
          ></s-input>
        </template>
        <template #input_CompanyID="{ item }">
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
        <template #input_WarehouseID="{ item }">
          <s-input
            ref="refWarehouseID"
            v-model="item.WarehouseID"
            :disabled="true"
            label="Warehouse ID"
            use-list
            :lookup-url="`/tenant/warehouse/find`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
          ></s-input>
        </template>
        <template #input_Dimension="{ item }">
          <div>
            <dimension-editor
              :key="data.keyDimGeneral"
              ref="FinancialDimension"
              v-model="item.Dimension"
              :default-list="profile.Dimension"
              :readOnly="data.formModeInventTrx == 'edit' ? false : true"
            ></dimension-editor>
          </div>
        </template>
        <template #buttons="{ item }">
          <span
            v-if="data.alertPosted"
            style="
              background-color: red;
              padding: 6px;
              color: white;
              font-weight: 700;
              border-radius: 6px;
            "
          >
            {{ `All Qty has been Received/Issued` }}
          </span>

          <s-button
            :icon="`rewind`"
            class="btn_warning back_btn"
            :label="'Back'"
            @click="onBackForm"
          />
        </template>
        <template #buttons_1="{ item }">
          <div class="flex gap-1" v-show="data.appLine == ''">
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
              v-if="!['SUBMITTED', 'POSTED', 'READY'].includes(item.Status)"
              :icon="`content-save`"
              class="btn_primary submit_btn"
              :label="'Save'"
              :disabled="data.loading.processItem"
              @click="onSave"
            />
            <form-buttons-trx
              v-if="data.valueLine.length > 0"
              :key="data.btnTrxId"
              :status="item.Status"
              :moduleid="`scm/new`"
              :autoPost="false"
              :journal-id="item._id"
              :posting-profile-id="item.PostingProfileID"
              :disabled="data.loading.processItem"
              :journal-type-id="
                data.trxType == 'Inventory Issuance'
                  ? 'Inventory Issuance'
                  : 'Inventory Receive'
              "
              @preSubmit="trxPreSubmit"
              @postSubmit="trxPostSubmit"
            />
          </div>
        </template>
      </s-form>
    </s-card>
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
    <PreviewReport
      v-if="data.isPreview"
      class="card w-full"
      title="Preview"
      :preview="data.preview"
      @close="closePreview"
      :SourceType="data.trxType"
      :SourceJournalID="data.records._id"
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
  vModelCheckbox,
} from "vue";
import { layoutStore } from "@/stores/layout.js";
import { useRoute, useRouter } from "vue-router";
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
  SModal,
  SPagination,
} from "suimjs";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import StatusText from "@/components/common/StatusText.vue";
import LogTrx from "@/components/common/LogTrx.vue";
// import ListApprovers from "@/components/common/ListApprovers.vue";
import moment from "moment";
import { authStore } from "@/stores/auth";
import PreviewReport from "@/components/common/PreviewReport.vue";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";
const router = useRouter();
const route = useRoute();
const featureID =
  route.query.type == "Inventory Issuance" ? "GoodIssuance" : "GoodReceive";
const profile = authStore().getRBAC(featureID);
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);
const auth = authStore();
const listControlTrx = ref(null);
const listControl = ref(null);
const formCtlInventTrx = ref(null);
const FinancialDimension = ref(null);
const LineControl = ref(null);
const BatchSNControl = ref(null);
const gridAttachment = ref(SGridAttachment);
const axios = inject("axios");
const pageCount = computed({
  get() {
    const count = Math.ceil(data.recordsCount / data.paging.pageSize);
    return count;
  },
});
const pageCountItem = computed({
  get() {
    return Math.ceil(data.countFulfillment / data.paging.pageSize);
  },
});
const sortIcon = computed({
  get() {
    switch (data.sortDirection) {
      case "asc":
        return "arrow-up";

      case "desc":
        return "arrow-down";

      default:
        return "radiobox-blank";
    }
  },
});
const cardTitle = computed(() => {
  let title =
    route.query.title == "Inventory Receive"
      ? "Goods Receive"
      : "Goods Issuance";
  if (data.appMode == "") {
    return `${title}`;
  } else {
    return data.records.Status == ""
      ? `${title}`
      : `${title} - ${data.records._id}`;
  }
});

const linesTag = computed({
  get() {
    const type = data.trxType == "Inventory Issuance" ? "GI" : "GR";
    let ReffNo = JSON.parse(
      JSON.stringify(data.records.ReffNo ? data.records.ReffNo : [])
    ).map((ref) => {
      return `${ref.slice(0, 2)}_${ref}`;
    });

    const tags =
      ReffNo && data.records._id
        ? [...[`${type}_${data.records._id}`], ...ReffNo]
        : data.records._id
        ? [`${type}_${data.records._id}`]
        : ReffNo;

    return tags;
  },
});

const addTags = computed({
  get() {
    const type = data.trxType == "Inventory Issuance" ? "GI" : "GR";
    return [`${type}_${data.records._id}`];
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

const data = reactive({
  isPreview: false,
  btnTrxId: util.uuid(),
  keyDimGeneral: util.uuid(),
  keyDimItem: util.uuid(),
  value: [],
  valueItem: [],
  valueLine: [],
  valueInventDim: {},
  valueBatchSN: [],
  valueItemSelected: [],
  recordsBatchSN: [],
  listReffNo: [],
  itemFulfillment: [],
  isDialogReffNo: false,
  recordsCount: 0,
  paging: {
    skip: 0,
    pageSize: 20,
    currentPage: 1,
  },
  pagingItems: {
    skip: 0,
    pageSize: 25,
    currentPage: 1,
  },
  batchSNID: "",
  titleForm: route.query.title || route.query.type,
  trxType: route.query.type,
  alertPosted: false,
  isProcess: false,
  appMode: "",
  appLine: "",
  sortDirection: "desc",
  sortable: ["_id", "Name", "Status", "TrxDate", "Created", "LastUpdate"],
  sortField: "LastUpdate",
  formMode: "edit",
  formInsert: "/scm/inventory/trx/receipt/save",
  formUpdate: "/scm/inventory/trx/receipt/save",
  journalID: "",
  formModeInventTrx: "edit",
  gridCfgTrx: {},
  gridCfg: {},
  frmCfg: {},
  gridLineCfg: {},
  fromCfgDim: {},
  gridCfgBatchSN: {},
  gridCfgReffNo: {},
  records: {},
  record: null,
  InventTrx: {
    by: "_id",
    search: "",
    Status: "All",

    Site: "",
    TrxDate: "",
  },
  WHSearch: util.uuid(),
  keyJournalType: util.uuid(),
  search: {
    CompanyID: auth.companyId,
    WarehouseID: "",
    SourceType: "",
    VendorID: "",
    SourceJournalID: "",
  },
  listChecked: [],
  queryParams: [],
  disableField: [],
  loading: {
    processItem: false,
  },
  preview: {},
  warehouseparams: undefined,
  defaultListWarehouse: [],
});

function changePage(page) {
  data.paging.currentPage = page;
  refreshRecords();
}

function changeSortDirection() {
  const sorts = ["asc", "desc"];
  const sortCount = sorts.length;

  let found = false;
  for (let i = 0; i < sortCount; i++) {
    if (found) {
      data.sortDirection = sorts[i];
      if (data.sortDirection != "") refreshRecords();
      return;
    }
    if (sorts[i] == data.sortDirection) found = true;
  }

  data.sortDirection = "asc";
  if (data.sortDirection != "") refreshRecords();
}

function refreshRecords(isPaging = false) {
  util.nextTickN(2, () => {
    const filters = [];
    if (data.InventTrx.Status != "All" && data.InventTrx.Status) {
      filters.push({
        Field: "Status",
        Op: "$eq",
        Value: data.InventTrx.Status,
      });
    }

    if (data.InventTrx.search != "") {
      filters.push({
        Op: "$or",
        Items: [
          {
            Field: "_id",
            Op: "$contains",
            Value: [data.InventTrx.search],
          },
          {
            Field: "Name",
            Op: "$contains",
            Value: [data.InventTrx.search],
          },
          {
            Field: "ReffNo",
            Op: "$contains",
            Value: [data.InventTrx.search],
          },
        ],
      });
    }

    if (
      data.InventTrx.TrxDate != "" &&
      data.InventTrx.TrxDate != "Invalid date"
    ) {
      filters.push(
        {
          Field: "TrxDate",
          Op: "$gte",
          Value: moment(data.InventTrx.TrxDate).format("YYYY-MM-DDT00:00:00Z"),
        },
        {
          Field: "TrxDate",
          Op: "$lte",
          Value: moment(data.InventTrx.TrxDate).format("YYYY-MM-DDT23:59:00Z"),
        }
      );
    }

    if (data.InventTrx.Site != "") {
      filters.push(
        {
          Field: "Dimension.Key",
          Op: "$eq",
          Value: "Site",
        },
        {
          Field: "Dimension.Value",
          Op: "$eq",
          Value: data.InventTrx.Site,
        }
      );
    }
    let Where = {};
    if (filters.length > 0) {
      Where = {
        Op: "$and",
        Items: filters,
      };
    }
    let param = {
      Skip: (data.paging.currentPage - 1) * data.paging.pageSize,
      Take: data.paging.pageSize,
      Where: Where,
    };
    if (isPaging) {
      param.Skip = 0;
    }

    if (Object.keys(Where).length == 0) {
      delete param.Where;
    }

    if (data.sortField) {
      param.Sort = [
        `${data.sortDirection == "asc" ? "" : "-"}${data.sortField}`,
      ];
    }
    GridTrxRefreshed(param);
  });
}

function onBackGridTrx() {
  data.appMode = "";
  util.nextTickN(2, () => {
    refreshRecords();
  });
}

function onCheckUncheckAll(checked) {
  let items = data.listChecked;
  for (let v = 0; v < data.valueItem.length; v++) {
    const exists = items.find(function (i) {
      return (
        i.Item._id == data.valueItem[v].Item._id &&
        i.SKU == data.valueItem[v].SKU &&
        i.SourceJournalID == data.valueItem[v].SourceJournalID &&
        i.SourceLineNo == data.valueItem[v].SourceLineNo
      );
    });
    if (checked && !exists) {
      items.push(data.valueItem[v]);
    } else {
      const newItem = items.filter(function (i) {
        return (
          i.Item._id == data.valueItem[v].Item._id &&
          i.SKU == data.valueItem[v].SKU &&
          i.SourceJournalID == data.valueItem[v].SourceJournalID &&
          i.SourceLineNo == data.valueItem[v].SourceLineNo
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
      `${i.Item._id}${i.SKU}${i.SourceJournalID}${i.SourceLineNo}` ==
      `${val.Item._id}${val.SKU}${val.SourceJournalID}${val.SourceLineNo}`
    );
  });

  if (val.isSelected && !exists) {
    items.push(val);
  } else {
    const newItem = items.filter(function (i) {
      return (
        `${i.Item._id}${i.SKU}${i.SourceJournalID}${i.SourceLineNo}` !=
        `${val.Item._id}${val.SKU}${val.SourceJournalID}${val.SourceLineNo}`
      );
    });
    items = newItem;
  }
  data.listChecked = items;
}

function onChangeWarehouse(v1, v2, item) {
  util.nextTickN(2, () => {
    GridRefreshed();
  });
}

function preSaveAttachment(payload) {
  const type = data.trxType == "Inventory Issuance" ? "GI" : "GR";

  payload.map((asset) => {
    asset.Asset.Tags = [`${type}_${data.records._id}`];
    return asset;
  });
}

function onChangeSourceType(v1, v2, item) {
  util.nextTickN(2, () => {
    GridRefreshed();
  });
}

function newRecords(record) {
  data.listChecked = [];
  data.isProcess = true;
  data.search = {
    CompanyID: auth.companyId,
    WarehouseID: "",
    SourceType: "",
    VendorID: "",
    SourceJournalID: "",
  };
  data.records = {
    _id: "",
    Name: "",
    Status: "",
    TrxType: data.trxType,
    CompanyID: auth.companyId,
    WarehouseID: "",
    TrxDate: new Date(),
    Created: new Date(),
    LastUpdate: new Date(),
    Attachment: [],
  };
  data.appMode = "grid";
  setDefaultDimension(profile.Dimension, "new");
}

function onCancelForm() {
  const isMode = ["SUBMITTED", "POSTED", "READY"].includes(data.records.Status);
  if (isMode) {
    onBackGridTrx();
  } else {
    data.appMode = "grid";
    util.nextTickN(2, () => {
      GridRefreshed(true);
    });
  }
}

function onBackForm(record) {
  const tab = document.querySelector(
    ".tab_container > div.tab_selected"
  ).textContent;
  if (data.appLine == "" || tab == "General") {
    if (data.formModeInventTrx == "view") {
      onBackGridTrx();
    } else {
      data.appMode = "grid";
      util.nextTickN(2, () => {
        GridRefreshed(true);
      });
    }
  } else if (data.appLine == "InventDim") {
    util.nextTickN(2, () => {
      data.appLine = "";
    });
  } else if (data.appLine == "BatchSN") {
    util.nextTickN(2, () => {
      data.appLine = "";
    });
  }
}

function setFormRequired(isRequired) {
  formCtlInventTrx.value.setFieldAttr("CompanyID", "required", isRequired);
  formCtlInventTrx.value.setFieldAttr("Name", "required", isRequired);
  formCtlInventTrx.value.setFieldAttr("WarehouseID", "required", isRequired);
  formCtlInventTrx.value.setFieldAttr("TrxDate", "required", isRequired);
  formCtlInventTrx.value.setFieldAttr(
    "PostingProfileID",
    "required",
    isRequired
  );
}

function getPostingProfile(record) {
  const trxType =
    data.trxType == "Inventory Issuance"
      ? "Inventory Issuance"
      : "Inventory Receive";
  util.nextTickN(2, () => {
    axios
      .post(`/scm/inventory/journal/type/find?TransactionType=${trxType}`)
      .then(
        (r) => {
          if (r.data.length > 0) {
            data.keyJournalType = util.uuid();
            record.JournalTypeID = r.data[0]._id;
            record.PostingProfileID = r.data[0].PostingProfileID;
          }
          if (record._id == "") {
            record.Dimension = data.valueLine[0].Dimension;
          }
          
          data.records = {...data.records, ...record};
          data.appMode = "form";
          data.appLine = "";

          util.nextTickN(2, () => {
            formCtlInventTrx.value.setFieldAttr(
              "_id",
              "hide",
              data.records._id == "" ? true : false
            );
            const el = document.querySelector(
              "#form-inventory-trx .form_inputs > div.flex.section_group_container > div:nth-child(1) > div > div > div:nth-child(1)"
            );
            if (el){
              if (data.records._id == "") {
                el.style.display = "none";
              } else {
                el.style.display = "block";
              }
            }
            if (route.query.type == "Inventory Receive") {
              formCtlInventTrx.value.setFieldAttr("ReffNo", "hide", false);
            } else {
              formCtlInventTrx.value.setFieldAttr("ReffNo", "hide", false);
            }
          });
        },
        (e) => util.showError(e)
      );
  });
}
function onSave() {
  setFormRequired(true);
  let valid = true;
  const pc = data.records.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = data.records.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = data.records.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;
  if (formCtlInventTrx.value) {
    valid = formCtlInventTrx.value.validate();
  }

  if (!data.records.Name) {
    return util.showError("filed Name is required");
  } else if (!data.records.TrxDate) {
    return util.showError("filed TrxDate is required");
  }
  if (FinancialDimension.value) {
    valid = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      valid = false;
    }
  }
  const line = data.valueLine;
  data.loading.processItem = true;
  formCtlInventTrx.value.setLoading(true);
  for (let l = 0; l < line.length; l++) {
    if (line[l].Qty == 0) {
      valid = false;
      util.showError("there is a line field qty item which is 0");
      break;
    }
  }
  if (valid) {
    data.records.Lines = line;
    let url = "/scm/inventory/receive/save";
    const payload = JSON.parse(JSON.stringify(data.records));
    payload.Status = "DRAFT";
    axios
      .post(url, payload)
      .then(
        (r) => {
          data.records = {
            ...payload,
            _id: r.data._id,
            Status: r.data.Status,
          };
          data.btnTrxId = util.uuid();
          util.nextTickN(2, () => {
            if (gridAttachment.value) {
              gridAttachment.value.Save();
            }
            postSaveAttachment();
          });
          if (route.query.type == "Inventory Receive") {
            return util.showInfo("Goods Receive has been successful save");
          } else {
            return util.showInfo("Goods Issuance  has been successful save");
          }
        },
        (e) => {
          util.showError(e);
        }
      )
      .finally(function () {
        data.loading.processItem = false;
        formCtlInventTrx.value.setLoading(false);
      });
  } else {
    data.loading.processItem = false;
    formCtlInventTrx.value.setLoading(false);
    return util.showError("field is required");
  }
}

function trxPostSubmit(record) {
  onBackGridTrx();
}

function trxPreSubmit(status, action, doSubmit) {
  setFormRequired(true);
  let valid = true;
  const pc = data.records.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = data.records.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = data.records.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;
  if (formCtlInventTrx.value) {
    valid = formCtlInventTrx.value.validate();
  }
  const line = data.valueLine;
  data.records.Lines = line;

  data.loading.processItem = true;
  formCtlInventTrx.value.setLoading(true);
  const payload = JSON.parse(JSON.stringify(data.records));
  payload.Lines.map(function (v) {
    if (route.query.type == "Inventory Issuance") {
      v.SettledQty = v.SettledQty - v.Qty;
    } else {
      v.SettledQty = v.SettledQty + v.Qty;
    }
    return v;
  });
  if (!data.records.JournalTypeID) {
    valid = false;
    data.loading.processItem = false;
    return util.showError("filed Journal Type is required");
  } else if (!data.records.PostingProfileID) {
    valid = false;
    data.loading.processItem = false;
    return util.showError("filed Posting Profile is required");
  }
  if (FinancialDimension.value) {
    valid = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      valid = false;
    }
  }
  if (valid) {
    if (payload.Status == "DRAFT") {
      axios
        .post("/scm/inventory/receive/save", payload)
        .then(
          (r) => {
            data.records = r.data;
            data.journalID = r.data._id;
            util.nextTickN(2, () => {
              if (gridAttachment.value) {
                gridAttachment.value.Save();
              }
              postSaveAttachment();
            });
            util.nextTickN(2, () => {
              doSubmit();
            });
          },
          (e) => {}
        )
        .finally(function () {
          data.loading.processItem = false;
          formCtlInventTrx.value.setLoading(false);
        });
    } else {
      util.nextTickN(2, () => {
        doSubmit();
      });
    }
  } else {
    data.loading.processItem = false;
    formCtlInventTrx.value.setLoading(false);
    return util.showError("field is required");
  }
}

function onSelectTex(record) {
  const isMode = ["SUBMITTED", "POSTED", "READY"].includes(record.Status);
  setDefaultDimension(profile.Dimension, "edit");
  data.records = record;
  data.search.WarehouseID =
    record.Lines.length > 0
      ? record.Lines[0].InventDim.WarehouseID
      : "W-MLG001";
  data.appMode = isMode ? "form" : "grid";
  data.listChecked = record.Lines;
  data.journalID = record._id;
  util.nextTickN(2, () => {
    if (isMode == "form") {
      if (record._id) {
        formCtlInventTrx.value.setFieldAttr("_id", "hide", true);
      } else {
        formCtlInventTrx.value.setFieldAttr("_id", "hide", false);
      }
      if (route.query.type == "Inventory Receive") {
        formCtlInventTrx.value.setFieldAttr("ReffNo", "hide", false);
      } else {
        formCtlInventTrx.value.setFieldAttr("ReffNo", "hide", true);
      }
    }
    data.btnTrxId = util.uuid();
    // GridRefreshed();
    if (["", "DRAFT"].includes(record.Status)) {
      GridRefreshed(true);
    } else {
      data.formModeInventTrx = "view";
      data.valueLine = record.Lines;
      data.valueLine.map(function (vl) {
        vl.TrxQty = vl.OriginalQty - vl.SettledQty;
        return vl;
      });
      data.keyDimGeneral = util.uuid();
      // data.records.ReffNo = data.valueLine.map(function (sj) {
      //   return sj.SourceJournalID;
      // });
    }
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
function onProcess(record, index) {
  if (data.search.WarehouseID == null || data.search.WarehouseID == "") {
    return util.showError("Please select warehouse");
  }
  data.records.WarehouseID = data.search.WarehouseID;
  // =========== remove jika suimjs udh update ===============
  // const Items = data.valueItem.filter((el) => el.isSelected === true);
  // for (let i = 0; i < Items.length; i++) {
  //   const Line = data.listChecked.find(function (r, idx) {
  //     return (
  //       r.Item._id == Items[i].Item._id &&
  //       r.SKU == Items[i].SKU &&
  //       r.SourceJournalID == Items[i].SourceJournalID &&
  //       r.SourceLineNo == Items[i].SourceLineNo
  //     );
  //   });
  //   if (!Line) {
  //     data.listChecked.push(Items[i]);
  //   }
  // }
  // =========== remove jika suimjs udh update ===============

  let group = data.listChecked.reduce((result, currentObject) => {
    const key = `${currentObject.InventDim.WarehouseID}`;
    // Create an array for the key if it doesn't exist
    result[key] = result[key] || [];
    // Push the current object to the array
    result[key].push(currentObject);
    return result;
  }, {});
  if (Object.keys(group).length > 1) {
    return util.showError("only in 1 same warehouse");
  }

  let vendor = data.listChecked.reduce((result, currentObject) => {
    const key = `${currentObject.VendorID}`;
    // Create an array for the key if it doesn't exist
    result[key] = result[key] || [];
    // Push the current object to the array
    result[key].push(currentObject);
    return result;
  }, {});

  if (Object.keys(vendor).length > 1) {
    return util.showError("only in 1 same vendor");
  }

  let Selected = [];
  const lines = JSON.parse(JSON.stringify(data.listChecked)).map((c) => {
    return {
      ItemID: c.ItemID,
      SKU: c.SKU,
      SourceJournalID: c.SourceJournalID,
      SourceLineNo: c.SourceLineNo,
    };
  });
  if (lines.length == 0) {
    return util.showError("Please choose inventory transaction");
  }

  let payload = {
    SelectedLines: lines,
    Statuses:
      route.query.type == "Inventory Receive" ? ["Planned"] : ["Reserved"],
  };
  data.loading.processItem = true;
  axios
    .post("/scm/inventory/trx/gets-filter", payload)
    .then(
      (r) => {
        r.data.data.map(function (i) {
          let isCheck = data.listChecked.find(function (v) {
            return (
              v.SKU == i.SKU &&
              v.Item._id == i.Item._id &&
              v.SourceJournalID == i.SourceJournalID &&
              v.SourceLineNo == i.SourceLineNo
            );
          });
          i.QtyLine = isCheck ? isCheck.Qty : 0;
          if (data.records.Status) {
            i.Qty = isCheck ? isCheck.Qty : 0;
          } else {
            i.Qty = i.TrxQty;
          }

          i.isSelected = isCheck ? true : false;
          i.BatchSerials = isCheck ? isCheck.BatchSerials : [];
          i.WarehouseID = i.InventDim.WarehouseID;
          i.AisleID = i.InventDim.AisleID;
          i.SectionID = i.InventDim.SectionID;
          i.BoxID = i.InventDim.BoxID;
          i.VariantID = i.InventDim.VariantID;
          i.Size = i.InventDim.Size;
          i.Grade = i.InventDim.Grade;
          i.TrxUnitID = i.InventJournalLine
            ? i.InventJournalLine.UnitID
            : i.TrxUnitID;
          i.UnitID = i.InventJournalLine
            ? i.InventJournalLine.UnitID
            : i.TrxUnitID;
          i.CostPerUnit = i.Item.CostUnit;
          i.UnitCost = i.InventJournalLine ? i.InventJournalLine.UnitCost : 0;
          i.ItemID = i.Item._id;
        });
        Selected = r.data.data;
        data.listChecked = Selected;
      },
      (e) => {
        data.loading.processItem = false;
        util.showError(e);
      }
    )
    .finally(() => {
      if (Selected.length > 0) {
        const payload = JSON.parse(JSON.stringify(Selected)).map((i) => {
          return {
            ItemID: i.ItemID,
            SourceJournalID: i.SourceJournalID,
            InventDim: i.InventDim,
          };
        });

        axios.post(`/scm/item/get-unit-cost`, payload).then(
          (r) => {
            Selected.map((i) => {
              if (
                i.SourceType == "INVENTORY" &&
                route.query.type == "Inventory Receive"
              ) {
                const items = r.data.find(function (r) {
                  return r.ItemID == i.ItemID;
                });
                if (items) {
                  i.UnitCost = items.Cost;
                }
              }
              return i;
            });

            data.valueLine = Selected;
            data.formModeInventTrx = "edit";
            if (!data.records._id) {
              data.records.Name = `${Object.keys(group)[0]} ${moment().format(
                "DDMMYYYY"
              )}`;
            }
            util.nextTickN(2, () => {
              const uniqReffNo = [
                ...new Set(
                  Selected.map(function (sj) {
                    return sj.SourceJournalID;
                  })
                ),
              ];
              data.records.ReffNo = uniqReffNo;
              getPostingProfile(data.records);
            });
            data.loading.processItem = false;
          },
          (e) => {
            data.loading.processItem = false;
            util.showError(e);
          }
        );
      } else {
        data.loading.processItem = false;
        return util.showError("Please choose inventory transaction");
      }
    });

  // Selected.map(function (val) {
  //   const Line = data.valueItem.find(function (r) {
  //     return (
  //       r.Item._id == val.Item._id &&
  //       r.SKU == val.SKU &&
  //       r.SourceJournalID == val.SourceJournalID &&
  //       r.SourceLineNo == val.SourceLineNo
  //     );
  //   });
  //   val.LineNo = val.SourceLineNo;
  //   val.SourceLine = val.SourceLineNo;
  //   val.ItemID = val.Item._id;
  //   if (Line) {
  //     val.SettledQty = Line.SettledQty;
  //     val.TrxQty = Line.TrxQty;
  //     val.Qty = Line.QtyLine;
  //     val.ItemName = Line.ItemName;
  //     val.Dimension = Line.Dimension;
  //     val.InventJournalLine = Line.InventJournalLine;
  //     if (val.BatchSerials.length == 0) {
  //       val.BatchSerials = Line ? Line.BatchSerials : [];
  //     }
  //   }
  //   return val;
  // });
}

function onSelectLineDIM(record, index) {
  data.appLine = "InventDim";
  data.batchSNID = `${record.ItemID}|${record.SKU}|${record.SourceJournalID}`;
  data.disableField = ["VariantID", "Size", "Grade", "SpecID"];
  util.nextTickN(2, () => {
    // data.valueInventDim = {
    //   InventDim: record.InventDim,
    //   Dimension: record.Dimension,
    // };
    data.valueInventDim = record;
    if (record.InventDim.WarehouseID) {
      data.disableField.push("WarehouseID");
    }
    if (!["", "DRAFT"].includes(data.records.Status)) {
      data.disableField.push("AisleID");
    }
    if (!["", "DRAFT"].includes(data.records.Status)) {
      data.disableField.push("SectionID");
    }
    if (!["", "DRAFT"].includes(data.records.Status)) {
      data.disableField.push("BoxID");
    }
    data.keyDimItem = util.uuid();
  });
}

function onSelectLineBatchSN(record, index) {
  data.appLine = "BatchSN";
  data.batchSNID = `${record.ItemID}|${record.SKU}|${record.SourceJournalID}`;
  util.nextTickN(2, () => {
    data.valueBatchSN = record.BatchSerials.map(function (bs) {
      bs._id = data.batchSNID;
      return bs;
    });
  });
}

function newRecordBatchSN() {
  const record = {};
  record._id = data.batchSNID;
  record.BatchID = "";
  record.SerialNumber = "";
  record.Qty = 1;
  BatchSNControl.value.setRecords([
    ...BatchSNControl.value.getRecords(),
    record,
  ]);
  data.valueBatchSN = BatchSNControl.value.getRecords();
}

function deleteRecordBatchSN(record, index) {
  const newRecords = record.items.filter((dt, idx) => {
    return idx != index;
  });
  data.valueBatchSN = newRecords;
}

function GridTrxRefreshed(param = {}) {
  if (listControlTrx.value) {
    listControlTrx.value.setLoading(true);
  }
  if (["Inventory Receive", "Inventory Issuance"].includes(route.query.type)) {
    let payload = param;
    axios
      .post(`/scm/inventory/receive/gets?TrxType=${data.trxType}`, payload)
      .then(
        (r) => {
          data.recordsCount = r.data.count;
          listControlTrx.value.setRecords(r.data.data);
        },
        (e) => util.showError(e)
      )
      .finally(() => {
        util.nextTickN(2, () => {
          if (listControlTrx.value) {
            listControlTrx.value.setLoading(false);
          }
        });
      });
  }
}

function changePageSizeItem(pageSize) {
  data.pagingItems.pageSize = pageSize;
  data.pagingItems.currentPage = 1;
  GridRefreshed();
}
function changePageItem(page) {
  data.pagingItems.currentPage = page;
  GridRefreshed();
}
function onFilterRefreshItems(val) {
  data.pagingItems = {
    skip: 0,
    pageSize: 25,
    currentPage: 1,
  };
  util.nextTickN(2, () => {
    GridRefreshed(true);
  });
}
function GridRefreshed(isLoad = false) {
  // =========== remove jika suimjs udh update ===============
  // data.valueLine = [];
  // =========================================================
  data.alertPosted = false;
  data.loading.processItem = true;
  if (listControl.value) {
    listControl.value.setLoading(true);
  }
  if (formCtlInventTrx.value) {
    formCtlInventTrx.value.setLoading(true);
  }
  const isMode = ["POSTED"].includes(data.records.Status);
  let status =
    route.query.type == "Inventory Receive" ? ["Planned"] : ["Reserved"];
  let SourceTrxTypes =
    route.query.type == "Inventory Receive"
      ? ["Transfer", "Purchase Order", "Inventory Receive"]
      : ["Transfer", "Sales Order", "Inventory Issuance"];

  let payload = {
    ...data.search,
    ...{ Statuses: status, SourceTrxTypes: SourceTrxTypes },
    Skip: (data.pagingItems.currentPage - 1) * data.pagingItems.pageSize,
    Take: data.pagingItems.pageSize,
    SortSourceJournalIDs: [],
  };
  if (data.listChecked.length > 0) {
    payload.SortSourceJournalIDs = [
      ...new Set(
        data.listChecked.map((i) => {
          return i.SourceJournalID;
        })
      ),
    ];
  }
  if (!payload.WarehouseID) {
    data.valueItem = [];
    data.loading.processItem = false;
    if (listControl.value) {
      listControl.value.setLoading(false);
    }
    if (formCtlInventTrx.value) {
      formCtlInventTrx.value.setLoading(false);
    }
    data.isProcess = false;
    return util.showError("Please select warehouse");
  }
  axios
    .post("/scm/inventory/trx/gets-filter", payload)
    .then(
      (r) => {
        r.data.data.map(function (i) {
          let isCheck = data.listChecked.find(function (v) {
            return (
              v.SKU == i.SKU &&
              v.Item._id == i.Item._id &&
              v.SourceJournalID == i.SourceJournalID &&
              v.SourceLineNo == i.SourceLineNo
            );
          });
          i.QtyLine = isCheck ? isCheck.Qty : 0;
          i.isSelected = isCheck ? true : false;
          i.BatchSerials = isCheck ? isCheck.BatchSerials : [];
          i.WarehouseID = i.InventDim.WarehouseID;
          i.AisleID = i.InventDim.AisleID;
          i.SectionID = i.InventDim.SectionID;
          i.BoxID = i.InventDim.BoxID;
          i.VariantID = i.InventDim.VariantID;
          i.Size = i.InventDim.Size;
          i.Grade = i.InventDim.Grade;
          i.TrxUnitID = i.InventJournalLine
            ? i.InventJournalLine.UnitID
            : i.TrxUnitID;
          i.UnitID = i.InventJournalLine
            ? i.InventJournalLine.UnitID
            : i.TrxUnitID;
          i.CostPerUnit = i.Item.CostUnit;
          i.UnitCost = i.InventJournalLine ? i.InventJournalLine.UnitCost : 0;
          i.ItemID = i.Item._id;
          if (isMode) {
            i.OriginalQty = isCheck ? isCheck.OriginalQty : 0;
            i.UnitCost = isCheck ? isCheck.UnitCost : 0;
            i.SettledQty = isCheck ? isCheck.SettledQty : 0;
          }
        });

        data.valueItem = r.data.data;
        const lines = JSON.parse(JSON.stringify(r.data.data));
        data.valueLine = lines
          .map(function (l) {
            l.Qty = l.QtyLine;
            return l;
          })
          .filter(function (val) {
            return val.isSelected == true;
          });
        const uniqReffNo = [
          ...new Set(
            data.valueLine.map(function (sj) {
              return sj.SourceJournalID;
            })
          ),
        ];
        data.records.ReffNo = uniqReffNo;

        if (["SUBMITTED", "POSTED", "READY"].includes(data.records.Status)) {
          data.formModeInventTrx = "view";
        }

        if (data.records.Status == "READY" && data.valueLine.length == 0) {
          data.alertPosted = true;
        }

        setTimeout(() => {
          data.countFulfillment = r.data.count;
          data.itemFulfillment = r.data.data;
        }, 500);
      },
      (e) => util.showError("Please select warehouse")
    )
    .finally(() => {
      util.nextTickN(2, () => {
        if (listControl.value) {
          listControl.value.setLoading(false);
          data.loading.processItem = false;
        }
        if (formCtlInventTrx.value) {
          formCtlInventTrx.value.setLoading(false);
        }
        data.isProcess = false;
      });
    });
}

function genCfgInventDim() {
  const cfg = createFormConfig("", true);
  cfg.addSection("General1", false).addRow(
    {
      field: "General1",
      kind: "text",
      label: "",
    },
    {
      field: "General2",
      kind: "text",
      label: "",
    },
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

function genCfgBatchSN() {
  const BatchSN = ["BatchID", "SerialNumber", "Qty"];
  let CfgBatchSN = [];
  for (let index = 0; index < BatchSN.length; index++) {
    CfgBatchSN.push({
      field: BatchSN[index],
      kind: "Text",
      label: BatchSN[index],
      readType: "show",
      input: {
        field: BatchSN[index],
        label: BatchSN[index],
        hint: "",
        hide: false,
        placeHolder: BatchSN[index],
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    });
  }
  data.gridCfgBatchSN = {
    setting: { idField: "", keywordFields: ["_id", "Name"], sortable: ["_id"] },
    fields: CfgBatchSN,
  };
}

function setDefaultDimension(nv, mode) {
  data.defaultListWarehouse = nv.filter((item) => item.Key === "WarehouseID");
  if (data.defaultListWarehouse.length > 0 && mode == "new") {
    data.search.WarehouseID = data.defaultListWarehouse[0].Value;
    util.nextTickN(2, () => {
      GridRefreshed();
    });
  } else if (mode == "new") {
    axios
      .post("/tenant/warehouse/find", lookupPayloadBuilder("", ["_id", "Name"]))
      .then(
        (r) => {
          if (r.data.length > 0) {
            data.search.WarehouseID = r.data[0]._id;
          }
          util.nextTickN(2, () => {
            GridRefreshed();
          });
        },
        (e) => util.showError(e)
      )
      .finally(function () {});
  }
}

function setDefaultListWH(nv) {
  const list = nv.map((o) => o.Value);
  const lookuplabels = ["Name", "_id"];
  data.warehouseparams = {};
  data.warehouseparams.Take = 20;
  data.warehouseparams.Sort = [lookuplabels[0]];
  data.warehouseparams.Select = lookuplabels;
  data.warehouseparams.Where = {
    Op: "$or",
    Items: [
      {
        Field: "_id",
        Op: "$contains",
        Value: list.length > 0 ? list : [""],
      },
      {
        Field: "Name",
        Op: "$contains",
        Value: list.length > 0 ? list : [""],
      },
    ],
  };
}

function lookupPayloadBuilder(search, select, value, item) {
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
      Op: "$in",
      Value: defaultList,
    },
  ];
  if (defaultList.length > 0) {
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
    if (defaultList.length > 0) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}

function onFieldChange(field, value1, value2, oldValue) {
  switch (field) {
    case "SectionID":
      data.valueLine.map(function (l) {
        l.SectionID = value1;
        l.InventDim.SectionID = value1;
        l.InventJournalLine.InventDim.SectionID = value1;
        return l;
      });
      break;
    default:
      break;
  }
}
function onAlterFormConfig(cfg) {
  cfg.sectionGroups = cfg.sectionGroups.map((sectionGroup) => {
    sectionGroup.sections = sectionGroup.sections.map((section) => {
      section.rows.map((row) => {
        row.inputs = row.inputs.map((input) => {
          if (["VariantID", "Size", "Grade", "SpecID"].includes(input.field)) {
            input.readOnly = true;
          }
          return input;
        });
        return row;
      });
      return section;
    });
    return sectionGroup;
  });
}

function getReffNo() {
  data.listReffNo = [];
  axios
    .post("/scm/postingprofile/find-journal-ref", {
      JournalType: route.query.title,
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
function loadGridConfigItems() {
  data.gridCfg = {};
  data.gridLineCfg = {};
  loadGridConfig(axios, `/scm/inventory/trx/receipt/gridconfig`).then(
    (r) => {
      let hideColm = [
        "_id",
        "OriginDestinationID",
        "InventDim",
        "InventJournalLine",
        "Text",
        "SKU",
        "WarehouseID",
        "UnitID",
        "SKUName",
        "SourceLineNo",
        "VendorID",
      ];
      const sortColm = [
        "TrxDate",
        "SourceType",
        "SourceTrxType",
        "SourceJournalID",
        "SourceLineNo",
        "VendorID",
        "VendorName",
        "OriginDestinationName",
        "Item",
        "SKU",
        "Qty",
        "TrxUnitID",
        "TrxQty",
        "SettledQty",
        "OriginalQty",
      ];
      const addColms = [];
      let colums = [];
      for (let index = 0; index < addColms.length; index++) {
        colums.push({
          field: addColms[index].field,
          kind: addColms[index].kind,
          label: addColms[index].label,
          readType: "show",
          input: {
            field: addColms[index].field,
            label: addColms[index].label,
            hint: "",
            hide: false,
            placeHolder: addColms[index].label,
            kind: addColms[index].kind,
            disable: false,
            required: false,
            multiple: false,
          },
        });
      }
      const _fields = [...r.fields, ...colums].filter((o) => {
        o.idx = sortColm.indexOf(o.field);
        return !hideColm.includes(o.field);
      });

      _fields.map((f) => {
        if (
          f.field == "OriginDestinationName" &&
          route.query.type == "Inventory Issuance"
        ) {
          f.label = "Destination";
        } else if (
          f.field == "OriginDestinationName" &&
          route.query.type == "Inventory Receive"
        ) {
          f.label = "Origin";
        } else if (f.field == "Item") {
          f.label = "Item Varian";
        }
        return f;
      });
      data.gridCfg = {
        ...r,
        fields: _fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
    },
    (e) => util.showError(e)
  );
  loadGridConfig(axios, `/scm/inventory/receive/line/gridconfig`).then(
    (r) => {
      let hideColm = [
        "OtherExpenses",
        "SourceLineNo",
        "InventJournalLine",
        "Item",
        "LineNo",
        "InventQty",
        "CostPerUnit",
        "RemainingQty",
        "Text",
        "SKU",
        "References",
        "TaxCodes",
        "DiscountGeneral",
        "DiscountAmount",
        "DiscountValue",
        "DiscountType",
        "OriginDestinationName",
        "VendorName",
      ];
      const Line = [
        "LineNo",
        "SKU",
        "UnitID",
        "Text",
        //"TrxQty",
        "UnitCost",
        "RemainingQty",
        "Qty",
      ];
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
        "RemainingQty",
        "InventQty",
        "OriginalQty",
        "SettledQty",
        "TrxQty",
        "CostPerUnit",
        "UnitCost",
        "Qty",
      ];
      let InventJournalLine = [];
      for (let index = 0; index < Line.length; index++) {
        let label = Line[index];
        if (Line[index].match(/[A-Z][a-z]+/g)) {
          label = Line[index].match(/[A-Z][a-z]+/g).join(" ");
        }
        if (label == "Unit") {
          label = "UOM";
        }
        InventJournalLine.push({
          field: Line[index],
          kind: [
            "RemainingQty",
            "OriginalQty",
            "SettledQty",
            "TrxQty",
            "UnitCost",
            "Qty",
          ].includes(Line[index])
            ? "number"
            : "text",
          label: label,
          readType: "show",
          labelField: "",
          input: {
            field: Line[index],
            label: label,
            hint: "",
            hide: false,
            placeHolder: label,
            kind: [
              "RemainingQty",
              "OriginalQty",
              "SettledQty",
              "TrxQty",
              "Qty",
            ].includes(Line[index])
              ? "number"
              : "text",
          },
        });
      }
      const _fields = [...r.fields, ...InventJournalLine].filter((o) => {
        if (
          [
            "RemainingQty",
            "OriginalQty",
            "SettledQty",
            "TrxQty",
            "CostPerUnit",
            "Qty",
          ].includes(o.field)
        ) {
          o.width = "300px";
        } else {
          o.width = "400px";
        }

        if (o.field == "VendorID" && route.query.type == "Inventory Issuance") {
          o.label = "Destination";
        } else if (
          o.field == "VendorID" &&
          route.query.type == "Inventory Receive"
        ) {
          o.label = "Origin";
        } else if (o.field == "ItemName") {
          o.label = "Item Varian";
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
}
watch(
  () => route.query.type,
  (nv) => {
    data.trxType = route.query.type;
    data.appMode = "";
    data.isPreview = false;
    data.InventTrx = {
      by: "_id",
      search: "",
      Status: "All",
      Site: "",
      TrxDate: "",
    };
    util.nextTickN(2, () => {
      if (nv) {
        data.listChecked = [];
        refreshRecords();
        loadGridConfigItems();
      }
    });
  }
);
watch(
  () => data.valueBatchSN,
  (nv) => {
    data.valueLine.map(function (l) {
      const batchSNID = `${l.Item._id}|${l.SKU}|${l.SourceJournalID}`;
      if (batchSNID == data.batchSNID) {
        l.BatchSerials = nv;
      }
      return l;
    });
  },
  { deep: true }
);

watch(
  () => data.valueInventDim,
  (nv) => {
    data.valueLine.map(function (l) {
      const batchSNID = `${l.Item._id}|${l.SKU}|${l.SourceJournalID}`;
      if (batchSNID == data.batchSNID) {
        l.Dimension = nv.Dimension;
        l.InventDim = nv.InventDim;
      }
      return l;
    });
  },
  { deep: true }
);

watch(
  () => data.defaultListWarehouse,
  (nv) => {
    if (nv.length > 0) {
      setDefaultListWH(nv);
    }
  },
  { deep: true }
);

onMounted(() => {
  if (defaultList.length == 1) {
    data.InventTrx.Site = defaultList[0];
  }
  createGridCfgRefNo();
  loadGridConfigItems();
  loadFormConfig(axios, "/scm/inventory/receive/formconfig").then(
    (r) => {
      r.sectionGroups = r.sectionGroups.map((sectionGroup) => {
        sectionGroup.sections = sectionGroup.sections.map((section) => {
          section.rows.map((row) => {
            row.inputs.map((input) => {
              if (input.field == "PostingProfileID") {
                input.hide = true;
              }
              if (input.field == "ReffNo") {
                input.readOnly = true;
              }
              return input;
            });
            return row;
          });
          return section;
        });
        return sectionGroup;
      });
      data.frmCfg = r;
      if (route.query.trxid !== undefined) {
        let currQuery = { ...route.query };
        listControl.value.selectData({ _id: currQuery.trxid }); //remark sementara tunggu suimjs update
        delete currQuery["trxid"];
        router.replace({ path: route.path, query: currQuery });
      }
    },
    (e) => util.showError(e)
  );
  loadGridConfig(axios, `/scm/inventory/receive/gridconfig`).then(
    (r) => {
      let hideColm = ["Dimension", "TrxType"];
      const Line = [
        {
          field: "Created",
          kind: "date",
          label: "Created date",
        },
        // {
        //   field: "Approvers",
        //   kind: "text",
        //   label: "Next Approval",
        // },
      ];
      let addColm = [];
      for (let index = 0; index < Line.length; index++) {
        addColm.push({
          field: Line[index].field,
          kind: Line[index].kind,
          label: Line[index].label,
          readType: "show",
          labelField: "",
          input: {
            field: Line[index].field,
            label: Line[index].label,
            hint: "",
            hide: false,
            placeHolder: Line[index].label,
            kind: Line[index].kind,
          },
        });
      }
      r.fields.map((f) => {
        if (f.field == "ReffNo") {
          f.width = "600px";
        }
        return r;
      });
      let sortColm = [
        "_id",
        "Name",
        "TrxDate",
        "Created",
        "ReffNo",
        "WarehouseID",
        "SectionID",
        "Approvers",
        "Status",
      ];
      let _fields = [...r.fields, ...addColm].filter((o) => {
        if (o.field == "WarehouseID") {
          o.label = "Warehouse";
        }
        return !hideColm.includes(o.field);
      });
      _fields.map((f) => {
        f.idx = sortColm.indexOf(f.field);
        return;
      });
      data.gridCfgTrx = {
        ...r,
        fields: _fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
      refreshRecords();
    },
    (e) => util.showError(e)
  );

  genCfgInventDim();
  genCfgBatchSN();
  if (route.query.id !== undefined) {
    let getUrlParam = route.query.id;
    axios
      .post(`/scm/inventory/receive/get`, [getUrlParam])
      .then(
        (r) => {
          onSelectTex(r.data);
        },
        (e) => util.showError(e)
      )
      .finally(() => {
        util.nextTickN(2, () => {
          router.replace({
            query: {
              type: route.query.type,
              title: route.query.title,
            },
          });
        });
      });
  }
});

function onPreview() {
  data.isPreview = true;
}

function closePreview() {
  data.isPreview = false;
}
</script>
<style>
.row_action {
  justify-content: center;
  align-items: center;
  gap: 2px;
}
</style>
