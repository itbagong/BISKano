<template>
  <div class="min-w-[55px] h-[30px]" v-if="data.loading">
    <loader kind="skeleton" skeleton-kind="input" />
  </div>
  <template v-else>
    <div
      v-if="data.showButton && journalId != '' && postingProfileId != ''"
      class="flex gap-1"
    >
      <s-button
        :disabled="disabled"
        class="btn_primary submit_btn"
        label="Submit"
        @click="onSubmit('Submit')"
        v-if="status === 'DRAFT' && data.isNeedApprovel === true"
      />

      <s-button
        :disabled="disabled"
        class="btn_primary submit_btn"
        label="Approve"
        @click="onSubmit('Approve')"
        v-else-if="
          status === 'SUBMITTED' &&
          data.isNeedApprovel === true &&
          data.isApprover === true
        "
      />
      <s-button
        :disabled="disabled"
        class="btn_primary submit_btn"
        label="Post"
        @click="onSubmit('Post')"
        v-else-if="
          (data.isPostinger === true && status === 'READY') ||
          (status === 'DRAFT' && data.isNeedApprovel === false)
        "
      />
      <s-button
        :disabled="disabled"
        class="btn_primary submit_btn"
        label="Reject"
        @click="showModalReject"
        v-if="
          showPost ||
          (status === 'SUBMITTED' &&
            data.isNeedApprovel === true &&
            data.isApprover === true)
        "
      />
      <template v-if="status === 'REJECTED'">
        <s-button
          v-if="
            ![
              'Site Entry Income',
              'SiteEntry Income',
              'Site Entry Expense',
            ].includes(trxType)
          "
          :disabled="disabled"
          class="btn_reopen submit_btn w-[80px]"
          label="Re-open"
          @click="onReopen()"
        />
      </template>
    </div>
  </template>
  <s-modal
    title="Reject"
    class="model-reject"
    hide-buttons
    :display="data.modalReject"
    ref="reject"
    @beforeHide="data.modalReject = false"
  >
    <div class="py-3 mb-3">
      <s-input
        ref="reasonReject"
        v-model="data.txtReject"
        label="Reason for Rejection"
        class="w-full"
        multiRow="5"
        :required="true"
        :rules="rulesTxtReject"
        :keepErrorSection="true"
      ></s-input>
    </div>
    <div class="mt-5">
      <s-button
        class="bg-primary text-white font-bold w-full flex justify-center"
        label="Submit"
        @click="onSubmit('Reject')"
      ></s-button>
    </div>
  </s-modal>
</template>
<script setup>
import { reactive, computed, inject, watch, onMounted } from "vue";

import { SButton, util, SModal, SInput } from "suimjs";

import Loader from "@/components/common/Loader.vue";

const props = defineProps({
  disabled: { type: Boolean, default: false },
  status: { type: String, default: "" },
  postingProfileId: { type: String, default: "" },
  journalId: { type: String, default: "" },
  journalTypeId: { type: String, default: "" },
  postUrl: { type: String, default: "" },
  moduleid: { type: String, default: "fico" },
  autoPost: { type: Boolean, default: true },
  autoReopen: { type: Boolean, default: true },
  trxType: { type: String, default: "" },
});
const emit = defineEmits({
  preSubmit: null,
  postSubmit: null,
  errorSubmit: null,
});

const axios = inject("axios");
const data = reactive({
  showButton: true,
  isApprover: false,
  isPostinger: false,
  isNeedApprovel: null,
  loading: false,
  modalReject: false,
  txtReject: "",
});
// const showSubmit = computed({
//   get(){
//     return props.status === 'DRAFT' && data.isNeedApprovel === true
//   }
// })
const showApprove = computed({
  get() {
    return (
      props.status === "SUBMITTED" &&
      data.isNeedApprovel === true &&
      data.isApprover === true
    );
  },
});
// const showApprove = computed({
//   get(){}
// })
const showPost = computed({
  get() {
    return (
      (data.isPostinger === true && props.status === "READY") ||
      (props.status === "DRAFT" && data.isNeedApprovel === false)
    );
  },
});
// const showReopen = computed({
//   get(){}
// })

const showReject = computed({
  get() {
    const showPostAlt =
      (data.isPostinger === true && props.status === "READY") ||
      (props.status === "DRAFT" && data.isNeedApprovel === false);
    const showApproveAlt =
      props.status === "SUBMITTED" &&
      data.isNeedApprovel === true &&
      data.isApprover === true;
    return showPostAlt || showApproveAlt;
  },
});

const rulesTxtReject = [
  (v) => {
    return v.length == 0 ? "reason for rejection is required" : "";
  },
];
function showModalReject() {
  data.txtReject = "";
  data.modalReject = true;
}
async function onSubmit(action) {
  emit("preSubmit", props.status, action, () => {
    if (props.status === "DRAFT") return doSubmit(action);
    else if (!props.autoPost) return doSubmit(action);
    else return null;
  });

  if (props.status !== "DRAFT" && props.autoPost) doSubmit(action);
}

async function onReopen() {
  emit("preReopen", props.status, () => {
    if (!props.autoReopen) {
      return doReopen();
    }
  });

  if (props.autoReopen) {
    return doReopen();
  }
}

function doSubmit(action) {
  data.loading = true;

  const param = {
    JournalType: props.journalTypeId,
    JournalID: props.journalId,
    Op: action,
    text: action == "Reject" ? data.txtReject : "",
  };
  // props.moduleid !== 'bagong' ? [param] : param
  let url = "" + props.moduleid + "/postingprofile/post";
  let paramBuilder = ["she", "bagong"].includes(props.moduleid)
    ? param
    : [param];

  if (props.postUrl) {
    url = props.postUrl;
  }
  axios
    .post(url, paramBuilder)
    .then(
      (r) => {
        emit("postSubmit", r.data, action);
      },
      (e) => {
        emit("errorSubmit", e, action);
        util.showError(e);
      }
    )
    .finally(() => {
      data.loading = false;
    });
}

function doReopen(action) {
  data.loading = true;

  const param = {
    JournalType: props.journalTypeId,
    JournalID: props.journalId,
  };
  axios
    .post("" + props.moduleid + "/postingprofile/reopen", [param])
    .then(
      (r) => {
        emit("postSubmit", r.data, action);
      },
      (e) => {
        emit("errorSubmit", e, action);
        util.showError(e);
      }
    )
    .finally(() => {
      data.loading = false;
    });
}

function fetchPostingProfile() {
  data.isNeedApprovel = null;

  if (props.postingProfileId == "" || props.postingProfileId == undefined)
    return;
  // if(props.status !== "DRAFT") return

  data.loading = true;
  axios
    .post("/fico/postingprofile/get", [props.postingProfileId])
    .then(
      (r) => {
        data.isNeedApprovel = r.data.NeedApproval;
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      data.loading = false;
    });
}
function fetchApproveSource() {
  data.isApprover = false;
  data.isPostinger = false;

  if (["SUBMITTED", "READY"].includes(props.status)) {
    data.loading = true;
    axios
      .post("/fico/postingprofile/get-approval-by-source-user", {
        JournalType: props.journalTypeId,
        JournalID: props.journalId,
      })
      .then(
        (r) => {
          data.isApprover = r.data.Approvers;
          data.isPostinger = r.data.Postingers;
          emit("approvalSource", r.data);
        },
        (e) => util.showError(e)
      )
      .finally(() => {
        data.loading = false;
      });
  }
}
defineExpose({
  submit: onSubmit,
});

watch(
  () => props.postingProfileId,
  (nv) => {
    fetchPostingProfile();
  }
);

onMounted(() => {
  console.log("test");
  fetchPostingProfile();
  fetchApproveSource();
});
</script>
<style>
.model-reject {
}
</style>
