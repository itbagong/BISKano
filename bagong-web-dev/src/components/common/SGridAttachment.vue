<template>
  <div>
    <loader v-if="data.loading" kind="skeleton" />

    <s-grid
      v-show="!data.loading"
      ref="gridref"
      class="w-full"
      editor
      :config="data.listCfg"
      hide-select
      hide-search
      hide-detail
      hide-sort
      hide-refresh-button
      :hide-new-button="readOnly || controlAdd"
      :hide-delete-button="readOnly"
      :no-confirm-delete="!singleSave"
      @new-data="onGridNewRecord"
      @delete-data="onGridRowDelete"
      @rowFieldChanged="onGridRowFieldChanged"
      autoCommitLine
      hidePaging
    >
      <template #item_FileName="{ item }">
        <a
          v-if="item._id"
          class="text-blue-600 hover:text-blue-800 visited:text-purple-600"
          :href="`/v1/asset/view?id=${item._id}`"
          target="_blank"
          >{{ item.FileName }}</a
        >
        <p v-else>{{ item.FileName }}</p>
      </template>
      <template #item_Content="{ item, header }">
        <template v-if="readOnly">&nbsp;</template>
        <template v-else>
          <input
            v-if="isSameJournal(item.RefID, item.Kind)"
            :type="header.input.kind"
            :placeholder="header.input.caption || header.input.label"
            class="input_field -mt-1"
            ref="control"
            @change="(file) => FileReadertoBase64(file, header.input, item)"
            :disabled="header.input.disabled"
          />
          <template v-else>&nbsp;</template>
        </template>
      </template>
      <template #item_ByTag="{ item }">
        <template v-if="item._id == '' || item._id == null"> &nbsp; </template>
        <template v-else>
          <mdicon
            name="close-thick"
            class="text-rose-500"
            v-if="isSameJournal(item.RefID, item.Kind)"
            size="16"
          />
          <mdicon name="check-bold" class="text-green-500" v-else size="16" />
        </template>
      </template>
      <template #item_Description="{ item }">
        <template v-if="readOnly"> {{ item.Description }} </template>
        <template v-else>
          <template v-if="!isSameJournal(item.RefID, item.Kind)">
            {{ item.Description }}
          </template>
        </template>
      </template>
      <template #item_button_recordchange="{ item, config }">
        <a
          href="#"
          v-if="singleSave && item.isChange"
          @click="saveAsset(item)"
          class="save_action"
        >
          <mdicon
            name="content-save"
            width="16"
            alt="edit"
            class="cursor-pointer hover:text-primary"
          />
        </a>
      </template>
    </s-grid>
  </div>
</template>

<script setup>
import { reactive, ref, inject, onMounted, watch, computed } from "vue";
import { SGrid, util, loadGridConfig } from "suimjs";
const axios = inject("axios");

import helper from "@/scripts/helper.js";
import Loader from "@/components/common/Loader.vue";
const gridref = ref(null);

const props = defineProps({
  journalId: { type: String, default: "" },
  journalIds: { type: Array, default: () => [] },
  journalType: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
  tags: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false },
  singleSave: { type: Boolean, default: false },
  gridConfig: { type: [String, Object, undefined], default: () => undefined },
  isSingleUpload: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  preSave: null,
  preGets: null,
  postSave: null,
  postDelete: null,
});

defineExpose({
  Save: saveAssetMultiple,
  refreshGrid,
});

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  _idDelete: [],
  listCfg: {},
  loading: false,
});
function isSameJournal(refId, kind) {
  return refId == props.journalId && kind == props.journalType;
}
function FileReadertoBase64(eventfile, config, value) {
  value.IsEdit = true;
  const file = eventfile.target.files[0];
  const reader = new FileReader();
  reader.readAsDataURL(file);

  reader.onloadend = () => {
    value[config.field] = reader.result.split(",")[1];
    value["OriginalFileName"] = file.name;
    value["FileName"] = file.name;
    value["ContentType"] = file.type;
  };
}

function updateItems() {
  if (!gridref.value) return;
  const r = data.records.filter((e) => e.IsDelete !== true);
  gridref.value.setRecords(r);
  emit("update:modelValue", data.records);
}

function onGridNewRecord() {
  const r = {};
  r.OriginalFileName = "";
  r.ContentType = "";
  r.Descriptions = "";
  r.UploadDate = new Date();
  r.Tags = [];
  if (props.singleSave) {
    r.isChange = true;
  }
  r.SameJournal = true;
  r.Kind = props.journalType;
  r.RefID = props.journalId;

  data.records.push(r);
  updateItems();
}

function onGridRowDelete(_, index) {
  const record = data.records[index];
  if (props.singleSave && record._id) {
    rowDelete(record);
  } else {
    if (data.records[index]._id) data._idDelete.push(data.records[index]._id);
    const newRecords = data.records.filter((dt, idx) => {
      return idx != index;
    });

    data.records = newRecords;
    updateItems();
  }
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
  record.isChange = true;
  updateItems();
}

async function GetDataAssets() {
  try {
    let param = {
      JournalType: props.journalType,
      JournalID: props.journalId,
      Tags: props.tags,
    };

    emit("preGets", param);

    const resp = await axios.post("/asset/read-by-journal", param);

    return resp.data.map((resp) => ({
      _id: resp._id,
      OriginalFileName: resp.OriginalFileName,
      Content: resp.Data.Content,
      ContentType: resp.ContentType,
      FileName: resp.OriginalFileName,
      UploadDate: resp.Data.UploadDate,
      Description: resp.Data.Description,
      URI: resp.URI,
      Kind: resp.Kind,
      Tags: resp.Tags,
      RefID: resp.RefID,
    }));
  } catch (error) {
    util.showError(error);
  }
}
async function generateGridConfig() {
  switch (typeof props.gridConfig) {
    case "string":
      data.listCfg = await loadGridConfig(axios, props.gridConfig).then(
        (r) => {
          alterGridConfig(r);
        },
        (e) => util.showError(e)
      );
      break;

    case "object":
      data.listCfg = props.gridConfig;
      break;

    default:
      const cfg = await loadGridConfig(axios, "/asset/gridconfig");
      data.listCfg = alterGridConfig(cfg);
      break;
  }
  let counting = 0;
  for (const item of InitGridConfig.fields) {
    switch (item.field) {
      case "FileName":
        counting++;
        break;

      case "Content":
        counting++;
        break;
    }
  }

  if (counting < 2) {
    util.showError("Content or FileName in Grid Config Not found");
    return;
  }
}
async function generateRecords() {
  gridref.value.setLoading(true);
  if (props.modelValue.length > 0) {
    data.records = props.modelValue;
  } else {
    data.records = await GetDataAssets();
  }
  gridref.value.setLoading(false);

  if (data.records.length == 0 && props.isSingleUpload) {
    onGridNewRecord();
  }
  updateItems();
}

onMounted(async () => {
  generateGridConfig();
  generateRecords();
});

async function refreshGrid() {
  gridref.value.setLoading(true);
  const resps = await GetDataAssets();
  data.records = resps;
  gridref.value.setLoading(false);
  util.nextTickN(2, () => {
    updateItems();
  });
}
async function saveAsset(record) {
  if (!record.Description) {
    return util.showError("Description is required");
  }
  const DataRecord = { ...record };
  delete DataRecord.OriginalFileName;
  delete DataRecord.ContentType;
  delete DataRecord.ReadOnly;
  delete DataRecord.Content;

  const param = {
    Content: record.Content,
    Asset: {
      _id: record._id,
      OriginalFileName: record.OriginalFileName,
      ContentType: record.ContentType,
      Kind: record.Kind,
      RefID: record.RefID,
      Data: DataRecord,
      URI: record.URI,

      Tags: props.tags,
    },
  };

  emit("preSave", param);
  try {
    await axios.post(`/asset/write-asset`, param);
    refreshGrid();
  } catch (error) {
    util.showError(error);
  } finally {
    console.log("w");
    emit("postSave");
  }
}
async function saveAssetMultiple(
  journalId,
  journalType,
  cbOK = () => {},
  cbFalse = () => {}
) {
  try {
    data.loading = true;
    journalId = props.journalId != "" ? props.journalId : journalId;
    journalType = props.journalType != "" ? props.journalType : journalType;

    if (journalId == "") {
      throw "Journal ID not found";
    }

    if (journalType == "") {
      throw "Journal Type not found";
    }

    if (data._idDelete.length > 0) {
      const promisebatch = [];
      for (const _id of data._idDelete) {
        promisebatch.push(deleteAsset(_id));
      }
      await Promise.all(promisebatch);
      data._idDelete = [];
    }
    if (data.records.length > 0) {
      const writecontents = [];
      for (const record of data.records) {
        if (record._id) {
          if (!record.isChange) continue;
        } else {
          if (!record.Content || record.Content == "") continue;
        }

        const DataRecord = {
          ...record,
          IsEdit: true,
          SameJournal: true,
          suimRecordChange: true,
          RefID: journalId,
          Kind: props.journalType,
        };
        delete DataRecord.OriginalFileName;
        delete DataRecord.ContentType;
        delete DataRecord.ReadOnly;
        delete DataRecord.Content;

        const fieldcfgs = data.listCfg.fields.map((field) => field.field);
        for (const key of Object.keys(DataRecord)) {
          if (!fieldcfgs.includes(key)) {
            delete DataRecord[key];
          }
        }

        const writecontent = {
          Content: record.Content,
          Asset: {
            _id: record._id,
            OriginalFileName: record.OriginalFileName,
            ContentType: record.ContentType,
            Kind: props.journalType,
            RefID: journalId,
            Data: DataRecord,
            Tags: record.Tags,
            URI: record.URI,
          },
        };

        writecontents.push(writecontent);
      }

      emit("preSave", writecontents);

      await axios.post(`/asset/write-batch-asset`, writecontents);
      emit("postSave", writecontents);

      const resps = await GetDataAssets();
      data.records = resps;
      emit("update:modelValue", resps);
      cbOK();
      return resps;
    }
  } catch (error) {
    cbFalse();
    util.showError(error);
  } finally {
    data.loading = false;
    updateItems();
  }
}

async function rowDelete(record) {
  if (!isSameJournal(record.RefID, record.Kind)) {
    deleteTag(record);
  } else {
    await deleteAsset(record._id);
    refreshGrid();
  }
}

function deleteTag(record) {
  record.Tags = record.Tags.reduce((arr, el) => {
    if (!props.tags.includes(el)) arr.push(el);
    return arr;
  }, []);
  saveAsset(record);
}
async function deleteAsset(id) {
  try {
    const resp = await axios.post(`/asset/delete`, id);

    return resp.data;
  } catch (error) {
    util.showError(error);
  } finally {
    emit("postDelete");
  }
}

function alterGridConfig(cfg) {
  cfg.fields.splice(
    4,
    0,
    helper.gridColumnConfig({ field: "ByTag", label: "By Tag" })
  );
  return cfg;
}

watch(
  () => props.modelValue,
  (nv) => {
    if (nv.length > 0) {
      data.records = nv;
      gridref.value.setRecords(data.records);
    }
  }
);

const controlAdd = computed({
  get() {
    let res = false;
    if (props.isSingleUpload && data.records.length == 1) res = true;
    return res;
  },
});

const InitGridConfig = {
  setting: {
    idField: "",
    keywordFields: ["_id", "Name"],
    sortable: ["_id"],
  },
  fields: [
    {
      field: "FileName",
      kind: "text",
      label: "File name",
      halign: "start",
      valign: "start",
      labelField: "",
      length: 0,
      width: "",
      pos: 1000,
      readType: "show",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      input: {
        field: "FileName",
        label: "File name",
        hint: "",
        hide: false,
        placeHolder: "File name",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
        multiRow: 1,
        minLength: 0,
        maxLength: 999,
        readOnly: true,
        readOnlyOnEdit: false,
        readOnlyOnNew: false,
        useList: false,
        allowAdd: false,
        items: [],
        useLookup: false,
        lookupUrl: "",
        lookupKey: "",
        lookupLabels: null,
        lookupSearchs: null,
        lookupFormat1: "",
        lookupFormat2: "",
        showTitle: false,
        showHint: false,
        showDetail: false,
        fixTitle: false,
        fixDetail: false,
        section: "General",
        sectionWidth: "",
        row: 0,
        col: 0,
        labelField: "",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        width: "",
        spaceBefore: 0,
        spaceAfter: 0,
      },
    },
    {
      field: "Description",
      kind: "text",
      label: "Description",
      halign: "start",
      valign: "start",
      labelField: "",
      length: 0,
      width: "",
      pos: 1000,
      readType: "show",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      input: {
        field: "Description",
        label: "Description",
        hint: "",
        hide: false,
        placeHolder: "Description",
        kind: "text",
        disable: false,
        required: true,
        multiple: false,
        multiRow: 1,
        minLength: 0,
        maxLength: 999,
        readOnly: false,
        readOnlyOnEdit: false,
        readOnlyOnNew: false,
        useList: false,
        allowAdd: false,
        items: [],
        useLookup: false,
        lookupUrl: "",
        lookupKey: "",
        lookupLabels: null,
        lookupSearchs: null,
        lookupFormat1: "",
        lookupFormat2: "",
        showTitle: false,
        showHint: false,
        showDetail: false,
        fixTitle: false,
        fixDetail: false,
        section: "General",
        sectionWidth: "",
        row: 0,
        col: 0,
        labelField: "",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        width: "",
        spaceBefore: 0,
        spaceAfter: 0,
      },
    },
    {
      field: "UploadDate",
      kind: "date",
      label: "Upload date",
      halign: "start",
      valign: "start",
      labelField: "",
      length: 0,
      width: "",
      pos: 1000,
      readType: "show",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      input: {
        field: "UploadDate",
        label: "Upload date",
        hint: "",
        hide: false,
        placeHolder: "Upload date",
        kind: "date",
        disable: false,
        required: false,
        multiple: false,
        multiRow: 1,
        minLength: 0,
        maxLength: 999,
        readOnly: true,
        readOnlyOnEdit: false,
        readOnlyOnNew: false,
        useList: false,
        allowAdd: false,
        items: [],
        useLookup: false,
        lookupUrl: "",
        lookupKey: "",
        lookupLabels: null,
        lookupSearchs: null,
        lookupFormat1: "",
        lookupFormat2: "",
        showTitle: false,
        showHint: false,
        showDetail: false,
        fixTitle: false,
        fixDetail: false,
        section: "General",
        sectionWidth: "",
        row: 0,
        col: 0,
        labelField: "",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        width: "",
        spaceBefore: 0,
        spaceAfter: 0,
      },
    },
    {
      field: "Content",
      kind: "file",
      label: "File",
      halign: "start",
      valign: "start",
      labelField: "",
      length: 0,
      width: "",
      pos: 1000,
      readType: "show",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      input: {
        field: "Content",
        label: "File",
        hint: "",
        hide: false,
        placeHolder: "File",
        kind: "file",
        disable: false,
        required: false,
        multiple: false,
        multiRow: 1,
        minLength: 0,
        maxLength: 999,
        readOnly: false,
        readOnlyOnEdit: false,
        readOnlyOnNew: false,
        useList: false,
        allowAdd: false,
        items: [],
        useLookup: false,
        lookupUrl: "",
        lookupKey: "",
        lookupLabels: null,
        lookupSearchs: null,
        lookupFormat1: "",
        lookupFormat2: "",
        showTitle: false,
        showHint: false,
        showDetail: false,
        fixTitle: false,
        fixDetail: false,
        section: "General",
        sectionWidth: "",
        row: 0,
        col: 0,
        labelField: "",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        width: "",
        spaceBefore: 0,
        spaceAfter: 0,
      },
    },
  ],
};
</script>
