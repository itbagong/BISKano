<template>
  <data-list
    v-show="data.appMode != 'preview'"
    class="card datalist-map-journal-type"
    ref="listControl"
    title="Contract"
    no-gap
    grid-hide-search
    grid-hide-sort
    grid-hide-footer
    grid-hide-delete
    gridHideNew
    grid-no-confirm-delete
    init-app-mode="grid"
    grid-mode="grid"
    form-keep-label
    grid-auto-commit-line
    :form-tabs-edit="data.tabs"
    :form-tabs-view="data.tabs"
    grid-config="/hcm/contract/gridconfig"
    form-config="/hcm/contract/formconfig"
    grid-read="/hcm/contract/get-contracts"
    form-read="/hcm/contract/get"
    grid-delete="/hcm/contract/delete"
    form-insert="/hcm/contract/save"
    form-update="/hcm/contract/save"
    :form-default-mode="data.formMode"
    :grid-custom-filter="customFilter"
    @gridRowFieldChanged="onGridRowFieldChanged"
    @formEditData="editRecord"
    :grid-fields="['Status']"
    :form-fields="[
      'JournalTypeID',
      'EmployeeID',
      'JobTitle',
      'Attendace',
      'ItemTemplateID',
      'ItemDetails',
      'IsProbationEnd',
      'ExtendedExpiredContractDate',
    ]"
    :formHideSubmit="readOnly"
    @alterGridConfig="alterGridConfig"
    @form-field-change="onFormFieldChange"
    stay-on-form-after-save
  >
    <template #grid_header_search>
      <div class="grow flex gap-3 justify-start grid-header-filter">
        <s-input
          class="w-[200px]"
          label="Employee"
          use-list
          lookup-url="/tenant/employee/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="data.filter.EmployeeID"
          @change="refreshGrid"
        />
        <s-input
          class="w-[200px]"
          label="Job Vacancy"
          use-list
          lookup-url="/hcm/manpowerrequest/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="data.filter.JobVacancyID"
          @change="refreshGrid"
        />
        <s-input
          class="min-w-[200px] filter-status"
          label="Status"
          use-list
          multiple
          :items="['DRAFT', 'SUBMITTED', 'READY', 'REJECTED', 'POSTED']"
          v-model="data.filter.Status"
          @change="refreshGrid"
        />
      </div>
    </template>
    <template #grid_item_buttons_1="{ item }">
      <log-trx :id="item._id" v-if="helper.isShowLog(item.Status)" />
    </template>
    <!-- grid status -->
    <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
    <template #form_buttons_1="{ item, inSubmission, loading, mode }">
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
      <template v-if="mode !== 'new'">
        <s-button
          :disabled="inSubmission || loading"
          class="bg-primary text-white font-bold w-full flex justify-center"
          label="Preview"
          @click="data.appMode = 'preview'"
        ></s-button>
      </template>
    </template>
    <template #form_tab_Summary="{ item, mode }"> <div>Summary</div> </template>
    <template #form_input_JournalTypeID="{ item, mode }">
      <s-input
        class="w-full"
        label="Journal type ID"
        use-list
        lookup-url="/hcm/journaltype/find?TransactionType=Contract"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="item.JournalTypeID"
        :read-only="readOnly || mode == 'view'"
      />
    </template>
    <template #form_input_EmployeeID="{ item, mode }">
      <s-input
        class="w-full"
        label="Employee ID"
        use-list
        lookup-url="/tenant/employee/find"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="item.EmployeeID"
        :read-only="readOnly || mode == 'view'"
      />
    </template>
    <template #form_input_JobTitle="{ item, mode }">
      <s-input
        class="w-full"
        label="Job title"
        use-list
        lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTE"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="item.JobTitle"
        :read-only="readOnly || mode == 'view'"
      />
    </template>
    <template #form_input_Attendace="{ item, mode }">
      <div class="suim_area_table">
        <table class="w-full table-auto suim_table">
          <thead class="grid_header">
            <tr class="border-b-[1px] border-slate-500">
              <th>Absent Name</th>
              <th
                v-for="(month, index) in item.Months"
                :key="`header-${index}`"
              >
                {{ month }}
              </th>
              <th>Total</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(category, key) in data.otherCategories"
              :key="`row-${key}`"
            >
              <td class="border border-gray-300 px-4 py-2">{{ category }}</td>
              <td
                v-for="(entry, index) in item.Attendace[key]"
                :key="`entry-${key}-${index}`"
                class="border border-gray-300 px-4"
              >
                <s-input
                  v-model.number="entry.Score"
                  kind="number"
                  class="w-full px-2 py-1"
                  :read-only="readOnly || mode == 'view'"
                />
              </td>
              <td class="border border-gray-300 px-4 py-2 font-semibold">
                {{ computeTotal(data.record.Attendace[key]) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
    <template #form_input_ItemTemplateID="{ item, mode }">
      <s-input
        class="w-[300px]"
        label="Item Template ID"
        use-list
        lookup-url="/she/mcuitemtemplate/find"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="item.ItemTemplateID"
        :read-only="readOnly || mode == 'view'"
        @change="(_, v) => renderItemDetails(v, item)"
      />
    </template>
    <template #form_input_ItemDetails="{ item, mode }">
      <div class="suim_area_table">
        <table class="w-full table-auto suim_table">
          <thead class="grid_header">
            <tr class="border-b-[1px] border-slate-500">
              <th>Aspect</th>
              <th>Max Score</th>
              <th>Achieved Score</th>
            </tr>
          </thead>
          <tbody v-if="data.loading">
            <tr>
              <td colspan="7">
                <div class="h-[300px] flex items-center justify-center">
                  <Loader kind="circle" />
                </div>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr v-for="(detail, key) in item.ItemDetails" :key="`row-${key}`">
              <td class="border border-gray-300 px-4 py-1">
                {{ detail.Aspect }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                {{ detail.MaxScore }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model.number="detail.AchievedScore"
                  :read-only="readOnly || mode == 'view'"
                  kind="number"
                  class="px-2"
                />
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="bg-gray-100">
              <td class="border border-gray-300 px-4 py-1 font-bold">
                Total Score
              </td>
              <td
                class="border border-gray-300 px-4 py-1 text-center font-bold"
              >
                {{ maxScoreTotal }}
              </td>
              <td
                class="border border-gray-300 px-4 py-1 text-center font-bold"
              >
                {{ achievedScoreTotal }}
              </td>
            </tr>
            <tr>
              <td
                class="border border-gray-300 px-4 py-1 font-bold"
                colspan="2"
              >
                Final Score
              </td>
              <td
                class="border border-gray-300 px-4 py-1 text-center font-bold"
              >
                {{ finalScoreTotal }}
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </template>
    <template #form_input_IsProbationEnd="{ item, mode }">
      <div>
        <div v-for="el in data.resultItems" :key="el.value" class="mb-2">
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              type="radio"
              :disabled="readOnly || mode == 'view'"
              :value="el.value"
              v-model="data.resultValue"
              class="form-radio text-blue-600"
              @change="onChangeResult(el, item)"
            />
            <span>{{ el.text }}</span>
          </label>
        </div>
      </div>
    </template>
    <template #form_input_ExtendedExpiredContractDate="{ item }">
      <div v-if="item.IsContractExtended">
        <s-input
          label="Contract Expired Date"
          class="w-full"
          keep-label
          kind="date"
          :read-only="readOnly || mode == 'view'"
          v-model="item.ExtendedExpiredContractDate"
        >
        </s-input>
      </div>
      <template v-else>&nbsp;</template>
    </template>
  </data-list>
  <PreviewReport
    v-if="data.appMode == 'preview'"
    class="card w-full"
    title="Preview"
    @close="closePreview"
    :disable-print="helper.isDisablePrintPreview(data.record.Status)"
    :SourceType="data.jType"
    :SourceJournalID="data.record._id"
    reload=1  
  >
    <template #buttons="props">
      <div class="flex gap-[1px] mr-2">
        <form-buttons-trx
          :disabled="inSubmission || loading"
          :status="data.record.Status"
          :journal-id="data.record._id"
          :posting-profile-id="data.record.PostingProfileID"
          :journal-type-id="data.jType"
          moduleid="hcm"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :auto-post="!waitTrxSubmit"
        />
        <!-- <div>
          <action-attachment
            ref="PreviewAction"
            :kind="data.jType"
            :ref-id="data.record._id"
            :tags="attachmentByTag(linesTag)"
            hide-button
            actionOnHeader
            read-only
          />
        </div> -->
      </div>
    </template>
  </PreviewReport>
</template>
<script setup>
import { reactive, ref, inject, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SButton, SInput } from "suimjs";
import { authStore } from "@/stores/auth";
import { useRoute } from "vue-router";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import StatusText from "@/components/common/StatusText.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import helper from "@/scripts/helper.js";

// import ActionAttachment from "@/components/common/ActionAttachment.vue";

const axios = inject("axios");

layoutStore().name = "tenant";

const FEATUREID = "Probation";
const profile = authStore().getRBAC(FEATUREID);
const auth = authStore();

const route = useRoute();
const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  records: [],
  record: {},
  tabs: ["Assessment"],
  filter: {
    JobVacancyID: "",
    EmployeeID: "",
    Status: [],
  },
  months: ["Month 1", "Month 2", "Month 3", "Month 4", "Month 5", "Month 6"],
  otherCategories: {
    Presence: "Kehadiran",
    Absent: "Alpha/Mangkir",
    Sick: "Sakit",
    Leave: "Izin",
    Late: "Terlambat",
  },
  loading: false,
  resultValue: null,
  resultItems: [
    { text: "Habis Probation/Kontrak", value: 1 },
    {
      text: "Diangkat menjadi karyawan tetap",
      value: 2,
    },
    { text: "Perpanjang Masa Kontrak", value: 3 },
  ],
  jType: "CONTRACT",
  journalType: {},
});
function editRecord(record) {
  if (record.Attendace.Presence.length === 0) {
    const initialData = data.months.map((month) => ({ Name: month, Score: 0 }));
    record.Attendace = {
      Presence: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
      Absent: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
      Sick: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
      Leave: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
      Late: initialData.map((o) => ({ Name: o.Name, Score: 0 })),
    };
  }
  record.Months = [...data.months];
  if (!record.CompanyID) {
    record.CompanyID = auth.appData.CompanyID;
  }
  setResultValue(record);
  openForm(record);
  data.record = record;
}

function openForm(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("PostingProfileID", "readOnly", true);
    if (readOnly.value === true) {
      listControl.value.setFormMode("view");
    }
  });
}
function computeTotal(entries) {
  return entries?.reduce((sum, entry) => sum + entry.Score, 0);
}

function setResultValue(record) {
  if (record.IsProbationEnd) {
    data.resultValue = data.resultItems[0].value;
  } else if (record.IsBecomeEmployee) {
    data.resultValue = data.resultItems[1].value;
  } else if (record.IsContractExtended) {
    data.resultValue = data.resultItems[2].value;
  } else {
    data.resultValue = 0;
  }
}
const customFilter = computed(() => {
  const filters = [];
  if (data.filter.JobVacancyID) {
    filters.push({
      Op: "$eq",
      Field: "JobID",
      Value: data.filter.JobVacancyID,
    });
  }
  if (data.filter.EmployeeID) {
    filters.push({
      Op: "$eq",
      Field: "EmployeeID",
      Value: data.filter.EmployeeID,
    });
  }
  if (data.filter.Status.length > 0) {
    filters.push({
      Op: "$in",
      Field: "Status",
      Value: [...data.filter.Status],
    });
  }
  return filters.length > 1
    ? { Op: "$and", Items: filters }
    : filters[0] || null;
});
function alterGridConfig(cfg) {
  //cfg.fields = cfg.fields.filter(el => el.field != 'Dimension')
  //console.log(cfg)
  const RemainingProbation = {
    field: "RemainingProbation",
    kind: "text",
    label: "Remaining Probation (Day)",
    halign: "start",
    valign: "start",
    labelField: "",
    length: 0,
    width: "",
    pos: 1000,
    readType: "show",
    decimal: 0,
    dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
    unit: "",
    input: {
      field: "RemainingProbation",
      label: "Remaining Probation",
      hint: "",
      hide: false,
      placeHolder: "Remaining Probation",
      kind: "text",
      disable: false,
      required: false,
      multiple: false,
      multiRow: 1,
      minLength: 0,
      maxLength: 999,
      readOnly: false,
      readOnlyOnEdit: false,
      readOnlyOnNew: false,
      useList: false,
      allowAdd: false,
      items: [],
      useLookup: false,
      lookupUrl: "",
      lookupKey: "",
      lookupLabels: null,
      lookupSearchs: null,
      lookupFormat1: "",
      lookupFormat2: "",
      showTitle: false,
      showHint: false,
      showDetail: false,
      fixTitle: false,
      fixDetail: false,
      section: "Assessment",
      sectionWidth: "",
      row: 0,
      col: 0,
      labelField: "",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      width: "",
      spaceBefore: 0,
      spaceAfter: 0,
    },
  };
  cfg.fields.splice(5, 0, RemainingProbation);
}
const maxScoreTotal = computed(() => {
  const result = data.record.ItemDetails.reduce(
    (total, aspect) => total + aspect.MaxScore,
    0
  );
  data.record.MaxScoreTotal = result;
  return result;
});
const achievedScoreTotal = computed(() => {
  const result = data.record.ItemDetails.reduce(
    (total, aspect) => total + aspect.AchievedScore,
    0
  );
  data.record.AchievedScoreTotal = result;
  return result;
});
const finalScoreTotal = computed(() => {
  if (maxScoreTotal.value === 0) {
    return 0;
  }
  const result = parseFloat(
    ((achievedScoreTotal.value / maxScoreTotal.value) * 100).toFixed(2)
  );
  data.record.FinalScore = result;
  return result;
});
function refreshGrid() {
  util.nextTickN(2, () => listControl.value.refreshGrid());
}
function renderItemDetails(v, record) {
  data.loading = true;
  axios
    .post("/she/mcuitemtemplate/get", [v])
    .then((r) => {
      record.ItemDetails = r.data.Lines.map((item) => {
        return {
          Aspect: item.Description,
          MaxScore: item.AnswerValue,
          AchievedScore: 0,
        };
      });
    })
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loading = false;
    });
}
function onChangeResult(v, record) {
  switch (v.value) {
    case 1:
      record.IsProbationEnd = true;
      record.IsBecomeEmployee = false;
      record.IsContractExtended = false;
      break;
    case 2:
      record.IsProbationEnd = false;
      record.IsBecomeEmployee = true;
      record.IsContractExtended = false;
      break;
    case 3:
      record.IsProbationEnd = false;
      record.IsBecomeEmployee = false;
      record.IsContractExtended = true;
      break;
    default:
      break;
  }
}
function getJurnalType(id) {
  if (id === "" || id === null) {
    data.journalType = {};
    data.record.PostingProfileID = "";
    return;
  }
  listControl.value.setFormLoading(true);
  axios
    .post("/hcm/journaltype/get", [id])
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
      listControl.value.setFormLoading(false);
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
</script>
