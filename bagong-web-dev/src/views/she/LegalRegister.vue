<template>
  <data-list
    class="card"
    ref="listControl"
    title="Legal Register"
    grid-config="/she/legalregister/gridconfig"
    form-config="/she/legalregister/formconfig"
    grid-read="/she/legalregister/gets"
    form-read="/she/legalregister/get"
    grid-mode="grid"
    grid-delete="/she/legalregister/delete"
    form-keep-label
    form-insert="/she/legalregister/save"
    form-update="/she/legalregister/save"
    :init-app-mode="data.appMode"
    gridHideSelect
    @form-new-data="newRecord"
    @form-edit-data="editRecord"
    :form-fields="['Dimension', 'Status']"
    :form-tabs-edit="['General', 'Detail', 'Reference', 'Attachments']"
    :grid-hide-new="!profile.canCreate"
    :grid-hide-edit="!profile.canUpdate"
    :grid-hide-delete="!profile.canDelete"
    @pre-save="onPreSave"
    @postSave="onPostSave"
    @form-field-change="onFormFieldChange"
    stay-on-form-after-save
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
          ref="refNo"
          v-model="data.search.No"
          class="w-[300px]"
          label="Legal No"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refType"
          v-model="data.search.Type"
          class="w-[300px]"
          use-list
          label="Type"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=LTY"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refFields"
          v-model="data.search.Category"
          class="w-[300px]"
          use-list
          label="Category"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=LCA"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refFields"
          v-model="data.search.Fields"
          class="w-[300px]"
          use-list
          multiple
          label="Fields"
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=LFI"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refStatus"
          v-model="data.search.Status"
          class="w-[300px]"
          label="Status"
          use-list
          :items="['Active', 'Inactive']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refSite"
          v-model="data.search.Site"
          lookup-key="_id"
          label="Site"
          class="w-[300px]"
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
    <template #form_input_Dimension="{ item }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :default-list="profile.Dimension"
      ></dimension-editor-vertical>
    </template>
    <template #form_input_Status="{ item }">
      <label class="input_label">
        <div>Status</div>
      </label>
      <s-toggle
        v-model="item.Status"
        class="w-[120px] mt-0.5"
        yes-label="active"
        no-label="inactive"
      />
    </template>
    <!-- <template #form_input_Reference="{ item, config }">
      <uploader
        ref="gridAttachment"
        :journalId="item._id"
        :config="config"
        journalType="LEGAL_REGISTER"
        :tags="[`LEGAL_COMPLIANCE_${item._id}`]"
      />
    </template> -->
    <template #form_tab_Detail="{ item }">
      <legal-detail v-model="item.LegalDetails" @countPlant="onCountPlant" />
    </template>
    <template #form_tab_Reference="{ item }">
      <s-grid-attachment
        :key="data.record._id"
        :journal-id="data.record._id"
        :tags="[`LEGAL_COMPLIANCE_${item._id}`]"
        journal-type="LEGAL_REGISTER"
        ref="gridLineReference"
      ></s-grid-attachment>
    </template>
    <template #form_tab_Attachments="{ item }">
      <s-grid-attachment
        :key="data.record._id"
        :journal-id="data.record._id"
        :tags="linesTag"
        journal-type="LEGAL_REGISTER"
        ref="gridLineAttachment"
        @pre-Save="preSaveAttachment"
      ></s-grid-attachment>
    </template>
  </data-list>
</template>
<script setup>
import { reactive, ref, inject, watch, computed, onMounted } from "vue";
import { DataList, SInput, util, SButton, SGrid } from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import SToggle from "@/components/common/SButtonToggle.vue";
import Uploader from "@/components/common/Uploader.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import helper from "@/scripts/helper.js";
import LegalDetail from "./widget/LegalDetailRegister.vue";

layoutStore().name = "tenant";
const axios = inject("axios");

const FEATUREID = "LegalRegister";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const gridAttachment = ref(null);
const gridLineAttachment = ref(null);
const gridLineReference = ref(null);

const FormMode = computed({
  get() {
    return listControl.value.getFormMode();
  },
});

const linesTag = computed({
  get() {
    const tags = [`LEGAL_REGISTER_${data.record._id}`];
    return tags;
  },
});

let customFilter = computed(() => {
  const filters = [];
  if (
    data.search.DateFrom !== null &&
    data.search.DateFrom !== "" &&
    data.search.DateFrom !== "Invalid date"
  ) {
    filters.push({
      Field: "Date",
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
      Field: "Date",
      Op: "$lte",
      Value: helper.formatFilterDate(data.search.DateTo, true),
      // Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
    });
  }

  if (data.search.No !== null && data.search.No !== "") {
    filters.push({
      Field: "LegalNo",
      Op: "$contains",
      Value: [data.search.No],
    });
  }

  if (data.search.Type !== null && data.search.Type !== "") {
    filters.push({
      Field: "Type",
      Op: "$eq",
      Value: data.search.Type,
    });
  }

  if (data.search.Category !== null && data.search.Category !== "") {
    filters.push({
      Field: "Category",
      Op: "$eq",
      Value: data.search.Category,
    });
  }

  if (data.search.Fields !== null && data.search.Fields.length > 0) {
    filters.push({
      Field: "Fields",
      Op: "$contains",
      Value: data.search.Fields,
    });
  }

  if (data.search.Status !== null && data.search.Status !== "") {
    let Status = data.search.Status === "Active" ? true : false;
    filters.push({
      Field: "Status",
      Op: "$eq",
      Value: Status,
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
  gridCfg: {},
  search: {
    DateFrom: null,
    DateTo: null,
    No: "",
    Type: "",
    Fields: [],
    Site: "",
    Category: "",
    Status: "",
  },
});

function newRecord(r) {
  r.Date = new Date();
  openForm(r);
}
function editRecord(r) {
  openForm(r);
}

function openForm(r) {
  util.nextTickN(2, () => {
    data.record = r;
    listControl.value.setFormFieldAttr("Reference", "hide", true);
  });
}

function onPreSave(r) {}

function onPostSave(r) {
  if (FormMode.value != "new") {
    // gridAttachment.value.Save(r._id, "LEGAL_REGISTER")
    gridLineReference.value.Save();
    gridLineAttachment.value.Save();
  }
}

function onCountPlant(count) {
  data.record.PlantCompliance = count;
}

function preSaveAttachment(payload) {
  payload.map((asset) => {
    asset.Asset.Tags = [`LEGAL_REGISTER_${data.record._id}`];
    return asset;
  });
}

function onFormFieldChange(field, v1, v2, old, record) {}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
onMounted(() => {});
</script>
