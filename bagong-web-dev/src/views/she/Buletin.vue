<template>
  <data-list
    class="card"
    ref="listControl"
    title="Buletin"
    grid-config="/she/bulletin/gridconfig"
    form-config="/she/bulletin/formconfig"
    grid-read="/she/bulletin/gets"
    form-read="/she/bulletin/get"
    grid-mode="grid"
    grid-delete="/she/bulletin/delete"
    form-keep-label
    form-insert="/she/bulletin/save"
    form-update="/she/bulletin/save"
    :init-app-mode="data.appMode"
    gridHideSelect
    stayOnFormAfterSave
    @form-new-data="newRecord"
    @form-edit-data="openForm"
    :form-fields="['Dimension', 'Banner', 'Tag']"
    :grid-hide-new="!profile.canCreate"
    :grid-hide-edit="!profile.canUpdate"
    :grid-hide-delete="!profile.canDelete"
    @postSave="onPostSave"
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
          label="ID"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refTitle"
          v-model="data.search.Title"
          label="Title"
          class="w-[200px]"
          @keyup.enter="refreshData"
        ></s-input>
        <s-input
          ref="refCategory"
          v-model="data.search.Category"
          lookup-key="_id"
          label="Category"
          class="w-[400px]"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=BCT`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refTag"
          v-model="data.search.Tag"
          lookup-key="_id"
          label="Tag"
          class="w-[400px]"
          use-list
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=TAGS`"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="refreshData"
        ></s-input>
        <s-input
          ref="refStatus"
          v-model="data.search.IsStatus"
          label="Status"
          class="w-[200px]"
          use-list
          :items="['Active', 'Inactive']"
          @change="refreshData"
        ></s-input>
        <div class="suim_input w-[200px]">
          <div>
            <label class="input_label"><div>Pin</div></label>
            <div>
              <s-toggle
                v-model="data.search.IsPin"
                class="w-[120px] mt-0.5"
                yes-label="pin"
                no-label="not pin"
                @change="refreshData"
              />
            </div>
          </div>
        </div>

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
    <template #form_input_Dimension="{ item }">
      <dimension-editor-vertical
        v-model="item.Dimension"
        :default-list="profile.Dimension"
      ></dimension-editor-vertical>
    </template>
    <template #form_input_Banner="{ item, config }">
      <uploader
        ref="gridAttachment"
        :journalId="item._id"
        :config="config"
        journalType="SHE_BULETIN"
        is-single-upload
        @pre-open="preOpenUploader"
      />
    </template>
    <template #form_input_Tag="{ item, config }">
      <s-input
        class="custom-tags-bulletin"
        :label="config.label"
        use-list
        multiple
        allowAdd
        lookup-url="/tenant/masterdata/find?MasterDataTypeID=TAGS"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookup-searchs="['_id', 'Name']"
        v-model="item.Tag"
      />
    </template>
  </data-list>
</template>

<script setup>
import { reactive, ref, inject, watch, computed, onMounted } from "vue";
import { DataList, SInput, util, SButton } from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import moment from "moment";
import Uploader from "@/components/common/Uploader.vue";
import SToggle from "@/components/common/SButtonToggle.vue";
import { authStore } from "@/stores/auth.js";

const listControl = ref(null);
const gridAttachment = ref(null);

layoutStore().name = "tenant";

const FEATUREID = "Buletin";
const profile = authStore().getRBAC(FEATUREID);

const axios = inject("axios");
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
});
const FormMode = computed({
  get() {
    return listControl.value.getFormMode();
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
  if (data.search.No !== null && data.search.No !== "") {
    filters.push({
      Field: "_id",
      Op: "$contains",
      Value: [data.search.No],
    });
  }
  if (data.search.Title !== null && data.search.Title !== "") {
    filters.push({
      Op: "$or",
      Items: [
        {
          Field: "Title",
          Op: "$contains",
          Value: [data.search.Title],
        },
      ],
    });
  }
  if (data.search.Category !== null && data.search.Category !== "") {
    filters.push({
      Field: "Category",
      Op: "$eq",
      Value: data.search.Category,
    });
  }
  if (data.search.Tag !== null && data.search.Tag !== "") {
    filters.push({
      Field: "Tag",
      Op: "$eq",
      Value: data.search.Tag,
    });
  }

  if (data.search.IsPin !== null && data.search.IsPin !== "") {
    filters.push({
      Field: "IsPin",
      Op: "$eq",
      Value: data.search.IsPin,
    });
  }

  if (data.search.IsStatus !== null && data.search.IsStatus !== "") {
    let IsStatus = data.search.IsStatus == "Active" ? true : false;
    filters.push({
      Field: "IsStatus",
      Op: "$eq",
      Value: IsStatus,
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
    No: "",
    Title: "",
    Category: "",
    Tag: "",
    IsStatus: null,
    IsPin: null,
    Site: "",
  },
});

const emit = defineEmits({
  "update:modelValue": null,
});

function onAlterFormConfig(config) {
  data.record = props.modelValue ?? {};
  data.record.Sign = data.record.Sign ?? [];
  data.record.Evidance = data.record.Evidance ?? [];
  listControl.value.setFormRecord(data.record);
}

function onPostSave(r) {
  for (let i in r.Tag) {
    let o = r.Tag[i];
    if (!o.includes("TAGS")) saveTags(o);
  }
  if (FormMode.value != "new") {
    gridAttachment.value.Save(r._id, "SHE_BULETIN");
  }
}

function saveTags(id) {
  const url = "/tenant/masterdata/save";
  let param = {
    IsActive: true,
    MasterDataTypeID: "TAGS",
    Name: id,
    ParentID: "",
    _id: "TAGS" + id,
  };
  axios.post(url, param).then(
    (r) => {},
    (e) => {}
  );
}
function preOpenUploader() {
  if (!data.record._id) {
    saveManual();
  }
}
function saveManual() {
  listControl.value.setFormLoading(true);
  listControl.value.submitForm(
    data.record,
    () => {
      listControl.value.setFormLoading(false);
    },
    () => {
      listControl.value.setFormLoading(false);
    }
  );
}
function refreshData() {
  util.nextTickN(2, () => {
    listControl.value.refreshGrid();
  });
}
watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>

<style>
.custom-tags-bulletin .vs__selected::before {
  content: "#";
}
</style>
