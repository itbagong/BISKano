<template>
  <div>
    <s-modal
      title="Pesan"
      :display="false"
      ref="confirmModal"
      :hideSubmit="!data.isProccesApproval"
      :hideCancel="!data.isProccesApproval"
      @submit="onSubmitModal"
    >
      <div v-if="data.isProccesApproval" class="my-2">
        <s-input
          label="Note"
          caption="Note Approval"
          kind="text"
          v-model="data.noteApproval"
        ></s-input>
      </div>
      {{ data.message }}
      <!-- You will Approve or Reject the existing data ! Are you sure ?<br />
    Please be noted, this can not be undone ! -->
      <template #buttons_1 v-if="!data.isProccesApproval">
        <s-button
          class="btn_warning"
          label="close"
          @click="onModalClose"
        ></s-button>
      </template>
    </s-modal>
    <div class="w-full card">
      <div
        v-show="data.appMode == 'grid'"
        class="grid grid-cols-1 gap-2 md:grid-cols-2 lg:grid-cols-4"
      >
        <div class="flex justify-start p-2 border rounded bg-indigo-50">
          <div class="w-full flex flex-col gap-4">
            <div class="my-2 flex flex-row">
              <mdicon
                name="calendar-filter-outline"
                class="text-indigo-700 rounded"
              />
              <h2 class="text-indigo-700 ml-3 font-semibold">Filter</h2>
            </div>
            <div class="w-[120px] items-start gap-2 grid gridCol1">
              <div class="flex flex-row gap-2">
                <s-input
                  kind="date"
                  label="Date From"
                  v-model="data.filter.Start"
                ></s-input>
                <s-input
                  kind="date"
                  label="Date To"
                  v-model="data.filter.End"
                ></s-input>
              </div>
            </div>
            <ul class="list-inside hover:list-outside">
              <li class="pb-2 font-semibold">Filter group</li>
              <li
                class="p-2 border-b-2 transition ease-in-out delay-150 hover:-translate-y-1 hover:scale-105 hover:font-semibold duration-300"
                :class="{ 'bg-blue-400': data.isGroupMenuSelected[0] }"
                @click="selectFilter(0, 'Module')"
              >
                <div class="cursor-pointer flex">
                  <!-- <mdicon name="menu-right" class="text-indigo-700 rounded" /> -->
                  Module
                </div>
              </li>
              <li
                class="p-2 border-b-2 transition ease-in-out delay-150 hover:-translate-y-1 hover:scale-105 hover:font-semibold duration-300"
                :class="{ 'bg-blue-400': data.isGroupMenuSelected[1] }"
                @click="selectFilter(1, 'Site')"
              >
                <div class="cursor-pointer flex">
                  <!-- <mdicon name="menu-right" class="text-indigo-700 rounded" /> -->
                  Site
                </div>
              </li>
              <li
                class="p-2 border-b-2 transition ease-in-out delay-150 hover:-translate-y-1 hover:scale-105 hover:font-semibold duration-300"
                :class="{ 'bg-blue-400': data.isGroupMenuSelected[2] }"
                @click="selectFilter(2, 'Object')"
              >
                <div class="cursor-pointer flex">
                  <!-- <mdicon name="menu-right" class="text-indigo-700 rounded"/> -->
                  Object
                </div>
              </li>
              <li class="p-2">
                <s-button
                  class="btn_primary"
                  label="Search"
                  icon="feature-search-outline"
                  @click="onFilter"
                ></s-button>
              </li>
            </ul>
          </div>
        </div>
        <div
          v-if="data.isLoading1"
          class="flex justify-center gap-3 items-center"
          style="min-height: calc(100vh - 300px)"
        >
          <loader kind="circle" />
        </div>
        <!-- show by group -->
        <div
          class="flex justify-center p-2 border rounded"
          v-if="data.isShowGroup"
        >
          <div class="w-full flex flex-col gap-2">
            <div class="w-full items-start gap-2 grid gridCol1">
              <div class="flex gap-2 border-b-4">
                <s-button
                  label="Need Approve"
                  class="flex-auto"
                  :class="{
                    'bg-orange-500': data.isTabSelected[0],
                    'text-white': data.isTabSelected[0],
                  }"
                  @click="onTabSelected(0, 'PENDING')"
                ></s-button>
                <s-button
                  label="Approved"
                  class="flex-auto"
                  :class="{
                    'bg-orange-500': data.isTabSelected[1],
                    'text-white': data.isTabSelected[1],
                  }"
                  @click="onTabSelected(1, 'APPROVED')"
                ></s-button>
                <s-button
                  label="Rejected"
                  class="btn flex-auto"
                  :class="{
                    'bg-orange-500': data.isTabSelected[2],
                    'text-white': data.isTabSelected[2],
                  }"
                  @click="onTabSelected(2, 'REJECTED')"
                ></s-button>
              </div>
            </div>
            <div
              class="w-full items-start gap-2 grid gridCol1 mb-4 approveList"
            >
              <!-- button approve / reject by group -->
              <div
                v-show="data.isNeedApproval"
                class="flex justify-end gap-4 mt-4"
              >
                <s-button
                  label="Approve"
                  class="btn_primary"
                  icon="file-document-check-outline"
                  @click="onApprove('Approve', 'group')"
                ></s-button>
                <s-button
                  label="Reject"
                  class="btn_warning"
                  icon="file-document-remove-outline"
                  @click="onApprove('Reject', 'group')"
                ></s-button>
              </div>
              <div class="p-2 flex justify-between">
                <div v-show="data.isNeedApproval" class="flex-none">
                  <s-input
                    kind="checkbox"
                    v-model="data.checkAll"
                    @change="checkAll"
                    label="select all"
                  ></s-input>
                </div>
              </div>
              <div class="relative overflow-y-auto max-h-[374px]">
                <s-list
                  :key="data.compKeyIn"
                  ref="listInCtl"
                  class="w-full"
                  :config="{ setting: {} }"
                  hide-delete-button
                  hide-new-button
                  hide-sort
                  hide-control
                  show-header
                  hideFooter
                  v-model="data.groupRecords"
                >
                  <template #item="{ item }">
                    <div class="flex">
                      <div v-show="data.isNeedApproval" class="flex-none w-6">
                        <s-input
                          kind="checkbox"
                          v-model="item.isChoise"
                        ></s-input>
                      </div>
                      <div class="grow">
                        <div class="flex justify-between mb-3 items-start">
                          <div class="font-semibold">
                            <!-- <div>{{item}}</div> -->
                            {{ item.Group }} [{{ item.SiteID }}]<br />
                            {{ util.formatMoney(item.Total, {}) }}
                          </div>
                          <div class="text-[0.7rem] flex gap-2 items-center">
                            <!-- {{ moment(item.Date).format("DD-MMM-yyyy") }} -->
                            <button
                              @click="onShowDetail(item)"
                              class="bg-primary text-white w-[40px] h-[24px] flex justify-center items-center"
                            >
                              <mdicon
                                name="chevron-right"
                                class="cursor-pointer"
                              />
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  </template>
                  <template #footer_1>
                    <div class="flex justify-end gap-4">
                      <s-button
                        label="Approved"
                        class="btn_primary"
                        icon="file-document-check-outline"
                        @click="onApprove('Approve', 'group')"
                      ></s-button>
                      <s-button
                        label="Reject"
                        class="btn_warning"
                        icon="file-document-remove-outline"
                        @click="onApprove('Reject', 'group')"
                      ></s-button>
                    </div>
                  </template>
                </s-list>
              </div>
            </div>
          </div>
        </div>
        <div
          v-if="data.isLoading2"
          class="flex justify-center gap-3 items-top m-3"
          style="min-height: calc(100vh - 300px)"
        >
          <loader kind="circle" />
        </div>
        <!-- show detail -->
        <div
          v-if="data.isShowDetail"
          class="flex justify-center p-2 border rounded sm:col-span-2"
        >
          <div class="w-full flex flex-col gap-4">
            <div class="w-full items-start gap-2 grid gridCol1">
              <div class="w-full">
                <div class="grid grid-flow-row auto-rows-max">
                  <div class="flex flex-row gap-2">
                    <div class="w-[100px]">{{ data.filter.GroupBy }}</div>
                    <div>: {{ data.headerDetail.field1 }}</div>
                  </div>
                  <!-- <div class="flex flex-row gap-2">
                    <div class="w-[100px]">Date</div>
                    <div>
                      :
                      {{
                        moment(data.headerDetail.field2).format("DD-MMM-yyyy")
                      }}
                    </div>
                  </div> -->
                  <div class="flex flex-row gap-2">
                    <div class="w-[100px]">Total Amount</div>
                    <div>
                      : {{ util.formatMoney(data.headerDetail.field3, {}) }}
                    </div>
                  </div>
                  <div
                    v-show="data.isNeedApproval"
                    class="flex justify-end gap-4"
                  >
                    <s-button
                      label="Approve"
                      class="btn_primary"
                      icon="file-document-check-outline"
                      @click="onApprove('Approve', 'detail')"
                    ></s-button>
                    <s-button
                      label="Reject"
                      class="btn_warning"
                      icon="file-document-remove-outline"
                      @click="onApprove('Reject', 'detail')"
                    ></s-button>
                  </div>
                </div>
                <!-- using browser -->
                <!-- <div
                class="mt-4 relative overflow-y-auto max-h-[416px] collapse sm:visible md:visible"
              > -->
                <div class="mt-4 relative overflow-y-auto max-h-[416px]">
                  <s-grid
                    title="Data Approve"
                    ref="refGrid"
                    :config="data.gridCfg"
                    :modelValue="data.detailRecords"
                    hideDetail
                    hideDeleteButton
                    hideControl
                    :hideSelect="!data.isNeedApproval"
                    @gridRefreshed="onGridRefreshed"
                  >
                    <template #item_Object="{ item }">
                      <div class="w-[180px]">
                        {{ item.SourceType }}
                      </div>
                    </template>
                    <template #item_TrxDate="{ item }">
                      <div class="w-[140px]">
                        {{ moment(item.TrxDate).format("DD-MMM-yyyy HH:mm") }}
                      </div>
                    </template>
                    <template #item_Text="{ item }">
                      <div class="w-[260px]">
                        {{ item.Text }}
                      </div>
                    </template>
                    <template #item_Amount="{ item }">
                      <div class="w-[130px]">
                        {{ util.formatMoney(item.Amount, {}) }}
                      </div>
                    </template>
                    <template #item_buttons_1="{ item }">
                      <!-- <action-attachment  
                        :kind="`${attchKind}`"
                        :ref-id="attchRefId" 
                        :tags="[`${attchTagPrefix}_EXPENSE_${attchRefId}_${item.ID}`]"
                        :read-only="hasJournalID(item.JournalID)"
                        icon-attach
                      /> -->
                      <s-button
                        class="bg-transparent hover:bg-blue-500 hover:text-black m-2 border"
                        icon="attachment"
                        label=""
                        tooltip="Attachment"
                        @click="
                          () => {
                            data.attach.refId = item.SourceID;
                            data.attach.kind = item.SourceType;
                            const tagPrefix = getPrefix(item.SourceType);
                            data.attach.tags = [
                              `${tagPrefix}_${item.SourceID}`,
                            ];
                            if (tagPrefix == 'VENDOR') {
                              data.attach.tags.push(
                                `${tagPrefix}_${tagPrefix} ${item.SourceID}`
                              );
                            }
                            util.nextTickN(2, () => {
                              data.attach.showContent = true;
                            });
                          }
                        "
                      ></s-button>
                      <s-button
                        class="bg-transparent hover:bg-blue-500 hover:text-black m-2 border"
                        icon="eye-outline"
                        label="Preview"
                        tooltip="Preview"
                        @click="onPreview(item)"
                      ></s-button>
                    </template>
                  </s-grid>
                </div>
                <!-- using mobile -->
                <!-- <div class="itemList sm:collapse">
                <s-list
                  :key="data.SourceID"
                  ref="listDetail"
                  class="w-full"
                  :config="{ setting: {} }"
                  hide-delete-button
                  hide-new-button
                  hide-sort
                  hide-control
                  show-header
                  v-model="data.detailRecords"
                >
                  <template #item="{ item }">
                    <div class="flex flex-col sm:flex-row flex-nowrap">
                      <div
                        class="flex-none w-[220]"
                        v-show="data.isNeedApproval"
                      >
                        <s-input
                          :label="item.Object"
                          kind="checkbox"
                          v-model="item.isSelected"
                        ></s-input>
                      </div>
                      <div class="w-[120px]">
                        <div class="sm:invisible inline-block font-semibold">
                          Date:
                        </div>
                        {{ moment(item.TrxDate).format("DD-MMM-yyyy") }}
                      </div>
                      <div class="w-full items-start">
                        <div class="sm:invisible inline-block font-semibold">
                          Text:
                        </div>
                        {{ item.Text }}
                      </div>
                      <div class="flex- w-18">
                        <div class="sm:invisible inline-block font-semibold">
                          Amount:
                        </div>
                        {{ util.formatMoney(item.Amount, {}) }}
                      </div>
                    </div>
                  </template>
                </s-list>
              </div> -->
              </div>
            </div>
          </div>
        </div>
      </div>
      <PreviewReport
        v-if="data.appMode == 'preview'"
        class="w-full"
        title="Preview"
        :preview="data.preview"
        @close="closePreview"
        disable-print
        :SourceType="data.recordDetail.SourceType"
        :SourceJournalID="data.recordDetail.SourceJournalID"
        :VoucherNo="data.recordDetail.VoucherNo"
      >
        <template #buttons="props">
          <div class="flex gap-[1px] mr-2">
            <form-buttons-trx
              :disabled="inSubmission || loading"
              :status="data.recordDetail.Status"
              :journal-id="data.recordDetail.SourceID"
              :posting-profile-id="data.recordDetail.PostingProfileID"
              :journal-type-id="getJournalType"
              :moduleid="getModuleId"
              :trx-type="data.recordDetail.TransactionType"
              @postSubmit="trxPostSubmit"
              @errorSubmit="trxErrorSubmit"
              :auto-post="!waitTrxSubmit"
            />
            <s-button
              class="btn_primary"
              icon="attachment"
              label="Attachment"
              tooltip="Attachment"
              @click="
                () => {
                  data.attach.refId = data.recordDetail.SourceID;
                  data.attach.kind = data.recordDetail.SourceType;
                  const tagPrefix = getPrefix(data.recordDetail.SourceType);
                  data.attach.tags = [
                    `${tagPrefix}_${data.recordDetail.SourceID}`,
                  ];
                  if (tagPrefix == 'VENDOR') {
                    data.attach.tags.push(
                      `${tagPrefix}_${tagPrefix} ${data.recordDetail.SourceID}`
                    );
                  }
                  util.nextTickN(2, () => {
                    data.attach.showContent = true;
                  });
                }
              "
            ></s-button>
          </div>
        </template>
      </PreviewReport>

      <s-modal
        v-if="data.attach.showContent"
        title="Attachment"
        class="model-reject"
        display
        ref="reject"
        @beforeHide="
          () => {
            data.attach.showContent = false;
          }
        "
        hideButtons
      >
        <div class="min-w-[500px] w-[1200px] max-h-[600px] overflow-auto">
          <s-grid-attachment
            :journalId="data.attach.refId"
            :journalType="data.attach.kind"
            :tags="data.attach.tags"
            ref="gridAttachmentCtl"
            read-only
          />
        </div>
      </s-modal>
    </div>
  </div>
</template>
<script setup>
import { reactive, onMounted, inject, ref, nextTick, computed } from "vue";
import {
  SCard,
  SGrid,
  SForm,
  loadGridConfig,
  loadFormConfig,
  util,
  SButton,
  SModal,
  SInput,
  SList,
} from "suimjs";
import moment from "moment";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { useRouter } from "vue-router";
import Loader from "@/components/common/Loader.vue";
import PreviewReport from "@/components/common/PreviewReport.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";

layoutStore().name = "tenant";

const FEATUREID = "ApprovalAllExpense";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const auth = authStore();
const router = useRouter();
const confirmModal = ref(null);
const listInCtl = ref(null);
const listDetail = ref(null);
const refGrid = ref(null);

const data = reactive({
  appMode: "grid",
  isLoading1: false,
  isLoading2: false,
  isShowGroup: false,
  isShowDetail: false,
  isProccesApproval: false,
  isGroupApproval: false,
  isNeedApproval: true,
  filter: {
    Start: "",
    End: "",
    GroupBy: "",
    Status: "",
  },
  isGroupMenuSelected: [false, false, false],
  isTabSelected: [true, false, false],
  checkAll: false,
  statusExpense: "general",
  groupRecords: [],
  detailRecords: [],
  compKeyIn: 0,
  idList: 0,
  gridCfg: {
    fields: [],
    setting: {
      // idField: "_id",
      // keywordFields: ["_id"],
    },
  },
  message: "",
  headerDetail: {
    field1: "",
    field2: "",
    field3: "",
  },
  operation: "",
  recordGroup: {},
  recordDetail: {},
  tabSelectedStatus: "",
  noteApproval: "",
  attach: {
    refId: "",
    kind: "",
    tags: [],
    showContent: false,
  },
});
function hasJournalID(journalID) {
  return journalID != "" && journalID != undefined;
}
function selectFilter(i, groupBy) {
  data.isGroupMenuSelected.map((_, x) => {
    if (x != i) {
      data.isGroupMenuSelected[x] = false;
    }
  });
  data.isGroupMenuSelected[i] = true;
  data.filter.GroupBy = groupBy;
  // data.isGroupMenuSelected[i] = !data.isGroupMenuSelected[i];
}

function checkAll() {
  data.groupRecords.map((d) => {
    d.isChoise = !data.checkAll;
  });
}

function onTabSelected(i, status) {
  data.tabSelectedStatus = status;
  data.isNeedApproval = i == 0 ? true : false;
  data.isTabSelected.map((_, x) => {
    if (x != i) {
      data.isTabSelected[x] = false;
    }
  });
  data.isTabSelected[i] = true;

  loadDataGroup(status);
}

function onGridRefreshed() {
  util.nextTickN(2, () => {
    refGrid.value.setRecords(data.detailRecords);
  });
}

function generateConfig(r) {
  data.gridCfg.fields = [
    {
      field: "SourceID",
      label: "Id",
      readType: "hide",
      labelField: "",
      input: {
        lookupUrl: "",
      },
    },
    {
      field: "Object",
      label: "Object",
      readType: "show",
      labelField: "",
      input: {
        lookupUrl: "",
      },
    },
    {
      field: "TrxDate",
      label: "Date",
      readType: "show",
      labelField: "",
      input: {
        lookupUrl: "",
      },
    },
    {
      field: "Text",
      label: "Text",
      readType: "show",
      labelField: "",
      input: {
        lookupUrl: "",
      },
    },
    {
      field: "Amount",
      label: "Amount",
      readType: "show",
      labelField: "",
      input: {
        lookupUrl: "",
      },
    },
  ];
}

function generateData(d) {
  util.nextTickN(2, () => {
    setTimeout(() => {
      // data.detailRecords = d;
      // refGrid.value.refreshData();
      refGrid.value.setRecords(data.detailRecords);
    }, 500);
  });
  // refGrid.value.refreshData();
}

function onFilter() {
  data.tabSelectedStatus = "PENDING";
  if (!data.filter.Start || !data.filter.End || !data.filter.GroupBy) {
    data.isProccesApproval = false;
    data.message = "Isi filter harus terisi untuk pencarian data!";
    confirmModal.value.show();
    return false;
  }
  loadDataGroup(data.tabSelectedStatus);
}

function loadDataGroup(status) {
  util.nextTickN(2, () => {
    data.isShowDetail = false;
    data.isShowGroup = false;
    data.isLoading1 = true;
    data.groupRecords = [];

    data.filter.Status = status;
    data.filter.Start = new Date(moment(data.filter.Start).startOf("day"));
    data.filter.End = new Date(moment(data.filter.End).endOf("day"));
    let payload = data.filter;
    const url = "/fico/approvalaggregator/group-by";
    axios
      .post(url, payload)
      .then(
        (r) => {
          r.data.forEach((d) => {
            // console.log("result:", d);
            let dt = {
              Group: d.Group == "" ? "-" : d.Group,
              Date: d.Date,
              Total: d.Total,
              isChoise: false,
              SiteID: d.SiteID == "" ? "-" : d.SiteID,
            };
            data.groupRecords.push(dt);
          });
        },
        (e) => util.showError(e)
      )
      .finally(() => {
        data.isShowGroup = true;
        data.isLoading1 = false;
      });
  });
}

function onShowDetail(d) {
  // console.log("d:", d);
  data.recordGroup = d;
  util.nextTickN(2, () => {
    data.headerDetail = {
      field1: `${d.Group} [${d.SiteID}]`,
      field2: d.Date,
      field3: d.Total,
    };

    data.isLoading2 = true;
    data.isShowDetail = true;
    data.detailRecords = [];

    let payload = {
      GroupBy: data.filter.GroupBy,
      Status: data.tabSelectedStatus,
      Type: data.filter.GroupBy == "Site" ? d.SiteID : d.Group,
      Date: d.Date,
    };

    const url = "/fico/approvalaggregator/get-journal";
    axios
      .post(url, payload)
      .then(
        (r) => {
          if (!r.error) {
            generateConfig(r.data);
            util.nextTickN(2, () => {
              r.data.forEach((d) => {
                d["isSelected"] = false;
                data.detailRecords.push(d);
              });
              data.isShowDetail = true;
            });
          }
        },
        (e) => util.showError(e)
      )
      .finally(() => {
        data.isLoading2 = false;
      });
  });
}

function onApprove(op, type) {
  data.operation = op;
  if (type == "group") {
    data.isGroupApproval = true;
  } else {
    data.isGroupApproval = false;
  }

  data.isProccesApproval = true;
  data.noteApproval = "";
  data.message = `You will ${op} the existing data ! Are you sure ? Please be noted, this can not be undone !`;
  confirmModal.value.show();
}

function onModalClose() {
  confirmModal.value.hide();
}

function onSubmitModal() {
  // console.log("operation:", data.operation);
  if (data.isGroupApproval) {
    operationGroup();
  } else {
    operationDetail();
  }

  confirmModal.value.hide();
}

function operationGroup() {
  let dataProcess = [];
  // console.log("approve:", data.groupRecords);
  dataProcess = data.groupRecords.filter((x) => {
    return x.isChoise == true;
  });
  console.log("dataProcess:", dataProcess);
  if (dataProcess.length == 0) {
    util.showError("No data selected to process!");
    return false;
  }

  let payload = [];
  dataProcess.forEach((r) => {
    payload.push({
      GroupBy: data.filter.GroupBy,
      Type: data.filter.GroupBy === "Site" ? r.SiteID : r.Group,
      Date: r.Date,
      Op: data.operation,
      Text: data.noteApproval,
    });
  });
  // console.log("Payload approve:", payload);
  // const url = "/fico/approvalaggregator/post-by-group";
  const url = "/mfg/approvalaggregator/post-by-group";
  axios
    .post(url, payload)
    .then(
      (r) => {
        util.showInfo(`Data has been ${data.operation}`);
        loadDataGroup(data.tabSelectedStatus);
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      data.isShowGroup = true;
      data.isLoading1 = false;
    });
}

function postDetail(url, payload) {
  return axios.post(url, payload);
}
function operationDetail() {
  let dataProcess = [];
  console.log("data:", data.detailRecords);
  dataProcess = data.detailRecords.filter((x) => {
    return x.isSelected == true;
  });
  // console.log("dataProcess:", dataProcess);
  if (dataProcess.length == 0) {
    util.showError("No data selected to process!");
    return false;
  }

  let payloadFico = [];
  let payloadSCM = [];
  let payloadMFG = [];
  let post = [];

  dataProcess.forEach((item) => {
    if (
      [
        "INVENTORY",
        "Inventory Receive",
        "Inventory Issuance",
        "Transfer",
        "Item Request",
        "Purchase Order",
        "Purchase Request",
        "Asset Acquisition",
      ].includes(item.SourceType)
    ) {
      payloadSCM.push({
        JournalID: item.SourceID,
        JournalType: item.SourceType,
        Op: data.operation,
        Text: data.noteApproval,
      });
    } else if (
      [
        "Work Request",
        "Work Order",
        "Work Order Report Consumption",
        "Work Order Report Resource",
        "Work Order Report Output",
      ].includes(item.SourceType)
    ) {
      payloadMFG.push({
        JournalID: item.SourceID,
        JournalType: item.SourceType,
        Op: data.operation,
        Text: data.noteApproval,
      });
    } else {
      payloadFico.push({
        JournalID: item.SourceID,
        Op: data.operation,
        Text: data.noteApproval,
      });
    }
  });
  if (payloadFico.length > 0) {
    post.push(postDetail("/fico/approvalaggregator/post", payloadFico));
  }
  if (payloadSCM.length > 0) {
    post.push(postDetail("/scm/postingprofile/post", payloadSCM));
  }
  if (payloadMFG.length > 0) {
    post.push(postDetail("/mfg/postingprofile/post", payloadMFG));
  }
  Promise.all(post)
    .then(() => {
      util.showInfo(`Data has been ${data.operation}`);
      onShowDetail(data.recordGroup);
    })
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.isShowGroup = true;
      data.isLoading1 = false;
    });
}

function onPreview(p) {
  data.recordDetail = p;
  data.recordDetail.SourceJournalID = p.SourceID;
  // data.recordDetail.SourceType = p.SourceType;

  data.appMode = "preview";
}

function closePreview() {
  data.appMode = "grid";
}
function getPrefix(type) {
  switch (type) {
    case "Purchase Request":
      return "PR";
    case "Purchase Order":
      return "PO";
    case "Movement In":
      return "MI";
    case "Movement Out":
      return "MO";
    case "Item Transfer":
      return "IT";
    case "Item Request":
      return "IR";
    case "Work Request":
      return "WR";
    case "Work Order":
      return "WO";
    case "Good Receive":
      return "GR";
    case "Good Issuance":
      return "GI";
    default:
      return type;
  }
}
const getJournalType = computed({
  get() {
    switch (data.recordDetail.SourceType) {
      case "Purchase Request":
        return "Purchase Request";
      case "Purchase Order":
        return "Purchase Order";
      case "Movement In":
        return "Movement In";
      case "Movement Out":
        return "Movement Out";
      case "Item Transfer":
        return "Item Transfer";
      case "Item Request":
        return "Item Request";
      case "Work Request":
        return "Work Request";
      case "Work Order":
        return "Work Order";
      case "Good Receive":
        return "Good Receive";
      case "Good Issuance":
        return "Good Issuance";
      default:
        return data.recordDetail.SourceType;
    }
  },
});
const getModuleId = computed({
  get() {
    switch (data.recordDetail.SourceType) {
      case "Purchase Request":
      case "Purchase Order":
      case "Movement In":
      case "Movement Out":
      case "Item Transfer":
      case "Item Request":
      case "Inventory Issuance":
      case "Inventory Receive":
        return "scm/new";

      case "Work Request":
      case "Work Order":
        return "mfg";

      case "Sales Quotation":
      case "Sales Order":
        return "sdp/new";

      case "Good Receive":
      case "Good Issuance":
        return "scm";

      default:
        return "fico";
    }
  },
});
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "READY"].includes(data.recordDetail.Status);
  },
});
function trxPostSubmit(record, action) {
  util.showInfo(`Data has been ${action}`);
  closePreview();
  util.nextTickN(2, () => {
    loadDataGroup(data.tabSelectedStatus);
  });
}
function trxErrorSubmit(e) {}
</script>

<style>
.approveList .suim_list ul,
.itemList .suim_list ul {
  grid-template-columns: repeat(1, minmax(0, 1fr)) !important;
  gap: 0 !important;
}
.approveList .suim_list ul li:hover > div > div,
.itemList .suim_list ul li:hover > div > div {
  @apply bg-slate-100;
  color: #222 !important;
}
.approveList .suim_list ul li:hover label.input_label,
.itemList .suim_list ul li:hover label.input_label {
  color: #222 !important;
}

.approveList.suim_list ul li > div,
.itemList.suim_list ul li > div {
  padding: 0 !important;
}
</style>
