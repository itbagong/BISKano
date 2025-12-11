<template>
  <div v-if="data.loading" class="h-[300px] flex items-center justify-center">
    <Loader kind="circle" />
  </div>
  <s-form
    v-else-if="data.appMode != 'preview'"
    v-model="data.record"
    :config="data.formCfg"
    :mode="'view'"
    :hide-submit="readOnly"
    keep-label
    hide-cancel
    :tabs="['Assesment', 'Psikotes', 'Interview']"
    ref="formControl"
    @submitForm="save"
  >
    <template #buttons_1="{ item, config, inSubmission }">
      <div class="flex flex-row gap-2">
        <form-buttons-trx
          :disabled="inSubmission || loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="data.jType"
          moduleid="hcm"
          :auto-post="!waitTrxSubmit"
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
      </div>
    </template>
    <template #input_Assesment="{ item }">
      <AssesmentTabAssesment
        v-model:journalTypeID="item.JournalTypeID"
        v-model="item.Assesment"
        v-model:status="item.Status"
        :read-only="readOnly"
        @journalTypeChange="(v1) => getJurnalType(v1, item)"
      ></AssesmentTabAssesment>
    </template>
    <template #tab_Psikotes="{ item }">
      <AssesmentTabPsikotes
        :generalRecords="item.PsychoTests"
        :talentDevelopmentAssesmentID="item._id"
        :read-only="readOnly"
      ></AssesmentTabPsikotes>
    </template>
    <template #tab_Interview="{ item }">
      <AssesmentTabInterview v-model="item.Interview" :read-only="readOnly"></AssesmentTabInterview>
    </template>
  </s-form>
  <PreviewReport
    v-else-if="data.appMode == 'preview'"
    class="card w-full"
    title="Preview"
    @close="closePreview"
    :disable-print="helper.isDisablePrintPreview(data.record.Status)"
    :SourceType="data.jType"
    :SourceJournalID="data.record._id"
    :VoucherNo="data.record.LedgerVoucherNo"
    reload="1"
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
      </div>
    </template>
  </PreviewReport>
</template>
<script setup>
import { reactive, onMounted, inject, ref, computed } from "vue";
import { SForm, loadFormConfig, util, SButton } from "suimjs";
import AssesmentTabAssesment from "./AssesmentTabAssesment.vue";
import AssesmentTabPsikotes from "./AssesmentTabPsikotes.vue";
import AssesmentTabInterview from "./AssesmentTabInterview.vue";
import Loader from "@/components/common/Loader.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import helper from "@/scripts/helper.js";
import PreviewReport from "@/components/common/PreviewReport.vue";

const axios = inject("axios");
const formControl = ref(null);

const props = defineProps({
  talentDevelopmentID: { type: String, default: "" },
});
const data = reactive({
  loading: false,
  appMode: "form",
  talentDevelopmentID: props.talentDevelopmentID,
  formCfg: {},
  record: {
    TalentDevelopmentID: props.talentDevelopmentID,
    Status: "DRAFT",
    Assesment: {
      ItemDetails: [],
    },
    Interview: {},
  },
  journalType: {},
  jType: "ASSESSMENT",
});
function closePreview() {
  data.appMode = "form";
}
function preSubmit(record) {}
function save(record, cb) {
  if (record.Status === "") {
    record.Status = "DRAFT";
  }
  axios.post("/hcm/talentdevelopmentassesment/save", record).then(
    (r) => {
      let _record = r.data;
      data.record = _record;
      cb();
    },
    (e) => {
      util.showError(e);
      setLoadingForm(false);
    }
  );
}
const getsData = () => {
  data.loading = true;
  axios
    .post(
      `/hcm/talentdevelopmentassesment/find?TalentDevelopmentID=${data.talentDevelopmentID}`,
      { Take: 1 }
    )
    .then(
      (r) => {
        if (r.data.length > 0) {
          data.record = r.data[0];
        }
        loadFormConfig(
          axios,
          "/hcm/talentdevelopmentassesment/formconfig"
        ).then(
          (r) => {
            data.formCfg = r;
            util.nextTickN(2, () => {});
          },
          (e) => util.showError(e)
        );
      },
      (e) => {
        util.showError(e);
      }
    )
    .finally(() => (data.loading = false));
  // axios.post("/hcm/employee/get", [data.employeeID]).then(
  //   (r) => {
  //     data.record = {
  //       ...r.data.Detail,
  //       Name: r.data.Name,
  //       NIK: r.data.Detail.IdentityCardNo,
  //       Group: r.data.Detail.IdentityCardNoGroup,
  //       Site: r.data.Dimension.find((o) => o.Key === "Site")?.Value,
  //     };
  //     console.log(data.record)

  //   },
  //   (e) => {
  //     util.showError(e);
  //   }
  // );
};
const readOnly = computed({
  get() {
    return !["", "DRAFT"].includes(data.record.Status);
  },
});
function setLoadingForm(loading) {
  formControl.value.setLoading(loading);
}

// function closePreview() {
//   data.appMode = "grid";
// }

function getJurnalType(id) {
  if (id === "" || id === null) {
    data.journalType = {};
    data.record.PostingProfileID = "";
    return;
  }
  formControl.value.setLoading(true);
  axios
    .post("/hcm/journaltype/get", [id])
    .then(
      (r) => {
        data.journalType = r.data;
        data.record.PostingProfileID = r.data.PostingProfileID;
        data.checklistTemplateID = r.data.ChecklistTemplate;
      },
      (e) => {
        data.journalType = {};
        data.record.PostingProfileID = "";
        util.showError(e);
      }
    )
    .finally(() => {
      formControl.value.setLoading(false);
    });
}
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "SUBMITTED", "READY"].includes(data.record.Status);
  },
});

function setFormRequired(required) {
  formControl.value.getAllField().forEach((e) => {
    formControl.value.setFieldAttr(e.field, "required", required);
  });
}
function trxPreSubmit(status, action, doSubmit) {
  if (waitTrxSubmit.value) {
    formControl.value.setCurrentTab(0);
    trxSubmit(doSubmit);
  }
}
function trxSubmit(doSubmit) {
  setFormRequired(true);
  util.nextTickN(2, () => {
    const valid = formControl.value.validate();
    if (valid) {
      setLoadingForm(true);
      save(data.record, () => {
        doSubmit();
      });
    }
    setFormRequired(false);
  });
}
function trxPostSubmit(record, action) {
  setLoadingForm(false);
  // closePreview();
  getsData();
}
function trxErrorSubmit(e) {
  setLoadingForm(false);
}
onMounted(() => {
  getsData();
});
</script>
<style></style>
