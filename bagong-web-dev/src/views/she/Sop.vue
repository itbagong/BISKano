<template>
  <data-list
    ref="listControl"
    title="SOP"
    class="SHE_SOP"
    grid-config="/she/mastersop/gridconfig"
    form-config="/she/mastersop/formconfig"
    grid-read="/she/mastersop/gets"
    form-read="/she/mastersop/get"
    grid-mode="grid"
    grid-delete="/she/mastersop/delete"
    form-keep-label
    form-insert="/she/mastersop/save"
    form-update="/she/mastersop/save"
    :init-app-mode="data.appMode"
    :grid-fields="['Attachment', 'Status']"
    :form-fields="['Dimension', 'Attachment']"
    @form-new-data="newData"
    @form-edit-data="openForm"
    @form-field-change="onChangeField"
    stay-on-form-after-save
    form-hide-submit
    :grid-custom-filter="customFilter"
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
          ref="refTitle"
          v-model="data.search.TitleOfDocument"
          label="Title"
          class="w-[200px]"
          keepLabel
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refPICDocument"
          v-model="data.search.PICDocument"
          lookup-key="_id"
          label="PIC Document"
          class="w-[400px]"
          use-list
          :lookup-url="`/tenant/employee/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refPICFacilitator"
          v-model="data.search.PICFacilitator"
          lookup-key="_id"
          label="PIC Facilitator"
          class="w-[400px]"
          use-list
          :lookup-url="`/tenant/employee/find`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refJob"
          v-model="data.search.JobPosition"
          lookup-key="_id"
          label="Job Position"
          class="w-[400px]"
          use-list
          multiple
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=PTE`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refStatus"
          v-model="data.search.Status"
          label="Status"
          class="w-[200px]"
          use-list
          :items="['DRAFT', 'SUBMITTED', 'READY', 'POSTED', 'REJECTED']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refSite"
          v-model="data.search.Site"
          lookup-key="_id"
          label="Site"
          class="w-[200px]"
          use-list
          :disabled="defaultList?.length === 1"
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
      </div>
    </template>
    <template #form_input_Dimension="{ item, mode }">
      <dimension-editor-vertical
        ref="dimenstionCtl"
        v-model="item.Dimension"
        :read-only="mode == 'view'"
      ></dimension-editor-vertical>
    </template>
    <template #form_input_Attachment="{ item, config }">
      <uploader
        :journalId="item._id"
        :config="config"
        journalType="SHE_SOP"
        single-save
        @modal="modalMode"
      />
    </template>
    <template #grid_Attachment="{ item, config }">
      <uploader
        :journalId="item._id"
        :config="{}"
        journalType="SHE_SOP"
        read-only
      />
    </template>
    <template #form_buttons_1="{ item, config }">
      <s-button
        icon="content-save"
        class="btn_primary submit_btn"
        label="Save"
        tooltip="Save"
        @click="onSubmit"
      />
    </template>
    <template #grid_Status="{ item }">
      <status-text :txt="item.Status" />
    </template>
  </data-list>
</template>
<script setup>
import {
  reactive,
  ref,
  inject,
  watch,
  computed,
  onMounted,
  nextTick,
} from "vue";

import { SInput, util, SButton, SGrid, SCard, DataList, SModal } from "suimjs";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import { layoutStore } from "@/stores/layout.js";
import Uploader from "@/components/common/Uploader.vue";
import StatusText from "@/components/common/StatusText.vue";

layoutStore().name = "tenant";
const listControl = ref(null);
const dimenstionCtl = ref(null);
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
      Value: moment(data.search.DateFrom).utc().format("YYYY-MM-DDT00:mm:00Z"),
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
      Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
    });
  }

  if (
    data.search.TitleOfDocument !== null &&
    data.search.TitleOfDocument !== ""
  ) {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "TitleOfDocument",
          Op: "$contains",
          Value: [data.search.TitleOfDocument],
        },
      ],
    });
  }
  if (data.search.PICDocument !== null && data.search.PICDocument !== "") {
    filters.push({
      Field: "PICDocument",
      Op: "$eq",
      Value: data.search.PICDocument,
    });
  }
  if (
    data.search.PICFacilitator !== null &&
    data.search.PICFacilitator !== ""
  ) {
    filters.push({
      Field: "PICFacilitator",
      Op: "$eq",
      Value: data.search.PICFacilitator,
    });
  }
  if (data.search.JobPosition !== null && data.search.JobPosition.length > 0) {
    filters.push({
      Field: "JobPosition",
      Op: "$contains",
      Value: data.search.JobPosition,
    });
  }
  if (data.search.Status !== null && data.search.Status !== "") {
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: data.search.Status,
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

  if (filters.length == 1) return filters[0];
  else if (filters.length > 1) return { Op: "$and", Items: filters };
  else return null;
});
const data = reactive({
  appMode: "grid",
  record: {},
  search: {
    DateFrom: null,
    DateTo: null,
    TitleOfDocument: "",
    PICDocument: "",
    PICFacilitator: "",
    JobPosition: [],
    Status: "",
    Site: "",
  },
});

function newData(r) {
  r.CreatedDate = new Date();
  r.CompletionDate = new Date();
  r.EffectiveDate = new Date();
  data.record = r;
}

function openForm(r) {
  data.record = r;
  nextTick(() => {
    let isPembuatan = r.NatureOfChange == "Pembuatan";
    setAttr("DocumentRefno", "required", !isPembuatan);
    setAttr("DocumentRefno", "hide", isPembuatan);
  });
}

function onChangeField(name, v1, v2, old, dt) {
  if (name == "NatureOfChange") {
    dt.DocumentRefno = "";
    let isPembuatan = v1 == "Pembuatan";
    setAttr("DocumentRefno", "required", !isPembuatan);
    setAttr("DocumentRefno", "hide", isPembuatan);
  }
}

function setAttr(field, attr, val) {
  listControl.value.setFormFieldAttr(field, attr, val);
}

function modalMode(v) {
  if (!data.record._id) saveManual(data.record, true);
}

function saveManual(r, isNotif) {
  listControl.value.submitForm(
    r,
    () => {},
    () => {},
    isNotif
  );
}

function onSubmit() {
  const validDimension = dimenstionCtl.value.validate();
  const validForm = listControl.value.formValidate();

  if (validDimension && validForm) saveManual(data.record, false);
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
</script>

<style>
.SHE_SOP .section_group.col.flex-col.grow:nth-child(1) {
  @apply basis-3/4;
}
.SHE_SOP .section_group.col.flex-col.grow:nth-child(2) {
  @apply basis-1/4;
}

.SHE_SOP .gap-2.grid.gridCol3 {
  @apply gap-4;
}
.SHE_SOP .gap-2.grid.gridCol2 {
  @apply gap-4;
}
</style>
