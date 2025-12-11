<template>
  <data-list
    class="card rsca_transaction"
    ref="listControl"
    :title="data.titleForm"
    grid-config="/she/induction/gridconfig"
    form-config="/she/induction/formconfig"
    grid-read="/she/induction/gets"
    form-read="/she/induction/get"
    grid-mode="grid"
    grid-delete="/she/induction/delete"
    form-keep-label
    form-insert="/she/induction/save"
    form-update="/she/induction/save"
    :init-app-mode="data.appMode"
    :form-fields="['Dimension']"
    :grid-fields="['Dimension', 'Status']"
    :form-tabs-edit="[
      'General',
      'Attendee',
      'Materials',
      'Assesment',
      'Attachments',
    ]"
    :form-tabs-view="[
      'General',
      'Attendee',
      'Materials',
      'Assesment',
      'Attachments',
    ]"
    :grid-custom-filter="customFilter"
    grid-hide-select
    stay-on-form-after-save
    @form-edit-data="editRecord"
    @form-new-data="newData"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    @alterGridConfig="alterGridConfig"
    @controlModeChanged="onControlModeChanged"
    @form-field-change="onFormFieldChange"
  >
    <template #grid_header_search="{ config }">
      <div class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter">
        <s-input
          kind="date"
          label="Date From"
          class="w-[200px]"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Date To"
          class="w-[200px]"
          v-model="data.search.DateTo"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refName"
          v-model="data.search.Name"
          class="w-[300px]"
          label="Name"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refCategory"
          v-model="data.search.Category"
          lookup-key="_id"
          label="Category"
          class="w-[200px]"
          use-list
          :items="['New Employee', 'Mutasi/Rotasi', 'Pulang Cuti']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refStatus"
          v-model="data.search.Status"
          lookup-key="_id"
          label="Status"
          class="w-[200px]"
          use-list
          :items="['DRAFT', 'SUBMITTED', 'READY', 'POSTED', 'REJECTED']"
          @change="refreshData"
        ></s-input>
      </div>
    </template>
    <template #form_tab_Attendee="{ item }">
      <s-grid
        class="attendee_lines"
        ref="gridAttendee"
        :config="data.cfgGridAttendee"
        hide-search
        hide-sort
        hide-refresh-button
        hide-edit
        hide-select
        :hideDeleteButton="readOnly"
        hide-paging
        :hide-control="readOnly"
        :editor="!readOnly"
        auto-commit-line
        no-confirm-delete
        @new-data="newAttendee"
      >
        <template #item_EmploymentType="{ item, idx }">
          <s-toggle
            :readOnly="readOnly"
            v-model="item.EmploymentType"
            class="w-[120px] mt-0.5"
            yes-label="Internal"
            no-label="External"
            @change="
              onFormFieldChange(
                'EmploymentType',
                item.PatientType,
                '',
                '',
                item
              )
            "
          />
        </template>
        <template #item_Name="{ item, idx }">
          <s-input
            :readOnly="readOnly"
            v-if="item.EmploymentType"
            ref="refName"
            v-model="item.Name"
            class="w-full"
            use-list
            allowAdd
            lookup-url="/tenant/employee/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="onFormFieldChange('AttendeeName', item.Name, '', '', item)"
          ></s-input>

          <s-input
            v-else
            :readOnly="readOnly"
            ref="refName"
            v-model="item.Name"
            class="w-full"
          ></s-input>
        </template>
        <template #item_Position="{ item, idx }">
          <s-input
            :readOnly="readOnly"
            v-if="!item.EmploymentType"
            ref="refName"
            v-model="item.Position"
            class="w-full"
          ></s-input>
        </template>
        <template #item_button_delete="{ item, idx }">
          <a
            v-if="!readOnly"
            @click="deleteAttendee(item)"
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
      </s-grid>
    </template>
    <template #form_tab_Materials="{ item }">
      <s-grid
        class="materials_lines"
        ref="gridMaterials"
        :config="data.cfgGridMaterials"
        hide-search
        hide-sort
        hide-refresh-button
        hide-edit
        :hide-control="readOnly"
        :hideDeleteButton="readOnly"
        hide-select
        hide-paging
        :editor="!readOnly"
        auto-commit-line
        no-confirm-delete
        @new-data="newMaterials"
      >
        <template #item_buttons_1="{ item }">
          <action-attachment
            v-if="!readOnly"
            kind="materials"
            :ref-id="data.record._id"
            :tags="buildTags(item, 'materials')"
            :tags-for-get="[`materials_${item._id}`]"
          />
        </template>
        <template #item_button_delete="{ item, idx }">
          <a
            v-if="!readOnly"
            @click="deleteMaterials(item)"
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
      </s-grid>
    </template>
    <template #form_tab_Assesment="{ item }">
      <s-grid
        class="assesment_lines"
        ref="gridAssesment"
        :config="data.cfgGridAssesment"
        hide-search
        hide-sort
        hide-refresh-button
        hide-edit
        :hide-control="readOnly"
        :hideDeleteButton="readOnly"
        hide-select
        hide-paging
        hide-new-button
        :editor="!readOnly"
        auto-commit-line
        no-confirm-delete
        @new-data="newAssesment"
        @delete-data="deleteAssesment"
      >
        <template #header_buttons_1="{ config }">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="input_label"> Induction Assessment </label>
            </div>
            <s-toggle
              :readOnly="readOnly"
              v-model="item.AssesmentStatus"
              class="w-[70px] mt-0.5"
              yes-label="Yess"
              no-label="No"
              @change="
                onFormFieldChange(
                  'AssesmentStatus',
                  item.AssesmentStatus,
                  '',
                  '',
                  item
                )
              "
            />
          </div>
        </template>
        <template #item_buttons_1="{ item }">
          <action-attachment
            :readOnly="readOnly"
            kind="assesment"
            :ref-id="data.record._id"
            :tags="buildTags(item, 'assesment')"
            :tags-for-get="[`assesment_${item._id}`]"
          />
        </template>
        <template #item_Attendee="{ item, idx }">
          <s-input
            :readOnly="readOnly"
            ref="refName"
            v-model="item.Attendee"
            :disabled="true"
            :readonly="false"
            class="w-full"
            use-list
            allowAdd
            lookup-url="/tenant/employee/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
          ></s-input>
        </template>
        <template #item_Result="{ item, idx }">
          <div class="flex gap-2">
            <div>
              <label class="input_label"> Passed? </label>
            </div>
            <s-toggle
              :readOnly="readOnly"
              v-model="item.Result"
              class="w-[120px] mt-0.5"
              yes-label="Yes"
              no-label="No"
              @change="onFormFieldChange('Result', item.Result, '', '', item)"
            />
          </div>
        </template>
      </s-grid>
    </template>
    <template #form_tab_Attachments="{ item }">
      <s-grid-attachment
        :readOnly="readOnly"
        :key="data.record._id"
        :journal-id="data.record._id"
        :tags="linesTag"
        journal-type="Induction"
        ref="gridAttachment"
        @pre-Save="preSaveAttachment"
      ></s-grid-attachment>
    </template>
    <template #grid_Status="{ item }">
      <status-text :txt="item.Status" />
    </template>
    <template #grid_Dimension="{ item }">
      <DimensionText :dimension="item.Dimension" />
    </template>
    <template #form_input_Dimension="{ item, mode }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :read-only="mode == 'view'"
      ></dimension-editor-vertical>
    </template>
    <template #form_buttons_1="{ item, inSubmission, loading }">
      <form-buttons-trx
        :disabled="loading"
        :status="item.Status"
        :journal-id="item._id"
        :posting-profile-id="item.PostingProfileID"
        :journal-type-id="'INDUCTION'"
        :moduleid="'she'"
        @preSubmit="trxPreSubmit"
        @postSubmit="trxPostSubmit"
        @errorSubmit="trxErrorSubmit"
        :auto-post="false"
      />
    </template>
    <template #grid_item_buttons_1="{ item }">
      <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
    </template>
  </data-list>
</template>
<script setup>
import { reactive, ref, inject, onMounted, computed, watch } from "vue";
import {
  DataList,
  SInput,
  SForm,
  SGrid,
  loadGridConfig,
  loadFormConfig,
  util,
} from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import ActionAttachment from "@/components/common/ActionAttachment.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import SInputSkuItem from "../scm/widget/SInputSkuItem.vue";
import SToggle from "@/components/common/SButtonToggle.vue";
import helper from "@/scripts/helper.js";
import moment from "moment";
import LogTrx from "@/components/common/LogTrx.vue";

layoutStore().name = "tenant";
const listControl = ref(null);
const gridMaterials = ref(null);
const gridAssesment = ref(null);
const gridAttendee = ref(null);
const gridAttachment = ref(null);
const axios = inject("axios");

const linesTag = computed({
  get() {
    const tags = [data.record._id];
    const tagsMaterials = JSON.parse(JSON.stringify(data.record.Materials)).map(
      (a) => {
        return `materials_${a._id}`;
      }
    );
    const tagsAssesment = JSON.parse(JSON.stringify(data.record.Assesment)).map(
      (a) => {
        return `assesment_${a._id}`;
      }
    );
    return [...tags, ...tagsMaterials, ...tagsAssesment];
  },
});
let currentTab = computed(() => {
  if (listControl.value == null) {
    return 0;
  } else {
    return listControl.value.getFormCurrentTab();
  }
});
let customFilter = computed(() => {
  const filters = [];
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "Created",
      Op: "$gte",
      Value: helper.formatFilterDate(data.search.DateFrom),
    });
  }
  if (
    data.search.DateTo !== null &&
    data.search.DateTo !== "" &&
    data.search.DateTo !== "Invalid date"
  ) {
    filters.push({
      Field: "Created",
      Op: "$lte",
      Value: helper.formatFilterDate(data.search.DateTo, true),
    });
  }

  if (data.search.Category !== null && data.search.Category !== "") {
    filters.push({
      Field: "Category",
      Op: "$eq",
      Value: data.search.Category,
    });
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
      ],
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
  titleForm: "Induction",
  record: {},
  cfgGridAttendee: {},
  cfgGridMaterials: {},
  cfgGridAssesment: {},
  formCfg: {},
  listJurnalType: [],
  search: {
    DateFrom: null,
    DateTo: null,
    Status: "",
    Category: "",
    Name: "",
  },
});
const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
function newData(r) {
  r.InductionDate = new Date();
  r.Attendee = [];
  r.Materials = [];
  r.Assesment = [];
  r.Attachments = [];
  if (data.listJurnalType.length > 0) {
    r.JournalTypeID = data.listJurnalType[0]._id;
    r.PostingProfileID = data.listJurnalType[0].PostingProfileID;
  }
  r.Status = "";
  data.record = r;
  data.titleForm = "Create New Induction";
  openForm();
}

function editRecord(r) {
  data.record = r;
  data.titleForm = `Edit Induction | ${r._id}`;
  util.nextTickN(2, () => {
    gridAttendee.value.setRecords(r.Attendee);
    gridMaterials.value.setRecords(r.Materials);
    gridAssesment.value.setRecords(r.Assesment);
    if (readOnly.value) {
      listControl.value.setFormMode("view");
    }
  });
}

function openForm(r) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("Name", "required", true);
  });
}

function onPreSave(record) {}
function onPostSave(record) {
  if (gridAttachment.value) {
    gridAttachment.value.Save();
  }
}
function newAttendee() {
  let r = {};
  const noLine = data.record.Attendee.length + 1;
  r._id = util.uuid();
  r.NoLine = noLine;
  r.EmployeeType = true;
  r.Name = "";
  r.Position = "";
  r.Presence = "";
  data.record.Attendee.push(r);
  updateGridLine(data.record.Attendee, "Attendee");
}
function deleteAttendee(r) {
  data.record.Attendee = data.record.Attendee.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.Attendee, "Attendee");
}
function newAssesment() {
  let r = {};
  const noLine = data.record.Assesment.length + 1;
  r._id = util.uuid();
  r.NoLine = noLine;
  r.Name = "";
  r.Position = "";
  r.Presence = "";
  data.record.Assesment.push(r);
  updateGridLine(data.record.Assesment, "Assesment");
}
function deleteAssesment(r) {
  data.record.Assesment = data.record.Assesment.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.Assesment, "Assesment");
}
function newMaterials() {
  let r = {};
  const noLine = data.record.Materials.length + 1;
  r._id = util.uuid();
  r.NoLine = noLine;
  r.Name = "";
  r.Position = "";
  r.Presence = "";
  data.record.Materials.push(r);
  updateGridLine(data.record.Materials, "Materials");
}
function deleteMaterials(r) {
  data.record.Materials = data.record.Materials.filter(
    (obj) => obj._id !== r._id
  );
  updateGridLine(data.record.Materials, "Materials");
}
function updateGridLine(record, type) {
  record.map((obj, idx) => {
    obj.NoLine = parseInt(idx) + 1;
    return obj;
  });
  if (type == "Attendee") {
    gridAttendee.value.setRecords(record);
  } else if (type == "Assesment") {
    gridAssesment.value.setRecords(record);
  } else {
    gridMaterials.value.setRecords(record);
  }
}
function buildTags(item, type) {
  let tags = [`${type}_${item._id}`];
  return tags;
}
function onFormFieldChange(field, v1, v2, old, record) {
  switch (field) {
    case "AssesmentStatus":
      if (record.AssesmentStatus) {
        record.Assesment = JSON.parse(JSON.stringify(record.Attendee)).map(
          (a) => {
            return {
              _id: a._id,
              NoLine: a.NoLine,
              Attendee: a.Name,
              Result: "",
              Attachment: "",
              Feedback: "",
            };
          }
        );

        gridAssesment.value.setRecords(record.Assesment);
      } else {
        record.Assesment = [];
        gridAssesment.value.setRecords([]);
      }
      break;
    case "AttendeeName":
      util.nextTickN(2, () => {
        axios.post("/bagong/employee/get", [record.Name]).then(
          (r) => {
            let emp = r.data;
            record.Position = emp.Detail.Position;
          },
          (e) => util.showError(e)
        );
      });
      break;
    case "JournalTypeID":
      util.nextTickN(2, () => {
        getJournalType(v1, record);
      });
      break;
    default:
      break;
  }
}

function preSaveAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [data.record._id];
    return asset;
  });
}

function alterGridConfig(cfg) {
  cfg.sortable = ["Created", "_id"];
  cfg.setting.idField = "Created";
  cfg.setting.sortable = ["Created", "_id"];
}
function loadGridMedicines() {
  let url = `/she/induction/attendee/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridAttendee = r;
    },
    (e) => {}
  );
}

function loadGridMaterials() {
  let url = `/she/induction/materials/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      data.cfgGridMaterials = r;
    },
    (e) => {}
  );
}

function loadGridAssesment() {
  let url = `/she/induction/assesment/gridconfig`;
  loadGridConfig(axios, url).then(
    (r) => {
      const fields = r.fields.filter((e) => e.field != "Attachment");
      data.cfgGridAssesment = { ...r, fields };
    },
    (e) => {}
  );
}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = `Induction`;
  }
}
function trxPreSubmit(status, action, doSubmit) {
  if (["DRAFT"].includes(status)) {
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
    doSubmit();
  }
}
function trxPostSubmit(data, action) {
  setLoadingForm(false);
  listControl.value.setControlMode("grid");
  listControl.value.refreshGrid();
}
function trxErrorSubmit() {
  setLoadingForm(false);
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function getJournalType(v1, record) {
  if (v1) {
    axios.post("/fico/shejournaltype/get", [v1]).then(
      (r) => {
        record.PostingProfileID = r.data.PostingProfileID;
      },
      (e) => util.showError(e)
    );
  }
}

function getsJournalType() {
  axios.post("/fico/shejournaltype/gets", {}).then(
    (r) => {
      data.listJurnalType = r.data.data;
    },
    (e) => util.showError(e)
  );
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
// watch(
//   () => currentTab.value,
//   (nv) => {
//     if (nv == 3) {
//       // let record = listControl.value.getFormRecord();
//       // onFormFieldChange(
//       //   "AssesmentStatus",
//       //   record.AssesmentStatus,
//       //   "",
//       //   "",
//       //   record
//       // );
//     }
//   }
// );

onMounted(() => {
  loadGridMedicines();
  loadGridMaterials();
  loadGridAssesment();
  getsJournalType();
});
</script>
<style lang="css" scoped></style>
