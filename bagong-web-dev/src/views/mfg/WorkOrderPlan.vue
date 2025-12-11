<template>
  <div class="w-full">
    <data-list
      v-show="data.isPreview == false"
      class="card grid-line-items"
      ref="listControl"
      :title="data.titleForm"
      grid-config="/mfg/workorderplan/gridconfig"
      form-config="/mfg/workorderplan/formconfig"
      grid-read="/mfg/workorderplan/gets"
      form-read="/mfg/workorderplan/get"
      grid-mode="grid"
      grid-delete="/mfg/workorderplan/delete"
      form-keep-label
      form-insert="/mfg/workorderplan/save"
      form-update="/mfg/workorderplan/save"
      grid-sort-field="LastUpdate"
      grid-sort-direction="desc"
      :form-hide-submit="true"
      :grid-fields="[
        'RequestorWOName',
        'RequestorDepartment',
        'Asset',
        'Dimension',
        'Status',
        'SiteName',
      ]"
      :form-fields="[
        'WOName',
        'RequestorName',
        'RequestorWOName',
        'JournalTypeID',
        'WoTypeKind',
        'RequestorDepartment',
        'TrxCreatedDate',
        'Asset',
        'WRDate',
        'StartDownTime',
        'ExpectedCompletedDate',
        'Merk',
        'UnitType',
        'BOM',
        'Dimension',
        'InventDim',
        'Summary',
        'TrxDate',
        'SafetyInstruction',
        'Status',
        'BreakdownType',
        'Kilometers',
        'WRDescription',
      ]"
      :form-tabs-new="data.formTabs"
      :form-tabs-edit="data.formTabs"
      :form-tabs-view="data.formTabs"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :stayOnFormAfterSave="data.isOnFormAfter"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :grid-custom-filter="customFilter"
      @alterGridConfig="onAlterGridConfig"
      @alterFormConfig="onAlterFormConfig"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @pre-save="preSave"
      @post-save="postSave"
      @form-field-change="onFormFieldChange"
      @controlModeChanged="onControlModeChanged"
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
          class="w-full"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="WO Date From"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="WO Date To"
          v-model="data.search.DateTo"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refPolisiNo"
          v-model="data.search.PolisiNo"
          lookup-key="_id"
          label="Police No"
          class="w-full"
          use-list
          :lookup-url="`/tenant/asset/find?GroupID=UNT`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :lookup-payload-builder="
            (search) =>
              lookupPayloadSearch(
                search,
                ['_id', 'Name'],
                data.search.Site,
                item
              )
          "
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refWOType"
          v-model="data.search.WoType"
          lookup-key="_id"
          label="WO Type"
          class="w-full"
          use-list
          :lookup-url="`/mfg/workorder/journal/type/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refStatus"
          v-model="data.search.Status"
          lookup-key="_id"
          label="Status"
          class="w-full"
          use-list
          :items="['DRAFT', 'SUBMITTED', 'READY', 'POSTED', 'REJECTED']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refSite"
          v-model="data.search.Site"
          lookup-key="_id"
          label="Site"
          class="w-full"
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
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #grid_RequestorWOName="{ item }">
        {{ item.RequestorNameFix }}
      </template>
      <template #grid_SiteName="{ item }">
        {{ item.SiteName }}
      </template>
      <template #grid_RequestorDepartment="{ item }">
        {{ item.DepartmentName }}
      </template>
      <template #grid_Dimension="{ item }">
        {{ item.SiteName }}
      </template>
      <template #grid_Asset="{ item }">
        {{ item.AssetName }}
      </template>
      <template #form_tab_Attachment="{ item }">
        <s-grid-attachment
          v-model="item.Attachment"
          ref="gridAttachment"
          journalType="WorkOrder"
          :isUpdateTags="reffTags.length > 0 ? true : false"
          :journalId="item._id"
          :tags="linesTag"
          :reff-tags="reffTags"
          :readOnly="!['', 'DRAFT'].includes(item.Status)"
          @pre-Save="preSaveAttachment"
        />
      </template>
      <template #form_tab_Report="{ item }">
        <s-grid
          v-if="data.isGridDailyReport"
          v-model="data.listDailyReport"
          ref="gridDailyRptControl"
          class="w-full grid-line-items"
          hide-search
          hide-sort
          :hide-new-button="
            item.Status != 'POSTED' || item.StatusOverall != 'IN PROGRESS'
          "
          :hide-delete-button="true"
          hide-refresh-button
          :hide-detail="false"
          :hide-action="false"
          hide-select
          auto-commit-line
          no-confirm-delete
          :config="data.gridCfgDailyReport"
          form-keep-label
          hide-paging
          @new-data="newRecordDailyReport"
          @select-data="getReport"
        >
        </s-grid>
        <s-form
          v-else
          ref="formDailyRptControl"
          v-model="data.dialogFrmDailyReport"
          :keep-label="true"
          :config="data.formCfgDailyReport"
          class="pt-2"
          :auto-focus="true"
          :hide-submit="true"
          :hide-cancel="true"
          mode="view"
        >
          <template #input_DailyStatus="{ item }">
            <s-input
              label="Daily Status"
              v-model="item.DailyStatus"
              class="w-full"
              use-list
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=WODailyStatus"
              lookup-key="_id"
              :lookup-labels="['_id', 'Name']"
              :lookup-searchs="['_id', 'Name']"
              :lookup-payload-builder="
                (search) =>
                  lookupDailyStatus(
                    search,
                    ['_id', 'Name'],
                    item.DailyStatus,
                    item
                  )
              "
              :disabled="
                !['', 'DRAFT'].includes(data.dialogFrmDailyReport.Status) ||
                data.record.StatusOverall == 'END'
              "
              :required="false"
              :keepErrorSection="false"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  item.MonitoringStatus = '';
                }
              "
            ></s-input>
          </template>
          <template #input_MonitoringStatus="{ item }">
            <s-input
              ref="refDailyStatus"
              v-model="item.MonitoringStatus"
              label="Monitoring Status"
              class="w-full"
              use-list
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=WODailyStatus"
              lookup-key="_id"
              :lookup-labels="['_id', 'Name']"
              :lookup-searchs="['_id', 'Name']"
              :lookup-payload-builder="
                (search) =>
                  lookupMonitoringStatus(
                    search,
                    ['_id', 'Name'],
                    item.MonitoringStatus,
                    item
                  )
              "
              :disabled="
                !['', 'DRAFT'].includes(data.dialogFrmDailyReport.Status) ||
                data.record.StatusOverall == 'END'
              "
              :required="false"
              :keepErrorSection="false"
            ></s-input>
          </template>
          <template #input_ComponentCategory="{ item }">
            <s-input
              ref="refComponentCategory"
              v-model="item.ComponentCategory"
              label="Component Category"
              class="w-full"
              use-list
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=WOComponentCategory"
              lookup-key="_id"
              :lookup-labels="['_id', 'Name']"
              :lookup-searchs="['_id', 'Name']"
              :disabled="
                !['', 'DRAFT'].includes(data.dialogFrmDailyReport.Status) ||
                data.record.StatusOverall == 'END'
              "
              :required="false"
              :keepErrorSection="false"
            ></s-input>
          </template>
          <template #input_Consumption="{ item }">
            <div class="material-consumption">
              <work-order-rpt-consumption
                ref="LineRptConsumptionPlan"
                typeMaterail="plan"
                :item="data.dialogFrmDailyReport"
                :general="data.record"
                :plan="data.planMaterial"
                :warehouseID="data.warehouseID"
                @preSubmit="preSubmitRptConsumption"
                @postSubmit="postSubmitRptConsumption"
                @errorSubmit="errorSubmitRptConsumption"
                @preReopen="preReopenRptConsumption"
              ></work-order-rpt-consumption>
              <work-order-rpt-consumption
                ref="LineRptConsumption"
                typeMaterail="additional"
                :item="data.dialogFrmDailyReport"
                :general="data.record"
                :plan="data.planMaterial"
                :warehouseID="data.warehouseID"
                @preSubmit="preSubmitRptConsumption"
                @postSubmit="postSubmitRptConsumption"
                @errorSubmit="errorSubmitRptConsumption"
                @createItemRequest="createItemRequest"
                @getAvailableStock="getAvailableStock"
              ></work-order-rpt-consumption>
            </div>
          </template>
          <template #input_Resource="{ item }">
            <work-order-rpt-resource
              ref="LineRptResource"
              :item="data.dialogFrmDailyReport"
              :general="data.record"
              :plan="data.planResource"
              :listPlan="data.listRptResource"
              @preSubmit="preSubmitRptResource"
              @postSubmit="postSubmitRptResource"
              @errorSubmit="errorSubmitRptResource"
              @preReopen="preReopenRptResource"
            ></work-order-rpt-resource>
          </template>
          <template #input_Output="{ item }">
            <work-order-rpt-output
              ref="LineRptOutput"
              :item="data.dialogFrmDailyReport"
              :general="data.record"
              :plan="data.planOutput"
              @preSubmit="preSubmitRptOutput"
              @postSubmit="postSubmitRptOutput"
              @errorSubmit="errorSubmitRptOutput"
              @preReopen="preReopenRptOutput"
            ></work-order-rpt-output>
          </template>
          <template #buttons="{ item }">
            <s-button
              :icon="`rewind`"
              class="btn_warning back_btn"
              :label="'Back to Report'"
              @click="backToReport"
            />
          </template>
          <template #buttons_1="{ item }">
            <div class="flex gap-2">
              <div class="w-[145px] h-[30px]" v-if="data.isLoadingBtnDaily">
                <loader kind="skeleton" skeleton-kind="input" />
              </div>
              <div class="flex gap-2" v-else>
                <div
                  v-if="
                    ['', 'DRAFT'].includes(item.Status) &&
                    data.record.StatusOverall == 'IN PROGRESS'
                  "
                  class="flex gap-1"
                >
                  <s-button
                    :icon="`content-save`"
                    class="btn_primary submit_btn"
                    :label="'Save As Draft'"
                    @click="saveAsReport('Save')"
                  />
                </div>
                <div
                  v-if="
                    item.Status == 'DRAFT' &&
                    data.record.StatusOverall == 'IN PROGRESS'
                  "
                  class="flex gap-1"
                >
                  <s-button
                    :icon="`content-save-all`"
                    class="btn_primary submit_btn"
                    :label="'Submit'"
                    @click="saveAsReport('Submit')"
                  />
                </div>
              </div>
            </div>
          </template>
        </s-form>
      </template>
      <template #form_tab_Material="{ item }">
        <work-order-sum
          ref="LineSUMMaterial"
          :item="item"
          gridConfig="/mfg/workorderplan/tab/material/gridconfig"
          type="Material"
        ></work-order-sum>
      </template>
      <template #form_tab_Resource="{ item }">
        <work-order-sum
          ref="LineSUMResource"
          :item="item"
          gridConfig="/mfg/workorderplan/tab/resource/gridconfig"
          type="Resource"
        ></work-order-sum>
      </template>
      <template #form_tab_Output="{ item }">
        <work-order-sum
          ref="LineSUMOutput"
          :item="item"
          gridConfig="/mfg/workorderplan/tab/output/gridconfig"
          type="Output"
        ></work-order-sum>
      </template>

      <template #form_input_WOName="{ item, config }">
        <s-input
          ref="refRequired"
          label="WO Name"
          v-model="item.WOName"
          class="w-full"
          :required="true"
          :use-list="item.JournalTypeID === 'WO_PrevMaintenance'"
          lookup-key="_id"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=WOPreventiveName`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_input_RequestorName="{ item, config }">
        <s-input
          ref="refRequired"
          label="WR Requestor"
          v-model="item.RequestorName"
          class="w-full"
          :disabled="true"
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onChangeRequestor(v1, item);
            }
          "
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilder(search, config, item.RequestorName, item)
          "
        ></s-input>
      </template>
      <template #form_input_JournalTypeID="{ item, config }">
        <div class="w-full items-start gap-2 grid gridCol2">
          <div class="col-auto">
            <s-input
              ref="refWOType"
              label="WO Type"
              v-model="item.JournalTypeID"
              class="w-full"
              :required="true"
              :disabled="
                data.statusDisabled.includes(item.Status) || data.disabledWOtype
              "
              use-list
              :lookup-url="`/mfg/workorder/journal/type/find`"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  item.WOName = '';
                  item.WoTypeKind = '';
                  data.listWoTypeKind = [];
                  getCfgMaterial();
                  getJournalType(v1, item);
                }
              "
              :lookup-payload-builder="
                (search) =>
                  lookupPayloadBuilderWOType(search, config, item.Asset, item)
              "
            ></s-input>
          </div>
          <div class="col-auto">
            <s-input
              ref="refRequestorWOName"
              label="WO Requestor"
              v-model="item.RequestorWOName"
              class="w-full"
              :disabled="
                data.statusDisabled.includes(item.Status) ||
                data.formMode == 'view'
              "
              use-list
              :lookup-url="`/tenant/employee/find`"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
            ></s-input>
          </div>
        </div>
      </template>
      <template #form_input_TrxDate="{ item, config }">
        <div class="w-full items-start gap-2 grid gridCol2">
          <div class="col-auto">
            <s-input
              ref="refWODate"
              label="WO Date"
              v-model="item.TrxDate"
              class="w-full"
              kind="date"
              :disabled="
                data.statusDisabled.includes(item.Status) ||
                data.formMode == 'view'
              "
            ></s-input>
          </div>
          <div class="col-auto">
            <s-input
              ref="refPriority"
              label="Priority"
              v-model="item.Priority"
              class="w-full"
              :disabled="
                data.statusDisabled.includes(item.Status) ||
                data.formMode == 'view'
              "
              use-list
              :items="['Top', 'Middle', 'Low']"
            ></s-input>
          </div>
        </div>
      </template>
      <template #form_input_SafetyInstruction="{ item, config }">
        <div class="w-full items-start gap-2 grid gridCol2">
          <div class="col-auto">
            <s-input
              ref="refSafetyInstruction"
              label="Safety Instruction"
              v-model="item.SafetyInstruction"
              class="w-full"
              :disabled="
                data.statusDisabled.includes(item.Status) ||
                data.formMode == 'view'
              "
              use-list
              :multiple="true"
              :items="[
                'Safety helmet',
                'Safety shoes',
                'Ear muff',
                'Ear plug',
                'Mask',
                'Face shield',
                'Reflector vest',
                'Apron',
                'Gloves',
              ]"
            ></s-input>
          </div>
          <div class="col-auto">
            <s-input
              ref="refBOM"
              v-model="item.BOM"
              :disabled="data.statusDisabled.includes(item.Status)"
              label="BOM"
              use-list
              :use-list="true"
              :lookup-url="`/mfg/bom/find`"
              lookup-key="_id"
              :lookup-labels="['_id', 'Title']"
              :lookup-searchs="['_id', 'Title']"
              keepLabel
              class="w-full"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  onFormFieldChange('BOM', v1, v2, old, item);
                }
              "
            ></s-input>
          </div>
        </div>
      </template>
      <template #form_input_Status="{ item, config }">
        <div class="w-full items-start gap-2 grid gridCol3">
          <div class="col-auto">
            <s-input
              ref="refStatus"
              label="Status"
              v-model="item.Status"
              class="w-full"
              :read-only="true"
            ></s-input>
          </div>
          <div class="col-auto">
            <s-input
              ref="refStatusOverall"
              label="Status Overall"
              v-model="item.StatusOverall"
              class="w-full"
              :read-only="true"
            ></s-input>
          </div>
          <div class="col-auto">
            <label class="input_label"><div>Trx Created Date</div></label>
            <div class="bg-transparent">
              {{
                item.TrxCreatedDate
                  ? moment(item.TrxCreatedDate).format("DD/MM/YYYY HH:mm")
                  : ""
              }}
            </div>
          </div>
        </div>
      </template>
      <template #form_input_BreakdownType="{ item, config }">
        <div class="w-full items-start gap-2 grid gridCol2">
          <div class="col-auto">
            <s-input
              ref="refBOM"
              v-model="item.BreakdownType"
              :disabled="data.statusDisabled.includes(item.Status)"
              label="Breakdown Type"
              use-list
              :use-list="true"
              :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=WOBreakdownType`"
              lookup-key="_id"
              :lookup-labels="['_id', 'Name']"
              :lookup-searchs="['_id', 'Name']"
              keepLabel
              class="w-full"
            ></s-input>
          </div>
          <div class="col-auto">
            <div class="w-full">
              <label class="input_label"><div>Trx Created Date</div></label>
              <div class="bg-transparent">
                {{
                  item.TrxCreatedDate
                    ? moment(item.TrxCreatedDate).format("DD/MM/YYYY HH:mm")
                    : ""
                }}
              </div>
            </div>
          </div>
        </div>
      </template>
      <template #form_input_WoTypeKind="{ item, config }">
        <div class="flex gap-2">
          <template v-for="(status, idx) in data.listWoTypeKind">
            <div class="radio-status">
              <input
                type="radio"
                class="bg-slate-800"
                v-model="item.WoTypeKind"
                :id="status._id + '_' + idx"
                :name="status.Name"
                :value="status._id"
              />
              <label :for="status._id + '_' + idx">{{ status.Name }}</label>
            </div>
          </template>
        </div>
      </template>
      <template #form_input_RequestorDepartment="{ item }">
        <s-input
          :key="data.keyDept"
          ref="refRequired"
          label="Requestor department"
          v-model="item.RequestorDepartment"
          class="w-full"
          :disabled="true"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=DME`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>

      <!-- start info WR/Unit -->
      <template #form_input_Asset="{ item, config }">
        <s-input
          ref="refAsset"
          v-model="item.Asset"
          :required="item.JournalTypeID == 'WOGeneral' ? false : true"
          :disabled="
            data.statusDisabled.includes(item.Status) ||
            item.WorkRequestID == undefined ||
            item.WorkRequestID != ''
          "
          label="Asset (NOPOL/Code Unit)"
          use-list
          :use-list="true"
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilder(search, config, item.Asset, item, true)
          "
          :lookup-url="`/tenant/asset/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          keepLabel
          class="w-full"
          @change="
            (field, v1, v2, old, ctlRef) => {
              getAsset(v1, item);
            }
          "
        ></s-input>
      </template>
      <template #form_input_WRDate="{ item, config }">
        <div
          v-if="
            [
              'WO_BLMaintenance',
              'WO_PrevMaintenance',
              'WO_PCR/OVHMaintenance',
            ].includes(item.JournalTypeID)
          "
        >
          <s-input
            v-if="!data.statusDisabled.includes(item.Status)"
            ref="refDate"
            v-model="item.Schedule"
            label="Schedule"
            kind="date"
            keepLabel
            class="w-full"
          ></s-input>
          <div v-else>
            <label class="input_label"><div>Schedule</div></label>
            <div class="bg-transparent">
              {{ moment(item.Schedule).format("DD/MM/YYYY") }}
            </div>
          </div>
        </div>
        <div v-else>
          <s-input
            v-if="!data.statusDisabled.includes(item.Status)"
            ref="refDate"
            v-model="item.WRDate"
            label="Report Date"
            kind="date"
            keepLabel
            class="w-full"
          ></s-input>
          <div v-else>
            <label class="input_label"><div>Report Date</div></label>
            <div class="bg-transparent">
              {{ moment(item.WRDate).format("DD/MM/YYYY") }}
            </div>
          </div>
        </div>
      </template>
      <template #form_input_Kilometers="{ item, config }">
        <s-input
          v-if="
            !data.statusDisabled.includes(item.Status) &&
            item.WorkRequestID == ''
          "
          ref="refKilometers"
          v-model="item.Kilometers"
          label="Kilometers"
          kind="number"
          keepLabel
          class="w-full"
        ></s-input>
      </template>
      <template #form_input_UnitType="{ item, config }">
        <s-input
          :key="data.keyDetailAsset"
          ref="refMerk"
          v-model="item.UnitType"
          :disabled="true"
          label="Unit Type"
          use-list
          :use-list="true"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=VTA`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          keepLabel
          class="w-full"
        ></s-input>
      </template>
      <template #form_input_Merk="{ item, config }">
        <s-input
          :key="data.keyDetailAsset"
          ref="refMerk"
          v-model="item.Merk"
          :disabled="true"
          label="Model Uint"
          use-list
          :use-list="true"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=MUA`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          keepLabel
          class="w-full"
        ></s-input>
      </template>
      <template #form_input_TrxCreatedDate="{ item, config }">
        <label class="input_label"><div>Trx Created Date</div></label>
        <div class="bg-transparent">
          {{
            item.TrxCreatedDate != "0001-01-01T00:00:00Z"
              ? moment(item.TrxCreatedDate).format("DD/MM/YYYY HH:mm")
              : "-"
          }}
        </div>
      </template>
      <template #form_input_WRDescription="{ item, config }">
        <!-- <s-input
          v-if="['WO_BDMaintenance'].includes(item.JournalTypeID)"
          ref="refDescription"
          v-model="item.WRDescription"
          label="BD Description"
          keepLabel
          multiRow="3"
          class="w-full"
        ></s-input> -->
      </template>
      <!-- end info WR/Unit -->

      <template #form_input_Dimension="{ item }">
        <dimension-editor
          :key="data.keyDimFinancial"
          ref="FinancialDimension"
          v-model="item.Dimension"
          sectionTitle="Financial Dimension"
          :default-list="profile.Dimension"
          :readOnly="data.formMode == 'view'"
          @change="
            (field, v1, v2) => {
              onChangeDimension(field, v1, v2, item);
            }
          "
        ></dimension-editor>
      </template>
      <template #form_input_InventDim="{ item }">
        <dimension-invent-jurnal
          ref="InventDimControl"
          :mandatory="['WarehouseID']"
          :defaultList="profile.Dimension"
          v-model="item.InventDim"
          :site="
            item.Dimension &&
            item.Dimension.find((_dim) => _dim.Key === 'Site') &&
            item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] != ''
              ? item.Dimension.find((_dim) => _dim.Key === 'Site')['Value']
              : undefined
          "
          :title-header="`Inventory Dimension`"
          :hide-field="[
            'BatchID',
            'SerialNumber',
            'SpecID',
            'InventDimID',
            'VariantID',
            'Size',
            'Grade',
          ]"
          :disabled="data.formMode == 'view'"
        ></dimension-invent-jurnal>
      </template>
      <template #form_footer_1="{ item }">
        <div>
          <div class="suim_form">
            <div class="mb-2 flex header">
              <div class="flex tab_container grow">
                <div
                  v-for="(tabTitle, tabIdx) in data.tabs"
                  @click="actionSubTabs(), (data.currentTab = tabIdx)"
                  :class="{
                    tab_selected: data.currentTab == tabIdx,
                    tab: data.currentTab != tabIdx,
                  }"
                >
                  {{ tabTitle }}
                </div>
              </div>
            </div>
          </div>
          <div class="list-plan">
            <div v-show="data.currentTab == 0">
              <work-order-plan-material
                ref="LinePlanMaterial"
                v-model="data.listPlanMaterial"
                :item="item"
                :hide-new-button="data.statusDisabled.includes(item.Status)"
                @OpentFromMaterial="OpentFromMaterial"
                @setPlan="setPlanMaterial"
              ></work-order-plan-material>
            </div>
            <div v-show="data.currentTab == 1">
              <work-order-plan-resource
                ref="LinePlanResource"
                v-model="data.listPlanResource"
                :item="item"
                :hide-new-button="data.statusDisabled.includes(item.Status)"
                :group-id-value="data.groupIdValue"
                @setPlan="setPlanResource"
              ></work-order-plan-resource>
            </div>
            <div v-show="data.currentTab == 2">
              <work-order-plan-output
                ref="LinePlanOutput"
                v-model="data.listPlanOutput"
                :item="item"
                :hide-new-button="data.statusDisabled.includes(item.Status)"
                @setPlan="setPlanOutput"
              ></work-order-plan-output>
            </div>
          </div>
        </div>
      </template>
      <template #form_footer_2="{ item }">
        <RejectionMessageList
          v-if="item.Status == 'REJECTED'"
          ref="listRejectionMessage"
          readUrl="/mfg/postingprofile/get-approval-by-source"
          :journalID="item._id"
          journalType="Work Order"
        ></RejectionMessageList>
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          class="bg-transparent hover:bg-blue-500 hover:text-black"
          label="Preview"
          icon="eye-outline"
          @click="onPreview"
        ></s-button>
        <s-button
          v-if="['DRAFT', ''].includes(item.Status)"
          :icon="`content-save`"
          class="btn_primary submit_btn"
          label="Save"
          :disabled="data.btnLoading"
          @click="onSave(item)"
        />
      </template>
      <template #form_buttons_2="{ item }">
        <div class="flex gap-1">
          <s-button
            v-if="
              data.record.Status == 'POSTED' &&
              data.record.StatusOverall == 'IN PROGRESS'
            "
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="END"
            @click="data.isDialogCloseWO = true"
          ></s-button>
          <form-buttons-trx
            :key="data.btnTrxId"
            :moduleid="`mfg`"
            :autoPost="false"
            :autoReopen="false"
            :status="item.Status"
            :journal-id="item._id"
            :posting-profile-id="item.PostingProfileID"
            :disabled="data.btnLoading"
            journal-type-id="Work Order"
            @preSubmit="trxPreSubmit"
            @pre-reopen="preReopen"
            @postSubmit="trxPostSubmit"
            @errorSubmit="trxErrorSubmit"
          />
        </div>
      </template>

      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
      <template #grid_item_button_delete="{ item }">
        <template v-if="!helper.isStatusDraft(item.Status)">&nbsp;</template>
      </template>
    </data-list>
    <dialog-from-material
      v-if="data.showDialogFrmMaterial"
      v-model="data.dialogFrmMaterial"
      @changeWork="onSaveMaterial"
      @close="data.showDialogFrmMaterial = false"
    />
    <dialog-from-daily-report
      v-if="data.showDialogFrmDailyReport"
      v-model="data.dialogFrmDailyReport"
      @changeWork="onOpenDailyReport"
      @close="data.showDialogFrmDailyReport = false"
    />
    <s-modal :display="data.isDialogCloseWO" ref="deleteModal">
      "Are you sure want to End this Work Order ?"
      <template #buttons="{ item }">
        <s-button
          class="w-[50px] btn_primary text-white font-bold flex justify-center"
          label="YES"
          @click="confirmEND"
        ></s-button>
        <s-button
          class="w-[50px] btn_warning text-white font-bold flex justify-center"
          label="NO"
          @click="data.isDialogCloseWO = false"
        ></s-button>
      </template>
    </s-modal>

    <PreviewReport
      v-if="data.isPreview"
      class="card w-full"
      title="Preview"
      :preview="data.preview"
      @close="closePreview"
      SourceType="Work Order"
      :SourceJournalID="data.record._id"
      :hideSignature="false"
    >
    </PreviewReport>
  </div>
</template>
<script setup>
import { reactive, ref, inject, watch, computed, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  loadGridConfig,
  loadFormConfig,
  DataList,
  util,
  SInput,
  SButton,
  SGrid,
  SForm,
  SModal,
} from "suimjs";
import { useRoute, useRouter } from "vue-router";
import { authStore } from "@/stores/auth";
import moment from "moment";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
// import DimensionInventory from "@/components/common/DimensionInventory.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import WorkOrderPlanMaterial from "./widget/workorder/WorkOrderPlanMaterial.vue";
import WorkOrderPlanResource from "./widget/workorder/WorkOrderPlanResource.vue";
import WorkOrderPlanOutput from "./widget/workorder/WorkOrderPlanOutput.vue";
import WorkOrderRptConsumption from "./widget/workorder/WorkOrderRptConsumption.vue";
import WorkOrderRptResource from "./widget/workorder/WorkOrderRptResource.vue";
import WorkOrderRptOutput from "./widget/workorder/WorkOrderRptOutput.vue";
import WorkOrderSum from "./widget/workorder/WorkOrderSum.vue";
import DialogFromMaterial from "./widget/workorder/DialogFromMaterial.vue";
import DialogFromDailyReport from "./widget/workorder/DialogFromDailyReport.vue";
import RejectionMessageList from "../scm/widget/RejectionMessageList.vue";
import Loader from "@/components/common/Loader.vue";
import StatusText from "@/components/common/StatusText.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import helper from "@/scripts/helper.js";

layoutStore().name = "tenant";
const headOffice = layoutStore().headOfficeID;
const featureID = "WorkOrder";
const profile = authStore().getRBAC(featureID);
const auth = authStore();
const route = useRoute();
const router = useRouter();
const defaultList = profile.Dimension.filter((v) => v.Key == "Site").map(
  (e) => e.Value
);

const listControl = ref(null);
const LinePlanMaterial = ref(null);
const LinePlanResource = ref(null);
const LinePlanOutput = ref(null);
const LineRptConsumption = ref(null);
const LineRptConsumptionPlan = ref(null);
const LineRptResource = ref(null);
const LineRptOutput = ref(null);
const LineSUMMaterial = ref(null);
const LineSUMResource = ref(null);
const LineSUMOutput = ref(null);
const gridDailyRptControl = ref(null);
const FinancialDimension = ref(null);
const InventDimControl = ref(null);
const refWOType = ref(null);
const refWoTypeKind = ref(null);
const gridAttachment = ref(null);
const refRequired = ref([]);
const axios = inject("axios");
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
        Value: [data.search.Text.trim()],
      },
      {
        Field: "WOName",
        Op: "$contains",
        Value: [data.search.Text.trim()],
      },
      {
        Field: "RequestorWOName",
        Op: "$contains",
        Value: [data.search.Text.trim()],
      },
      {
        Field: "WorkRequestID",
        Op: "$contains",
        Value: [data.search.Text.trim()],
      },
      {
        Field: "HullNo",
        Op: "$contains",
        Value: [data.search.Text.trim()],
      },
      {
        Field: "NoHullCustomer",
        Op: "$contains",
        Value: [data.search.Text.trim()],
      },
      {
        Field: "Asset",
        Op: "$contains",
        Value: [data.search.Text.trim()],
      }
    );
  }

  if (data.search.requestor !== null && data.search.requestor !== "") {
    filters.push({
      Field: "RequestorWOName",
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

  if (data.search.PolisiNo !== null && data.search.PolisiNo !== "") {
    filters.push({
      Field: "Asset",
      Op: "$eq",
      Value: data.search.PolisiNo,
    });
  }
  if (data.search.Status !== null && data.search.Status !== "") {
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: data.search.Status,
    });
  }

  if (data.search.WoType !== null && data.search.WoType !== "") {
    filters.push({
      Field: "JournalTypeID",
      Op: "$eq",
      Value: data.search.WoType,
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
    let ReffNo = data.record.WorkRequestID
      ? [
          `${data.record.WorkRequestID.slice(0, 2)}_${
            data.record.WorkRequestID
          }`,
        ]
      : [];
    const tags =
      ReffNo && data.record._id
        ? [...[`WO_${data.record._id}`], ...ReffNo]
        : data.record._id
        ? [`WO_${data.record._id}`]
        : ReffNo;

    return tags;
  },
});

const addTags = computed({
  get() {
    return [`WO_${data.record._id}`];
  },
});

const reffTags = computed({
  get() {
    let ReffNo = data.record.WorkRequestID
      ? [
          `${data.record.WorkRequestID.slice(0, 2)}_${
            data.record.WorkRequestID
          }`,
        ]
      : [];
    return ReffNo;
  },
});
const data = reactive({
  isPreview: false,
  appMode: "grid",
  formMode: "edit",
  titleForm: "Work Order",
  isOnFormAfter: true,
  isGridDailyReport: true,
  isLoadingBtnDaily: false,
  isDialogCloseWO: false,
  btnTrxId: util.uuid(),
  keyDept: util.uuid(),
  keyDetailAsset: util.uuid(),
  keyDimFinancial: util.uuid(),
  record: {
    _id: "",
  },
  formTabs: ["General"],
  tabs: ["Material", "Resource", "Output"],
  currentTab: 0,
  statusDisabled: [
    "WAITING",
    "READY",
    "POSTED",
    "SUBMITTED",
    "IN PROGRESS",
    "COMPLETED",
    "REJECTED",
  ],
  preview: "",
  warehouseID: "",
  gridCfgDailyReport: {},
  formCfgDailyReport: {},
  listPlanMaterial: [],
  listPlanResource: [],
  listPlanOutput: [],
  listDailyReport: [],
  listActivity: [],
  planMaterial: [],
  planResource: [],
  listRptResource: [],
  planOutput: [],
  showDialogFrmMaterial: false,
  showDialogFrmResource: false,
  showDialogFrmOutput: false,
  showDialogFrmDailyReport: false,
  isEdit: true,
  btnLoading: false,
  dialogFrmMaterial: {
    ItemID: "",
    SKU: "",
    Required: 0,
  },
  dialogFrmDailyReport: {
    WorkDate: new Date(),
  },
  InventDimKey: "",
  disabledWOtype: false,
  groupIdValue: ["EXG0008"],
  siteUser: "",
  search: {
    Text: "",
    requestor: "",
    DateFrom: null,
    DateTo: null,
    PolisiNo: "",
    Status: "",
    WoType: "",
    Site: "",
  },
  listTTD: [],
  requestorName: "",
  listWoTypeKind: [],
});

function getRequestor(record) {
  axios.post("/bagong/employee/get", [record.RequestorWOName]).then(
    (r) => {
      data.requestorName = r.data.Name;
    },
    (e) => util.showError(e)
  );
}

function newRecord(record) {
  data.disabledWOtype = false;
  data.formMode = "new";
  data.titleForm = `Create New Work Order`;
  data.currentTab = 0;
  data.listWoTypeKind = [];
  data.listPlanMaterial = [];
  data.listPlanResource = [];
  data.listPlanOutput = [];
  record._id = "";
  record.Status = "";
  record.CompanyID = auth.companyId;
  record.Kilometers = 0;
  record.WorkRequestID = "";
  record.JournalTypeID = "";
  record.TrxDate = new Date();
  record.WRDate = new Date();
  record.Schedule = new Date();
  record.ExpectedCompletedDate = moment().format("YYYY-MM-DDTHH:mm");
  record.StartDownTime = moment().format("YYYY-MM-DDTHH:mm");
  record.InventDim = {
    WarehouseID: "",
    AisleID: "",
    SectionID: "",
    BoxID: "",
  };
  getDetailEmployee("", record);
  util.nextTickN(2, () => {
    openForm(record);
  });
}

function editRecord(record) {
  data.record = record;
  data.listPlanMaterial = [];
  data.listPlanResource = [];
  data.listPlanOutput = [];
  record.StartDownTime = moment(
    moment(record.StartDownTime ? record.StartDownTime : new Date()).format(
      "YYYY-MM-DDTHH:mm:00Z"
    )
  ).format("YYYY-MM-DDTHH:mm");
  data.formMode = "edit";
  data.titleForm = `Edit Work Order | ${record._id}`;

  record.ExpectedCompletedDate = moment(
    moment(
      record.ExpectedCompletedDate ? record.ExpectedCompletedDate : new Date()
    ).format("YYYY-MM-DDTHH:mm:00Z")
  ).format("YYYY-MM-DDTHH:mm");

  openForm(record);
  getsReport(record);
  if (record.WorkRequestID) {
    getAsset(record.Asset, record);
  }

  data.disabledWOtype = false;
  if (record.WorkRequestType === "Production" && record.WorkRequestID !== "") {
    data.disabledWOtype = true;
    record.JournalTypeID = "WO_Production";
    disabledJournalType(record);
  }
  getApproval(record);
  getJournalType(record.JournalTypeID, record);
}
function openForm(record) {
  util.nextTickN(2, () => {
    // listControl.value.setFormFieldAttr("_id", "rules", roleID);
    if (data.statusDisabled.includes(record.Status)) {
      listControl.value.setFormMode("view");
      data.formMode = "view";
    }
    data.isGridDailyReport = true;
    data.record = record;
    data.btnTrxId = util.uuid();
    data.isEdit = false;
    listControl.value.setFormFieldAttr("_id", "readOnly", true);
    if (!record.WorkRequestID) {
      listControl.value.setFormFieldAttr("WorkRequestID", "hide", true);
      listControl.value.setFormFieldAttr("WRDate", "hide", false);
      listControl.value.setFormFieldAttr("RequestorName", "hide", true);
      listControl.value.setFormFieldAttr("RequestorDepartment", "hide", true);
      listControl.value.setFormFieldAttr("WRDescription", "hide", false);
      document.querySelector(
        "div:nth-child(2) > div > div.title.section_title"
      ).textContent = "Unit Information";
    } else {
      listControl.value.setFormFieldAttr("WorkRequestID", "hide", false);
      listControl.value.setFormFieldAttr("WRDate", "hide", false);
      listControl.value.setFormFieldAttr("RequestorName", "hide", false);
      listControl.value.setFormFieldAttr("RequestorDepartment", "hide", false);
      listControl.value.setFormFieldAttr("WRDescription", "hide", false);
      document.querySelector(
        "div:nth-child(2) > div > div.title.section_title"
      ).textContent = "Work Request Information";
    }

    if (
      [
        "WO_BDMaintenance",
        "WO_Production",
        "WO_PCR/OVHMaintenance",
        "WOGeneral",
      ].includes(record.JournalTypeID)
    ) {
      listControl.value.setFormFieldAttr("WoTypeKind", "hide", false);
      listControl.value.setFormFieldAttr("WoTypeKind", "required", true);
    } else {
      listControl.value.setFormFieldAttr("WoTypeKind", "hide", true);
      listControl.value.setFormFieldAttr("WoTypeKind", "required", false);
    }

    if (["WO_Production", "WO_BDMaintenance"].includes(record.JournalTypeID)) {
      listControl.value.setFormFieldAttr("WRDescription", "hide", false);
    } else {
      listControl.value.setFormFieldAttr("WRDescription", "hide", true);
    }

    if (
      ["WO_Production"].includes(record.JournalTypeID) &&
      record.WorkRequestID
    ) {
      listControl.value.setFormFieldAttr("Year", "hide", false);
      listControl.value.setFormFieldAttr("ACSystem", "hide", false);
    } else {
      listControl.value.setFormFieldAttr("Year", "hide", true);
      listControl.value.setFormFieldAttr("ACSystem", "hide", true);
    }

    if (data.statusDisabled.includes(record.Status)) {
      data.formTabs = [
        "General",
        "Report",
        "Material",
        "Resource",
        "Output",
        "Attachment",
      ];
      getWarehouseOnSite();
      actionTabs();
    } else {
      data.formTabs = record._id ? ["General", "Attachment"] : ["General"];
      util.nextTickN(6, () => {
        document.querySelector(
          ".form_inputs > div.flex.section_group_container > div:nth-child(1)"
        ).style.width = "37.5%";
        document.querySelector(
          ".form_inputs > div.flex.section_group_container > div:nth-child(2)"
        ).style.width = "37.5%";
        document.querySelector(
          ".form_inputs > div.flex.section_group_container > div:nth-child(3)"
        ).style.width = "25%";
      });
    }
    setMinDate(record);
  });
}
function setMinDate(record) {
  const el = document.querySelector('input[placeholder="WO Date"]');
  if (el) {
    el.setAttribute("min", moment(record.TrxDate).format("YYYY-MM-DD"));
  }
}
function actionTabs() {
  util.nextTickN(6, () => {
    let tabs = document.querySelector(".tab_container");
    tabs.addEventListener("click", function (event) {
      let tab = document.querySelector(
        ".tab_container > div.tab_selected"
      ).textContent;
      if (tab == "General") {
        document.querySelector(
          ".form_inputs > div.flex.section_group_container > div:nth-child(1)"
        ).style.width = "37.5%";
        document.querySelector(
          ".form_inputs > div.flex.section_group_container > div:nth-child(2)"
        ).style.width = "37.5%";
        document.querySelector(
          ".form_inputs > div.flex.section_group_container > div:nth-child(3)"
        ).style.width = "25%";
      }
    });
  });
}
function actionSubTabs() {
  document.querySelector(
    ".form_inputs > div.flex.section_group_container > div:nth-child(1)"
  ).style.width = "37.5%";
  document.querySelector(
    ".form_inputs > div.flex.section_group_container > div:nth-child(2)"
  ).style.width = "37.5%";
  document.querySelector(
    ".form_inputs > div.flex.section_group_container > div:nth-child(3)"
  ).style.width = "25%";
}

function preSaveAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [`WO_${data.record._id}`];
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
  let sortColm = [
    "ID",
    "RequestorWOName",
    "WOName",
    "Priority",
    "JournalTypeID",
    "BreakdownType",
    "TrxDate",
    "SafetyInstruction",
    "BOM",
    "WorkDescription",
    "WorkRequestID",
    "Asset",
    "WRDate",
    "HullNo",
    "RequestorName",
    "NoHullCustomer",
    "RequestorDepartment",
    "SunID",
    "StartDownTime",
    "CaroseryCode",
    "ExpectedCompletedDate",
    "Kilometers",
    "Dimension",
    "InventDim",
    "Status",
    "StatusOverall",
  ];

  cfg.setting.idField = "LastUpdate";
  cfg.setting.sortable = ["LastUpdate", "Created", "TrxDate", "Status", "_id"];
  cfg.setting.keywordFields = ["_id", "WOName", "WorkRequestID"];
  cfg.fields.map((f) => {
    f.idx = sortColm.indexOf(f.field);
    if (f.field == "Dimension") {
      f.label = "Site";
    } else if (f.field == "RequestorDepartment") {
      f.label = "Department";
    } else if (f.field == "Status") {
      f.idx = 100;
    } else if (f.field == "StatusOverall") {
      f.idx = 200;
    }
    return f;
  });

  cfg.fields.sort((a, b) => (a.idx > b.idx ? 1 : -1));
}

function onAlterFormConfig(cfg) {
  if (route.query.id !== undefined) {
    let currQuery = { ...route.query };
    listControl.value.selectData({ _id: currQuery.id });
    delete currQuery["id"];
    router.replace({ path: route.path, query: currQuery });
  }
}

function getCfgMaterial() {
  LinePlanMaterial.value.createGridCfgMaterial();
}
function getJournalType(_id, item) {
  listControl.value.setFormFieldAttr("Year", "hide", true);
  listControl.value.setFormFieldAttr("ACSystem", "hide", true);
  if (_id == "WO_Production" && item.WorkRequestID) {
    listControl.value.setFormFieldAttr("Year", "hide", false);
    listControl.value.setFormFieldAttr("ACSystem", "hide", false);
  }
  if (_id == "WO_Production") {
    listControl.value.setFormFieldAttr("SunID", "readOnly", false);
    if (!item.WorkRequestID && !item.Asset) {
      item.Asset = "";
      item.SunID = "";
      item.HullNo = "";
      item.CaroseryCode = "";
      item.Merk = "";
      item.UnitType = "";
      data.keyDetailAsset = util.uuid();
    }
  } else {
    listControl.value.setFormFieldAttr("SunID", "readOnly", true);
  }
  data.listWoTypeKind = [];
  if (
    ["WO_BDMaintenance", "WO_Production", "WO_PCR/OVHMaintenance"].includes(_id)
  ) {
    listControl.value.setFormFieldAttr("WoTypeKind", "hide", false);
    listControl.value.setFormFieldAttr("WoTypeKind", "required", true);
  } else {
    listControl.value.setFormFieldAttr("WoTypeKind", "hide", true);
    listControl.value.setFormFieldAttr("WoTypeKind", "required", false);
  }

  if (["WO_Production", "WO_BDMaintenance"].includes(_id)) {
    listControl.value.setFormFieldAttr("WRDescription", "hide", false);
  } else {
    listControl.value.setFormFieldAttr("WRDescription", "hide", true);
  }

  item.PostingProfileID = "";
  axios
    .post("/mfg/workorder/journal/type/find?_id=" + _id, { sort: ["-_id"] })
    .then(
      (r) => {
        if (r.data.length > 0) {
          data.btnTrxId = util.uuid();
          item.PostingProfileID = r.data[0].PostingProfileID;
        }
      },
      (e) => util.showError(e)
    );

  axios
    .post(
      `/tenant/masterdata/find?MasterDataTypeID=WOBreakdownTypeKind&ParentID=${_id}`
    )
    .then(
      (r) => {
        if (r.data.length > 0) {
          data.listWoTypeKind = r.data;
          item.WoTypeKind = r.data
            .map((v) => {
              return v._id;
            })
            .includes(item.WoTypeKind)
            ? item.WoTypeKind
            : r.data[0]._id;
        }
      },
      (e) => util.showError(e)
    );
}
function getAsset(_id, item) {
  item.SunID = "";
  item.HullNo = "";
  item.CaroseryCode = "";
  item.Merk = "";
  item.UnitType = "";
  if (_id) {
    axios.post("/bagong/asset/get", [_id]).then(
      (r) => {
        if (r.data.DetailUnit) {
          util.nextTickN(2, () => {
            item.SunID = r.data.DetailUnit.PurchaseCode;
            item.HullNo = r.data.DetailUnit.HullNum;
            item.CaroseryCode = r.data.DetailUnit.CaroseriCode;
            item.Merk = r.data.DetailUnit.Merk;
            item.UnitType = r.data.DetailUnit.UnitType;
            data.keyDetailAsset = util.uuid();
          });
        }
      },
      (e) => util.showError(e)
    );
    if (item.WorkRequestID) {
      axios.post("/tenant/asset/get", [_id]).then(
        (r) => {
          if (r.data.GroupID === "UNT") {
            // UNT CC terset Plant
            item.Dimension.map((dim) => {
              if (dim.Key == "CC") {
                dim.Value = "PLN";
              }
            });
          } else if (["PRT", "ELC"].includes(r.data.GroupID)) {
            //Property atau Office Equipment CC terset  HCGS
            item.Dimension.map((dim) => {
              if (dim.Key == "CC") {
                dim.Value = "HCG";
              }
            });
          }
          data.keyDimFinancial = util.uuid();
        },
        (e) => util.showError(e)
      );
    }
  } else {
    data.keyDetailAsset = util.uuid();
  }
}
function onChangeDimension(field, v1, v2, item) {
  switch (field) {
    case "Site":
      if (!item.WorkRequestID) {
        item.RequestorName = "";
        item.RequestorDepartment = "";
        item.Asset = "";
        item.SunID = "";
        item.HullNo = "";
        if (item.InventDim) {
          item.InventDim.WarehouseID = "";
        }

        util.nextTickN(2, () => {
          data.keyDept = util.uuid();
        });
      }

      break;
    case "PC":
      if (v1) {
        getWorkOrderExpTypeGroup(v1);
      }
      break;
    default:
      break;
  }
}
function disabledJournalType(record) {
  if (record.JournalTypeID) {
    axios
      .post("/mfg/workorder/journal/type/find?_id=" + record.JournalTypeID, {
        sort: ["-_id"],
      })
      .then(
        (r) => {
          if (r.data.length > 0) {
            record.PostingProfileID = r.data[0].PostingProfileID;
            data.btnTrxId = util.uuid();
          }
        },
        (e) => util.showError(e)
      );
  }
}
function getDetailEmployee(_id, record) {
  let payload = [];
  if (_id) {
    payload = [_id];
  }
  record.RequestorName = _id;
  axios.post("/tenant/employee/get-emp-warehouse", payload).then(
    (r) => {
      if (!_id) {
        record.RequestorName = r.data._id;
      }
    },
    (e) => util.showError(e)
  );
}
function onChangeRequestor(_id, record) {
  record.RequestorDepartment = "";
  if (typeof _id == "string") {
    axios.post("/bagong/employee/get", [_id]).then(
      (r) => {
        record.RequestorDepartment = r.data.Detail.Department;
        data.keyDept = util.uuid();
      },
      (e) => util.showError(e)
    );
  } else {
    data.keyDept = util.uuid();
  }
}
function postSave(record) {
  record.StartDownTime = moment(
    moment(record.StartDownTime ? record.StartDownTime : new Date()).format(
      "YYYY-MM-DDTHH:mm:00Z"
    )
  ).format("YYYY-MM-DDTHH:mm");

  record.ExpectedCompletedDate = moment(
    moment(
      record.ExpectedCompletedDate ? record.ExpectedCompletedDate : new Date()
    ).format("YYYY-MM-DDTHH:mm:00Z")
  ).format("YYYY-MM-DDTHH:mm");

  data.record = record;
}
function preSave(record) {
  record.Status = "DRAFT";
  record.StatusOverall = "DRAFT";
  record.StartDownTime = moment(record.StartDownTime).format(
    "YYYY-MM-DDTHH:mm:00Z"
  );

  record.ExpectedCompletedDate = moment(record.ExpectedCompletedDate).format(
    "YYYY-MM-DDTHH:mm:00Z"
  );
}
function onSave(record) {
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
  let validateWoTypeKind = true;

  if (FinancialDimension.value) {
    validate = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      validate = false;
    }
  }

  const LineMaterial = LinePlanMaterial.value.getDataValue();
  const LineResource = LinePlanResource.value.getDataValue();
  const LineOutpu = LinePlanOutput.value.getDataValue();

  LineMaterial.map((m, idx) => {
    m.LineNo = idx;
    if (record.JournalTypeID !== "WOGeneral") {
      m.Remarks = "";
    }
    return m;
  });
  LineResource.map((m, idx) => {
    m.LineNo = idx;
    return m;
  });
  LineOutpu.map((m, idx) => {
    m.LineNo = idx;
    return m;
  });

  record.WorkOrderSummaryMaterial = LineMaterial;
  record.WorkOrderSummaryResource = LineResource;
  record.WorkOrderSummaryOutput = LineOutpu;

  let payload = JSON.parse(JSON.stringify(record));
  payload.IsFirsttimeSave = false;
  if (!payload.StatusOverall) {
    payload.IsFirsttimeSave = true;
  }

  if (payload.JournalTypeID != "WOGeneral" && !payload.Asset) {
    return util.showError("field Asset is required");
  }

  if (!payload.WOName) {
    return util.showError("field WO Name is required");
  }
  if (validateWoTypeKind && validate && listControl.value.formValidate()) {
    data.btnLoading = true;
    listControl.value.setFormLoading(true);
    payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
    LinePlanMaterial.value.getAvailableStock(
      () => {
        listControl.value.submitForm(
          payload,
          () => {
            data.btnLoading = false;
            util.nextTickN(2, () => {
              if (gridAttachment.value) {
                gridAttachment.value.Save();
              }
              postSaveAttachment();
            });
            listControl.value.setFormLoading(false);
          },
          (e) => {
            data.btnLoading = false;
            listControl.value.setFormLoading(false);
          }
        );
      },
      () => {
        data.btnLoading = false;
        listControl.value.setFormLoading(false);
      }
    );
  } else {
    return util.showError("field is required");
  }
}

function trxPreSubmit(status, action, doSubmit) {
  let payload = JSON.parse(JSON.stringify(data.record));
  payload.IsFirsttimeSave = false;
  if (!payload.StatusOverall) {
    payload.IsFirsttimeSave = true;
  }

  payload.StartDownTime = moment(payload.StartDownTime).format(
    "YYYY-MM-DDTHH:mm:00Z"
  );

  payload.ExpectedCompletedDate = moment(payload.ExpectedCompletedDate).format(
    "YYYY-MM-DDTHH:mm:00Z"
  );

  const pc = payload.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = payload.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = payload.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;

  let validate = true;
  let validateWoTypeKind = true;

  if (FinancialDimension.value) {
    validate = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      validate = false;
    }
  }

  if (payload.JournalTypeID != "WOGeneral" && !payload.Asset) {
    return util.showError("field Asset is required");
  }

  if (!payload.WOName) {
    return util.showError("field WO Name is required");
  }

  if (validateWoTypeKind && validate && listControl.value.formValidate()) {
    if (status == "DRAFT" || status == "READY") {
      data.btnLoading = true;
      LinePlanMaterial.value.getAvailableStock(
        () => {
          payload.StartDownTime = moment(data.record.StartDownTime).format(
            "YYYY-MM-DDTHH:mm:00Z"
          );

          const LineMaterial = LinePlanMaterial.value.getDataValue();
          const LineResource = LinePlanResource.value.getDataValue();
          const LineOutpu = LinePlanOutput.value.getDataValue();

          LineMaterial.map((m, idx) => {
            m.LineNo = idx;
            return m;
          });
          LineResource.map((m, idx) => {
            m.LineNo = idx;
            return m;
          });
          LineOutpu.map((m, idx) => {
            m.LineNo = idx;
            return m;
          });

          payload.WorkOrderSummaryMaterial = LineMaterial;
          payload.WorkOrderSummaryResource = LineResource;
          payload.WorkOrderSummaryOutput = LineOutpu;
          payload.WorkOrderSummaryOutput.map((o) => {
            if (o.Type != "Waste Ledger") {
              o.InventoryLedgerAccID = o.InventoryLedgerAccID.split("~~").at(0);
            }
            return o;
          });

          if (status == "READY") {
            let material = payload.WorkOrderSummaryMaterial;
            for (let i in material) {
              if (material[i].AvailableStock > 0 && !material[i].InventDim) {
                data.btnLoading = false;
                return util.showError(
                  "AvailableStock greater than 0 Warehouse Location is required"
                );
              }

              if (material[i].AvailableStock == 0 && !material[i].InventDim) {
                material[i].InventDim = {
                  WarehouseID: data.warehouseID,
                };
              }
            }
          }
          payload.TrxDate = helper.dateTimeNow(payload.TrxDate);
          axios.post("/mfg/workorderplan/save", payload).then(
            (r) => {
              util.nextTickN(2, () => {
                if (gridAttachment.value) {
                  gridAttachment.value.Save();
                }
                postSaveAttachment();
              });
              doSubmit();
            },
            (e) => {
              data.btnLoading = false;
              util.showError(e);
            }
          );
        },
        () => {
          data.btnLoading = false;
        }
      );
    } else {
      doSubmit();
    }
  } else {
    return util.showError("field is required");
  }
}
function trxPostSubmit(record) {
  listControl.value.setControlMode("grid");
  listControl.value.refreshGrid();
  data.btnLoading = false;
  return util.showInfo("Work order has been successful submit");
}
function preReopen() {
  let payload = JSON.parse(JSON.stringify(data.record));
  payload.Status = "DRAFT";
  payload.StartDownTime = moment(data.record.StartDownTime).format(
    "YYYY-MM-DDTHH:mm:00Z"
  );
  payload.ExpectedCompletedDate = moment(
    data.record.ExpectedCompletedDate
  ).format("YYYY-MM-DDTHH:mm:00Z");
  payload.WorkOrderSummaryMaterial = LinePlanMaterial.value.getDataValue();
  payload.WorkOrderSummaryResource = LinePlanResource.value.getDataValue();
  payload.WorkOrderSummaryOutput = LinePlanOutput.value.getDataValue();
  payload.WorkOrderSummaryOutput.map((o) => {
    if (o.Type != "Waste Ledger") {
      o.InventoryLedgerAccID = o.InventoryLedgerAccID.split("~~").at(0);
    }
    return o;
  });

  data.btnLoading = true;
  axios.post("/mfg/workorderplan/save", payload).then(
    (r) => {
      util.nextTickN(2, () => {
        listControl.value.setControlMode("grid");
        listControl.value.refreshGrid();
        data.btnLoading = false;
      });
    },
    (e) => {
      data.btnLoading = false;
      return util.showError(e);
    }
  );
}
function trxErrorSubmit(record) {
  data.btnLoading = false;
}

function OpentFromMaterial() {
  data.showDialogFrmMaterial = true;
  data.dialogFrmMaterial = {
    ItemID: "",
    SKU: "",
    Required: 0,
  };
}
function onSaveMaterial(record) {
  const general = JSON.parse(JSON.stringify(data.record));
  const payload = { ...record, WorkOrderPlanID: general._id };
  if (general.Status == "DRAFT" || general.Status == "") {
    payload.ItemName = payload.Item.Name;
    payload.UnitID = payload.Item.DefaultUnitID;
    payload.SKUName = payload.ItemSpec?.SKU;
    LinePlanMaterial.value.addDataValue(payload);
    data.showDialogFrmMaterial = false;
    return util.showInfo("Plan Material has been successful add");
  } else {
    axios.post("/mfg/workorderplan/summary/material/insert", payload).then(
      (r) => {
        data.showDialogFrmMaterial = false;
        LinePlanMaterial.value.getsPlanMaterial();
        return util.showInfo("Plan Material has been successful save");
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}
function getsBOMMaterial(_id) {
  return axios.post(`/mfg/bom/material/gets?BoMID=${_id}`, {}).then(
    (r) => {
      return r.data.data;
    },
    (e) => {
      return [];
    }
  );
}
function getsBOMResource(_id) {
  return axios.post(`/mfg/bom/manpower/gets?BoMID=${_id}`, {}).then(
    (r) => {
      return r.data.data;
    },
    (e) => {
      return [];
    }
  );
}
function getsBOMOutput(_id) {
  return axios.post(`/mfg/bom/get`, [_id]).then(
    (r) => {
      return r.data;
    },
    (e) => {
      return {};
    }
  );
}
function onFormFieldChange(field, v1, v2, old, record) {
  switch (field) {
    case "BOM":
      data.listPlanMaterial = [];
      data.listPlanResource = [];
      data.listPlanOutput = [];
      if (typeof v1 == "string") {
        Promise.all([
          getsBOMMaterial(v1),
          getsBOMResource(v1),
          getsBOMOutput(v1),
        ]).then((res) => {
          const listPlanMaterial = res[0].map(function (m) {
            delete m._id;
            delete m.BoMID;
            m.UnitID = m.UoM;
            m.SKUName = m.Description.split("-").at(1);
            m.ItemName = m.Description.split("-").at(0);
            m.Required = m.Qty;
            return m;
          });
          const listPlanResource = res[1].map(function (r) {
            delete r._id;
            delete r.BoMID;
            r.TargetHour = r.StandartHour;
            return r;
          });
          const listPlanOutput = [res[2]].map(function (o) {
            if (o.OutputType == "Item") {
              o.Type = "WO Output";
              o.InventoryLedgerAccID = o.ItemID;
            } else {
              o.Type = "Waste Ledger";
              o.InventoryLedgerAccID = o.LedgerID;
            }
            delete o._id;
            o.QtyAmount = 0;
            o.AchievedQtyAmount = 0;
            o.Group = "";
            o.UnitID = "";
            return o;
          });
          LinePlanMaterial.value.setDataValue(listPlanMaterial);
          LinePlanResource.value.setDataValue(listPlanResource);
          LinePlanOutput.value.setDataValue(listPlanOutput);
        });
      } else {
        LinePlanMaterial.value.setDataValue([]);
        LinePlanResource.value.setDataValue([]);
        LinePlanOutput.value.setDataValue([]);
      }

      break;
    case "RequestorName":
      record.RequestorDepartment = "";
      axios.post("/bagong/employee/get", [v1]).then(
        (r) => {
          record.RequestorDepartment = r.data.Detail.Department;
          data.keyDept = util.uuid();
        },
        (e) => util.showError(e)
      );
      break;
    case "Asset":
      record.SunID = "";
      record.HullNo = "";
      if (!v1) {
        axios.post("/bagong/asset/get", [v1]).then(
          (r) => {
            if (r.data.DetailUnit) {
              record.SunID = r.data.DetailUnit.PurchaseCode;
              record.HullNo = r.data.DetailUnit.HullNum;
            }
          },
          (e) => util.showError(e)
        );
      }

      break;
    default:
      break;
  }
}
function newRecordDailyReport() {
  data.showDialogFrmDailyReport = true;
  data.dialogFrmDailyReport = {
    WorkDate: new Date(),
    Status: "",
  };
}

function onOpenDailyReport(record) {
  const rpt = data.listDailyReport.find((r) => {
    return (
      moment(r.WorkDate).format("DD-MMM-YYYY") ==
      moment(record.WorkDate).format("DD-MMM-YYYY")
    );
  });
  if (rpt) {
    return util.showError(`Please choose different date`);
  }
  const general = JSON.parse(JSON.stringify(data.record));
  const payload = {
    ...record,
    WorkOrderPlanID: general._id,
    PostingProfileID: general.PostingProfileID,
    WorkOrderPlanReportConsumptionID: "",
    WorkOrderPlanReportConsumptionStatus: "DRAFT",
    WorkOrderPlanReportOutputID: "",
    WorkOrderPlanReportOutputStatus: "DRAFT",
    WorkOrderPlanReportResourceID: "",
    WorkOrderPlanReportResourceStatus: "DRAFT",
    WorkOrderPlanReportConsumptionLines: [],
    WorkOrderPlanReportConsumptionAdditionalLines: [],
    WorkOrderPlanReportOutputLines: [],
    WorkOrderPlanReportResourceLines: [],
  };
  payload.Status = "";
  data.dialogFrmDailyReport = payload;
  data.showDialogFrmDailyReport = false;
  data.isGridDailyReport = false;
}
function getsReport(record) {
  axios
    .post(`/mfg/workorderplan/report/gets?WorkOrderPlanID=${record._id}`, {
      Sort: ["-WorkDate"],
    })
    .then(
      (r) => {
        util.nextTickN(2, () => {
          data.listDailyReport = r.data.data;
        });
      },
      (e) => {
        util.showError(e);
      }
    );
}
function getReport(record) {
  listControl.value.setFormLoading(true);
  axios.post(`/mfg/workorderplan/report/get`, [record._id]).then(
    (r) => {
      listControl.value.setFormLoading(false);
      util.nextTickN(2, () => {
        data.isGridDailyReport = false;
        data.dialogFrmDailyReport = r.data;
      });
    },
    (e) => {
      listControl.value.setFormLoading(false);
      util.showError(e);
    }
  );
}

function saveAsReport(ststus, cbOK, cbFalse) {
  listControl.value.setFormLoading(true);
  getAvailableStock(
    () => {
      data.isLoadingBtnDaily = true;
      const daily = JSON.parse(JSON.stringify(data.dialogFrmDailyReport));
      daily.WorkOrderPlanReportConsumptionAdditionalLines =
        LineRptConsumption.value.getDataValue();
      daily.WorkOrderPlanReportConsumptionLines =
        LineRptConsumptionPlan.value.getDataValue();
      daily.WorkOrderPlanReportResourceLines =
        LineRptResource.value.getDataValue();
      daily.WorkOrderPlanReportOutputLines = LineRptOutput.value.getDataValue();
      daily.Status = "DRAFT";
      if (ststus == "Submit") {
        daily.WorkOrderPlanReportConsumptionStatus = "SUBMITTED";
        daily.WorkOrderPlanReportOutputStatus = "SUBMITTED";
        daily.WorkOrderPlanReportResourceStatus = "SUBMITTED";
      }
      axios.post(`/mfg/workorderplan/report/save`, daily).then(
        (r) => {
          util.nextTickN(2, () => {
            if (ststus == "Submit") {
              submitReport(daily);
            } else {
              data.dialogFrmDailyReport = r.data;
              listControl.value.setFormLoading(false);
              data.isLoadingBtnDaily = false;
              if (cbOK) {
                cbOK();
              }

              return util.showInfo(
                "Work order report has been successful save"
              );
            }
          });
        },
        (e) => {
          listControl.value.setFormLoading(false);
          data.isLoadingBtnDaily = false;
          if (cbFalse) {
            cbFalse();
          }
          util.showError(e);
        }
      );
    },
    (e) => {
      listControl.value.setFormLoading(false);
      data.isLoadingBtnDaily = false;
      util.showError(e);
    }
  );
}
function submitReport() {
  const payload = JSON.parse(JSON.stringify(data.record));
  const daily = JSON.parse(JSON.stringify(data.dialogFrmDailyReport));
  axios
    .post(`/mfg/workorderplan/report/submit`, {
      ID: daily._id,
    })
    .then(
      (r) => {
        util.nextTickN(2, () => {
          data.isGridDailyReport = true;
          data.isLoadingBtnDaily = false;
          listControl.value.setFormLoading(false);
          data.dialogFrmDailyReport = r.data;
          getsReport(payload);
          return util.showInfo("Item Request has been sent");
        });
      },
      (e) => {
        data.isLoadingBtnDaily = false;
        listControl.value.setFormLoading(false);
        util.showError(e);
      }
    );
}
function confirmEND() {
  const payload = JSON.parse(JSON.stringify(data.record));
  axios.post(`/mfg/workorderplan/end`, payload._id).then(
    (r) => {
      util.nextTickN(2, () => {
        data.isDialogCloseWO = false;
        listControl.value.cancelForm();
        return util.showInfo("Work order report has been successful end");
      });
    },
    (e) => {
      util.showError(e);
    }
  );
}
function backToReport() {
  const payload = JSON.parse(JSON.stringify(data.record));
  data.isGridDailyReport = true;
  getsReport(payload);
}
function getStock(Items) {
  return axios
    .post(`/mfg/workorderplan/gets-available-stock`, {
      InventDim: data.record.InventDim,
      Items: Items,
      BalanceFilter: { WarehouseIDs: [data.record.InventDim.WarehouseID] },
    })
    .then(
      (r) => {
        return r.data;
      },
      (e) => {
        util.showError(e);
        return [];
      }
    );
}
function getsPlanMaterial() {
  return axios
    .post(
      `/mfg/workorderplan/summary/material/gets?WorkOrderPlanID=${data.record._id}`,
      {}
    )
    .then(
      (r) => {
        r.data.data.map((c) => {
          c.ItemVarian = helper.ItemVarian(c.ItemID, c.SKU);
          return c;
        });
        return r.data.data;
      },
      (e) => {
        return [];
      }
    );
}
function getAvailableStock(cbOK, cbFalse) {
  const additional = LineRptConsumption.value.getDataValue();
  const plan = LineRptConsumptionPlan.value.getDataValue();
  LineRptConsumption.value.getDataValue();
  Promise.all([getStock(additional), getsPlanMaterial()]).then(
    (res) => {
      const itemAdditional = res[0];
      const itemPlan = res[1];
      additional.map((r, i) => {
        const totalItem = itemAdditional
          .filter(function (v) {
            return v.ItemID == r.ItemID && v.SKU == r.SKU;
          })
          .reduce((accumulator, object) => {
            return accumulator + object.Qty;
          }, 0);

        r.QtyAvailable = totalItem;
        return r;
      });
      plan.map((r, i) => {
        const rptItem = itemPlan.find(function (v) {
          return v.ItemID == r.ItemID && v.SKU == r.SKU;
        });
        let QtyAvailable = 0;
        if (rptItem) {
          QtyAvailable = rptItem.Required - (rptItem.Used ? rptItem.Used : 0);
        }
        r.QtyAvailable = QtyAvailable;
        return r;
      });
      LineRptConsumption.value.setRecords(additional);
      LineRptConsumptionPlan.value.setRecords(plan);
      if (cbOK) {
        cbOK();
      }
    },
    (e) => {
      if (cbFalse) {
        cbFalse(emit);
      }
      util.showError(e);
    }
  );
}

function createItemRequest() {
  const additional = LineRptConsumption.value.getDataValue();
  const plan = LineRptConsumptionPlan.value.getDataValue();
  if (additional.length == 0 || !data.dialogFrmDailyReport._id) {
    util.showError("additional item no lines or please save");
    return;
  }
  saveAsReport(
    "Save",
    () => {
      axios
        .post(`/mfg/workorderplan/report/create-additional-item-request`, {
          WorkOrderPlanReportID: data.dialogFrmDailyReport._id,
        })
        .then(
          (r) => {
            util.showInfo("additional item request has been successful create");
          },
          (e) => {
            util.showError(e);
          }
        );
    },
    null
  );
}
function preSubmitRptConsumption(status, action, doSubmit) {
  listControl.value.setFormLoading(true);
  const daily = JSON.parse(JSON.stringify(data.dialogFrmDailyReport));
  if (
    status == "READY" ||
    daily.WorkOrderPlanReportConsumptionStatus == "DRAFT"
  ) {
    if (daily.WorkOrderPlanReportConsumptionStatus == "READY") {
      let material = LineRptConsumption.value.getDataValue();
      for (let i in material) {
        if (material[i].QtyAvailable > 0 && !material[i].WarehouseLocation) {
          data.btnLoading = false;
          listControl.value.setFormLoading(false);
          return util.showError(
            "Qty Available greater than 0 Warehouse Location is required"
          );
        }

        if (material[i].QtyAvailable == 0 && !material[i].WarehouseLocation) {
          material[i].InventDim = {
            WarehouseID: data.warehouseID,
          };
        }
      }
    }

    daily.WorkOrderPlanReportConsumptionAdditionalLines =
      LineRptConsumption.value.getDataValue();
    daily.WorkOrderPlanReportConsumptionLines =
      LineRptConsumptionPlan.value.getDataValue();
    daily.WorkOrderPlanReportResourceLines =
      LineRptResource.value.getDataValue();
    daily.WorkOrderPlanReportOutputLines = LineRptOutput.value.getDataValue();

    getAvailableStock(
      () => {
        axios.post(`/mfg/workorderplan/report/save`, daily).then(
          (r) => {
            util.nextTickN(2, () => {
              doSubmit();
            });
          },
          (e) => {
            listControl.value.setFormLoading(false);
            util.showError(e);
          }
        );
      },
      () => {
        listControl.value.setFormLoading(false);
        util.showError(e);
      }
    );
  } else {
    doSubmit();
  }
}

function preReopenRptConsumption(status, action, doSubmit) {
  listControl.value.setFormLoading(true);
  const daily = JSON.parse(JSON.stringify(data.dialogFrmDailyReport));
  daily.WorkOrderPlanReportConsumptionAdditionalLines =
    LineRptConsumption.value.getDataValue();
  daily.WorkOrderPlanReportConsumptionLines =
    LineRptConsumptionPlan.value.getDataValue();
  daily.WorkOrderPlanReportResourceLines = LineRptResource.value.getDataValue();
  daily.WorkOrderPlanReportOutputLines = LineRptOutput.value.getDataValue();
  daily.WorkOrderPlanReportConsumptionStatus = "DRAFT";
  axios.post(`/mfg/workorderplan/report/save`, daily).then(
    (r) => {
      listControl.value.setFormLoading(false);
      backToReport();
    },
    (e) => {
      listControl.value.setFormLoading(false);
      util.showError(e);
    }
  );
}
function postSubmitRptConsumption() {
  data.isGridDailyReport = true;
  listControl.value.setFormLoading(false);
  const payload = JSON.parse(JSON.stringify(data.record));
  getsReport(payload);
  LinePlanMaterial.value.getsPlanMaterial();
  LineSUMMaterial.value.getsSummary();
}
function errorSubmitRptConsumption() {
  listControl.value.setFormLoading(false);
}

function preSubmitRptResource(status, action, doSubmit) {
  if (status == "DRAFT") {
    listControl.value.setFormLoading(true);
    data.isLoadingBtnDaily = true;
    const daily = JSON.parse(JSON.stringify(data.dialogFrmDailyReport));
    daily.WorkOrderPlanReportConsumptionAdditionalLines =
      LineRptConsumption.value.getDataValue();
    daily.WorkOrderPlanReportConsumptionLines =
      LineRptConsumptionPlan.value.getDataValue();
    daily.WorkOrderPlanReportResourceLines =
      LineRptResource.value.getDataValue();
    daily.WorkOrderPlanReportOutputLines = LineRptOutput.value.getDataValue();
    daily.WorkOrderPlanReportResourceStatus = "SUBMITTED";

    axios.post(`/mfg/workorderplan/report/save`, daily).then(
      (r) => {
        data.isLoadingBtnDaily = false;
        util.nextTickN(2, () => {
          doSubmit();
        });
      },
      (e) => {
        data.isLoadingBtnDaily = false;
        listControl.value.setFormLoading(false);
        util.showError(e);
      }
    );
  }
}
function postSubmitRptResource() {
  data.isGridDailyReport = true;
  listControl.value.setFormLoading(false);
  const payload = JSON.parse(JSON.stringify(data.record));
  getsReport(payload);
  LinePlanResource.value.getsPlanResource();
  LineSUMResource.value.getsSummary();
}
function errorSubmitRptResource() {
  listControl.value.setFormLoading(false);
}

function preReopenRptResource(status, action, doSubmit) {
  listControl.value.setFormLoading(true);
  const daily = JSON.parse(JSON.stringify(data.dialogFrmDailyReport));
  daily.WorkOrderPlanReportConsumptionAdditionalLines =
    LineRptConsumption.value.getDataValue();
  daily.WorkOrderPlanReportConsumptionLines =
    LineRptConsumptionPlan.value.getDataValue();
  daily.WorkOrderPlanReportResourceLines = LineRptResource.value.getDataValue();
  daily.WorkOrderPlanReportOutputLines = LineRptOutput.value.getDataValue();
  daily.WorkOrderPlanReportResourceStatus = "DRAFT";
  axios.post(`/mfg/workorderplan/report/save`, daily).then(
    (r) => {
      listControl.value.setFormLoading(false);
      backToReport();
    },
    (e) => {
      listControl.value.setFormLoading(false);
      util.showError(e);
    }
  );
}
function preSubmitRptOutput(status, action, doSubmit) {
  if (status == "DRAFT") {
    listControl.value.setFormLoading(true);
    data.isLoadingBtnDaily = true;
    const daily = JSON.parse(JSON.stringify(data.dialogFrmDailyReport));
    daily.WorkOrderPlanReportConsumptionAdditionalLines =
      LineRptConsumption.value.getDataValue();
    daily.WorkOrderPlanReportConsumptionLines =
      LineRptConsumptionPlan.value.getDataValue();
    daily.WorkOrderPlanReportResourceLines =
      LineRptResource.value.getDataValue();
    daily.WorkOrderPlanReportOutputLines = LineRptOutput.value.getDataValue();
    daily.WorkOrderPlanReportOutputStatus = "SUBMITTED";

    axios.post(`/mfg/workorderplan/report/save`, daily).then(
      (r) => {
        data.isLoadingBtnDaily = false;
        util.nextTickN(2, () => {
          doSubmit();
        });
      },
      (e) => {
        data.isLoadingBtnDaily = false;
        listControl.value.setFormLoading(false);
        util.showError(e);
      }
    );
  }
}
function postSubmitRptOutput() {
  data.isGridDailyReport = true;
  listControl.value.setFormLoading(false);
  const payload = JSON.parse(JSON.stringify(data.record));
  getsReport(payload);
  LinePlanOutput.value.getsPlanOutput();
  LineSUMOutput.value.getsSummary();
}
function errorSubmitRptOutput() {
  listControl.value.setFormLoading(false);
}

function preReopenRptOutput(status, action, doSubmit) {
  listControl.value.setFormLoading(true);
  const daily = JSON.parse(JSON.stringify(data.dialogFrmDailyReport));
  daily.WorkOrderPlanReportConsumptionAdditionalLines =
    LineRptConsumption.value.getDataValue();
  daily.WorkOrderPlanReportConsumptionLines =
    LineRptConsumptionPlan.value.getDataValue();
  daily.WorkOrderPlanReportResourceLines = LineRptResource.value.getDataValue();
  daily.WorkOrderPlanReportOutputLines = LineRptOutput.value.getDataValue();
  daily.WorkOrderPlanReportOutputStatus = "DRAFT";
  axios.post(`/mfg/workorderplan/report/save`, daily).then(
    (r) => {
      listControl.value.setFormLoading(false);
      backToReport();
    },
    (e) => {
      listControl.value.setFormLoading(false);
      util.showError(e);
    }
  );
}
function onControlModeChanged(mode) {
  data.appMode = mode;
  if (mode === "grid") {
    data.titleForm = "Work Order";
    data.isEdit = true;
  }
}
function lookupPayloadBuilder(search, config, value, item, isSite = false) {
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
          { Field: "Name", Op: "$contains", Value: [search] },
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
function lookupPayloadBuilderWOType(search, config, value, item) {
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

  const query = [
    {
      Field: "_id",
      Op: "$ne",
      Value: "WO_BDMaintenance",
    },
  ];

  if (item.WorkRequestID) {
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

    if (Site) {
      items = [...items, ...query];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }

  return qp;
}
function lookupDailyStatus(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

  //setting search
  const query = [
    {
      Field: "ParentID",
      Op: "$eq",
    },
  ];
  qp.Where = {
    Op: "$and",
    items: query,
  };
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
    items = [...items, ...query];
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}
function lookupMonitoringStatus(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;
  //setting search
  let query = [
    {
      Field: "ParentID",
      Op: "$ne",
    },
  ];
  if (item.DailyStatus) {
    query = [
      {
        Field: "ParentID",
        Op: "$eq",
        value: item.DailyStatus,
      },
    ];
  }
  qp.Where = {
    Op: "$and",
    items: query,
  };
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
    items = [...items, ...query];
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}

function getWorkOrderExpTypeGroup(pcValue) {
  if (pcValue && pcValue != "") {
    let url = `/tenant/expensetypegroup/gets?Dimension.Key=PC&Dimension.Value=${pcValue}`;
    axios.post(url, { Sort: ["-_id"] }).then((r) => {
      const records = r.data.data;
      const find = records.find((v) => v.Name.includes("WORK ORDER"));
      if (find) {
        data.groupIdValue = [data.groupIdValue[0], find._id];
      }
    });
  } else {
    data.groupIdValue = [data.groupIdValue[0]];
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
    },
    (e) => util.showError(e)
  );
}
function setPlanMaterial(record) {
  data.planMaterial = record;
}
function setPlanResource(record) {
  const resource = record.map((t) => {
    return t.ExpenseType;
  });
  data.listRptResource = record;
  data.planResource = resource;
}
function setPlanOutput(record) {
  data.planOutput = record;
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
function getWarehouseOnSite(siteId, cbOK, cbFalse) {
  let Site =
    data.record.Dimension &&
    data.record.Dimension.find((_dim) => _dim.Key === "Site") &&
    data.record.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? data.record.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : undefined;

  if (!siteId) {
    siteId = Site;
  }
  axios
    .post(`/tenant/warehouse/find`, {
      Where: {
        Op: "$and",
        Items: [
          {
            Field: "Dimension.Key",
            Op: "$eq",
            Value: "Site",
          },
          {
            Field: "Dimension.Value",
            Op: "$eq",
            Value: siteId,
          },
        ],
      },
    })
    .then(
      (r) => {
        if (r.data.length > 0) {
          data.warehouseID = r.data[0]._id;
        }
        if (cbOK) {
          cbOK(r.data);
        }
      },
      (e) => {
        util.showError(e);
        if (cbFalse) {
          cbFalse();
        }
      }
    );
}
onMounted(() => {
  getByCurrentUser();
  loadGridConfig(axios, "/mfg/workorderplan/report/gridconfig").then(
    (r) => {
      data.gridCfgDailyReport = r;
    },
    (e) => util.showError(e)
  );
  loadFormConfig(axios, "/mfg/workorderplan/report/formconfig").then(
    (r) => {
      data.formCfgDailyReport = r;
    },
    (e) => util.showError(e)
  );
});

function onPreview() {
  getRequestor(data.record);
  data.isPreview = true;
}

function closePreview() {
  data.isPreview = false;
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
</script>
<style scoped>
.material-consumption {
  display: grid;
  grid-row-gap: 20px;
}
.status-overall {
  width: 49%;
}
</style>
<style scoped>
.radio-status input {
  display: none;
}
.radio-status label {
  position: relative;
  cursor: pointer;
  color: #666;
  font-weight: 400;
  font-size: 14px;
}
.radio-status label:before {
  content: " ";
  display: inline-block;
  position: relative;
  top: 5px;
  margin: 0 5px 0 0;
  width: 20px;
  height: 20px;
  border-radius: 11px;
  border: 2px solid;
  background-color: inherit;
}

.radio-status input[type="radio"]:checked + label:after {
  border-radius: 11px;
  position: absolute;
  width: 12px;
  height: 12px;
  top: 4px;
  left: 4px;
  content: " ";
  display: block;
  background-color: #fd6e76;
}
</style>
