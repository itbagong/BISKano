<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Coaching"
      grid-config="/she/coaching/gridconfig"
      form-config="/she/coaching/formconfig"
      grid-read="/she/coaching/gets"
      form-read="/she/coaching/get"
      grid-mode="grid"
      grid-delete="/she/coaching/delete"
      form-keep-label
      form-insert="/she/coaching/save"
      form-update="/she/coaching/save"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      grid-hide-select
      @form-new-data="newRecord"
      @form-edit-data="openForm"
      :form-fields="[
        'Dimension',
        'Date',
        'CoachTitle',
        'CoacheeTitle',
        'RequestorID',
        'EmployeeID',
      ]"
      :grid-fields="['Status', 'Feedback']"
      :grid-custom-filter="customFilter"
      @pre-save="onPreSave"
      @form-field-change="onFormFieldChange"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      stayOnFormAfterSave
    >
      <template #grid_header_search="{ config }">
        <div
          class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter"
        >
          <s-input
            ref="refResponse"
            v-model="data.search.No"
            lookup-key="_id"
            label="No."
            class="w-[200px]"
            @keyup.enter="refreshData"
          ></s-input>
          <s-input
            ref="refCoach"
            v-model="data.search.Coach"
            class="w-[300px]"
            use-list
            label="Coach"
            lookup-url="/tenant/employee/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refCoachee"
            v-model="data.search.Coachee"
            class="w-[300px]"
            use-list
            label="Coachee"
            lookup-url="/tenant/employee/find"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            @change="refreshData"
          ></s-input>
          <s-input
            ref="refTopic"
            v-model="data.search.Topic"
            lookup-key="_id"
            label="Topic"
            class="w-[200px]"
            @keyup.enter="refreshData"
          ></s-input>
          <s-input
            ref="refTarget"
            v-model="data.search.Target"
            lookup-key="_id"
            label="Target"
            class="w-[200px]"
            @keyup.enter="refreshData"
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
      <template #form_input_RequestorID="{ item }">
        <s-input
          v-if="readOnly"
          class="w-full"
          label="Requestor ID"
          use-list
          lookup-url="/tenant/employee/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookupSearchs="['Name']"
          v-model="item.RequestorID"
          :read-only="readOnly || mode == 'view'"
        />
      </template>
      <template #form_input_EmployeeID="{ item }">
        <s-input
          v-if="readOnly"
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
      <template #form_input_Dimension="{ item }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :read-only="
            ['SUBMITTED', 'READY', 'POSTED', 'REJECTED'].includes(item.Status)
          "
          :default-list="profile.Dimension"
        ></dimension-editor-vertical>
      </template>
      <template #form_input_Date="{ item }">
        <label class="input_label">
          <div>Date</div>
        </label>
        <div
          class=""
          v-if="
            ['SUBMITTED', 'READY', 'POSTED', 'REJECTED'].includes(item.Status)
          "
        >
          {{ moment(item.Date).format("DD MMM YYYY HH:mm") }}
        </div>
        <div class="flex gap-2" v-else>
          <s-input kind="date" v-model="data.records.bookDate" />
          <s-input kind="time" v-model="data.records.bookTime" />
        </div>
      </template>
      <template #form_input_CoachTitle="{ item }">
        <s-input
          v-model="item.CoachTitle"
          label="Coach Title"
          use-list
          :lookup-url="'/tenant/masterdata/find?MasterDataTypeID=PTE'"
          lookup-key="_id"
          :lookup-labels="['Name']"
          read-only
          :key="item.CoachTitle"
        />
      </template>
      <template #form_input_CoacheeTitle="{ item }">
        <s-input
          v-model="item.CoacheeTitle"
          label="Coachee Title"
          use-list
          :lookup-url="'/tenant/masterdata/find?MasterDataTypeID=PTE'"
          lookup-key="_id"
          :lookup-labels="['Name']"
          read-only
          :key="item.CoacheeTitle"
        />
      </template>
      <template #grid_Feedback="{ item }">
        <s-button
          v-if="['SUBMITTED'].includes(item.Status)"
          class="btn_success"
          label="Feedback"
          @click="onFeedback(item)"
        />
      </template>
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #form_buttons_1="{ item, inSubmission, loading }">
        <s-button
          v-if="['SUBMITTED'].includes(item.Status)"
          :disabled="false"
          class="btn_primary submit_btn"
          label="Submit"
          @click="onSubmit('Approve')"
        />
        <form-buttons-trx
          v-if="['DRAFT', 'READY'].includes(item.Status)"
          :disabled="loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="'COACHING'"
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
  </div>
</template>
<script setup>
import {
  reactive,
  ref,
  inject,
  watch,
  nextTick,
  onMounted,
  computed,
} from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SInput, util, SButton } from "suimjs";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import moment from "moment";
import StatusText from "@/components/common/StatusText.vue";
import helper from "@/scripts/helper.js";
import { authStore } from "@/stores/auth.js";
import LogTrx from "@/components/common/LogTrx.vue";

layoutStore().name = "tenant";
const FEATUREID = "Coaching";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const listControl = ref(null);
let customFilter = computed(() => {
  const filters = [];
  if (data.search.No !== null && data.search.No !== "") {
    filters.push({
      Field: "_id",
      Op: "$contains",
      Value: [data.search.No],
    });
  }
  if (data.search.Coach !== null && data.search.Coach !== "") {
    filters.push({
      Field: "Coach",
      Op: "$eq",
      Value: data.search.Coach,
    });
  }
  if (data.search.Coachee !== null && data.search.Coachee !== "") {
    filters.push({
      Field: "Coachee",
      Op: "$eq",
      Value: data.search.Coachee,
    });
  }
  if (data.search.Topic !== null && data.search.Topic !== "") {
    filters.push({
      Field: "Topic",
      Op: "$contains",
      Value: [data.search.Topic],
    });
  }
  if (data.search.Target !== null && data.search.Target !== "") {
    filters.push({
      Field: "Target",
      Op: "$contains",
      Value: [data.search.Target],
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
  records: {},
  isFeedBack: false,
  listJurnalType: [],
  search: {
    No: "",
    Coach: "",
    Coachee: "",
    Topic: "",
    Target: "",
    Status: "",
  },
});

function resetBookDate(params) {
  data.records.bookDate = new Date();
  data.records.bookTime = new Date();
}

function newRecord(r) {
  resetBookDate();
  data.isFeedBack = false;
  util.nextTickN(2, () => {
    setReadOnlyAll(false);
    hideFeedback(true);
  });
  if (data.listJurnalType.length > 0) {
    r.JournalTypeID = data.listJurnalType[0]._id;
    r.PostingProfileID = data.listJurnalType[0].PostingProfileID;
  }
  data.records = r;
}

function openForm(r) {
  r.bookDate = moment(r.Date);
  r.bookTime = moment(r.Date).format("HH:mm");
  data.isFeedBack = false;
  util.nextTickN(2, () => {
    if (["SUBMITTED", "READY", "POSTED", "REJECTED"].includes(r.Status)) {
      setReadOnlyAll(true);
    } else {
      setReadOnlyAll(false);
    }

    if (["DRAFT", "SUBMITTED"].includes(r.Status)) {
      hideFeedback(true);
    } else {
      hideFeedback(false);
    }
  });
  data.records = r;
}

function onPreSave(r) {
  let bookDate = moment(data.records.bookDate).format("YYYY-MM-DD");
  let bookTime = data.records.bookTime
    ? data.records.bookTime + ":00"
    : moment(data.records.bookDate).format("HH:mm") + ":00";
  let booking = moment(bookDate + " " + bookTime).format();
  r.Date = booking;
}

function onFormFieldChange(name, v1, v2, old, record) {
  if (name == "Coachee" || name == "Coach") {
    getEmployeeID(name, v1);
  } else if (name == "JournalTypeID") {
    getJournalType(v1, record);
  }
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

function getEmployeeID(field, id) {
  const url = "/bagong/employeedetail/find?EmployeeID=" + id;
  axios.post(url).then(
    (r) => {
      if (r.data.length > 0) {
        data.records[field + "Title"] = r.data[0].Position;
      }
    },
    (e) => {
      util.showError(e.error);
    }
  );
}

function onFeedback(r) {
  data.isFeedBack = true;
  listControl.value.setControlMode("form");
  r.bookDate = moment(r.Date);
  r.bookTime = moment(r.Date).format("HH:mm:ss");
  listControl.value.setFormRecord(r);
  data.records = r;

  util.nextTickN(2, () => {
    setReadOnlyAll(true);
    hideFeedback(false);
    listControl.value.setFormFieldAttr("Feedback", "readOnly", false);
  });
}

function setReadOnlyAll(value) {
  const fields = [
    "Coach",
    "Coachee",
    "Topic",
    "Goal",
    "Benefit",
    "Target",
    "ProblemClarification",
    "Improvement",
    "ObstacleIdentification",
    "JournalTypeID",
    "Feedback",
  ];

  for (let i in fields) {
    let o = fields[i];
    listControl.value.setFormFieldAttr(o, "readOnly", value);
  }
}

function hideFeedback(value) {
  listControl.value.setFormFieldAttr("Feedback", "hide", value);
}
function trxPreSubmit(status, action, doSubmit) {
  if (["DRAFT"].includes(status)) {
    listControl.value.submitForm(
      data.records,
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
function onSubmit() {
  const param = {
    JournalType: "COACHING",
    JournalID: data.records._id,
    Op: "Approve",
    text: "",
  };
  setLoadingForm(true);
  listControl.value.submitForm(
    data.records,
    () => {
      axios
        .post("she/postingprofile/post", param)
        .then(
          (r) => {
            setLoadingForm(false);
            listControl.value.setControlMode("grid");
            listControl.value.refreshGrid();
          },
          (e) => {
            util.showError(e);
            setLoadingForm(false);
          }
        )
        .finally(() => {
          setLoadingForm(false);
        });
    },
    () => {
      setLoadingForm(false);
    }
  );
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
onMounted(() => {
  getsJournalType();
});
</script>
