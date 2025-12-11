<template>
  <div class="w-full Pica">
    <data-list
      class="card"
      ref="listControl"
      title="PICA"
      grid-config="/she/pica/gridconfig"
      form-config="/she/pica/formconfig"
      grid-read="/she/pica/gets"
      form-read="/she/pica/get"
      grid-mode="grid"
      grid-delete="/she/pica/delete"
      form-keep-label
      form-insert="/she/pica/take-action"
      form-update="/she/pica/take-action"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      gridHideSelect
      gridHideNew
      @form-new-data="newRecord"
      @form-edit-data="openForm"
      :grid-custom-filter="customFilter"
      :form-fields="['Evidence', 'Status', 'FindingDescription', 'Dimension']"
      :grid-fields="['EmployeeID']"
      :form-tabs-edit="['General', 'Attachment']"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @postSave="onPostSave"
      @alter-form-config="onCfgForm"
    >
      <template #grid_header_search="{ config }">
        <s-input
          kind="date"
          label="Date From"
          v-model="data.search.DateFrom"
          @change="refreshData"
        ></s-input>
        <s-input
          kind="date"
          label="Date To"
          v-model="data.search.DateTo"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refSourceModule"
          v-model="data.search.SourceModule"
          lookup-key="_id"
          label="Source Module"
          class="w-full"
          use-list
          :items="['MEETING', 'SAFETYCARD']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refSourceNo"
          v-model="data.search.SourceNo"
          lookup-key="_id"
          label="Source No"
          class="w-full"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refSite"
          v-model="data.search.Site"
          lookup-key="_id"
          label="Site"
          class="w-full"
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
        <s-input
          ref="refStatus"
          v-model="data.search.Status"
          lookup-key="_id"
          label="Status"
          class="w-full"
          use-list
          :items="['Completed', 'Open']"
          @change="refreshData"
        ></s-input>
      </template>
      <template #form_buttons_1="{ item, inSubmission, loading }">
        <form-buttons-trx
          :disabled="inSubmission || loading"
          status="DRAFT"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          journal-type-id="PICA"
          moduleid="she"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
        />
      </template>
      <template #form_input_Evidence="{ item, config }">
        <uploader
          ref="gridAttachment"
          :journalId="item._id"
          :config="config"
          :journalType="prefixUpload"
          :tags="[`${prefixUpload}_EVIDANCE_${item._id}`]"
          single-save
        />
      </template>
      <template #form_input_Status="{ item }">
        <label class="input_label flex"> Status </label>
        <status-text :txt="item.Status" />
      </template>
      <template #grid_EmployeeID="{ item }">
        <s-input
          v-model="item.EmployeeID"
          hide-label
          use-list
          lookup-url="/tenant/employee/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          read-only
        />
      </template>
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor-vertical
          v-model="item.Dimension"
          :read-only="mode == 'view'"
        ></dimension-editor-vertical>
      </template>
      <template #form_input_FindingDescription_footer="{ item }">
        <div class="mb-4"></div>
      </template>

      <template #form_tab_Attachment="{ item }">
        <attachment
          :journal-id="item._id"
          :tags="[`${prefixUpload}_${item._id}`]"
          :journalType="prefixUpload"
          single-save
        />
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, inject, watch, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SInput, util, SButton, createFormConfig } from "suimjs";
import StatusText from "@/components/common/StatusText.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import { authStore } from "@/stores/auth.js";
import Uploader from "@/components/common/Uploader.vue";
import Attachment from "@/components/common/SGridAttachment.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";

layoutStore().name = "tenant";
const FEATUREID = "TPICA";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const gridAttachment = ref(null);
const prefixUpload = "SHE_PICA";
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

  if (data.search.SourceModule !== null && data.search.SourceModule !== "") {
    filters.push({
      Field: "SourceModule",
      Op: "$contains",
      Value: [data.search.SourceModule],
    });
  }

  if (data.search.SourceNo !== null && data.search.SourceNo !== "") {
    filters.push({
      Field: "SourceNumber",
      Op: "$eq",
      Value: data.search.SourceNo,
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
  formMode: "edit",
  formCfg: {},
  record: {},
  search: {
    DateFrom: null,
    DateTo: null,
    SourceModule: "",
    SourceNo: "",
    Status: "",
    Site: "",
  },
  jType: "",
});
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "READY"].includes(data.record.Status);
  },
});
function takeAction(isTA) {
  listControl.value.setFormSectionAttr("TakeAction", "visible", false);
}
function setModeGrid() {
  listControl.value.setControlMode("grid");
  listControl.value.gridResetFilter();
  listControl.value.refreshList();
}
function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    listControl.value.setFormFieldAttr(e.field, "required", required);
  });
}
function openForm(record) {
  const readOnlyField = [
    "EmployeeID",
    "DueDate",
    "FindingDescription",
    "Status",
  ];
  data.record = record;
  data.record.ActionDate = new Date();
  util.nextTickN(2, () => {
    setFormAttr(readOnlyField, "readOnly", true);
  });
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function trxPreSubmit(status, action, doSubmit) {
  if (waitTrxSubmit.value) {
    trxSubmit(doSubmit);
  }
  doSubmit();
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

function trxPostSubmit(record) {
  setLoadingForm(false);
  setModeGrid();
}
function trxErrorSubmit(e) {
  setLoadingForm(false);
}
function setFormAttr(fields, attr, val) {
  for (let i in fields) {
    let f = fields[i];
    listControl.value.setFormFieldAttr(f, attr, val);
  }
}

function onPostSave(r) {
  for (let i in r.Tag) {
    let o = r.Tag[i];
    if (!o.includes("TAGS")) saveTags(o);
  }
  gridAttachment.value.Save(r._id, "SHE_PICA");
}

function onCfgForm(cfg) {
  for (let i in cfg.sectionGroups[0].sections) {
    let o = cfg.sectionGroups[0].sections[i];
    if (o.name == "TakeAction") o.title = "Take Action";
  }
}

function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
</script>

<style scoped>
.Pica .box-upload {
  @apply border-2 border-dashed border-zinc-200 w-20 h-20 mr-2 rounded-md grid place-content-center relative;
}
</style>
