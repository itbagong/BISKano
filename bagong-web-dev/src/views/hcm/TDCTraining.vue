<template>
  <div class="w-full">
    <data-list
      v-show="data.view == 'tdc'"
      class="card"
      ref="listControl"
      title="TDC Training"
      grid-config="/hcm/tdc/gridconfig"
      form-config="/hcm/tdc/formconfig"
      grid-read="/hcm/tdc/gets"
      form-read="/hcm/tdc/get"
      grid-mode="grid"
      grid-delete="/hcm/tdc/delete"
      grid-sort-field="RequestDate"
      form-keep-label
      form-insert="/hcm/tdc/save"
      form-update="/hcm/tdc/save"
      :grid-fields="[
        'TrainingTitle',
        'Status',
        'StatusDetail',
        'TrainingStatus',
        'TrainingRequestor',
        'AssessmentType',
      ]"
      :form-tabs-new="['General']"
      :form-tabs-edit="data.tabsEdit"
      :form-tabs-view="data.tabsEdit"
      :form-fields="[
        'TrainingTitle',
        'TrainingRequestor',
        'Dimension',
        'JournalTypeID',
      ]"
      :init-app-mode="data.appMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      @post-save="onPostSave"
      :formHideSubmit="readOnly"
      @form-field-change="onFormFieldChange"
      @alterGridConfig="alterGridConfig"
      :grid-custom-filter="data.customFilter"
      stay-on-form-after-save
    >
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilter"
          v-model="data.customFilter"
          hideAll
          customTextLabel="Search"
          :fieldsText="['_id', 'TrainingTitle']"
          @initNewItem="initNewItemFilter"
          @preChange="changeFilter"
          @change="refreshGrid"
        >
          <template #filter_1="{ item }">
            <s-input
              class="w-full filter-text"
              label="Search"
              v-model="item.Text"
            />
          </template>
        </grid-header-filter>
      </template>
      <!-- form slot -->
      <template #form_buttons_1="{ item, inSubmission, loading, mode }">
        <s-button
          v-show="!item.TrainingStatus"
          class="btn_primary"
          label="Close"
          icon="Close"
          @click="onClosingTraining(item)"
        ></s-button>
        <s-button
          v-show="showSaveTrainingDetail"
          class="btn_primary"
          label="Save"
          icon="content-save"
          @click="saveDetail"
        ></s-button>
        <s-button
          v-show="showSaveAttendance"
          class="btn_primary"
          label="Save"
          icon="content-save"
          @click="saveAttendance"
        ></s-button>
        <s-button
          v-show="showSaveMaterial"
          class="btn_primary"
          label="Save"
          icon="content-save"
          @click="saveMaterial"
        ></s-button>
        <s-button
          v-show="mode !== 'new' && showPreview !== ''"
          :disabled="inSubmission || loading"
          class="bg-primary text-white font-bold w-full flex justify-center"
          label="Preview"
          @click="
            () => {
              if (showPreview == 'previewGeneral') {
                data.view = 'preview';
              } else if (showPreview == 'previewDetail') {
                data.view = 'previewDetail';
              }
            }
          "
        ></s-button>
        <s-button
          v-show="mode !== 'new'"
          :disabled="inSubmission || loading"
          class="bg-primary text-white font-bold w-full flex justify-center"
          label="Download"
          @click="downloadTraining"
        ></s-button>
        <!-- <s-button
          v-show="showSaveAssesment"
          class="btn_primary"
          label="Save"
          icon="content-save"
          @click="saveAssesment"
        ></s-button> -->
        <form-buttons-trx
          :disabled="inSubmission || loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="data.jType"
          :auto-post="!waitTrxSubmit"
          moduleid="hcm"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
        />
        <form-buttons-trx
          :key="'trainingdetail_' + data.trainingDetail._id"
          :disabled="inSubmission || loading"
          :status="data.trainingDetail.Status"
          :journal-id="data.trainingDetail._id"
          :posting-profile-id="data.trainingDetail.PostingProfileID"
          :journal-type-id="data.jTypeTraningDetail"
          :auto-post="!waitTrxSubmitDetail"
          moduleid="hcm"
          @preSubmit="trxPreSubmitDetail"
          @postSubmit="trxPostSubmitDetail"
          @errorSubmit="trxErrorSubmitDetail"
        />
      </template>
      <template #form_input_TrainingTitle="{ item }">
        <s-input
          ref="refTrainingTitle"
          label="Training Title"
          v-model="item.TrainingTitle"
          class="w-50"
          use-list
          :lookup-url="`/hcm/tdctitle/find`"
          lookup-key="_id"
          :lookup-labels="['Alias', 'Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #form_input_TrainingRequestor="{ item }">
        <s-input
          ref="refTrainingRequestor"
          label="Training Requestor"
          v-model="item.TrainingRequestor"
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          :read-only="readOnly || mode == 'view'"
          class="w-100"
          @change="
            (field, v1, v2, old, ctlRef) => {
              onGridRowFieldChanged('TrainingRequestor', v1, v2, old, item);
            }
          "
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :read-only="readOnly || mode == 'view'"
        ></dimension-editor-vertical>
      </template>
      <template #form_input_JournalTypeID="{ item, mode }">
        <s-input
          class="w-full"
          required
          label="Journal type ID"
          use-list
          lookup-url="/hcm/tdcjournaltype/find?TransactionType=Training%20-%20General%20%26%20Participant"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.JournalTypeID"
          :read-only="readOnly || mode == 'view'"
          @change="
            (_, v) => {
              getJurnalType(v, item);
            }
          "
        />
      </template>

      <!-- form tab -->
      <template #form_tab_Participant="{ item }">
        <data-list
          ref="gridParticipant"
          hide-title
          no-gap
          :grid-editor="!readOnly"
          grid-hide-search
          grid-hide-sort
          grid-hide-refresh
          grid-hide-detail
          grid-hide-select
          grid-hide-paging
          :grid-hide-new="readOnly"
          :grid-hide-delete="readOnly"
          grid-no-confirm-delete
          :init-app-mode="data.appMode"
          :grid-mode="data.appMode"
          new-record-type="grid"
          grid-config="/hcm/tdcparticipant/gridconfig"
          :grid-read="`/hcm/tdcparticipant/gets?TrainingCenterID=${item._id}`"
          grid-auto-commit-line
          @alter-grid-config="alterGridParticipantConfig"
          @grid-row-add="newParticipantRecord"
          @grid-row-delete="onGridRowDeleteParticipant"
          @grid-row-field-changed="onGridRowFieldChangedParticipant"
          :grid-fields="['EmployeeID', 'Site', 'Department', 'NIK']"
        >
          <template #grid_EmployeeID="{ item }">
            <s-input
              ref="refEmployeeID"
              hide-label
              v-model="item.EmployeeID"
              :read-only="readOnly"
              use-list
              :lookup-url="`/tenant/employee/find`"
              lookup-key="_id"
              :lookup-labels="['_id', 'Name']"
              :lookup-searchs="['_id', 'Name']"
              class="w-100"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  onGridRowFieldChangedParticipant(
                    'EmployeeID',
                    v1,
                    v2,
                    old,
                    item
                  );
                }
              "
            ></s-input>
          </template>
          <template #grid_NIK="{ item }">
            {{ item.NIK }}
          </template>
          <template #grid_Site="{ item }">
            <s-input
              ref="refSite"
              use-list
              hide-label
              lookup-url="/tenant/dimension/find?DimensionType=Site"
              lookup-key="_id"
              :lookup-labels="['Label']"
              :lookup-searchs="['_id', 'Label']"
              v-model="item.Site"
              class="min-w-[180px]"
              disabled
              multiple
              @change="
                (field, v1, v2, old, ctlRef) => {
                  onGridRowFieldChangedParticipant('Site', v1, v2, old, item);
                }
              "
            />
          </template>
          <template #grid_Department="{ item }">
            <s-input
              ref="refDepartment"
              use-list
              hide-label
              lookup-url="/tenant/dimension/find?DimensionType=CC"
              lookup-key="_id"
              :lookup-labels="['Label']"
              :lookup-searchs="['_id', 'Label']"
              v-model="item.Department"
              class="min-w-[180px]"
              disabled
              @change="
                (field, v1, v2, old, ctlRef) => {
                  onGridRowFieldChangedParticipant(
                    'Department',
                    v1,
                    v2,
                    old,
                    item
                  );
                }
              "
            />
          </template>

          <template #grid_header_buttons_1="{ config }">
            <s-button
              v-if="!readOnly"
              class="bg-primary text-white font-bold w-full flex justify-center"
              label="Import"
              @click="triggerFileInputParticipant"
            ></s-button>
            <input
              type="file"
              ref="fileInputParticipant"
              @change="handleFileUploadParticipant"
              hidden
            />
            <s-button
              v-if="!readOnly"
              class="bg-primary text-white font-bold w-full flex justify-center"
              label="Action"
              @click="() => (data.modal = true)"
            ></s-button>
          </template>
        </data-list>
      </template>
      <template #form_tab_Training_Detail="{}">
        <s-form
          ref="formTrainingDetail"
          v-model="data.trainingDetail"
          :config="data.trainingDetailFormCfg"
          :mode="
            ['', 'DRAFT'].includes(data.trainingDetail.Status) ? 'edit' : 'view'
          "
          keep-label
          only-icon-top
          hide-submit
          hide-cancel
        >
          <template #input_TrainerType="{ item, mode }">
            <s-input
              v-if="item.TrainerType == 'Internal'"
              ref="refTrainerType"
              label="Trainer Type"
              v-model="item.TrainerType"
              use-list
              :read-only="readOnlyDetail || mode == 'view'"
              :items="['Internal', 'External']"
              class="w-100"
              @change="
                (field, v1, v2, old, ctlRef) => {
                  item.TrainerName = '';
                }
              "
            ></s-input>
          </template>
          <template #input_JournalTypeID="{ item, mode }">
            <s-input
              :key="item.JournalTypeID"
              class="w-full"
              required
              label="Journal type ID"
              :read-only="readOnlyDetail || mode == 'view'"
              use-list
              lookup-url="/hcm/tdcjournaltype/find?TransactionType=Training%20-%20Training%20Detail%20%26%20Attachment"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookupSearchs="['Name']"
              v-model="item.JournalTypeID"
              @change="
                (_, v) => {
                  getJurnalTypeDetail(v, item);
                }
              "
            />
          </template>
          <template #input_TrainerName="{ item, mode }">
            <s-input
              v-if="item.TrainerType == 'Internal'"
              :read-only="readOnlyDetail || mode == 'view'"
              ref="refTrainerName"
              label="Trainer Name"
              v-model="item.TrainerName"
              use-list
              :lookup-url="`/tenant/employee/find`"
              lookup-key="_id"
              :lookup-labels="['_id', 'Name']"
              :lookup-searchs="['_id', 'Name']"
              class="w-100"
            ></s-input>
            <s-input
              v-else
              :read-only="readOnlyDetail || mode == 'view'"
              ref="refTrainerName"
              label="Trainer Name"
              v-model="item.TrainerName"
              class="w-100"
            ></s-input>
          </template>
        </s-form>
      </template>
      <template #form_tab_Attendance="{ item }">
        <!-- grid -->
        <data-list
          v-show="data.viewAttendance == 'grid'"
          ref="gridAttendance"
          hide-title
          no-gap
          grid-hide-search
          grid-hide-sort
          grid-hide-refresh
          grid-hide-select
          grid-hide-paging
          :init-app-mode="data.appMode"
          :grid-mode="data.appMode"
          new-record-type="grid"
          grid-config="/hcm/tdcattendance/gridconfig"
          :grid-read="`/hcm/tdcattendance/gets?TrainingCenterID=${item._id}`"
          grid-delete="/hcm/tdcattendance/delete"
          grid-auto-commit-line
          @grid-row-add="addAttendance"
          @form-edit-data="editAttendance"
          :grid-fields="['Date']"
        >
          <template #grid_Date="{ item }">
            {{ moment(item.Date).format("DD MM YYYY").toString() }}
          </template>
        </data-list>
        <!-- form -->
        <div v-show="data.viewAttendance == 'form'">
          <div class="flex justify-end gap-2">
            <s-button
              icon="content-save"
              class="btn_primary"
              label="Save"
              @click="saveAttendance"
            />
            <s-button
              class="btn_warning back_btn"
              label="Back"
              icon="rewind"
              @click="backAttendance"
            />
          </div>
          <s-form
            ref="formAttendance"
            form-config="/hcm/tdc/formconfig"
            v-model="data.attendance"
            :config="data.attendanceFormCfg"
            keep-label
            only-icon-top
            hide-submit
            hide-cancel
          >
            <template #input_LocationID="{ item, mode }">
              <s-input
                ref="refActivityName"
                label="Location"
                v-model="item.LocationID"
                use-list
                lookup-url="/tenant/masterdata/find?MasterDataTypeID=LOC"
                lookup-key="_id"
                :lookup-labels="['Name']"
                :lookup-searchs="['_id', 'Name']"
                class="w-100"
              ></s-input>
            </template>
            <template #input_TrainerType="{ item, mode }">
              <s-input
                ref="refTrainerType"
                label="Trainer Type"
                v-model="item.TrainerType"
                use-list
                :items="['Internal', 'External']"
                class="w-100"
                @change="
                  (field, v1, v2, old, ctlRef) => {
                    item.TrainerName = '';
                  }
                "
              ></s-input>
            </template>
            <template #input_Trainer="{ item, mode }">
              <s-input
                v-if="item.TrainerType == 'Internal'"
                ref="refTrainerName"
                label="Trainer Name"
                v-model="item.Trainer"
                use-list
                :lookup-url="`/tenant/employee/find`"
                lookup-key="_id"
                :lookup-labels="['_id', 'Name']"
                :lookup-searchs="['_id', 'Name']"
                class="w-100"
              ></s-input>
              <s-input
                v-else
                ref="refTrainerName"
                label="Trainer Name"
                v-model="item.Trainer"
                class="w-100"
              ></s-input>
            </template>
          </s-form>
          <s-grid
            ref="gridAttendanceCheck"
            hide-select
            hide-search
            hide-sort
            hide-refresh-button
            hide-new-button
            hide-delete-button
            hide-action
            :config="data.cfgAttendanceCheck"
            :grid-fields="['IsPresent']"
          >
            <template #item_IsPresent="{ item, mode }">
              <s-input
                ref="refIsPresent"
                kind="checkbox"
                v-model="item.IsPresent"
                class="w-100"
              ></s-input>
            </template>
          </s-grid>
          <div
            v-if="data.isLoadingAttendance"
            class="flex justify-center gap-3 items-center loading"
          >
            <loader kind="circle" />
          </div>
        </div>
      </template>
      <template #form_tab_Material="{}">
        <Material
          test-schedule-type="TDC"
          v-model="data.material"
          :test-id="data.record._id"
        ></Material>
        <div
          v-if="data.isLoadingMaterial"
          class="flex justify-center gap-3 items-center loading"
        >
          <loader kind="circle" />
        </div>
      </template>
      <template #form_tab_Attachment="{ item, mode }">
        <s-grid-attachment
          :key="item._id"
          :journal-id="item._id"
          journal-type="TDC Training"
          :tags="[`TDC Training_${item._id}`]"
          :read-only="readOnlyDetail && readOnly"
          v-model="item.Attachment"
          single-save
        ></s-grid-attachment>
      </template>
      <template #form_tab_Assesment="{ item }">
        <!-- Staff -->
        <div
          v-show="data.trainingDetail?.AssessmentType === 'Assessment Staff'"
        >
          <s-grid
            ref="gridAssesmentStaff"
            hide-select
            hide-search
            hide-sort
            hide-refresh-button
            hide-new-button
            hide-delete-button
            :config="data.cfgAssesmentStaff"
            @select-data="editAssesmentStaffRecord"
            :grid-fields="['WrittenTest', 'PracticeTestScore', 'Certificate']"
          >
            <template #item_WrittenTest="{ item }">
              <div
                class="font-underline"
                @click="openModalAssesmentDetail(item, 'WrittenTest')"
              >
                {{ item.WrittenTest }}
              </div>
            </template>
            <template #item_PracticeTestScore="{ item }">
              {{ item.PracticeTestScore }}
            </template>
            <template #item_Certificate="{ item }">
              <s-button
                class="btn_primary w-fit"
                icon="Download"
                @click="generateCertificate(item)"
              ></s-button>
            </template>
          </s-grid>
        </div>
        <!-- Driver -->
        <div
          v-show="data.trainingDetail?.AssessmentType === 'Assessment Driver'"
        >
          <div v-show="data.viewAssesmentDrive == 'grid'">
            <div class="flex justify-end">
              <s-button
                icon="content-save"
                class="btn_primary"
                label="Practice Duration"
                @click="refreshGridPracticeDuration"
              />
            </div>
            <s-grid
              ref="gridAssesmentDriver"
              hide-select
              hide-search
              hide-sort
              hide-refresh-button
              hide-new-button
              hide-delete-button
              :config="data.cfgAssesmentDriver"
              @select-data="editAssesmentDriverRecord"
              :grid-fields="[
                'WrittenTest',
                'PracticeTestDuration',
                'PracticeTestScore',
                'Certificate',
              ]"
            >
              <template #item_WrittenTest="{ item }">
                <div
                  class="font-underline"
                  @click="openModalAssesmentDetail(item, 'WrittenTest')"
                >
                  {{ item.WrittenTest }}
                </div>
              </template>
              <template #item_PracticeTestDuration="{ item }">
                {{ item.PracticeTestDuration / 60 }}
              </template>
              <template #item_PracticeTestScore="{ item }">
                <div
                  class="font-underline"
                  @click="openModalAssesmentDetail(item, 'PracticeTestScore')"
                >
                  {{ item.PracticeTestScore }}
                </div>
              </template>
              <template #item_Certificate="{ item }">
                <s-button
                  class="btn_primary w-fit"
                  icon="Download"
                  @click="generateCertificate(item)"
                ></s-button>
              </template>
            </s-grid>
          </div>
          <!-- Grid Practice Duration -->
          <div v-show="data.viewAssesmentDrive == 'gridPracticeDuration'">
            <div class="flex justify-end gap-1">
              <s-button
                icon="plus"
                class="btn_primary"
                label="Add"
                @click="addPracticeDuration"
              />
              <s-button
                class="btn_warning back_btn"
                label="Back"
                icon="rewind"
                @click="backAssesmentDrive"
              />
            </div>
            <s-grid
              ref="gridPracticeDuration"
              hide-select
              hide-search
              hide-sort
              hide-refresh-button
              hide-new-button
              :config="data.cfgPracticeDuration"
              @select-data="editPracticeDuration"
              @delete-data="deletePracticeDuration"
              :grid-fields="['Date']"
            >
              <template #item_Date="{ item }">
                {{ moment(item.Date).format("YYYY-MM-DD").toString() }}
              </template>
            </s-grid>
          </div>
          <!-- Grid Practice Duration Details -->
          <div
            v-show="data.viewAssesmentDrive == 'gridDetailsPracticeDuration'"
          >
            <div class="flex justify-between gap-1">
              <s-input
                ref="refpracticeDurationDate"
                hide-label
                v-model="data.practiceDurationDate"
                class="w-100"
                kind="date"
              ></s-input>

              <div class="flex justify-end gap-1">
                <s-button
                  icon="content-save"
                  class="btn_primary"
                  label="Save"
                  @click="savePracticeDuration"
                />
                <s-button
                  class="btn_warning back_btn"
                  label="Back"
                  icon="rewind"
                  @click="
                    () => (data.viewAssesmentDrive = 'gridPracticeDuration')
                  "
                />
              </div>
            </div>
            <s-grid
              ref="gridDetailsPracticeDuration"
              hide-select
              hide-search
              hide-sort
              hide-refresh-button
              hide-new-button
              hide-delete-button
              hide-action
              :config="data.cfgDetailsPracticeDuration"
              :grid-fields="[
                'P2H',
                'PoliceNo',
                'ActivityName',
                'StartTime',
                'EndTime',
                'Duration',
                'Note',
              ]"
            >
              <template #item_P2H="{ item }">{{
                item.P2H ? "Yes" : "No"
              }}</template>
              <template #item_PoliceNo="{ item }">
                <s-input
                  ref="refPoliceNo"
                  hide-label
                  v-model="item.PoliceNo"
                  use-list
                  :lookup-url="`/tenant/asset/find`"
                  lookup-key="_id"
                  :lookup-labels="['Name']"
                  :lookup-searchs="['_id', 'Name']"
                  class="w-100"
                ></s-input>
              </template>
              <template #item_ActivityName="{ item }">
                <s-input
                  ref="refActivityName"
                  hide-label
                  v-model="item.ActivityName"
                  use-list
                  multiple
                  lookup-url="/tenant/masterdata/find?MasterDataTypeID=ActivityName"
                  lookup-key="_id"
                  :lookup-labels="['Name']"
                  :lookup-searchs="['_id', 'Name']"
                  class="w-100"
                ></s-input>
              </template>
              <template #item_StartTime="{ item }">
                <s-input
                  ref="refStartTime"
                  hide-label
                  v-model="item.StartTime"
                  class="w-100"
                  kind="time"
                  @change="
                    (field, v1, v2, old, ctlRef) => {
                      onGridRowFieldChangedPracticeDuration(
                        'StartTime',
                        v1,
                        v2,
                        old,
                        item
                      );
                    }
                  "
                ></s-input>
              </template>
              <template #item_EndTime="{ item }">
                <s-input
                  ref="refEndTime"
                  hide-label
                  v-model="item.EndTime"
                  class="w-100"
                  kind="time"
                  @change="
                    (field, v1, v2, old, ctlRef) => {
                      onGridRowFieldChangedPracticeDuration(
                        'EndTime',
                        v1,
                        v2,
                        old,
                        item
                      );
                    }
                  "
                ></s-input>
              </template>
              <template #item_Note="{ item }">
                <s-input
                  ref="refNote"
                  hide-label
                  v-model="item.Note"
                  class="w-100"
                ></s-input>
              </template>
            </s-grid>
          </div>
        </div>
        <!-- Mechanic -->
        <div
          v-show="data.trainingDetail?.AssessmentType === 'Assessment Mechanic'"
        >
          <s-grid
            ref="gridAssesmentMechanic"
            hide-select
            hide-search
            hide-sort
            hide-refresh-button
            hide-new-button
            hide-delete-button
            :config="data.cfgAssesmentStaff"
            @select-data="editAssesmentMechanicRecord"
            :grid-fields="['WrittenTest', 'PracticeTestScore', 'Certificate']"
          >
            <template #item_WrittenTest="{ item }">
              <div
                class="font-underline"
                @click="openModalAssesmentDetail(item, 'WrittenTest')"
              >
                {{ item.WrittenTest }}
              </div>
            </template>
            <template #item_PracticeTestScore="{ item }">
              {{ item.PracticeTestScore }}
            </template>
            <template #item_Certificate="{ item }">
              <s-button
                class="btn_primary w-fit"
                icon="Download"
                @click="generateCertificate(item)"
              ></s-button>
            </template>
          </s-grid>
        </div>
      </template>

      <!-- grid status -->
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #grid_TrainingTitle="{ item }">
        <s-input
          ref="refTrainingTitle"
          read-only
          v-model="item.TrainingTitle"
          class="w-50"
          use-list
          :lookup-url="`/hcm/tdctitle/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #grid_StatusDetail="{ item }">
        <div class="flex items-center">
          <log-trx
            :id="item.TrainingDevelopmentDetailID"
            v-if="helper.isShowLog(item.StatusDetail)"
          />
          <status-text :txt="item.StatusDetail" />
        </div>
      </template>
      <template #grid_TrainingStatus="{ item }">
        <div>{{ item.TrainingStatus ? "Close" : "Open" }}</div>
      </template>
      <template #grid_TrainingRequestor="{ item }">
        <s-input
          ref="refTrainingRequestor"
          v-model="item.TrainingRequestor"
          class="w-50"
          read-only
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
      <template #grid_AssessmentType="{ item }">
        {{ assessmentTypes[item._id] }}
      </template>

      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
      </template>
    </data-list>

    <!-- Preview -->
    <PreviewReport
      v-if="data.view == 'preview'"
      class="card w-full"
      title="Preview"
      @close="closePreview"
      :disable-print="helper.isDisablePrintPreview(data.record.Status)"
      :SourceType="data.jType"
      :SourceJournalID="data.record._id"
    >
      <template #buttons="props">
        <div class="flex gap-[1px] mr-2">
          <form-buttons-trx
            :disabled="inSubmission || loading"
            :status="data.record.Status"
            :journal-id="data.record._id"
            :posting-profile-id="data.record.PostingProfileID"
            :journal-type-id="data.jType"
            :auto-post="!waitTrxSubmit"
            moduleid="hcm"
            @preSubmit="trxPreSubmit"
            @postSubmit="trxPostSubmit"
            @errorSubmit="trxErrorSubmit"
          />
        </div>
      </template>
    </PreviewReport>
    <!-- Preview -->
    <PreviewReport
      v-if="data.view == 'previewDetail'"
      class="card w-full"
      title="Preview Detail"
      @close="closePreview"
      :disable-print="helper.isDisablePrintPreview(data.trainingDetail.Status)"
      :SourceType="data.jTypeTraningDetail"
      :SourceJournalID="data.trainingDetail._id"
    >
      <template #buttons="props">
        <div class="flex gap-[1px] mr-2">
          <form-buttons-trx
            :key="'trainingdetail_' + data.trainingDetail._id"
            :disabled="inSubmission || loading"
            :status="data.trainingDetail.Status"
            :journal-id="data.trainingDetail._id"
            :posting-profile-id="data.trainingDetail.PostingProfileID"
            :journal-type-id="data.jTypeTraningDetail"
            :auto-post="!waitTrxSubmitDetail"
            moduleid="hcm"
            @preSubmit="trxPreSubmitDetail"
            @postSubmit="trxPostSubmitDetail"
            @errorSubmit="trxErrorSubmitDetail"
          />
        </div>
      </template>
    </PreviewReport>

    <!-- Assesment Staff -->
    <AssesmentStaff
      v-show="data.view == 'formAssesmentStaff'"
      :view="data.view"
      :selectedAssesmentStaff="data.selectedAssesmentStaff"
      :assesmentStage="data.assesmentStage"
      :selectedStage="data.selectedStage"
      :selectedTemplate="data.selectedTemplate"
      :activeStage="data.activeStage"
      :activeQuestionTab="data.activeQuestionTab"
      @refreshGridAssesment="refreshGridAssesment"
      @handleClickStage="handleClickStage"
      @handleClickTemplate="handleClickTemplate"
      @update:view="handleChangeView"
    ></AssesmentStaff>

    <!-- Assesment Driver -->
    <AssesmentDriver
      v-show="data.view == 'formAssesmentDriver'"
      :view="data.view"
      :selectedAssesmentDriver="data.selectedAssesmentDriver"
      :assesmentStage="data.assesmentStage"
      :selectedStage="data.selectedStage"
      :selectedTemplate="data.selectedTemplate"
      :activeStage="data.activeStage"
      :activeQuestionTab="data.activeQuestionTab"
      @refreshGridAssesment="refreshGridAssesment"
      @handleClickStage="handleClickStage"
      @handleClickTemplate="handleClickTemplate"
      @update:view="handleChangeView"
    ></AssesmentDriver>

    <!-- Assesment Mechanic -->
    <AssesmentMechanic
      v-show="data.view == 'formAssesmentMechanic'"
      :view="data.view"
      :selectedAssesmentMechanic="data.selectedAssesmentMechanic"
      :assesmentStage="data.assesmentStage"
      :selectedStage="data.selectedStage"
      :selectedTemplate="data.selectedTemplate"
      :activeStage="data.activeStage"
      :activeQuestionTab="data.activeQuestionTab"
      @refreshGridAssesment="refreshGridAssesment"
      @handleClickStage="handleClickStage"
      @handleClickTemplate="handleClickTemplate"
      @update:view="handleChangeView"
    ></AssesmentMechanic>

    <!-- Modal Participant -->
    <s-modal
      :display="data.modal"
      hideButtons
      title="Action - Manpower Request"
      @beforeHide="data.modal = false"
    >
      <action-participant
        v-model="data.record._id"
        @generate="onGenerateParticipant"
      />
    </s-modal>

    <!-- Modal Assesment -->
    <s-modal
      :display="data.modalAssesment"
      hideButtons
      :title="`Detail - ${
        data.titleModal == 'WrittenTest' ? 'Written Test' : 'Practice Test'
      }`"
      @beforeHide="data.modalAssesment = false"
    >
      <div class="w-full flex flex-col">
        <!-- <pre>{{ data.titleModal == 'WrittenTest' ? data.selectedAssesmentDetail?.WrittenTestDetail : data.selectedAssesmentDetail?.PracticeTestScoreDetails }}</pre> -->
        <div
          class="mb-3 mx-5"
          v-if="data.titleModal == 'WrittenTest'"
          v-for="(writenTest, idx) in data.selectedAssesmentDetail
            ?.WrittenTestDetail"
          :key="idx"
        >
          <span class="font-bold mb-1">{{ writenTest.Stage }}</span>
          <br />
          <div v-for="(test, index) in writenTest.TestDetails" :key="index">
            <span>{{ index + 1 }}. </span>
            <span>{{ test.TemplateName }}</span
            >:
            <span>{{ test.Score }}</span>
          </div>
        </div>
        <div
          v-else
          v-for="(practiceTest, idx) in data.selectedAssesmentDetail
            ?.PracticeTestScoreDetails"
          :key="idx + '_practiceTest'"
        >
          <span>{{ idx + 1 }}. </span>
          <span>{{
            practiceTest.Type == "basic-movement"
              ? "Basic Movement"
              : "Main Road"
          }}</span
          >:
          <span>{{ practiceTest.FinalScore }}</span>
        </div>
      </div>
    </s-modal>
  </div>
</template>

<script setup>
// import { authStore } from "@/stores/auth";
import { reactive, ref, inject, onMounted, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import {
  createFormConfig,
  DataList,
  util,
  SForm,
  SButton,
  loadGridConfig,
  SGrid,
  SInput,
  SModal,
} from "suimjs";
import moment from "moment";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import { authStore } from "@/stores/auth";

import helper from "@/scripts/helper.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import StatusText from "@/components/common/StatusText.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import Material from "./widget/Training/Material.vue";
import AssesmentStaff from "./widget/Training/AssesmentStaff.vue";
import AssesmentDriver from "./widget/Training/AssesmentDriver.vue";
import AssesmentMechanic from "./widget/Training/AssesmentMechanic.vue";
import ActionParticipant from "./widget/Training/ActionParticipant.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";

import Loader from "@/components/common/Loader.vue";

layoutStore().name = "tenant";

const axios = inject("axios");
const auth = authStore();

const listControl = ref(null);
const gridParticipant = ref(null);
const formTrainingDetail = ref(null);
//Attendance
const gridAttendance = ref(null);
const formAttendance = ref(null);
const gridAttendanceCheck = ref(null);
//Assesment
const gridAssesmentStaff = ref(null);
const gridAssesmentDriver = ref(null);
const gridAssesmentMechanic = ref(null);
//Practice Duration
const gridPracticeDuration = ref(null);
const gridDetailsPracticeDuration = ref(null);

const gridAttachment = ref(SGridAttachment);

const assessmentTypes = ref({});

//File Import Participant
const fileInputParticipant = ref(null);

const FEATUREID = "DownloadCertificateTraining";
const assessmentProfile = authStore().getRBAC(FEATUREID);

const data = reactive({
  appMode: "grid",
  customFilter: null,
  tabsEdit: [],
  isLoading: false,
  isLoadingAttendance: false,
  isLoadingMaterial: false,
  record: {},
  participant: [],
  trainingDetail: {
    Status: "DRAFT",
  },
  material: [],
  attachment: [],
  trainingDetailFormCfg: {},
  viewAttendance: "grid",
  attendance: {},
  attendanceFormCfg: {},
  cfgAttendanceCheck: {},
  view: "tdc",
  jType: "TRAININGDEVELOPMENT",
  jTypeTraningDetail: "TRAININGDEVELOPMENTDETAIL",
  journalType: {},
  journalTypeTraningDetail: {},
  viewAssesmentDrive: "grid",
  cfgAssesmentStaff: {},
  cfgAssesmentDriver: {},
  cfgPracticeDuration: {},
  cfgDetailsPracticeDuration: {},
  recordAssesmentStaff: [],
  recordAssesmentDriver: [],
  recordAssesmentMechanic: [],
  selectedAssesmentStaff: {},
  selectedAssesmentDriver: {},
  selectedAssesmentMechanic: {},
  assesmentStage: [
    { name: "Pre-test", templates: [] },
    { name: "Post-test", templates: [] },
  ],
  selectedStage: {},
  selectedTemplate: {},
  activeStage: "Pre-test",
  activeQuestionTab: "",
  practiceDurationDate: "",
  modal: true,
  modalAssesment: true,
  selectedAssesmentDetail: {},
  titleModal: "",
});

watch(
  () => data.record,
  (nv) => {
    // console.log("Watch: ", nv)
    data.trainingDetail.TrainingDateFrom = nv.RequestTrainingDateFrom;
    data.trainingDetail.TrainingDateTo = nv.RequestTrainingDateTo;
  },
  { deep: true }
);

function newRecord(record) {
  record.Status = "DRAFT";
  const dateNow = moment(Date.now()).utc().format("YYYY-MM-DDTHH:mm:ss[Z]");
  record.RequestDate = dateNow;
  record.RequestTrainingDateFrom = dateNow;
  record.RequestTrainingDateTo = dateNow;
  openForm(record);
}

function openForm(record) {
  if (!record.CompanyID) {
    record.CompanyID = auth.appData.CompanyID;
  }
  data.record = record;
  // data.tabsEdit = [
  //   "General",
  //   "Participant",
  //   "Training Detail",
  //   "Attachment",
  //   "Attendance",
  //   "Material",
  // ];
  data.tabsEdit =
    record.Status === "POSTED"
      ? [
          "General",
          "Participant",
          "Training Detail",
          "Attachment",
          // "Attendance",
          // "Material",
        ]
      : ["General", "Participant"];

  util.nextTickN(2, () => {
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
  });
  refreshTrainingDetail();
}

function genFormCfg() {
  // cfg form training detail
  const cfgTrainingDetail = createFormConfig("", true);
  cfgTrainingDetail.addSection("", true).addRow(
    {
      field: "TrainingDateFrom",
      label: "Training Date From",
      kind: "date",
    },
    {
      field: "TrainingDateTo",
      label: "Training Date To",
      kind: "date",
    },
    {
      field: "Batch",
      label: "Batch",
      kind: "number",
    }
  );
  cfgTrainingDetail.addSection("", true).addRow(
    {
      field: "AssessmentType",
      label: "Assessment Type",
      kind: "text",
      kind: "text",
      useList: true,
      allowAdd: false,
      items: ["Assessment Driver", "Assessment Mechanic", "Assessment Staff"],
    },
    {
      field: "TrainerType",
      label: "Trainer Type",
      kind: "text",
      useList: true,
      allowAdd: false,
      items: ["Internal", "External"],
    },
    {
      field: "TrainerName",
      label: "Trainer Name",
      kind: "text",
    }
  );
  cfgTrainingDetail.addSection("", true).addRow(
    {
      field: "ExternalTraining",
      label: "External Training",
      kind: "checkbox",
    },
    {
      field: "ScheduledTraining",
      label: "Scheduled Training",
      kind: "checkbox",
    },
    {
      field: "OnlineTraining",
      label: "Online Training",
      kind: "checkbox",
    }
  );
  cfgTrainingDetail.addSection("", true).addRow(
    {
      field: "CostPerPerson",
      label: "Cost Per Person",
      kind: "number",
    },
    {
      field: "Location",
      label: "Location",
      kind: "text",
      useList: true,
      allowAdd: false,
      items: ["HO", "Local"],
    },
    {
      field: "Site",
      label: "Site",
      kind: "text",
      useList: true,
      allowAdd: false,
      lookupKey: "_id",
      lookupLabels: ["Label"],
      lookupSearchs: ["_id", "Label"],
      lookupUrl: "/tenant/dimension/find?DimensionType=Site",
    }
  );
  cfgTrainingDetail.addSection("", true).addRow(
    {
      field: "JournalTypeID",
      label: "Journal type ID",
      kind: "text",
      useList: true,
      allowAdd: false,
      lookupKey: "_id",
      lookupLabels: ["Name"],
      lookupSearchs: ["_id", "Name"],
      lookupUrl: "/hcm/tdcjournaltype/find",
      required: true,
    },
    {
      field: "Status",
      label: "Status",
      kind: "text",
      readOnly: true,
    }
  );
  cfgTrainingDetail.addSection("", true).addRow(
    {
      field: "Description",
      label: "Description",
      kind: "text",
      multiRow: 3,
    },
    {
      field: "DiscussionScope",
      label: "Discussion Scope",
      kind: "text",
      multiRow: 3,
    },
    {
      field: "RequiredTool",
      label: "Required Tool",
      kind: "text",
      multiRow: 3,
    }
  );
  cfgTrainingDetail.addSection("", true).addRow(
    {
      field: "MaterialClass",
      label: "Material Class",
      kind: "text",
      multiRow: 3,
    },
    {
      field: "PracticeClass",
      label: "Practice Class",
      kind: "text",
      multiRow: 3,
    }
  );
  data.trainingDetailFormCfg = cfgTrainingDetail.generateConfig();

  // cfg form attendance
  const cfgAttendance = createFormConfig("", true);
  cfgAttendance.addSection("", true).addRow(
    {
      field: "Date",
      label: "Date",
      kind: "date",
    },
    {
      field: "Time",
      label: "Time",
      kind: "time",
    },
    {
      field: "Topic",
      label: "Topic",
      kind: "text",
    }
  );
  cfgAttendance.addSection("", true).addRow(
    {
      field: "LocationID",
      label: "Location",
      kind: "text",
    },
    {
      field: "TrainerType",
      label: "Trainer Type",
      kind: "text",
    },
    {
      field: "Trainer",
      label: "Trainer",
      kind: "text",
    }
    // {
    //   field: "List",
    //   label: "List",
    //   kind: "text",
    // },
    // {
    //   field: "Attendace",
    //   label: "Attendace",
    //   kind: "number",
    // }
  );
  data.attendanceFormCfg = cfgAttendance.generateConfig();

  // cfg grid attendance check
  data.cfgAttendanceCheck = {
    fields: [
      {
        field: "EmployeeName",
        kind: "text",
        label: "Employee Name",
        labelField: "",
        readType: "show",
        input: {
          field: "EmployeeName",
          kind: "text",
          label: "Employee Name",
          lookupUrl: "",
          placeHolder: "Employee Name",
        },
      },
      {
        field: "IsPresent",
        kind: "text",
        label: "Present",
        labelField: "",
        readType: "show",
        input: {
          field: "IsPresent",
          kind: "text",
          label: "Present",
          lookupUrl: "",
          placeHolder: "Present",
        },
      },
    ],
    setting: {
      idField: "",
      keywordFields: ["_id", "Name"],
      sortable: ["_id"],
    },
  };
}

function saveParticipant() {
  const recordsParticipant = gridParticipant.value?.getGridRecords();
  recordsParticipant?.forEach(async (e) => {
    await axios.post("/hcm/tdcparticipant/save", e).then(
      (r) => {
        util.nextTickN(2, () => {
          gridParticipant.value.refreshGrid();
        });
      },
      (e) => {
        util.showError(e);
      }
    );
  });
}

async function saveDetail() {
  if (!data.trainingDetail.JournalTypeID && data.record.Status === "POSTED") {
    return util.showError("Journal type ID is required");
  }
  //Save Details
  setLoadingForm(true);
  await axios.post("/hcm/tdcdetail/save", data.trainingDetail).then(
    (r) => {
      setLoadingForm(false);
      util.nextTickN(2, () => {
        refreshTrainingDetail();
      });
    },
    (e) => {
      setLoadingForm(false);
      util.showError(e);
    }
  );
}

function saveMaterial() {
  data.isLoadingMaterial = true;
  axios.post("hcm/testschedule/save-training-schedule", data.material).then(
    async (r) => {
      util.showInfo("Material has been successful save");
      data.isLoadingMaterial = false;

      util.nextTickN(2, () => {
        refreshGridAssesment();
      });
    },
    (e) => {
      data.isLoadingMaterial = false;
      util.showError(e);
    }
  );
}

async function onPostSave(record) {
  saveParticipant();
  saveDetail();
  saveMaterial();
  // if (record.Status === "POSTED") {
  //   gridAttachment.value?.Save();
  // }
}

function handleChangeView(params) {
  data.view = params;
}

function onClosingTraining(item) {
  listControl.value.setFormLoading(true);
  item.TrainingStatus = true;
  // console.log("item: ",item)
  axios
    .post("/hcm/tdc/save", item)
    .then(
      (r) => {},
      (e) => {
        util.showError(e);
      }
    )
    .finally(() => {
      listControl.value.setFormLoading(false);
      listControl.value.setControlMode("grid");
      listControl.value.refreshGrid();
    });
}

async function onGridRowFieldChanged(name, v1, v2, old, record) {
  // change to filled automatically
  // if (name == "TrainingRequestor") {
  //   await axios.post("/tenant/employee/get", [v1]).then(
  //     (r) => {
  //       const dtResult = r.data;
  //       record.Dimension = dtResult.Dimension || [];
  //     },
  //     (e) => {
  //       util.showError(e);
  //     }
  //   );
  // }
}

async function getAssesmentType(item) {
  if (!item._id) return;

  const url = `/hcm/tdcdetail/find`;
  const param = {
    Take: 1,
    Where: { Field: "TrainingCenterID", Op: "$eq", Value: item._id },
  };

  try {
    const response = await axios.post(url, param);
    const dtResult = response.data;
    assessmentTypes.value[item._id] = dtResult[0]?.AssessmentType || "-";
  } catch (error) {
    console.error(error);
    assessmentTypes.value[item._id] = "-";
  }
}

function downloadTraining() {
  const param = {
    TrainingCenterID: data.record._id,
  };
  axios
    .post("hcm/tdc/generate-training", param, { responseType: "blob" })
    .then((response) => {
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const a = document.createElement("a");
      a.href = url;
      a.download = "Training_Report.xlsx";
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    })
    .catch((error) => {
      util.showError(error);
    });
}

// Participant
function newParticipantRecord(r) {
  const records = gridParticipant.value.getGridRecords();
  records.push({
    _id: "",
    TrainingCenterID: data.record._id,
    EmployeeID: "",
    ManpowerRequestID: "",
  });

  data.participant = records;

  gridParticipant.value.setGridRecords(records);
  updateItemsParticipant();
}

async function onGridRowFieldChangedParticipant(name, v1, v2, old, record) {
  if (name == "EmployeeID") {
    await axios.post("/tenant/employee/get", [v1]).then(
      (r) => {
        const dtResult = r.data;
        const Site = dtResult.Dimension?.find((e) => {
          return e.Key == "Site";
        });
        const Department = dtResult.Dimension?.find((e) => {
          return e.Key == "CC";
        });
        record.Site = Site ? Site.Value : "";
        record.Department = Department ? Department.Value : "";
      },
      (e) => {
        util.showError(e);
      }
    );
  }
  gridParticipant.value.setGridRecord(
    record,
    gridParticipant.value.getGridCurrentIndex()
  );
}

async function onGridRowDeleteParticipant(record, index) {
  // grid-delete="/hcm/tdcparticipant/delete"
  const records = gridParticipant.value.getGridRecords();
  const dtParticipant = records[index];
  if (dtParticipant._id !== "") {
    await axios.post("/hcm/tdcparticipant/delete", dtParticipant).then(
      (r) => {
        util.nextTickN(10, () => {
          // gridParticipant.value.refreshGrid();
          const newRecords = records
            .filter((dt, idx) => {
              return idx != index;
            })
            .map((nr, index) => ({ ...nr, index: index }));

          data.participant = newRecords;
          gridParticipant.value.setGridRecords(newRecords);
          updateItemsParticipant();
          // refreshTrainingDetail()
        });
      },
      (e) => {
        util.showError(e);
      }
    );
  } else {
    const newRecords = records
      .filter((dt, idx) => {
        return idx != index;
      })
      .map((nr, index) => ({ ...nr, index: index }));

    data.participant = newRecords;
    // console.log(newRecords)
    gridParticipant.value.setGridRecords(newRecords);
    updateItemsParticipant();
  }
}

function updateItemsParticipant() {
  gridParticipant.value.setGridRecords(data.participant);
}

function alterGridParticipantConfig(cfg) {
  cfg.fields = cfg.fields.filter(
    (item) =>
      item.field != "TrainingCenterID" && item.field != "ManpowerRequestID"
  );

  cfg.fields.splice(
    2,
    0,
    helper.gridColumnConfig({
      field: "NIK",
      label: "Employee No.",
      kind: "text",
    })
  );
  cfg.fields.splice(
    3,
    0,
    helper.gridColumnConfig({
      field: "Site",
      label: "Site",
      kind: "text",
    })
  );
  cfg.fields.splice(
    4,
    0,
    helper.gridColumnConfig({
      field: "Department",
      label: "Department",
      kind: "text",
    })
  );
}

function onGenerateParticipant(participant) {
  data.modal = false;
  participant.map((e) => {
    e.TrainingCenterID = data.record._id;
    return e;
  });
  const old = data.participant;
  const newRecords = old.concat(participant);
  // console.log("newRecords: ", newRecords)

  gridParticipant.value.setGridRecords(newRecords);
  data.participant = newRecords;

  util.nextTickN(2, () => {
    saveParticipant();
  });
}

function triggerFileInputParticipant() {
  fileInputParticipant.value.click();
}

function handleFileUploadParticipant(event) {
  const file = event.target.files[0];

  readFileAsBase64(file)
    .then((base64) => {
      let fileBase64 = {
        FileBase64: base64,
        TrainingCenterID: data.record._id,
      };

      axios.post("/hcm/tdc/import-participant", fileBase64).then(
        (r) => {
          util.showInfo("Success import participant");
          gridParticipant.value.refreshGrid();
        },
        (e) => {
          util.showError(e);
        }
      );
    })
    .catch((error) => {
      console.error("Error reading file:", error);
    });
}

function readFileAsBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);

    reader.onloadend = () => {
      resolve(reader.result.split(",")[1]);
    };

    reader.onerror = (error) => {
      reject(error);
    };
  });
}
// End Participant

// Detail
function refreshTrainingDetail() {
  const url = `/hcm/tdcdetail/find`;
  let param = {
    Take: 20,
    Sort: ["Created"],
    Where: {
      Op: "$or",
      items: [{ Op: "$eq", Field: "TrainingCenterID", Value: data.record._id }],
    },
  };

  axios.post(url, param).then(
    (r) => {
      const dtResult = r.data;
      if (dtResult.length > 0) {
        data.trainingDetail = dtResult[0];
      } else {
        data.trainingDetail = {
          _id: "",
          TrainingCenterID: data.record._id,
          TrainingDateFrom: data.record.RequestTrainingDateFrom,
          TrainingDateTo: data.record.RequestTrainingDateTo,
          TrainingType: "",
          ExternalTraining: false,
          AssessmentType: "",
          ScheduledTraining: false,
          TrainerType: "",
          TrainerName: "",
          Description: "",
          DiscussionScope: "",
          ParticipantTarget: "",
          CostPerPerson: 0,
          RequiredTool: "",
          Batch: 0,
          Status: "DRAFT",
          MaterialClass: "",
          PracticeClass: "",
          OnlineTraining: false,
          Location: "",
          Site: "",
        };
      }

      const tabs = data.trainingDetail?.AssessmentType;
      if (tabs !== "") {
        data.tabsEdit =
          data.trainingDetail.Status === "POSTED"
            ? [
                "General",
                "Participant",
                "Training Detail",
                "Attachment",
                "Attendance",
                "Material",
                "Assesment",
              ]
            : ["General", "Participant", "Training Detail", "Attachment"];
        listControl.value.refreshForm();
        // util.nextTickN(2, () => {
        //   if (readOnlyDetail.value === true) {
        //     formTrainingDetail.value.setMode("view");
        //   }
        // });
        util.nextTickN(2, () => {
          refreshGridAssesment();
        });
      }
    },
    (e) => {
      util.showError(e);
    }
  );
}
// End Detail

// Attendance
async function editAttendance(record) {
  data.viewAttendance = "form";
  data.attendance = record;

  const dtMapParticipant = await Promise.all(
    data.participant.map(async (e) => {
      const list = data.attendance?.List?.find(
        (o) => o.EmployeeID == e.EmployeeID
      );
      return {
        EmployeeID: list?.EmployeeID || e.EmployeeID,
        EmployeeName: e.Name,
        IsPresent: list?.IsPresent ?? true,
      };
    })
  );

  util.nextTickN(2, () => {
    gridAttendanceCheck.value.setRecords(dtMapParticipant);
  });
}

function addAttendance() {
  data.viewAttendance = "form";
  const dateNow = moment(Date.now()).utc().format("YYYY-MM-DD") + "T00:00:00Z";
  const timeNow = moment(Date.now()).format("HH:mm");
  data.attendance = {
    Date: dateNow,
    Time: timeNow,
  };
  const dtMapParticipant = data.participant.map((e) => {
    const dtParticipant = {
      EmployeeID: e.EmployeeID,
      EmployeeName: e.Name,
      IsPresent: true,
    };
    return dtParticipant;
  });

  util.nextTickN(2, () => {
    gridAttendanceCheck.value.setRecords(dtMapParticipant || []);
  });
}

function backAttendance() {
  data.viewAttendance = "grid";
  data.attendance = {};
}

async function saveAttendance() {
  // data.viewAttendance = 'grid'
  data.isLoadingAttendance = true;
  const dtAttendance = gridAttendance.value.getGridRecords();
  const find = dtAttendance.find((e) => {
    return e.Date == data.attendance.Date;
  });
  if (find) {
    util.showError("Data with the selected date is available!");
    data.isLoadingAttendance = false;
    return;
  }

  const dtAttendanceCheck = gridAttendanceCheck.value.getRecords();
  const Attendance = dtAttendanceCheck.filter((e) => {
    return e.IsPresent == true;
  });
  const dtMap = data.attendance;
  dtMap.List = dtAttendanceCheck;
  dtMap.TrainingCenterID = data.record._id;
  dtMap.Attendace = Attendance?.length || 0;
  await axios.post("/hcm/tdcattendance/save", dtMap).then(
    (r) => {
      backAttendance();
      util.nextTickN(2, () => {
        gridAttendance.value.refreshGrid();
      });
      data.isLoadingAttendance = false;
    },
    (e) => {
      data.isLoadingAttendance = false;
      util.showError(e);
    }
  );
}
// End Attendance

//Assesment
function averageWrittenTestScore(item) {
  if (!item?.TestDetails || !Array.isArray(item.TestDetails)) {
    return "N/A";
  }

  const doneTests = item.TestDetails.filter(
    (test) => test.Stage === "Pre-Test" && test.Status === "DONE"
  );

  if (doneTests.length === 0) {
    return "0";
  }

  const totalScore = doneTests.reduce((sum, test) => sum + test.Score, 0);
  const averageScore = (totalScore / doneTests.length).toFixed(2);
  return averageScore;
}

function getQuestionAnswer(type) {
  //map assesment question and answwer from API
  const url = "/hcm/tdc/get-question-answer";
  let dtTestDetails = [];

  if (type == "staff") {
    dtTestDetails = data.selectedAssesmentStaff.TestDetails;
  } else if (type == "driver") {
    dtTestDetails = data.selectedAssesmentDriver.TestDetails;
  } else if (type == "mechanic") {
    dtTestDetails = data.selectedAssesmentMechanic.TestDetails;
  }

  dtTestDetails.map(async (e) => {
    let param = {
      ParticipantID:
        type == "staff"
          ? data.selectedAssesmentStaff._id
          : type == "driver"
          ? data.selectedAssesmentDriver._id
          : data.selectedAssesmentMechanic._id,
      TemplateTestID: e.TemplateID,
    };
    await axios.post(url, param).then(
      (r) => {
        e.answer = r.data;
      },
      (e) => {
        e.answer = [];
        util.showError(e);
      }
    );
    return e;
  });

  if (type == "staff") {
    data.recordAssesmentStaff.TestDetails = dtTestDetails;
  } else if (type == "driver") {
    data.recordAssesmentDriver.TestDetails = dtTestDetails;
  } else if (type == "mechanic") {
    data.recordAssesmentMechanic.TestDetails = dtTestDetails;
  }

  //push assesment tracking
  data.assesmentStage.map((e) => {
    dtTestDetails.forEach((o) => {
      if (e.name.toLowerCase() === o.Stage.toLowerCase()) {
        e.templates.push(o);
      }
    });
    return e;
  });

  const defaultTemplate =
    data.assesmentStage[0].templates?.length > 0
      ? data.assesmentStage[0].templates[0]
      : {};
  handleClickStage(data.assesmentStage[0]);
  handleClickTemplate(defaultTemplate);
}

function handleClickStage(stage) {
  data.selectedStage = stage;
  data.activeStage = stage.name;

  const dtTemplates = stage.templates?.length > 0 ? stage.templates[0] : {};
  handleClickTemplate(dtTemplates);
}

function handleClickTemplate(template) {
  data.activeQuestionTab = template.TemplateID;
  data.selectedTemplate = template;
}

function refreshGridAssesment() {
  data.isLoading = true;
  const tabs = data.trainingDetail?.AssessmentType;
  if (tabs === "Assessment Staff") {
    gridAssesmentStaff.value?.setLoading(true);
  } else if (tabs === "Assessment Driver") {
    gridAssesmentDriver.value?.setLoading(true);
  } else if (tabs === "Assessment Mechanic") {
    gridAssesmentMechanic.value?.setLoading(true);
  }

  const url = `/hcm/tdcparticipant/gets?TrainingCenterID=${data.record._id}`;
  let param = {
    Skip: 0,
    Sort: ["-_id"],
    Take: 15,
  };

  axios.post(url, param).then(
    (r) => {
      let dtParticipant = r.data.data;
      data.participant = helper.cloneObject(r.data.data);
      data.recordAssesmentStaff = helper.cloneObject(r.data.data);
      data.recordAssesmentDriver = helper.cloneObject(r.data.data);

      let urlDetail =
        tabs === "Assessment Staff"
          ? "/hcm/tdc/assesment-staff"
          : tabs === "Assessment Driver"
          ? "/hcm/tdc/assesment-driver"
          : tabs === "Assessment Mechanic"
          ? "/hcm/tdc/assesment-mechanic"
          : "";

      if (urlDetail !== "") {
        const paramDetail = {
          TrainingCenterID: data.record._id,
          Take: -1,
          Skip: 0,
        };

        axios.post(urlDetail, paramDetail).then(
          (res) => {
            // console.log("assesment: ", res.data)
            const resDetail = res.data.data;
            dtParticipant.map((e) => {
              const find = resDetail.find((o) => {
                return o.ParticipantID == e._id;
              });
              e.WrittenTest = find.WrittenTest || 0;
              e.WrittenTestDetail = find.WrittenTestDetail || [];
              e.PracticeTestDuration = find.PracticeTestDuration || 0;
              e.PracticeTestScore =
                find.PracticeTestScore || find.PracticeTest || 0;
              e.PracticeTestScoreDetails = find.PracticeTestScoreDetails || [];
              return e;
            });

            if (tabs === "Assessment Staff") {
              gridAssesmentStaff.value.setLoading(false);
              gridAssesmentStaff.value.setRecords(dtParticipant);
              urlDetail = "/hcm/tdc/assesment-staff";
            } else if (tabs === "Assessment Driver") {
              gridAssesmentDriver.value?.setLoading(false);
              gridAssesmentDriver.value?.setRecords(dtParticipant);
              urlDetail = "/hcm/tdc/assesment-driver";
            } else if (tabs === "Assessment Mechanic") {
              gridAssesmentMechanic.value?.setLoading(false);
              gridAssesmentMechanic.value?.setRecords(dtParticipant);
              urlDetail = "/hcm/tdc/assesment-mechanic";
            }
            data.isLoading = false;
            util.nextTickN(2, () => {
              if (readOnlyDetail.value === true) {
                gridAssesmentMechanic.value?.setMode("view");
              }
            });
          },
          (err) => {
            util.showError(err);
          }
        );
      }
    },
    (e) => {
      if (tabs === "Assesment Staff") {
        gridAssesmentStaff.value.setLoading(false);
      } else if (tabs === "Assesment Driver") {
        gridAssesmentDriver.value.setLoading(false);
      } else if (tabs === "Assesment Mechanic") {
        gridAssesmentMechanic.value.setLoading(false);
      }

      data.isLoading = false;
      util.showError(e);
    }
  );
}

function openModalAssesmentDetail(item, title) {
  data.selectedAssesmentDetail = item;
  data.titleModal = title;
  data.modalAssesment = true;
}
//End Assesment

//Assesment Staff
function alterGridAssesmentStaffConfig(cfg) {
  cfg.fields = cfg.fields.filter(
    (item) =>
      item.field != "TrainingCenterID" && item.field != "ManpowerRequestID"
  );
  cfg.fields.splice(
    1,
    0,
    helper.gridColumnConfig({
      field: "Name",
      label: "Employee Name",
      kind: "text",
    })
  );
  cfg.fields.splice(
    2,
    0,
    helper.gridColumnConfig({
      field: "NIK",
      label: "Employee No.",
      kind: "text",
    })
  );
  cfg.fields.splice(
    3,
    0,
    helper.gridColumnConfig({
      field: "WrittenTest",
      label: "Written Test",
      kind: "text",
    })
  );
  cfg.fields.splice(
    4,
    0,
    helper.gridColumnConfig({
      field: "PracticeTestScore",
      label: "Practice Test",
      kind: "text",
    })
  );
  if (!assessmentProfile.canSpecial1) {
    cfg.fields.splice(
      5,
      0,
      helper.gridColumnConfig({
        field: "Certificate",
        label: "Certificate",
        kind: "text",
      })
    );
  }
}

function editAssesmentStaffRecord(dt, index) {
  data.selectedAssesmentStaff = dt;
  data.assesmentStage = [
    {
      name: "Pre-test",
      templates: [],
    },
    {
      name: "Post-test",
      templates: [],
    },
  ];

  util.nextTickN(2, () => {
    handleChangeView("formAssesmentStaff");
    // data.view = 'formAssesmentStaff'
    getQuestionAnswer("staff");
  });
}
//End Assesment Staff

//Assesment Driver
function alterGridAssesmentDriverConfig(cfg) {
  cfg.fields = cfg.fields.filter(
    (item) =>
      item.field != "TrainingCenterID" && item.field != "ManpowerRequestID"
  );
  cfg.fields.splice(
    1,
    0,
    helper.gridColumnConfig({
      field: "Name",
      label: "Employee Name",
      kind: "text",
    })
  );
  cfg.fields.splice(
    2,
    0,
    helper.gridColumnConfig({
      field: "NIK",
      label: "Employee No.",
      kind: "text",
    })
  );
  cfg.fields.splice(
    3,
    0,
    helper.gridColumnConfig({
      field: "WrittenTest",
      label: "Written Test",
      kind: "text",
    })
  );
  cfg.fields.splice(
    4,
    0,
    helper.gridColumnConfig({
      field: "PracticeTestDuration",
      label: "Practice Test Duration (Hour)",
      kind: "text",
    })
  );
  cfg.fields.splice(
    5,
    0,
    helper.gridColumnConfig({
      field: "PracticeTestScore",
      label: "Practice Test",
      kind: "text",
    })
  );
  if (assessmentProfile.canSpecial1) {
    cfg.fields.splice(
      6,
      0,
      helper.gridColumnConfig({
        field: "Certificate",
        label: "Certificate",
        kind: "text",
      })
    );
  }
}

function genGridPracticeDuration() {
  data.cfgPracticeDuration = {
    fields: [
      {
        field: "Date",
        kind: "text",
        label: "Date",
        labelField: "",
        readType: "show",
        input: {
          field: "Date",
          kind: "text",
          label: "Date",
          lookupUrl: "",
          placeHolder: "Date",
        },
      },
    ],
    setting: {
      idField: "",
      keywordFields: ["_id", "Date"],
      sortable: ["_id"],
    },
  };

  data.cfgDetailsPracticeDuration = {
    fields: [
      {
        field: "EmployeeID",
        kind: "text",
        label: "Employee ID",
        labelField: "",
        readType: "show",
        input: {
          field: "EmployeeID",
          kind: "text",
          label: "Employee ID",
          lookupUrl: "",
          placeHolder: "Employee ID",
        },
      },
      {
        field: "Name",
        kind: "text",
        label: "Employee Name",
        labelField: "",
        readType: "show",
        input: {
          field: "Name",
          kind: "text",
          label: "Employee Name",
          lookupUrl: "",
          placeHolder: "Employee Name",
        },
      },
      {
        field: "P2H",
        kind: "text",
        label: "P2H",
        labelField: "",
        readType: "show",
        input: {
          field: "P2H",
          kind: "text",
          label: "P2H",
          lookupUrl: "",
          placeHolder: "P2H",
        },
      },
      {
        field: "PoliceNo",
        kind: "text",
        label: "Police No.",
        labelField: "",
        readType: "show",
        input: {
          field: "PoliceNo",
          kind: "text",
          label: "Police No.",
          lookupUrl: "",
          placeHolder: "Police No.",
        },
      },
      {
        field: "ActivityName",
        kind: "text",
        label: "Activity Name",
        labelField: "",
        readType: "show",
        input: {
          field: "ActivityName",
          kind: "text",
          label: "Activity Name",
          lookupUrl: "",
          placeHolder: "Activity Name",
        },
      },
      {
        field: "StartTime",
        kind: "text",
        label: "Start Time",
        labelField: "",
        readType: "show",
        input: {
          field: "StartTime",
          kind: "text",
          label: "Start Time",
          lookupUrl: "",
          placeHolder: "Start Time",
        },
      },
      {
        field: "EndTime",
        kind: "text",
        label: "End Time",
        labelField: "",
        readType: "show",
        input: {
          field: "EndTime",
          kind: "text",
          label: "End Time",
          lookupUrl: "",
          placeHolder: "End Time",
        },
      },
      {
        field: "Duration",
        kind: "text",
        label: "Duration (Minutes)",
        labelField: "",
        readType: "show",
        input: {
          field: "Duration",
          kind: "text",
          label: "Duration",
          lookupUrl: "",
          placeHolder: "Duration",
        },
      },
      {
        field: "Note",
        kind: "text",
        label: "Note",
        labelField: "",
        readType: "show",
        input: {
          field: "Note",
          kind: "text",
          label: "Note",
          lookupUrl: "",
          placeHolder: "Note",
        },
      },
    ],
    setting: {
      idField: "",
      keywordFields: ["_id", "Name"],
      sortable: ["_id"],
    },
  };
}

function editAssesmentDriverRecord(dt, index) {
  data.selectedAssesmentDriver = dt;
  data.assesmentStage = [
    {
      name: "Pre-test",
      templates: [],
    },
    {
      name: "Post-test",
      templates: [],
    },
  ];

  util.nextTickN(2, () => {
    handleChangeView("formAssesmentDriver");
    getQuestionAnswer("driver");
  });
}

function refreshGridPracticeDuration() {
  data.viewAssesmentDrive = "gridPracticeDuration";

  gridPracticeDuration.value.setLoading(true);
  const url = `/hcm/tdcpracticeduration/get-date`; //?TrainingCenterID=${data.record._id}`;
  let param = {
    Skip: 0,
    Sort: ["-_id"],
    Take: -1,
    Where: {
      Op: "$or",
      items: [{ Op: "$eq", Field: "TrainingCenterID", Value: data.record._id }],
    },
  };

  axios.post(url, param).then(
    (r) => {
      gridPracticeDuration.value.setLoading(false);
      gridPracticeDuration.value.setRecords(r.data);
    },
    (e) => {
      gridPracticeDuration.value.setLoading(false);
      util.showError(e);
    }
  );
}

function editPracticeDuration(dt, index) {
  data.viewAssesmentDrive = "gridDetailsPracticeDuration";
  data.practiceDurationDate = dt.Date;

  gridDetailsPracticeDuration.value.setLoading(true);
  const url = `/hcm/tdcpracticeduration/gets?Date=${dt.Date}`;
  let param = {
    Skip: 0,
    Sort: ["-_id"],
    Take: -1,
  };

  axios.post(url, param).then(
    (r) => {
      const dt = r.data.data;
      dt.map((e) => {
        const participant = data.recordAssesmentDriver.find((o) => {
          return o._id == e.ParticipantID;
        });

        e.EmployeeID = participant?.EmployeeID || "-";
        e.Name = participant?.Name || "-";

        e.StartTime = moment(e.StartTime).format("HH:mm").toString();
        e.EndTime = moment(e.EndTime).format("HH:mm").toString();

        return e;
      });

      gridDetailsPracticeDuration.value.setLoading(false);
      gridDetailsPracticeDuration.value.setRecords(r.data.data);
    },
    (e) => {
      gridDetailsPracticeDuration.value.setLoading(false);
      util.showError(e);
    }
  );
}

function addPracticeDuration() {
  data.viewAssesmentDrive = "gridDetailsPracticeDuration";

  gridDetailsPracticeDuration.value.setLoading(true);

  data.practiceDurationDate = "";
  const dtAssesmentDriver = helper.cloneObject(data.recordAssesmentDriver);
  const dtMap = [];
  dtAssesmentDriver.forEach((e) => {
    const map = {
      _id: "",
      Date: "",
      ParticipantID: e._id || "",
      TrainingCenterID: data.record._id,
      EmployeeID: e.EmployeeID || "",
      Name: e.Name || "",
      P2H: false,
      PoliceNo: "",
      ActivityName: [],
      StartTime: "",
      EndTime: "",
      Duration: 0,
      Note: "",
    };
    dtMap.push(map);
  });

  gridDetailsPracticeDuration.value.setRecords(dtMap);
  gridDetailsPracticeDuration.value.setLoading(false);
}

function savePracticeDuration() {
  const dt = gridDetailsPracticeDuration.value.getRecords();
  dt.map((e) => {
    const dateNow = moment(Date.now()).format("YYYY-MM-DD").toString(); //2025-01-21";
    // const dateTimeNow = moment(`${dateNow}T00:00+07:00`).format("YYYY-MM-DDTHH:mm:ss.SSSSSSZ").toString();

    const dateStartTime = `${dateNow}T${
      e.StartTime == "" ? "00:00" : e.StartTime
    }`;
    const dateEndTime = `${dateNow}T${e.EndTime == "" ? "00:00" : e.EndTime}`;

    const formattedStartTime = moment(dateStartTime + "+07:00").format(
      "YYYY-MM-DDTHH:mm:ss.SSSSSSZ"
    );
    const formattedEndTime = moment(dateEndTime + "+07:00").format(
      "YYYY-MM-DDTHH:mm:ss.SSSSSSZ"
    );

    e.StartTime = formattedStartTime;
    e.EndTime = formattedEndTime;

    const formatPracticeDurationDate =
      data.practiceDurationDate == ""
        ? dateNow
        : moment(data.practiceDurationDate).format("YYYY-MM-DD").toString();
    const formatDate = moment(`${formatPracticeDurationDate}T00:00+07:00`)
      .format("YYYY-MM-DDTHH:mm:ss.SSSSSSZ")
      .toString();
    e.Date = formatDate;

    return e;
  });

  const params = {
    Details: dt,
  };

  axios.post("/hcm/tdcpracticeduration/save", params).then(
    (res) => {
      refreshGridPracticeDuration();
    },
    (err) => {
      util.showError(err);
    }
  );
}

function deletePracticeDuration(items, index) {
  const dtDate = items.items[index];
  const newDt = items.items.filter((e, i) => {
    return i !== index;
  });
  gridPracticeDuration.value.setLoading(true);
  const url = `/hcm/tdcpracticeduration/gets?Date=${dtDate.Date}`;
  let param = {
    Skip: 0,
    Sort: ["-_id"],
    Take: -1,
  };

  axios.post(url, param).then(
    async (r) => {
      const dt = r.data.data;
      await dt.forEach(async (e) => {
        axios.post("/hcm/tdcpracticeduration/delete", e).then(
          (res) => {},
          (err) => {
            util.showError(err);
          }
        );
      });
      // console.log(items, index)
      gridPracticeDuration.value.setRecords(newDt);
      gridPracticeDuration.value.setLoading(false);
    },
    (e) => {
      gridPracticeDuration.value.setLoading(false);
      util.showError(e);
    }
  );
}

function onGridRowFieldChangedPracticeDuration(name, v1, v2, old, record) {
  if (name == "StartTime" && record.EndTime !== "") {
    const start = moment(v1, "HH:mm", true);
    const end = moment(record.EndTime, "HH:mm", true);
    const durationInMinutes = end.diff(start, "minutes");

    record.Duration = durationInMinutes;
  } else if (name == "EndTime" && record.StartTime !== "") {
    const start = moment(record.StartTime, "HH:mm", true);
    const end = moment(v1, "HH:mm", true);
    const durationInMinutes = end.diff(start, "minutes");

    record.Duration = durationInMinutes;
  }
}

function backAssesmentDrive() {
  data.viewAssesmentDrive = "grid";

  util.nextTickN(2, () => {
    refreshGridAssesment();
  });
}
//End Assesment Driver

//Assesment Staff
function editAssesmentMechanicRecord(dt, index) {
  data.selectedAssesmentMechanic = dt;
  data.assesmentStage = [
    {
      name: "Pre-test",
      templates: [],
    },
    {
      name: "Post-test",
      templates: [],
    },
  ];

  util.nextTickN(2, () => {
    handleChangeView("formAssesmentMechanic");
    // data.view = 'formAssesmentStaff'
    getQuestionAnswer("mechanic");
  });
}
// approval training
function getJurnalType(id) {
  if (id === "" || id === null) {
    data.journalType = {};
    data.record.PostingProfileID = "";
    return;
  }
  // listControl.value.setFormLoading(true);
  axios
    .post("/hcm/tdcjournaltype/get", [id])
    .then(
      (r) => {
        data.journalType = r.data;
        data.record.PostingProfileID = r.data.PostingProfileID;
      },
      (e) => {
        data.journalType = {};
        data.record.PostingProfileID = "";
        util.showError(e);
      }
    )
    .finally(() => {
      // listControl.value.setFormLoading(false);
    });
}
function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
  }
}
function closePreview() {
  data.appMode = "grid";
  data.view = "tdc";
}

const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "SUBMITTED", "READY"].includes(data.record.Status);
  },
});
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    listControl.value.setFormFieldAttr(e.field, "required", required);
  });
}
function trxPreSubmit(status, action, doSubmit) {
  if (waitTrxSubmit.value) {
    listControl.value.setFormCurrentTab(0);
    trxSubmit(doSubmit);
  }
}
function trxSubmit(doSubmit) {
  setFormRequired(true);
  util.nextTickN(2, () => {
    const valid = listControl.value.formValidate();
    if (valid) {
      setLoadingForm(true);
      listControl.value.submitForm(
        data.record,
        () => {
          doSubmit();
        },
        () => {
          setLoadingForm(false);
        }
      );
    }
    setFormRequired(false);
  });
}
function trxPostSubmit(record, action) {
  setLoadingForm(false);
  closePreview();
  setModeGrid();
}
function trxErrorSubmit(e) {
  setLoadingForm(false);
}
function setModeGrid() {
  listControl.value.setControlMode("grid");
  listControl.value.refreshList();
}

// approval training detail
function getJurnalTypeDetail(id) {
  if (id === "" || id === null) {
    data.journalType = {};
    data.trainingDetail.PostingProfileID = "";
    return;
  }
  // listControl.value.setFormLoading(true);
  axios
    .post("/hcm/tdcjournaltype/get", [id])
    .then(
      (r) => {
        data.journalTypeTraningDetail = r.data;
        data.trainingDetail.PostingProfileID = r.data.PostingProfileID;
      },
      (e) => {
        data.journalTypeTraningDetail = {};
        data.trainingDetail.PostingProfileID = "";
        util.showError(e);
      }
    )
    .finally(() => {
      // listControl.value.setFormLoading(false);
    });
}

const readOnlyDetail = computed({
  get() {
    return !["", "DRAFT"].includes(data.trainingDetail.Status);
  },
});
const waitTrxSubmitDetail = computed({
  get() {
    return ["DRAFT", "SUBMITTED", "READY"].includes(data.trainingDetail.Status);
  },
});
function setFormRequiredDetail(required) {
  formTrainingDetail.value.getAllField().forEach((e) => {
    formTrainingDetail.value.setFieldAttr(e.field, "required", required);
  });
}
function trxPreSubmitDetail(status, action, doSubmit) {
  if (waitTrxSubmitDetail.value) {
    trxSubmitDetail(doSubmit);
  }
}
function trxSubmitDetail(doSubmit) {
  // setFormRequiredDetail(true);
  util.nextTickN(2, () => {
    const valid = formTrainingDetail.value.validate();
    if (valid) {
      setLoadingForm(true);
      saveDetail()
        .then(() => {
          doSubmit();
        })
        .catch(() => {
          setLoadingForm(false);
        });
    }
    // setFormRequired(false);
  });
}
function trxPostSubmitDetail(record, action) {
  setLoadingForm(false);
  closePreview();
  setModeGrid();
}
function trxErrorSubmitDetail(e) {
  setLoadingForm(false);
}
const showSaveTrainingDetail = computed({
  get() {
    return (
      ["", "DRAFT"].includes(data.trainingDetail.Status) &&
      listControl.value.getFormCurrentTab() === 2
    );
  },
});
const showSaveAttendance = computed({
  get() {
    return (
      ["POSTED"].includes(data.trainingDetail.Status) &&
      listControl.value.getFormCurrentTab() === 4
    );
  },
});
const showSaveMaterial = computed({
  get() {
    return (
      ["POSTED"].includes(data.trainingDetail.Status) &&
      listControl.value.getFormCurrentTab() === 5
    );
  },
});
const showSaveAssesment = computed({
  get() {
    return (
      ["POSTED"].includes(data.trainingDetail.Status) &&
      listControl.value.getFormCurrentTab() === 6
    );
  },
});
const showPreview = computed({
  get() {
    if (
      listControl.value?.getFormCurrentTab() == 0 ||
      listControl.value?.getFormCurrentTab() == 1
    ) {
      return "previewGeneral";
    } else if (
      listControl.value?.getFormCurrentTab() == 2 ||
      listControl.value?.getFormCurrentTab() == 3
    ) {
      return "previewDetail";
    } else {
      return "";
    }
  },
});
function alterGridConfig(cfg) {
  cfg.fields.splice(
    7,
    0,
    helper.gridColumnConfig({
      field: "StatusDetail",
      label: "Status training detail",
      kind: "text",
    })
  );
  cfg.fields.splice(
    4,
    0,
    helper.gridColumnConfig({
      field: "AssessmentType",
      label: "Assessment Type",
      kind: "text",
    })
  );
}
//End Assesment Staff

//Generate Certificate
function goto(destination) {
  // window.location.href = import.meta.env.VITE_API_URL + destination;
  window.open(import.meta.env.VITE_API_URL + destination, "_blank");
}
function generateCertificate(item) {
  const tabs = data.trainingDetail?.AssessmentType;
  if (tabs === "Assessment Staff") {
    gridAssesmentStaff.value?.setLoading(true);
  } else if (tabs === "Assessment Driver") {
    gridAssesmentDriver.value?.setLoading(true);
  } else if (tabs === "Assessment Mechanic") {
    gridAssesmentMechanic.value?.setLoading(true);
  }
  const param = {
    ParticipantID: item._id,
  };
  axios.post("/hcm/tdc/generate-certificate", param).then(
    async (r) => {
      if (tabs === "Assessment Staff") {
        gridAssesmentStaff.value?.setLoading(false);
      } else if (tabs === "Assessment Driver") {
        gridAssesmentDriver.value?.setLoading(false);
      } else if (tabs === "Assessment Mechanic") {
        gridAssesmentMechanic.value?.setLoading(false);
      }

      util.showInfo("Certificate has been successfuly generate");
      goto(`/asset/view?id=${item._id}`);
    },
    (e) => {
      if (tabs === "Assessment Staff") {
        gridAssesmentStaff.value?.setLoading(false);
      } else if (tabs === "Assessment Driver") {
        gridAssesmentDriver.value?.setLoading(false);
      } else if (tabs === "Assessment Mechanic") {
        gridAssesmentMechanic.value?.setLoading(false);
      }

      util.showError(e);
    }
  );
}
//End Generate Certificate

//filter
function initNewItemFilter(item) {
  item.Text = "";
}
function changeFilter(item, filters) {
  if (item.Text != "") {
    filters.push({
      Op: "$or",
      Items: ["_id", "TrainingTitle"].map((e) => {
        return {
          Op: "$contains",
          Field: e,
          Value: [item.Text],
        };
      }),
    });
  }
}
function refreshGrid() {
  listControl.value.refreshGrid();
}
onMounted(() => {
  data.modal = false;
  data.modalAssesment = false;
  genFormCfg();
  //get data participant grid
  loadGridConfig(axios, "/hcm/tdcparticipant/gridconfig").then(
    (r) => {
      data.cfgAssesmentStaff = helper.cloneObject(r);
      data.cfgAssesmentDriver = helper.cloneObject(r);
      alterGridAssesmentStaffConfig(data.cfgAssesmentStaff);
      alterGridAssesmentDriverConfig(data.cfgAssesmentDriver);

      genGridPracticeDuration();
    },
    (e) => util.showError(e)
  );

  util.nextTickN(500, () => {
    if (listControl.value) {
      watch(
        () => listControl.value.getGridRecords(),
        (nv) => {
          if (nv?.length) {
            nv.forEach((item) => {
              if (!assessmentTypes.value[item._id]) {
                getAssesmentType(item);
              }
            });
          }
        },
        { deep: true, immediate: true }
      );
    }
  });
});
</script>

<style>
.gridPracticeAssesmentStaff .suim_table > tbody > tr:last-child {
  background-color: #d3d3d3;
}

.loading {
  min-height: calc(-300px + 100vh);
  position: fixed;
  width: 100%;
  background: #8080803b;
  top: 0px;
  bottom: 0px;
  left: 0px;
  right: 0px;
}
</style>

<style scoped>
.font-underline {
  text-decoration-line: underline !important;
}
.font-underline:hover {
  color: blue !important;
}
</style>
