<template>
  <s-modal
    :display="false"
    ref="confirmModal"
    title="Request Confirmation"
    @submit="onRequest"
  >
    You will requesting this data ! Are you sure ?<br />
    Please be noted, this can not be undone !
  </s-modal>
  <s-card
    :title="props.title"
    class="w-full bg-white suim_datalist"
    hide-footer
  >
    <template #title>
      <div class="grow flex justify-between items-center mb-5">
        <div class="card_title grow flex gap-2 items-center">
          <mdicon
            name="arrow-left"
            size="28"
            class="cursor-pointer"
            @click="emit('back')"
          />{{ props.title }}
        </div>
        <div class="flex gap-2">
          <s-button
            :icon="`download`"
            class="btn_primary submit_btn"
            label="Download as PDF"
            :disabled="false"
            :no-tooltip="true"
            @click="downloadPDF"
          />
          <div
            v-show="!data.records.RoutineChecklist.IsAlreadyRequest"
            class="flex gap-2"
          >
            <div class="w-[55px] h-[30px]" v-if="data.loadingRequest">
              <loader kind="skeleton" skeleton-kind="input" />
            </div>
            <s-button
              v-else
              :disabled="
                disableRequest || data.records.RoutineChecklist.IsAlreadyRequest
              "
              icon="content-save"
              label="Request"
              @click="showDialogRequest"
              class="bg-primary text-white"
              :no-tooltip="true"
            />
            <div class="w-[55px] h-[30px]" v-if="data.loadingSave">
              <loader kind="skeleton" skeleton-kind="input" />
            </div>
            <s-button
              v-else
              icon="content-save"
              label="Save"
              @click="onSave"
              class="bg-primary text-white"
              :no-tooltip="true"
            />
          </div>
        </div>
      </div>
    </template>
    <div class="suim_form">
      <div class="mb-2 flex header">
        <div class="flex tab_container grow">
          <div
            v-for="(tabTitle, tabIdx) in data.tabs"
            @click="data.currentTab = tabIdx"
            :class="{
              tab_selected: data.currentTab == tabIdx,
              tab: data.currentTab != tabIdx,
            }"
          >
            {{ tabTitle }}
          </div>
        </div>
      </div>
      <div id="form_inputs_checlist" v-show="data.currentTab == 0">
        <!-- <div
          class="mb-5 flex border border-gray-300 p-5 justify-between items-center gap-5"
        > -->
        <div class="mb-5 border-gray-300 p-5 grid grid-cols-1 md:grid-cols-2">
          <div class="flex flex-col">
            <span class="font-semibold text-lg">{{
              props.dataRoutineAsset._id +
              " | " +
              props.dataRoutineAsset.AssetName
            }}</span>
            <span class="text-base">{{
              moment(props.dataRoutine?.ExecutionDate).format("DD-MMM-YYYY")
            }}</span>
          </div>
          <div class="flex gap-2">
            <span class="font-semibold text-lg">Status Condition:</span>
            <span class="text-lg">{{
              props.dataRoutineAsset?.StatusCondition?.replace(
                /([A-Z])/g,
                " $1"
              ).trim()
            }}</span>
          </div>
        </div>
        <s-form
          ref="formAssetRef"
          :mode="data.formMode"
          hide-cancel
          hide-submit
          v-model="data.records.RoutineChecklist"
          :config="data.formCfg"
          :keepLabel="true"
          @field-change="(name, v1, v2, old) => onFieldChange(name, v1)"
        >
          <template #input_Name="{ item }">
            <s-input
              :key="data.Name"
              label="Name"
              v-model="item.Name"
              class="w-full"
              :required="data.loadingRequest"
              :disabled="true"
              use-list
              :lookup-url="`/tenant/employee/find`"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
            ></s-input>
          </template>
          <template #input_Department="{ item }">
            <s-input
              :key="data.keyDepartment"
              label="Department"
              v-model="item.Department"
              class="w-full"
              :required="data.loadingRequest"
              :disabled="data.formMode == 'view' || data.disableDepartment"
              use-list
              :lookup-url="`/tenant/dimension/find?DimensionType=CC`"
              lookup-key="_id"
              :lookup-labels="['Label']"
              :lookup-searchs="['_id', 'Label']"
            ></s-input>
          </template>
          <template #input_WorkLocation="{ item }">
            <s-input
              :key="data.keyWO"
              label="Work Location"
              v-model="item.WorkLocation"
              class="w-full"
              :required="data.loadingRequest"
              :disabled="data.formMode == 'view' || data.disableWO"
              use-list
              :lookup-url="`/tenant/dimension/find?DimensionType=Site`"
              lookup-key="_id"
              :lookup-labels="['Label']"
              :lookup-searchs="['_id', 'Label']"
            ></s-input>
          </template>
          <template #input_Marketing="{ item }">
            <div class="pr-6">
              <s-input
                label="Marketing"
                v-model="item.Marketing"
                class="w-[25%]"
                :required="data.loadingRequest"
                :disabled="data.formMode == 'view'"
                use-list
                :lookup-url="`/tenant/employee/find`"
                lookup-key="_id"
                :lookup-labels="['Name']"
                :lookup-searchs="['_id', 'Name']"
              ></s-input>
            </div>
          </template>
        </s-form>
        <div class="mb-2 flex header">
          <div class="flex tab_container grow">
            <div
              v-for="(category, tabIdx) in data.records
                .RoutineChecklistCategories"
              @click="onChangeTabChecklist(category, tabIdx)"
              :class="{
                tab_selected: data.currentTabCategories == tabIdx,
                tab: data.currentTabCategories != tabIdx,
              }"
            >
              {{ category.CategoryName }}
            </div>
          </div>
        </div>
        <div id="form_inputs">
          <s-grid
            :editor="data.formMode !== 'view'"
            class="w-full r-grid grid-line-items"
            ref="gridChecklistDetails"
            hideNewButton
            hide-action
            hide-control
            hide-select
            v-if="data.gridChecklistDetailCfg.setting"
            :config="data.gridChecklistDetailCfg"
            form-keep-label
            auto-commit-line
            :model-value="data.listChecklistCategories"
          >
            <template #item_Status="{ item }">
              <div class="flex gap-2">
                <template v-for="(status, idx) in data.status">
                  <div class="radio-status">
                    <input
                      type="radio"
                      :name="item.uuid"
                      class="bg-slate-800"
                      :id="item.uuid + '_' + idx"
                      v-model="item.Status"
                      :value="status.value"
                      :disabled="data.formMode === 'view'"
                    />
                    <label :for="item.uuid + '_' + idx">{{
                      status.text
                    }}</label>
                  </div>
                </template>
              </div>
            </template>
          </s-grid>
        </div>
      </div>
      <div id="form_inputs_attachment" v-show="data.currentTab == 1">
        <!-- <grid-attachment
          gridConfig="/mfg/routine/checklist/attachment/gridconfig"
          :gridFields="['FileName', 'UploadDate', 'URI']"
          v-model="data.records.RoutineChecklistAttachments"
        ></grid-attachment> -->
        <s-grid-attachment
          :journal-id="props.dataRoutineAsset._id"
          journal-type="P2H"
          ref="gridAttachment"
        ></s-grid-attachment>
      </div>
    </div>
  </s-card>
</template>
<script setup>
import { reactive, onMounted, inject, ref, computed } from "vue";
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
} from "suimjs";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import Loader from "@/components/common/Loader.vue";

import moment from "moment";

const axios = inject("axios");

const props = defineProps({
  title: { type: String, default: "" },
  dataRoutine: { type: Object, default: undefined },
  dataRoutineAsset: { type: Object, default: undefined },
  disableDepartment: false,
  disableWO: false,
});

const formAssetRef = ref(null);
const confirmModal = ref(null);
const gridAttachment = ref(SGridAttachment);

const emit = defineEmits({
  back: null,
  postSave: null,
});
const gridChecklistDetails = ref(null);

const data = reactive({
  records: {
    RoutineChecklist: {
      Name: "",
      KmToday: 0,
      TimeBreakdown: moment().format("HH:mm"),
      Departure: moment().format("YYYY-MM-DDTHH:mm"),
      Arrive: moment().format("YYYY-MM-DDTHH:mm"),
    },
    RoutineChecklistAttachments: [],
    RoutineChecklistCategories: [],
    RoutineDetailID: "",
  },
  formCfg: {},
  gridChecklistDetailCfg: {},
  tabs: ["Checklist", "Attachment"],
  currentTab: 0,
  currentTabCategories: 0,
  listChecklistCategories: [],
  status: [
    { value: "Normal", text: "Normal" },
    { value: "Damaged", text: "Damaged" },
  ],
  formMode: "edit",
  loadingRequest: false,
  loadingSave: false,
  keyDepartment: util.uuid(),
  keyWO: util.uuid(),
});

const onFieldChange = (name, v1) => {
  if (name === "Name" && typeof v1 === "string") {
    axios.post("/bagong/employee/get", [v1]).then((r) => {
      // set department
      data.records.RoutineChecklist.Department = r.data.Dimension.reduce(
        (val, obj) => {
          if (obj.Key === "CC") val = obj.Value;
          return val;
        },
        ""
      );
      if (data.records.RoutineChecklist.Department != "") {
        data.disableDepartment = true;
      } else {
        data.disableDepartment = false;
      }
      data.keyDepartment = util.uuid();

      // set work location
      data.records.RoutineChecklist.WorkLocation = r.data.Dimension.reduce(
        (val, obj) => {
          if (obj.Key === "Site") val = obj.Value;
          return val;
        },
        ""
      );
      if (data.records.RoutineChecklist.WorkLocation != "") {
        data.disableWO = true;
      } else {
        data.disableWO = false;
      }
      data.keyWO = util.uuid();
    });
  } else if (name === "Name") {
    data.records.RoutineChecklist.Department = "";
    data.disableDepartment = false;
    data.records.RoutineChecklist.WorkLocation = "";
    data.disableWO = false;
  }
};
const onChangeTabChecklist = (items, tabIndex) => {
  data.currentTabCategories = tabIndex;
  const listChecklistCategories =
    data.records?.RoutineChecklistCategories[data.currentTabCategories]
      ?.RoutineChecklistDetails;

  listChecklistCategories.map(function (check) {
    check.uuid = util.uuid();
    return check;
  });
  data.listChecklistCategories = listChecklistCategories;
  // gridChecklistDetails.value.setRecords(items.RoutineChecklistDetails);
};
function onSave(cbOK, cbFalse) {
  data.loadingSave = true;
  setRequiredAllField(true);
  const status = checkChecklist();
  if (status === "NotCheckedYet") {
    data.loadingSave = false;
    if (cbFalse) {
      cbFalse();
    }
    return util.showError(
      "Oops! It seems that some status on the checklist remain unchecked. Please complete all items before proceeding further."
    );
  }
  const attachments = data.records.RoutineChecklistAttachments.map((item) => {
    return {
      ...item,
      RoutineDetailID: props.dataRoutine._id,
      RoutineChecklistID: props.dataRoutineAsset._id,
    };
  });
  const payload = {
    ...data.records,
    RoutineChecklistAttachments: attachments,
  };

  let valid = true;
  if (formAssetRef.value) {
    valid = formAssetRef.value.validate();
  }

  if (!valid) {
    if (cbFalse) {
      cbFalse();
    }
    data.loadingSave = false;
    return util.showError("Please check required field");
  }
  payload.RoutineChecklist.Departure =
    moment(data.records.RoutineChecklist.Departure).format(
      "YYYY-MM-DDTHH:mm:00Z"
    ) == "Invalid date"
      ? moment(new Date()).format("YYYY-MM-DDTHH:mm:00Z")
      : moment(data.records.RoutineChecklist.Departure).format(
          "YYYY-MM-DDTHH:mm:00Z"
        );
  payload.RoutineChecklist.Arrive =
    moment(data.records.RoutineChecklist.Arrive).format(
      "YYYY-MM-DDTHH:mm:00Z"
    ) == "Invalid date"
      ? moment(new Date()).format("YYYY-MM-DDTHH:mm:00Z")
      : moment(data.records.RoutineChecklist.Arrive).format(
          "YYYY-MM-DDTHH:mm:00Z"
        );
  axios
    .post("/mfg/routine/checklist/save-checklist", payload)
    .then(
      (r) => {
        gridAttachment.value.Save();

        r.data.RoutineChecklist.Departure = moment(
          moment(
            r.data.RoutineChecklist.Departure
              ? r.data.RoutineChecklist.Departure
              : new Date()
          ).format("YYYY-MM-DDTHH:mm:00Z")
        ).format("YYYY-MM-DDTHH:mm");
        r.data.RoutineChecklist.Arrive = moment(
          moment(
            r.data.RoutineChecklist.Arrive
              ? r.data.RoutineChecklist.Arrive
              : new Date()
          ).format("YYYY-MM-DDTHH:mm:00Z")
        ).format("YYYY-MM-DDTHH:mm");
        data.records = r.data;
        if (cbOK) {
          emit("postSave", status, cbOK);
        } else {
          emit("postSave", status);
          util.showInfo("data has been saved");
        }
      },
      (e) => {
        if (cbFalse) {
          cbFalse();
        }
        data.loadingSave = false;
        util.showError(e);
      }
    )
    .finally(() => {
      data.loadingSave = false;
      setRequiredAllField(true);
    });
}
function checkChecklist() {
  let status = "NotCheckedYet";
  let checklists = [];
  data.records.RoutineChecklistCategories.map((item) => {
    checklists = [...checklists, ...item.RoutineChecklistDetails];
  });
  if (checklists.find((o) => o.Status === "")) {
    status = "NotCheckedYet";
  } else if (checklists.find((o) => o.Status === "Damaged")) {
    status = "NeedRepair";
  } else if (checklists.find((o) => o.Status === "Normal")) {
    status = "RunningWell";
  }
  return status;
}

function setRequiredAllField(required) {
  formAssetRef.value.getAllField().forEach((e) => {
    if (
      ![
        "TimeBreakdown",
        "Departure",
        "Arrive",
        "HelperName",
        "BBMLevel",
      ].includes(e.field)
    ) {
      formAssetRef.value.setFieldAttr(e.field, "required", required);
    }
  });
  formAssetRef.value.setFieldAttr("KmToday", "rules", [
    (v) => {
      const errorStr = "value must be greater than 0";
      if (v <= 0) {
        return errorStr;
      }
      return "";
    },
  ]);
}
function onRequest() {
  onSave(
    () => {
      data.loadingRequest = true;
      setRequiredAllField(true);
      const attachments = data.records.RoutineChecklistAttachments.map(
        (item) => {
          return {
            ...item,
            RoutineDetailID: props.dataRoutine._id,
            RoutineChecklistID: props.dataRoutineAsset._id,
          };
        }
      );
      const assets = {
        ...data.records,
        RoutineChecklistAttachments: attachments,
      };

      let payload = {
        RoutineChecklistID: data.records.RoutineChecklist._id,
        RoutineID: props.dataRoutine._id,
        EquipmentNo: props.dataRoutineAsset.AssetID,
        Kilometers: data.records.RoutineChecklist.KmToday,
        Description: "",
        Site: props.dataRoutine.SiteID,
        ...assets,
      };
      util.nextTickN(2, () => {
        const valid = formAssetRef.value.validate();
        if (valid) {
          axios
            .post("/mfg/routine/create-for-wo", payload)
            .then(
              (r) => {
                confirmModal.value.hide();
                emit("back");
                util.showInfo("data has been saved");
              },
              (e) => {
                confirmModal.value.hide();
                return util.showError(e);
              }
            )
            .finally(() => {
              data.loadingRequest = false;
            });
        } else {
          util.showError("Please check required field");
          confirmModal.value.hide();
          data.loadingRequest = false;
        }
        setRequiredAllField(false);
      });
    },
    () => {
      data.loadingSave = true;
    }
  );
}
const disableRequest = computed({
  get: () => {
    let checklists = [];
    data.records.RoutineChecklistCategories.map((item) => {
      checklists = [...checklists, ...item.RoutineChecklistDetails];
    });
    if (checklists.find((o) => o.Status === "Damaged")) {
      return false;
    }
    return true;
  },
});
function showDialogRequest() {
  confirmModal.value.show();
}
function downloadPDF() {
  const link = document.createElement("a");
  link.href = `${window.location.origin}/v1/mfg/routine/checklist/download-as-pdf?RoutineDetailID=${props.dataRoutineAsset._id}`;
  link.target = "_blank";
  link.click();
  link.remove();
}
onMounted(() => {
  loadFormConfig(axios, "/mfg/routine/checklist/formconfig").then(
    (r) => {
      data.formCfg = r;
    },
    (e) => util.showError(e)
  );
  axios
    .post("/mfg/routine/checklist/get-checklist", {
      RoutineDetailID: props.dataRoutineAsset._id,
    })
    .then(
      (r) => {
        r.data.RoutineChecklist.TimeBreakdown = r.data.RoutineChecklist
          .TimeBreakdown
          ? r.data.RoutineChecklist.TimeBreakdown
          : moment(moment(new Date()).format("YYYY-MM-DDTHH:mm:00Z")).format(
              "HH:mm"
            );

        r.data.RoutineChecklist.Departure = moment(
          moment(
            r.data.RoutineChecklist.Departure
              ? r.data.RoutineChecklist.Departure
              : new Date()
          ).format("YYYY-MM-DDTHH:mm:00Z")
        ).format("YYYY-MM-DDTHH:mm");
        r.data.RoutineChecklist.Arrive = moment(
          moment(
            r.data.RoutineChecklist.Arrive
              ? r.data.RoutineChecklist.Arrive
              : new Date()
          ).format("YYYY-MM-DDTHH:mm:00Z")
        ).format("YYYY-MM-DDTHH:mm");

        data.records = r.data;
        let listChecklistCategories = [];
        if (data.records?.RoutineChecklistCategories.length > 0) {
          listChecklistCategories =
            data.records?.RoutineChecklistCategories[data.currentTabCategories]
              ?.RoutineChecklistDetails;

          listChecklistCategories.map(function (check) {
            check.uuid = util.uuid();
            return check;
          });
        }
        data.listChecklistCategories = listChecklistCategories;
        if (data.records.RoutineChecklist.IsAlreadyRequest) {
          data.formMode = "view";
        }
        // gridChecklistDetails.value.setRecords(
        //   r.data.RoutineChecklistCategories[0].RoutineChecklistDetails
        // );
        // set deparment and work location
        if (data.records.RoutineChecklist.Department != "") {
          data.disableDepartment = true;
        } else {
          data.disableDepartment = false;
        }

        if (data.records.RoutineChecklist.WorkLocation != "") {
          data.disableWO = true;
        } else {
          data.disableWO = false;
        }
      },
      (e) => util.showError(e)
    );
  loadGridConfig(axios, "/mfg/routine/checklist/detail/gridconfig").then(
    (r) => {
      const _fields = r.fields.filter((o) =>
        ["Name", "Status", "Code", "Note"].includes(o.field)
      );
      const mappingFields = _fields.map((o) => {
        if (!["Status", "Code", "Note"].includes(o.field)) {
          o.input.readOnly = true;
          o.input.disable = true;
        }
        return o;
      });
      data.gridChecklistDetailCfg = {
        ...r,
        fields: [...mappingFields],
      };
    },
    (e) => util.showError(e)
  );
});
</script>
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
  background-color: #000;
}
</style>
