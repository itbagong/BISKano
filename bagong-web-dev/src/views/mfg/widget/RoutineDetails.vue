<template>
  <s-card :title="props.title" class="w-full bg-white suim_datalist">
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
        <!-- <s-button
          icon="content-save"
          label="Save"
          @click="onSave"
          class="bg-primary text-white"
        /> -->
      </div>
    </template>
    <!-- <div class="mb-5 flex border border-gray-300 p-5 items-center gap-5"> -->
    <div
      class="mb-5 border border-gray-300 p-5 grid grid-cols-1 md:grid-cols-2"
    >
      <div class="flex gap-2">
        <span class="font-semibold text-lg">Site:</span>
        <span class="text-lg">{{ props.dataRoutine?.SiteName }}</span>
      </div>
      <div class="flex md:border-l md:border-gray-300 md:pl-5 gap-2">
        <span class="font-semibold text-lg">Date:</span>
        <span class="text-lg">{{
          moment(props.dataRoutine?.ExecutionDate).format("DD-MMM-YYYY")
        }}</span>
      </div>
    </div>
    <s-grid
      class="w-full r-grid grid-line-items"
      ref="gridRoutineDetails"
      hideNewButton
      hide-action
      hide-control
      hide-select
      :read-url="'/mfg/routine/detail/gets?RoutineID=' + props.dataRoutine._id"
      v-if="data.gridCfg.setting"
      :config="data.gridCfg"
      form-keep-label
      auto-commit-line
    >
      <template #item_AssetID="{ item, i }">
        <a
          class="font-medium text-blue-600 hover:underline"
          @click="onSelect(item)"
          >{{ item.AssetID }}</a
        >
      </template>
      <template #item_TypeID="{ item }">
        <div class="bg-transparent">{{ item.AssetTypeName }}</div>
      </template>
      <template #item_StatusCondition="{ item }">
        <div class="bg-transparent">
          {{
            item.StatusCondition
              ? data.statusOptions.find((o) => o.key === item.StatusCondition)
                  .text
              : "&nbsp;"
          }}
        </div>
        <!-- <s-input
          :items="data.statusOptions"
          v-model="item.StatusCondition"
          read-only
          lookup-key="key"
          keep-label
          :lookup-labels="['text']"
          :lookup-searchs="['key', 'text']"
          use-list
          class="w-full"
          @change="(field, v1, v2, old, ctlRef) => onChangeStatusCondition(item, v1)"
        ></s-input> -->
      </template>
      <template #item_Download="{ item }">
        <div class="flex gap-2 justify-center items-center my-2">
          <s-button
            :icon="`download`"
            class="btn_primary submit_btn"
            label="Download as PDF"
            :disabled="false"
            :no-tooltip="true"
            @click="downloadPDF(item)"
          />
        </div>
      </template>
    </s-grid>
  </s-card>
</template>

<script setup>
import { reactive, onMounted, inject, ref } from "vue";
import {
  SCard,
  SGrid,
  SForm,
  loadGridConfig,
  util,
  SInput,
  SButton,
} from "suimjs";
import moment from "moment";

const axios = inject("axios");

const props = defineProps({
  title: { type: String, default: "" },
  dataRoutine: { type: Object, default: undefined },
});

const emit = defineEmits({
  selectdata_detail: null,
  back: null,
});
const gridRoutineDetails = ref(null);
const data = reactive({
  gridCfg: {},
  statusOptions: [
    {
      key: "NotCheckedYet",
      text: "Not Checked Yet",
    },
    {
      key: "NeedRepair",
      text: "Need Repair",
    },
    {
      key: "RunningWell",
      text: "Running Well",
    },
  ],
});
function onSelect(dt) {
  emit("selectdata_detail", dt);
}
// function onSave() {
//   const records = gridRoutineDetails.value.getRecords();
//   const payload = {
//     RoutineID: props.dataRoutine._id,
//     RoutineDetails: [...records],
//   };
//   axios.post("/mfg/routine/detail/save-multiple", payload).then(
//     (r) => {
//       emit("back");
//       util.showInfo("data has been saved");
//     },
//     (e) => util.showError(e)
//   );
// }
function saveRoutineDetail(item) {
  axios.post("/mfg/routine/detail/save", item).then(
    (r) => {},
    (e) => util.showError(e)
  );
}
function onChangeStatusCondition(item, v1) {
  saveRoutineDetail({ ...item, StatusCondition: v1 });
}

function downloadPDF(asset) {
  const link = document.createElement("a");
  link.href = `${window.location.origin}/v1/mfg/routine/checklist/download-as-pdf?RoutineDetailID=${asset._id}`;
  link.target = "_blank";
  link.click();
  link.remove();
}
onMounted(() => {
  loadGridConfig(axios, "/mfg/routine/detail/gridconfig").then(
    (r) => {
      const _fields = r.fields.filter((o) =>
        ["TypeID", "PhysicalAvailability", "StatusCondition"].includes(o.field)
      );

      const mappingFields = _fields.map((o) => {
        if (!["PhysicalAvailability", "StatusCondition"].includes(o.field)) {
          o.input.readOnly = true;
          o.input.disable = true;
        }
        return o;
      });
      const newField = [
        {
          field: "AssetName",
          kind: "text",
          label: "Asset name",
          labelField: "",
          readType: "show",
          pos: 1000,
          input: {
            field: "AssetName",
            label: "Asset name",
            hint: "",
            hide: false,
            placeHolder: "Asset name",
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
            readOnly: true,
          },
        },
        {
          field: "AssetTypeName",
          kind: "text",
          label: "Asset Type",
          labelField: "",
          readType: "show",
          pos: 1000,
          input: {
            field: "AssetTypeName",
            label: "Asset Type",
            hint: "",
            hide: false,
            placeHolder: "Asset Type",
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
            readOnly: true,
          },
        },
        {
          field: "DriveType",
          kind: "text",
          label: "Drive Type",
          labelField: "",
          readType: "show",
          pos: 1000,
          input: {
            field: "DriveType",
            label: "Drive Type",
            hint: "",
            hide: false,
            placeHolder: "Drive Type",
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
            readOnly: true,
          },
        },
      ];

      const download = {
        field: "Download",
        kind: "text",
        label: "Download",
        labelField: "",
        readType: "show",
        width: "200px",
        input: {
          field: "Download",
          label: "Download",
          hint: "",
          width: "200px",
          hide: false,
          placeHolder: "Download",
          kind: "text",
          disable: false,
          required: false,
          multiple: false,
          readOnly: true,
        },
      };
      const assetIDField = r.fields.find((o) => o.field === "AssetID");
      data.gridCfg = {
        ...r,
        fields: [assetIDField, ...newField, ...mappingFields, download],
      };
    },
    (e) => util.showError(e)
  );
  // axios.post("/mfg/routine/detail/gets", [props.dataRoutine._id]).then(
  //   (r) => {
  //     console.log('r', r)
  //   },
  //   (e) => util.showError(e)
  // );
});
</script>
